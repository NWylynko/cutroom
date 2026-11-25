package cli

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/cutroom/cutroom/internal/media"
	"github.com/cutroom/cutroom/internal/render"
	"github.com/cutroom/cutroom/internal/timeline"
)

func init() {
	Register(&Command{
		Name:     "overlay",
		Args:     "<add|set|list|rm> [text|image|id] [content]",
		MinArgs:  1,
		MaxArgs:  3,
		Category: "edit",
		Summary:  "Text cards and images over the edit, anchored to quoted words (they follow re-cuts)",
		Help: `An overlay appears when its anchor words are said and lasts --duration seconds (or until the --to
words). It is anchored to the words, not to a time, so later cuts never move it off them.
  overlay add text "Install Linux" --at "Then install" --duration 3
  overlay add text "Nick Wylie" --sub "Builds personal software" --at "Hi, today" --to "home server."
  overlay add image logo.png --at "Finally" --duration 4 --layout corner --size 0.2
  overlay add clip terminal.mp4 --at "open the terminal" --to "update command." [--clip-in 2.5]
                                  B-roll: the clip covers the frame (--layout full, default) or sits
                                  in a corner (--layout corner --size 0.35); muted — the voice goes on
  overlay set o2 --duration 2.2   change one in place (also --to, --at, --layout, --text, --sub, --fade, --size)
  overlay list                    every overlay with its output time, or why it won't show
  overlay rm o2
Layouts: lower-third (default for text), title, top, corner (default for images), center;
full (clips only, their default).
Text is drawn by cutroom (no ffmpeg text filters needed); look at it before rendering:
` + "`cutroom look --output <time>`" + ` shows the edit with its overlays.`,
		Flags: []Flag{
			{Name: "at", Kind: "string", Help: "Words it appears on (quoted from the transcript)"},
			{Name: "to", Kind: "string", Help: "Words it disappears after (instead of --duration)"},
			{Name: "duration", Kind: "float", Help: "Seconds on screen (default 4)"},
			{Name: "sub", Kind: "string", Help: "text: a smaller second line"},
			{Name: "layout", Kind: "string", Help: "lower-third | title | top | corner | center"},
			{Name: "size", Kind: "float", Help: "image: width as a fraction of the frame (default 0.25)"},
			{Name: "fade", Kind: "float", Default: "0.2", Help: "Fade in/out seconds (0 = cut in and out)"},
			{Name: "near", Kind: "string", Help: "Pick the --at match closest to this source time when words repeat"},
			{Name: "occurrence", Kind: "int", Help: "Pick the Nth --at match when words repeat"},
			{Name: "source", Kind: "string", Help: "Only match words in this source"},
			{Name: "text", Kind: "string", Help: "set: new text for a text overlay"},
			{Name: "clip-in", Kind: "float", Default: "0", Help: "clip: where in the clip to start (s)"},
			{Name: "dry-run", Kind: "bool", Help: "Show the result without saving"},
		},
		Subs: map[string]Sub{
			"set": {Usage: `  overlay set o2 --duration 2.2        keep it on screen 2.2 s
  overlay set o2 --to "on it."         … or until these words end
  overlay set o2 --at "Then install"   move it to other words
  overlay set o2 --layout top --text "Step 2" --sub "…" --fade 0
Changes only what you give; the id stays.`,
				Flags: []string{"duration", "to", "at", "near", "occurrence", "source", "layout", "text", "sub", "fade", "size", "clip-in", "dry-run"}},
			"add": {Usage: `  overlay add text "Install Linux" --at "Then install" [--duration 3 | --to "on it."] [--sub "…"]
  overlay add image logo.png --at "Finally" --duration 4 [--layout corner] [--size 0.2]
  overlay add clip broll.mp4 --at "open the terminal" [--to "…" | --duration 3] [--clip-in 2.5] [--layout full|corner] [--size 0.35]`,
				Flags: []string{"at", "to", "duration", "sub", "layout", "size", "fade", "clip-in", "near", "occurrence", "source", "dry-run"}},
			"list": {Usage: `  overlay list   every overlay: id, kind, output time, layout, anchor words`},
			"rm":   {Usage: `  overlay rm o2  remove an overlay (undo-able with ` + "`cutroom edit undo`" + `)`, Flags: []string{"dry-run"}},
		},
		Run: runOverlay,
	})
}

func runOverlay(c *Ctx) (*Result, error) {
	op := c.Args[0]
	switch op {
	case "add", "set", "list", "rm":
	default:
		fix := "ops: add, set, list, rm"
		if s := suggest(op, []string{"add", "set", "list", "rm"}); s != "" {
			fix = fmt.Sprintf("did you mean `cutroom overlay %s`? (%s)", s, fix)
		}
		return nil, Errf(ExitUsage, "unknown-op", fix, "no overlay op %q", op)
	}
	e, err := loadEditEnv(c, op == "add")
	if err != nil {
		return nil, err
	}
	var lines []string
	label := ""
	switch op {
	case "list":
		return overlayList(e), nil
	case "rm":
		if len(c.Args) != 2 {
			return nil, Errf(ExitUsage, "bad-args", "`cutroom overlay rm o1` (ids: `cutroom overlay list`)", "overlay rm takes one id")
		}
		k := e.tl.OverlayByID(c.Args[1])
		if k < 0 {
			return nil, Errf(ExitUsage, "no-overlay", "`cutroom overlay list` shows the ids", "no overlay %q", c.Args[1])
		}
		o := e.tl.Overlays[k]
		e.tl.Overlays = append(e.tl.Overlays[:k], e.tl.Overlays[k+1:]...)
		label = "overlay rm " + o.ID
		lines = append(lines, fmt.Sprintf("removed %s (%s %q)", o.ID, o.Kind, ovDesc(o)))
	case "set":
		if len(c.Args) != 2 {
			return nil, Errf(ExitUsage, "bad-args", "`cutroom overlay set o1 --duration 3` (ids: `cutroom overlay list`)", "overlay set takes one id")
		}
		k := e.tl.OverlayByID(c.Args[1])
		if k < 0 {
			return nil, Errf(ExitUsage, "no-overlay", "`cutroom overlay list` shows the ids", "no overlay %q", c.Args[1])
		}
		before := e.tl.Overlays[k]
		o, changed, err := e.overlaySet(c, before)
		if err != nil {
			return nil, err
		}
		if len(changed) == 0 {
			return nil, Errf(ExitUsage, "nothing-to-set", "give what to change, e.g. --duration 3, --to \"…\", --at \"…\", --layout top, --text \"…\"", "overlay set %s: nothing to change", o.ID)
		}
		e.tl.Overlays[k] = o
		a0, b0, ok0, _ := e.tl.Window(before)
		a, b, ok, why := e.tl.Window(o)
		if !ok {
			return nil, Errf(ExitUsage, "overlay-not-shown", "anchor it to words the edit keeps (`cutroom timeline` shows them)", "the overlay wouldn't show: %s", why)
		}
		was := "not shown"
		if ok0 {
			was = fmt.Sprintf("%s–%s", clock(a0), clock(b0))
		}
		label = fmt.Sprintf("overlay set %s %s", o.ID, strings.Join(changed, " "))
		lines = append(lines, fmt.Sprintf("set %s: %s; now output %s–%s (%s), was %s", o.ID, strings.Join(changed, ", "), clock(a), clock(b), secs(b-a), was))
		lines = append(lines, fmt.Sprintf("look at it: `cutroom look --output %.2f`", (a+b)/2))
	case "add":
		o, err := e.overlayFromFlags(c)
		if err != nil {
			return nil, err
		}
		id := e.tl.AddOverlay(o)
		o.ID = id
		a, b, ok, why := e.tl.Window(o)
		if !ok {
			return nil, Errf(ExitUsage, "overlay-not-shown", "anchor it to words the edit keeps (`cutroom timeline` shows them)", "the overlay wouldn't show: %s", why)
		}
		label = fmt.Sprintf("overlay add %s %q --at %q", o.Kind, ovDesc(o), o.AnchorText)
		lines = append(lines, fmt.Sprintf("added %s: %s %q, %s, output %s–%s (%s), appears on %q (%s %s)",
			id, o.Kind, ovDesc(o), o.Layout, clock(a), clock(b), secs(b-a), o.AnchorText, o.Source, clock(o.At)))
		lines = append(lines, fmt.Sprintf("look at it: `cutroom look --output %.2f`", (a+b)/2))
	}
	for _, is := range e.validate() {
		if strings.HasPrefix(is.Code, "overlay-") {
			lines = append(lines, fmt.Sprintf("WARN[%s] %s → %s", is.Code, is.Message, is.Fix))
		}
	}
	if c.Bool("dry-run") {
		lines = append(lines, "DRY RUN: nothing was changed; run the same command without --dry-run to apply it")
	} else {
		if err := timeline.Checkpoint(e.p.Dir, cacheDir(e.p), label); err != nil {
			return nil, err
		}
		if err := e.tl.Save(e.p.Dir); err != nil {
			return nil, err
		}
	}
	next := []string{"cutroom overlay list", "cutroom render --preview"}
	return &Result{Text: strings.Join(lines, "\n"), Data: map[string]any{"overlays": e.tl.Overlays}, Next: next}, nil
}

func ovDesc(o timeline.Overlay) string {
	if o.Kind == "captions" {
		return "burned-in captions (the whole edit)"
	}
	if o.Kind == "image" || o.Kind == "clip" {
		return o.Path
	}
	return o.Text
}

func (e *editEnv) overlayFromFlags(c *Ctx) (timeline.Overlay, error) {
	if len(c.Args) != 3 {
		return timeline.Overlay{}, Errf(ExitUsage, "bad-args", `overlay add text "Install Linux" --at "Then install"  or  overlay add image logo.png --at "…"`, "overlay add takes a kind (text|image) and its content")
	}
	o := timeline.Overlay{Kind: c.Args[1], Fade: c.Float("fade"), Duration: 4, Layout: c.Str("layout"), Sub: c.Str("sub"), Size: c.Float("size")}
	switch o.Kind {
	case "text":
		o.Text = c.Args[2]
		if strings.TrimSpace(o.Text) == "" {
			return o, Errf(ExitUsage, "bad-args", `give the text: overlay add text "Install Linux" --at "…"`, "empty text")
		}
		if o.Layout == "" {
			o.Layout = "lower-third"
		}
	case "image", "clip":
		p := c.Args[2]
		abs := userPath(p)
		if _, err := os.Stat(abs); err != nil {
			return o, Errf(ExitUsage, "file-not-found", "give the file's path (relative to your current folder, like any path)", "%s not found", p)
		}
		if rel, err := filepath.Rel(e.p.Dir, abs); err == nil && !strings.HasPrefix(rel, "..") {
			p = rel
		} else {
			p = abs
		}
		o.Path = p
		if o.Sub != "" {
			return o, Errf(ExitUsage, "bad-flag", "--sub is for text overlays", "%s overlays have no --sub", o.Kind)
		}
		if o.Layout == "" {
			o.Layout = "corner"
		}
		if o.Kind == "clip" {
			f, err := media.Probe(abs, media.Options{})
			if err != nil || f.Video == nil {
				return o, Errf(ExitUsage, "not-a-video", "give a video file (mp4, mov, …) for a clip overlay", "%s has no video stream", p)
			}
			o.ClipDur, o.ClipW, o.ClipH = f.Video.Duration, f.Video.Width, f.Video.Height
			if o.ClipDur <= 0 {
				o.ClipDur = f.Duration
			}
			o.ClipIn = c.Float("clip-in")
			if o.ClipIn < 0 || o.ClipIn >= o.ClipDur {
				return o, Errf(ExitUsage, "bad-flag-value", fmt.Sprintf("--clip-in is a time inside the clip (0–%.2f s)", o.ClipDur), "--clip-in %.2f is outside the clip", o.ClipIn)
			}
			if !c.Has("layout") {
				o.Layout = "full" // a cutaway over the whole frame
			}
			if !c.Has("fade") {
				o.Fade = 0 // B-roll cuts in and out
			}
			if !c.Has("duration") && !c.Has("to") {
				o.Duration = math.Min(4, o.ClipDur-o.ClipIn)
			}
		}
	default:
		return o, Errf(ExitUsage, "bad-kind", "overlay add text \"…\", overlay add image <file> or overlay add clip <video>", "no overlay kind %q (text, image or clip)", o.Kind)
	}
	if !contains(render.Layouts, o.Layout) && !(o.Kind == "clip" && o.Layout == "full") {
		return o, Errf(ExitUsage, "bad-layout", "layouts: "+strings.Join(render.Layouts, ", ")+" (clips also: full)", "no layout %q", o.Layout)
	}
	if o.Fade < 0 || o.Fade > 2 {
		return o, Errf(ExitUsage, "bad-flag-value", "--fade 0–2 seconds", "--fade %.2f out of range", o.Fade)
	}
	if !c.Has("at") {
		return o, Errf(ExitUsage, "missing-flag", `--at "words it appears on" (quoted from the transcript)`, "overlay add needs --at")
	}
	src, i, _, err := e.resolve(c, c.Str("at"), true, "", 0, 0)
	if err != nil {
		return o, err
	}
	ws := e.tr[src].Words
	o.Source, o.At, o.AnchorText = src, ws[i].Start, c.Str("at")
	if c.Has("to") {
		if c.Has("duration") {
			return o, Errf(ExitUsage, "bad-flags", "give --to or --duration, not both", "--to and --duration both given")
		}
		_, _, j, err := e.resolve(c, c.Str("to"), true, src, i, i)
		if err != nil {
			return o, err
		}
		o.Until, o.Duration = ws[j].End, 0
	} else if c.Has("duration") {
		o.Duration = c.Float("duration")
		if o.Duration <= 0 {
			return o, Errf(ExitUsage, "bad-flag-value", "--duration in seconds, e.g. 3", "--duration must be > 0")
		}
	}
	return o, nil
}

func overlayList(e *editEnv) *Result {
	if len(e.tl.Overlays) == 0 {
		return &Result{Text: "no overlays yet", Data: map[string]any{"overlays": []any{}},
			Next: []string{`cutroom overlay add text "…" --at "words it appears on"`}}
	}
	var lines []string
	type row struct {
		timeline.Overlay
		Start float64 `json:"output_start"`
		End   float64 `json:"output_end"`
		Shown bool    `json:"shown"`
		Why   string  `json:"why,omitempty"`
	}
	var rows []row
	for _, o := range e.tl.Overlays {
		a, b, ok, why := e.tl.Window(o)
		rows = append(rows, row{o, a, b, ok, why})
		when := fmt.Sprintf("%s–%s", clock(a), clock(b))
		if !ok {
			when = "NOT SHOWN: " + why
		}
		sub := ""
		if o.Sub != "" {
			sub = fmt.Sprintf(" / %q", o.Sub)
		}
		if o.Kind == "captions" {
			lines = append(lines, fmt.Sprintf("%-4s %-5s burned in over the whole edit (%s), groups from the cut at render time — `cutroom captions --burn --dry-run` shows them", o.ID, "captions", when))
			continue
		}
		lines = append(lines, fmt.Sprintf("%-4s %-5s %q%s  %s  %s  on %q (%s %s)", o.ID, o.Kind, ovDesc(o), sub, o.Layout, when, o.AnchorText, o.Source, clock(o.At)))
	}
	return &Result{Text: strings.Join(lines, "\n"), Data: map[string]any{"overlays": rows}, Next: []string{"cutroom look --output <time>   (see one)"}}
}

// overlaySet applies the given flags to an overlay; changed lists what changed.
func (e *editEnv) overlaySet(c *Ctx, o timeline.Overlay) (timeline.Overlay, []string, error) {
	var ch []string
	if o.Kind == "captions" {
		return o, nil, Errf(ExitUsage, "bad-kind", "burned-in captions follow the cut; remove them with `cutroom overlay rm "+o.ID+"`", "captions have nothing to set")
	}
	if c.Has("at") {
		src, i, _, err := e.resolve(c, c.Str("at"), true, "", 0, 0)
		if err != nil {
			return o, nil, err
		}
		o.Source, o.At, o.AnchorText = src, e.tr[src].Words[i].Start, c.Str("at")
		if o.Until > 0 && o.Until <= o.At { // its old end is before the new start
			o.Until, o.Duration = 0, 4
		}
		ch = append(ch, fmt.Sprintf("--at %q", o.AnchorText))
	}
	if c.Has("to") && c.Has("duration") {
		return o, nil, Errf(ExitUsage, "bad-flags", "give --to or --duration, not both", "--to and --duration both given")
	}
	if c.Has("to") {
		ws := e.tr[o.Source].Words
		first := 0
		for first < len(ws) && ws[first].Start < o.At {
			first++
		}
		_, _, j, err := e.resolve(c, c.Str("to"), true, o.Source, first, first)
		if err != nil {
			return o, nil, err
		}
		o.Until, o.Duration = ws[j].End, 0
		ch = append(ch, fmt.Sprintf("--to %q", c.Str("to")))
	}
	if c.Has("duration") {
		if c.Float("duration") <= 0 {
			return o, nil, Errf(ExitUsage, "bad-flag-value", "--duration in seconds, e.g. 3", "--duration must be > 0")
		}
		o.Duration, o.Until = c.Float("duration"), 0
		ch = append(ch, fmt.Sprintf("--duration %g", o.Duration))
	}
	if c.Has("layout") {
		if !contains(render.Layouts, c.Str("layout")) && !(o.Kind == "clip" && c.Str("layout") == "full") {
			return o, nil, Errf(ExitUsage, "bad-layout", "layouts: "+strings.Join(render.Layouts, ", "), "no layout %q", c.Str("layout"))
		}
		o.Layout = c.Str("layout")
		ch = append(ch, "--layout "+o.Layout)
	}
	if c.Has("text") || c.Has("sub") {
		if o.Kind != "text" {
			return o, nil, Errf(ExitUsage, "bad-flag", "--text/--sub are for text overlays", "%s is an image overlay", o.ID)
		}
		if c.Has("text") {
			if strings.TrimSpace(c.Str("text")) == "" {
				return o, nil, Errf(ExitUsage, "bad-flag-value", `--text "new text"`, "empty --text")
			}
			o.Text = c.Str("text")
			ch = append(ch, fmt.Sprintf("--text %q", o.Text))
		}
		if c.Has("sub") {
			o.Sub = c.Str("sub")
			ch = append(ch, fmt.Sprintf("--sub %q", o.Sub))
		}
	}
	if c.Has("fade") {
		if f := c.Float("fade"); f < 0 || f > 2 {
			return o, nil, Errf(ExitUsage, "bad-flag-value", "--fade 0–2 seconds", "--fade %.2f out of range", f)
		}
		o.Fade = c.Float("fade")
		ch = append(ch, fmt.Sprintf("--fade %g", o.Fade))
	}
	if c.Has("clip-in") {
		if o.Kind != "clip" {
			return o, nil, Errf(ExitUsage, "bad-flag", "--clip-in is for clip overlays", "%s is a %s overlay", o.ID, o.Kind)
		}
		if v := c.Float("clip-in"); v < 0 || v >= o.ClipDur {
			return o, nil, Errf(ExitUsage, "bad-flag-value", fmt.Sprintf("--clip-in is a time inside the clip (0–%.2f s)", o.ClipDur), "--clip-in %.2f is outside the clip", v)
		}
		o.ClipIn = c.Float("clip-in")
		ch = append(ch, fmt.Sprintf("--clip-in %g", o.ClipIn))
	}
	if c.Has("size") {
		if o.Kind == "text" {
			return o, nil, Errf(ExitUsage, "bad-flag", "--size is for image and clip overlays", "%s is a text overlay", o.ID)
		}
		o.Size = c.Float("size")
		ch = append(ch, fmt.Sprintf("--size %g", o.Size))
	}
	return o, ch, nil
}
