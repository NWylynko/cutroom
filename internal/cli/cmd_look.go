package cli

import (
	"fmt"
	"image"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cutroom/cutroom/internal/deps"
	"github.com/cutroom/cutroom/internal/perceive"
	"github.com/cutroom/cutroom/internal/project"
	"github.com/cutroom/cutroom/internal/raster"
	"github.com/cutroom/cutroom/internal/speech"
	"github.com/cutroom/cutroom/internal/timecode"
)

func init() {
	Register(&Command{
		Name:     "look",
		Args:     "[file] [time | start-end]",
		MinArgs:  0,
		MaxArgs:  2,
		Category: "perceive",
		Summary:  "See video: a frame, a labelled contact sheet, every cut (--cuts), the edit (--output) or a render (--file)",
		Help: `A time gives one frame (as a viewer would see it: rotation applied). A range gives a contact
sheet: --count frames evenly spaced, each labelled with its time, frame number and the word being
spoken. Open the PNG it prints (e.g. with your file/image reading tool) — the text output alone does
not show the picture.
Examples:
  cutroom look 12.5
  cutroom look 0:06-0:12 --count 12
  cutroom look 3 --source cam --crop 400,100,600,600     (full-resolution crop, e.g. to judge a face)
  cutroom look --cuts                    every cut: last frame before | first frame after, with the words
  cutroom look --output 0:03-0:06        OUTPUT times: frames of the edit, cuts marked in red
  cutroom look out/final.mp4 3.9        a frame of a rendered file (file time; same as --file)
Times are source times unless you pass --output (edit time) or --file (that file's time).`,
		Flags: []Flag{
			{Name: "source", Kind: "string", Help: "Source id (default: the first video source)"},
			{Name: "count", Kind: "int", Default: "12", Help: "Frames in a contact sheet (1–36)"},
			{Name: "width", Kind: "int", Default: "1280", Help: "Max width of a single frame in pixels"},
			{Name: "crop", Kind: "string", Help: "x,y,w,h in displayed source pixels; shown at full resolution"},
			{Name: "output", Kind: "bool", Help: "Times are output (edit) times, mapped through the timeline"},
			{Name: "cuts", Kind: "bool", Help: "Show every cut: frame before | frame after, 12 cuts per page"},
			{Name: "page", Kind: "int", Default: "1", Help: "Page of --cuts"},
			{Name: "file", Kind: "string", Help: "Look at this video file (e.g. a render); times are file times"},
		},
		Run: runLook,
	})
	Register(&Command{
		Name:     "listen",
		Args:     "[file] <start-end>",
		MinArgs:  1,
		MaxArgs:  2,
		Category: "perceive",
		Summary:  "Hear a stretch of audio (a source, the edit with --output, or a file): spectrogram + loudness + words as a PNG, and what an isolated re-transcription hears",
		Help: `Use it to settle doubts before cutting: a suspect sound, a word the transcript may have missed,
the exact end of a word, whether a pause is clean. Ranges up to 30 s.
The isolated re-transcription often hears words the long pass skipped (it has no context to smooth over).
Red areas in the image are sounds with no transcript words.
Examples:
  cutroom listen 16.5-17.8
  cutroom listen 0:06-0:11 --source mic
  cutroom listen --output 3-5        the RENDERED edit around output 3–5 s, cuts marked (needs a current render)
  cutroom listen out/final.mp4 3-5   (same as --file)`,
		Flags: []Flag{
			{Name: "output", Kind: "bool", Help: "Range is in output time; listens to the current render (out/final.mp4 or preview)"},
			{Name: "file", Kind: "string", Help: "Listen to this file (e.g. a render); range in file time"},
			{Name: "source", Kind: "string", Help: "Source id (default: the first transcribed source, else the first with audio)"},
			{Name: "pad", Kind: "float", Default: "0", Help: "Widen the range by this many seconds on each side"},
		},
		Run: runListen,
	})
}

var rangeRe = regexp.MustCompile(`^\s*([0-9:.f]+s?)\s*(?:-|–|\.\.)\s*([0-9:.f]+s?)\s*$`)

// parseRange accepts "6-12", "0:06-0:12", "6..12", "f180-f360".
func parseRange(s string, r timecode.Rate) (float64, float64, error) {
	m := rangeRe.FindStringSubmatch(s)
	if m == nil {
		return 0, 0, fmt.Errorf("%q is not a range (want start-end, e.g. 6-12 or 0:06-0:12)", s)
	}
	a, err := timecode.Parse(m[1], r)
	if err != nil {
		return 0, 0, err
	}
	b, err := timecode.Parse(m[2], r)
	if err != nil {
		return 0, 0, err
	}
	if b.Cmp(a) <= 0 {
		return 0, 0, fmt.Errorf("range end must be after its start: %s", s)
	}
	return a.Float(), b.Float(), nil
}

func sourceRate(s *project.Source) timecode.Rate {
	if s.Facts != nil && s.Facts.Video != nil {
		if r, err := timecode.ParseRate(s.Facts.Video.FPS); err == nil {
			return r
		}
	}
	return timecode.Rate{Num: 30, Den: 1}
}

func defaultVideo(p *project.Project, id string) (*project.Source, error) {
	if id != "" {
		s, err := sourceByID(p, id)
		if err != nil {
			return nil, err
		}
		if s.Facts.Video == nil {
			return nil, Errf(ExitUsage, "no-video", "pick a video source; `cutroom inspect` lists them", "source %q has no video", id)
		}
		return s, nil
	}
	// the longest video: with several takes, the main recording (ties: first id)
	var best *project.Source
	for _, id := range p.IDs() {
		s := p.Sources[id]
		if s.Facts.Video != nil && (best == nil || s.Facts.Duration > best.Facts.Duration+1e-6) {
			best = s
		}
	}
	if best != nil {
		return best, nil
	}
	return nil, Errf(ExitInvalid, "no-video", "add a video: `cutroom init <file>`", "the project has no video source")
}

func lookDir(p *project.Project) string {
	d := filepath.Join(cacheDir(p), "look")
	os.MkdirAll(d, 0o755)
	return d
}

func relToCwd(path string) string {
	wd, _ := os.Getwd()
	if r, err := filepath.Rel(wd, path); err == nil && !strings.HasPrefix(r, "..") {
		return r
	}
	return path
}

func stamp(t float64) string { return strings.ReplaceAll(strconv.FormatFloat(t, 'f', 3, 64), ".", "_") }

func runLook(c *Ctx) (*Result, error) {
	if err := fileArg(c); err != nil {
		return nil, err
	}
	if c.Bool("cuts") || c.Bool("output") || c.Has("file") {
		return lookEdit(c)
	}
	if len(c.Args) == 0 {
		return nil, Errf(ExitUsage, "bad-args", "give a time or range (`cutroom look 12.5`, `cutroom look 6-12`), or use --cuts", "look needs a time")
	}
	p, err := needProject(c.Dir)
	if err != nil {
		return nil, err
	}
	s, err := defaultVideo(p, c.Str("source"))
	if err != nil {
		return nil, err
	}
	rate := sourceRate(s)
	dur := s.Facts.Duration
	src := p.Abs(s)
	var words []speech.Word
	if t, err := speech.Load(cacheDir(p), s.ID); err == nil {
		words = t.Words
	}
	arg := c.Args[0]
	if strings.ContainsAny(arg, "-–") || strings.Contains(arg, "..") {
		a, b, err := parseRange(arg, rate)
		if err != nil {
			return nil, Errf(ExitUsage, "bad-range", "write a range as start-end, e.g. 6-12 or 0:06-0:12", "%v", err)
		}
		if a >= dur {
			return nil, Errf(ExitUsage, "out-of-range", fmt.Sprintf("%s is %.3f s long", s.ID, dur), "range starts after the end of %s", s.ID)
		}
		b = math.Min(b, dur)
		n := c.Int("count")
		if n < 1 || n > 36 {
			return nil, Errf(ExitUsage, "bad-flag-value", "use --count between 1 and 36; for longer stretches look at several ranges", "--count %d out of range", n)
		}
		img, meta, err := perceive.Sheet(src, a, b, n, rate, words)
		if err != nil {
			return nil, Errf(ExitExternal, "frame-failed", "check the times are inside the source", "%v", err)
		}
		out := filepath.Join(lookDir(p), fmt.Sprintf("%s_%s-%s_x%d.png", s.ID, stamp(a), stamp(b), n))
		if err := raster.SavePNG(out, img); err != nil {
			return nil, err
		}
		step := 0.0
		if n > 1 {
			step = (b - a) / float64(n-1)
		}
		text := fmt.Sprintf("wrote %s\n  contact sheet of %s: %d frames %s–%s (every %.3f s), %dx%d px\n  each frame is labelled with time, frame number and the word being spoken\n  view the image to see it",
			relToCwd(out), s.ID, n, clock(a), clock(b), step, img.Bounds().Dx(), img.Bounds().Dy())
		if s.Facts.Video.VFR {
			text += "\n  note: variable frame rate source — frame numbers are nominal (time × fps), the picture is the frame actually shown at that time"
		}
		return &Result{Text: text, Data: map[string]any{"image": out, "source": s.ID, "frames": meta}}, nil
	}
	tm, err := timecode.Parse(arg, rate)
	if err != nil {
		return nil, Errf(ExitUsage, "bad-time", "times look like 12.5, 1:02.5, f372, or a range 6-12", "%v", err)
	}
	t := tm.Float()
	if t >= dur {
		return nil, Errf(ExitUsage, "out-of-range", fmt.Sprintf("%s is %.3f s long; the last frame is near %.3f", s.ID, dur, dur-1/rate.FPS()), "%s is past the end of %s", clock(t), s.ID)
	}
	var crop *image.Rectangle
	maxW := c.Int("width")
	if c.Has("crop") {
		var x, y, w, h int
		if n, _ := fmt.Sscanf(c.Str("crop"), "%d,%d,%d,%d", &x, &y, &w, &h); n != 4 || w <= 0 || h <= 0 {
			return nil, Errf(ExitUsage, "bad-crop", "write --crop x,y,w,h in pixels, e.g. --crop 400,100,600,600", "bad --crop %q", c.Str("crop"))
		}
		v := s.Facts.Video
		if x < 0 || y < 0 || x+w > v.Width || y+h > v.Height {
			return nil, Errf(ExitUsage, "bad-crop", fmt.Sprintf("the frame is %dx%d", v.Width, v.Height), "crop %s is outside the frame", c.Str("crop"))
		}
		r := image.Rect(x, y, x+w, y+h)
		crop, maxW = &r, 0
	}
	img, err := perceive.Frame(src, t, maxW, crop)
	if err != nil {
		return nil, Errf(ExitExternal, "frame-failed", "check the time is inside the source", "%v", err)
	}
	f := tm.Frame(rate)
	label := fmt.Sprintf("%s %s f%d", s.ID, clock(t), f)
	out := filepath.Join(lookDir(p), fmt.Sprintf("%s_%s.png", s.ID, stamp(t)))
	if crop != nil {
		out = strings.TrimSuffix(out, ".png") + fmt.Sprintf("_crop%d-%d-%d-%d.png", crop.Min.X, crop.Min.Y, crop.Dx(), crop.Dy())
	}
	if err := raster.SavePNG(out, perceive.Annotate(img, label)); err != nil {
		return nil, err
	}
	say := ""
	for _, w := range words {
		if t >= w.Start && t < w.End {
			say = fmt.Sprintf(", during the word %q", w.Text)
		}
	}
	text := fmt.Sprintf("wrote %s\n  frame %d of %s at %s%s, %dx%d px\n  view the image to see it", relToCwd(out), f, s.ID, clock(t), say, img.Bounds().Dx(), img.Bounds().Dy())
	return &Result{Text: text, Data: map[string]any{"image": out, "source": s.ID, "time": t, "frame": f}}, nil
}

func runListen(c *Ctx) (*Result, error) {
	if err := fileArg(c); err != nil {
		return nil, err
	}
	if c.Bool("output") || c.Has("file") {
		return listenFile(c)
	}
	p, err := needProject(c.Dir)
	if err != nil {
		return nil, err
	}
	var s *project.Source
	if id := c.Str("source"); id != "" {
		if s, err = sourceByID(p, id); err != nil {
			return nil, err
		}
	} else {
		for _, id := range p.IDs() {
			if _, err := speech.Load(cacheDir(p), id); err == nil {
				s = p.Sources[id]
				break
			}
		}
		if s == nil {
			for _, id := range p.IDs() {
				if p.Sources[id].Facts.Audio != nil {
					s = p.Sources[id]
					break
				}
			}
		}
	}
	if s == nil || s.Facts.Audio == nil {
		return nil, Errf(ExitInvalid, "no-audio", "add a source with audio: `cutroom init <file>`", "no source with audio")
	}
	a, b, err := parseRange(c.Args[0], sourceRate(s))
	if err != nil {
		return nil, Errf(ExitUsage, "bad-range", "write a range as start-end, e.g. 16.5-17.8 or 0:06-0:11", "%v", err)
	}
	pad := c.Float("pad")
	a, b = math.Max(0, a-pad), math.Min(s.Facts.Duration, b+pad)
	if b-a > 30 {
		return nil, Errf(ExitUsage, "range-too-long", "listen to at most 30 s at a time; split it", "range is %.1f s", b-a)
	}
	if b-a < 0.2 {
		return nil, Errf(ExitUsage, "range-too-short", "use at least 0.2 s (add --pad 1 for context)", "range is %.2f s", b-a)
	}
	var words []speech.Word
	room, peak := math.NaN(), 0.0
	if t, err := speech.Load(cacheDir(p), s.ID); err == nil {
		words, room, peak = t.Words, t.RoomTone, t.SpeechDB
	}
	model, _ := deps.FindModel()
	tmp, err := os.MkdirTemp(cacheDir(p), "listen")
	if err != nil {
		os.MkdirAll(cacheDir(p), 0o755)
		if tmp, err = os.MkdirTemp(cacheDir(p), "listen"); err != nil {
			return nil, err
		}
	}
	defer os.RemoveAll(tmp)
	out := filepath.Join(lookDir(p), fmt.Sprintf("%s_listen_%s-%s.png", s.ID, stamp(a), stamp(b)))
	L, err := perceive.Listen(p.Abs(s), a, b, words, nil, room, peak, model, out, tmp)
	if err != nil {
		return nil, Errf(ExitExternal, "listen-failed", "check the range is inside the source", "%v", err)
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("wrote %s  (view it: spectrogram, loudness, words)", relToCwd(L.Image)))
	lines = append(lines, fmt.Sprintf("%s %s–%s", s.ID, clock(a), clock(b)))
	if words != nil {
		lines = append(lines, fmt.Sprintf("  transcript: %s", orNone(L.Transcript)))
	}
	if model != "" {
		lines = append(lines, fmt.Sprintf("  heard (isolated clip): %s", orNone(L.Heard)))
	}
	lines = append(lines, fmt.Sprintf("  loudness: peak %.0f dB, room tone %.0f dB", L.PeakDB, L.RoomTone))
	var reg []string
	for _, r := range L.Sounds {
		reg = append(reg, fmt.Sprintf("%s–%s", clock(r.Start), clock(r.End)))
	}
	lines = append(lines, "  sound at: "+orNone(strings.Join(reg, ", ")))
	for _, r := range L.Unworded {
		lines = append(lines, fmt.Sprintf("  ! sound with no transcript words %s–%s (peak %.0f dB)", clock(r.Start), clock(r.End), r.PeakDB))
	}
	if words != nil && L.Heard != "" && !sameWords(L.Heard, L.Transcript) {
		lines = append(lines, "  ! the isolated clip heard different words than the transcript; at the clip's edges this is normal (cut-off words)")
	}
	return &Result{Text: strings.Join(lines, "\n"), Data: L}, nil
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(nothing)"
	}
	return s
}

func sameWords(a, b string) bool {
	norm := func(s string) string {
		var o []string
		for _, f := range strings.Fields(s) {
			if n := speech.Norm(f); n != "" {
				o = append(o, n)
			}
		}
		return strings.Join(o, " ")
	}
	return norm(a) == norm(b)
}

// fileArg lets `look out/final.mp4 3.9` mean `look 3.9 --file out/final.mp4`: a positional argument
// that names an existing file (and isn't a time) becomes --file.
func fileArg(c *Ctx) error {
	if !c.Has("file") {
		takeFileArg(c)
	}
	if len(c.Args) > 1 && c.Bool("output") {
		return Errf(ExitUsage, "bad-args", fmt.Sprintf("--output is a switch, the time follows it once: `cutroom %s --output %s`", c.Cmd.Name, c.Args[len(c.Args)-1]),
			"two times given (%s); --output takes no value", strings.Join(c.Args, ", "))
	}
	if len(c.Args) > 1 {
		return Errf(ExitUsage, "bad-args", fmt.Sprintf("one time or range, optionally after a file: `cutroom %s out/final.mp4 3-5`", c.Cmd.Name),
			"%q is not a file and %s takes one time or range", c.Args[0], c.Cmd.Name)
	}
	return nil
}

func takeFileArg(c *Ctx) {
	for i, a := range c.Args {
		if isNumberish(a) || !strings.ContainsAny(a, "./\\") {
			continue
		}
		if st, err := os.Stat(a); err == nil && !st.IsDir() {
			c.vals["file"] = []string{a}
			c.Args = append(append([]string{}, c.Args[:i]...), c.Args[i+1:]...)
			return
		}
	}
}
