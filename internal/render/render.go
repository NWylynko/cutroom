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
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cutroom/cutroom/internal/align"
	"github.com/cutroom/cutroom/internal/deps"
	"github.com/cutroom/cutroom/internal/media"
	"github.com/cutroom/cutroom/internal/timeline"
)

// Version is part of every cache key; bump when output would change.
const Version = 2 // 2: HDR sources tone-mapped to SDR bt709 (L23)

type Source struct {
	Path       string
	AudioStart float64
	HasAudio   bool
	Size       int64
	ModTime    time.Time
	Sync       *align.Sync // set when this source's audio is aligned to another source
}

type Options struct {
	Preview  bool
	Range    [2]float64 // output seconds; zero = all
	Jobs     int
	Target   float64 // chunk length (s)
	CacheDir string
	Loudness float64 // LUFS
	TruePeak float64 // dBTP
	Chain    string
	DryRun   bool
	Progress func(string)
}

type Chunk struct {
	Index    int     `json:"index"`
	Segments []int   `json:"segments"`
	Start    float64 `json:"output_start"`
	Duration float64 `json:"duration"`
	Frames   int64   `json:"frames"`
	Key      string  `json:"key"`
	Cached   bool    `json:"cached"`
	Seconds  float64 `json:"render_seconds,omitempty"`
	file     string
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
		if acc >= o.Target {
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
		c.Key = hash([]any{Version, tl.FPS, w, h, o.Preview, encArgs(o.Preview), segs, tl.Grid})
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
		sdr := media.ReadSDR(s.Path, even(float64(w)*z), even(float64(h)*z))
		args = append(args, "-ss", f6(seek), "-t", f6(dur))
		args = append(args, sdr.Pre...)
		args = append(args, "-i", s.Path)
		chain := fmt.Sprintf("[%d:v]", n) + sdr.Chain(fmt.Sprintf("fps=fps=%s:start_time=%s:round=near,trim=end_frame=%d,setpts=PTS-STARTPTS", fps, f6(lead), nf))
		if z > 1 { // punch-in (L40): crop around the anchor, kept inside the frame, then scale back up
			ax, ay := g.Anchor()
			chain += "," + ZoomCrop(z, ax, ay)
		}
		chain += fmt.Sprintf(",scale=%d:%d:force_original_aspect_ratio=decrease:flags=lanczos,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,setsar=1", w, h, w, h)
		chain += ",format=yuv420p" + fmt.Sprintf("[v%d]", n)
		fc = append(fc, chain)
		labels += fmt.Sprintf("[v%d]", n)
	}
	// restamp after concat so every frame sits exactly on the grid (L08)
	fc = append(fc, fmt.Sprintf("%sconcat=n=%d:v=1:a=0,settb=AVTB,setpts=N/(FRAME_RATE*TB),setparams=colorspace=bt709:color_primaries=bt709:color_trc=bt709:range=tv[out]", labels, len(c.Segments)))
	args = append(args, "-filter_complex", strings.Join(fc, ";"), "-map", "[out]", "-an")
	args = append(args, encArgs(preview)...)
	args = append(args, "-pix_fmt", "yuv420p", "-r", fps, "-fps_mode", "cfr",
		"-colorspace", "bt709", "-color_primaries", "bt709", "-color_trc", "bt709", "-color_range", "tv",
		"-video_track_timescale", "90000", "-movflags", "+faststart", out)
	return args
}

func renderChunk(tl *timeline.Timeline, srcs map[string]Source, c *Chunk, preview bool) error {
	if err := ensureDir(c.file); err != nil {
		return err
	}
	tmp := strings.TrimSuffix(c.file, ".mp4") + ".tmp.mp4"
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
	all := Plan(tl, srcs, o)
	chunks := all
	if o.Range[1] > 0 {
		chunks = nil
		for _, c := range all {
			if c.Start < o.Range[1] && c.Start+c.Duration > o.Range[0] {
				chunks = append(chunks, c)
			}
		}
	}
	w, h := size(tl, o.Preview)
	res := &Result{Output: out, Width: w, Height: h, Preview: o.Preview, DryRun: o.DryRun}
	for _, c := range chunks {
		res.Expected += c.Duration
		res.Frames += c.Frames
	}
	res.Expected = round6(res.Expected)
	if len(chunks) > 0 {
		res.RangeStart = chunks[0].Start
	}
	res.Estimate = estimate(o, chunks, w*h)
	res.Chunks = chunks
	if o.DryRun {
		return res, nil
	}

	// audio (whole timeline measured once; L17)
	audioDone := make(chan error, 1)
	go func() {
		ar, err := renderAudio(tl, srcs, chunks, all, o)
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
			errs[i] = renderChunk(tl, srcs, &chunks[i], o.Preview)
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
	if _, err := deps.Run("ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-f", "concat", "-safe", "0", "-i", list,
		"-i", res.Audio.File, "-map", "0:v", "-map", "1:a", "-c", "copy", "-movflags", "+faststart", tmp); err != nil {
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
	if err := writeManifest(tl, srcs, chunks, res, o); err != nil {
		return nil, err
	}
	return res, nil
}

func renderAudio(tl *timeline.Timeline, srcs map[string]Source, chunks, all []Chunk, o Options) (AudioResult, error) {
	src, ok := srcs[tl.Audio]
	if tl.Audio == "" || !ok || !src.HasAudio {
		return AudioResult{}, fmt.Errorf("the timeline has no audio source")
	}
	cache := filepath.Join(o.CacheDir, "render")
	os.MkdirAll(cache, 0o755)
	full := filepath.Join(cache, "src-"+hash([]any{src.Path, src.Size, src.ModTime.UnixNano()})+".wav")
	if _, err := AudioWav(src.Path, full); err != nil {
		return AudioResult{}, err
	}
	pieces := func(cs []Chunk) []Piece {
		var ps []Piece
		for _, c := range cs {
			for _, k := range c.Segments {
				g := tl.Segments[k]
				// frames × (SR/fps) samples exactly, so audio length always equals video length
				n := int64(math.Round(float64(tl.Frames(g)) * SR / tl.Rate()))
				if g.Source == tl.Audio {
					from := int64(math.Round((g.In - src.AudioStart) * SR))
					ps = append(ps, Piece{from, from + n})
					continue
				}
				// audio from another device: map through its measured offset + drift (L05), in
				// pieces of ≤ 10 s so drift never accumulates beyond half a sample per piece
				const step = 10 * SR
				for o := int64(0); o < n; o += step {
					m := min(step, n-o)
					t := g.In + float64(o)/SR
					from := int64(math.Round((src.Sync.Map(t) - src.AudioStart) * SR))
					ps = append(ps, Piece{from, from + m})
				}
			}
		}
		return ps
	}
	for _, c := range all {
		for _, k := range c.Segments {
			if g := tl.Segments[k]; g.Source != tl.Audio && (src.Sync == nil || src.Sync.To != g.Source) {
				return AudioResult{}, fmt.Errorf("audio source %q is not aligned to %q: run `cutroom sync %s --to %s`", tl.Audio, g.Source, tl.Audio, g.Source)
			}
		}
	}
	key := hash([]any{Version, full, pieces(all), o.Chain, o.Loudness, o.TruePeak, src.Sync})
	allWav := filepath.Join(cache, "timeline-"+key+".wav")
	if _, err := os.Stat(allWav); err != nil {
		if _, err := Assemble(full, allWav, pieces(all), 20); err != nil {
			return AudioResult{}, err
		}
	}
	// gain is measured on the whole timeline even for a range render (L17)
	rkey := hash([]any{key, chunkKeys(chunks)})
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
	if len(chunks) != len(all) {
		in = filepath.Join(cache, "range-"+rkey+".wav")
		if _, err := Assemble(full, in, pieces(chunks), 20); err != nil {
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

func chunkKeys(cs []Chunk) []string {
	var k []string
	for _, c := range cs {
		k = append(k, c.Key)
	}
	return k
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
