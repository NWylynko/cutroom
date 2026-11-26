package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NWylynko/cutroom/internal/export"
	"github.com/NWylynko/cutroom/internal/render"
)

func init() {
	Register(&Command{
		Name:     "export",
		MaxArgs:  0,
		Category: "render",
		Summary:  "Hand the edit to another editor: OpenTimelineIO (--otio) and/or CMX 3600 EDL (--edl)",
		Help: `Writes the edit as it is now, in whole frames of the timeline's rate:
  --otio   out/<name>.otio — V1 the picture, V2 overlays and V3 burned captions (as their image files;
           text cards as the PNGs cutroom draws, copied to out/<name>_cards/), A1 the sound (a
           recorder's sound at the recorder's own clock). Opens in DaVinci Resolve, Premiere (via
           OTIO), Kdenlive and anything that reads OpenTimelineIO.
  --edl    out/<name>.edl — CMX 3600: a B event per segment (picture and its own sound), or a V and
           an A event when the sound comes from a recorder. Record time from 01:00:00:00. Overlays
           can't be expressed in an EDL; they are listed as comments.
Examples:
  cutroom export --otio --edl
  cutroom export --otio --name final_cut`,
		Flags: []Flag{
			{Name: "otio", Kind: "bool", Help: "Write OpenTimelineIO (.otio JSON)"},
			{Name: "edl", Kind: "bool", Help: "Write a CMX 3600 EDL"},
			{Name: "name", Kind: "string", Default: "edit", Help: "File name (without extension) in out/"},
		},
		Run: runExport,
	})
}

func runExport(c *Ctx) (*Result, error) {
	if !c.Bool("otio") && !c.Bool("edl") {
		return nil, Errf(ExitUsage, "missing-flag", "`cutroom export --otio --edl` (either or both)", "export needs --otio and/or --edl")
	}
	e, err := loadEditEnv(c, false)
	if err != nil {
		return nil, err
	}
	for _, is := range e.validate() {
		if is.Level == "error" {
			return nil, Errf(ExitInvalid, "invalid-timeline", orStr(is.Fix, "`cutroom timeline` lists the problems"), "timeline error: %s", is.Message)
		}
	}
	name := c.Str("name")
	if name == "" || strings.ContainsAny(name, `/\`) || strings.HasPrefix(name, ".") {
		return nil, Errf(ExitUsage, "bad-name", "a plain file name, e.g. --name final_cut (files go in out/)", "--name %q is not a plain file name", name)
	}
	outDir := filepath.Join(e.p.Dir, "out")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	srcs := map[string]export.Source{}
	for id, s := range e.p.Sources {
		srcs[id] = export.Source{Path: e.p.Abs(s), Sync: s.Sync}
	}
	rs := renderSources(e.p)
	audioFor := func(v string) string {
		a, _, err := render.AudioFor(e.tl, rs, v)
		if err != nil || a == "" {
			return v
		}
		return a
	}
	es := export.Events(e.tl, audioFor, srcs)

	// overlays and burned captions as the pictures they are drawn from
	var cards []export.Card
	cardDir := filepath.Join(outDir, name+"_cards")
	if len(e.tl.Overlays) > 0 {
		ps, err := render.Placements(e.tl, e.tr, e.p.Dir, cacheDir(e.p), e.tl.Width, e.tl.Height)
		if err != nil {
			return nil, Errf(ExitInvalid, "overlay-failed", "`cutroom overlay list`; fix or remove it", "%v", err)
		}
		copyCard := func(src, as string) (string, error) {
			if err := os.MkdirAll(cardDir, 0o755); err != nil {
				return "", err
			}
			b, err := os.ReadFile(src)
			if err != nil {
				return "", err
			}
			dst := filepath.Join(cardDir, as)
			return dst, os.WriteFile(dst, b, 0o644)
		}
		for _, p := range ps {
			switch p.Kind {
			case "captions":
				animFrom := -1.0 // entrance frames are left out: the group's card starts where they did
				for i, g := range p.Groups {
					if g.Anim {
						if animFrom < 0 {
							animFrom = g.Start
						}
						continue
					}
					if animFrom >= 0 {
						g.Start, animFrom = animFrom, -1
					}
					dst, err := copyCard(g.PNG, fmt.Sprintf("caption_%03d.png", i+1))
					if err != nil {
						return nil, err
					}
					cards = append(cards, export.Card{Name: g.Text, Path: dst, Start: g.Start, End: g.End, Track: "captions"})
				}
			case "clip":
				cards = append(cards, export.Card{Name: p.ID, Path: p.Clip, Start: p.Start, End: p.End, ClipIn: p.ClipIn, Track: "overlays"})
			default:
				path := p.PNG
				if p.Kind == "text" || p.Kind == "image" {
					if path, err = copyCard(p.PNG, p.ID+".png"); err != nil {
						return nil, err
					}
				}
				cards = append(cards, export.Card{Name: p.ID, Path: path, Start: p.Start, End: p.End, Track: "overlays"})
			}
		}
	}

	var lines, wrote []string
	if c.Bool("otio") {
		b, err := export.OTIO(name, e.tl, es, srcs, cards)
		if err != nil {
			return nil, err
		}
		p := filepath.Join(outDir, name+".otio")
		if err := os.WriteFile(p, b, 0o644); err != nil {
			return nil, err
		}
		wrote = append(wrote, relToCwd(p))
		tracks := "V1 picture, A1 sound"
		if len(cards) > 0 {
			tracks = "V1 picture, V2/V3 overlays and captions (" + relToCwd(cardDir) + "), A1 sound"
		}
		lines = append(lines, fmt.Sprintf("wrote %s: %d clips, %s", relToCwd(p), len(es), tracks))
	}
	if c.Bool("edl") {
		p := filepath.Join(outDir, name+".edl")
		if err := os.WriteFile(p, []byte(export.EDL(name, e.tl, es, srcs, cards)), 0o644); err != nil {
			return nil, err
		}
		wrote = append(wrote, relToCwd(p))
		n := 0
		for _, ev := range es {
			n++
			if ev.Audio != ev.Source {
				n++
			}
		}
		note := ""
		if len(e.tl.Overlays) > 0 {
			note = " (overlays listed as comments: an EDL can't carry them)"
		}
		lines = append(lines, fmt.Sprintf("wrote %s: %d events, CMX 3600, %s fps non-drop, record from 01:00:00:00%s", relToCwd(p), n, strings.TrimSuffix(e.tl.FPS, "/1"), note))
	}
	lines = append(lines, fmt.Sprintf("  %s, %d segment(s) in whole frames; sources referenced where they are (not copied)", secs(e.tl.Duration()), len(es)))
	return &Result{Text: strings.Join(lines, "\n"), Data: map[string]any{"files": wrote, "events": es}, Next: []string{"cutroom status"}}, nil
}
