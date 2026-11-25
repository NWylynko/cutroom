package cli

import (
	"fmt"
	"github.com/cutroom/cutroom/internal/media"
	"math"
	"path/filepath"
	"strings"

	"github.com/cutroom/cutroom/internal/render"
	"github.com/cutroom/cutroom/internal/timecode"
	"github.com/cutroom/cutroom/internal/timeline"
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
  cutroom render --preview --range 0:20-0:40   (just that OUTPUT range, to whole frames, into
                                               out/preview-range.mp4; loudness as in the full render)
  cutroom render                        (final quality)`,
		Flags: []Flag{
			{Name: "preview", Kind: "bool", Help: "540p, fast encode"},
			{Name: "output", Kind: "string", Help: "Output path (same as the positional [output.mp4])"},
			{Name: "range", Kind: "string", Help: "Render only this OUTPUT time range (start-end), to whole frames"},
			{Name: "dry-run", Kind: "bool", Help: "Print the plan and time estimate; render nothing"},
			{Name: "jobs", Kind: "int", Default: "4", Help: "Chunks rendered in parallel"},
			{Name: "loudness", Kind: "float", Default: "-16", Help: "Integrated loudness target (LUFS)"},
			{Name: "true-peak", Kind: "float", Default: "-1.5", Help: "Peak ceiling (dBTP)"},
			{Name: "size", Kind: "string", Help: "Set the output size first (saved in the timeline, undoable): 1080p, 720p, 2160p or WxH"},
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
	var sizeNote string
	if c.Has("size") {
		w, h, err := parseSize(c.Str("size"), e.tl.Width, e.tl.Height)
		if err != nil {
			return nil, err
		}
		if w != e.tl.Width || h != e.tl.Height {
			if !c.Bool("dry-run") {
				if err := timeline.Checkpoint(e.p.Dir, cacheDir(e.p), fmt.Sprintf("edit frame --size %dx%d (render --size)", w, h)); err != nil {
					return nil, err
				}
			}
			sizeNote = fmt.Sprintf("output size %dx%d → %dx%d (saved in the timeline, like `edit frame --size`; `cutroom edit undo` reverts it)", e.tl.Width, e.tl.Height, w, h)
			e.tl.Width, e.tl.Height = w, h
			if !c.Bool("dry-run") {
				if err := e.tl.Save(e.p.Dir); err != nil {
					return nil, err
				}
			}
		}
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
		Progress: func(s string) { fmt.Fprintln(c.ErrW, s) }, ProjDir: e.p.Dir, Transcripts: e.tr}
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
	if c.Has("output") {
		if len(c.Args) == 1 && c.Args[0] != c.Str("output") {
			return nil, Errf(ExitUsage, "bad-args", "give the path once: `cutroom render out/final.mp4`", "two output paths: %q and --output %q", c.Args[0], c.Str("output"))
		}
		out = c.Str("output")
	} else if len(c.Args) == 1 {
		out = c.Args[0]
	}
	if out != "" {
		out = userPath(out) // like any CLI: relative to the shell's folder, --project or not
	}
	if out == "" {
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
	if err := writeTarget(e.p, out); err != nil {
		return nil, err
	}
	if !o.Preview && !o.DryRun && e.tl.Width*e.tl.Height > 1920*1080 {
		fmt.Fprintf(c.ErrW, "rendering at %dx%d (the timeline's size, from the source); for 1080p stop and run `cutroom edit frame --size 1080p` (or `render --size 1080p`)\n", e.tl.Width, e.tl.Height)
	}
	res, err := render.Render(e.tl, srcs, out, o)
	if err != nil {
		fix := "re-run the same command (finished chunks are kept); if it repeats, `cutroom timeline` and `cutroom doctor`"
		if strings.Contains(err.Error(), "(L14)") {
			fix = "a source has fewer frames there than the timeline expects (re-running won't help): `cutroom inspect <source>` shows where its picture and sound end and any gaps; cut the edit back from that point (`cutroom edit cut <start>-<end> --source <id>`), then render again"
		}
		return nil, Errf(ExitExternal, "render-failed", fix, "%v", err)
	}
	var lines []string
	if sizeNote != "" {
		lines = append(lines, sizeNote)
	}
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
	if c.Has("range") {
		rangeNote = fmt.Sprintf("  range: output %s–%s of the %s edit (file time 0 = output %s)",
			clock(res.RangeStart), clock(res.RangeStart+res.Expected), secs(e.tl.Duration()), clock(res.RangeStart))
	}
	if res.DryRun {
		lines = append(lines, fmt.Sprintf("dry run → %s (%s, %dx%d): %d chunk(s), %d cached, %s of video", relToCwd(out), mode, res.Width, res.Height, len(res.Chunks), cached, secs(res.Expected)))
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
		return &Result{Text: strings.Join(lines, "\n"), Data: res, Next: []string{strings.Join(strings.Fields(strings.Join([]string{"cutroom render", flagIf(o.Preview, "--preview"), flagIf(c.Has("range"), "--range "+c.Str("range")), customOut(c)}, " ")), " ")}}, nil
	}
	lines = append(lines, fmt.Sprintf("wrote %s (%s, %dx%d)", relToCwd(res.Output), mode, res.Width, res.Height))
	lines = append(lines, fmt.Sprintf("  %s, %d frames (timeline says %s)  rendered in %.1f s, %d of %d chunk(s) cached",
		secs(res.Duration), res.Frames, secs(res.Expected), res.Seconds, cached, len(res.Chunks)))
	if math.Abs(res.Duration-res.Expected) > 0.5/e.tl.Rate() {
		lines = append(lines, fmt.Sprintf("WARN duration differs from the timeline by %.3f s", res.Duration-res.Expected))
	}
	for _, id := range e.p.IDs() {
		f := e.p.Sources[id].Facts
		used := false
		for _, g := range e.tl.Segments {
			used = used || g.Source == id
		}
		if f != nil && f.Video != nil && f.Video.HDR != "" && used {
			if sdr := media.ReadSDR(e.p.Abs(e.p.Sources[id]), 0, 0); sdr.Method != "none" && sdr.Method != "" {
				lines = append(lines, fmt.Sprintf("  picture: %s is HDR (%s): tone-mapped to SDR bt709 (%s); `cutroom check` compares its colour with the source", id, strings.ToUpper(f.Video.HDR), sdr.Method))
			} else {
				lines = append(lines, fmt.Sprintf("  WARN picture: %s is HDR (%s) but this ffmpeg can't tone-map: the render looks washed out (`cutroom doctor`)", id, strings.ToUpper(f.Video.HDR)))
			}
		}
		if f != nil && f.Video != nil && f.Video.VFR && used {
			lines = append(lines, fmt.Sprintf("  picture: %s has a variable frame rate (%d gaps): each frame was placed by its timestamp on a constant %s fps grid, so sound and picture can't drift apart (L02)", id, f.Video.Gaps, strings.TrimSuffix(e.tl.FPS, "/1")))
		}
		if f != nil && f.Video != nil && f.Audio != nil && used && math.Abs(f.AVOffset) >= 0.005 {
			lines = append(lines, fmt.Sprintf("  audio: %s's sound starts %+.3f s from its picture; kept as recorded, so lip sync is as the camera captured it", id, f.AVOffset))
		}
		if f != nil && f.Video != nil && f.Video.Rotation != 0 && used {
			v := f.Video
			lines = append(lines, fmt.Sprintf("  picture: %s's %d° display rotation is baked into the pixels; the file has no rotation metadata (upright in every player)", id, v.Rotation))
		}
	}
	a := res.Audio
	lines = append(lines, fmt.Sprintf("  audio: %.1f LUFS (target %.0f), peak %.1f dBTP, gain %+.1f dB", a.LUFS, a.Target, a.TruePeak, a.GainDB))
	if c.Has("range") {
		lines = append(lines, rangeNote)
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

// customOut repeats a user-given output path in a suggested command.
func customOut(c *Ctx) string {
	if c.Has("output") {
		return c.Str("output")
	}
	if len(c.Args) == 1 {
		return c.Args[0]
	}
	return ""
}
