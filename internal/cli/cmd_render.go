package cli

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/cutroom/cutroom/internal/render"
	"github.com/cutroom/cutroom/internal/timecode"
)

func init() {
	Register(&Command{
		Name:     "render",
		Args:     "[output.mp4]",
		MaxArgs:  1,
		Category: "render",
		Summary:  "Render the timeline to a video (cached chunks; --preview is fast and small)",
		Help: `Output: the positional path, default out/final.mp4 (out/preview.mp4 with --preview).
Renders in chunks of ~45 s split at cuts; unchanged chunks are reused, so re-rendering after a small
edit only redoes the chunks it touched. Audio is cut sample-exactly, run through a voice chain and
normalised to --loudness (measured, then limited to --true-peak).
Cuts land on whole frames and audio samples; sources are decoded once and the video is encoded once.
Examples:
  cutroom render --dry-run              (what would run and how long it should take)
  cutroom render --preview              (540p, fast: check the edit)
  cutroom render --preview --range 0:20-0:40   (only the ~45 s chunks covering that output range;
                                               a short timeline is one chunk, so it renders all of it)
  cutroom render                        (final quality)`,
		Flags: []Flag{
			{Name: "preview", Kind: "bool", Help: "540p, fast encode"},
			{Name: "range", Kind: "string", Help: "Only chunks covering this OUTPUT time range (start-end)"},
			{Name: "dry-run", Kind: "bool", Help: "Print the plan and time estimate; render nothing"},
			{Name: "jobs", Kind: "int", Default: "4", Help: "Chunks rendered in parallel"},
			{Name: "loudness", Kind: "float", Default: "-16", Help: "Integrated loudness target (LUFS)"},
			{Name: "true-peak", Kind: "float", Default: "-1.5", Help: "Peak ceiling (dBTP)"},
			{Name: "audio-chain", Kind: "string", Default: "voice", Help: "voice (default speech chain), none, or an ffmpeg audio filter chain"},
		},
		Run: runRender,
	})
}

func runRender(c *Ctx) (*Result, error) {
	e, err := loadEditEnv(c, false)
	if err != nil {
		return nil, err
	}
	issues := e.validate()
	for _, is := range issues {
		if is.Level == "error" {
			fix := "fix it first: `cutroom timeline` lists the problems"
			if is.Fix != "" {
				fix = is.Fix
			}
			return nil, Errf(ExitInvalid, "invalid-timeline", fix, "timeline error: %s", is.Message)
		}
	}
	srcs := map[string]render.Source{}
	for id, s := range e.p.Sources {
		srcs[id] = render.Source{Path: e.p.Abs(s), HasAudio: s.Facts.Audio != nil, Size: s.Size, ModTime: s.ModTime, Sync: s.Sync}
		if s.Facts.Audio != nil {
			v := srcs[id]
			v.AudioStart = s.Facts.Audio.Start
			srcs[id] = v
		}
		if stale, _ := e.p.Stale(s); stale {
			return nil, Errf(ExitInvalid, "stale-source", "run `cutroom init` to re-inspect it, then check the timeline still fits", "source %q changed since it was inspected", id)
		}
	}
	chain := c.Str("audio-chain")
	switch chain {
	case "voice":
		chain = render.VoiceChain
	case "none":
		chain = ""
	}
	o := render.Options{Preview: c.Bool("preview"), Jobs: c.Int("jobs"), Target: 45, CacheDir: cacheDir(e.p),
		Loudness: c.Float("loudness"), TruePeak: c.Float("true-peak"), Chain: chain, DryRun: c.Bool("dry-run"),
		Progress: func(s string) { fmt.Fprintln(c.ErrW, s) }}
	if c.Has("range") {
		a, b, err := parseRange(c.Str("range"), timecode.Rate{Num: 30, Den: 1})
		if err != nil {
			return nil, Errf(ExitUsage, "bad-range", "an OUTPUT time range like 0:20-0:40 (see `cutroom timeline` for output times)", "%v", err)
		}
		if a >= e.tl.Duration() {
			return nil, Errf(ExitUsage, "out-of-range", fmt.Sprintf("the timeline is %.3f s long", e.tl.Duration()), "range starts after the end")
		}
		o.Range = [2]float64{a, b}
	}
	out := ""
	if len(c.Args) == 1 {
		out = c.Args[0]
		if !filepath.IsAbs(out) {
			out = filepath.Join(e.p.Dir, out)
		}
	} else {
		name := "final.mp4"
		if o.Preview {
			name = "preview.mp4"
		}
		if c.Has("range") {
			name = strings.TrimSuffix(name, ".mp4") + "-range.mp4"
		}
		out = filepath.Join(e.p.Dir, "out", name)
	}
	if !strings.HasSuffix(strings.ToLower(out), ".mp4") {
		return nil, Errf(ExitUsage, "bad-output", "name the output something.mp4", "output must be .mp4")
	}
	res, err := render.Render(e.tl, srcs, out, o)
	if err != nil {
		return nil, Errf(ExitExternal, "render-failed", "re-run the same command (finished chunks are kept); if it repeats, `cutroom timeline` and `cutroom doctor`", "%v", err)
	}
	var lines []string
	cached := 0
	for _, ch := range res.Chunks {
		if ch.Cached {
			cached++
		}
	}
	mode := "final"
	if o.Preview {
		mode = "preview"
	}
	rangeNote := ""
	if c.Has("range") && len(res.Chunks) > 0 {
		first, last := res.Chunks[0], res.Chunks[len(res.Chunks)-1]
		a, b := first.Start, last.Start+last.Duration
		rangeNote = fmt.Sprintf("  range %s–%s: renders the whole chunk(s) covering it, output %s–%s (%s)",
			clock(o.Range[0]), clock(math.Min(o.Range[1], e.tl.Duration())), clock(a), clock(b), secs(b-a))
		if a <= 0 && b >= e.tl.Duration()-0.001 {
			rangeNote += " = the whole timeline (it fits in one chunk)"
		}
	}
	if res.DryRun {
		lines = append(lines, fmt.Sprintf("dry run (%s, %dx%d): %d chunk(s), %d cached, %s of video", mode, res.Width, res.Height, len(res.Chunks), cached, secs(res.Expected)))
		for _, ch := range res.Chunks {
			st := "render"
			if ch.Cached {
				st = "cached"
			}
			lines = append(lines, fmt.Sprintf("  chunk %d  %s–%s  %d segment(s)  %s", ch.Index+1, clock(ch.Start), clock(ch.Start+ch.Duration), len(ch.Segments), st))
		}
		if rangeNote != "" {
			lines = append(lines, rangeNote)
		}
		lines = append(lines, fmt.Sprintf("estimate: ~%.0f s", math.Ceil(res.Estimate)))
		return &Result{Text: strings.Join(lines, "\n"), Data: res, Next: []string{strings.TrimSpace("cutroom render " + flagIf(o.Preview, "--preview"))}}, nil
	}
	lines = append(lines, fmt.Sprintf("wrote %s (%s, %dx%d)", relToCwd(res.Output), mode, res.Width, res.Height))
	lines = append(lines, fmt.Sprintf("  %s, %d frames (timeline says %s)  rendered in %.1f s, %d of %d chunk(s) cached",
		secs(res.Duration), res.Frames, secs(res.Expected), res.Seconds, cached, len(res.Chunks)))
	if math.Abs(res.Duration-res.Expected) > 0.5/e.tl.Rate() {
		lines = append(lines, fmt.Sprintf("WARN duration differs from the timeline by %.3f s", res.Duration-res.Expected))
	}
	a := res.Audio
	lines = append(lines, fmt.Sprintf("  audio: %.1f LUFS (target %.0f), peak %.1f dBTP, gain %+.1f dB", a.LUFS, a.Target, a.TruePeak, a.GainDB))
	if c.Has("range") {
		lines = append(lines, rangeNote, fmt.Sprintf("  file time 0 = output %s", clock(res.RangeStart)))
	}
	next := []string{"cutroom check " + relToCwd(res.Output)}
	return &Result{Text: strings.Join(lines, "\n"), Data: res, Next: next}, nil
}

func flagIf(b bool, f string) string {
	if b {
		return f
	}
	return ""
}
