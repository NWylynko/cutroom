package cli

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/cutroom/cutroom/internal/project"
	"github.com/cutroom/cutroom/internal/speech"
	"github.com/cutroom/cutroom/internal/timecode"
	"github.com/cutroom/cutroom/internal/timeline"
)

var editOps = []string{"cut", "keep", "tighten", "audio", "undo", "history", "reset"}

func init() {
	Register(&Command{
		Name:     "edit",
		Args:     "<cut|keep|tighten|audio|undo|history|reset> [range|source]",
		MinArgs:  1,
		MaxArgs:  2,
		Category: "edit",
		Summary:  "Change the timeline: cut or keep by quoted words or times, tighten pauses, undo",
		Help: `The timeline starts as the whole first video source (created on the first edit).
Every edit is checked afterwards and problems are printed. Use --dry-run to preview without saving.

  edit cut --text "um, so"                       remove exactly these words (cut lands in the pauses)
  edit cut --from "By the way" --to "keyboard."  remove from the first words through the last words
  edit cut 16.9-17.4                             remove a source time range (seconds or m:ss.mmm)
  edit keep --from "Hi, today" --to "server is ready."   keep only this span (e.g. for a short)
  edit tighten [--min-cut 0.4]                   shorten every pause longer than the pads + min-cut
  edit audio mic                                 take the sound from source "mic" (after ` + "`cutroom sync mic`" + `)
  edit undo | edit history | edit reset [--source id]

Words are matched ignoring case, punctuation and um/uh. --from must match once (add --near <time>
or --occurrence N if it repeats); --to matches the first occurrence after --from.
A cut by words leaves --pad-after (default 0.22 s) after the last kept word and --pad-before
(0.10 s) before the next one, so joins sound natural. Edges you cut explicitly are pinned: tighten
never moves them.`,
		Flags: []Flag{
			{Name: "text", Kind: "string", Help: "Exact words to cut/keep"},
			{Name: "from", Kind: "string", Help: "First words of the span"},
			{Name: "to", Kind: "string", Help: "Last words of the span (first match after --from)"},
			{Name: "near", Kind: "string", Help: "Pick the match closest to this source time when words repeat"},
			{Name: "occurrence", Kind: "int", Help: "Pick the Nth match when words repeat (1-based)"},
			{Name: "source", Kind: "string", Help: "Source id (for ranges, reset, or to restrict word matching)"},
			{Name: "reason", Kind: "string", Help: "Why (stored with the edit; shown in the timeline)"},
			{Name: "min-cut", Kind: "float", Default: "0.4", Help: "tighten: only shorten pauses where at least this much can go"},
			{Name: "dry-run", Kind: "bool", Help: "Show the result without saving"},
		},
		Run: runEdit,
	})
	Register(&Command{
		Name:     "timeline",
		Category: "edit",
		Summary:  "Show the timeline: segments with output times and text, every join, and problems",
		Help: `Default: one line per segment (output time, source range, first and last words) and any problems.
  --joins   every cut point with the words on both sides and the pause across it — read these
            before rendering; a bad join is the most common editing mistake
Examples:
  cutroom timeline
  cutroom timeline --joins`,
		Flags: []Flag{
			{Name: "joins", Kind: "bool", Help: "List every join (cut point)"},
			{Name: "all", Kind: "bool", Help: "Don't truncate"},
		},
		Run: runTimeline,
	})
}

// edit context: project, timeline, transcripts, durations.
type editEnv struct {
	p    *project.Project
	tl   *timeline.Timeline
	tr   timeline.Transcripts
	dur  map[string]float64
	made bool
}

func loadEditEnv(c *Ctx, create bool) (*editEnv, error) {
	p, err := needProject(c.Dir)
	if err != nil {
		return nil, err
	}
	e := &editEnv{p: p, tr: timeline.Transcripts{}, dur: map[string]float64{}}
	for _, id := range p.IDs() {
		s := p.Sources[id]
		e.dur[id] = s.Facts.Duration
		if t, err := speech.Load(cacheDir(p), id); err == nil {
			if stale, _ := p.Stale(s); !stale {
				e.tr[id] = t
			}
		}
	}
	e.tl, err = timeline.Load(p.Dir)
	if errors.Is(err, timeline.ErrNone) {
		if !create {
			return nil, Errf(ExitInvalid, "no-timeline", "make the first edit (e.g. `cutroom edit tighten`), which creates it", "no timeline yet")
		}
		s, err := defaultVideo(p, c.Str("source"))
		if err != nil {
			return nil, err
		}
		e.tl = newTimeline(s)
		e.made = true
	} else if err != nil {
		return nil, Errf(ExitInvalid, "bad-timeline", "fix "+timeline.File+" or run `cutroom edit undo`", "%v", err)
	}
	return e, nil
}

func newTimeline(s *project.Source) *timeline.Timeline {
	v := s.Facts.Video
	audio, grid := "", 0.0
	if s.Facts.Audio != nil {
		audio, grid = s.ID, s.Facts.Audio.Start
	}
	return timeline.New(v.FPS, v.Width, v.Height, s.ID, audio, grid, s.Facts.Duration)
}

func runEdit(c *Ctx) (*Result, error) {
	op := c.Args[0]
	if !contains(editOps, op) {
		fix := "ops: " + strings.Join(editOps, ", ")
		if s := suggest(op, editOps); s != "" {
			fix = fmt.Sprintf("did you mean `cutroom edit %s`? (%s)", s, fix)
		}
		return nil, Errf(ExitUsage, "unknown-op", fix, "no edit op %q", op)
	}
	p, err := needProject(c.Dir)
	if err != nil {
		return nil, err
	}
	switch op {
	case "undo":
		what, err := timeline.Undo(p.Dir, cacheDir(p))
		if err != nil {
			return nil, Errf(ExitInvalid, "nothing-to-undo", "`cutroom edit history` lists edits", "%v", err)
		}
		text := "undid: " + what
		if tl, err := timeline.Load(p.Dir); err == nil {
			text += fmt.Sprintf("\ntimeline: %d segments, %s", len(tl.Segments), secs(tl.Duration()))
		} else {
			text += "\nno timeline now (the undone edit created it)"
		}
		return &Result{Text: text, Data: map[string]string{"undone": what}, Next: []string{"cutroom timeline"}}, nil
	case "history":
		h := timeline.History(cacheDir(p))
		var lines []string
		for _, e := range h {
			lines = append(lines, fmt.Sprintf("%3d  %s  %s", e.N, e.Time, e.Op))
		}
		if len(lines) == 0 {
			lines = []string{"no edits yet"}
		}
		return &Result{Text: strings.Join(lines, "\n"), Data: h}, nil
	}

	e, err := loadEditEnv(c, true)
	if err != nil {
		return nil, err
	}
	var lines []string
	var data map[string]any
	label := "edit " + op
	if e.made {
		verb := "created"
		if c.Bool("dry-run") {
			verb = "would create"
		}
		lines = append(lines, fmt.Sprintf("%s timeline from %s (whole source, %s)", verb, e.tl.Video, secs(e.tl.Duration())))
	}
	before := e.tl.Duration()
	soundsBefore := e.keptSounds()
	var issuesBefore []timeline.Issue // warnings already open before this edit are shown in one line
	if !e.made {
		issuesBefore = e.validate()
	}

	switch op {
	case "reset":
		s, err := defaultVideo(e.p, c.Str("source"))
		if err != nil {
			return nil, err
		}
		e.tl = newTimeline(s)
		lines = append(lines, fmt.Sprintf("timeline reset to the whole of %s", s.ID))
		data = map[string]any{"source": s.ID}
	case "tighten":
		minCut := c.Float("min-cut")
		if len(e.tr) == 0 {
			return nil, Errf(ExitInvalid, "no-transcript", "run `cutroom transcribe` first: tighten needs word timings", "no transcripts")
		}
		r := e.tl.Tighten(e.tr, minCut)
		label = fmt.Sprintf("edit tighten --min-cut %g", minCut)
		lines = append(lines, fmt.Sprintf("tightened: %d pauses shortened, %d segment edges trimmed, %s removed (pauses longer than pads %.2f+%.2f s + --min-cut %g s)",
			r.Pauses, r.Edges, secs(r.Removed), e.tl.PadAfter, e.tl.PadBefore, minCut))
		var ps []string
		for _, pc := range r.Shortened {
			ps = append(ps, fmt.Sprintf("%s %.2f→%.2f s", clock(pc.At), pc.Before, pc.After))
		}
		if len(ps) > 0 {
			if len(ps) > 8 {
				ps = append(ps[:8], fmt.Sprintf("… %d more (--json)", len(ps)-8))
			}
			lines = append(lines, "  pauses (source time): "+strings.Join(ps, ", "))
		}
		if r.Skipped > 0 {
			lines = append(lines, fmt.Sprintf("kept %d pinned edge(s) from your explicit cuts", r.Skipped))
		}
		data = map[string]any{"tighten": r}
	case "audio":
		if len(c.Args) != 2 {
			return nil, Errf(ExitUsage, "bad-args", "name the source: `cutroom edit audio mic`", "edit audio needs a source id")
		}
		s, err := sourceByID(e.p, c.Args[1])
		if err != nil {
			return nil, err
		}
		if s.Facts.Audio == nil {
			return nil, Errf(ExitUsage, "no-audio", "pick a source with sound (`cutroom inspect`)", "source %q has no audio", s.ID)
		}
		for _, g := range e.tl.Segments {
			if g.Source != s.ID && (s.Sync == nil || s.Sync.To != g.Source) {
				return nil, Errf(ExitInvalid, "audio-not-synced", fmt.Sprintf("align it first: `cutroom sync %s --to %s`", s.ID, g.Source),
					"%s has not been aligned to %s, so its sound would be out of sync", s.ID, g.Source)
			}
		}
		e.tl.Audio = s.ID
		label = "edit audio " + s.ID
		lines = append(lines, fmt.Sprintf("the timeline now takes its sound from %s", s.ID))
		if s.Sync != nil {
			lines = append(lines, fmt.Sprintf("  aligned to %s: offset %+.3f s, drift %+.1f ppm", s.Sync.To, s.Sync.Offset, s.Sync.PPM))
		}
	case "cut", "keep":
		res, desc, err := e.applySpan(c, op)
		if err != nil {
			return nil, err
		}
		label = desc
		lines = append(lines, res...)
	}
	e.tl.Recompute()
	for _, sd := range soundsBefore {
		if !e.tl.Kept(sd.src, sd.mid) {
			lines = append(lines, fmt.Sprintf("  this also removes the sound with no words at source %s–%s", clock(sd.a), clock(sd.b)))
		}
	}
	issues := e.validate()
	lines = append(lines, fmt.Sprintf("timeline: %s, %s (was %s)", plural(len(e.tl.Segments), "segment"), secs(e.tl.Duration()), secs(before)))
	fresh, open := splitIssues(issues, issuesBefore)
	lines = append(lines, issueLines(fresh, 6)...)
	if len(open) > 0 {
		var cs []string
		for _, x := range open {
			cs = append(cs, fmt.Sprintf("%s at output %s", x.Code, clock(x.At)))
		}
		lines = append(lines, "still open (unchanged by this edit): "+strings.Join(truncList(cs, 4), "; ")+" — `cutroom timeline` shows them with fixes")
	}
	if c.Bool("dry-run") {
		for i, l := range lines {
			lines[i] = strings.Replace(l, "this also removes", "this would also remove", 1)
			l = lines[i]
			for _, v := range [][2]string{{"cut ", "would cut "}, {"kept only ", "would keep only "}, {"tightened: ", "would tighten: "}, {"timeline: ", "timeline would be: "}} {
				if strings.HasPrefix(l, v[0]) {
					lines[i] = v[1] + strings.TrimPrefix(l, v[0])
				}
			}
		}
		lines = append(lines, "DRY RUN: nothing was changed; run the same command without --dry-run to apply it")
	} else {
		if err := timeline.Checkpoint(e.p.Dir, cacheDir(e.p), label); err != nil {
			return nil, err
		}
		if err := e.tl.Save(e.p.Dir); err != nil {
			return nil, err
		}
	}
	if data == nil {
		data = map[string]any{}
	}
	data["duration"] = e.tl.Duration()
	data["segments"] = len(e.tl.Segments)
	data["issues"] = nzIssues(issues)
	data["saved"] = !c.Bool("dry-run")
	return &Result{Text: strings.Join(lines, "\n"), Data: data, Next: []string{"cutroom timeline --joins"}}, nil
}

func nzIssues(is []timeline.Issue) []timeline.Issue {
	if is == nil {
		return []timeline.Issue{}
	}
	return is
}

func issueLines(is []timeline.Issue, n int) []string {
	var out []string
	for i, x := range is {
		if i == n {
			out = append(out, fmt.Sprintf("… %d more problem(s): `cutroom timeline`", len(is)-n))
			break
		}
		s := fmt.Sprintf("%s[%s] at output %s: %s", strings.ToUpper(x.Level), x.Code, clock(x.At), x.Message)
		if x.Fix != "" {
			s += " → " + x.Fix
		}
		out = append(out, s)
	}
	return out
}

// applySpan handles cut/keep by range, --text or --from/--to.
func (e *editEnv) applySpan(c *Ctx, op string) ([]string, string, error) {
	hasRange := len(c.Args) == 2
	hasText := c.Has("text")
	hasFrom := c.Has("from") || c.Has("to")
	n := 0
	for _, b := range []bool{hasRange, hasText, hasFrom} {
		if b {
			n++
		}
	}
	if n != 1 {
		return nil, "", Errf(ExitUsage, "bad-span", `give exactly one of: a range (`+"`cutroom edit "+op+" 6.3-9.2`"+`), --text "words", or --from "words" --to "words"`, "%s needs one span", op)
	}
	if hasFrom && !(c.Has("from") && c.Has("to")) {
		return nil, "", Errf(ExitUsage, "bad-span", `use both --from "first words" and --to "last words" (or --text for a short phrase)`, "--from and --to go together")
	}
	reason := c.Str("reason")
	if hasRange {
		src := c.Str("source")
		if src == "" {
			src = e.tl.Video
		}
		s, err := sourceByID(e.p, src)
		if err != nil {
			return nil, "", err
		}
		a, b, err := parseRange(c.Args[1], sourceRate(s))
		if err != nil {
			return nil, "", Errf(ExitUsage, "bad-range", "write a range as start-end in source time, e.g. 16.9-17.4", "%v", err)
		}
		desc := fmt.Sprintf("edit %s %s %s", op, s.ID, c.Args[1])
		if op == "cut" {
			rm := e.tl.Cut(s.ID, a, b, orStr(reason, "range "+c.Args[1]), true)
			if rm == 0 {
				return nil, "", Errf(ExitUsage, "nothing-cut", "that range is not in the timeline (already cut?); `cutroom timeline` shows what is kept", "nothing to cut in %s %s", s.ID, c.Args[1])
			}
			line := fmt.Sprintf("cut %s %s–%s (%s removed)", s.ID, clock(e.tl.Snap(a)), clock(e.tl.Snap(b)), secs(rm))
			if tx := e.tr[s.ID]; tx != nil {
				line += "  words removed: " + orNone(wordsIn(tx.Words, a, b))
			}
			return []string{line}, desc, nil
		}
		e.tl.Keep(s.ID, a, b, orStr(reason, "range "+c.Args[1]))
		return []string{fmt.Sprintf("kept only %s %s–%s", s.ID, clock(a), clock(b))}, desc, nil
	}

	// text
	var src string
	var i, j int
	var err error
	keptOnly := op == "cut"
	if hasText {
		src, i, j, err = e.resolve(c, c.Str("text"), keptOnly, "", -1, -1)
	} else {
		var iLast int
		src, i, iLast, err = e.resolve(c, c.Str("from"), keptOnly, "", -1, -1)
		if err == nil {
			// --to: the first match that starts after --from starts and ends at or after it ends
			_, _, j, err = e.resolve(c, c.Str("to"), keptOnly, src, i+1, iLast)
		}
	}
	if err != nil {
		return nil, "", err
	}
	words := e.tr[src].Words
	desc := fmt.Sprintf("edit %s %q", op, speech.Text(words, i, j))
	if op == "keep" {
		a := words[i].Start - e.tl.PadBefore
		b := words[j].End + e.tl.PadAfter
		if i > 0 {
			a = clampF(a, words[i-1].End, words[i].Start)
		}
		if j+1 < len(words) {
			b = clampF(b, words[j].End, words[j+1].Start)
		}
		e.tl.Keep(src, a, b, orStr(reason, "keep \""+speech.Text(words, i, j)+"\""))
		return []string{fmt.Sprintf("kept only %s %s–%s: %q", src, clock(a), clock(b), ellipsize(speech.Text(words, i, j), 120))}, desc, nil
	}
	r := e.tl.CutWords(src, words, i, j, orStr(reason, "cut \""+ellipsize(speech.Text(words, i, j), 60)+"\""))
	if r.Removed == 0 {
		return nil, "", Errf(ExitUsage, "nothing-cut", "those words are not in the timeline any more; `cutroom timeline` shows what is kept", "nothing to cut")
	}
	out := []string{fmt.Sprintf("cut %s %s–%s (%s removed): %q", src, clock(r.From), clock(r.To), secs(r.Removed), ellipsize(r.Text, 140))}
	if r.JoinPause >= 0 {
		out = append(out, fmt.Sprintf("join: …%s | %s… (pause %.2f s)", r.Before, r.After, r.JoinPause))
	}
	for _, w := range r.Warnings {
		out = append(out, "WARN "+w)
	}
	return out, desc, nil
}

// resolve finds a quoted phrase in the transcripts. If afterSrc != "", only matches in that source
// starting at or after word index afterIdx count, and the first one wins.
func (e *editEnv) resolve(c *Ctx, phrase string, keptOnly bool, afterSrc string, afterIdx, minLast int) (string, int, int, error) {
	if len(e.tr) == 0 {
		return "", 0, 0, Errf(ExitInvalid, "no-transcript", "run `cutroom transcribe` first, or cut by time range", "word matching needs a transcript")
	}
	type hit struct {
		src string
		m   speech.Match
	}
	var hits []hit
	for _, id := range e.p.IDs() {
		tx := e.tr[id]
		if tx == nil || (c.Str("source") != "" && id != c.Str("source")) || (afterSrc != "" && id != afterSrc) {
			continue
		}
		for _, m := range speech.Search(tx.Words, phrase) {
			if afterSrc != "" && (m.First < afterIdx || m.Last < minLast) {
				continue
			}
			if keptOnly {
				all := true
				for k := m.First; k <= m.Last; k++ {
					if !e.tl.Kept(id, (tx.Words[k].Start+tx.Words[k].End)/2) {
						all = false
					}
				}
				if !all {
					continue
				}
			}
			hits = append(hits, hit{id, m})
		}
	}
	if len(hits) == 0 {
		where := ""
		if keptOnly {
			where = " in what the timeline still keeps"
		}
		fix := "check the words with `cutroom transcript --search \"" + phrase + "\"` (quote what the transcript says, not what you expect)"
		if afterSrc != "" {
			fix = "--to must come after --from in the transcript; check both with `cutroom transcript --search`"
		}
		return "", 0, 0, Errf(ExitUsage, "text-not-found", fix, "%q not found%s", phrase, where)
	}
	if afterSrc != "" {
		return hits[0].src, hits[0].m.First, hits[0].m.Last, nil
	}
	if len(hits) > 1 {
		if c.Has("occurrence") {
			k := c.Int("occurrence")
			if k < 1 || k > len(hits) {
				return "", 0, 0, Errf(ExitUsage, "bad-occurrence", fmt.Sprintf("use 1–%d", len(hits)), "--occurrence %d out of range", k)
			}
			h := hits[k-1]
			return h.src, h.m.First, h.m.Last, nil
		}
		if c.Has("near") {
			nt, err := timecode.Parse(c.Str("near"), timecode.Rate{Num: 30, Den: 1})
			if err != nil {
				return "", 0, 0, Errf(ExitUsage, "bad-time", "--near takes a source time like 10.3 or 0:10.3", "%v", err)
			}
			best, bd := 0, math.Inf(1)
			for k, h := range hits {
				if d := math.Abs(h.m.Start - nt.Float()); d < bd {
					best, bd = k, d
				}
			}
			h := hits[best]
			return h.src, h.m.First, h.m.Last, nil
		}
		var opts []string
		for k, h := range hits {
			w := e.tr[h.src].Words
			opts = append(opts, fmt.Sprintf("  %d: %s %s  …%s [%s] %s…", k+1, h.src, clock(h.m.Start),
				speech.Text(w, max(0, h.m.First-4), h.m.First-1), h.m.Text, speech.Text(w, h.m.Last+1, min(len(w)-1, h.m.Last+4))))
		}
		return "", 0, 0, Errf(ExitUsage, "ambiguous-text", "add --occurrence N or --near <time>, or quote more words",
			"%q occurs %d times:\n%s", phrase, len(hits), strings.Join(opts, "\n"))
	}
	return hits[0].src, hits[0].m.First, hits[0].m.Last, nil
}

func wordsIn(ws []speech.Word, a, b float64) string {
	var o []string
	for _, w := range ws {
		m := (w.Start + w.End) / 2
		if m >= a && m < b {
			o = append(o, w.Text)
		}
	}
	return ellipsize(strings.Join(o, " "), 140)
}

func clampF(x, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, x)) }

func orStr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func ellipsize(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func secs(s float64) string { return fmt.Sprintf("%.3f s", s) }

func runTimeline(c *Ctx) (*Result, error) {
	e, err := loadEditEnv(c, false)
	if err != nil {
		return nil, err
	}
	tl := e.tl
	issues := e.validate()
	var lines []string
	srcTotal := 0.0
	for _, d := range e.dur {
		_ = d
	}
	if d, ok := e.dur[tl.Video]; ok {
		srcTotal = d
	}
	lines = append(lines, fmt.Sprintf("timeline: %d segment(s), %s output (%s fps, %dx%d), from %s of %s; sound from %s",
		len(tl.Segments), secs(tl.Duration()), fpsStr(tl.FPS), tl.Width, tl.Height, secs(srcTotal), tl.Video, orNone(tl.Audio)))
	if c.Bool("joins") {
		js := tl.Joins(e.tr)
		if len(js) == 0 {
			lines = append(lines, "no joins (one segment)")
		}
		for _, j := range js {
			p := "?"
			if j.Pause >= 0 {
				p = fmt.Sprintf("%.2f s", j.Pause)
			}
			lines = append(lines, fmt.Sprintf("%s  %s|%s  …%s | %s…  (pause %s)", clock(j.At), j.Left, j.Right, j.Before, j.After, p))
		}
		lines = truncLines(lines, c.Bool("all"))
		lines = append(lines, issueLines(issues, 10)...)
		return &Result{Text: strings.Join(lines, "\n"), Data: map[string]any{"joins": js, "issues": nzIssues(issues)},
			Next: []string{"cutroom listen <source range around a join>   (for any join you doubt)"}}, nil
	}
	type segOut struct {
		timeline.Segment
		OutStart float64 `json:"out_start"`
		Duration float64 `json:"duration"`
		Text     string  `json:"text,omitempty"`
	}
	var segs []segOut
	for _, g := range tl.Segments {
		d := float64(tl.Frames(g)) / tl.Rate()
		text := ""
		if tx := e.tr[g.Source]; tx != nil {
			text = wordsIn(tx.Words, g.In, g.Out)
		}
		segs = append(segs, segOut{g, g.OutStart, math.Round(d*1e6) / 1e6, text})
		lines = append(lines, fmt.Sprintf("%s  %-4s %s %s–%s  %s", clock(g.OutStart), g.ID, g.Source, clock(g.In), clock(g.Out), ellipsize(orNone(text), 90)))
	}
	lines = truncLines(lines, c.Bool("all"))
	if len(issues) == 0 {
		lines = append(lines, "no problems found")
	}
	lines = append(lines, issueLines(issues, 10)...)
	return &Result{Text: strings.Join(lines, "\n"), Data: map[string]any{"duration": tl.Duration(), "fps": tl.FPS, "width": tl.Width, "height": tl.Height,
		"video": tl.Video, "audio": tl.Audio, "segments": segs, "issues": nzIssues(issues)}, Next: []string{"cutroom timeline --joins"}}, nil
}

func truncLines(lines []string, all bool) []string {
	if all || len(lines) <= maxLines {
		return lines
	}
	n := len(lines) - maxLines
	return append(lines[:maxLines], fmt.Sprintf("… %d more (--all, or --json)", n))
}

type keptSound struct {
	src       string
	a, b, mid float64
}

// keptSounds lists the transcript's "sound" suspects the timeline currently keeps.
func (e *editEnv) keptSounds() []keptSound {
	var out []keptSound
	for _, id := range e.p.IDs() {
		tx := e.tr[id]
		if tx == nil {
			continue
		}
		for _, s := range tx.Suspects {
			if m := (s.Start + s.End) / 2; s.Kind == "sound" && e.tl.Kept(id, m) {
				out = append(out, keptSound{id, s.Start, s.End, m})
			}
		}
	}
	return out
}

// splitIssues separates issues this edit introduced from ones that were already open (same code and
// message), so an edit doesn't repeat a warning the agent has already seen in full.
func splitIssues(now, before []timeline.Issue) (fresh, open []timeline.Issue) {
	seen := map[string]bool{}
	for _, x := range before {
		seen[x.Code+"|"+x.Message] = true
	}
	for _, x := range now {
		if seen[x.Code+"|"+x.Message] && x.Level != "error" {
			open = append(open, x)
		} else {
			fresh = append(fresh, x)
		}
	}
	return
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return fmt.Sprintf("%d %ss", n, word)
}
