package cli

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/NWylynko/cutroom/internal/project"
	"github.com/NWylynko/cutroom/internal/speech"
	"github.com/NWylynko/cutroom/internal/timeline"
)

// Short-form defaults (L41): tight pads at cuts, pauses down to about 0.2 s, a vertical frame.
const (
	shortPadBefore = 0.06
	shortPadAfter  = 0.16
	shortMinCut    = 0.1
	shortMaxLen    = 60.0
)

func init() {
	Register(&Command{
		Name:     "short",
		MaxArgs:  0,
		Category: "edit",
		Summary:  "Make a vertical short from a span of the talk: its own project in shorts/<name>, short-form defaults",
		Help: `Takes the words from --from through --to and makes a short of them in shorts/<name>/ — a project of
its own with the same sources and transcripts, so every command works on it with --project:
  cutroom short --from "The first thing you need" --to "server is ready." --name laptop
  cutroom timeline --project shorts/laptop
  cutroom edit cut --text "By the way…" --project shorts/laptop
  cutroom render --project shorts/laptop        → shorts/laptop/out/final.mp4
  cutroom render out/short.mp4 --project shorts/laptop   → ./out/short.mp4 (paths you give are
                                                relative to your current folder, like any path)
  cutroom edit frame --size 720x1280 --project shorts/laptop   (resize; a 720p source enlarges less)
Short-form defaults (L41): 1080×1920 (--size), the picture scaled to fill the frame and cropped around
--anchor x,y (0–1; default 0.5,0.5: move it onto the speaker with ` + "`cutroom edit frame --anchor 0.4,0.5`" + `),
pads 0.06 s before / 0.16 s after each cut, every pause tightened to ~0.2 s. Open on the claim and
drop every aside: cut them in the short like in any edit. Overlays and burned captions keep to the
vertical safe zones (nothing above y 260, below y 1490 or in the right 120 px of 1080×1920).`,
		Flags: []Flag{
			{Name: "from", Kind: "string", Help: "First words of the short (quoted from the transcript)"},
			{Name: "to", Kind: "string", Help: "Last words of the short"},
			{Name: "name", Kind: "string", Default: "short", Help: "Its folder: shorts/<name>"},
			{Name: "size", Kind: "string", Help: "Output size WxH (vertical; default 1080x1920, or 720x1280 when the source is too small to fill 1080x1920 sharply)"},
			{Name: "anchor", Kind: "string", Default: "0.5,0.5", Help: "Point of the source kept in frame (0–1 across, down)"},
			{Name: "near", Kind: "string", Help: "Pick the --from match closest to this source time when words repeat"},
			{Name: "occurrence", Kind: "int", Help: "Pick the Nth --from match when words repeat"},
			{Name: "source", Kind: "string", Help: "Only match words in this source"},
			{Name: "look", Kind: "string", Help: "A saved look for its captions and cards (`cutroom style looks`)"},
		},
		Run: runShort,
	})
}

func runShort(c *Ctx) (*Result, error) {
	p, err := needProject(c.Dir)
	if err != nil {
		return nil, err
	}
	if !c.Has("from") || !c.Has("to") {
		return nil, Errf(ExitUsage, "missing-flag", `cutroom short --from "first words" --to "last words." --name <name>`, "short needs --from and --to")
	}
	name := c.Str("name")
	if !project.ValidID(name) {
		return nil, Errf(ExitUsage, "bad-name", "a short name: letters, digits, - and _ (e.g. --name laptop)", "bad --name %q", name)
	}
	w, h := 1080, 1920
	if c.Has("size") {
		if _, err := fmt.Sscanf(c.Str("size"), "%dx%d", &w, &h); err != nil || w < 64 || h < 64 || w%2+h%2 != 0 {
			return nil, Errf(ExitUsage, "bad-flag-value", "--size WxH with even numbers, e.g. 1080x1920", "bad --size %q", c.Str("size"))
		}
	}
	ax, ay, err := parseAnchor(c.Str("anchor"))
	if err != nil {
		return nil, err
	}
	var look *Look
	if c.Has("look") {
		lk, _, err := loadLook(c.Str("look"))
		if err != nil {
			return nil, err
		}
		look = &lk
	}
	dir := filepath.Join(p.Dir, "shorts", name)
	if _, err := os.Stat(filepath.Join(dir, project.File)); err == nil {
		return nil, Errf(ExitUsage, "short-exists", fmt.Sprintf("work on it with `--project %s`, or give another --name", relToCwd(dir)), "shorts/%s already exists", name)
	}
	// words come from the parent's transcripts (any source, kept by its edit or not)
	e := &editEnv{p: p, tr: timeline.Transcripts{}, dur: map[string]float64{}}
	for _, id := range p.IDs() {
		e.dur[id] = p.Sources[id].Facts.End()
		if t, err := speech.Load(cacheDir(p), id); err == nil {
			e.tr[id] = t
		}
	}
	src, i, _, err := e.resolve(c, c.Str("from"), false, "", 0, 0)
	if err != nil {
		return nil, err
	}
	_, _, j, err := e.resolve(c, c.Str("to"), false, src, i, i)
	if err != nil {
		return nil, err
	}
	s := p.Sources[src]
	if s.Facts.Video == nil {
		return nil, Errf(ExitUsage, "no-video", "quote words from a video source (the transcript of a recorder is not a picture)", "%s has no picture", src)
	}
	ws := e.tr[src].Words
	sizeNote := ""
	if v := s.Facts.Video; v != nil && !c.Has("size") {
		if ch := cropHeight(v.Width, v.Height, w, h); ch < 1080*0.75 { // a 720p-class source
			w, h = 720, 1280
			sizeNote = fmt.Sprintf("size 720x1280: the %dx%d source's 9:16 crop is %.0f px tall, so 1080x1920 would enlarge it %.1f× and look soft (--size 1080x1920 to have it anyway)", v.Width, v.Height, ch, 1920/ch)
		}
	}

	// the short's own project: same sources (paths rewritten for its folder), same transcripts
	sp := project.New(dir)
	for id, x := range p.Sources {
		cp := *x
		cp.Path = sp.RelPath(p.Abs(x))
		sp.Sources[id] = &cp
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	if err := sp.Save(); err != nil {
		return nil, err
	}
	for id := range e.tr {
		b, err := os.ReadFile(speech.TranscriptPath(cacheDir(p), id))
		if err != nil {
			continue
		}
		dst := speech.TranscriptPath(cacheDir(sp), id)
		os.MkdirAll(filepath.Dir(dst), 0o755)
		os.WriteFile(dst, b, 0o644)
	}

	se := &editEnv{p: sp, tr: e.tr, dur: e.dur}
	se.tl = newTimeline(sp.Sources[src])
	tl := se.tl
	tl.Width, tl.Height, tl.Fill, tl.FillX, tl.FillY = w, h, true, ax, ay
	tl.PadBefore, tl.PadAfter = shortPadBefore, shortPadAfter
	if look != nil {
		applyLook(tl, *look)
	}
	if a := p.Sources[src]; a.Facts.Audio == nil {
		tl.Audio = "" // a silent video: the parent's audio choice is made in the short with edit audio
	}
	a := math.Max(0, ws[i].Start-shortPadBefore)
	b := ws[j].End + shortPadAfter
	tl.Keep(src, a, b, "short")
	tr := tl.Tighten(e.tr, shortMinCut)
	if err := timeline.Checkpoint(dir, cacheDir(sp), fmt.Sprintf("short --from %q --to %q", c.Str("from"), c.Str("to"))); err != nil {
		return nil, err
	}
	if err := tl.Save(dir); err != nil {
		return nil, err
	}
	rel := relToCwd(dir)
	lines := []string{
		fmt.Sprintf("made %s: %s %s–%s %q … %q", rel, src, clock(ws[i].Start), clock(ws[j].End), speech.Text(ws, i, min(i+5, j)), speech.Text(ws, max(i, j-4), j)),
		fmt.Sprintf("  %dx%d, picture filled and cropped around %.2f,%.2f; pads %.2f/%.2f s; %d pause(s) tightened (%.1f s out)", w, h, ax, ay, shortPadBefore, shortPadAfter, tr.Pauses, tr.Removed),
		fmt.Sprintf("  timeline: %s, %s", plural(len(tl.Segments), "segment"), secs(tl.Duration())),
	}
	if d := tl.Duration(); d > shortMaxLen {
		lines = append(lines, fmt.Sprintf("WARN short is %.1f s: shorts work best under %.0f s — cut asides and side points, or end earlier (--to)", d, shortMaxLen))
	}
	if v := s.Facts.Video; v != nil {
		// how much the 9:16 crop of the source is enlarged (L40: above ~1.5× it looks soft)
		cropH := cropHeight(v.Width, v.Height, w, h)
		if sizeNote != "" {
			lines = append(lines, sizeNote)
		}
		if look != nil {
			lines = append(lines, fmt.Sprintf("look %q: %s (burn captions with `cutroom captions --burn --project %s`)", c.Str("look"), lookSummary(*look), relToCwd(dir)))
		}
		if up := float64(h) / cropH; up > 1.5 {
			tip := "a 4K recording keeps a short sharp"
			if h > 1280 {
				tip += "; --size 720x1280 enlarges it less"
			}
			lines = append(lines, fmt.Sprintf("WARN the %dx%d source is enlarged %.1f× to fill %dx%d: it will look soft (%s)", v.Width, v.Height, up, w, h, tip))
		}
	}
	for _, x := range se.validate() {
		if x.Level == "warn" || x.Level == "error" {
			lines = append(lines, fmt.Sprintf("%s[%s] at output %s: %s", strings.ToUpper(x.Level), x.Code, clock(x.At), x.Message))
		}
	}
	return &Result{Text: strings.Join(lines, "\n"), Data: map[string]any{"project": rel, "duration": tl.Duration(), "segments": len(tl.Segments)},
		Next: []string{"cutroom timeline --project " + rel + "   (read every sentence: cut asides and side points)",
			"cutroom look --output 0:01 --project " + rel + "   (check the framing)",
			"cutroom render out/short.mp4 --project " + rel + "   (paths are relative to your current folder)"}}, nil
}

func parseAnchor(s string) (float64, float64, error) {
	var x, y float64
	if _, err := fmt.Sscanf(s, "%g,%g", &x, &y); err != nil || x < 0 || x > 1 || y < 0 || y > 1 {
		return 0, 0, Errf(ExitUsage, "bad-anchor", "--anchor x,y with 0–1 across and down, e.g. 0.4,0.5", "bad --anchor %q", s)
	}
	return x, y, nil
}

// cropHeight is the height of the largest w:h crop of a vw×vh picture.
func cropHeight(vw, vh, w, h int) float64 {
	if float64(vw)/float64(vh) < float64(w)/float64(h) {
		return float64(vw) * float64(h) / float64(w)
	}
	return float64(vh)
}
