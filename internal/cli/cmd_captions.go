package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cutroom/cutroom/internal/captions"
	"github.com/cutroom/cutroom/internal/render"
	"github.com/cutroom/cutroom/internal/timeline"
)

func init() {
	Register(&Command{
		Name:     "captions",
		MaxArgs:  0,
		Category: "render",
		Summary:  "Write the edit's captions (SRT/VTT, output time) and check them: reading, sentence seams, words",
		Help: `Writes out/captions.srt from the words the timeline keeps (um/uh left out), in output time.
Cues: ≤ 2 lines of --max-chars, ≤ 7 s, a new sentence or a pause ≥ 0.7 s starts a new cue, no cue or
line ends on a weak word, no 1–2 word sentence tail opens a cue (L48, L49).
Checks printed every time: reading limits, timing, words vs the timeline, sentence seams at cuts.
Examples:
  cutroom captions                       out/captions.srt
  cutroom captions --vtt                 out/captions.vtt (WebVTT)
  cutroom captions --check out/captions.srt   check a file (e.g. after editing it by hand) against the timeline
  cutroom captions --dry-run             show the cues and checks, write nothing
  cutroom captions --output out/final.srt   name the file
  cutroom captions --burn                burn short captions into the video (1–3 words at a time,
                                         as on shorts) and write the subtitle file too, in one call;
                                         they follow the cut. Remove: cutroom overlay rm <id>`,
		Flags: []Flag{
			{Name: "vtt", Kind: "bool", Help: "WebVTT instead of SRT"},
			{Name: "output", Kind: "string", Help: "Path (default out/captions.srt or .vtt)"},
			{Name: "check", Kind: "string", Help: "Check this SRT/VTT file against the timeline instead of writing one"},
			{Name: "max-chars", Kind: "int", Default: "42", Help: "Characters per line"},
			{Name: "lines", Kind: "int", Default: "2", Help: "Lines per cue (1 for shorts)"},
			{Name: "dry-run", Kind: "bool", Help: "Print the cues and checks; write nothing"},
			{Name: "all", Kind: "bool", Help: "Print every cue"},
			{Name: "burn", Kind: "bool", Help: "Burn captions into the picture: short 1–3 word groups (L48), drawn by the next render"},
		},
		Run: runCaptions,
	})
}

// captionsMeta is written next to a caption file so status can tell whether it is current.
type captionsMeta struct {
	TimelineHash string    `json:"timeline_hash"`
	Cues         int       `json:"cues"`
	Words        int       `json:"words"`
	WrittenAt    time.Time `json:"written_at"`
}

func captionsMetaPath(f string) string { return f + ".cutroom.json" }

func runCaptions(c *Ctx) (*Result, error) {
	e, err := loadEditEnv(c, false)
	if err != nil {
		return nil, err
	}
	o := captions.Default
	o.MaxChars, o.MaxLines = c.Int("max-chars"), c.Int("lines")
	if o.MaxChars < 10 || o.MaxLines < 1 || o.MaxLines > 3 {
		return nil, Errf(ExitUsage, "bad-flag-value", "e.g. --max-chars 42 --lines 2 (shorts: --max-chars 20 --lines 1)", "--max-chars must be ≥ 10 and --lines 1–3")
	}
	ws, missing := captions.OutputWords(e.tl, e.tr)
	if len(missing) > 0 {
		return nil, Errf(ExitInvalid, "no-transcript", "cutroom transcribe "+missing[0], "the edit uses %s, which has no transcript", strings.Join(missing, ", "))
	}
	if len(ws) == 0 {
		return nil, Errf(ExitInvalid, "no-words", "check the transcript: `cutroom transcript`", "the timeline keeps no words to caption")
	}
	dur := e.tl.Duration()
	if c.Bool("burn") {
		return burnCaptions(c, e, ws, dur, o)
	}
	var lines []string
	var cues []captions.Cue
	out := ""
	if c.Has("check") {
		out = c.Str("check")
		if !filepath.IsAbs(out) {
			out = filepath.Join(e.p.Dir, out)
		}
		b, err := os.ReadFile(out)
		if err != nil {
			return nil, Errf(ExitUsage, "no-file", "`cutroom captions` writes out/captions.srt", "can't read %s", relToCwd(out))
		}
		if cues, err = captions.Parse(string(b)); err != nil {
			return nil, Errf(ExitInvalid, "bad-captions", "the file must be SRT or WebVTT", "%s: %v", relToCwd(out), err)
		}
		lines = append(lines, fmt.Sprintf("%s: %d cues, checked against the timeline (%d words, %s)", relToCwd(out), len(cues), len(ws), clock(dur)))
	} else {
		cues = captions.Build(ws, o, dur)
		out = captionPath(c, e)
		minD, maxD := 1e9, 0.0
		for _, q := range cues {
			minD, maxD = min(minD, q.End-q.Start), max(maxD, q.End-q.Start)
		}
		verb := "wrote"
		if c.Bool("dry-run") {
			verb = "would write"
		}
		lines = append(lines, fmt.Sprintf("%s %s: %d cues, %d words over %s (≤ %d chars × %d line(s), %.1f–%.1f s each)",
			verb, relToCwd(out), len(cues), len(ws), clock(dur), o.MaxChars, o.MaxLines, minD, maxD))
	}
	show := len(cues)
	if !c.Bool("all") && show > 8 {
		show = 6
	}
	for k, q := range cues[:show] {
		for i, l := range q.Lines {
			if i == 0 {
				lines = append(lines, fmt.Sprintf("  %3d  %s–%s  %s", k+1, clock(q.Start), clock(q.End), l))
			} else {
				lines = append(lines, fmt.Sprintf("  %3s  %17s  %s", "", "", l))
			}
		}
	}
	if show < len(cues) {
		lines = append(lines, fmt.Sprintf("  … %d more cues (--all, or read the file)", len(cues)-show))
	}
	issues := captions.Check(cues, ws, o, dur)
	fails := 0
	for _, is := range issues {
		st := map[string]string{"pass": "ok  ", "warn": "WARN", "fail": "FAIL"}[is.Status]
		lines = append(lines, fmt.Sprintf("%s %-8s %s", st, is.Check, is.Detail))
		if is.Fix != "" && is.Status != "pass" {
			lines = append(lines, "     → "+is.Fix)
		}
		if is.Status == "fail" {
			fails++
		}
	}
	data := map[string]any{"path": relToCwd(out), "cues": cues, "checks": issues, "words": len(ws), "duration": dur}
	if c.Has("check") {
		next := []string{"cutroom render"}
		if fails > 0 {
			next = []string{"cutroom captions   (writes a fresh file from the timeline)"}
			return &Result{Text: strings.Join(lines, "\n"), Data: data, Next: next, Exit: ExitCheckFail}, nil
		}
		return &Result{Text: strings.Join(lines, "\n"), Data: data, Next: next}, nil
	}
	if !c.Bool("dry-run") {
		if err := writeCaptionFile(e, out, cues, len(ws)); err != nil {
			return nil, err
		}
	}
	next := []string{"cutroom status"}
	if c.Bool("dry-run") {
		next = []string{"cutroom captions"}
	}
	return &Result{Text: strings.Join(lines, "\n"), Data: data, Next: next}, nil
}

func burnCaptions(c *Ctx, e *editEnv, ws []captions.Word, dur float64, o captions.Options) (*Result, error) {
	gs := captions.Groups(ws, captions.DefaultGroups, dur)
	var lines []string
	verb := "burning in"
	if c.Bool("dry-run") {
		verb = "would burn in"
	}
	existing := e.tl.Captions()
	lines = append(lines, fmt.Sprintf("%s captions: %d groups of 1–3 words (≤ 20 chars), %d words over %s", verb, len(gs), len(ws), clock(dur)))
	show := len(gs)
	if !c.Bool("all") && show > 12 {
		show = 10
	}
	for _, g := range gs[:show] {
		lines = append(lines, fmt.Sprintf("  %s–%s  %s", clock(g.Start), clock(g.End), g.Text))
	}
	if show < len(gs) {
		lines = append(lines, fmt.Sprintf("  … %d more (--all)", len(gs)-show))
	}
	id := ""
	switch {
	case existing != nil:
		id = existing.ID
		lines = append(lines, fmt.Sprintf("already on (%s): the groups follow the cut; the next render draws them", id))
	case c.Bool("dry-run"):
		lines = append(lines, "DRY RUN: nothing was changed; run the same command without --dry-run to apply it")
	default:
		id = e.tl.AddOverlay(timeline.Overlay{Kind: "captions", Layout: "bottom"})
		if err := timeline.Checkpoint(e.p.Dir, cacheDir(e.p), "captions --burn"); err != nil {
			return nil, err
		}
		if err := e.tl.Save(e.p.Dir); err != nil {
			return nil, err
		}
		lines = append(lines, fmt.Sprintf("added %s: burned-in captions over the whole edit (lower thirds move up above them); remove with `cutroom overlay rm %s`", id, id))
	}
	// the sidecar file comes with it: burned captions for muted viewers, the file for YouTube
	side := captionPath(c, e)
	cues := captions.Build(ws, o, dur)
	if c.Bool("dry-run") {
		lines = append(lines, fmt.Sprintf("would also write %s (%d cues; the subtitle file for YouTube — name it with --output)", relToCwd(side), len(cues)))
	} else {
		if err := writeCaptionFile(e, side, cues, len(ws)); err != nil {
			return nil, err
		}
		lines = append(lines, fmt.Sprintf("also wrote %s (%d cues; the subtitle file for YouTube — name it with --output, e.g. --output out/final.srt)", relToCwd(side), len(cues)))
	}
	mid := 0.0
	if len(gs) > 0 {
		mid = (gs[0].Start + gs[0].End) / 2
	}
	return &Result{Text: strings.Join(lines, "\n"), Data: map[string]any{"groups": gs, "overlay": id, "file": relToCwd(side)},
		Next: []string{fmt.Sprintf("cutroom look --output %.2f   (see them)", mid), "cutroom render --preview"}}, nil
}

// writeCaptionFile writes cues as SRT (or VTT by extension) with the sidecar status reads.
func writeCaptionFile(e *editEnv, out string, cues []captions.Cue, words int) error {
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(out, []byte(captions.Format(cues, strings.HasSuffix(strings.ToLower(out), ".vtt"))), 0o644); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(captionsMeta{TimelineHash: render.CutHash(e.tl), Cues: len(cues), Words: words, WrittenAt: time.Now().UTC().Truncate(time.Second)}, "", "  ")
	return os.WriteFile(captionsMetaPath(out), b, 0o644)
}

// captionPath: --output, else out/captions.srt (.vtt with --vtt).
func captionPath(c *Ctx, e *editEnv) string {
	if c.Has("output") {
		if out := c.Str("output"); filepath.IsAbs(out) {
			return out
		} else {
			return filepath.Join(e.p.Dir, out)
		}
	}
	ext := ".srt"
	if c.Bool("vtt") {
		ext = ".vtt"
	}
	return filepath.Join(e.p.Dir, "out", "captions"+ext)
}
