// Command grade checks an agent's eval output independently of cutroom (it shares no code with it).
//
//	go run ./evals/grade [-fixtures evals/work/fixtures] <task> <workdir>
//
// It runs its own whisper pass, matches output frames against the original reference video
// (talk.mp4: testsrc frames carry a counter, so every frame is unique), and cross-correlates the
// output audio against the reference audio at the matched source time to measure A/V sync.
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type spec struct {
	Fixtures       []string  `json:"fixtures"`
	Output         string    `json:"output"`
	MustContain    []string  `json:"must_contain"`
	MustNotContain []string  `json:"must_not_contain"`
	ExactlyOnce    []string  `json:"exactly_once"`
	Duration       []float64 `json:"duration"`
	Checks         []string  `json:"checks"`
	// Ref is the fixture whose frames the output must come from (default talk.mp4). RefSDR decodes
	// it through VideoToolbox HLG→bt709 (the recipe real productions used), sequentially and without
	// seeking, so the reference timing is the grader's own.
	Ref    string `json:"ref,omitempty"`
	RefSDR bool   `json:"ref_sdr,omitempty"`
	// FrameErr is the luma MSE under which an output frame counts as matching a reference frame
	// (default 30, tuned on SDR fixtures; a flipped frame scores in the thousands).
	FrameErr float64 `json:"frame_err,omitempty"`
	// SyncLimit (ms, default 45). The HDR fixture is HEVC-encoded test pattern whose neighbouring
	// frames differ less than its encoding noise, so the grader can place a frame only to ±1 (33 ms).
	SyncLimit float64 `json:"sync_limit_ms,omitempty"`
	// AltRef is a second source the output may also draw from (a pickup take). Frames and sync are
	// judged against whichever reference a frame matches; `alt_used` requires it to appear.
	AltRef string `json:"alt_ref,omitempty"`
	// Cards: a card (text or image overlay) must cover the picture while each phrase's line is said;
	// Clean: no card while these lines are said (burned captions in the bottom band excepted).
	Cards []string `json:"cards,omitempty"`
	Clean []string `json:"clean,omitempty"`
	// BurnedCaptions: the bottom band carries captions while lines are said.
	BurnedCaptions bool `json:"burned_captions,omitempty"`
	// SRT: a caption file whose words and timing must match the output's speech.
	SRT string `json:"srt,omitempty"`
	// VerticalCrop: the output is a 9:16 crop of the reference at an anchor the agent chose; the
	// grader finds it (21 positions) and matches frames against the reference cropped the same way.
	VerticalCrop bool `json:"vertical_crop,omitempty"`
	// SafeBottom: nothing drawn in the bottom sixth of the frame while lines are said (a vertical
	// video's own buttons and description sit there).
	SafeBottom bool `json:"safe_bottom,omitempty"`
	// CaptionColor: text the brief asked for in this colour (#rrggbb) is on screen while lines are
	// said — pixels near the colour in most sampled frames (the fixture has none of it).
	CaptionColor string `json:"caption_color,omitempty"`
	// AltDuring: while these phrases are said, the picture comes from AltRef (a B-roll cutaway).
	AltDuring []string `json:"alt_during,omitempty"`
}

type result struct {
	Name   string `json:"name"`
	Pass   bool   `json:"pass"`
	Detail string `json:"detail"`
}

const (
	gw, gh = 160, 90 // grey thumbnail size for frame matching
	refFPS = 30.0
	sr     = 16000
)

var fixtures string

func main() {
	flag.StringVar(&fixtures, "fixtures", "evals/work/fixtures", "fixture directory (reference media + manifests)")
	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "usage: grade <task> <workdir>")
		os.Exit(1)
	}
	task, dir := flag.Arg(0), flag.Arg(1)
	var sp spec
	b, err := os.ReadFile(filepath.Join("evals/tasks", task+".json"))
	must(err)
	must(json.Unmarshal(b, &sp))

	var res []result
	add := func(name string, pass bool, f string, a ...any) {
		res = append(res, result{name, pass, fmt.Sprintf(f, a...)})
	}
	out := filepath.Join(dir, sp.Output)
	if _, err := os.Stat(out); err != nil {
		add("output-exists", false, "%s not found", sp.Output)
		finish(task, res)
	}
	add("output-exists", true, "%s", sp.Output)

	vd, ad := streamDur(out, "v:0"), streamDur(out, "a:0")
	if len(sp.Duration) == 2 {
		add("duration", vd >= sp.Duration[0] && vd <= sp.Duration[1], "%.3f s (want %.1f–%.1f)", vd, sp.Duration[0], sp.Duration[1])
	}
	add("av-lengths", math.Abs(vd-ad) <= 0.1, "video %.3f s, audio %.3f s", vd, ad)

	// Independent transcript.
	words := transcribe(out)
	text := strings.Join(words, " ")
	for _, p := range sp.MustContain {
		n := count(words, p)
		add("contains: "+p, n >= 1, "found %d", n)
	}
	for _, p := range sp.MustNotContain {
		n := count(words, p)
		add("absent: "+p, n == 0, "found %d", n)
	}
	for _, p := range sp.ExactlyOnce {
		n := count(words, p)
		add("once: "+p, n == 1, "found %d", n)
	}

	for _, c := range sp.Checks {
		switch {
		case c == "no_rotation_meta":
			r := rotation(out)
			add(c, r == 0, "display rotation %d", r)
		case c == "upright", c == "frames_from_source", c == "av_sync", c == "alt_used":
			// computed together below
		case c == "sdr_bt709":
			tr, pr := colourTags(out)
			add(c, tr == "bt709" && pr == "bt709", "colour transfer %s, primaries %s (want bt709/bt709)", orDash(tr), orDash(pr))
		case c == "tone_mapped":
			// L23: a plain matrix conversion of HLG keeps the tags right but flattens the picture.
			// Compare colourfulness and luma spread with the reference decoded through a tone map.
			ref := filepath.Join(fixtures, sp.Ref)
			os_, or_ := toneStats(out, nil, "")
			rs, rr := toneStats(ref, []string{"-hwaccel", "videotoolbox", "-hwaccel_output_format", "videotoolbox_vld", "-display_rotation", "0"},
				"scale_vt=color_matrix=bt709:color_primaries=bt709:color_transfer=bt709,hwdownload,format=p010le")
			ok := rs > 0 && rr > 0 && math.Abs(os_/rs-1) < 0.1 && math.Abs(or_/rr-1) < 0.1
			add(c, ok, "saturation %.1f vs tone-mapped source %.1f; luma range %.0f vs %.0f (within 10%%)", os_, rs, or_, rr)
		case strings.HasPrefix(c, "no_tone="):
			// a B-roll clip's own sound (a pure tone above the talk's band) must not be in the output
			f := strings.TrimPrefix(c, "no_tone=")
			tone, voice := bandRMS(out, f+":100"), bandRMS(out, "1500:1000")
			add(c, tone-voice < -30, "energy at %s Hz %.1f dB vs speech band %.1f dB (want 30 dB under)", f, tone, voice)
		case c == "vertical_9x16":
			w, h := size(out)
			add(c, h > w && math.Abs(float64(w)/float64(h)-9.0/16) < 0.01, "%dx%d", w, h)
		case strings.HasPrefix(c, "size="):
			w, h := size(out)
			add(c, fmt.Sprintf("size=%dx%d", w, h) == c, "%dx%d", w, h)
		case strings.HasPrefix(c, "audio_from="):
			// Compare the output's high-band share with both candidate sources measured the same way.
			db := func(p string) float64 { return 10 * math.Log10(highBandShare(p)+1e-12) }
			o, want, cam := db(out), db(filepath.Join(fixtures, strings.TrimPrefix(c, "audio_from="))), db(filepath.Join(fixtures, "talk.mp4"))
			add(c, math.Abs(o-want) < math.Abs(o-cam), "speech energy above 4 kHz: output %.1f dB, %s %.1f dB, camera %.1f dB", o, strings.TrimPrefix(c, "audio_from="), want, cam)
		default:
			add(c, false, "unknown check")
		}
	}
	if sp.SRT != "" {
		srtCheck(filepath.Join(dir, sp.SRT), out, vd, add)
	}
	if has(sp.Checks, "upright") || has(sp.Checks, "frames_from_source") || has(sp.Checks, "av_sync") || has(sp.Checks, "alt_used") || len(sp.Cards)+len(sp.Clean) > 0 || sp.BurnedCaptions || sp.SafeBottom || len(sp.AltDuring) > 0 {
		frameAndSync(out, sp, add)
	}
	if sp.CaptionColor != "" {
		n, hit, best := colorFrames(out, sp.CaptionColor)
		add("caption_color", n > 0 && float64(hit) >= 0.6*float64(n), "%d of %d sampled frames show %s text (most: %d px)", hit, n, sp.CaptionColor, best)
	}
	res = append(res, result{"transcript", true, truncate(text, 400)})
	finish(task, res)
}

func finish(task string, res []result) {
	fails := 0
	for _, r := range res {
		mark := "ok  "
		if !r.Pass {
			mark, fails = "FAIL", fails+1
		}
		fmt.Printf("%s %-40s %s\n", mark, r.Name, r.Detail)
	}
	j, _ := json.Marshal(map[string]any{"task": task, "pass": fails == 0, "results": res})
	fmt.Printf("JSON %s\n", j)
	if fails > 0 {
		fmt.Printf("GRADE %s: FAIL (%d)\n", task, fails)
		os.Exit(4)
	}
	fmt.Printf("GRADE %s: PASS\n", task)
}

// --- frames and sync -------------------------------------------------------------------------

func frameAndSync(out string, sp spec, add func(string, bool, string, ...any)) {
	checks := sp.Checks
	maxErr := 30.0
	if sp.FrameErr > 0 {
		maxErr = sp.FrameErr
	}
	ref := filepath.Join(fixtures, "talk.mp4")
	if sp.Ref != "" {
		ref = filepath.Join(fixtures, sp.Ref)
	}
	var refFrames [][]byte
	if sp.RefSDR {
		refFrames = greyFramesPre(ref, []string{"-hwaccel", "videotoolbox", "-hwaccel_output_format", "videotoolbox_vld"},
			"scale_vt=color_matrix=bt709:color_primaries=bt709:color_transfer=bt709,hwdownload,format=p010le,transpose=clock,transpose=clock")
	} else if sp.VerticalCrop {
		var anchor float64
		refFrames, anchor = verticalRef(ref, out)
		add("vertical_crop", refFrames != nil, "output is a 9:16 crop of the source around x = %.2f", anchor)
	} else {
		refFrames = greyFrames(ref, "")
	}
	refSets := [][][]byte{refFrames}
	refPaths := []string{ref}
	if sp.AltRef != "" {
		refSets = append(refSets, greyFrames(filepath.Join(fixtures, sp.AltRef), ""))
		refPaths = append(refPaths, filepath.Join(fixtures, sp.AltRef))
	}
	outFrames := greyFrames(out, "fps=30") // resample to a 30 fps grid in output time
	// with cards or burned captions over the picture, match on the best 3/4 of the frame's tiles
	errf := mse
	if len(sp.Cards)+len(sp.Clean) > 0 || sp.BurnedCaptions || sp.SafeBottom {
		errf = robust
	}
	if len(outFrames) == 0 || len(refFrames) == 0 {
		add("frames", false, "could not decode frames")
		return
	}
	type match struct {
		outT, refT, err, flippedErr float64
		ref                         int // which reference the frame came from
		covered, bottom             int // tiles that differ from the reference (cards), of them in the bottom band
		lowest                      int // of them in the bottom sixth (row 6 of 6)
	}
	var ms []match
	for i := 0; i < len(outFrames); i += 6 { // every 0.2 s
		// match a short run of frames, not one: neighbouring frames of a near-static picture differ
		// less than encoding noise, and a ±1 frame error is a ±33 ms sync error. The run is taken on
		// whichever side of frame i fits better, so a cut next to i doesn't spoil it.
		bestRun, bi, br := math.Inf(1), 0, 0
		for r, refFrames := range refSets {
			run := func(j, d0, d1 int) float64 {
				t, n := 0.0, 0
				for d := d0; d <= d1; d++ {
					if i+d < 0 || i+d >= len(outFrames) || j+d < 0 || j+d >= len(refFrames) {
						continue
					}
					t += errf(outFrames[i+d], refFrames[j+d])
					n++
				}
				if n == 0 {
					return math.Inf(1)
				}
				return t / float64(n)
			}
			for j, rf := range refFrames {
				if errf(outFrames[i], rf) > 4*maxErr+1000 {
					continue // cheap reject before scoring the run
				}
				if e := math.Min(run(j, -2, 0), run(j, 0, 2)); e < bestRun {
					bestRun, bi, br = e, j, r
				}
			}
		}
		if math.IsInf(bestRun, 1) { // nothing close: fall back to the main reference's nearest frame
			for j, rf := range refSets[0] {
				if e := errf(outFrames[i], rf); e < bestRun {
					bestRun, bi, br = e, j, 0
				}
			}
		}
		best := errf(outFrames[i], refSets[br][bi])
		fl := errf(flip(outFrames[i]), refSets[br][bi])
		cov, bot, low := covered(outFrames[i], refSets[br][bi])
		ms = append(ms, match{float64(i) / refFPS, float64(bi) / refFPS, best, fl, br, cov, bot, low})
	}
	// upright: best match error small and much smaller than the flipped comparison
	bad := 0
	for _, m := range ms {
		if !(m.err < maxErr && m.err < m.flippedErr) {
			bad++
		}
	}
	if has(checks, "upright") || has(checks, "frames_from_source") {
		es, fs := make([]float64, len(ms)), make([]float64, len(ms))
		for i, m := range ms {
			es[i], fs[i] = m.err, m.flippedErr
		}
		sort.Float64s(es)
		sort.Float64s(fs)
		add("frames_from_source/upright", bad == 0, "%d/%d sampled frames match an upright source frame (err median %.1f, worst %.1f; flipped median %.1f)", len(ms)-bad, len(ms), es[len(es)/2], es[len(es)-1], fs[len(fs)/2])
	}
	if has(checks, "alt_used") {
		n := 0
		for _, m := range ms {
			if m.ref == 1 && m.err < maxErr {
				n++
			}
		}
		add("alt_used", n >= 5, "%d sampled frames (≈%.1f s) come from %s", n, float64(n)*0.2, sp.AltRef)
	}
	if len(sp.AltDuring) > 0 {
		ws := transcribeTimed(out)
		for _, ph := range sp.AltDuring {
			a, b, ok := phraseSpan(ws, ph)
			if !ok {
				add("cutaway during: "+ph, false, "phrase not heard in the output")
				continue
			}
			in, alt, outside, main := 0, 0, 0, 0
			for _, m := range ms {
				switch {
				case m.outT > a+0.2 && m.outT < b-0.2:
					in++
					if m.ref == 1 && m.err < maxErr {
						alt++
					}
				case m.outT < a-0.5 || m.outT > b+0.5:
					outside++
					if m.ref == 0 && m.err < maxErr {
						main++
					}
				}
			}
			add("cutaway during: "+ph, in > 0 && float64(alt) >= 0.8*float64(in), "%d of %d sampled frames while it is said (%.2f–%.2f s) come from %s", alt, in, a, b, sp.AltRef)
			add("speaker elsewhere", outside > 0 && float64(main) >= 0.9*float64(outside), "%d of %d sampled frames away from it come from the talk", main, outside)
		}
	}
	if len(sp.Cards)+len(sp.Clean) > 0 || sp.BurnedCaptions || sp.SafeBottom {
		lines := fixtureLines(ref)
		during := func(phrase string) (n, card, bottom int) {
			for _, l := range lines {
				if !strings.Contains(strings.Join(tokens(l.Say), " "), strings.Join(tokens(phrase), " ")) {
					continue
				}
				for _, m := range ms {
					if m.ref == 0 && m.refT > l.Start+0.3 && m.refT < l.End-0.3 {
						n++
						if m.covered-m.bottom >= 2 || (m.covered >= 3 && !sp.BurnedCaptions) {
							card++
						}
						if m.bottom >= 1 {
							bottom++
						}
					}
				}
			}
			return
		}
		for _, ph := range sp.Cards {
			n, c, _ := during(ph)
			add("card during: "+ph, n > 0 && float64(c) >= 0.6*float64(n), "%d of %d sampled frames while it is said carry a card", c, n)
		}
		for _, ph := range sp.Clean {
			n, c, _ := during(ph)
			add("no card during: "+ph, n > 0 && float64(c) <= 0.2*float64(n), "%d of %d sampled frames while it is said carry a card", c, n)
		}
		if sp.SafeBottom {
			n, bad := 0, 0
			for _, m := range ms {
				if m.ref == 0 && m.err < maxErr {
					n++
					if m.lowest > 0 {
						bad++
					}
				}
			}
			add("safe_bottom", n > 0 && bad == 0, "%d of %d sampled frames have something drawn in the bottom sixth", bad, n)
		}
		if sp.BurnedCaptions {
			n, b := 0, 0
			for _, l := range lines {
				if l.Say == "" {
					continue
				}
				x, _, y := during(l.Say)
				n, b = n+x, b+y
			}
			add("burned_captions", n > 0 && float64(b) >= 0.6*float64(n), "%d of %d sampled speech frames have something drawn in the bottom band", b, n)
		}
	}
	if !has(checks, "av_sync") {
		return
	}
	outA := pcm(out)
	refAs := make([][]float64, len(refPaths))
	for i, p := range refPaths {
		refAs[i] = pcm(p)
	}
	var lags []float64
	win, maxLag := int(0.5*sr), int(0.25*sr)
	for _, m := range ms {
		refA := refAs[m.ref]
		o := int(m.outT * sr)
		r := int(m.refT * sr)
		if o-win/2 < 0 || o+win/2 > len(outA) || r-win/2-maxLag < 0 || r+win/2+maxLag > len(refA) {
			continue
		}
		seg := outA[o-win/2 : o+win/2]
		if rms(seg) < 0.02 { // not speech; correlation meaningless
			continue
		}
		lag, c := xcorr(seg, refA[r-win/2-maxLag:r+win/2+maxLag], maxLag)
		if c < 0.6 {
			continue // window probably spans a cut
		}
		lags = append(lags, float64(lag)/sr*1000)
	}
	if len(lags) < 5 {
		add("av_sync", false, "only %d confident sync windows (need ≥5)", len(lags))
		return
	}
	sort.Float64s(lags)
	med := lags[len(lags)/2]
	mx := math.Max(math.Abs(lags[0]), math.Abs(lags[len(lags)-1]))
	lim := 45.0
	if sp.SyncLimit > 0 {
		lim = sp.SyncLimit
	}
	add("av_sync", mx <= lim, "%d windows, median %+.0f ms, worst %.0f ms (limit %.0f)", len(lags), med, mx, lim)
}

func worst[T any](xs []T, f func(T) float64) float64 {
	w := 0.0
	for _, x := range xs {
		w = math.Max(w, f(x))
	}
	return w
}

func greyFrames(p, extra string) [][]byte { return greyFramesPre(p, nil, extra) }

func greyFramesPre(p string, pre []string, extra string) [][]byte {
	// raw luma plane: independent of colour tags (untagged sources vs bt709-tagged renders)
	vf := fmt.Sprintf("scale=%d:%d,format=yuv420p,extractplanes=y", gw, gh)
	if extra != "" {
		vf = extra + "," + vf
	}
	args := append([]string{"-v", "error"}, pre...)
	if len(pre) > 0 {
		args = append(args, "-display_rotation", "0") // hardware frames: rotate by hand in extra
	}
	args = append(args, "-i", p, "-vf", vf, "-f", "rawvideo", "-")
	b, err := exec.Command("ffmpeg", args...).Output()
	if err != nil {
		return nil
	}
	n := len(b) / (gw * gh)
	fr := make([][]byte, n)
	for i := range fr {
		fr[i] = b[i*gw*gh : (i+1)*gw*gh]
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

// tileErrs: MSE of each of 8×6 tiles (20×15 px of the 160×90 grey frame).
func tileErrs(a, b []byte) []float64 {
	es := make([]float64, 48)
	for ty := 0; ty < 6; ty++ {
		for tx := 0; tx < 8; tx++ {
			s := 0.0
			for y := ty * 15; y < ty*15+15; y++ {
				for x := tx * 20; x < tx*20+20; x++ {
					d := float64(a[y*gw+x]) - float64(b[y*gw+x])
					s += d * d
				}
			}
			es[ty*8+tx] = s / 300
		}
	}
	return es
}

// robust: mean error of the best 36 of 48 tiles — a card or caption over part of the frame
// doesn't stop it matching its source frame.
func robust(a, b []byte) float64 {
	es := tileErrs(a, b)
	sort.Float64s(es)
	t := 0.0
	for _, e := range es[:36] {
		t += e
	}
	return t / 36
}

// covered counts tiles that clearly differ from the reference frame (something drawn over them),
// and how many of those are in the bottom band (rows 5–6 of 6: where captions sit).
func covered(a, b []byte) (n, bottom, lowest int) {
	for i, e := range tileErrs(a, b) {
		if e > 400 {
			n++
			if i >= 32 {
				bottom++
			}
			if i >= 40 {
				lowest++
			}
		}
	}
	return
}

type fixtureLine struct {
	Say        string
	Start, End float64
}

// fixtureLines reads the ground-truth line times of a reference fixture (<name>.json).
func fixtureLines(ref string) []fixtureLine {
	b, err := os.ReadFile(strings.TrimSuffix(ref, filepath.Ext(ref)) + ".json")
	if err != nil {
		return nil
	}
	var m struct {
		Lines []struct {
			Say   string  `json:"say"`
			Start float64 `json:"start"`
			End   float64 `json:"end"`
		} `json:"lines"`
	}
	json.Unmarshal(b, &m)
	var ls []fixtureLine
	for _, l := range m.Lines {
		ls = append(ls, fixtureLine{l.Say, l.Start, l.End})
	}
	return ls
}

func flip(a []byte) []byte {
	o := make([]byte, len(a))
	for i := range a {
		o[len(a)-1-i] = a[i]
	}
	return o
}

func pcm(p string) []float64 {
	b, err := exec.Command("ffmpeg", "-v", "error", "-i", p, "-vn", "-ac", "1", "-ar", strconv.Itoa(sr),
		"-af", "highpass=f=200,lowpass=f=3400", "-f", "s16le", "-").Output()
	if err != nil {
		return nil
	}
	o := make([]float64, len(b)/2)
	for i := range o {
		o[i] = float64(int16(binary.LittleEndian.Uint16(b[2*i:]))) / 32768
	}
	return o
}

func rms(x []float64) float64 {
	s := 0.0
	for _, v := range x {
		s += v * v
	}
	return math.Sqrt(s / float64(len(x)))
}

// xcorr finds the lag in [-maxLag, maxLag] where seg best matches ref (ref is seg-length + 2*maxLag,
// centred). Positive lag = output audio is late relative to the picture. Returns normalised corr.
func xcorr(seg, ref []float64, maxLag int) (int, float64) {
	best, bl := -2.0, 0
	es := 0.0
	for _, v := range seg {
		es += v * v
	}
	for lag := -maxLag; lag <= maxLag; lag += 4 {
		off := maxLag - lag
		s, er := 0.0, 0.0
		for i, v := range seg {
			r := ref[off+i]
			s += v * r
			er += r * r
		}
		c := s / math.Sqrt(es*er+1e-12)
		if c > best {
			best, bl = c, lag
		}
	}
	// refine ±4 samples
	for lag := bl - 4; lag <= bl+4; lag++ {
		if lag < -maxLag || lag > maxLag {
			continue
		}
		off := maxLag - lag
		s, er := 0.0, 0.0
		for i, v := range seg {
			r := ref[off+i]
			s += v * r
			er += r * r
		}
		if c := s / math.Sqrt(es*er+1e-12); c > best {
			best, bl = c, lag
		}
	}
	return bl, best
}

// highBandShare: energy above 4 kHz / total, over the loudest 30% of 50 ms windows (speech).
func highBandShare(p string) float64 {
	read := func(af string) []float64 {
		b, _ := exec.Command("ffmpeg", "-v", "error", "-i", p, "-vn", "-ac", "1", "-ar", "48000", "-af", af, "-f", "s16le", "-").Output()
		o := make([]float64, len(b)/2)
		for i := range o {
			o[i] = float64(int16(binary.LittleEndian.Uint16(b[2*i:]))) / 32768
		}
		return o
	}
	full, hi := read("anull"), read("highpass=f=4000,highpass=f=4000,highpass=f=4000")
	n := 2400
	type w struct{ f, h float64 }
	var ws []w
	for i := 0; i+n <= len(full) && i+n <= len(hi); i += n {
		ws = append(ws, w{rms(full[i : i+n]), rms(hi[i : i+n])})
	}
	sort.Slice(ws, func(i, j int) bool { return ws[i].f > ws[j].f })
	ws = ws[:max(1, len(ws)*3/10)]
	sf, sh := 0.0, 0.0
	for _, x := range ws {
		sf += x.f * x.f
		sh += x.h * x.h
	}
	return sh / (sf + 1e-12)
}

// --- transcript ------------------------------------------------------------------------------

func transcribe(p string) []string {
	tmp, _ := os.MkdirTemp("", "grade")
	defer os.RemoveAll(tmp)
	wav := filepath.Join(tmp, "a.wav")
	if err := exec.Command("ffmpeg", "-v", "error", "-i", p, "-vn", "-ac", "1", "-ar", "16000", wav).Run(); err != nil {
		return nil
	}
	model := os.Getenv("GRADE_WHISPER_MODEL")
	if model == "" {
		home, _ := os.UserHomeDir()
		model = filepath.Join(home, ".cache/cutroom/models/ggml-base.en.bin")
	}
	if err := exec.Command("whisper-cli", "-m", model, "-f", wav, "-oj", "-of", filepath.Join(tmp, "t"), "-np").Run(); err != nil {
		return nil
	}
	b, _ := os.ReadFile(filepath.Join(tmp, "t.json"))
	var d struct {
		Transcription []struct{ Text string } `json:"transcription"`
	}
	json.Unmarshal(b, &d)
	var all []string
	for _, s := range d.Transcription {
		all = append(all, s.Text)
	}
	return tokens(strings.Join(all, " "))
}

var nonWord = regexp.MustCompile(`[^a-z0-9' ]+`)

func tokens(s string) []string {
	s = nonWord.ReplaceAllString(strings.ToLower(s), " ")
	return strings.Fields(s)
}

// count: non-overlapping fuzzy occurrences (≥80% of the phrase's tokens, in order, within a
// window of 1.5× its length). Single-token phrases must match exactly.
func count(words []string, phrase string) int {
	p := tokens(phrase)
	if len(p) == 0 {
		return 0
	}
	need := int(math.Ceil(0.8 * float64(len(p))))
	if len(p) == 1 {
		need = 1
	}
	win := int(math.Ceil(1.5 * float64(len(p))))
	n := 0
	for i := 0; i < len(words); {
		if words[i] != p[0] && (len(p) < 2 || words[i] != p[1]) {
			i++
			continue
		}
		k, j := 0, i
		for ; j < len(words) && j < i+win && k < len(p); j++ {
			for kk := k; kk < len(p); kk++ {
				if words[j] == p[kk] {
					k = kk + 1
					break
				}
			}
		}
		if k >= need && matched(words[i:j], p) >= need {
			n++
			i = j
		} else {
			i++
		}
	}
	return n
}

func matched(ws, p []string) int {
	k := 0
	for _, w := range ws {
		if k < len(p) && w == p[k] {
			k++
		} else if k+1 < len(p) && w == p[k+1] {
			k += 2
		}
	}
	return k
}

// --- probes ----------------------------------------------------------------------------------

func streamDur(p, s string) float64 {
	b, _ := exec.Command("ffprobe", "-v", "error", "-select_streams", s, "-show_entries", "stream=duration", "-of", "csv=p=0", p).Output()
	f, _ := strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
	return f
}

func rotation(p string) int {
	b, _ := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream_side_data=rotation", "-of", "csv=p=0", p).Output()
	r, _ := strconv.Atoi(strings.TrimSpace(strings.Split(string(b), "\n")[0]))
	return r
}

func size(p string) (int, int) {
	b, _ := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "csv=p=0", p).Output()
	var w, h int
	fmt.Sscanf(strings.TrimSpace(string(b)), "%d,%d", &w, &h)
	return w, h
}

func has(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func colourTags(p string) (transfer, primaries string) {
	b, _ := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=color_transfer,color_primaries", "-of", "default=nw=1", p).Output()
	for _, l := range strings.Split(string(b), "\n") {
		k, v, _ := strings.Cut(strings.TrimSpace(l), "=")
		switch k {
		case "color_transfer":
			transfer = v
		case "color_primaries":
			primaries = v
		}
	}
	return
}

func orDash(s string) string {
	if s == "" || s == "unknown" {
		return "untagged"
	}
	return s
}

var satRe, lowRe, highRe = regexp.MustCompile(`SATAVG=([0-9.]+)`), regexp.MustCompile(`YLOW=([0-9.]+)`), regexp.MustCompile(`YHIGH=([0-9.]+)`)

// toneStats returns mean saturation and mean (YHIGH−YLOW) over one frame per second, in 8-bit units.
func toneStats(p string, pre []string, vf string) (sat, rng float64) {
	f := "fps=1,format=yuv420p,signalstats,metadata=print:file=-"
	if vf != "" {
		f = vf + "," + f
	}
	args := append([]string{"-v", "error"}, pre...)
	args = append(args, "-i", p, "-vf", f, "-f", "null", "-")
	b, err := exec.Command("ffmpeg", args...).Output()
	if err != nil {
		return 0, 0
	}
	mean := func(re *regexp.Regexp) float64 {
		ms := re.FindAllStringSubmatch(string(b), -1)
		t := 0.0
		for _, m := range ms {
			v, _ := strconv.ParseFloat(m[1], 64)
			t += v
		}
		if len(ms) == 0 {
			return 0
		}
		return t / float64(len(ms))
	}
	return mean(satRe), mean(highRe) - mean(lowRe)
}

// --- SRT ---------------------------------------------------------------------------------------

type cue struct {
	a, b float64
	text string
}

var srtTime = regexp.MustCompile(`(\d+):(\d{2}):(\d{2})[,.](\d{3})\s*-->\s*(\d+):(\d{2}):(\d{2})[,.](\d{3})`)

func parseSRT(s string) []cue {
	var cs []cue
	for _, blk := range strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n\n") {
		ls := strings.Split(strings.TrimSpace(blk), "\n")
		for i, l := range ls {
			m := srtTime.FindStringSubmatch(l)
			if m == nil {
				continue
			}
			t := func(h, mi, se, ms string) float64 {
				a, _ := strconv.Atoi(h)
				b, _ := strconv.Atoi(mi)
				c, _ := strconv.Atoi(se)
				d, _ := strconv.Atoi(ms)
				return float64(a*3600+b*60+c) + float64(d)/1000
			}
			cs = append(cs, cue{t(m[1], m[2], m[3], m[4]), t(m[5], m[6], m[7], m[8]), strings.Join(ls[i+1:], " ")})
			break
		}
	}
	return cs
}

// srtCheck: the file's cues are in order, inside the video, carry the words the output says
// (in order), and show each word while it is being said (±0.3 s).
func srtCheck(path, out string, dur float64, add func(string, bool, string, ...any)) {
	b, err := os.ReadFile(path)
	if err != nil {
		add("srt", false, "%s not found", filepath.Base(path))
		return
	}
	cs := parseSRT(string(b))
	if len(cs) == 0 {
		add("srt", false, "no cues in %s", filepath.Base(path))
		return
	}
	ordered := true
	for i, c := range cs {
		if c.b <= c.a || c.b > dur+0.05 || (i > 0 && c.a < cs[i-1].b-0.001) {
			ordered = false
		}
	}
	add("srt: timing", ordered, "%d cues, in order, no overlaps, inside the %.3f s video", len(cs), dur)
	var cueWords []string
	for _, c := range cs {
		cueWords = append(cueWords, tokens(c.text)...)
	}
	ws := transcribeTimed(out)
	var spoken []string
	for _, w := range ws {
		spoken = append(spoken, w.w)
	}
	l := lcs(spoken, cueWords)
	ok := len(spoken) > 0 && float64(l) >= 0.85*float64(len(spoken)) && float64(l) >= 0.85*float64(len(cueWords))
	add("srt: words", ok, "%d of %d spoken words in the file in order (file has %d)", l, len(spoken), len(cueWords))
	hit := 0
	for _, w := range ws {
		mid := (w.a + w.b) / 2
		for _, c := range cs {
			if mid >= c.a-0.3 && mid <= c.b+0.3 && has(tokens(c.text), w.w) {
				hit++
				break
			}
		}
	}
	add("srt: in sync", len(ws) > 0 && float64(hit) >= 0.8*float64(len(ws)), "%d of %d spoken words are on screen (±0.3 s) while said", hit, len(ws))
}

type tword struct {
	w    string
	a, b float64
}

// transcribeTimed: one whisper segment per word (-ml 1) with its times.
func transcribeTimed(p string) []tword {
	tmp, _ := os.MkdirTemp("", "grade")
	defer os.RemoveAll(tmp)
	wav := filepath.Join(tmp, "a.wav")
	if err := exec.Command("ffmpeg", "-v", "error", "-i", p, "-vn", "-ac", "1", "-ar", "16000", wav).Run(); err != nil {
		return nil
	}
	model := os.Getenv("GRADE_WHISPER_MODEL")
	if model == "" {
		home, _ := os.UserHomeDir()
		model = filepath.Join(home, ".cache/cutroom/models/ggml-base.en.bin")
	}
	if err := exec.Command("whisper-cli", "-m", model, "-f", wav, "-oj", "-ml", "1", "-sow", "-of", filepath.Join(tmp, "t"), "-np").Run(); err != nil {
		return nil
	}
	b, _ := os.ReadFile(filepath.Join(tmp, "t.json"))
	var d struct {
		Transcription []struct {
			Text    string
			Offsets struct{ From, To int } `json:"offsets"`
		} `json:"transcription"`
	}
	json.Unmarshal(b, &d)
	var ws []tword
	for _, s := range d.Transcription {
		for _, t := range tokens(s.Text) {
			ws = append(ws, tword{t, float64(s.Offsets.From) / 1000, float64(s.Offsets.To) / 1000})
		}
	}
	return ws
}

func lcs(a, b []string) int {
	dp := make([][]int, len(a)+1)
	for i := range dp {
		dp[i] = make([]int, len(b)+1)
	}
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else {
				dp[i][j] = max(dp[i+1][j], dp[i][j+1])
			}
		}
	}
	return dp[0][0]
}

// verticalRef finds where the agent cropped the 9:16 frame from the reference: it tries 21 anchors
// on three output frames and returns the reference cropped at the best one.
func verticalRef(ref, out string) ([][]byte, float64) {
	o := greyFrames(out, "fps=30")
	if len(o) < 10 {
		return nil, 0
	}
	probe := []int{len(o) / 5, len(o) / 2, 4 * len(o) / 5}
	best, bestA := math.Inf(1), 0.0
	var bestFr [][]byte
	for k := 0; k <= 20; k++ {
		a := float64(k) / 20
		fr := greyFrames(ref, fmt.Sprintf("crop=ih*9/16:ih:'clip(%.2f*iw-ow/2,0,iw-ow)':0", a))
		t := 0.0
		for _, i := range probe {
			e := math.Inf(1)
			for _, rf := range fr {
				e = math.Min(e, mse(o[i], rf))
			}
			t += e
		}
		if t < best {
			best, bestA, bestFr = t, a, fr
		}
	}
	return bestFr, bestA
}

// phraseSpan finds a phrase in timed words (fuzzy like count): its output start and end.
func phraseSpan(ws []tword, phrase string) (float64, float64, bool) {
	p := tokens(phrase)
	for i := range ws {
		k, j := 0, i
		for ; j < len(ws) && j < i+len(p)+3 && k < len(p); j++ {
			if ws[j].w == p[k] {
				k++
			}
		}
		if k >= int(math.Ceil(0.8*float64(len(p)))) && ws[i].w == p[0] {
			return ws[i].a, ws[j-1].b, true
		}
	}
	return 0, 0, false
}

// bandRMS is the RMS level (dB) of the output's audio in a band "centre:width" (Hz).
func bandRMS(p, band string) float64 {
	parts := strings.Split(band, ":")
	b, _ := exec.Command("ffmpeg", "-v", "info", "-i", p, "-vn", "-af", fmt.Sprintf("bandpass=f=%s:width_type=h:w=%s,astats=metadata=0", parts[0], parts[1]), "-f", "null", "-").CombinedOutput()
	m := regexp.MustCompile(`RMS level dB: (-?[0-9.]+|-inf)`).FindAllStringSubmatch(string(b), -1)
	if len(m) == 0 || m[len(m)-1][1] == "-inf" {
		return -200
	}
	v, _ := strconv.ParseFloat(m[len(m)-1][1], 64)
	return v
}

// colorFrames samples 12 frames over the middle 80% of p (270 px wide) and counts, per frame, pixels
// within a small distance of hex; a frame "shows" the colour with ≥ 40 of them.
func colorFrames(p, hex string) (n, hit, best int) {
	var r, g, b int
	if _, err := fmt.Sscanf(strings.TrimPrefix(hex, "#"), "%02x%02x%02x", &r, &g, &b); err != nil {
		return 0, 0, 0
	}
	d := streamDur(p, "v:0")
	if d <= 0 {
		return 0, 0, 0
	}
	const W, H = 270, 480
	for i := range 12 {
		t := d * (0.1 + 0.8*float64(i)/11)
		px, err := exec.Command("ffmpeg", "-v", "error", "-ss", fmt.Sprintf("%.3f", t), "-i", p, "-frames:v", "1",
			"-vf", fmt.Sprintf("scale=%d:%d", W, H), "-pix_fmt", "rgb24", "-f", "rawvideo", "-").Output()
		if err != nil || len(px) < W*H*3 {
			continue
		}
		n++
		c := 0
		for k := 0; k+2 < len(px); k += 3 {
			dr, dg, db := int(px[k])-r, int(px[k+1])-g, int(px[k+2])-b
			if dr*dr+dg*dg+db*db < 40*40 {
				c++
			}
		}
		best = max(best, c)
		if c >= 40 {
			hit++
		}
	}
	return
}
