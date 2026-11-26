package cli

import (
	"fmt"
	"strings"

	"github.com/NWylynko/cutroom/internal/raster"
	"github.com/NWylynko/cutroom/internal/render"
	"github.com/NWylynko/cutroom/internal/timeline"
)

func init() {
	flags := []Flag{}
	var names []string
	for _, f := range timeline.StyleFields {
		help := f.Help
		if f.For != "" {
			help = f.For + ": " + help
		}
		flags = append(flags, Flag{Name: f.Name, Kind: "string", Help: help + ` ("default" unsets it)`})
		names = append(names, f.Name)
	}
	flags = append(flags,
		Flag{Name: "reset", Kind: "bool", Help: "Back to cutroom's default look (for what you name)"},
		Flag{Name: "dry-run", Kind: "bool", Help: "Show the result without saving"})
	setFlags := append(append([]string{}, names...), "reset", "dry-run")
	Register(&Command{
		Name:     "style",
		Args:     "[captions|cards|<overlay id>|fonts|save|use|looks] [name]",
		MaxArgs:  2,
		Category: "edit",
		Summary:  "The look of burned-in captions and text cards: font, size, colours, box, outline, place",
		Help: `cutroom has a default look, but none of it is fixed: set what you want, keep the rest.
  style                                    both looks: what you set, and the defaults under it
  style captions --font "Arial Black" --size 0.08 --color white --outline black --box none
  style captions --highlight "#ffd400" --words 2 --case upper --y 0.62
  style cards --box "#f3f6f7f2" --color "#101718" --accent "#307351" --radius 0.4
  style o3 --x 0.3 --y 0.2 --align left    one text card (over the card style)
  style captions --color default           unset one field; --reset unsets them all
  style fonts [avenir]                     installed fonts you can name with --font
  style save channel                       keep this project's caption + card look as "channel"
  style use channel                        give another project (or short) that look in one call;
                                           also: cutroom short … --look channel
  style looks                              saved looks (a path like looks/brand.json works too)
Colours: #rrggbb, #rrggbbaa (aa = opacity, e.g. #00000099) or a name; "none" turns a box, outline,
shadow, accent or highlight off. Sizes are fractions of the frame's short side; --x/--y put the centre
of the box at that point of the frame (0–1 across, down); --pad/--radius/--outline-width are fractions
of the font size. Changes are undoable (` + "`cutroom edit undo`" + `) and the next render draws them;
look first: ` + "`cutroom look --output <time>`" + `. On a vertical frame, check warns if text
reaches where the apps draw their own buttons (L41).`,
		Flags: flags,
		Subs: map[string]Sub{
			"captions": {Usage: "  style captions [--font …] [--size …] [--color …] [--highlight …] [--words …] …   burned-in captions", Flags: setFlags},
			"cards":    {Usage: "  style cards [--font …] [--box …] [--accent …] …   every text card (each can override: style <id>)", Flags: setFlags},
			"fonts":    {Usage: "  style fonts [query]   installed fonts (name them with --font)"},
			"save":     {Usage: "  style save <name|file.json>   keep this project's caption and card look (fonts by file are stored as absolute paths)", Flags: []string{"dry-run"}},
			"use":      {Usage: "  style use <name|file.json>    replace this project's caption and card look with a saved one (undoable; per-card styles stay)", Flags: []string{"dry-run"}},
			"looks":    {Usage: "  style looks   saved looks and what each sets"},
		},
		Run: runStyle,
	})
}

func runStyle(c *Ctx) (*Result, error) {
	what := ""
	if len(c.Args) > 0 {
		what = c.Args[0]
	}
	if what == "fonts" {
		return styleFonts(c)
	}
	if what == "looks" {
		names := lookNames()
		if len(names) == 0 {
			return &Result{Text: fmt.Sprintf("no saved looks in %s: set one up in a project (`cutroom style captions …`), then `cutroom style save <name>`", looksDir())}, nil
		}
		lines := []string{fmt.Sprintf("saved looks (%s):", looksDir())}
		for _, n := range names {
			lk, _, err := loadLook(n)
			if err != nil {
				lines = append(lines, fmt.Sprintf("  %-14s (unreadable: %v)", n, err))
				continue
			}
			lines = append(lines, fmt.Sprintf("  %-14s %s", n, lookSummary(lk)))
		}
		return &Result{Text: strings.Join(lines, "\n"), Next: []string{"cutroom style use <name>"}}, nil
	}
	e, err := loadEditEnv(c, false)
	if err != nil {
		return nil, err
	}
	if what == "save" || what == "use" {
		if len(c.Args) != 2 {
			return nil, Errf(ExitUsage, "bad-args", fmt.Sprintf("`cutroom style %s channel` (a name, or a file like looks/brand.json)", what), "style %s takes a name", what)
		}
		if anyStyleFlag(c) {
			return nil, Errf(ExitUsage, "bad-flags", "save/use the look first, then change fields with `cutroom style captions …`", "style %s takes no style fields", what)
		}
		name := c.Args[1]
		if what == "save" {
			if c.Bool("dry-run") {
				return &Result{Text: fmt.Sprintf("would save %s: %s\nDRY RUN: nothing was written", lookPath(name), lookSummary(Look{e.tl.CaptionStyle, e.tl.CardStyle}))}, nil
			}
			p, err := saveLook(name, e.tl, e.p.Dir)
			if err != nil {
				return nil, err
			}
			lk, _, _ := loadLook(name)
			return &Result{Text: fmt.Sprintf("saved look %q to %s\n  %s", name, relToCwd(p), lookSummary(lk)),
				Next: []string{fmt.Sprintf("cutroom style use %s --project <another>   (or: cutroom short … --look %s)", name, name)}}, nil
		}
		lk, p, err := loadLook(name)
		if err != nil {
			return nil, err
		}
		if err := render.CheckStyle(lk.Captions.Over(timeline.Style{}), e.p.Dir); err != nil {
			return nil, Errf(ExitUsage, "bad-font", "install the font, or change the look's font (`cutroom style captions --font …`, then save again)", "look %q: %v", name, err)
		}
		if err := render.CheckStyle(lk.Cards.Over(timeline.Style{}), e.p.Dir); err != nil {
			return nil, Errf(ExitUsage, "bad-font", "install the font, or change the look's font (`cutroom style cards --font …`, then save again)", "look %q: %v", name, err)
		}
		applyLook(e.tl, lk)
		lines := []string{fmt.Sprintf("look %q (%s): %s", name, relToCwd(p), lookSummary(lk))}
		if !lk.Captions.Empty() && e.tl.Captions() == nil {
			lines = append(lines, "note: no captions are burned in yet — the look applies once they are: `cutroom captions --burn`")
		}
		for _, is := range e.validate() {
			if strings.HasPrefix(is.Code, "overlay-") {
				lines = append(lines, fmt.Sprintf("WARN[%s] %s → %s", is.Code, is.Message, is.Fix))
			}
		}
		if c.Bool("dry-run") {
			lines = append(lines, "DRY RUN: nothing was changed; run the same command without --dry-run to apply it")
		} else {
			if err := timeline.Checkpoint(e.p.Dir, cacheDir(e.p), "style use "+name); err != nil {
				return nil, err
			}
			if err := e.tl.Save(e.p.Dir); err != nil {
				return nil, err
			}
		}
		return &Result{Text: strings.Join(lines, "\n"), Next: []string{"cutroom look --overlays   (see it)", "cutroom render --preview"}}, nil
	}
	if what == "" {
		if anyStyleFlag(c) {
			return nil, Errf(ExitUsage, "missing-target", "`cutroom style captions --color yellow`, `cutroom style cards …` or `cutroom style o3 …`", "say what to style: captions, cards or an overlay id")
		}
		lines := []string{"captions:"}
		lines = append(lines, styleLines(e.tl.CaptionStyle, render.CaptionStyle(e.tl), "captions", e.p.Dir)...)
		lines = append(lines, "cards (lower-third defaults; title/top/corner differ in size and alignment):")
		lines = append(lines, styleLines(e.tl.CardStyle, e.tl.CardStyle.Over(render.CardDefaults("lower-third")), "cards", e.p.Dir)...)
		for _, o := range e.tl.Overlays {
			if o.Style != nil && !o.Style.Empty() {
				lines = append(lines, fmt.Sprintf("%s (%q) sets its own: %s", o.ID, o.Text, setSummary(o.Style)))
			}
		}
		return &Result{Text: strings.Join(lines, "\n"), Data: map[string]any{"captions": render.CaptionStyle(e.tl), "card_style": e.tl.CardStyle},
			Next: []string{"cutroom style captions --help"}}, nil
	}
	var target **timeline.Style
	var full func() timeline.Style
	kind := "cards"
	switch what {
	case "captions":
		target, kind = &e.tl.CaptionStyle, "captions"
		full = func() timeline.Style { return render.CaptionStyle(e.tl) }
	case "cards":
		target = &e.tl.CardStyle
		full = func() timeline.Style { return e.tl.CardStyle.Over(render.CardDefaults("lower-third")) }
	default:
		k := e.tl.OverlayByID(what)
		if k < 0 {
			return nil, Errf(ExitUsage, "no-target", "style captions, cards, or a text overlay's id (`cutroom overlay list`)", "nothing called %q to style", what)
		}
		o := &e.tl.Overlays[k]
		switch o.Kind {
		case "captions":
			target, kind = &e.tl.CaptionStyle, "captions"
			full = func() timeline.Style { return render.CaptionStyle(e.tl) }
		case "text":
			target = &o.Style
			full = func() timeline.Style { return render.CardStyle(e.tl, *o) }
		case "image", "clip":
			target, kind = &o.Style, "media"
			full = func() timeline.Style {
				if o.Style == nil {
					return timeline.Style{}
				}
				return *o.Style
			}
		default:
			return nil, Errf(ExitUsage, "not-text", "style captions, cards or a text/image/clip overlay", "%s can't be styled", o.ID)
		}
	}
	if !anyStyleFlag(c) && !c.Bool("reset") {
		lines := styleLines(*target, full(), kind, e.p.Dir)
		return &Result{Text: strings.Join(lines, "\n"), Data: map[string]any{"style": full()}, Next: []string{fmt.Sprintf("cutroom style %s --help", what)}}, nil
	}
	st := timeline.Style{}
	if *target != nil {
		st = **target
	}
	var changed []string
	if c.Bool("reset") {
		st = timeline.Style{}
		changed = append(changed, "reset")
	}
	for _, f := range timeline.StyleFields {
		if !c.Has(f.Name) {
			continue
		}
		if kind == "media" && !contains(render.MediaStyleFields, f.Name) {
			return nil, Errf(ExitUsage, "wrong-field", "an image or clip takes "+strings.Join(render.MediaStyleFields, ", ")+" (--radius and --outline-width are fractions of its short side)", "--%s doesn't apply to an image or clip", f.Name)
		}
		if f.For != "" && f.For != kind && kind != "media" {
			return nil, Errf(ExitUsage, "wrong-field", fmt.Sprintf("--%s is for %s", f.Name, f.For), "--%s doesn't apply to %s", f.Name, kind)
		}
		if err := st.Set(f.Name, c.Str(f.Name)); err != nil {
			return nil, Errf(ExitUsage, "bad-style", "see `cutroom style --help`", "%v", err)
		}
		changed = append(changed, fmt.Sprintf("%s=%s", f.Name, c.Str(f.Name)))
	}
	if err := render.CheckStyle(st, e.p.Dir); err != nil {
		return nil, Errf(ExitUsage, "bad-font", "`cutroom style fonts` lists installed fonts; or give a .ttf/.otf file", "%v", err)
	}
	if st.Empty() {
		*target = nil
	} else {
		*target = &st
	}
	label := fmt.Sprintf("style %s %s", what, strings.Join(changed, " "))
	lines := []string{fmt.Sprintf("%s: %s", what, strings.Join(changed, ", "))}
	if *target != nil {
		lines = append(lines, "now set: "+setSummary(*target)+" (the rest default; `cutroom style "+what+"` shows every field)")
	} else {
		lines = append(lines, "all back to cutroom's default look")
	}
	for _, f := range timeline.StyleFields {
		if f.Kind == "font" && c.Has(f.Name) && c.Str(f.Name) != "default" {
			lines = append(lines, fmt.Sprintf("%s → %s", f.Name, render.FontName(c.Str(f.Name), e.p.Dir)))
		}
	}
	if kind == "captions" && e.tl.Captions() == nil {
		lines = append(lines, "note: no captions are burned in yet — the style applies once they are: `cutroom captions --burn`")
	}
	if c.Has("animate") || c.Has("animate-time") {
		lines = append(lines, "`look` draws text settled; to see the motion, render and look at the frames: `cutroom render --preview`, then `cutroom look --file out/preview.mp4 <start>-<start+0.5> --count 8`")
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
	at := 0.0
	for _, o := range e.tl.Overlays {
		if a, b, ok, _ := e.tl.Window(o); ok && (kind == "captions") == (o.Kind == "captions") && (o.Kind == "captions" || o.Kind == "text") {
			at = (a + b) / 2
			if o.Kind == "captions" {
				at = a + 1
			}
			break
		}
	}
	return &Result{Text: strings.Join(lines, "\n"), Data: map[string]any{"style": full()},
		Next: []string{fmt.Sprintf("cutroom look --output %.2f   (see it)", at), "cutroom render --preview"}}, nil
}

func anyStyleFlag(c *Ctx) bool {
	for _, f := range timeline.StyleFields {
		if c.Has(f.Name) {
			return true
		}
	}
	return false
}

// styleLines prints each field that applies: its value and whether it was set or is the default.
func styleLines(set *timeline.Style, full timeline.Style, kind, dir string) []string {
	var out []string
	for _, f := range timeline.StyleFields {
		if (f.For != "" && f.For != kind && kind != "media") || (kind == "media" && !contains(render.MediaStyleFields, f.Name)) {
			continue
		}
		v := full.Get(f.Name)
		mark := "default"
		if set.Get(f.Name) != "" {
			mark = "set"
		}
		if v == "" {
			v = "-"
			switch f.Name {
			case "x":
				v = "centred"
				if kind == "cards" {
					v = "from --layout"
				}
			case "y":
				v = "the caption band"
				if kind == "cards" {
					v = "from --layout"
				}
			case "width":
				v = "the safe width"
			case "outline-width":
				v = "0.08 (with an outline)"
			case "radius":
				v = "0 (square)"
			case "words":
				v = "3 (a 4th only to avoid a weak ending; set it to make it a hard limit)"
			}
		}
		if f.Kind == "font" && v != "-" {
			v = fmt.Sprintf("%s → %s", v, render.FontName(v, dir))
		}
		out = append(out, fmt.Sprintf("  %-14s %-40s %s", f.Name, v, mark))
	}
	return out
}

func setSummary(s *timeline.Style) string {
	var parts []string
	for _, f := range timeline.StyleFields {
		if v := s.Get(f.Name); v != "" {
			parts = append(parts, f.Name+"="+v)
		}
	}
	return strings.Join(parts, " ")
}

func styleFonts(c *Ctx) (*Result, error) {
	q := ""
	if len(c.Args) > 1 {
		q = strings.ToLower(c.Args[1])
	}
	lines := []string{"built in (always available): bold, regular, mono"}
	n, shown := 0, 0
	seen := map[string]bool{}
	for _, f := range raster.InstalledFonts() {
		if seen[f.Name] || (q != "" && !strings.Contains(strings.ToLower(f.Name), q)) {
			continue
		}
		seen[f.Name] = true
		n++
		if shown < 60 {
			lines = append(lines, fmt.Sprintf("  %s", f.Name))
			shown++
		}
	}
	if n > shown {
		lines = append(lines, fmt.Sprintf("  … %d more: narrow it, e.g. `cutroom style fonts helvetica`", n-shown))
	}
	if n == 0 {
		lines = append(lines, fmt.Sprintf("no installed font matches %q (looked in %s); any .ttf/.otf/.ttc file works too: --font path/to/font.ttf", q, strings.Join(raster.FontDirs(), ", ")))
	}
	lines = append(lines, `use one: cutroom style captions --font "<name>"`)
	return &Result{Text: strings.Join(lines, "\n")}, nil
}
