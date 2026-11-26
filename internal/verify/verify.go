// Package verify checks a rendered file against the manifest of what it should contain: frame count,
// picture (render frames vs source frames), A/V sync per segment, words (re-transcription vs the
// timeline's words), sound at cut edges (L33), loudness and output tags. It measures the file; it does
// not trust the renderer.
package verify

import (
	"bytes"
	"fmt"
	"github.com/NWylynko/cutroom/internal/align"
	"github.com/NWylynko/cutroom/internal/perceive"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/NWylynko/cutroom/internal/deps"
	"github.com/NWylynko/cutroom/internal/media"
	"github.com/NWylynko/cutroom/internal/render"
	"github.com/NWylynko/cutroom/internal/speech"
)

type Item struct {
	Check   string  `json:"check"`
	Status  string  `json:"status"` // pass, warn, fail
	Detail  string  `json:"detail"`
	At      float64 `json:"output_time,omitempty"`
	Segment string  `json:"segment,omitempty"`
	Fix     string  `json:"fix,omitempty"`
}

type SegSync struct {
	Segment     string  `json:"segment"`
	OutStart    float64 `json:"output_start"`
	AudioLagMS  float64 `json:"audio_lag_ms"` // + = render audio late vs source
	VideoOffset float64 `json:"video_offset_frames"`
	Measured    bool    `json:"measured"`
}

type Report struct {
	File     string    `json:"file"`
	Pass     bool      `json:"pass"`
	Items    []Item    `json:"items"`
	Sync     []SegSync `json:"segments"`
	Heard    string    `json:"heard"`
	Fails    int       `json:"fails"`
	Warnings int       `json:"warnings"`
}

// Add records a result measured outside this package (e.g. a subtitle file next to the render).
func (r *Report) Add(it Item) {
	r.add(it)
	r.Pass = r.Fails == 0
}

func (r *Report) add(it Item) {
	r.Items = append(r.Items, it)
	switch it.Status {
	case "fail":
		r.Fails++
	case "warn":
		r.Warnings++
	}
}

type Options struct {
	Transcripts map[string]*speech.Transcript
	Model       string
	CacheDir    string
	CurrentHash string // hash of the timeline now
}

const (
	gw, gh = 160, 90
	sr     = 16000
)

func Check(file string, m *render.Manifest, o Options) (*Report, error) {
	r := &Report{File: file}
	fps := rate(m.FPS)
	tmp, err := os.MkdirTemp(o.CacheDir, "check")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	// 1. streams, frames, tags
	vframes, w, h, rot, cs := probeVideo(file)
	adur := streamDur(file, "a:0")
	vdur := float64(vframes) / fps
	st := "pass"
	if vframes != m.Frames {
		st = "fail"
	}
	r.add(Item{Check: "frames", Status: st, Detail: fmt.Sprintf("%d frames (%.3f s); the timeline rendered had %d", vframes, vdur, m.Frames),
		Fix: ifs(st == "fail", "re-render: `cutroom render`")})
	st = "pass"
	if math.Abs(adur-vdur) > 1.5/fps {
		st = "fail"
	}
	r.add(Item{Check: "av-length", Status: st, Detail: fmt.Sprintf("audio %.3f s, video %.3f s", adur, vdur)})
	st = "pass"
	if w != m.Width || h != m.Height || rot != 0 {
		st = "fail"
	}
	rotNote := "no rotation tag: upright in any player, including ones that ignore rotation"
	if rot != 0 {
		rotNote = fmt.Sprintf("a %d° rotation tag (players that ignore it show it sideways)", rot)
	}
	r.add(Item{Check: "format", Status: st, Detail: fmt.Sprintf("%dx%d, display rotation %d° (%s), colour %s", w, h, rot, rotNote, orDash(cs))})
	if o.CurrentHash != "" && o.CurrentHash != m.TimelineHash {
		r.add(Item{Check: "stale", Status: "warn", Detail: "the timeline has changed since this file was rendered", Fix: "`cutroom render` again, then check"})
	}

	// 2. picture: sampled render frames vs source frames (±3)
	pictureCheck(file, m, fps, r)
	overlayCheck(file, m, fps, r)
	safeZoneCheck(m, r)
	framingCheck(m, r)
	toneCheck(file, m, fps, r)

	// 3. sync per segment + 4. words + 5. edges — all need audio
	renderWav := filepath.Join(tmp, "render.wav")
	if _, err := deps.Run("ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-i", file, "-vn", "-ac", "1", "-ar", "16000", "-c:a", "pcm_s16le", renderWav); err != nil {
		return nil, err
	}
	rx, _ := speech.ReadWav(renderWav)
	// WAV extraction starts at the first audio sample: put it back on the file's clock (a late
	// audio stream must show up as late, not be silently re-aligned)
	if ast := streamStart(file, "a:0") - streamStart(file, "v:0"); ast > 0 {
		pad := make([]float64, int(math.Round(ast*sr)))
		rx = append(pad, rx...)
	} else if ast < 0 {
		n := min(len(rx), int(math.Round(-ast*sr)))
		rx = rx[n:]
	}
	// each segment's sound came from its own audio source (multi-take); load each once
	tracks := map[string][]float64{}
	load := func(id string) []float64 {
		if x, ok := tracks[id]; ok {
			return x
		}
		var x []float64
		if s, ok := m.Sources[id]; ok {
			w := speech.WavPath(o.CacheDir, id)
			if _, err := os.Stat(w); err != nil {
				w = filepath.Join(tmp, "src-"+id+".wav")
				if err := speech.ExtractWav(s.Path, w); err != nil {
					w = ""
				}
			}
			if w != "" {
				x, _ = speech.ReadWav(w)
			}
		}
		tracks[id] = x
		return x
	}
	audioTrack := func(g render.MSeg) ([]float64, float64, func(float64) float64, bool) {
		a := m.AudioOf(g)
		s := m.Sources[a]
		x := load(a)
		if len(x) == 0 {
			return nil, 0, nil, false
		}
		if g.Source != a && s.Sync != nil {
			return x, s.AudioStart, s.Sync.Map, true
		}
		return x, s.AudioStart, func(t float64) float64 { return t }, true
	}
	syncCheck(rx, audioTrack, m, r)
	cameraCheck(rx, m, load, r)
	if o.Model != "" {
		wordCheck(renderWav, tmp, m, o, r)
	}
	for id := range audioSources(m) {
		if tr := o.Transcripts[id]; tr != nil && len(load(id)) > 0 {
			edgeCheck(load(id), m.Sources[id].AudioStart, tr, id, m, r)
			cutThroughWords(tr, id, m, r)
		}
	}

	// 6. loudness
	if lufs, peak, err := render.Loudness(file, ""); err == nil {
		st := "pass"
		if math.Abs(lufs-m.Loudness) > 1 || peak > m.TruePeak+0.5 {
			st = "warn"
		}
		r.add(Item{Check: "loudness", Status: st, Detail: fmt.Sprintf("%.1f LUFS (target %.0f), peak %.1f dBTP (ceiling %.1f)", lufs, m.Loudness, peak, m.TruePeak)})
	}
	r.Pass = r.Fails == 0
	return r, nil
}

// --- picture ---------------------------------------------------------------------------------

// pictureNoise is the grey-frame difference an encoded frame has from its own source frame (scaling
// and encoding): measured 3.3–3.9 on 4K HEVC → 1080p; a wrong moment with any motion is ≥ 10.
const pictureNoise = 6.0

func pictureCheck(file string, m *render.Manifest, fps float64, r *Report) {
	frames := grey(file, "", 0, 0)
	if len(frames) == 0 {
		r.add(Item{Check: "picture", Status: "fail", Detail: "could not decode the render"})
		return
	}
	type sample struct {
		seg     render.MSeg
		f       int     // render frame
		srcT    float64 // source time
		offset  int
		err     float64
		ambig   bool
		skipped bool
		distMS  float64
	}
	var ss []sample
	for _, g := range m.Segments {
		start := int(math.Round(g.OutStart * fps))
		picks := []int{1, int(g.Frames) / 2, int(g.Frames) - 2}
		if g.Frames < 4 {
			picks = []int{int(g.Frames) / 2}
		}
		for _, k := range picks {
			ss = append(ss, sample{seg: g, f: start + k, srcT: g.In + float64(k)/fps})
		}
	}
	if len(ss) > 120 { // keep it fast on long timelines: spread samples evenly
		step := float64(len(ss)) / 120
		var keep []sample
		for i := 0.0; int(i) < len(ss); i += step {
			keep = append(keep, ss[int(i)])
		}
		ss = keep
	}
	var wg sync.WaitGroup
	sem := make(chan struct{}, 6)
	for i := range ss {
		if ss[i].skipped || ss[i].f >= len(frames) {
			ss[i].skipped = true
			continue
		}
		wg.Add(1)
		go func(s *sample) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			path := m.Sources[s.seg.Source].Path
			crop := m.SegVF(s.seg) // the same punch-in (L40) and frame crop the render used
			burst, pts := greyBurst(path, crop, s.srcT-4/fps, 9)
			if len(burst) == 0 {
				s.skipped = true
				return
			}
			errs := make([]float64, len(burst))
			best := math.Inf(1)
			boxes := overlayBoxes(m, s.f, fps) // overlays aren't in the source: compare around them
			for j, b := range burst {
				errs[j] = mseMasked(frames[s.f], b, boxes, false)
				best = math.Min(best, errs[j])
			}
			// the frame that should be on screen: the one whose display interval [pts_j, pts_j+1)
			// contains the expected time (real timestamps, so VFR held frames are handled)
			exp, bj := -1, 0
			for j := range burst {
				hi := pts[j] + 1/fps
				if j+1 < len(pts) {
					hi = pts[j+1]
				}
				if s.srcT >= pts[j]-0.5/fps && s.srcT < hi-0.5/fps+1e-6 {
					exp = j
					break
				}
			}
			for j := range errs {
				if errs[j] < errs[bj] {
					bj = j
				}
			}
			s.err = best
			bestDist := math.Abs(pts[bj]-s.srcT) * 1000
			// the expected frame (or one a viewer can't tell from it) is shown: within 1 of the best,
			// or at the encode-noise level itself — at 50–60 fps a near-still speaker's neighbouring
			// frames differ by less than the encode noise (r63: 3.5 vs 1.9 "two frames off")
			if exp >= 0 && (errs[exp]-best < 1.0 || errs[exp] <= pictureNoise) {
				s.offset, bestDist = 0, 0
			} else {
				s.offset = int(math.Round((pts[bj] - s.srcT) * fps))
			}
			s.ambig = exp >= 0 && bj != exp
			s.distMS = bestDist
		}(&ss[i])
	}
	wg.Wait()
	bad, n, worst := 0, 0, 0.0
	var first *sample
	for i := range ss {
		s := &ss[i]
		if s.skipped {
			continue
		}
		n++
		worst = math.Max(worst, s.err)
		if s.err > 60 || abs(s.offset) >= 2 {
			bad++
			if first == nil {
				first = s
			}
		}
	}
	st := "pass"
	d := fmt.Sprintf("%d sampled frames match their source frames as a player shows them — upright, right moment (worst difference %.1f)", n, worst)
	it := Item{Check: "picture"}
	if n == 0 && len(ss) > 0 {
		zoomed := true
		for _, s := range ss {
			if s.seg.Zoom <= 1.001 {
				zoomed = false
			}
		}
		if !zoomed {
			r.add(Item{Check: "picture", Status: "fail", Detail: "no frame could be measured (source frames could not be decoded); the picture is unverified",
				Fix: "make sure the sources are still readable; if they are, this is a cutroom bug"})
			return
		}
	}
	if bad > 0 {
		st = "fail"
		d = fmt.Sprintf("%d of %d sampled frames show the wrong moment; first at output %.3f (segment %s): about %+d frames (%.0f ms) off, difference %.1f",
			bad, n, float64(first.f)/fps, first.seg.ID, first.offset, first.distMS, first.err)
		if first.err > 500 { // nothing nearby matches: not a timing problem
			d = fmt.Sprintf("%d of %d sampled frames match no nearby source frame; first at output %.3f (segment %s), difference %.1f: wrong picture, framing or zoom, not timing",
				bad, n, float64(first.f)/fps, first.seg.ID, first.err)
		}
		it.At, it.Segment = float64(first.f)/fps, first.seg.ID
		it.Fix = "look at it: `cutroom look " + strconv.FormatFloat(first.srcT, 'f', 2, 64) + " --source " + first.seg.Source + "` vs the render frame"
	}
	it.Status, it.Detail = st, d
	r.add(it)
}

var ptsRe = regexp.MustCompile(`pts_time:(-?[0-9.]+)`)

// greyBurst decodes n frames from seek with their real presentation times (source clock, -copyts),
// so held/dropped frames on VFR sources are judged by time, not by frame count.
func greyBurst(p, crop string, seek float64, n int) ([][]byte, []float64) {
	sdr := media.ReadSDR(p, 0, 0) // compare with the source as the render read it (L23)
	args := []string{"-hide_banner", "-loglevel", "info", "-ss", strconv.FormatFloat(math.Max(0, seek), 'f', 4, 64), "-copyts"}
	args = append(args, sdr.Pre...)
	vf := fmt.Sprintf("scale=%d:%d,format=yuv420p,extractplanes=y,showinfo", gw, gh)
	if crop != "" {
		vf = crop + "," + vf
	}
	args = append(args, "-i", p, "-vf", sdr.Chain(vf),
		"-frames:v", strconv.Itoa(n), "-fps_mode", "passthrough", "-f", "rawvideo", "-")
	cmd := exec.Command("ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	if err != nil {
		return nil, nil
	}
	var fr [][]byte
	for i := 0; (i+1)*gw*gh <= len(b); i++ {
		fr = append(fr, b[i*gw*gh:(i+1)*gw*gh])
	}
	var pts []float64
	for _, m := range ptsRe.FindAllStringSubmatch(stderr.String(), -1) {
		f, _ := strconv.ParseFloat(m[1], 64)
		pts = append(pts, f)
	}
	// showinfo logs frames that -frames:v then discards; frames come out in order, so trim
	if len(pts) < len(fr) {
		return nil, nil
	}
	return fr, pts[:len(fr)]
}

func grey(p, extra string, seek float64, n int) [][]byte {
	args := []string{"-v", "error"}
	if seek > 0 {
		args = append(args, "-ss", strconv.FormatFloat(seek, 'f', 4, 64))
	}
	vf := fmt.Sprintf("scale=%d:%d,format=yuv420p,extractplanes=y", gw, gh)
	if extra != "" {
		vf = extra + "," + vf
	}
	args = append(args, "-i", p, "-vf", vf)
	if n > 0 {
		args = append(args, "-frames:v", strconv.Itoa(n))
	}
	args = append(args, "-fps_mode", "passthrough", "-f", "rawvideo", "-")
	b, err := deps.Run("ffmpeg", args...)
	if err != nil {
		return nil
	}
	var fr [][]byte
	for i := 0; (i+1)*gw*gh <= len(b); i++ {
		fr = append(fr, b[i*gw*gh:(i+1)*gw*gh])
	}
	return fr
}

func mse(a, b []byte) float64 {
	s := 0.0
	for i := range a {
		d := float64(a[i]) - float64(b[i])
		s += d * d
	}
	return s / float64(len(a))
}

// --- sync ------------------------------------------------------------------------------------

// measureSync cross-correlates the render's sound with a source's sound at the moment each segment's
// picture shows; minC is the correlation a window needs to count.
// trackFor gives, for a segment, the sound to compare with: samples, their start on the source clock,
// and the map from the segment's picture time to that clock. ok = false skips the segment.
type trackFor func(g render.MSeg) (sx []float64, start float64, mapT func(float64) float64, ok bool)

func measureSync(rx []float64, get trackFor, m *render.Manifest, minC float64) (all, measured []SegSync, worst float64, worstSeg SegSync) {
	win, maxLag := int(0.5*sr), int(0.15*sr)
	for _, g := range m.Segments {
		d := float64(g.Frames) / rate(m.FPS)
		ss := SegSync{Segment: g.ID, OutStart: g.OutStart}
		sx, audioStart, mapT, ok := get(g)
		if !ok {
			continue
		}
		// try a few windows inside the segment; keep the most confident speech window
		bestC := 0.0
		for _, frac := range []float64{0.5, 0.3, 0.7} {
			if d < 0.6 {
				break
			}
			to := g.OutStart + d*frac
			so := mapT(g.In+d*frac) - audioStart
			oi, si := int(to*sr)-win/2, int(so*sr)-win/2
			if oi < 0 || oi+win > len(rx) || si-maxLag < 0 || si+win+maxLag > len(sx) {
				continue
			}
			seg := rx[oi : oi+win]
			if rms(seg) < 0.01 {
				continue
			}
			lag, c := xcorr(seg, sx[si-maxLag:si+win+maxLag], maxLag)
			if c > bestC && c > minC {
				bestC = c
				ss.AudioLagMS = round1(float64(lag) / sr * 1000)
				ss.Measured = true
			}
		}
		if ss.Measured {
			measured = append(measured, ss)
			if math.Abs(ss.AudioLagMS) > math.Abs(worst) {
				worst, worstSeg = ss.AudioLagMS, ss
			}
		}
		all = append(all, ss)
	}
	return all, measured, worst, worstSeg
}

func syncCheck(rx []float64, get trackFor, m *render.Manifest, r *Report) {
	all, measured, worst, worstSeg := measureSync(rx, get, m, 0.5)
	r.Sync = append(r.Sync, all...)
	if len(measured) == 0 {
		r.add(Item{Check: "sync", Status: "warn", Detail: "no segment had enough speech to measure sync"})
		return
	}
	lags := make([]float64, len(measured))
	for i, s := range measured {
		lags[i] = s.AudioLagMS
	}
	sort.Float64s(lags)
	st := "pass"
	it := Item{Check: "sync"}
	if math.Abs(worst) > 45 {
		st = "fail"
		it.At, it.Segment = worstSeg.OutStart, worstSeg.Segment
		it.Fix = "audio and picture disagree; if the audio comes from another device, align it first (see `cutroom guide`)"
	}
	it.Status = st
	ref := m.Audio + "'s sound"
	if s, ok := m.Sources[m.Audio]; ok && s.Sync != nil {
		ref = m.Audio + "'s sound, aligned to " + s.Sync.To + " by sync,"
	}
	if as := audioSources(m); len(as) > 1 {
		var ids []string
		for id := range as {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		ref = "each segment's own sound source (" + strings.Join(ids, ", ") + ")"
	}
	it.Detail = fmt.Sprintf("%d of %d segments measured: the render's sound vs %s at the source moment its picture shows: median %+.0f ms, worst %+.0f ms (limit 45)", len(measured), len(m.Segments), ref, lags[len(lags)/2], worst)
	if st == "pass" {
		off := ""
		if src, ok := m.Sources[m.Audio]; ok && math.Abs(src.AudioStart) >= 0.005 && src.Sync == nil {
			off = fmt.Sprintf(" — %s's audio starts at %+.3f s, and that offset is in the render", m.Audio, src.AudioStart)
		}
		it.Detail += "; with picture ok, lip sync is as the source plays (its audio start offset included" + off + ")"
	}
	r.add(it)
}

func xcorr(seg, ref []float64, maxLag int) (int, float64) {
	es := 0.0
	for _, v := range seg {
		es += v * v
	}
	score := func(lag int) float64 {
		off := maxLag - lag
		s, er := 0.0, 0.0
		for i, v := range seg {
			x := ref[off+i]
			s += v * x
			er += x * x
		}
		return s / math.Sqrt(es*er+1e-12)
	}
	best, bl := -2.0, 0
	for lag := -maxLag; lag <= maxLag; lag += 8 {
		if c := score(lag); c > best {
			best, bl = c, lag
		}
	}
	for lag := bl - 8; lag <= bl+8; lag++ {
		if lag >= -maxLag && lag <= maxLag {
			if c := score(lag); c > best {
				best, bl = c, lag
			}
		}
	}
	// positive lag = render content appears later than the source at the mapped time
	return bl, best
}

func rms(x []float64) float64 {
	s := 0.0
	for _, v := range x {
		s += v * v
	}
	return math.Sqrt(s / float64(len(x)))
}

// --- words -----------------------------------------------------------------------------------

func wordCheck(renderWav, tmp string, m *render.Manifest, o Options, r *Report) {
	raw, err := speech.Whisper(o.Model, renderWav, "", tmp)
	if err != nil {
		r.add(Item{Check: "words", Status: "warn", Detail: "could not re-transcribe the render: " + err.Error()})
		return
	}
	var heard []string
	var heardT []float64
	for _, w := range raw {
		if n := speech.Norm(w.Text); n != "" && !strings.HasPrefix(w.Text, "[") {
			heard = append(heard, n)
			heardT = append(heardT, w.Start)
		}
	}
	var hs []string
	for _, w := range raw {
		hs = append(hs, w.Text)
	}
	r.Heard = strings.Join(hs, " ")
	type ew struct {
		w    string
		outT float64
	}
	var exp []ew
	for _, g := range m.Segments {
		tr := o.Transcripts[g.Source]
		if tr == nil {
			continue
		}
		for _, w := range tr.Words {
			mid := (w.Start + w.End) / 2
			if mid < g.In || mid >= g.Out {
				continue
			}
			// a --range render's own edges aren't cuts: a word they cross isn't expected whole
			if (g.RangeIn && w.Start < g.In) || (g.RangeOut && w.End > g.Out) {
				continue
			}
			if n := speech.Norm(w.Text); n != "" {
				exp = append(exp, ew{n, g.OutStart + math.Max(w.Start, g.In) - g.In})
			}
		}
	}
	if len(exp) == 0 {
		r.add(Item{Check: "words", Status: "warn", Detail: "no transcript for the audio source; run `cutroom transcribe` to enable the word check", Fix: "cutroom transcribe"})
		return
	}
	// LCS diff
	n, k := len(exp), len(heard)
	dp := make([][]int32, n+1)
	for i := range dp {
		dp[i] = make([]int32, k+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := k - 1; j >= 0; j-- {
			if exp[i].w == heard[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else {
				dp[i][j] = max(dp[i+1][j], dp[i][j+1])
			}
		}
	}
	type gap struct {
		missing, extra []string
		at             float64
	}
	var gaps []gap
	var cur gap
	flush := func() {
		if len(cur.missing)+len(cur.extra) > 0 {
			gaps = append(gaps, cur)
		}
		cur = gap{}
	}
	i, j := 0, 0
	for i < n || j < k {
		switch {
		case i < n && j < k && exp[i].w == heard[j]:
			flush()
			i, j = i+1, j+1
		case j >= k || (i < n && dp[i+1][j] >= dp[i][j+1]):
			if len(cur.missing)+len(cur.extra) == 0 {
				cur.at = exp[i].outT
			}
			cur.missing = append(cur.missing, exp[i].w)
			i++
		default:
			if len(cur.missing)+len(cur.extra) == 0 {
				cur.at = heardT[j]
			}
			cur.extra = append(cur.extra, heard[j])
			j++
		}
	}
	flush()
	matched := int(dp[0][0])
	// say which model listened, and whether it is the one the transcript used (a shared mishearing
	// would pass unnoticed)
	model := speech.ModelName(o.Model)
	same := false
	for _, g := range m.Segments {
		if tr := o.Transcripts[g.Source]; tr != nil && tr.Model == model {
			same = true
		}
	}
	via := fmt.Sprintf(" (re-transcribed with %s)", model)
	if same {
		via = fmt.Sprintf(" (re-transcribed with %s, the model the transcript used: it confirms the cut, not the recogniser — a larger model via `cutroom transcribe --model` checks both)", model)
	}
	// a --range render starts/ends mid-speech: words at its own edges are heard in part or misheard
	edgeNote := ""
	if lo, hi, ok := rangeEdges(m); ok {
		kept := gaps[:0]
		for _, g := range gaps {
			if (lo >= 0 && g.at < lo+0.5) || (hi >= 0 && g.at > hi-0.5) {
				continue
			}
			kept = append(kept, g)
		}
		if d := len(gaps) - len(kept); d > 0 {
			edgeNote = fmt.Sprintf("; %d difference(s) within 0.5 s of the range's own start/end not counted (words cut by the range, not by the edit)", d)
		}
		gaps = kept
	}
	if len(gaps) == 0 {
		d := fmt.Sprintf("all %d expected words heard in order in the render", n)
		if edgeNote != "" {
			d = fmt.Sprintf("%d of %d expected words heard in order%s", matched, n, edgeNote)
		}
		r.add(Item{Check: "words", Status: "pass", Detail: d + via})
		return
	}
	r.add(Item{Check: "words", Status: "pass", Detail: fmt.Sprintf("%d of %d expected words heard in order; %d difference(s) listed below%s%s", matched, n, len(gaps), edgeNote, via)})
	for _, g := range gaps {
		st := "warn"
		var d string
		switch {
		case len(g.extra) == 0:
			d = fmt.Sprintf("expected but not heard: %q", strings.Join(g.missing, " "))
			if len(g.missing) >= 3 {
				st = "fail"
			}
		case len(g.missing) == 0:
			d = fmt.Sprintf("heard but not in the timeline: %q (a fragment of cut material, or a misheard word)", strings.Join(g.extra, " "))
		default:
			d = fmt.Sprintf("heard %q where the transcript has %q (usually the recogniser; listen if unsure)", strings.Join(g.extra, " "), strings.Join(g.missing, " "))
		}
		seg := segAt(m, g.at)
		r.add(Item{Check: "words", Status: st, Detail: d, At: round3(g.at), Segment: seg,
			Fix: fmt.Sprintf("`cutroom listen` the source around it (output %.2f s is in segment %s)", g.at, seg)})
	}
}

// rangeEdges gives the file times where a --range render starts/ends mid-segment (-1: a real edge).
func rangeEdges(m *render.Manifest) (lo, hi float64, ok bool) {
	if !m.Partial || len(m.Segments) == 0 {
		return 0, 0, false
	}
	lo, hi = -1, -1
	if m.Segments[0].RangeIn {
		lo = 0
	}
	if last := m.Segments[len(m.Segments)-1]; last.RangeOut {
		hi = last.OutStart + float64(last.Frames)/rate(m.FPS)
	}
	return lo, hi, lo >= 0 || hi >= 0
}

func segAt(m *render.Manifest, t float64) string {
	for _, g := range m.Segments {
		if t >= g.OutStart && t < g.OutStart+float64(g.Frames)/rate(m.FPS) {
			return g.ID
		}
	}
	return ""
}

// --- edges (L33) -----------------------------------------------------------------------------

func edgeCheck(sx []float64, audioStart float64, tr *speech.Transcript, id string, m *render.Manifest, r *Report) {
	env := speech.NewEnvelope(sx)
	env.Floor, env.Peak = tr.RoomTone, tr.SpeechDB
	env.Thr = env.Floor + math.Max(10, 0.35*(env.Peak-env.Floor))
	regions := env.Regions()
	found, breaths := 0, 0
	for _, g := range m.Segments {
		if g.Source != id || m.AudioOf(g) != id {
			continue
		}
		for _, rg := range regions {
			a, b := rg.Start+audioStart, rg.End+audioStart
			ov := math.Min(b, g.Out) - math.Max(a, g.In)
			if ov < 0.06 {
				continue
			}
			worded := false
			for _, w := range tr.Words {
				if w.Start < math.Min(b, g.Out) && w.End > math.Max(a, g.In) {
					worded = true
					break
				}
			}
			if worded {
				continue
			}
			if speech.IsBreath(rg.PeakDB, tr.SpeechDB, rg.End-rg.Start) {
				breaths++
				continue
			}
			found++
			at := g.OutStart + math.Max(a, g.In) - g.In
			r.add(Item{Check: "stray-sound", Status: "warn", At: round3(at), Segment: g.ID,
				Detail: fmt.Sprintf("%.2f s of sound with no words kept in segment %s (source %.2f–%.2f, peak %.0f dB): a fragment of cut speech, a breath, or a noise", ov, g.ID, math.Max(a, g.In), math.Min(b, g.Out), rg.PeakDB),
				Fix:    fmt.Sprintf("`cutroom listen %.2f-%.2f`; cut it with `cutroom edit cut %.2f-%.2f` if unwanted", math.Max(a, g.In)-0.5, math.Min(b, g.Out)+0.5, math.Max(a, g.In), math.Min(b, g.Out))})
		}
	}
	if found == 0 && !r.has("stray-sound") {
		d := "no sound without words kept at any cut or inside segments"
		if breaths > 0 {
			d += fmt.Sprintf(" (%d breath(s) kept: quiet, normal)", breaths)
		}
		r.add(Item{Check: "stray-sound", Status: "pass", Detail: d})
	}
}

// cutThroughWords reports segment edges that fall inside a word (a clipped word in the render).
func cutThroughWords(tr *speech.Transcript, id string, m *render.Manifest, r *Report) {
	fps := rate(m.FPS)
	for _, g := range m.Segments {
		if g.Source != id || m.AudioOf(g) != id {
			continue
		}
		for _, w := range tr.Words {
			for _, e := range []struct {
				t    float64
				side string
				out  float64
				rng  bool
			}{{g.In, "start", g.OutStart, g.RangeIn}, {g.Out, "end", g.OutStart + float64(g.Frames)/fps, g.RangeOut}} {
				if !e.rng && e.t > w.Start+0.04 && e.t < w.End-0.04 {
					r.add(Item{Check: "clipped-word", Status: "warn", At: round3(e.out), Segment: g.ID,
						Detail: fmt.Sprintf("segment %s %ss inside the word %q (source %.2f–%.2f): it is cut off in the render", g.ID, e.side, w.Text, w.Start, w.End),
						Fix:    fmt.Sprintf("cut by words instead (`cutroom edit cut --text ...`), or `cutroom listen %.2f-%.2f`", w.Start-0.5, w.End+0.5)})
				}
			}
		}
	}
}

// --- probes ----------------------------------------------------------------------------------

func probeVideo(p string) (frames int64, w, h, rot int, cs string) {
	b, _ := deps.Run("ffprobe", "-v", "error", "-select_streams", "v:0", "-count_packets",
		"-show_entries", "stream=nb_read_packets,width,height,color_space:stream_side_data=rotation", "-of", "default=nw=1", p)
	for _, l := range strings.Split(string(b), "\n") {
		k, v, _ := strings.Cut(strings.TrimSpace(l), "=")
		switch k {
		case "nb_read_packets":
			frames, _ = strconv.ParseInt(v, 10, 64)
		case "width":
			w, _ = strconv.Atoi(v)
		case "height":
			h, _ = strconv.Atoi(v)
		case "rotation":
			rot, _ = strconv.Atoi(v)
		case "color_space":
			cs = v
		}
	}
	return
}

func streamStart(p, s string) float64 {
	b, _ := deps.Run("ffprobe", "-v", "error", "-select_streams", s, "-show_entries", "stream=start_time", "-of", "csv=p=0", p)
	f, _ := strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
	return f
}

func streamDur(p, s string) float64 {
	b, _ := deps.Run("ffprobe", "-v", "error", "-select_streams", s, "-show_entries", "stream=duration", "-of", "csv=p=0", p)
	f, _ := strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
	return f
}

func rate(s string) float64 {
	if n, d, ok := strings.Cut(s, "/"); ok {
		a, _ := strconv.ParseFloat(n, 64)
		b, _ := strconv.ParseFloat(d, 64)
		return a / b
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func ifs(c bool, s string) string {
	if c {
		return s
	}
	return ""
}
func orDash(s string) string {
	if s == "" {
		return "untagged"
	}
	return s
}
func round1(f float64) float64 { return math.Round(f*10) / 10 }
func round3(f float64) float64 { return math.Round(f*1000) / 1000 }

// cameraCheck verifies a separate recorder's alignment independently of `sync`: the render's sound
// (from the recorder) against the camera's own sound at the same moment. Both are band-passed so
// two different microphones correlate. A wrong sync offset passes syncCheck (it uses the same
// offset) but fails here.
func cameraCheck(rx []float64, m *render.Manifest, load func(string) []float64, r *Report) {
	// segments whose sound came from a recorder, compared with their camera's own sound
	cams := map[string]bool{}
	bp := map[string][]float64{}
	get := func(g render.MSeg) ([]float64, float64, func(float64) float64, bool) {
		if m.AudioOf(g) == g.Source {
			return nil, 0, nil, false
		}
		cx, ok := bp[g.Source]
		if !ok {
			if x := load(g.Source); len(x) > 0 {
				cx = align.Bandpass(x)
			}
			bp[g.Source] = cx
		}
		if len(cx) == 0 {
			return nil, 0, nil, false // camera has no audio: nothing independent to compare with
		}
		cams[g.Source] = true
		return cx, m.Sources[g.Source].AudioStart, func(t float64) float64 { return t }, true
	}
	_, measured, worst, worstSeg := measureSync(align.Bandpass(rx), get, m, 0.3)
	if len(cams) == 0 {
		return
	}
	var camList []string
	for c := range cams {
		camList = append(camList, c)
	}
	sort.Strings(camList)
	cam := strings.Join(camList, ", ")
	if len(measured) == 0 {
		r.add(Item{Check: "sync-camera", Status: "warn", Detail: "could not compare with " + cam + "'s own sound (too little shared speech)"})
		return
	}
	lags := make([]float64, len(measured))
	for i, s := range measured {
		lags[i] = s.AudioLagMS
	}
	sort.Float64s(lags)
	it := Item{Check: "sync-camera", Status: "pass"}
	if math.Abs(worst) > 45 {
		it.Status, it.At, it.Segment = "fail", worstSeg.OutStart, worstSeg.Segment
		it.Fix = fmt.Sprintf("the recorder's alignment looks wrong: re-run `cutroom sync %s`, then render again", m.AudioOf(segByID(m, worstSeg.Segment)))
	}
	it.Detail = fmt.Sprintf("%d of %d segments: the render's sound vs %s's own camera sound (independent of sync's offset): median %+.0f ms, worst %+.0f ms (limit 45)", len(measured), len(m.Segments), cam, lags[len(lags)/2], worst)
	r.add(it)
}

// toneCheck (L23): for segments from an HDR source, the render's colourfulness and brightness range
// must match the source read through the tone map. A plain conversion (or a render made before tone
// mapping existed) keeps the right tags but looks washed out; this catches it.
func toneCheck(file string, m *render.Manifest, fps float64, r *Report) {
	var outS, srcS, outR, srcR float64
	n := 0
	hdr := ""
	for _, g := range m.Segments {
		src, ok := m.Sources[g.Source]
		if !ok {
			continue
		}
		sdr := media.ReadSDR(src.Path, 0, 0)
		if sdr.HDR == "" {
			continue
		}
		hdr = strings.ToUpper(sdr.HDR)
		if sdr.Method == "none" || n >= 5 {
			continue
		}
		d := float64(g.Frames) / fps
		crop := ""
		if vf := m.SegVF(g); vf != "" {
			crop = vf + ","
		}
		os_, or_, ok1 := frameTone(file, g.OutStart+d/2, media.SDR{}, "")
		ss, sr, ok2 := frameTone(src.Path, g.In+d/2, sdr, crop)
		if ok1 && ok2 {
			outS, outR, srcS, srcR, n = outS+os_, outR+or_, srcS+ss, srcR+sr, n+1
		}
	}
	if hdr == "" {
		return
	}
	if n == 0 {
		r.add(Item{Check: "tone", Status: "warn", Detail: "HDR (" + hdr + ") source, but its colour could not be compared (no tone mapper, or frames unreadable)",
			Fix: "`cutroom doctor`; look at a frame with `cutroom look --output <time>`"})
		return
	}
	ds, dr := outS/srcS-1, outR/srcR-1
	it := Item{Check: "tone", Status: "pass", Detail: fmt.Sprintf("HDR (%s) source tone-mapped to SDR: saturation %.0f vs source %.0f, brightness range %.0f vs %.0f (%d frame(s), within 15%%)",
		hdr, outS/float64(n), srcS/float64(n), outR/float64(n), srcR/float64(n), n)}
	if math.Abs(ds) > 0.15 || math.Abs(dr) > 0.15 {
		it.Status = "fail"
		it.Fix = "the render's colour doesn't match the tone-mapped source (washed out or over-saturated): render again with this cutroom, then `cutroom look --output <time>`"
	}
	r.add(it)
}

var satRe, yLowRe, yHighRe = regexp.MustCompile(`SATAVG=([0-9.]+)`), regexp.MustCompile(`YLOW=([0-9.]+)`), regexp.MustCompile(`YHIGH=([0-9.]+)`)

// frameTone returns one frame's mean saturation and YHIGH−YLOW (8-bit units), read through sdr.
func frameTone(p string, t float64, sdr media.SDR, crop string) (sat, rng float64, ok bool) {
	args := []string{"-v", "error", "-ss", strconv.FormatFloat(math.Max(0, t), 'f', 3, 64)}
	args = append(args, sdr.Pre...)
	args = append(args, "-i", p, "-frames:v", "1", "-vf", sdr.Chain(crop+"format=yuv420p,signalstats,metadata=print:file=-"), "-f", "null", "-")
	b, err := deps.Run("ffmpeg", args...)
	if err != nil {
		return 0, 0, false
	}
	get := func(re *regexp.Regexp) (float64, bool) {
		m := re.FindStringSubmatch(string(b))
		if m == nil {
			return 0, false
		}
		v, _ := strconv.ParseFloat(m[1], 64)
		return v, true
	}
	s, ok1 := get(satRe)
	lo, ok2 := get(yLowRe)
	hi, ok3 := get(yHighRe)
	return s, hi - lo, ok1 && ok2 && ok3 && hi > lo
}

// audioSources lists the audio sources the render's segments used.
func audioSources(m *render.Manifest) map[string]bool {
	out := map[string]bool{}
	for _, g := range m.Segments {
		out[m.AudioOf(g)] = true
	}
	return out
}

func segByID(m *render.Manifest, id string) render.MSeg {
	for _, g := range m.Segments {
		if g.ID == id {
			return g
		}
	}
	return render.MSeg{}
}

func (r *Report) has(check string) bool {
	for _, it := range r.Items {
		if it.Check == check {
			return true
		}
	}
	return false
}

// --- overlays (L16, L20) -----------------------------------------------------------------------

// overlayBoxes: the areas (in gw×gh pixels, padded) covered by overlays at render frame f.
func overlayBoxes(m *render.Manifest, f int, fps float64) [][4]int {
	var bs [][4]int
	for _, p := range m.Overlays {
		if f >= int(math.Round(p.Start*fps)) && f < int(math.Round(p.End*fps)) {
			bs = append(bs, boxPx(p))
		}
	}
	return bs
}

func boxPx(p render.Placement) [4]int {
	return [4]int{int(math.Floor(p.Box[0]*gw)) - 1, int(math.Floor(p.Box[1]*gh)) - 1, int(math.Ceil(p.Box[2]*gw)) + 1, int(math.Ceil(p.Box[3]*gh)) + 1}
}

func inBox(x, y int, b [4]int) bool { return x >= b[0] && x < b[2] && y >= b[1] && y < b[3] }

// mseMasked compares grey frames outside the boxes (inside = false) or only inside them (true).
func mseMasked(a, b []byte, boxes [][4]int, inside bool) float64 {
	if len(boxes) == 0 && !inside {
		return mse(a, b)
	}
	s, n := 0.0, 0
	for y := 0; y < gh; y++ {
		for x := 0; x < gw; x++ {
			in := false
			for _, bx := range boxes {
				in = in || inBox(x, y, bx)
			}
			if in != inside {
				continue
			}
			d := float64(a[y*gw+x]) - float64(b[y*gw+x])
			s += d * d
			n++
		}
	}
	if n == 0 {
		return 0
	}
	return s / float64(n)
}

// overlayCheck: each overlay is visible inside its window and gone on the frame before it starts
// and the frame after it ends — a frame-exact check against the source picture under the card.
func overlayCheck(file string, m *render.Manifest, fps float64, r *Report) {
	if len(m.Overlays) == 0 {
		return
	}
	frames := grey(file, "", 0, 0)
	// diffIn: the box's difference from the source frame shown there (found by matching around it)
	diffIn := func(f int, box [4]int) (float64, bool) {
		if f < 0 || f >= len(frames) {
			return 0, false
		}
		var seg *render.MSeg
		for i := range m.Segments {
			g := &m.Segments[i]
			a := int(math.Round(g.OutStart * fps))
			if f >= a && f < a+int(g.Frames) {
				seg = g
			}
		}
		if seg == nil {
			return 0, false
		}
		srcT := seg.In + (float64(f)-math.Round(seg.OutStart*fps))/fps
		burst, _ := greyBurst(m.Sources[seg.Source].Path, m.SegVF(*seg), srcT-2/fps, 5)
		if len(burst) == 0 {
			return 0, false
		}
		// find the source frame shown there by the picture around the box; a box over most of the
		// frame (a full-frame clip) leaves too little around it, so match on the whole frame
		mask := [][4]int{box}
		if (box[2]-box[0])*(box[3]-box[1]) > gw*gh/2 {
			mask = nil
		}
		bj, be := 0, math.Inf(1)
		for j, b := range burst {
			if e := mseMasked(frames[f], b, mask, false); e < be {
				bj, be = j, e
			}
		}
		return mseMasked(frames[f], burst[bj], [][4]int{box}, true), true
	}
	for _, p := range m.Overlays {
		if p.Kind == "captions" {
			captionCheck(p, fps, diffIn, r)
			continue
		}
		box := boxPx(p)
		fa, fb := int(math.Round(p.Start*fps)), int(math.Round(p.End*fps))
		nf := int(math.Round(p.Fade * fps))
		var bad []string
		in := []int{fa + nf, (fa + fb) / 2, fb - 1 - nf}
		if nf == 0 {
			in[0], in[2] = fa, fb-1
		}
		minIn := math.Inf(1)
		for _, f := range in {
			if d, ok := diffIn(f, box); ok {
				minIn = math.Min(minIn, d)
				if d < 150 {
					bad = append(bad, fmt.Sprintf("not visible at %.3f (frame %d, difference %.0f)", float64(f)/fps, f, d))
				}
			}
		}
		edge := ""
		for _, f := range []int{fa - 1, fb} {
			d, ok := diffIn(f, box)
			if !ok {
				continue
			}
			if d > 60 {
				bad = append(bad, fmt.Sprintf("still on screen at %.3f (frame %d, outside its window, difference %.0f)", float64(f)/fps, f, d))
			} else {
				edge = "; gone on the frame before and after"
			}
		}
		clipNote := ""
		if p.Kind == "clip" {
			note, problem := clipFrames(file, m, p, fps)
			if problem != "" {
				bad = append(bad, problem)
			}
			clipNote = note
		}
		it := Item{Check: "overlay", At: round3(p.Start)}
		if len(bad) > 0 {
			it.Status = "fail"
			it.Detail = fmt.Sprintf("%s %q (%.3f–%.3f): %s", p.ID, p.Name, p.Start, p.End, strings.Join(bad, "; "))
			it.Fix = "`cutroom look --file " + m.Output + " " + strconv.FormatFloat(p.Start, 'f', 2, 64) + "` to see it; re-render (`cutroom render`); if it repeats, this is a cutroom bug"
		} else {
			it.Status = "pass"
			it.Detail = fmt.Sprintf("%s %q visible over frames %d–%d (%.3f–%.3f, card differs from the picture by %.0f)%s%s", p.ID, p.Name, fa, fb-1, p.Start, p.End, minIn, edge, clipNote)
		}
		r.add(it)
	}
}

// captionCheck: burned-in caption groups are on screen during their groups (a sample spread over
// the edit) and nothing is drawn in the gaps between them (L48).
func captionCheck(p render.Placement, fps float64, diffIn func(int, [4]int) (float64, bool), r *Report) {
	var bad []string
	var picks []render.CapGroup
	for _, g := range p.Groups {
		if math.Round(g.End*fps)-math.Round(g.Start*fps) >= 3 {
			picks = append(picks, g)
		}
	}
	if len(picks) > 6 {
		var keep []render.CapGroup
		for i := 0; i < 6; i++ {
			keep = append(keep, picks[i*(len(picks)-1)/5])
		}
		picks = keep
	}
	minIn := math.Inf(1)
	for _, g := range picks {
		f := int(math.Round((g.Start + g.End) / 2 * fps))
		if d, ok := diffIn(f, boxPx(render.Placement{Box: g.Box})); ok {
			minIn = math.Min(minIn, d)
			if d < 150 {
				bad = append(bad, fmt.Sprintf("%q not visible at %.3f (difference %.0f)", g.Text, float64(f)/fps, d))
			}
		}
	}
	// gaps: before the first group, between groups, after the last
	gaps, clean := 0, 0
	prevEnd := 0.0
	ends := append(append([]render.CapGroup{}, p.Groups...), render.CapGroup{Start: p.End})
	for _, g := range ends {
		a, b := int(math.Round(prevEnd*fps)), int(math.Round(g.Start*fps))
		if b-a >= 1 {
			gaps++
			if gaps <= 4 {
				f := (a + b - 1) / 2
				if d, ok := diffIn(f, boxPx(p)); ok {
					if d > 60 {
						bad = append(bad, fmt.Sprintf("something drawn at %.3f, between caption groups (difference %.0f)", float64(f)/fps, d))
					} else {
						clean++
					}
				}
			}
		}
		prevEnd = g.End
	}
	it := Item{Check: "captions"}
	if len(bad) > 0 {
		it.Status, it.Detail = "fail", fmt.Sprintf("burned-in captions: %s", strings.Join(bad, "; "))
		it.Fix = "re-render (`cutroom render`); if it repeats, this is a cutroom bug"
	} else {
		it.Status = "pass"
		it.Detail = fmt.Sprintf("burned-in captions: %d groups; %d sampled on screen (card differs from the picture by ≥ %.0f), %d gap(s) checked clean", len(p.Groups), len(picks), minIn, clean)
	}
	r.add(it)
}

// clipFrames: a clip overlay shows the clip's own frames, at the moment --clip-in says, inside its
// box. Three render frames in the window are cropped to the box at the render's own pixels (as the
// render placed it) and compared, at 64 px wide, with the clip scaled the same way.
func clipFrames(file string, m *render.Manifest, p render.Placement, fps float64) (note, problem string) {
	W, H := float64(m.Width), float64(m.Height)
	bx, by := int(math.Round(p.Box[0]*W)), int(math.Round(p.Box[1]*H))
	bw, bh := evenInt(p.Box[2]*W-float64(bx)), evenInt(p.Box[3]*H-float64(by))
	sw := 64
	sh := max(8, int(math.Round(float64(sw)*float64(bh)/float64(bw))))
	// luma planes as stored: no colour-matrix conversion on either side (a tagged render and an
	// untagged clip would otherwise be converted differently and hide or fake a colour shift)
	small := fmt.Sprintf("scale=%d:%d,format=yuv420p,extractplanes=y", sw, sh)
	fa, fb := int(math.Round(p.Start*fps)), int(math.Round(p.End*fps))
	worst, n := 0.0, 0
	for _, f := range []int{fa + (fb-fa)/4, (fa + fb) / 2, fa + 3*(fb-fa)/4} {
		r, err := deps.Run("ffmpeg", "-v", "error", "-ss", strconv.FormatFloat((float64(f)-0.5)/fps, 'f', 5, 64), "-i", file,
			"-vf", fmt.Sprintf("crop=%d:%d:%d:%d,%s", bw, bh, bx, by, small), "-frames:v", "1", "-f", "rawvideo", "-")
		if err != nil || len(r) < sw*sh {
			continue
		}
		t := p.ClipIn + float64(f-fa)/fps
		sdr := media.ReadSDR(p.Clip, 0, 0)
		args := append([]string{"-v", "error", "-ss", strconv.FormatFloat(math.Max(0, t-2/fps), 'f', 5, 64)}, sdr.Pre...)
		args = append(args, "-i", p.Clip, "-vf", sdr.Chain(fmt.Sprintf("fps=%g,scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d,%s", fps, bw, bh, bw, bh, small)),
			"-frames:v", "5", "-f", "rawvideo", "-")
		b, err := deps.Run("ffmpeg", args...)
		if err != nil || len(b) < sw*sh {
			continue // past the clip's end: its last frame is held (clip-short warns)
		}
		best := math.Inf(1)
		for k := 0; (k+1)*sw*sh <= len(b); k++ {
			best = math.Min(best, mse(r[:sw*sh], b[k*sw*sh:(k+1)*sw*sh]))
		}
		n++
		worst = math.Max(worst, best)
		if best > 60 {
			problem = fmt.Sprintf("the box at %.3f doesn't show the clip's frame at %.2f s (difference %.0f): wrong clip, --clip-in, placement or colours", float64(f)/fps, t, best)
		}
	}
	if n > 0 && problem == "" {
		note = fmt.Sprintf("; shows the clip's own frames from %.2f s (worst difference %.0f)", p.ClipIn, worst)
	}
	return note, problem
}

func evenInt(x float64) int { return int(math.Round(x/2)) * 2 }

// safeZoneCheck: on a vertical render, every overlay (caption groups included; full-frame clips
// excepted) stays inside the safe zones (L41).
func safeZoneCheck(m *render.Manifest, r *Report) {
	if m.Height <= m.Width || len(m.Overlays) == 0 {
		return
	}
	top, bottom, right := render.SafeZone(m.Width, m.Height)
	W, H := float64(m.Width), float64(m.Height)
	var bad []string
	n := 0
	for _, p := range m.Overlays {
		if p.Kind == "clip" && p.Full {
			continue
		}
		boxes := [][4]float64{p.Box}
		if p.Kind == "captions" {
			boxes = nil
			for _, g := range p.Groups {
				boxes = append(boxes, g.Box)
			}
		}
		for _, b := range boxes {
			n++
			x1, y0, y1 := int(b[2]*W), int(b[1]*H), int(b[3]*H)
			if y0 < top || y1 > bottom || x1 > right {
				bad = append(bad, fmt.Sprintf("%s spans x ..%d, y %d–%d", p.ID, x1, y0, y1))
				break
			}
		}
	}
	it := Item{Check: "safe-zones"}
	if len(bad) > 0 {
		it.Status = "warn"
		it.Detail = fmt.Sprintf("outside the vertical safe zones (y %d–%d, x ≤ %d): %s", top, bottom, right, strings.Join(bad, "; "))
		it.Fix = "`cutroom timeline` (overlay-unsafe) says which layout to change"
	} else {
		it.Status = "pass"
		it.Detail = fmt.Sprintf("%d overlay box(es) inside the vertical safe zones (y %d–%d, x ≤ %d)", n, top, bottom, right)
	}
	r.add(it)
}

// framingCheck: a filled (cropped) render keeps the speaker the source shows in its crop. The
// speaker is found in the source over the span the edit uses (skin-coloured movement); nothing is
// said when the source has no one in it.
func framingCheck(m *render.Manifest, r *Report) {
	if !m.Fill || len(m.Segments) == 0 {
		return
	}
	dur := map[string]float64{}
	lo, hi := map[string]float64{}, map[string]float64{}
	main := ""
	for _, g := range m.Segments {
		if _, ok := lo[g.Source]; !ok {
			lo[g.Source], hi[g.Source] = g.In, g.Out
		}
		lo[g.Source], hi[g.Source] = math.Min(lo[g.Source], g.In), math.Max(hi[g.Source], g.Out)
		dur[g.Source] += g.Out - g.In
		if main == "" || dur[g.Source] > dur[main] {
			main = g.Source
		}
	}
	src, ok := m.Sources[main]
	if !ok {
		return
	}
	_, vw, vh, rot, _ := probeVideo(src.Path)
	if rot == 90 || rot == 270 {
		vw, vh = vh, vw
	}
	if vw == 0 || vh == 0 {
		return
	}
	sp := perceive.FindSpeaker(src.Path, lo[main], hi[main]-lo[main])
	if !sp.Found {
		return
	}
	a, b := render.CropSpan(m.FillX, vw, vh, m.Width, m.Height)
	if render.CropHolds(m.FillX, sp.X, vw, vh, m.Width, m.Height) {
		r.add(Item{Check: "framing", Status: "pass", Detail: fmt.Sprintf("the speaker (found at x %.2f of %s) is inside the crop (%.2f–%.2f)", sp.X, main, a, b)})
		return
	}
	r.add(Item{Check: "framing", Status: "warn", Detail: fmt.Sprintf("the speaker is at x %.2f of %s, but the crop keeps %.2f–%.2f: they are cut off or out of the picture", sp.X, main, a, b),
		Fix: "`cutroom edit frame --anchor auto` (or --anchor x,y), render again, look at it"})
}
