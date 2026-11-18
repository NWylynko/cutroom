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
		case c == "upright", c == "frames_from_source", c == "av_sync":
			// computed together below
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
	if has(sp.Checks, "upright") || has(sp.Checks, "frames_from_source") || has(sp.Checks, "av_sync") {
		frameAndSync(out, sp.Checks, add)
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

func frameAndSync(out string, checks []string, add func(string, bool, string, ...any)) {
	ref := filepath.Join(fixtures, "talk.mp4")
	refFrames := greyFrames(ref, "")
	outFrames := greyFrames(out, "fps=30") // resample to a 30 fps grid in output time
	if len(outFrames) == 0 || len(refFrames) == 0 {
		add("frames", false, "could not decode frames")
		return
	}
	type match struct {
		outT, refT, err, flippedErr float64
	}
	var ms []match
	for i := 0; i < len(outFrames); i += 6 { // every 0.2 s
		best, bi := math.Inf(1), 0
		for j, rf := range refFrames {
			if e := mse(outFrames[i], rf); e < best {
				best, bi = e, j
			}
		}
		fl := mse(flip(outFrames[i]), refFrames[bi])
		ms = append(ms, match{float64(i) / refFPS, float64(bi) / refFPS, best, fl})
	}
	// upright: best match error small and much smaller than the flipped comparison
	bad := 0
	for _, m := range ms {
		if !(m.err < 30 && m.err < m.flippedErr) {
			bad++
		}
	}
	if has(checks, "upright") || has(checks, "frames_from_source") {
		add("frames_from_source/upright", bad == 0, "%d/%d sampled frames match an upright source frame (worst err %.1f)", len(ms)-bad, len(ms), worst(ms, func(m match) float64 { return m.err }))
	}
	if !has(checks, "av_sync") {
		return
	}
	outA := pcm(out)
	refA := pcm(ref)
	var lags []float64
	win, maxLag := int(0.5*sr), int(0.25*sr)
	for _, m := range ms {
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
	add("av_sync", mx <= 45, "%d windows, median %+.0f ms, worst %.0f ms (limit 45)", len(lags), med, mx)
}

func worst[T any](xs []T, f func(T) float64) float64 {
	w := 0.0
	for _, x := range xs {
		w = math.Max(w, f(x))
	}
	return w
}

func greyFrames(p, extra string) [][]byte {
	vf := fmt.Sprintf("scale=%d:%d,format=gray", gw, gh)
	if extra != "" {
		vf = extra + "," + vf
	}
	b, err := exec.Command("ffmpeg", "-v", "error", "-i", p, "-vf", vf, "-f", "rawvideo", "-").Output()
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
