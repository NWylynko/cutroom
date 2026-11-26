// Package render turns a timeline into a video. Video is rendered in chunks split at cuts (L10,
// L11), each segment read with its own seek and resampled onto the audio-anchored frame grid (L02,
// L09), encoded once (L12), frame-count checked (L14), cached by content and joined by stream copy.
// Audio is assembled sample-exactly in Go and loudness-normalised by measurement (L17, L43).
package render

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/NWylynko/cutroom/internal/align"
	"github.com/NWylynko/cutroom/internal/deps"
	"github.com/NWylynko/cutroom/internal/media"
	"github.com/NWylynko/cutroom/internal/timeline"
)

// Version is part of every cache key; bump when output would change.
const Version = 3 // 2: HDR sources tone-mapped to SDR bt709 (L23); 3: GPU pre-scale keeps the source shape (filled HDR frames were squashed)

type Source struct {
	Path       string
	AudioStart float64
	HasAudio   bool
	Size       int64
	ModTime    time.Time
	Sync       *align.Sync // set when this source's audio is aligned to another source
	W, H       int         // picture as displayed (rotation applied); 0 = unknown
}

type Options struct {
	Preview     bool
	Range       [2]float64 // output seconds; zero = all
	Jobs        int
	Target      float64 // chunk length (s)
	CacheDir    string
	Loudness    float64 // LUFS
	TruePeak    float64 // dBTP
	Chain       string
	DryRun      bool
	Progress    func(string)
	ProjDir     string               // overlay image paths are relative to it
	Transcripts timeline.Transcripts // for burned-in captions
	Overlays    []Placement          // set by Render: overlays in the planned timeline's output time
}

type Chunk struct {
	Index    int            `json:"index"`
	Segments []int          `json:"segments"`
	Start    float64        `json:"output_start"`
	Duration float64        `json:"duration"`
	Frames   int64          `json:"frames"`
	Key      string         `json:"key"`
	Cached   bool           `json:"cached"`
	Seconds  float64        `json:"render_seconds,omitempty"`
	Overlays []ChunkOverlay `json:"overlays,omitempty"`
	file     string
}

// ChunkOverlay is a placement's part inside one chunk, in chunk frames [From, To).
type ChunkOverlay struct {
	P        Placement `json:"placement"`
	From, To int64
	FadeIn   bool `json:"fade_in"`
	FadeOut  bool `json:"fade_out"`
	AnimIn   bool `json:"anim_in,omitempty"`
	AnimOut  bool `json:"anim_out,omitempty"`
	// Frames: burned-in captions as runs of chunk frames, each showing one card ("" = nothing)
	Frames []CapRun `json:"frames,omitempty"`
}

// CapRun is a run of chunk frames [From, To) showing one caption card (PNG "" = none).
type CapRun struct {
	PNG      string `json:"png"`
	From, To int64
}

// captionFrames lays the caption groups over a chunk's frames [c0, c0+n) as runs, gaps included.
func captionFrames(p Placement, c0, n int64, fps float64) []CapRun {
	var rs []CapRun
	at := int64(0)
	for _, g := range p.Groups {
		a, b := int64(math.Round(g.Start*fps))-c0, int64(math.Round(g.End*fps))-c0
		a, b = max(a, at), min(b, n)
		if b <= a {
			continue
		}
		if a > at {
			rs = append(rs, CapRun{"", at, a})
		}
		rs = append(rs, CapRun{g.PNG, a, b})
		at = b
	}
	if len(rs) == 0 {
		return nil
	}
	if at < n {
		rs = append(rs, CapRun{"", at, n})
	}
	return rs
}

// capList is the ffconcat list that plays caption runs as one clip: each card for its run's frames.
func capList(runs []CapRun, blank string, fps float64) string {
	var b strings.Builder
	b.WriteString("ffconcat version 1.0\n")
	q := func(p string) string { return "'" + strings.ReplaceAll(p, "'", `'\''`) + "'" }
	last := ""
	for _, r := range runs {
		p := r.PNG
		if p == "" {
			p = blank
		}
		fmt.Fprintf(&b, "file %s\nduration %s\n", q(p), f6(float64(r.To-r.From)/fps))
		last = p
	}
	fmt.Fprintf(&b, "file %s\n", q(last)) // the concat demuxer needs the last file once more
	return b.String()
}

func capListPath(c Chunk, k int) string {
	return strings.TrimSuffix(c.file, ".mp4") + fmt.Sprintf(".cap%d.ffconcat", k)
}

type Result struct {
	Output     string      `json:"output"`
	Duration   float64     `json:"duration"`
	Expected   float64     `json:"expected_duration"`
	Frames     int64       `json:"frames"`
	Width      int         `json:"width"`
	Height     int         `json:"height"`
	Chunks     []Chunk     `json:"chunks"`
	Audio      AudioResult `json:"audio"`
	Seconds    float64     `json:"render_seconds"`
	Estimate   float64     `json:"estimate_seconds"`
	Preview    bool        `json:"preview"`
	RangeStart float64     `json:"range_start"`
	DryRun     bool        `json:"dry_run"`
	Overlays   []Placement `json:"overlays,omitempty"` // in the file's time
}

// size returns output dimensions (even), preview at 540 p.
func size(tl *timeline.Timeline, preview bool) (int, int) {
	w, h := tl.Width, tl.Height
	if preview && h > 540 {
		w = int(math.Round(float64(w)*540/float64(h)/2)) * 2
		h = 540
	}
	return w - w%2, h - h%2
}

func encArgs(preview bool) []string {
	if preview {
		return []string{"-c:v", "libx264", "-preset", "veryfast", "-crf", "26"}
	}
	return []string{"-c:v", "libx264", "-preset", "medium", "-crf", "18"}
}

func hash(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return fmt.Sprintf("%x", h[:10])
}

// Plan splits the timeline into chunks at segment boundaries of about target seconds.
func Plan(tl *timeline.Timeline, srcs map[string]Source, o Options) []Chunk {
	var chunks []Chunk
	cur := Chunk{}
	acc := 0.0
	flush := func() {
		if len(cur.Segments) > 0 {
			cur.Index = len(chunks)
			chunks = append(chunks, cur)
		}
		cur = Chunk{}
		acc = 0
	}
	for i, g := range tl.Segments {
		if len(cur.Segments) == 0 {
			cur.Start = g.OutStart
		}
		cur.Segments = append(cur.Segments, i)
		f := tl.Frames(g)
		cur.Frames += f
		acc += float64(f) / tl.Rate()
		end := int64(math.Round(cur.Start*tl.Rate())) + cur.Frames
		if acc >= o.Target && !inFade(o.Overlays, end, tl.Rate()) {
			flush()
		}
	}
	flush()
	w, h := size(tl, o.Preview)
	for i := range chunks {
		c := &chunks[i]
		c.Duration = round6(float64(c.Frames) / tl.Rate())
		var segs []any
		for _, k := range c.Segments {
			g := tl.Segments[k]
			s := srcs[g.Source]
			segs = append(segs, []any{g.Source, g.In, g.Out, g.Zoom, g.AnchorX, g.AnchorY, s.Size, s.ModTime.UnixNano(), s.AudioStart})
		}
		// overlays over this chunk, in chunk frames
		c0 := int64(math.Round(c.Start * tl.Rate()))
		var ovs []any
		for _, p := range o.Overlays {
			pa, pb := int64(math.Round(p.Start*tl.Rate())), int64(math.Round(p.End*tl.Rate()))
			from, to := max(pa, c0)-c0, min(pb, c0+c.Frames)-c0
			if to <= from {
				continue
			}
			co := ChunkOverlay{P: p, From: from, To: to, FadeIn: pa >= c0 && p.Fade > 0, FadeOut: pb <= c0+c.Frames && p.Fade > 0}
			if p.Kind == "captions" {
				co.Frames = captionFrames(p, c0, c.Frames, tl.Rate())
				if len(co.Frames) == 0 {
					continue
				}
				ovs = append(ovs, []any{"captions", co.Frames})
				c.Overlays = append(c.Overlays, co)
				continue
			}
			co.AnimIn, co.AnimOut = p.Anim != "" && pa >= c0, p.Anim != "" && pb <= c0+c.Frames
			c.Overlays = append(c.Overlays, co)
			ovs = append(ovs, []any{p.Key, from, to, co.FadeIn, co.FadeOut, p.Fade, pa - c0, pb - c0, p.Anim, p.AnimTime, co.AnimIn, co.AnimOut})
		}
		parts := []any{Version, tl.FPS, w, h, o.Preview, encArgs(o.Preview), segs, tl.Grid}
		if tl.Bars != "" && !tl.Fill {
			parts = append(parts, []any{"bars", tl.Bars})
		}
		if tl.Fill { // keys of fitted timelines are unchanged
			parts = append(parts, []any{"fill", tl.FillX, tl.FillY})
		}
		if len(ovs) > 0 { // keys of chunks without overlays are unchanged
			parts = append(parts, ovs)
		}
		c.Key = hash(parts)
		c.file = filepath.Join(o.CacheDir, "chunks", c.Key+".mp4")
		if _, err := os.Stat(c.file); err == nil {
			c.Cached = true
		}
	}
	return chunks
}

// chunkCmd builds the ffmpeg args for one chunk.
func chunkCmd(tl *timeline.Timeline, srcs map[string]Source, c Chunk, preview bool, out string) []string {
	w, h := size(tl, preview)
	fps := tl.FPS
	args := []string{"-hide_banner", "-loglevel", "error", "-y"}
	var fc []string
	var labels string
	for n, k := range c.Segments {
		g := tl.Segments[k]
		s := srcs[g.Source]
		margin := 0.25 // seek early so a VFR frame stored before its slot is still decoded (L09)
		seek := math.Max(0, g.In-margin)
		lead := g.In - seek
		nf := tl.Frames(g)
		dur := lead + float64(nf+3)/tl.Rate() // 3 spare frames (L09)
		z := g.ZoomOf()
		// HDR → SDR bt709 on the way in (L23); a zoomed segment needs z× the pixels before its crop
		gw, gh := gpuSize(tl.Fill, s.W, s.H, w, h, z)
		sdr := media.ReadSDR(s.Path, gw, gh)
		args = append(args, "-ss", f6(seek), "-t", f6(dur))
		args = append(args, sdr.Pre...)
		args = append(args, "-i", s.Path)
		chain := fmt.Sprintf("[%d:v]", n) + sdr.Chain(fmt.Sprintf("fps=fps=%s:start_time=%s:round=near,trim=end_frame=%d,setpts=PTS-STARTPTS", fps, f6(lead), nf))
		if z > 1 { // punch-in (L40): crop around the anchor, kept inside the frame, then scale back up
			ax, ay := g.Anchor()
			chain += "," + ZoomCrop(z, ax, ay)
		}
		chain += "," + FrameFit(tl.Fill, tl.FillX, tl.FillY, tl.Bars, fmt.Sprintf("s%d", n), w, h)
		chain += ",format=yuv420p" + fmt.Sprintf("[v%d]", n)
		fc = append(fc, chain)
		labels += fmt.Sprintf("[v%d]", n)
	}
	// restamp after concat so every frame sits exactly on the grid (L08)
	base := "out"
	if len(c.Overlays) > 0 {
		base = "base0"
	}
	fc = append(fc, fmt.Sprintf("%sconcat=n=%d:v=1:a=0,settb=AVTB,setpts=N/(FRAME_RATE*TB),setparams=colorspace=bt709:color_primaries=bt709:color_trc=bt709:range=tv[%s]", labels, len(c.Segments), base))
	// overlays: each card is a looped still, shown by frame number (exact on the grid) and repeated
	// rather than ended (L16); fades are alpha fades counted in frames
	for k, co := range c.Overlays {
		in := len(c.Segments) + k
		var chain string
		if co.P.Kind == "captions" { // every caption group of the chunk as one alpha clip (L48)
			args = append(args, "-f", "concat", "-safe", "0", "-i", capListPath(c, k))
			fc = append(fc, fmt.Sprintf("[%d:v]fps=fps=%s:round=near,format=rgba,trim=end_frame=%d,setpts=PTS-STARTPTS[o%d]", in, fps, c.Frames, k))
			next := fmt.Sprintf("base%d", k+1)
			if k == len(c.Overlays)-1 {
				next = "ovl"
			}
			fc = append(fc, fmt.Sprintf("[base%d][o%d]overlay=0:0:eof_action=repeat[%s]", k, k, next))
			continue
		}
		x0, y0 := 0, 0
		if co.P.Kind == "clip" { // B-roll: its frames from the right moment, scaled into its box, muted
			bx, by := int(math.Round(co.P.Box[0]*float64(w))), int(math.Round(co.P.Box[1]*float64(h)))
			bw, bh := even(co.P.Box[2]*float64(w)-float64(bx)), even(co.P.Box[3]*float64(h)-float64(by))
			x0, y0 = bx, by
			n := co.To - co.From
			seek := co.P.ClipIn + (c.Start + float64(co.From)/tl.Rate() - co.P.Start)
			sdr := media.ReadSDR(co.P.Clip, 0, 0) // HDR clips tone-mapped, rotation applied, like sources
			args = append(args, "-ss", f6(math.Max(0, seek)), "-t", f6(float64(n+3)/tl.Rate()))
			args = append(args, sdr.Pre...)
			args = append(args, "-i", co.P.Clip)
			// a clip shorter than its window holds its last frame (tpad clone); then it is shifted to
			// start on chunk frame From (transparent before it)
			chain = fmt.Sprintf("[%d:v]", in) + sdr.Chain(fmt.Sprintf("fps=fps=%s:round=near,scale=%d:%d:force_original_aspect_ratio=increase:flags=lanczos,crop=%d:%d,setsar=1,setparams=colorspace=bt709:color_primaries=bt709:color_trc=bt709:range=tv,format=yuva420p,tpad=stop=-1:stop_mode=clone,trim=end_frame=%d,setpts=PTS-STARTPTS,tpad=start=%d:start_mode=add:color=black@0",
				fps, bw, bh, bw, bh, n, co.From))
		} else {
			args = append(args, "-loop", "1", "-framerate", fps, "-t", f6(float64(c.Frames+3)/tl.Rate()), "-i", co.P.PNG)
			chain = fmt.Sprintf("[%d:v]format=rgba", in)
		}
		nf := int64(math.Max(1, math.Round(co.P.Fade*tl.Rate())))
		if co.FadeIn {
			chain += fmt.Sprintf(",fade=t=in:s=%d:n=%d:alpha=1", co.From, nf)
		}
		if co.FadeOut {
			chain += fmt.Sprintf(",fade=t=out:s=%d:n=%d:alpha=1", co.To-nf, nf)
		}
		next := fmt.Sprintf("base%d", k+1)
		if k == len(c.Overlays)-1 {
			next = "ovl"
		}
		if co.P.Kind != "clip" && (co.AnimIn || co.AnimOut) {
			crop, ox, oy := cardMotion(co, w, h, tl.Rate())
			fc = append(fc, chain+crop+fmt.Sprintf("[o%d]", k),
				fmt.Sprintf("[base%d][o%d]overlay=x='%s':y='%s':eval=frame:eof_action=repeat:enable='between(n,%d,%d)'[%s]", k, k, ox, oy, co.From, co.To-1, next))
			continue
		}
		fc = append(fc, chain+fmt.Sprintf("[o%d]", k),
			fmt.Sprintf("[base%d][o%d]overlay=%d:%d:eof_action=repeat:enable='between(n,%d,%d)'[%s]", k, k, x0, y0, co.From, co.To-1, next))
	}
	if len(c.Overlays) > 0 {
		// the looped stills run a little longer than the chunk: end on the chunk's last frame (L14)
		fc = append(fc, fmt.Sprintf("[ovl]trim=end_frame=%d,format=yuv420p,setparams=colorspace=bt709:color_primaries=bt709:color_trc=bt709:range=tv[out]", c.Frames))
	}
	args = append(args, "-filter_complex", strings.Join(fc, ";"), "-map", "[out]", "-an")
	args = append(args, encArgs(preview)...)
	args = append(args, "-pix_fmt", "yuv420p", "-r", fps, "-fps_mode", "cfr",
		"-colorspace", "bt709", "-color_primaries", "bt709", "-color_trc", "bt709", "-color_range", "tv",
		"-video_track_timescale", "90000", "-movflags", "+faststart", out)
	return args
}

// inFade says whether output frame boundary b falls inside an overlay's fade in or fade out, where a
// chunk boundary would show as a seam (L13).
func inFade(ps []Placement, b int64, fps float64) bool {
	for _, p := range ps {
		d := p.Fade
		if p.Anim != "" {
			d = math.Max(d, p.AnimTime)
		}
		if d <= 0 {
			continue
		}
		nf := int64(math.Max(1, math.Round(d*fps)))
		pa, pb := int64(math.Round(p.Start*fps)), int64(math.Round(p.End*fps))
		if (b > pa && b < pa+nf) || (b > pb-nf && b < pb) {
			return true
		}
	}
	return false
}

// joinArgs joins encoded chunks by stream copy and muxes the finished audio — no second encode (L12).
func joinArgs(list, audio, out string) []string {
	return []string{"-hide_banner", "-loglevel", "error", "-y", "-f", "concat", "-safe", "0", "-i", list,
		"-i", audio, "-map", "0:v", "-map", "1:a", "-c", "copy", "-movflags", "+faststart", out}
}

// renderChunk encodes one chunk from the sources and checks its frame count (L14).
func renderChunk(tl *timeline.Timeline, srcs map[string]Source, c *Chunk, preview bool) error {
	if err := ensureDir(c.file); err != nil {
		return err
	}
	tmp := strings.TrimSuffix(c.file, ".mp4") + ".tmp.mp4"
	for k, co := range c.Overlays {
		if co.P.Kind != "captions" {
			continue
		}
		w, h := size(tl, preview)
		blank, err := blankPNG(filepath.Dir(filepath.Dir(c.file)), w, h)
		if err != nil {
			return err
		}
		if err := os.WriteFile(capListPath(*c, k), []byte(capList(co.Frames, blank, tl.Rate())), 0o644); err != nil {
			return err
		}
	}
	t0 := time.Now()
	if _, err := deps.Run("ffmpeg", chunkCmd(tl, srcs, *c, preview, tmp)...); err != nil {
		return err
	}
	n, err := countFrames(tmp)
	if err != nil {
		return err
	}
	if n != c.Frames {
		os.Remove(tmp)
		return fmt.Errorf("chunk %d has %d frames, expected %d (L14) — a source may end early or be unreadable around output %.3f s", c.Index, n, c.Frames, c.Start)
	}
	c.Seconds = round3(time.Since(t0).Seconds())
	return os.Rename(tmp, c.file)
}

func countFrames(p string) (int64, error) {
	out, err := deps.Run("ffprobe", "-v", "error", "-select_streams", "v:0", "-count_packets", "-show_entries", "stream=nb_read_packets", "-of", "csv=p=0", p)
	if err != nil {
		return 0, err
	}
	first := strings.TrimSpace(strings.Split(string(out), "\n")[0])
	first, _, _ = strings.Cut(first, ",") // side data adds empty csv fields
	return strconv.ParseInt(first, 10, 64)
}

// Render renders the timeline (or a range) to out.
func Render(tl *timeline.Timeline, srcs map[string]Source, out string, o Options) (*Result, error) {
	t0 := time.Now()
	if o.Jobs <= 0 {
		o.Jobs = 4
	}
	if o.Target <= 0 {
		o.Target = 45
	}
	say := o.Progress
	if say == nil {
		say = func(string) {}
	}
	w, h := size(tl, o.Preview)
	ps, err := Placements(tl, o.Transcripts, o.ProjDir, o.CacheDir, w, h)
	if err != nil {
		return nil, err
	}
	o.Overlays = ps
	all := Plan(tl, srcs, o)
	// a range renders exactly that part of the edit (whole frames): vtl is the trimmed edit
	vtl, chunks := tl, all
	res := &Result{Output: out, Width: w, Height: h, Preview: o.Preview, DryRun: o.DryRun, Overlays: ps}
	if o.Range[1] > 0 {
		vtl, res.RangeStart = tl.Trim(o.Range[0], o.Range[1])
		ro := o
		ro.Overlays = nil
		for _, p := range ps { // into the range file's time
			p.Start, p.End = round6(math.Max(0, p.Start-res.RangeStart)), round6(math.Min(vtl.Duration(), p.End-res.RangeStart))
			var gs []CapGroup
			for _, g := range p.Groups {
				g.Start, g.End = round6(math.Max(0, g.Start-res.RangeStart)), round6(math.Min(vtl.Duration(), g.End-res.RangeStart))
				if g.End > g.Start {
					gs = append(gs, g)
				}
			}
			p.Groups = gs
			if p.End > p.Start {
				ro.Overlays = append(ro.Overlays, p)
			}
		}
		res.Overlays = ro.Overlays
		chunks = Plan(vtl, srcs, ro)
	}
	for _, c := range chunks {
		res.Expected += c.Duration
		res.Frames += c.Frames
	}
	res.Expected = round6(res.Expected)
	res.Estimate = estimate(o, chunks, w*h)
	res.Chunks = chunks
	if o.DryRun {
		return res, nil
	}

	// audio (whole timeline measured once; L17)
	audioDone := make(chan error, 1)
	go func() {
		ar, err := renderAudio(tl, vtl, srcs, chunks, all, o)
		res.Audio = ar
		audioDone <- err
	}()

	// video chunks in parallel
	var wg sync.WaitGroup
	sem := make(chan struct{}, o.Jobs)
	errs := make([]error, len(chunks))
	var mu sync.Mutex
	done := 0
	for i := range chunks {
		if chunks[i].Cached {
			continue
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			errs[i] = renderChunk(vtl, srcs, &chunks[i], o.Preview)
			mu.Lock()
			done++
			say(fmt.Sprintf("chunk %d/%d rendered (%.1f s of video in %.1f s)", chunks[i].Index+1, len(all), chunks[i].Duration, chunks[i].Seconds))
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	for _, e := range errs {
		if e != nil {
			return nil, e
		}
	}
	if err := <-audioDone; err != nil {
		return nil, fmt.Errorf("audio: %w", err)
	}
	recordTimings(o, chunks, w*h)

	// join by stream copy (L12) and mux
	if err := ensureDir(out); err != nil {
		return nil, err
	}
	list := filepath.Join(o.CacheDir, "render-list.txt")
	var b strings.Builder
	for _, c := range chunks {
		fmt.Fprintf(&b, "file '%s'\n", strings.ReplaceAll(c.file, "'", `'\''`))
	}
	if err := os.WriteFile(list, []byte(b.String()), 0o644); err != nil {
		return nil, err
	}
	tmp := out + ".tmp.mp4"
	if _, err := deps.Run("ffmpeg", joinArgs(list, res.Audio.File, tmp)...); err != nil {
		return nil, err
	}
	if err := os.Rename(tmp, out); err != nil {
		return nil, err
	}
	n, err := countFrames(out)
	if err != nil {
		return nil, err
	}
	res.Frames = n
	res.Duration = round6(float64(n) / tl.Rate())
	res.Seconds = round3(time.Since(t0).Seconds())
	if err := writeManifest(tl, vtl, srcs, chunks, res, o); err != nil {
		return nil, err
	}
	return res, nil
}

// AudioFor resolves which audio source a video source's segments use: its mapped recorder
// (timeline audio_for), else the timeline's audio when that is the source itself or is synced to it,
// else — for another take in a multi-take edit — the take's own sound (own = true). err is set when
// the timeline's audio can't be used for any segment (a hand-set, unaligned recorder, L05).
func AudioFor(tl *timeline.Timeline, srcs map[string]Source, video string) (audio string, own bool, err error) {
	if a := tl.AudioFor[video]; a != "" {
		s, ok := srcs[a]
		if !ok || !s.HasAudio {
			return "", false, fmt.Errorf("audio source %q for %q has no sound", a, video)
		}
		if a != video && (s.Sync == nil || s.Sync.To != video) {
			return "", false, fmt.Errorf("audio source %q is not aligned to %q: run `cutroom sync %s --to %s`", a, video, a, video)
		}
		return a, false, nil
	}
	if tl.Audio == video {
		return video, false, nil
	}
	main, ok := srcs[tl.Audio]
	if ok && main.Sync != nil && main.Sync.To == video {
		return tl.Audio, false, nil
	}
	// the timeline's audio belongs to some take in this edit? then other takes fall back to their own sound
	related := false
	for _, g := range tl.Segments {
		if g.Source == tl.Audio || (ok && main.Sync != nil && main.Sync.To == g.Source) {
			related = true
		}
	}
	if !related {
		return "", false, fmt.Errorf("audio source %q is not aligned to %q: run `cutroom sync %s --to %s`", tl.Audio, video, tl.Audio, video)
	}
	if v, ok := srcs[video]; ok && v.HasAudio {
		return video, true, nil
	}
	return "", false, fmt.Errorf("%q has no sound of its own and no recorder mapped to it: `cutroom edit audio <recorder> --for %s`", video, video)
}

// renderAudio assembles the sound for chunks (planned on vtl: tl itself, or tl trimmed to a range);
// loudness is always measured on the whole of tl (L17).
func renderAudio(tl, vtl *timeline.Timeline, srcs map[string]Source, chunks, all []Chunk, o Options) (AudioResult, error) {
	if tl.Audio == "" {
		return AudioResult{}, fmt.Errorf("the timeline has no audio source")
	}
	// which audio source each segment uses
	audioOf := map[string]string{}
	var ids []string
	for _, g := range tl.Segments {
		if _, done := audioOf[g.Source]; done {
			continue
		}
		a, _, err := AudioFor(tl, srcs, g.Source)
		if err != nil {
			return AudioResult{}, err
		}
		audioOf[g.Source] = a
		if !slices.Contains(ids, a) {
			ids = append(ids, a)
		}
	}
	cache := filepath.Join(o.CacheDir, "render")
	os.MkdirAll(cache, 0o755)
	chans := 0
	if len(ids) > 1 {
		chans = 1 // several sources: one channel layout so their pieces join
	}
	files := make([]string, len(ids))
	idx := map[string]int{}
	for i, id := range ids {
		src := srcs[id]
		files[i] = filepath.Join(cache, "src-"+hash([]any{src.Path, src.Size, src.ModTime.UnixNano(), chans})+".wav")
		if chans == 0 {
			files[i] = filepath.Join(cache, "src-"+hash([]any{src.Path, src.Size, src.ModTime.UnixNano()})+".wav")
		}
		if _, err := AudioWavCh(src.Path, files[i], chans); err != nil {
			return AudioResult{}, err
		}
		idx[id] = i
	}
	pieces := func(tl *timeline.Timeline, cs []Chunk) []Piece {
		var ps []Piece
		for _, c := range cs {
			for _, k := range c.Segments {
				g := tl.Segments[k]
				aid := audioOf[g.Source]
				src := srcs[aid]
				// frames × (SR/fps) samples exactly, so audio length always equals video length
				n := int64(math.Round(float64(tl.Frames(g)) * SR / tl.Rate()))
				if g.Source == aid {
					from := int64(math.Round((g.In - src.AudioStart) * SR))
					ps = append(ps, Piece{idx[aid], from, from + n})
					continue
				}
				// audio from another device: map through its measured offset + drift (L05), in
				// pieces of ≤ 10 s so drift never accumulates beyond half a sample per piece
				const step = 10 * SR
				for o := int64(0); o < n; o += step {
					m := min(step, n-o)
					t := g.In + float64(o)/SR
					from := int64(math.Round((src.Sync.Map(t) - src.AudioStart) * SR))
					ps = append(ps, Piece{idx[aid], from, from + m})
				}
			}
		}
		return ps
	}
	var syncs []any
	for _, id := range ids {
		syncs = append(syncs, srcs[id].Sync)
	}
	key := hash([]any{Version, files, pieces(tl, all), o.Chain, o.Loudness, o.TruePeak, syncs})
	allWav := filepath.Join(cache, "timeline-"+key+".wav")
	if _, err := os.Stat(allWav); err != nil {
		if _, err := Assemble(files, allWav, pieces(tl, all), 20); err != nil {
			return AudioResult{}, err
		}
	}
	// gain is measured on the whole timeline even for a range render; the key holds audio inputs only,
	// so a picture change (overlay, zoom, framing) never re-runs the audio (L17)
	rkey := hash([]any{key, pieces(vtl, chunks)})
	m4a := filepath.Join(cache, "audio-"+rkey+".m4a")
	meta := m4a + ".json"
	if b, err := os.ReadFile(meta); err == nil {
		var r AudioResult
		if json.Unmarshal(b, &r) == nil {
			if _, err := os.Stat(m4a); err == nil {
				r.Cached = true
				return r, nil
			}
		}
	}
	in := allWav
	if vtl != tl {
		in = filepath.Join(cache, "range-"+rkey+".wav")
		if _, err := Assemble(files, in, pieces(vtl, chunks), 20); err != nil {
			return AudioResult{}, err
		}
	}
	lufs, _, err := Loudness(allWav, o.Chain)
	if err != nil {
		return AudioResult{}, err
	}
	r, err := FinishAudio(in, m4a, o.Chain, o.Loudness, o.TruePeak, lufs)
	if err != nil {
		return r, err
	}
	b, _ := json.Marshal(r)
	os.WriteFile(meta, b, 0o644)
	return r, nil
}

// --- timing history → estimates (SPEC rule 8) ------------------------------------------------

type timing struct {
	Preview bool    `json:"preview"`
	Pixels  int     `json:"pixels"`
	PerSec  float64 `json:"render_s_per_output_s"`
}

func timingsPath(o Options) string { return filepath.Join(o.CacheDir, "timings.json") }

func recordTimings(o Options, cs []Chunk, px int) {
	var ts []timing
	if b, err := os.ReadFile(timingsPath(o)); err == nil {
		json.Unmarshal(b, &ts)
	}
	for _, c := range cs {
		if !c.Cached && c.Seconds > 0 && c.Duration > 0 {
			ts = append(ts, timing{o.Preview, px, round3(c.Seconds / c.Duration)})
		}
	}
	if len(ts) > 200 {
		ts = ts[len(ts)-200:]
	}
	b, _ := json.Marshal(ts)
	os.WriteFile(timingsPath(o), b, 0o644)
}

// estimate: wall seconds for the uncached chunks, from this project's history, else defaults
// measured on an M1 Pro (L56): 1080p final ≈ 0.35 s per output s, preview ≈ 0.08.
func estimate(o Options, cs []Chunk, px int) float64 {
	per := 0.35 * float64(px) / (1920 * 1080)
	if o.Preview {
		per = 0.08
	}
	var ts []timing
	if b, err := os.ReadFile(timingsPath(o)); err == nil {
		json.Unmarshal(b, &ts)
	}
	var xs []float64
	for _, t := range ts {
		if t.Preview == o.Preview {
			xs = append(xs, t.PerSec*float64(px)/float64(max(1, t.Pixels)))
		}
	}
	if len(xs) > 0 {
		sort.Float64s(xs)
		per = xs[len(xs)/2]
	}
	work := 0.0
	n := 0
	for _, c := range cs {
		if !c.Cached {
			work += c.Duration * per
			n++
		}
	}
	jobs := math.Min(float64(max(1, o.Jobs)), math.Max(1, float64(n)))
	return round3(work/jobs + 2)
}

func f6(x float64) string      { return strconv.FormatFloat(x, 'f', 6, 64) }
func round6(f float64) float64 { return math.Round(f*1e6) / 1e6 }
func round3(f float64) float64 { return math.Round(f*1e3) / 1e3 }

func even(x float64) int { return int(math.Round(x/2)) * 2 }

// FrameFit puts a (zoomed) source picture into the w×h output: fitted with bars, or — fill — scaled
// to cover it and cropped centred on (fx, fy) (0–1 across, down; the same meaning as ZoomCrop's
// anchor), clamped inside the picture, as a vertical short from landscape.
//
// A fitted picture's bars are black, a colour (bars "#rrggbb"), or "blur": the picture itself scaled
// to cover the frame, blurred and darkened, under the fitted one (tag keeps the graph's labels
// unique when several segments are in one graph).
func FrameFit(fill bool, fx, fy float64, bars, tag string, w, h int) string {
	if !fill {
		fit := fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease:flags=lanczos", w, h)
		switch {
		case bars == "blur":
			blur := max(8, min(w, h)/24)
			return fmt.Sprintf("split=2[%[1]sfg][%[1]sbg];[%[1]sbg]scale=%[2]d:%[3]d:force_original_aspect_ratio=increase,crop=%[2]d:%[3]d,boxblur=%[4]d:2,eq=brightness=-0.1:saturation=0.9[%[1]sbb];[%[1]sfg]%[5]s[%[1]sff];[%[1]sbb][%[1]sff]overlay=(W-w)/2:(H-h)/2,setsar=1",
				tag, w, h, blur, fit)
		case strings.HasPrefix(bars, "#"):
			return fmt.Sprintf("%s,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=0x%s,setsar=1", fit, w, h, strings.TrimPrefix(bars, "#"))
		}
		return fmt.Sprintf("%s,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,setsar=1", fit, w, h)
	}
	return fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=increase:flags=lanczos,crop=%d:%d:'clip(%s*iw-ow/2,0,iw-ow)':'clip(%s*ih-oh/2,0,ih-oh)',setsar=1", w, h, w, h, f6(fx), f6(fy))
}

// PictureVF is the filter that turns a source frame into what the render shows in a segment, before
// any final scaling: the punch-in crop, then (fill timelines) the crop to the output's shape.
// Comparisons (look, check) run source frames through it so they see what the render saw.
func PictureVF(zoom, ax, ay float64, fill bool, fx, fy float64, bars string, w, h int) string {
	var vf []string
	if zoom > 1.001 {
		vf = append(vf, ZoomCrop(zoom, ax, ay))
	}
	if fill || bars != "" { // a fitted frame with bars of its own: compare with them drawn
		vf = append(vf, FrameFit(fill, fx, fy, bars, "pv", w, h))
	}
	return strings.Join(vf, ",")
}

// ZoomCrop is the crop for a punch-in of z centred on (ax, ay) (normalised display coordinates),
// clamped inside the frame. Render and check use the same one.
func ZoomCrop(z, ax, ay float64) string {
	return fmt.Sprintf("crop=w='trunc(iw/%[1]g/2)*2':h='trunc(ih/%[1]g/2)*2':x='clip(%[2]g*iw-ow/2,0,iw-ow)':y='clip(%[3]g*ih-oh/2,0,ih-oh)'", z, ax, ay)
}

// Anchor returns a manifest segment's zoom centre (default 0.5, 0.4).
func (g MSeg) Anchor() (float64, float64) {
	if g.AnchorX == 0 && g.AnchorY == 0 {
		return 0.5, 0.4
	}
	return g.AnchorX, g.AnchorY
}

// cardMotion is a moving card's filter tail and overlay position: the card is cropped to its box
// (with room for its shadow), scaled per frame for pop, and placed per frame (x, y expressions in
// the chunk's frame number n) — easing in over AnimTime after it appears and out before it goes.
func cardMotion(co ChunkOverlay, w, h int, fps float64) (crop, x, y string) {
	pad := 0.06 * float64(min(w, h))
	bx0 := math.Max(0, math.Floor(co.P.Box[0]*float64(w)-pad))
	by0 := math.Max(0, math.Floor(co.P.Box[1]*float64(h)-pad))
	bx1 := math.Min(float64(w), math.Ceil(co.P.Box[2]*float64(w)+pad))
	by1 := math.Min(float64(h), math.Ceil(co.P.Box[3]*float64(h)+pad))
	cw, ch := int(bx1-bx0), int(by1-by0)
	bw, bh := (co.P.Box[2]-co.P.Box[0])*float64(w), (co.P.Box[3]-co.P.Box[1])*float64(h)
	nf := math.Max(1, math.Round(co.P.AnimTime*fps))
	pin, pout := "1", "1"
	if co.AnimIn {
		pin = fmt.Sprintf("clip((n-%d+1)/%g,0,1)", co.From, nf+1)
	}
	if co.AnimOut {
		pout = fmt.Sprintf("clip((%d-n)/%g,0,1)", co.To, nf+1)
	}
	P := fmt.Sprintf("min(%s,%s)", pin, pout)
	E := fmt.Sprintf("(1-pow(1-%s,3))", P)
	crop = fmt.Sprintf(",crop=%d:%d:%d:%d", cw, ch, int(bx0), int(by0))
	cx, cy := fmt.Sprintf("%d+(%d-w)/2", int(bx0), cw), fmt.Sprintf("%d+(%d-h)/2", int(by0), ch)
	switch co.P.Anim {
	case "pop":
		S := fmt.Sprintf("(0.6+0.4*(1+2.70158*pow(%[1]s-1,3)+1.70158*pow(%[1]s-1,2)))", P)
		crop += fmt.Sprintf(",scale=w='max(2,trunc(iw*%[1]s/2)*2)':h='max(2,trunc(ih*%[1]s/2)*2)':eval=frame", S)
		return crop, cx, cy
	case "slide-up":
		return crop, cx, fmt.Sprintf("%s+%g*(1-%s)", cy, 0.6*bh, E)
	case "slide-down":
		return crop, cx, fmt.Sprintf("%s-%g*(1-%s)", cy, 0.6*bh, E)
	case "slide-left":
		return crop, fmt.Sprintf("%s+%g*(1-%s)", cx, 0.3*bw, E), cy
	case "slide-right":
		return crop, fmt.Sprintf("%s-%g*(1-%s)", cx, 0.3*bw, E), cy
	}
	return crop, cx, cy
}

// CropSpan is the part of a vw×vh picture (0–1 across) a w:h fill crop centred on ax keeps.
func CropSpan(ax float64, vw, vh, w, h int) (lo, hi float64) {
	cw := 1.0
	if float64(vw)/float64(vh) > float64(w)/float64(h) {
		cw = float64(vh) * float64(w) / float64(h) / float64(vw)
	}
	lo = math.Max(0, math.Min(1-cw, ax-cw/2))
	return lo, lo + cw
}

// CropHolds says whether that crop keeps x inside it, with a tenth of the crop's width spare each side.
func CropHolds(ax, x float64, vw, vh, w, h int) bool {
	lo, hi := CropSpan(ax, vw, vh, w, h)
	m := 0.1 * (hi - lo)
	return x >= lo+m && x <= hi-m
}

// gpuSize is the size a hardware-decoded (HDR) source may be scaled to on the GPU before the rest
// of the chain: the output size × zoom when the shapes match, but with the source's own aspect kept
// — covering the output for a filled (cropped) frame, inside it for a fitted one — so FrameFit's
// crop sees an undistorted picture. Unknown source size: no GPU scale (0, 0).
func gpuSize(fill bool, sw, sh, w, h int, z float64) (int, int) {
	if sw <= 0 || sh <= 0 {
		if fill {
			return 0, 0
		}
		return even(float64(w) * z), even(float64(h) * z)
	}
	fx, fy := float64(w)/float64(sw), float64(h)/float64(sh)
	f := math.Min(fx, fy)
	if fill {
		f = math.Max(fx, fy)
	}
	return even(float64(sw) * f * z), even(float64(sh) * f * z)
}
