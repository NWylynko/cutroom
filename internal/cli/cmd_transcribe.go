package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/cutroom/cutroom/internal/captions"
	"github.com/cutroom/cutroom/internal/deps"
	"github.com/cutroom/cutroom/internal/media"
	"github.com/cutroom/cutroom/internal/project"
	"github.com/cutroom/cutroom/internal/speech"
	"github.com/cutroom/cutroom/internal/timecode"
	"github.com/cutroom/cutroom/internal/timeline"
)

func init() {
	Register(&Command{
		Name:     "transcribe",
		Args:     "[source-id…]",
		Category: "perceive",
		Summary:  "Word-level transcript of each source with audio; flags retakes, unworded sounds and doubtful timing",
		Help: `With no ids: every source that has audio. Results are cached; re-running is instant unless the
file, model or --passes changed.
Word times are measured against the audio's actual loudness, so pauses between words are reliable.
By default two passes run (normal + one prompted to keep stumbles) and their differences become
suspects; --passes 1 halves the time.
Examples:
  cutroom transcribe
  cutroom transcribe talk --passes 1`,
		Flags: []Flag{
			{Name: "passes", Kind: "int", Default: "2", Help: "1 or 2 recognition passes"},
			{Name: "model", Kind: "string", Help: "whisper.cpp model file (default: best found by `cutroom doctor`)"},
			{Name: "force", Kind: "bool", Help: "Ignore the cache"},
		},
		Run: runTranscribe,
	})
	Register(&Command{
		Name:     "transcript",
		Args:     "[source-id]",
		MaxArgs:  1,
		Category: "perceive",
		Summary:  "Read a transcript: phrases with times and pauses, words, search, suspects",
		Help: `Default view: one line per phrase (split at pauses ≥ --gap), its time range, and the pause after it.
Markers in [brackets] point at suspects; --suspects explains them.
Times are source times (seconds or m:ss.mmm), usable with every other command.
Examples:
  cutroom transcript
  cutroom transcript --search "first thing you need"
  cutroom transcript --suspects
  cutroom transcript --words --from 0:06 --to 0:12
  cutroom transcript --retakes                 every repeated take, which one the edit keeps
  cutroom transcript --output                  the edit's words in output time ("|" marks a cut)
  cutroom transcript --range 0:06-0:12         same as --from 0:06 --to 0:12
  cutroom transcript --file out/final.mp4     read the render's own words (to confirm sentences are whole)`,
		Flags: []Flag{
			{Name: "words", Kind: "bool", Help: "One line per word with start/end"},
			{Name: "search", Kind: "string", Help: "Find a phrase (case/punctuation/fillers ignored); prints every occurrence"},
			{Name: "suspects", Kind: "bool", Help: "List suspects: retakes, unworded sounds, long words, pass differences"},
			{Name: "retakes", Kind: "bool", Help: "Only repeats: each take, and whether the edit keeps it"},
			{Name: "output", Kind: "bool", Help: "The edit's words in output time, through the timeline (no render needed)"},
			{Name: "phrases", Kind: "bool", Help: "One line per phrase (the default view)"},
			{Name: "range", Kind: "string", Help: "Range start-end (e.g. 6-12 or 0:06-0:12); same as --from/--to"},
			{Name: "from", Kind: "string", Help: "Start of range (time)"},
			{Name: "to", Kind: "string", Help: "End of range (time)"},
			{Name: "gap", Kind: "float", Default: "0.5", Help: "Pause length (s) that splits phrases"},
			{Name: "all", Kind: "bool", Help: "Don't truncate long output"},
			{Name: "file", Kind: "string", Help: "Transcribe and read any file, e.g. your render (cached); times are file times"},
		},
		Run: runTranscript,
	})
}

const maxLines = 40

func runTranscribe(c *Ctx) (*Result, error) {
	p, err := needProject(c.Dir)
	if err != nil {
		return nil, err
	}
	passes := c.Int("passes")
	if passes != 1 && passes != 2 {
		return nil, Errf(ExitUsage, "bad-flag-value", "use --passes 1 or --passes 2", "--passes must be 1 or 2")
	}
	model := c.Str("model")
	if model == "" {
		model, _ = deps.FindModel()
	}
	if model == "" {
		m := deps.Model()
		return nil, Errf(ExitDep, "no-model", m.Fix, "no whisper model found")
	}
	if _, err := os.Stat(model); err != nil {
		return nil, Errf(ExitUsage, "model-not-found", "pass a ggml model file path, or omit --model", "model %s not found", model)
	}
	ids := c.Args
	if len(ids) == 0 {
		for _, id := range p.IDs() {
			if f := p.Sources[id].Facts; f != nil && f.Audio != nil {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil, Errf(ExitInvalid, "no-audio", "add a source with audio: `cutroom init <file>`", "no source has audio")
		}
	}
	var lines []string
	type sum struct {
		ID       string  `json:"id"`
		Words    int     `json:"words"`
		Phrases  int     `json:"phrases"`
		Suspects int     `json:"suspects"`
		Cached   bool    `json:"cached"`
		Seconds  float64 `json:"seconds"`
		File     string  `json:"file"`
	}
	var sums []sum
	for _, id := range ids {
		s, err := sourceByID(p, id)
		if err != nil {
			return nil, err
		}
		if s.Facts.Audio == nil {
			return nil, Errf(ExitUsage, "no-audio", "pick a source with audio (`cutroom inspect`)", "source %q has no audio", id)
		}
		if err := notChanged(p, s); err != nil {
			return nil, err
		}
		t0 := time.Now()
		key := speech.Key(s.Size, s.ModTime, model, passes)
		t, cached, err := speech.Transcribe(id, p.Abs(s), s.Facts.Audio.Start, s.Facts.Duration, key,
			speech.Options{Model: model, Passes: passes, CacheDir: cacheDir(p), Force: c.Bool("force")})
		if err != nil {
			return nil, Errf(ExitExternal, "transcribe-failed", "run `cutroom doctor`; if it passes, retry with --passes 1", "%s: %v", id, err)
		}
		el := time.Since(t0).Seconds()
		ph := speech.Phrases(t.Words, 0.5)
		note := fmt.Sprintf("%.1f s", el)
		if cached {
			note = "cached"
		}
		lines = append(lines, fmt.Sprintf("%-10s %d words, %d phrases, %d suspects  (%s, %d pass(es), %s)", id, len(t.Words), len(ph), len(t.Suspects), t.Model, t.Passes, note))
		sums = append(sums, sum{id, len(t.Words), len(ph), len(t.Suspects), cached, el, speech.TranscriptPath(cacheDir(p), id)})
	}
	if strings.HasPrefix(speech.ModelName(model), "tiny") || strings.HasPrefix(speech.ModelName(model), "base") {
		lines = append(lines, "note: small model ("+speech.ModelName(model)+"); expect misheard words. `cutroom doctor` shows how to get a larger one.")
	}
	return &Result{Text: strings.Join(lines, "\n"), Data: sums, Next: []string{"cutroom transcript " + ids[0], "cutroom transcript " + ids[0] + " --suspects"}}, nil
}

// transcribeFile transcribes any media file (1 pass), cached by path, size and mtime.
func transcribeFile(p *project.Project, path string) (*speech.Transcript, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, Errf(ExitUsage, "file-not-found", "check the path", "%s not found", path)
	}
	f, err := media.Probe(path, media.Options{})
	if err != nil || f.Audio == nil {
		return nil, Errf(ExitUsage, "no-audio", "pick a file with sound", "%s has no audio", path)
	}
	model, _ := deps.FindModel()
	if model == "" {
		return nil, Errf(ExitDep, "no-model", deps.Model().Fix, "no whisper model found")
	}
	abs, _ := filepath.Abs(path)
	id := "file-" + speech.Key(st.Size(), st.ModTime(), abs, 1)
	t, _, err := speech.Transcribe(id, path, f.Audio.Start, f.Duration, speech.Key(st.Size(), st.ModTime(), model, 1),
		speech.Options{Model: model, Passes: 1, CacheDir: cacheDir(p)})
	if err != nil {
		return nil, Errf(ExitExternal, "transcribe-failed", "run `cutroom doctor`", "%v", err)
	}
	t.Source = relToCwd(path)
	return t, nil
}

func cacheDir(p *project.Project) string { return filepath.Join(p.Dir, ".cutroom") }

func sourceByID(p *project.Project, id string) (*project.Source, error) {
	if s := p.Sources[id]; s != nil {
		return s, nil
	}
	fix := "source ids: " + strings.Join(p.IDs(), ", ")
	if s := suggest(id, p.IDs()); s != "" {
		fix = fmt.Sprintf("did you mean %q? (%s)", s, fix)
	}
	if _, err := os.Stat(id); err == nil {
		if src := p.FindByPath(id); src != nil {
			return src, nil
		}
		if _, err := os.Stat(id + ".cutroom.json"); err == nil {
			// A render (it has a manifest): adding it as a source would be wrong. Read it as a file.
			fix = fmt.Sprintf("%s is a render, not a source; read it as a file: `cutroom transcript --file %s`, `cutroom look %s <time>`, `cutroom check %s`", id, id, id, id)
		} else {
			fix = fmt.Sprintf("to edit it, add it as a source: `cutroom init %s`; to just read it: `cutroom transcript --file %s`", id, id)
		}
	}
	return nil, Errf(ExitUsage, "unknown-source", fix, "no source %q", id)
}

// loadTranscript finds the transcript for id (or the default source) and errors helpfully.
func loadTranscript(p *project.Project, id string) (*speech.Transcript, error) {
	if id == "" {
		var have []string
		for _, sid := range p.IDs() {
			if _, err := speech.Load(cacheDir(p), sid); err == nil {
				have = append(have, sid)
			}
		}
		if len(have) == 0 {
			return nil, Errf(ExitInvalid, "no-transcript", "run `cutroom transcribe`", "nothing has been transcribed yet")
		}
		id = defaultTranscript(p, have)
	}
	s, err := sourceByID(p, id)
	if err != nil {
		return nil, err
	}
	t, err := speech.Load(cacheDir(p), s.ID)
	if err != nil {
		return nil, Errf(ExitInvalid, "no-transcript", "run `cutroom transcribe "+s.ID+"`", "source %q has not been transcribed", s.ID)
	}
	if stale, _ := p.Stale(s); stale {
		return nil, Errf(ExitInvalid, "stale-transcript", "run `cutroom init` then `cutroom transcribe "+s.ID+"`", "source %q changed since it was transcribed", s.ID)
	}
	return t, nil
}

// defaultTranscript picks the transcript `transcript` shows with no source named: the edit's main
// video, else the longest video, else the longest source (a short pickup sorting first must not
// hide the main take, and a recorder's clock must not stand in for the camera's).
func defaultTranscript(p *project.Project, have []string) string {
	if tl, err := timeline.Load(p.Dir); err == nil && slices.Contains(have, tl.Video) {
		return tl.Video
	}
	// the longest video (a recorder's clock isn't the one edits use), else the longest source
	best, bestD := have[0], -1.0
	for _, video := range []bool{true, false} {
		for _, id := range have {
			f := p.Sources[id].Facts
			if f != nil && (f.Video != nil) == video && f.Duration > bestD {
				best, bestD = id, f.Duration
			}
		}
		if bestD >= 0 {
			return best
		}
	}
	return best
}

// otherTranscripts lists the other transcribed sources, for a header line.
func otherTranscripts(p *project.Project, shown string) []string {
	var o []string
	for _, id := range p.IDs() {
		if _, err := speech.Load(cacheDir(p), id); err == nil && id != shown {
			o = append(o, id)
		}
	}
	return o
}

func runTranscript(c *Ctx) (*Result, error) {
	p, err := needProject(c.Dir)
	if err != nil {
		return nil, err
	}
	id := ""
	if len(c.Args) == 1 {
		id = c.Args[0]
	}
	// `transcript out/final.mp4` means `transcript --file out/final.mp4` when the path isn't a source.
	if !c.Has("file") && id != "" && p.Sources[id] == nil && p.FindByPath(id) == nil {
		if takeFileArg(c); c.Has("file") {
			id = ""
		}
	}
	if c.Bool("output") {
		return transcriptOutput(c)
	}
	var t *speech.Transcript
	if c.Has("file") && !c.Has("project") {
		projectFromFile(c, c.Str("file"))
		if p2, err := needProject(c.Dir); err == nil {
			p = p2
		}
	}
	if c.Has("file") {
		if t, err = transcribeFile(p, c.Str("file")); err != nil {
			return nil, err
		}
	} else if t, err = loadTranscript(p, id); err != nil {
		return nil, err
	}
	lo, hi := 0.0, 1e18
	if c.Has("range") {
		a, b, err := parseRange(c.Str("range"), timecode.Rate{Num: 30, Den: 1})
		if err != nil {
			return nil, Errf(ExitUsage, "bad-range", "--range 6-12 or --range 0:06-0:12", "%v", err)
		}
		c.vals["from"], c.vals["to"] = []string{fmt.Sprint(a)}, []string{fmt.Sprint(b)}
	}
	if c.Has("from") {
		v, err := timecode.Parse(c.Str("from"), timecode.Rate{Num: 30, Den: 1})
		if err != nil {
			return nil, Errf(ExitUsage, "bad-time", "times look like 12.5, 1:02.5 or f372", "--from: %v", err)
		}
		lo = v.Float()
	}
	if c.Has("to") {
		v, err := timecode.Parse(c.Str("to"), timecode.Rate{Num: 30, Den: 1})
		if err != nil {
			return nil, Errf(ExitUsage, "bad-time", "times look like 12.5, 1:02.5 or f372", "--to: %v", err)
		}
		hi = v.Float()
	}
	inRange := func(a, b float64) bool { return b > lo && a < hi }
	if lo >= t.Duration || (c.Has("to") && hi <= 0) {
		return nil, Errf(ExitUsage, "out-of-range", fmt.Sprintf("%s is %.3f s long; --from/--to are times in seconds (or m:ss), not word numbers", t.Source, t.Duration),
			"--from %s is past the end of %s", c.Str("from"), t.Source)
	}
	tl, _ := timeline.Load(p.Dir)
	cutNote := func(a, b float64) string {
		if tl == nil {
			return ""
		}
		used := false
		for _, g := range tl.Segments {
			used = used || g.Source == t.Source
		}
		if !used { // audio-only (a recorder) or not in the edit: "already cut" would be wrong
			return ""
		}
		if !tl.Kept(t.Source, a+0.01) && !tl.Kept(t.Source, (a+b)/2) && !tl.Kept(t.Source, b-0.01) {
			return "  (already cut)"
		}
		return ""
	}
	var lines []string
	trunc := func(note string) {
		if !c.Bool("all") && len(lines) > maxLines {
			n := len(lines) - maxLines
			lines = append(lines[:maxLines], fmt.Sprintf("… %d more lines (%s, or --all)", n, note))
		}
	}
	head := fmt.Sprintf("%s — %d words, %s  (%s, %d pass(es))", t.Source, len(t.Words), clock(t.Duration), t.Model, t.Passes)
	if id == "" && !c.Has("file") {
		if o := otherTranscripts(p, t.Source); len(o) > 0 {
			head += fmt.Sprintf("\nalso transcribed: %s — `cutroom transcript %s` (this is %s; times are each source's own)", strings.Join(o, ", "), o[0], t.Source)
		}
	}

	switch {
	case c.Has("search"):
		q := c.Str("search")
		ms := speech.Search(t.Words, q)
		var hits []map[string]any
		for _, m := range ms {
			if !inRange(m.Start, m.End) {
				continue
			}
			ctxA, ctxB := max(0, m.First-6), min(len(t.Words)-1, m.Last+6)
			lines = append(lines, fmt.Sprintf("%s–%s  …%s [%s] %s…", clock(m.Start), clock(m.End),
				speech.Text(t.Words, ctxA, m.First-1), m.Text, speech.Text(t.Words, m.Last+1, ctxB)))
			hits = append(hits, map[string]any{"start": m.Start, "end": m.End, "text": m.Text, "first_word": m.First, "last_word": m.Last})
		}
		if len(hits) == 0 {
			return &Result{Text: fmt.Sprintf("no match for %q in %s", q, t.Source), Data: map[string]any{"matches": []any{}},
				Next: []string{"cutroom transcript " + t.Source + "   (read it; the recogniser may have heard different words)"}}, nil
		}
		lines = append([]string{fmt.Sprintf("%d match(es) for %q:", len(hits), q)}, lines...)
		trunc("narrow with --from/--to")
		return &Result{Text: strings.Join(lines, "\n"), Data: map[string]any{"matches": hits}}, nil

	case c.Bool("suspects"):
		var sel []speech.Suspect
		breaths := 0
		for _, s := range t.Suspects {
			if s.Kind == "breath" && !c.Bool("all") {
				if inRange(s.Start, s.End) {
					breaths++
				}
				continue
			}
			if inRange(s.Start, s.End) {
				sel = append(sel, s)
				note := cutNote(s.Start, s.End)
				if s.Kind == "repeat" && tl != nil {
					kept := 0
					for _, at := range s.Related {
						if tl.Kept(t.Source, at+0.05) {
							kept++
						}
					}
					if kept <= 1 {
						note = "  (resolved: only one take is kept)"
					}
				}
				lines = append(lines, fmt.Sprintf("%s–%s  %-14s %s%s", clock(s.Start), clock(s.End), s.Kind, s.Detail, note))
			}
		}
		if len(sel) == 0 {
			lines = append(lines, "no suspects")
		}
		if breaths > 0 {
			lines = append(lines, fmt.Sprintf("(%d breath(s) not listed: quiet, normal; --all shows them)", breaths))
		}
		trunc("narrow with --from/--to")
		return &Result{Text: head + "\n" + strings.Join(lines, "\n"), Data: map[string]any{"suspects": nzs(sel)},
			Next: []string{"cutroom transcript " + t.Source + " --words --from <start> --to <end>   (look closer at one)"}}, nil

	case c.Bool("retakes"):
		type take struct {
			At   float64 `json:"at"`
			Kept *bool   `json:"kept,omitempty"`
		}
		type retake struct {
			Start  float64 `json:"start"`
			End    float64 `json:"end"`
			Text   string  `json:"text"`
			Takes  []take  `json:"takes"`
			Status string  `json:"status"`
		}
		var rs []retake
		open := 0
		for _, s := range t.Suspects {
			if s.Kind != "repeat" || !inRange(s.Start, s.End) {
				continue
			}
			r := retake{Start: s.Start, End: s.End, Text: s.Text}
			ats := append([]float64{}, s.Related...)
			if len(ats) == 0 {
				ats = []float64{s.Start}
			}
			kept := 0
			var marks []string
			for _, at := range ats {
				tk := take{At: at}
				m := clock(at)
				if tl != nil {
					k := tl.Kept(t.Source, at+0.05)
					tk.Kept = &k
					if k {
						kept++
						m += " kept"
					} else {
						m += " cut"
					}
				}
				r.Takes = append(r.Takes, tk)
				marks = append(marks, m)
			}
			switch {
			case tl == nil:
				r.Status = "no edit yet"
			case kept <= 1:
				r.Status = "resolved"
			default:
				r.Status = fmt.Sprintf("open: %d takes kept", kept)
				open++
			}
			rs = append(rs, r)
			lines = append(lines, fmt.Sprintf("%s–%s  %-20s takes: %s  — %s", clock(s.Start), clock(s.End), r.Status, strings.Join(marks, ", "), s.Detail))
		}
		if len(rs) == 0 {
			lines = append(lines, "no retakes (no repeated phrase found)")
		}
		trunc("narrow with --from/--to")
		next := []string{"cutroom transcript " + t.Source + " --words --from <start> --to <end>   (compare the takes)"}
		if open > 0 {
			next = append([]string{"cutroom edit cut --text \"<the take you drop>\" --near <its time>   (keep the best take, cut the others)"}, next...)
		}
		if rs == nil {
			rs = []retake{}
		}
		return &Result{Text: head + "\n" + strings.Join(lines, "\n"), Data: map[string]any{"retakes": rs, "open": open}, Next: next}, nil

	case c.Bool("words"):
		var sel []speech.Word
		for _, w := range t.Words {
			if inRange(w.Start, w.End) {
				sel = append(sel, w)
				flags := ""
				if len(w.Flags) > 0 {
					flags = "  [" + strings.Join(w.Flags, ",") + "]"
				}
				lines = append(lines, fmt.Sprintf("%5d  %s–%s  %s%s", w.I, clock(w.Start), clock(w.End), w.Text, flags))
			}
		}
		trunc("narrow with --from/--to")
		return &Result{Text: head + "\n" + strings.Join(lines, "\n"), Data: map[string]any{"words": sel}}, nil
	}

	// default: phrases
	ph := speech.Phrases(t.Words, c.Float("gap"))
	if c.Has("file") { // a render: tight pauses (~0.2 s), so a line per sentence or per pause ≥ 0.15 s
		gap := 0.15
		if c.Has("gap") {
			gap = c.Float("gap")
		}
		ph = speech.Sentences(t.Words, gap)
	}
	type phraseOut struct {
		Start      float64  `json:"start"`
		End        float64  `json:"end"`
		Text       string   `json:"text"`
		PauseAfter float64  `json:"pause_after"`
		Suspects   []string `json:"suspects,omitempty"`
	}
	var out []phraseOut
	for _, x := range ph {
		if !inRange(x.Start, x.End) {
			continue
		}
		var marks []string
		for _, s := range t.Suspects {
			if s.Start < x.End && s.End > x.Start && !contains(marks, s.Kind) {
				marks = append(marks, s.Kind)
			}
		}
		text := speech.Text(t.Words, x.First, x.Last)
		m := ""
		if len(marks) > 0 {
			m = "   [" + strings.Join(marks, ", ") + "]"
		}
		lines = append(lines, fmt.Sprintf("%s–%s  %s%s", clock(x.Start), clock(x.End), text, m))
		if x.PauseAfter > 0 {
			gap := fmt.Sprintf("             · %.2f s pause", x.PauseAfter)
			for _, s := range t.Suspects {
				if s.Kind == "sound" && s.Start >= x.End && s.End <= x.End+x.PauseAfter+0.05 {
					gap += fmt.Sprintf(" (contains a sound at %s)", clock(s.Start))
				}
			}
			lines = append(lines, gap)
		}
		out = append(out, phraseOut{x.Start, x.End, text, x.PauseAfter, marks})
	}
	trunc("narrow with --from/--to")
	ref := t.Source
	if c.Has("file") {
		ref = "--file " + c.Str("file")
	}
	foot := fmt.Sprintf("%d suspect(s): `cutroom transcript %s --suspects`", len(t.Suspects), ref)
	return &Result{Text: head + "\n" + strings.Join(lines, "\n") + "\n" + foot, Data: map[string]any{"source": t.Source, "phrases": out, "suspects": len(t.Suspects)},
		Next: []string{"cutroom transcript " + ref + " --suspects"}}, nil
}

func nzs(s []speech.Suspect) []speech.Suspect {
	if s == nil {
		return []speech.Suspect{}
	}
	return s
}

// clock formats seconds as m:ss.mmm.
func clock(s float64) string {
	return timecode.FromFloat(s).Clock()
}

// notChanged refuses a source whose file changed since `init` inspected it: its facts, cached audio
// and transcript would describe the old file.
func notChanged(p *project.Project, s *project.Source) error {
	if stale, err := p.Stale(s); stale {
		if err != nil {
			return Errf(ExitInvalid, "source-missing", "put the file back, or `cutroom init --remove "+s.ID+"`", "source %q (%s) is missing", s.ID, s.Path)
		}
		return Errf(ExitInvalid, "stale-source", "run `cutroom init` to re-inspect it", "source %q changed on disk since it was inspected", s.ID)
	}
	return nil
}

// transcriptOutput reads the edit as it plays: the kept words in output time, a line per sentence
// (or pause ≥ --gap, default 0.15 s), "|" where a cut joins two pieces. No render is needed.
func transcriptOutput(c *Ctx) (*Result, error) {
	e, err := loadEditEnv(c, false)
	if err != nil {
		return nil, err
	}
	ows, missing := captions.OutputWords(e.tl, e.tr)
	ws := make([]speech.Word, len(ows))
	for i, w := range ows {
		text := w.Text
		if w.AfterCut {
			text = "| " + text
		}
		ws[i] = speech.Word{I: i, Text: text, Start: w.Start, End: w.End}
	}
	gap := 0.15
	if c.Has("gap") {
		gap = c.Float("gap")
	}
	lo, hi := 0.0, 1e18
	if c.Has("range") {
		if lo, hi, err = parseRange(c.Str("range"), timecode.Rate{Num: 30, Den: 1}); err != nil {
			return nil, Errf(ExitUsage, "bad-range", "--range 6-12 or --range 0:06-0:12 (output time)", "%v", err)
		}
	}
	for k, v := range map[string]*float64{"from": &lo, "to": &hi} {
		if c.Has(k) {
			t, err := timecode.Parse(c.Str(k), timecode.Rate{Num: 30, Den: 1})
			if err != nil {
				return nil, Errf(ExitUsage, "bad-time", "times look like 12.5, 1:02.5 or f372", "--%s: %v", k, err)
			}
			*v = t.Float()
		}
	}
	type phraseOut struct {
		Start float64 `json:"start"`
		End   float64 `json:"end"`
		Text  string  `json:"text"`
	}
	var out []phraseOut
	var lines []string
	for _, x := range speech.Sentences(ws, gap) {
		if x.End <= lo || x.Start >= hi {
			continue
		}
		text := speech.Text(ws, x.First, x.Last)
		lines = append(lines, fmt.Sprintf("%s–%s  %s", clock(x.Start), clock(x.End), text))
		out = append(out, phraseOut{x.Start, x.End, text})
	}
	if !c.Bool("all") && len(lines) > maxLines {
		n := len(lines) - maxLines
		lines = append(lines[:maxLines], fmt.Sprintf("… %d more lines (narrow with --from/--to, or --all)", n))
	}
	head := fmt.Sprintf("the edit in output time — %d words, %s, %d segment(s); \"|\" marks a cut", len(ws), clock(e.tl.Duration()), len(e.tl.Segments))
	if len(missing) > 0 {
		head += fmt.Sprintf("\nno transcript for %s: its words are missing here (`cutroom transcribe %s`)", strings.Join(missing, ", "), missing[0])
	}
	if out == nil {
		out = []phraseOut{}
	}
	return &Result{Text: head + "\n" + strings.Join(lines, "\n"), Data: map[string]any{"phrases": out, "duration": e.tl.Duration()},
		Next: []string{"cutroom timeline --joins   (each cut, with the words either side)"}}, nil
}
