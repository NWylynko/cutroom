package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cutroom/cutroom/internal/deps"
	"github.com/cutroom/cutroom/internal/render"
	"github.com/cutroom/cutroom/internal/verify"
)

func init() {
	Register(&Command{
		Name:     "check",
		Args:     "[render.mp4]",
		MaxArgs:  1,
		Category: "verify",
		Summary:  "Verify a render: frames, picture vs source, A/V sync per segment, every word, stray sounds, loudness",
		Help: `Measures the rendered file itself (it does not trust the renderer):
  frames      frame count equals what the timeline rendered; audio length = video length
  picture     sampled frames match their source frames (catches shifted or frozen video)
  sync        per segment, the render's sound vs the source's sound at the moment its picture shows
              (limit 45 ms). picture + sync together prove lip sync is as the source plays in a
              player, timestamps and audio start offsets included. They can't repair a source
              recorded out of sync; to judge that, look at a face with ` + "`cutroom look --output`" + `.
  words       re-transcribes the render and compares with the timeline's words, in order
              (missing words = a clipped cut; extra words = a fragment of something you cut)
  stray-sound sound with no words kept in the video (breaths, bumps, pieces of cut takes)
  loudness    integrated loudness and peak vs the render settings
Exit code 4 if anything fails. Each problem says where (output time, segment) and how to look at it.
Word differences are often the recogniser mishearing; settle them with ` + "`cutroom listen`" + `.
Default file: out/final.mp4 (else out/preview.mp4).`,
		Flags: []Flag{{Name: "all", Kind: "bool", Help: "Don't truncate the list of problems"}},
		Run:   runCheck,
	})
}

func runCheck(c *Ctx) (*Result, error) {
	e, err := loadEditEnv(c, false)
	if err != nil {
		return nil, err
	}
	file := ""
	if len(c.Args) == 1 {
		file = c.Args[0]
		if !filepath.IsAbs(file) {
			file, _ = filepath.Abs(file)
		}
	} else {
		for _, n := range []string{"final.mp4", "preview.mp4"} {
			p := filepath.Join(e.p.Dir, "out", n)
			if _, err := os.Stat(p); err == nil {
				file = p
				break
			}
		}
		if file == "" {
			return nil, Errf(ExitUsage, "no-render", "render first: `cutroom render --preview`", "no out/final.mp4 or out/preview.mp4")
		}
	}
	if _, err := os.Stat(file); err != nil {
		fix := "check the path, or render first"
		if have := rendersIn(filepath.Join(e.p.Dir, "out")); len(have) > 0 {
			fix = "renders here: " + strings.Join(have, ", ") + " (`render --range` writes out/preview-range.mp4)"
		}
		return nil, Errf(ExitUsage, "no-render", fix, "%s not found", file)
	}
	m, err := render.LoadManifest(file)
	if err != nil {
		return nil, Errf(ExitUsage, "no-manifest", "check only works on files made by `cutroom render` (it reads "+filepath.Base(render.ManifestPath(file))+")", "%s has no render manifest", relToCwd(file))
	}
	model, _ := deps.FindModel()
	rep, err := verify.Check(file, m, verify.Options{Transcripts: e.tr, Model: model, CacheDir: cacheDir(e.p), CurrentHash: render.TimelineHash(e.tl)})
	if err != nil {
		return nil, Errf(ExitExternal, "check-failed", "make sure the file plays; re-render if it was interrupted", "%v", err)
	}
	var lines []string
	verdict := "PASS"
	if !rep.Pass {
		verdict = "FAIL"
	}
	lines = append(lines, fmt.Sprintf("%s %s — %d failure(s), %d warning(s)", verdict, relToCwd(file), rep.Fails, rep.Warnings))
	for _, it := range rep.Items {
		mark := map[string]string{"pass": "ok  ", "warn": "WARN", "fail": "FAIL"}[it.Status]
		where := ""
		if it.At > 0 || it.Segment != "" {
			where = fmt.Sprintf(" [%s %s]", clock(it.At), it.Segment)
		}
		l := fmt.Sprintf("%s %-11s %s%s", mark, it.Check, it.Detail, where)
		if it.Fix != "" && it.Status != "pass" {
			l += "\n     → " + it.Fix
		}
		lines = append(lines, l)
	}
	lines = truncLines(lines, c.Bool("all"))
	res := &Result{Text: strings.Join(lines, "\n"), Data: rep}
	if !rep.Pass {
		res.Exit = ExitCheckFail
		res.Next = []string{"fix the FAIL lines, render again, then `cutroom check`"}
	} else if rep.Warnings > 0 {
		res.Next = []string{"read each WARN: decide it is fine or fix it (they don't fail the check)"}
	}
	return res, nil
}

// rendersIn lists files in dir that have a render manifest, as paths relative to the cwd.
func rendersIn(dir string) []string {
	ms, _ := filepath.Glob(filepath.Join(dir, "*.cutroom.json"))
	var out []string
	for _, m := range ms {
		f := strings.TrimSuffix(m, ".cutroom.json")
		if _, err := os.Stat(f); err == nil {
			out = append(out, relToCwd(f))
		}
	}
	return out
}
