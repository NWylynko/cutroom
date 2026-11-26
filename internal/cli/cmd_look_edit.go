package cli

import (
	"fmt"
	"github.com/NWylynko/cutroom/internal/timeline"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/NWylynko/cutroom/internal/deps"
	"github.com/NWylynko/cutroom/internal/media"
	"github.com/NWylynko/cutroom/internal/perceive"
	"github.com/NWylynko/cutroom/internal/raster"
	"github.com/NWylynko/cutroom/internal/render"
	"github.com/NWylynko/cutroom/internal/speech"
	"github.com/NWylynko/cutroom/internal/timecode"
)

// outputWords maps every kept transcript word to output time.
func (e *editEnv) outputWords() []speech.Word {
	var out []speech.Word
	for _, g := range e.tl.Segments {
		tx := e.tr[g.Source]
		if tx == nil {
			continue
		}
		for _, w := range tx.Words {
			if m := (w.Start + w.End) / 2; m >= g.In && m < g.Out {
				w.Start = g.OutStart + w.Start - g.In
				w.End = g.OutStart + w.End - g.In
				out = append(out, w)
			}
		}
	}
	return out
}

// manifestWords maps transcript words into a rendered file's time.
func manifestWords(m *render.Manifest, tr map[string]*speech.Transcript) ([]speech.Word, []float64) {
	var out []speech.Word
	var cuts []float64
	for i, g := range m.Segments {
		if i > 0 {
			cuts = append(cuts, g.OutStart)
		}
		tx := tr[g.Source]
		if tx == nil {
			continue
		}
		for _, w := range tx.Words {
			if mid := (w.Start + w.End) / 2; mid >= g.In && mid < g.Out {
				w.Start = g.OutStart + w.Start - g.In
				w.End = g.OutStart + w.End - g.In
				out = append(out, w)
			}
		}
	}
	return out, cuts
}

// currentRender returns the newest render whose manifest matches the timeline now.
func (e *editEnv) currentRender() (string, *render.Manifest, error) {
	h := render.TimelineHash(e.tl)
	stale := ""
	for _, n := range []string{"final.mp4", "preview.mp4"} {
		f := filepath.Join(e.p.Dir, "out", n)
		m, err := render.LoadManifest(f)
		if err != nil {
			continue
		}
		if m.TimelineHash == h && !m.Partial {
			return f, m, nil
		}
		stale = n
	}
	if stale != "" {
		return "", nil, Errf(ExitInvalid, "stale-render", "render again (`cutroom render --preview` is fast), or pass --file to listen to the old one", "out/%s was rendered from an older timeline", stale)
	}
	return "", nil, Errf(ExitInvalid, "no-render", "render first: `cutroom render --preview`", "no render of the current timeline")
}

func lookEdit(c *Ctx) (*Result, error) {
	e, err := loadEditEnv(c, false)
	if err != nil {
		return nil, err
	}
	tl := e.tl
	fps := tl.Rate()
	src := func(id string) string { return e.p.Abs(e.p.Sources[id]) }
	srcWords := func(id string, t float64) string {
		if tx := e.tr[id]; tx != nil {
			for _, w := range tx.Words {
				if t >= w.Start && t < w.End {
					return w.Text
				}
			}
		}
		return ""
	}
	var cells []perceive.Cell
	var desc, name string
	var ts []float64 // output times to show (--output, --overlays)
	cols := 0
	switch {
	case c.Bool("overlays"):
		ps, err := render.Placements(tl, e.tr, e.p.Dir, cacheDir(e.p), tl.Width, tl.Height)
		if err != nil {
			return nil, Errf(ExitInvalid, "overlay-failed", "`cutroom overlay list`; fix or remove the overlay", "%v", err)
		}
		for _, p := range ps {
			a, b := p.Start, p.End
			if p.Kind == "captions" && len(p.Groups) > 0 { // the first caption group on screen
				a, b = p.Groups[0].Start, p.Groups[0].End
			}
			ts = append(ts, math.Min((a+b)/2, tl.Duration()-0.5/fps))
		}
		if len(ts) == 0 {
			return &Result{Text: "no overlays: `cutroom overlay add` puts one on", Data: map[string]any{"overlays": 0}, Next: []string{"cutroom guide overlays"}}, nil
		}
		sort.Float64s(ts)
		desc = "overlay(s), each at the middle of its window, drawn as rendered; labelled output time = source time"
		name = "overlays.png"
	case c.Bool("cuts"):
		js := tl.Joins(e.tr)
		if len(js) == 0 {
			return &Result{Text: "no cuts: the timeline is one segment", Data: map[string]any{"cuts": 0}}, nil
		}
		per := 12
		page := max(1, c.Int("page"))
		pages := (len(js) + per - 1) / per
		if page > pages {
			return nil, Errf(ExitUsage, "bad-page", fmt.Sprintf("there are %d page(s)", pages), "no page %d", page)
		}
		from, to := (page-1)*per, min(len(js), page*per)
		for k, j := range js[from:to] {
			l, r := tl.Find(j.Left), tl.Find(j.Right)
			lt := l.Out - 0.5/fps
			rt := r.In + 0.5/fps
			cells = append(cells,
				perceive.Cell{Path: src(l.Source), T: lt, Label: fmt.Sprintf("%s %d @%s  before (%s %.2f)", j.Kind, from+k+1, clock(j.At), l.Source, lt), Say: "…" + j.Before, VF: segVF(tl, *l)},
				perceive.Cell{Path: src(r.Source), T: rt, Label: fmt.Sprintf("after (%s %.2f)  pause %.2f s", r.Source, rt, j.Pause), Say: j.After + "…", Cut: true, VF: segVF(tl, *r)})
		}
		cols = 4
		desc = fmt.Sprintf("cuts %d–%d of %d: each pair is the last frame before | first frame after a cut (red bar), with the words either side", from+1, to, len(js))
		if pages > 1 {
			desc += fmt.Sprintf("; page %d of %d (--page)", page, pages)
		}
		name = fmt.Sprintf("cuts_p%d.png", page)
	case c.Has("file"):
		f := c.Str("file")
		if _, err := os.Stat(f); err != nil {
			return nil, fileNotFound(f)
		}
		var words []speech.Word
		if m, err := render.LoadManifest(f); err == nil {
			words, _ = manifestWords(m, e.tr)
		}
		var a, b float64
		isRange := true
		fx, err := media.Probe(f, media.Options{})
		if err != nil || fx.Video == nil {
			return nil, Errf(ExitUsage, "not-a-video", "give a video file", "%s has no picture", f)
		}
		last := math.Max(0, fx.Duration-1/fps) // the last frame's time (a range end at the file's end)
		if len(c.Args) == 0 {                  // the whole file as a contact sheet
			a, b = 0, last
		} else {
			a, b, isRange, err = timeOrRange(c.Args[0])
			if err != nil {
				return nil, err
			}
			if a > last {
				return nil, Errf(ExitUsage, "out-of-range", fmt.Sprintf("%s is %.3f s long", relToCwd(f), fx.Duration), "%s is past the end of the file", clock(a))
			}
			b = math.Min(b, last)
		}
		if !isRange {
			b = a
		}
		n := 1
		if isRange {
			n = c.Int("count")
		}
		r := timecode.Rate{Num: int64(math.Round(fps)), Den: 1}
		img, meta, err := perceive.Sheet(f, a, b, n, r, words)
		if err != nil {
			return nil, Errf(ExitExternal, "frame-failed", "check the time is inside the file", "%v", err)
		}
		out := filepath.Join(lookDir(e.p), fmt.Sprintf("file_%s_%s-%s.png", strings.TrimSuffix(filepath.Base(f), filepath.Ext(f)), stamp(a), stamp(b)))
		if err := raster.SavePNG(out, img); err != nil {
			return nil, err
		}
		when := clock(a)
		if isRange {
			when = clock(a) + "–" + clock(b)
		}
		return &Result{Text: fmt.Sprintf("wrote %s\n  %d frame(s) of %s at %s (file time), labelled with the word being spoken\n  view the image to see it", relToCwd(out), n, relToCwd(f), when),
			Data: map[string]any{"image": out, "frames": meta}}, nil
	default: // --output
		if len(c.Args) == 0 {
			return nil, Errf(ExitUsage, "bad-args", "give an OUTPUT time or range, e.g. `cutroom look --output 0:03-0:06`", "look --output needs a time")
		}
		a, b, isRange, err := timeOrRange(c.Args[0])
		if err != nil {
			return nil, err
		}
		d := tl.Duration()
		if a >= d {
			return nil, Errf(ExitUsage, "out-of-range", fmt.Sprintf("the edit is %.3f s long", d), "output %s is past the end", clock(a))
		}
		b = math.Min(b, d-0.5/fps)
		n := 1
		if isRange {
			n = c.Int("count")
			if n < 2 || n > 36 {
				return nil, Errf(ExitUsage, "bad-flag-value", "use --count 2–36", "--count %d out of range", n)
			}
		}
		for i := 0; i < n; i++ {
			t := a
			if n > 1 {
				t = a + (b-a)*float64(i)/float64(n-1)
			}
			ts = append(ts, t)
		}
		desc = fmt.Sprintf("frame(s) of the edit at output %s–%s; each labelled output time = source time; red bar = a cut happened before it", clock(a), clock(b))
		name = fmt.Sprintf("output_%s-%s_x%d.png", stamp(a), stamp(b), n)
	}
	if ts != nil {
		prev := ""
		var prevSeg *timeline.Segment
		// overlays drawn as the render will (cards at the output size)
		ps, err := render.Placements(tl, e.tr, e.p.Dir, cacheDir(e.p), tl.Width, tl.Height)
		if err != nil {
			return nil, Errf(ExitInvalid, "overlay-failed", "`cutroom overlay list`; fix or remove the overlay", "%v", err)
		}
		shown := map[string]bool{}
		for _, t := range ts {
			g, st, ok := tl.ToSource(t)
			if !ok {
				continue
			}
			label := fmt.Sprintf("out %s = %s %.2f", clock(t), g.Source, st)
			if g.ZoomOf() > 1 {
				label += fmt.Sprintf(" zoom %.2f×", g.ZoomOf())
			}
			var over []string
			for _, p := range ps {
				if p.Kind == "captions" {
					for _, g := range p.Groups {
						if t >= g.Start && t < g.End {
							over = append(over, g.PNG)
							label += " +cc"
							shown[p.ID] = true
						}
					}
					continue
				}
				if p.Kind == "clip" && t >= p.Start && t < p.End {
					cc, err := render.ClipCard(p, t, cacheDir(e.p), tl.Width, tl.Height)
					if err != nil {
						return nil, Errf(ExitExternal, "frame-failed", "check the clip is readable: `cutroom overlay list`", "overlay %s: %v", p.ID, err)
					}
					over = append(over, cc)
					label += " +" + p.ID
					shown[p.ID] = true
					continue
				}
				if t >= p.Start && t < p.End {
					over = append(over, p.PNG)
					label += " +" + p.ID
					shown[p.ID] = true
				}
			}
			cells = append(cells, perceive.Cell{Path: src(g.Source), T: st, Label: label, Over: over,
				Say: srcWords(g.Source, st), VF: segVF(tl, *g),
				// a cut, not just a zoom change on continuous picture
				Cut: !c.Bool("overlays") && prev != "" && prev != g.ID && !(prevSeg != nil && prevSeg.Source == g.Source && math.Abs(prevSeg.Out-g.In) < 1e-6)})
			prev, prevSeg = g.ID, g
		}
		desc = fmt.Sprintf("%d %s", len(cells), desc)
		if len(shown) > 0 {
			var ids []string
			for _, p := range ps {
				if shown[p.ID] {
					if p.Kind == "captions" {
						ids = append(ids, p.ID+" (burned-in captions, +cc)")
					} else {
						ids = append(ids, p.ID)
					}
				}
			}
			var uniq []string
			for _, id := range ids {
				if len(uniq) == 0 || uniq[len(uniq)-1] != id {
					uniq = append(uniq, id)
				}
			}
			desc += "; with overlay(s) " + strings.Join(uniq, ", ") + " drawn as rendered (+oN in the label)"
		}
	}
	img, err := perceive.Cells(cells, cols)
	if err != nil {
		return nil, Errf(ExitExternal, "frame-failed", "check the sources are readable", "%v", err)
	}
	out := filepath.Join(lookDir(e.p), name)
	if err := raster.SavePNG(out, img); err != nil {
		return nil, err
	}
	return &Result{Text: fmt.Sprintf("wrote %s\n  %s\n  view the image to see it", relToCwd(out), desc), Data: map[string]any{"image": out, "cells": cells}}, nil
}

func timeOrRange(s string) (float64, float64, bool, error) {
	r := timecode.Rate{Num: 30, Den: 1}
	if strings.ContainsAny(s, "-–") || strings.Contains(s, "..") {
		a, b, err := parseRange(s, r)
		if err != nil {
			return 0, 0, false, Errf(ExitUsage, "bad-range", "write a range as start-end, e.g. 3-6 or 0:03-0:06", "%v", err)
		}
		return a, b, true, nil
	}
	t, err := timecode.Parse(s, r)
	if err != nil {
		return 0, 0, false, Errf(ExitUsage, "bad-time", "times look like 12.5, 1:02.5 or f372", "%v", err)
	}
	return t.Float(), t.Float(), false, nil
}

func listenFile(c *Ctx) (*Result, error) {
	e, err := loadEditEnv(c, false)
	if err != nil {
		return nil, err
	}
	var f string
	var m *render.Manifest
	if c.Has("file") {
		f = c.Str("file")
		if _, err := os.Stat(f); err != nil {
			return nil, fileNotFound(f)
		}
		m, _ = render.LoadManifest(f)
	} else if f, m, err = e.currentRender(); err != nil {
		return nil, err
	}
	if len(c.Args) == 0 {
		return nil, Errf(ExitUsage, "bad-args", "give a range in the file, e.g. `cutroom listen "+relToCwd(f)+" 3-5`", "listen needs a range")
	}
	a, b, err := parseRange(c.Args[0], timecode.Rate{Num: 30, Den: 1})
	if err != nil {
		return nil, Errf(ExitUsage, "bad-range", "write a range as start-end, e.g. 3-5", "%v", err)
	}
	pad := c.Float("pad")
	a = math.Max(0, a-pad)
	b += pad
	if b-a > 30 {
		return nil, Errf(ExitUsage, "range-too-long", "listen to at most 30 s at a time; split it", "range is %.1f s", b-a)
	}
	var words []speech.Word
	var cuts []float64
	room, peak := math.NaN(), 0.0
	if m != nil {
		words, cuts = manifestWords(m, e.tr)
		if tx := e.tr[m.Audio]; tx != nil {
			room, peak = tx.RoomTone, tx.SpeechDB
		}
	}
	model, _ := deps.FindModel()
	os.MkdirAll(cacheDir(e.p), 0o755)
	tmp, err := os.MkdirTemp(cacheDir(e.p), "listen")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	out := filepath.Join(lookDir(e.p), fmt.Sprintf("file_%s_listen_%s-%s.png", strings.TrimSuffix(filepath.Base(f), filepath.Ext(f)), stamp(a), stamp(b)))
	L, err := perceive.Listen(f, a, b, words, cuts, room, peak, model, out, tmp)
	if err != nil {
		return nil, Errf(ExitExternal, "listen-failed", "check the range is inside the file", "%v", err)
	}
	var ic []string
	for _, ct := range cuts {
		if ct > a && ct < b {
			ic = append(ic, clock(ct))
		}
	}
	lines := []string{
		fmt.Sprintf("wrote %s  (view it: spectrogram, loudness, words, cuts in red)", relToCwd(L.Image)),
		fmt.Sprintf("%s %s–%s (file time)", relToCwd(f), clock(a), clock(b)),
		"  cuts in range: " + orNone(strings.Join(ic, ", ")),
	}
	if m != nil {
		lines = append(lines, "  words planned here: "+orNone(L.Transcript))
	}
	if model != "" {
		lines = append(lines, "  heard (isolated clip): "+orNone(L.Heard))
	}
	for _, r := range L.Unworded {
		lines = append(lines, fmt.Sprintf("  ! sound with no planned words %s–%s (peak %.0f dB)", clock(r.Start), clock(r.End), r.PeakDB))
	}
	return &Result{Text: strings.Join(lines, "\n"), Data: L}, nil
}

// segVF is what a segment's picture goes through in the render (punch-in, fill crop), so look shows
// what the render will.
func segVF(tl *timeline.Timeline, g timeline.Segment) string {
	ax, ay := g.Anchor()
	return render.PictureVF(g.ZoomOf(), ax, ay, tl.Fill, tl.FillXOf(g), tl.FillY, tl.Bars, tl.Width, tl.Height)
}
