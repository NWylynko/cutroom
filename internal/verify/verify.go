// Package verify checks a rendered file against the manifest of what it should contain: frame count,
// picture (render frames vs source frames), A/V sync per segment, words (re-transcription vs the
// timeline's words), sound at cut edges (L33), loudness and output tags. It measures the file; it does
// not trust the renderer.
package verify

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/cutroom/cutroom/internal/deps"
	"github.com/cutroom/cutroom/internal/render"
	"github.com/cutroom/cutroom/internal/speech"
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
	r.add(Item{Check: "format", Status: st, Detail: fmt.Sprintf("%dx%d, display rotation %d°, colour %s", w, h, rot, orDash(cs))})
	if o.CurrentHash != "" && o.CurrentHash != m.TimelineHash {
		r.add(Item{Check: "stale", Status: "warn", Detail: "the timeline has changed since this file was rendered", Fix: "`cutroom render` again, then check"})
	}

	// 2. picture: sampled render frames vs source frames (±3)
	pictureCheck(file, m, fps, r)

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
	src := m.Sources[m.Audio]
	srcWav := speech.WavPath(o.CacheDir, m.Audio)
	if _, err := os.Stat(srcWav); err != nil {
		srcWav = filepath.Join(tmp, "src.wav")
		if err := speech.ExtractWav(src.Path, srcWav); err != nil {
			return nil, err
		}
	}
	sx, _ := speech.ReadWav(srcWav)
	syncCheck(rx, sx, src.AudioStart, m, r)
	if o.Model != "" {
		wordCheck(renderWav, tmp, m, o, r)
	}
	if tr := o.Transcripts[m.Audio]; tr != nil && len(sx) > 0 {
		edgeCheck(sx, src.AudioStart, tr, m, r)
		cutThroughWords(tr, m, r)
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
	}
	var ss []sample
	for _, g := range m.Segments {
		start := int(math.Round(g.OutStart * fps))
		picks := []int{1, int(g.Frames) / 2, int(g.Frames) - 2}
		if g.Frames < 4 {
			picks = []int{int(g.Frames) / 2}
		}
		for _, k := range picks {
			ss = append(ss, sample{seg: g, f: start + k, srcT: g.In + float64(k)/fps, skipped: g.Zoom > 1.001})
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
			seek := s.srcT - 3/fps
			lead := 3
			if seek < 0 {
				lead = int(math.Round(s.srcT * fps))
				seek = 0
			}
			burst := grey(path, "", seek, 7)
			best, bi := math.Inf(1), 0
			errs := make([]float64, len(burst))
			for j, b := range burst {
				errs[j] = mse(frames[s.f], b)
				if errs[j] < best {
					best, bi = errs[j], j
				}
			}
			if len(burst) == 0 {
				s.skipped = true
				return
			}
			s.offset, s.err = bi-lead, best
			if lead < len(errs) && errs[lead]-best < 1.0 {
				s.offset, s.ambig = 0, true // neighbouring frames look alike (static shot)
			}
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
	d := fmt.Sprintf("%d sampled frames match their source frames (worst difference %.1f)", n, worst)
	it := Item{Check: "picture"}
	if bad > 0 {
		st = "fail"
		d = fmt.Sprintf("%d of %d sampled frames don't match their source frame; first at output %.3f (segment %s): offset %+d frames, difference %.1f",
			bad, n, float64(first.f)/fps, first.seg.ID, first.offset, first.err)
		it.At, it.Segment = float64(first.f)/fps, first.seg.ID
		it.Fix = "look at it: `cutroom look " + strconv.FormatFloat(first.srcT, 'f', 2, 64) + " --source " + first.seg.Source + "` vs the render frame"
	}
	it.Status, it.Detail = st, d
	r.add(it)
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

func syncCheck(rx, sx []float64, audioStart float64, m *render.Manifest, r *Report) {
	win, maxLag := int(0.5*sr), int(0.15*sr)
	var measured []SegSync
	worst := 0.0
	var worstSeg SegSync
	for _, g := range m.Segments {
		d := float64(g.Frames) / rate(m.FPS)
		ss := SegSync{Segment: g.ID, OutStart: g.OutStart}
		// try a few windows inside the segment; keep the most confident speech window
		bestC := 0.0
		for _, frac := range []float64{0.5, 0.3, 0.7} {
			if d < 0.6 {
				break
			}
			to := g.OutStart + d*frac
			so := g.In + d*frac - audioStart
			oi, si := int(to*sr)-win/2, int(so*sr)-win/2
			if oi < 0 || oi+win > len(rx) || si-maxLag < 0 || si+win+maxLag > len(sx) {
				continue
			}
			seg := rx[oi : oi+win]
			if rms(seg) < 0.01 {
				continue
			}
			lag, c := xcorr(seg, sx[si-maxLag:si+win+maxLag], maxLag)
			if c > bestC && c > 0.5 {
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
		r.Sync = append(r.Sync, ss)
	}
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
	it.Detail = fmt.Sprintf("%d of %d segments measured: audio vs source median %+.0f ms, worst %+.0f ms (limit 45)", len(measured), len(m.Segments), lags[len(lags)/2], worst)
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
			if mid >= g.In && mid < g.Out {
				if n := speech.Norm(w.Text); n != "" {
					exp = append(exp, ew{n, g.OutStart + w.Start - g.In})
				}
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
	if len(gaps) == 0 {
		r.add(Item{Check: "words", Status: "pass", Detail: fmt.Sprintf("all %d expected words heard in order in the render", n)})
		return
	}
	r.add(Item{Check: "words", Status: "pass", Detail: fmt.Sprintf("%d of %d expected words heard in order; %d difference(s) listed below", matched, n, len(gaps))})
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

func segAt(m *render.Manifest, t float64) string {
	for _, g := range m.Segments {
		if t >= g.OutStart && t < g.OutStart+float64(g.Frames)/rate(m.FPS) {
			return g.ID
		}
	}
	return ""
}

// --- edges (L33) -----------------------------------------------------------------------------

func edgeCheck(sx []float64, audioStart float64, tr *speech.Transcript, m *render.Manifest, r *Report) {
	env := speech.NewEnvelope(sx)
	env.Floor, env.Peak = tr.RoomTone, tr.SpeechDB
	env.Thr = env.Floor + math.Max(10, 0.35*(env.Peak-env.Floor))
	regions := env.Regions()
	found := 0
	for _, g := range m.Segments {
		if g.Source != m.Audio {
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
			found++
			at := g.OutStart + math.Max(a, g.In) - g.In
			r.add(Item{Check: "stray-sound", Status: "warn", At: round3(at), Segment: g.ID,
				Detail: fmt.Sprintf("%.2f s of sound with no words kept in segment %s (source %.2f–%.2f, peak %.0f dB): a fragment of cut speech, a breath, or a noise", ov, g.ID, math.Max(a, g.In), math.Min(b, g.Out), rg.PeakDB),
				Fix:    fmt.Sprintf("`cutroom listen %.2f-%.2f`; cut it with `cutroom edit cut %.2f-%.2f` if unwanted", math.Max(a, g.In)-0.5, math.Min(b, g.Out)+0.5, math.Max(a, g.In), math.Min(b, g.Out))})
		}
	}
	if found == 0 {
		r.add(Item{Check: "stray-sound", Status: "pass", Detail: "no sound without words kept at any cut or inside segments"})
	}
}

// cutThroughWords reports segment edges that fall inside a word (a clipped word in the render).
func cutThroughWords(tr *speech.Transcript, m *render.Manifest, r *Report) {
	fps := rate(m.FPS)
	for _, g := range m.Segments {
		if g.Source != m.Audio {
			continue
		}
		for _, w := range tr.Words {
			for _, e := range []struct {
				t    float64
				side string
				out  float64
			}{{g.In, "start", g.OutStart}, {g.Out, "end", g.OutStart + float64(g.Frames)/fps}} {
				if e.t > w.Start+0.04 && e.t < w.End-0.04 {
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
