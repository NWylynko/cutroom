package cli

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/NWylynko/cutroom/internal/project"
	"github.com/NWylynko/cutroom/internal/render"
	"github.com/NWylynko/cutroom/internal/speech"
	"github.com/NWylynko/cutroom/internal/timecode"
	"github.com/NWylynko/cutroom/internal/timeline"
)

var editOps = []string{"cut", "keep", "add", "replace", "move", "split", "pin", "zoom", "frame", "tighten", "audio", "undo", "redo", "history", "reset"}

func init() {
	Register(&Command{
		Name:     "edit",
		Args:     "<cut|keep|add|replace|move|split|pin|zoom|frame|tighten|audio|undo|redo|history|reset> [factor|segment|time] [range|source]",
		MinArgs:  1,
		MaxArgs:  3,
		Category: "edit",
		Summary:  "Change the timeline: cut or keep by quoted words or times, tighten pauses, undo",
		Help: `The timeline starts as the whole first video source (created on the first edit).
Every edit is checked afterwards and problems are printed. Use --dry-run to preview without saving.

  edit cut --text "um, so"                       remove exactly these words (cut lands in the pauses)
  edit cut --from "By the way" --to "keyboard."  remove from the first words through the last words
  edit cut 16.9-17.4                             remove a source time range (seconds or m:ss.mmm)
  edit keep --from "Hi, today" --to "server is ready."   keep only this span (e.g. for a short)
  edit zoom 1.15 --from "Then install" --to "on it."   punch in on a span (1 = zoom back out)
        --anchor 0.5,0.4   the point to zoom on (0–1 across, 0–1 down; default 0.5,0.4 = a face)
        above 1.2× upscaling looks soft (a 1080p source at 1.3×); 4K sources can go to 2× for 1080p
  edit add --source take2 --from "So the" --to "works." [--after s3 | --before s1]
                                                 put a span of another take (or source) into the edit
                                                 (default: at the end); its edges are pinned
  edit replace --text "Then install Linux on it." --with pickup
                                                 swap a line for another take's version of it (cut +
                                                 add in one step; --from/--to for longer spans)
  edit move s5 --after s2 | --before s1 | --after start      reorder a segment
  edit split --text "Then install" | edit split 12.4        split a segment in two (nothing removed)
  edit pin s3 [--in|--out] [--unpin]             keep tighten off a segment's edges
  edit frame --anchor 0.4,0.5                    where a filled (vertical) frame is cropped (auto = on the speaker); --fill / --fit
  edit tighten [--min-cut 0.4]                   shorten every pause longer than the pads + min-cut
  edit audio mic                                 take the sound from source "mic" (after ` + "`cutroom sync mic`" + `)
  edit audio mic2 --for take2                    multi-take: segments from take2 use recorder mic2
                                                 (unmapped takes use their own sound; timeline warns)
  edit undo | edit redo | edit history | edit reset [--source id]
  edit <sub> --help                              just one subcommand and its flags

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
			{Name: "anchor", Kind: "string", Help: "zoom: point to zoom on as x,y in 0–1 (default 0.5,0.4); frame: where a filled frame is cropped"},
			{Name: "bars", Kind: "string", Help: "frame: what fills the frame around a fitted picture — black (default), blur, or a colour #rrggbb"},
			{Name: "for", Kind: "string", Help: "audio: only for segments from this video source (multi-take)"},
			{Name: "with", Kind: "string", Help: "replace: the source whose version of the words goes in (a pickup or another take)"},
			{Name: "after", Kind: "string", Help: "add/move: put it after this segment id (or \"start\")"},
			{Name: "before", Kind: "string", Help: "add/move: put it before this segment id (default for add: at the end)"},
			{Name: "fill", Kind: "bool", Help: "frame: scale the picture to fill the output frame (crop the rest)"},
			{Name: "fit", Kind: "bool", Help: "frame: fit the whole picture (bars where shapes differ)"},
			{Name: "size", Kind: "string", Help: "frame: output size: 1080p, 720p, 2160p, or WxH (e.g. 1920x1080, 1080x1920)"},
			{Name: "in", Kind: "bool", Help: "pin: only the segment's start"},
			{Name: "out", Kind: "bool", Help: "pin: only the segment's end"},
			{Name: "unpin", Kind: "bool", Help: "pin: release instead (tighten may move it)"},
			{Name: "dry-run", Kind: "bool", Help: "Show the result without saving"},
		},
		Subs: map[string]Sub{
			"frame": {Usage: `  edit frame --anchor 0.4,0.5   a filled frame (a vertical short) is cropped around this point of the
                                source (0–1 across, down): move it onto the speaker
  edit frame --fill | --fit     fill the output frame (crop) or fit the whole picture (bars)
  edit frame --fit --bars blur  the bars around a fitted picture: black, blur (the picture blurred
                                behind itself) or a colour #rrggbb
  edit frame --size 720x1280    change the output size (overlays and captions follow)
Look at the result with ` + "`cutroom look --output <time>`" + `.`, Flags: []string{"anchor", "fill", "fit", "bars", "size", "dry-run"}},
			"cut": {Usage: `  edit cut --text "um, so"                       remove exactly these words (cut lands in the pauses)
  edit cut --from "By the way" --to "keyboard."  remove from the first words through the last words
  edit cut 16.9-17.4 [--source id]               remove a source time range (seconds or m:ss.mmm)
Words ignore case, punctuation and um/uh; --from must match once (--near/--occurrence if it repeats).
A cut by words leaves 0.22 s after the last kept word and 0.10 s before the next. Its edges are pinned.`,
				Flags: []string{"text", "from", "to", "near", "occurrence", "source", "reason", "dry-run"}},
			"keep": {Usage: `  edit keep --from "Hi, today" --to "server is ready."   keep only this span (e.g. for a short)
  edit keep 0:12-0:58 [--source id]                      keep only a source time range`,
				Flags: []string{"text", "from", "to", "near", "occurrence", "source", "reason", "dry-run"}},
			"add": {Usage: `  edit add --source take2 --from "So the" --to "works." [--after s3 | --before s1 | --after start]
  edit add 3.2-9.8 --source take2
Puts a span of another take (or source) into the edit; default at the end. Its edges are pinned.
Warns when it duplicates words already kept (to swap a line, use edit replace).`,
				Flags: []string{"source", "text", "from", "to", "near", "occurrence", "after", "before", "reason", "dry-run"}},
			"replace": {Usage: `  edit replace --text "Then install Linux on it." --with pickup
  edit replace --from "Then install" --to "on it." --with pickup
Cuts the edit's version of the words and puts --with's version in the same place, in one step.`,
				Flags: []string{"with", "text", "from", "to", "near", "occurrence", "source", "reason", "dry-run"}},
			"move": {Usage: `  edit move s5 --after s2 | --before s1 | --after start      reorder a segment (ids: cutroom timeline)
  edit move --text "Then install" --after s2                 ... the segment that keeps these words`,
				Flags: []string{"text", "near", "occurrence", "after", "before", "dry-run"}},
			"split": {Usage: `  edit split 12.4 [--source id]          split the segment holding this source time in two
  edit split --text "Then install"        ... just before these words (in the pause before them)
Nothing is removed; both new edges are pinned. Then move, zoom or cut each part on its own.`,
				Flags: []string{"text", "source", "near", "occurrence", "dry-run"}},
			"pin": {Usage: `  edit pin s3 [--in | --out] [--unpin]   pin a segment's edges so tighten never moves them
  edit pin --text "Then install"          ... the segment that keeps these words
Explicit cuts pin the edges they make already; pin is for edges you set some other way.`,
				Flags: []string{"text", "near", "occurrence", "in", "out", "unpin", "dry-run"}},
			"zoom": {Usage: `  edit zoom 1.15 --from "Then install" --to "on it."   punch in on a span
  edit zoom 1.15 12.0-15.5 [--source id]               ... or a source time range
  edit zoom 1 --from ... --to ...                      zoom back out
Above 1.2× upscaling looks soft (a 1080p source at 1.3×); 4K sources can go to 2× for 1080p.`,
				Flags: []string{"anchor", "text", "from", "to", "near", "occurrence", "source", "dry-run"}},
			"tighten": {Usage: `  edit tighten [--min-cut 0.4]   shorten every pause longer than the pads + min-cut
Never moves pinned edges (your explicit cuts). Prints the sounds inside removed pauses.`,
				Flags: []string{"min-cut", "dry-run"}},
			"audio": {Usage: `  edit audio mic                take the sound from source "mic" (after ` + "`cutroom sync mic`" + `)
  edit audio mic2 --for take2   multi-take: segments from take2 use recorder mic2`,
				Flags: []string{"for", "dry-run"}},
			"undo": {Usage: `  edit undo            put the timeline back as it was before the last edit (repeat to go further back)
  edit undo --dry-run  say what it would undo and the timeline it would leave, change nothing
` + "`cutroom edit redo`" + ` puts an undone edit back (until the next edit). ` + "`cutroom edit history`" + ` lists both.`, Flags: []string{"dry-run"}},
			"redo":    {Usage: `  edit redo   put back the edit the last undo took away (repeat for earlier undos; a new edit clears them)`, Flags: []string{"dry-run"}},
			"history": {Usage: `  edit history   every edit in order, newest last (what edit undo would undo next)`},
			"reset": {Usage: `  edit reset [--source id]   start over: the timeline becomes the whole video source again
(undo-able: edit undo brings the old timeline back)`,
				Flags: []string{"source"}},
		},
		Run: runEdit,
	})
	Register(&Command{
		Name:     "timeline",
		Args:     "[validate]",
		MaxArgs:  1,
		Category: "edit",
		Summary:  "Show the timeline: segments with output times and text, every join, and problems",
		Help: `Default: one line per segment (output time, source range, first and last words) and any problems.
  --joins   every cut point with the words on both sides and the pause across it — read these
            before rendering; a bad join is the most common editing mistake
Each warning is decided one of two ways:
  --apply retake-kept@4.09    run the warning's own suggested edit (the time picks one when a code repeats)
  --accept retake-kept@4.09 --reason "both takes wanted"   keep it as it is: listed, no longer open
  validate  only the problems; exit status 3 when any is an error (for scripts and pre-render checks)
Examples:
  cutroom timeline
  cutroom timeline --joins
  cutroom timeline --apply retake-kept
  cutroom timeline --accept audio-mixed --reason "same mic"`,
		Flags: []Flag{
			{Name: "joins", Kind: "bool", Help: "List every join (cut point)"},
			{Name: "all", Kind: "bool", Help: "Don't truncate"},
			{Name: "apply", Kind: "string", Help: "Run a warning's suggested edit: <code>[@<output time>]"},
			{Name: "accept", Kind: "string", Help: "Keep a warning as decided: <code>[@<output time>] (with --reason)"},
			{Name: "reason", Kind: "string", Help: "accept: why it is fine (stored with the timeline)"},
			{Name: "dry-run", Kind: "bool", Help: "apply: show the edit without saving"},
		},
		Run: runTimeline,
	})
}

// edit context: project, timeline, transcripts, durations.
type editEnv struct {
	zoom, ax, ay float64 // edit zoom: factor and anchor
	p            *project.Project
	tl           *timeline.Timeline
	tr           timeline.Transcripts
	dur          map[string]float64
	made         bool
	notes        []timeline.Issue // said once: segments trimmed to where their source really ends
	videoOnly    bool             // quotes resolve only in sources with a picture (a short needs one)
}

func loadEditEnv(c *Ctx, create bool) (*editEnv, error) {
	p, err := needProject(c.Dir)
	if err != nil {
		return nil, err
	}
	e := &editEnv{p: p, tr: timeline.Transcripts{}, dur: map[string]float64{}}
	for _, id := range p.IDs() {
		s := p.Sources[id]
		e.dur[id] = s.Facts.End()
		if t, err := speech.Load(cacheDir(p), id); err == nil {
			if stale, _ := p.Stale(s); !stale {
				e.tr[id] = t
			}
		}
	}
	e.tl, err = timeline.Load(p.Dir)
	if errors.Is(err, timeline.ErrNone) {
		if !create {
			fix := "make the first edit (e.g. `cutroom edit tighten`), which creates it"
			if ms, _ := filepath.Glob(filepath.Join(p.Dir, "shorts", "*", timeline.File)); len(ms) > 0 {
				var ds []string
				for _, m := range ms {
					ds = append(ds, "--project "+relToCwd(filepath.Dir(m)))
				}
				fix += "; or work on a short: " + strings.Join(ds, ", ")
			}
			return nil, Errf(ExitInvalid, "no-timeline", fix, "no timeline yet in %s", relToCwd(p.Dir))
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
	e.trimToEnds()
	heardWords(p, e.tl, e.tr, e.dur)
	e.tl.Quiet = quietLevels(p, e.tl)
	return e, nil
}

// trimToEnds stops a segment that runs a little past its source's real end (a container longer than
// its streams; timelines made before cutroom measured that) at the last frame the source has.
func (e *editEnv) trimToEnds() {
	changed := false
	for i := range e.tl.Segments {
		g := &e.tl.Segments[i]
		d, ok := e.dur[g.Source]
		if !ok || g.Out <= d+1e-6 || g.Out-d > 0.5 || g.In >= d-0.1 {
			continue
		}
		end := e.tl.SnapDown(d)
		e.notes = append(e.notes, timeline.Issue{Level: "info", Code: "trimmed-to-source-end", Segment: g.ID, At: g.OutStart,
			Message: fmt.Sprintf("segment %s ran to %s %.3f, past where its picture and sound end (%.3f); it now stops at %.3f", g.ID, g.Source, g.Out, d, end)})
		g.Out = end
		changed = true
	}
	if changed {
		e.tl.Recompute()
	}
}

func newTimeline(s *project.Source) *timeline.Timeline {
	v := s.Facts.Video
	audio, grid := "", 0.0
	if s.Facts.Audio != nil {
		audio, grid = s.ID, s.Facts.Audio.Start
	}
	return timeline.New(v.FPS, v.Width, v.Height, s.ID, audio, grid, s.Facts.End())
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
		if c.Bool("dry-run") { // say what it would restore; change nothing
			what, restores, err := timeline.UndoPreview(cacheDir(p))
			if err != nil {
				return nil, Errf(ExitInvalid, "nothing-to-undo", "`cutroom edit history` lists edits", "%v", err)
			}
			text := "would undo: " + what
			if restores != nil {
				text += fmt.Sprintf("\ntimeline would be: %s, %s", plural(len(restores.Segments), "segment"), secs(restores.Duration()))
			} else {
				text += "\nthere would be no timeline (the undone edit created it)"
			}
			text += "\nDRY RUN: nothing was changed; `cutroom edit undo` does it (and `cutroom edit redo` puts it back)"
			return &Result{Text: text, Data: map[string]any{"would_undo": what}, Next: []string{"cutroom edit undo"}}, nil
		}
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
	case "redo":
		if c.Bool("dry-run") {
			if r := timeline.Redoable(cacheDir(p)); len(r) > 0 {
				return &Result{Text: "would redo: " + r[0] + "\nDRY RUN: nothing was changed", Next: []string{"cutroom edit redo"}}, nil
			}
			return nil, Errf(ExitInvalid, "nothing-to-redo", "redo puts back what `cutroom edit undo` took away, until the next edit", "nothing to redo")
		}
		what, err := timeline.Redo(p.Dir, cacheDir(p))
		if err != nil {
			return nil, Errf(ExitInvalid, "nothing-to-redo", "redo puts back what `cutroom edit undo` took away, until the next edit", "%v", err)
		}
		text := "redid: " + what
		if tl, err := timeline.Load(p.Dir); err == nil {
			text += fmt.Sprintf("\ntimeline: %d segments, %s", len(tl.Segments), secs(tl.Duration()))
		}
		return &Result{Text: text, Data: map[string]string{"redone": what}, Next: []string{"cutroom timeline"}}, nil
	case "history":
		h := timeline.History(cacheDir(p))
		var lines []string
		for _, e := range h {
			lines = append(lines, fmt.Sprintf("%3d  %s  %s", e.N, e.Time, e.Op))
		}
		if len(lines) == 0 {
			lines = []string{"no edits yet"}
		}
		if r := timeline.Redoable(cacheDir(p)); len(r) > 0 {
			lines = append(lines, fmt.Sprintf("undone (`cutroom edit redo` puts back, next first): %s", strings.Join(r, "; ")))
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
		l := fmt.Sprintf("%s timeline from %s (whole source, %s); output %dx%d at %s fps, the source's", verb, e.tl.Video, secs(e.tl.Duration()), e.tl.Width, e.tl.Height, fpsStr(e.tl.FPS))
		if e.tl.Width*e.tl.Height > 1920*1080 {
			l += fmt.Sprintf(" — for a 1080p video: `cutroom edit frame --size 1080p`")
		}
		videos := 0
		for _, id := range e.p.IDs() {
			if s := e.p.Sources[id]; s.Facts != nil && s.Facts.Video != nil {
				videos++
			}
		}
		if videos > 1 && !c.Has("source") {
			l += "; it is the longest video — other takes go in with `cutroom edit add --source <take> …` (or start from another with --source)"
		}
		lines = append(lines, l)
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
		if c.Has("for") { // multi-take: this recorder supplies the sound for one video source
			v := c.Str("for")
			vs, err := sourceByID(e.p, v)
			if err != nil {
				return nil, err
			}
			if s.ID != vs.ID && (s.Sync == nil || s.Sync.To != vs.ID) {
				return nil, Errf(ExitInvalid, "audio-not-synced", fmt.Sprintf("align it first: `cutroom sync %s --to %s`", s.ID, vs.ID),
					"%s has not been aligned to %s, so its sound would be out of sync", s.ID, vs.ID)
			}
			if e.tl.AudioFor == nil {
				e.tl.AudioFor = map[string]string{}
			}
			if s.ID == vs.ID {
				delete(e.tl.AudioFor, vs.ID) // its own sound: the default for an unmapped take
				e.tl.AudioFor[vs.ID] = vs.ID
			} else {
				e.tl.AudioFor[vs.ID] = s.ID
			}
			label = fmt.Sprintf("edit audio %s --for %s", s.ID, vs.ID)
			lines = append(lines, fmt.Sprintf("segments from %s now take their sound from %s", vs.ID, s.ID))
			break
		}
		ok := false
		for _, g := range e.tl.Segments {
			if g.Source == s.ID || (s.Sync != nil && s.Sync.To == g.Source) {
				ok = true
			}
		}
		if !ok && len(e.tl.Segments) > 0 {
			g := e.tl.Segments[0]
			return nil, Errf(ExitInvalid, "audio-not-synced", fmt.Sprintf("align it first: `cutroom sync %s --to %s`", s.ID, g.Source),
				"%s has not been aligned to %s, so its sound would be out of sync", s.ID, g.Source)
		}
		e.tl.Audio = s.ID
		label = "edit audio " + s.ID
		lines = append(lines, fmt.Sprintf("the timeline now takes its sound from %s", s.ID))
		if s.Sync != nil {
			lines = append(lines, fmt.Sprintf("  aligned to %s: offset %+.3f s, drift %+.1f ppm", s.Sync.To, s.Sync.Offset, s.Sync.PPM))
		}
	case "zoom":
		if len(c.Args) < 2 {
			return nil, Errf(ExitUsage, "bad-args", `give the factor first: `+"`cutroom edit zoom 1.15 --from \"…\" --to \"…\"`", "edit zoom needs a factor")
		}
		z, err := strconv.ParseFloat(c.Args[1], 64)
		if err != nil || z < 1 || z > 4 {
			return nil, Errf(ExitUsage, "bad-zoom", "a zoom factor between 1 (none) and 4, e.g. 1.15", "bad zoom factor %q", c.Args[1])
		}
		e.zoom, e.ax, e.ay = z, 0.5, 0.4
		if c.Has("anchor") {
			xs, ys, ok := strings.Cut(c.Str("anchor"), ",")
			x, err1 := strconv.ParseFloat(strings.TrimSpace(xs), 64)
			y, err2 := strconv.ParseFloat(strings.TrimSpace(ys), 64)
			if !ok || err1 != nil || err2 != nil || x < 0 || x > 1 || y < 0 || y > 1 {
				return nil, Errf(ExitUsage, "bad-anchor", "x,y between 0 and 1, e.g. --anchor 0.5,0.4 (0,0 = top left)", "bad --anchor %q", c.Str("anchor"))
			}
			e.ax, e.ay = x, y
		}
		c.Args = append([]string{c.Args[0]}, c.Args[2:]...) // the rest is the span, as for cut
		res, desc, err := e.applySpan(c, op)
		if err != nil {
			return nil, err
		}
		label = desc
		lines = append(lines, res...)
	case "add":
		res, desc, err := e.applyAdd(c)
		if err != nil {
			return nil, err
		}
		label = desc
		lines = append(lines, res...)
	case "replace":
		res, desc, err := e.applyReplace(c)
		if err != nil {
			return nil, err
		}
		label = desc
		lines = append(lines, res...)
	case "frame":
		if !c.Has("anchor") && !c.Bool("fill") && !c.Bool("fit") && !c.Has("size") && !c.Has("bars") {
			return nil, Errf(ExitUsage, "missing-flag", "`cutroom edit frame --anchor 0.4,0.5` (where to crop a filled frame), --fill, --fit [--bars blur], or --size 720x1280", "edit frame needs --anchor, --fill, --fit, --bars or --size")
		}
		if c.Has("bars") {
			b := strings.ToLower(c.Str("bars"))
			switch {
			case b == "black" || b == "default":
				b = ""
			case b == "blur":
			default:
				col, err := timeline.ParseColor(b)
				if err != nil {
					return nil, Errf(ExitUsage, "bad-bars", "--bars black, blur, or a colour like #101718", "%v", err)
				}
				b = fmt.Sprintf("#%02x%02x%02x", col[0], col[1], col[2])
			}
			e.tl.Bars = b
			if !c.Bool("fill") && !c.Has("anchor") {
				e.tl.Fill = false // bars only show around a fitted picture
			}
		}
		if c.Has("size") {
			w, h, err := parseSize(c.Str("size"), e.tl.Width, e.tl.Height)
			if err != nil {
				return nil, err
			}
			e.tl.Width, e.tl.Height = w, h
		}
		if c.Bool("fill") && c.Bool("fit") {
			return nil, Errf(ExitUsage, "bad-flags", "give --fill or --fit, not both", "--fill and --fit both given")
		}
		if c.Bool("fill") || c.Has("anchor") {
			e.tl.Fill = true
		}
		if c.Bool("fit") {
			e.tl.Fill = false
		}
		if c.Has("anchor") && c.Str("anchor") == "auto" {
			src, a, b := e.mainPicture()
			if src == "" {
				return nil, Errf(ExitUsage, "no-picture", "give the point: --anchor x,y", "the edit keeps no picture to find a speaker in")
			}
			ax, note := speakerAnchor(e.p.Abs(e.p.Sources[src]), a, b-a)
			e.tl.FillX, e.tl.FillY = ax, 0.5
			lines = append(lines, note)
		} else if c.Has("anchor") {
			ax, ay, err := parseAnchor(c.Str("anchor"))
			if err != nil {
				return nil, err
			}
			e.tl.FillX, e.tl.FillY = ax, ay
		}
		label = fmt.Sprintf("edit frame %dx%d fill=%v anchor %.2f,%.2f", e.tl.Width, e.tl.Height, e.tl.Fill, e.tl.FillX, e.tl.FillY)
		if e.tl.Fill {
			lines = append(lines, fmt.Sprintf("frame: the picture fills %dx%d, cropped around %.2f,%.2f (0–1 across, down) — check it with `cutroom look --output <time>`", e.tl.Width, e.tl.Height, e.tl.FillX, e.tl.FillY))
		} else {
			bars := "black"
			if e.tl.Bars != "" {
				bars = e.tl.Bars
			}
			lines = append(lines, fmt.Sprintf("frame: the whole picture fits %dx%d (%s bars where the shapes differ; --bars blur|black|#rrggbb)", e.tl.Width, e.tl.Height, bars))
		}
	case "split":
		src, at := "", 0.0
		switch {
		case c.Has("text"):
			sid, i, _, err := e.resolve(c, c.Str("text"), true, "", 0, 0)
			if err != nil {
				return nil, err
			}
			ws := e.tr[sid].Words
			src, at = sid, ws[i].Start // just before the quoted words
			if i > 0 {
				at = (ws[i-1].End + ws[i].Start) / 2 // in the gap before them
			}
		case len(c.Args) == 2:
			v, err := timecode.Parse(c.Args[1], timecode.Rate{Num: 30, Den: 1})
			if err != nil {
				return nil, Errf(ExitUsage, "bad-time", "a source time like 12.4 or 0:12.4, or --text \"words\" to split before them", "%v", err)
			}
			src, at = orStr(c.Str("source"), e.tl.Video), v.Float()
		default:
			return nil, Errf(ExitUsage, "bad-args", "`cutroom edit split 12.4` (source time) or `cutroom edit split --text \"Then install\"`", "edit split needs a time or --text")
		}
		l, r, ok := e.tl.Split(src, at)
		if !ok {
			return nil, Errf(ExitUsage, "not-in-edit", "split inside a kept segment (`cutroom timeline` lists them with source ranges)", "%s %.3f is not inside a segment of the edit", src, at)
		}
		label = fmt.Sprintf("edit split %s %.3f", src, at)
		lines = append(lines, fmt.Sprintf("split at %s %s: %s | %s (nothing removed; both edges pinned). Move, zoom or cut each part on its own.", src, clock(at), l, r))
	case "pin":
		id, err := e.segmentArg(c, "`cutroom edit pin s3` or `cutroom edit pin --text \"Then install\"`")
		if err != nil {
			return nil, err
		}
		g := e.tl.Find(id)
		in, out := !c.Bool("out"), !c.Bool("in")
		v := !c.Bool("unpin")
		if in {
			g.PinIn = v
		}
		if out {
			g.PinOut = v
		}
		what := map[bool]string{true: "pinned", false: "unpinned"}[v]
		edges := map[[2]bool]string{{true, true}: "both edges", {true, false}: "the start", {false, true}: "the end"}[[2]bool{in, out}]
		label = fmt.Sprintf("edit pin %s", g.ID)
		lines = append(lines, fmt.Sprintf("%s %s of %s: tighten %s move %s", what, edges, g.ID, map[bool]string{true: "won't", false: "may"}[v], map[bool]string{true: "them", false: "it"}[in && out]))
	case "move":
		id, err := e.segmentArg(c, "`cutroom edit move s5 --after s2` or `cutroom edit move --text \"Then install\" --after s2`")
		if err != nil {
			return nil, err
		}
		pos, where, err := e.position(c, id)
		if err != nil {
			return nil, err
		}
		e.tl.Move(id, pos)
		label = fmt.Sprintf("edit move %s %s", id, where)
		lines = append(lines, fmt.Sprintf("moved %s %s; it now starts at output %s", id, where, clock(e.tl.Find(id).OutStart)))
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
	for _, x := range fresh {
		if x.Level == "warn" {
			lines = append(lines, "decide each WARN: `cutroom timeline --apply <code>@<time>` runs its suggested edit; `--accept <code>@<time> --reason \"…\"` keeps it")
			break
		}
	}
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
	var shown []timeline.Issue
	for _, x := range is {
		if x.Level != "accepted" { // decided: `cutroom timeline` lists them apart
			shown = append(shown, x)
		}
	}
	is = shown
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
	for _, f := range []string{"from", "to", "text"} {
		if v := c.Str(f); v != "" && isTimeish(v) {
			return nil, "", Errf(ExitUsage, "bad-span", fmt.Sprintf("--%s takes quoted words; for times give a range: `cutroom edit %s %s-<end>`", f, op, v), "--%s %q is a time, not words", f, v)
		}
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
		if op == "zoom" {
			desc = fmt.Sprintf("edit zoom %g %s %s", e.zoom, s.ID, c.Args[1])
			return e.applyZoom(s.ID, a, b), desc, nil
		}
		if op == "cut" {
			why := "range " + c.Args[1]
			if tx := e.tr[s.ID]; tx != nil {
				if w := wordsIn(tx.Words, a, b); w != "" {
					why += ": \"" + ellipsize(w, 60) + "\"" // so `timeline` and history show what it took
				}
			}
			rm := e.tl.Cut(s.ID, a, b, orStr(reason, why), true)
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
	keptOnly := op == "cut" || op == "zoom"
	if hasText {
		src, i, j, err = e.resolve(c, c.Str("text"), keptOnly, "", -1, -1)
	} else {
		var iLast int
		src, i, iLast, err = e.resolve(c, c.Str("from"), keptOnly, "", -1, -1)
		if err == nil {
			// --to: the first match that starts at or after --from and ends at or after it ends (--to = --from: that phrase)
			_, _, j, err = e.resolve(c, c.Str("to"), keptOnly, src, i, iLast)
		}
	}
	if err != nil {
		return nil, "", err
	}
	words := e.tr[src].Words
	desc := fmt.Sprintf("edit %s %q", op, speech.Text(words, i, j))
	if op == "zoom" {
		// from just before the first word to just after the last, inside the pauses around them
		a, b := words[i].Start-e.tl.PadBefore, words[j].End+e.tl.PadAfter
		if i > 0 {
			a = e.tl.Boundary(src, words[i-1].End, words[i].Start, clampF(a, words[i-1].End, words[i].Start))
		}
		if j+1 < len(words) {
			b = e.tl.Boundary(src, words[j].End, words[j+1].Start, clampF(b, words[j].End, words[j+1].Start))
		}
		desc = fmt.Sprintf("edit zoom %g %q", e.zoom, ellipsize(speech.Text(words, i, j), 60))
		return e.applyZoom(src, a, b), desc, nil
	}
	if op == "keep" {
		a := words[i].Start - e.tl.PadBefore
		b := words[j].End + e.tl.PadAfter
		if i > 0 {
			a = e.tl.Boundary(src, words[i-1].End, words[i].Start, clampF(a, words[i-1].End, words[i].Start))
		}
		if j+1 < len(words) {
			b = e.tl.Boundary(src, words[j].End, words[j+1].Start, clampF(b, words[j].End, words[j+1].Start))
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
		if e.videoOnly && c.Str("source") == "" {
			if s := e.p.Sources[id]; s == nil || s.Facts == nil || s.Facts.Video == nil {
				continue
			}
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
		fix := "add --occurrence N or --near <time>, or quote more words"
		srcs := map[string]bool{}
		for _, h := range hits {
			srcs[h.src] = true
		}
		if len(srcs) > 1 && c.Str("source") == "" {
			fix = "the words are in more than one source: add --source <id> (" + strings.Join(e.p.IDs(), ", ") + "), or --occurrence N / --near <time>"
		}
		return "", 0, 0, Errf(ExitUsage, "ambiguous-text", fix,
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
	if c.Has("apply") || c.Has("accept") {
		return decideIssue(c, e)
	}
	tl := e.tl
	issues := e.validate()
	if len(c.Args) == 1 {
		if c.Args[0] != "validate" {
			return nil, Errf(ExitUsage, "bad-args", "`cutroom timeline` or `cutroom timeline validate`", "unknown timeline argument %q", c.Args[0])
		}
		return validateOnly(issues)
	}
	var lines []string
	srcTotal := 0.0
	for _, d := range e.dur {
		_ = d
	}
	if d, ok := e.dur[tl.Video]; ok {
		srcTotal = d
	}
	lines = append(lines, fmt.Sprintf("timeline: %d segment(s), %s output (%s fps, %dx%d), from %s of %s; sound from %s",
		len(tl.Segments), secs(tl.Duration()), fpsStr(tl.FPS), tl.Width, tl.Height, secs(srcTotal), tl.Video, e.soundFrom()))
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
			if j.Kind == "zoom" {
				l, r := tl.Find(j.Left), tl.Find(j.Right)
				lines = append(lines, fmt.Sprintf("%s  %s|%s  zoom %.2f× → %.2f× (no cut; the picture continues)", clock(j.At), j.Left, j.Right, l.ZoomOf(), r.ZoomOf()))
				continue
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
		zt := ""
		if z := g.ZoomOf(); z > 1 {
			ax, ay := g.Anchor()
			zt = fmt.Sprintf("  [zoom %.2f× at %.2f,%.2f]", z, ax, ay)
		}
		lines = append(lines, fmt.Sprintf("%s  %-4s %s %s–%s  %s%s", clock(g.OutStart), g.ID, g.Source, clock(g.In), clock(g.Out), ellipsize(orNone(text), 90), zt))
	}
	lines = truncLines(lines, c.Bool("all"))
	if n := len(tl.Overlays); n > 0 {
		var ids []string
		for _, o := range tl.Overlays {
			if a, _, ok, _ := tl.Window(o); ok {
				ids = append(ids, fmt.Sprintf("%s at %s", o.ID, clock(a)))
			} else {
				ids = append(ids, o.ID+" (not shown)")
			}
		}
		lines = append(lines, fmt.Sprintf("overlays: %s — `cutroom overlay list`", strings.Join(truncList(ids, 6), ", ")))
	}
	open, warns := 0, 0
	var accepted []string
	for _, x := range issues {
		switch x.Level {
		case "accepted":
			_, why := tl.IsAccepted(x)
			accepted = append(accepted, fmt.Sprintf("%s at %s (%q)", x.Code, clock(x.At), why))
		case "warn":
			warns++
			open++
		case "error":
			open++
		}
	}
	if open == 0 {
		lines = append(lines, "no problems found")
	}
	lines = append(lines, issueLines(issues, 10)...)
	if len(accepted) > 0 {
		lines = append(lines, "accepted (decided, not open): "+strings.Join(truncList(accepted, 4), "; "))
	}
	if warns > 0 {
		lines = append(lines, "decide each WARN: fix it (`cutroom timeline --apply <code>@<time>` runs its suggested edit) or keep it (`cutroom timeline --accept <code>@<time> --reason \"…\"`)")
	}
	return &Result{Text: strings.Join(lines, "\n"), Data: map[string]any{"duration": tl.Duration(), "fps": tl.FPS, "width": tl.Width, "height": tl.Height,
		"video": tl.Video, "audio": tl.Audio, "segments": segs, "issues": nzIssues(issues)}, Next: []string{"cutroom timeline --joins"}}, nil
}

// segmentArg is the segment an op names: an id (`s5`), or --text "words" — the segment that keeps them.
func (e *editEnv) segmentArg(c *Ctx, usage string) (string, error) {
	if len(c.Args) == 2 {
		if e.tl.Find(c.Args[1]) == nil {
			return "", Errf(ExitUsage, "unknown-segment", "segment ids: "+strings.Join(e.tl.IDs(), ", "), "no segment %q", c.Args[1])
		}
		return c.Args[1], nil
	}
	if !c.Has("text") {
		return "", Errf(ExitUsage, "bad-args", usage, "name the segment: an id from `cutroom timeline`, or --text \"words it keeps\"")
	}
	sid, i, _, err := e.resolve(c, c.Str("text"), true, "", 0, 0)
	if err != nil {
		return "", err
	}
	w := e.tr[sid].Words[i]
	m := (w.Start + w.End) / 2
	for _, g := range e.tl.Segments {
		if g.Source == sid && m >= g.In && m < g.Out {
			return g.ID, nil
		}
	}
	return "", Errf(ExitUsage, "not-in-edit", "those words are cut; quote words the edit keeps (`cutroom transcript --output`)", "%q (source %s %s) is not in the edit", c.Str("text"), sid, clock(w.Start))
}

// validateOnly is `timeline validate`: the problems alone, an error exit when any is an error.
func validateOnly(issues []timeline.Issue) (*Result, error) {
	errs, warns, acc := 0, 0, 0
	var first string
	for _, x := range issues {
		switch x.Level {
		case "error":
			errs++
			if first == "" {
				first = x.Message
			}
		case "warn":
			warns++
		case "accepted":
			acc++
		}
	}
	summary := fmt.Sprintf("%d error(s), %d open warning(s), %d accepted", errs, warns, acc)
	lines := issueLines(issues, 50)
	if errs > 0 {
		return nil, Errf(ExitInvalid, "invalid-timeline", "fix the errors (`cutroom timeline` shows each with its fix)", "%s — first: %s\n%s", summary, first, strings.Join(lines, "\n"))
	}
	verdict := "valid"
	if warns > 0 {
		verdict = "valid (warnings to decide: --apply or --accept)"
	}
	return &Result{Text: strings.Join(append([]string{verdict + ": " + summary}, lines...), "\n"), Data: map[string]any{"valid": true, "errors": 0, "warnings": warns, "accepted": acc, "issues": nzIssues(issues)},
		Next: []string{"cutroom timeline"}}, nil
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
// parseSize reads an output size: WxH, or 2160p / 1440p / 1080p / 720p / 4k in the timeline's
// orientation (vertical timelines get 1080x1920 for 1080p).
func parseSize(s string, curW, curH int) (int, int, error) {
	short := map[string]int{"2160p": 2160, "4k": 2160, "1440p": 1440, "1080p": 1080, "720p": 720, "540p": 540}
	if n, ok := short[strings.ToLower(s)]; ok {
		long := n * 16 / 9
		if curH > curW {
			return n, long, nil
		}
		return long, n, nil
	}
	var w, h int
	if _, err := fmt.Sscanf(s, "%dx%d", &w, &h); err != nil || w < 64 || h < 64 || w%2+h%2 != 0 {
		return 0, 0, Errf(ExitUsage, "bad-flag-value", "--size 1080p (or 720p, 1440p, 2160p), or WxH with even numbers, e.g. 1920x1080 or 1080x1920", "bad --size %q", s)
	}
	return w, h, nil
}

func splitIssues(now, before []timeline.Issue) (fresh, open []timeline.Issue) {
	seen := map[string]bool{}
	for _, x := range before {
		seen[x.Code+"|"+x.Message] = true
	}
	for _, x := range now {
		if x.Level == "accepted" { // decided: neither new nor open
			continue
		}
		if seen[x.Code+"|"+x.Message] && x.Level != "error" {
			if x.Level != "info" { // notes are said once, not carried as open problems
				open = append(open, x)
			}
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

// applyZoom sets the zoom on a source span and reports it, warning when it upscales (L40).
func (e *editEnv) applyZoom(src string, a, b float64) []string {
	n := e.tl.Zoom(src, a, b, e.zoom, e.ax, e.ay)
	if e.zoom == 1 {
		return []string{fmt.Sprintf("zoomed out %s %s–%s (%d segment(s))", src, clock(e.tl.Snap(a)), clock(e.tl.Snap(b)), n)}
	}
	out := []string{fmt.Sprintf("zoom %.2f× on %s %s–%s, centred at %.2f,%.2f (%d segment(s))", e.zoom, src, clock(e.tl.Snap(a)), clock(e.tl.Snap(b)), e.ax, e.ay, n)}
	if s := e.p.Sources[src]; s != nil && s.Facts != nil && s.Facts.Video != nil && s.Facts.Video.Height > 0 {
		up := e.zoom * float64(e.tl.Height) / float64(s.Facts.Video.Height)
		if up > 1.2+1e-9 {
			out = append(out, fmt.Sprintf("WARN this upscales the source %.2f× (%dp source, %.2f× zoom, %dp output): above 1.2× it looks soft (L40) — try `cutroom edit zoom %.2f …`",
				up, s.Facts.Video.Height, e.zoom, e.tl.Height, 1.2*float64(s.Facts.Video.Height)/float64(e.tl.Height)))
		}
	}
	out = append(out, "check the framing: `cutroom look --output <time>` shows the zoomed picture")
	return out
}

// position resolves --after/--before into an index in the timeline without segment `moving` ("" for add).
func (e *editEnv) position(c *Ctx, moving string) (int, string, error) {
	ids := []string{}
	for _, g := range e.tl.Segments {
		if g.ID != moving {
			ids = append(ids, g.ID)
		}
	}
	find := func(id string) (int, error) {
		for i, x := range ids {
			if x == id {
				return i, nil
			}
		}
		return 0, Errf(ExitUsage, "unknown-segment", "segment ids: "+strings.Join(ids, ", ")+" (or --after start)", "no segment %q", id)
	}
	switch {
	case c.Has("after") && c.Has("before"):
		return 0, "", Errf(ExitUsage, "bad-args", "give one of --after or --before", "both --after and --before")
	case c.Has("after"):
		if c.Str("after") == "start" {
			return 0, "to the start", nil
		}
		i, err := find(c.Str("after"))
		return i + 1, "after " + c.Str("after"), err
	case c.Has("before"):
		if c.Str("before") == "end" {
			return len(ids), "to the end", nil
		}
		i, err := find(c.Str("before"))
		return i, "before " + c.Str("before"), err
	}
	if moving != "" {
		return 0, "", Errf(ExitUsage, "bad-args", "say where: --after <id>, --before <id>, --after start or --before end", "edit move needs --after or --before")
	}
	return len(ids), "at the end", nil
}

// applyAdd inserts a span of any source (another take) by range or by words.
func (e *editEnv) applyAdd(c *Ctx) ([]string, string, error) {
	pos, where, err := e.position(c, "")
	if err != nil {
		return nil, "", err
	}
	reason := c.Str("reason")
	if len(c.Args) == 2 { // range: needs --source
		if !c.Has("source") {
			return nil, "", Errf(ExitUsage, "bad-args", "say which source: `cutroom edit add 12.0-20.5 --source take2`", "a range needs --source")
		}
		s, err := sourceByID(e.p, c.Str("source"))
		if err != nil {
			return nil, "", err
		}
		a, b, err := parseRange(c.Args[1], sourceRate(s))
		if err != nil {
			return nil, "", Errf(ExitUsage, "bad-range", "write a range as start-end in source time, e.g. 12.0-20.5", "%v", err)
		}
		id := e.tl.Insert(s.ID, a, b, pos, orStr(reason, "add "+c.Args[1]))
		e.dur[s.ID] = s.Facts.End()
		return []string{fmt.Sprintf("added %s %s–%s as %s %s", s.ID, clock(a), clock(b), id, where)}, fmt.Sprintf("edit add %s %s", s.ID, c.Args[1]), nil
	}
	if !(c.Has("text") || (c.Has("from") && c.Has("to"))) {
		return nil, "", Errf(ExitUsage, "bad-span", `give --from "first words" --to "last words" (or --text, or a range with --source)`, "edit add needs a span")
	}
	var src string
	var i, j int
	if c.Has("text") {
		src, i, j, err = e.resolve(c, c.Str("text"), false, "", -1, -1)
	} else {
		var iLast int
		src, i, iLast, err = e.resolve(c, c.Str("from"), false, "", -1, -1)
		if err == nil {
			_, _, j, err = e.resolve(c, c.Str("to"), false, src, i, iLast)
		}
	}
	if err != nil {
		return nil, "", err
	}
	words := e.tr[src].Words
	a, b := words[i].Start-e.tl.PadBefore, words[j].End+e.tl.PadAfter
	if i > 0 {
		a = e.tl.Boundary(src, words[i-1].End, words[i].Start, clampF(a, words[i-1].End, words[i].Start))
	}
	if j+1 < len(words) {
		b = e.tl.Boundary(src, words[j].End, words[j+1].Start, clampF(b, words[j].End, words[j+1].Start))
	}
	text := speech.Text(words, i, j)
	id := e.tl.Insert(src, a, b, pos, orStr(reason, "add \""+ellipsize(text, 60)+"\""))
	out := []string{fmt.Sprintf("added %s %s–%s as %s %s: %q", src, clock(a), clock(b), id, where, ellipsize(text, 120))}
	if e.tl.Kept(src, (a+b)/2) && e.tl.Index(id) >= 0 {
		for _, g := range e.tl.Segments {
			if g.ID != id && g.Source == src && g.In < b && g.Out > a {
				out = append(out, fmt.Sprintf("WARN %s already contains part of this (segment %s): it will play twice", src, g.ID))
				break
			}
		}
	}
	return out, fmt.Sprintf("edit add %q", ellipsize(text, 60)), nil
}

// applyReplace swaps the kept words of a span for another source's version of them: the original is
// cut by words and the other take's span goes in where it was.
func (e *editEnv) applyReplace(c *Ctx) ([]string, string, error) {
	with := c.Str("with")
	if with == "" {
		return nil, "", Errf(ExitUsage, "bad-args", `name the other take: `+"`cutroom edit replace --text \"…\" --with pickup`", "edit replace needs --with")
	}
	if _, err := sourceByID(e.p, with); err != nil {
		return nil, "", err
	}
	if e.tr[with] == nil {
		return nil, "", Errf(ExitInvalid, "no-transcript", "run `cutroom transcribe "+with+"`", "source %q has not been transcribed", with)
	}
	if !(c.Has("text") || (c.Has("from") && c.Has("to"))) {
		return nil, "", Errf(ExitUsage, "bad-span", `give --text "words" (or --from "first words" --to "last words")`, "edit replace needs the words to swap")
	}
	span := func(keptOnly bool) (string, int, int, error) {
		if c.Has("text") {
			return e.resolve(c, c.Str("text"), keptOnly, "", -1, -1)
		}
		src, i, iLast, err := e.resolve(c, c.Str("from"), keptOnly, "", -1, -1)
		if err != nil {
			return "", 0, 0, err
		}
		_, _, j, err := e.resolve(c, c.Str("to"), keptOnly, src, i, iLast)
		return src, i, j, err
	}
	// 1. the version in the edit now (kept words, any source but --with)
	src, i, j, err := span(true)
	if err != nil {
		return nil, "", err
	}
	if src == with {
		return nil, "", Errf(ExitUsage, "bad-args", "the edit already uses "+with+"'s version of these words", "nothing to replace")
	}
	words := e.tr[src].Words
	// where the original sits: the segment holding its first word, and that segment's index
	sid, sidx := "", len(e.tl.Segments)
	for k, g := range e.tl.Segments {
		if g.Source == src && words[i].Start >= g.In-1e-6 && words[i].Start < g.Out {
			sid, sidx = g.ID, k
			break
		}
	}
	r := e.tl.CutWords(src, words, i, j, "replaced by "+with)
	if r.Removed == 0 {
		return nil, "", Errf(ExitUsage, "nothing-cut", "those words are not in the timeline; `cutroom timeline` shows what is kept", "nothing to replace")
	}
	// 2. the other take's version, inserted where the original was
	old := c.vals["source"]
	c.vals["source"] = []string{with}
	_, wi, wj, err := span(false)
	c.vals["source"] = old
	if err != nil {
		return nil, "", err
	}
	ww := e.tr[with].Words
	a, b := ww[wi].Start-e.tl.PadBefore, ww[wj].End+e.tl.PadAfter
	if wi > 0 {
		a = clampF(a, ww[wi-1].End, ww[wi].Start)
	}
	if wj+1 < len(ww) {
		b = clampF(b, ww[wj].End, ww[wj+1].Start)
	}
	pos := min(sidx, len(e.tl.Segments))
	if k := e.tl.Index(sid); k >= 0 {
		if g := e.tl.Segments[k]; g.Out <= r.From+1e-6 {
			pos = k + 1 // the original's segment kept its start: the new take goes right after it
		} else {
			pos = k // it kept only what follows the cut: the new take goes before it
		}
	}
	text := speech.Text(ww, wi, wj)
	id := e.tl.Insert(with, a, b, pos, "replaces "+src+": \""+ellipsize(text, 60)+"\"")
	return []string{
		fmt.Sprintf("cut %s %s–%s: %q", src, clock(r.From), clock(r.To), ellipsize(r.Text, 100)),
		fmt.Sprintf("in its place: %s %s–%s as %s: %q", with, clock(a), clock(b), id, ellipsize(text, 100)),
	}, fmt.Sprintf("edit replace %q with %s", ellipsize(text, 50), with), nil
}

// soundFrom names the sound each take in the edit uses: "mic" for one source, "mic1 (for take1),
// mic2 (for take2)" when takes differ (T3 r50: the header said "sound from sec-take1" after both
// takes were mapped to their recorders).
func (e *editEnv) soundFrom() string {
	rs := renderSources(e.p)
	var parts, seen []string
	for _, g := range e.tl.Segments {
		if contains(seen, g.Source) {
			continue
		}
		seen = append(seen, g.Source)
		a, _, err := render.AudioFor(e.tl, rs, g.Source)
		if err != nil || a == "" {
			a = "none"
		}
		parts = append(parts, a+" (for "+g.Source+")")
	}
	switch {
	case len(parts) == 0:
		return orNone(e.tl.Audio)
	case len(parts) == 1:
		return strings.SplitN(parts[0], " (", 2)[0]
	}
	return strings.Join(parts, ", ")
}

// isTimeish: "12.5", "0:12.5", "1:02:03" — a time given where words were expected.
func isTimeish(v string) bool {
	v = strings.TrimSpace(v)
	if v == "" {
		return false
	}
	for _, r := range v {
		if !(r >= '0' && r <= '9' || r == '.' || r == ':') {
			return false
		}
	}
	return strings.ContainsAny(v, "0123456789")
}

// mainPicture is the video source most of the edit shows and the source span it covers.
func (e *editEnv) mainPicture() (string, float64, float64) {
	e.tl.Recompute()
	dur := map[string]float64{}
	lo, hi := map[string]float64{}, map[string]float64{}
	best := ""
	for _, g := range e.tl.Segments {
		if s := e.p.Sources[g.Source]; s == nil || s.Facts.Video == nil {
			continue
		}
		if _, ok := lo[g.Source]; !ok {
			lo[g.Source], hi[g.Source] = g.In, g.Out
		}
		lo[g.Source], hi[g.Source] = math.Min(lo[g.Source], g.In), math.Max(hi[g.Source], g.Out)
		dur[g.Source] += g.Out - g.In
		if best == "" || dur[g.Source] > dur[best] {
			best = g.Source
		}
	}
	if best == "" {
		return "", 0, 0
	}
	return best, lo[best], hi[best]
}
