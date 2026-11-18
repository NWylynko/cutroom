package cli

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/cutroom/cutroom/internal/media"
	"github.com/cutroom/cutroom/internal/project"
	"github.com/cutroom/cutroom/internal/timeline"
)

func init() {
	Register(&Command{
		Name:     "init",
		Args:     "[file…]",
		Category: "start",
		Summary:  "Create a project here (or add sources to it) and inspect each source",
		Help: `Sources are referenced, never copied or modified. Each gets a short id you use in other
commands. Ids default to the file name ("talk.mp4" → talk); set one with --source id=path.
Running init again adds new sources and re-inspects changed ones; it is safe to repeat.
Examples:
  cutroom init talk.mp4
  cutroom init --source cam=raw/IMG_0412.MOV --source mic=raw/zoom.wav`,
		Flags: []Flag{{Name: "source", Kind: "string", Many: true, Help: "id=path of a source file"},
			{Name: "remove", Kind: "string", Help: "Remove a source id from the project (the file itself is not touched)"}},
		Run: runInit,
	})
	Register(&Command{
		Name:     "inspect",
		Args:     "[source-id|file…]",
		Category: "perceive",
		Summary:  "Facts about sources (frame rate, VFR, rotation, HDR, audio offset, loudness) and what they mean for editing",
		Help: `With no argument: every source in the project. Warnings marked "handled" need nothing from
you; the others say what to do.
Examples:
  cutroom inspect
  cutroom inspect talk
  cutroom inspect some/other/file.mov`,
		Flags: []Flag{{Name: "fast", Kind: "bool", Help: "Skip the loudness measurement (it decodes all audio)"}},
		Run:   runInspect,
	})
}

func loadOrNew(dir string) (*project.Project, bool, error) {
	p, err := project.Load(dir)
	if errors.Is(err, project.ErrNoProject) {
		return project.New(dir), true, nil
	}
	if err != nil {
		return nil, false, Errf(ExitInvalid, "bad-project", "fix or delete "+filepath.Join(dir, project.File), "%v", err)
	}
	return p, false, nil
}

func needProject(dir string) (*project.Project, error) {
	p, err := project.Load(dir)
	if errors.Is(err, project.ErrNoProject) {
		return nil, Errf(ExitInvalid, "no-project", "run `cutroom init <video file>` in your working directory first (or pass --project <dir>)",
			"no %s in %s", project.File, absDir(dir))
	}
	if err != nil {
		return nil, Errf(ExitInvalid, "bad-project", "fix or delete "+filepath.Join(dir, project.File), "%v", err)
	}
	return p, nil
}

func absDir(d string) string { a, _ := filepath.Abs(d); return a }

func runInit(c *Ctx) (*Result, error) {
	if err := os.MkdirAll(c.Dir, 0o755); err != nil {
		return nil, Errf(ExitInvalid, "bad-project-dir", "pass a writable --project directory", "%v", err)
	}
	p, created, err := loadOrNew(c.Dir)
	if err != nil {
		return nil, err
	}
	if id := c.Str("remove"); id != "" {
		if created {
			return nil, Errf(ExitUsage, "no-project", "nothing to remove from", "no project here")
		}
		if _, err := sourceByID(p, id); err != nil {
			return nil, err
		}
		if tl, err := timeline.Load(p.Dir); err == nil {
			for _, g := range tl.Segments {
				if g.Source == id || tl.Audio == id {
					return nil, Errf(ExitInvalid, "source-in-use", "cut its segments or `cutroom edit audio <other>` first", "source %q is used by the timeline", id)
				}
			}
		}
		delete(p.Sources, id)
		if err := p.Save(); err != nil {
			return nil, err
		}
		return &Result{Text: fmt.Sprintf("removed source %q from %s (the file is untouched)", id, project.File), Data: map[string]string{"removed": id}}, nil
	}
	type pending struct{ id, path string }
	var todo []pending
	for _, s := range c.Strs("source") {
		id, path, ok := strings.Cut(s, "=")
		if !ok || id == "" || path == "" {
			return nil, Errf(ExitUsage, "bad-source", "write it as --source id=path, e.g. --source cam=talk.mp4", "bad --source %q", s)
		}
		if !project.ValidID(id) {
			return nil, Errf(ExitUsage, "bad-source-id", "ids are lowercase letters, digits, - and _, starting with a letter (max 32), e.g. cam, mic, take-2", "invalid source id %q", id)
		}
		todo = append(todo, pending{id, path})
	}
	for _, a := range c.Args {
		todo = append(todo, pending{"", a})
	}
	if len(todo) == 0 && created {
		return nil, Errf(ExitUsage, "no-sources", "name at least one file: `cutroom init talk.mp4`", "init needs source files")
	}

	var lines []string
	var added, refreshed, unchanged []string
	for _, t := range todo {
		path := t.path
		if !filepath.IsAbs(path) {
			// paths are relative to the current directory, like any CLI
			path, _ = filepath.Abs(path)
		}
		if _, err := os.Stat(path); err != nil {
			return nil, Errf(ExitUsage, "source-not-found", "check the path; it is relative to your current directory", "%s: no such file", t.path)
		}
		existing := p.FindByPath(path)
		id := t.id
		if existing != nil {
			if id != "" && id != existing.ID {
				return nil, Errf(ExitUsage, "source-exists", fmt.Sprintf("refer to it as %q", existing.ID), "%s is already source %q", t.path, existing.ID)
			}
			id = existing.ID
			if stale, _ := p.Stale(existing); !stale {
				unchanged = append(unchanged, id)
				continue
			}
		} else {
			if id == "" {
				id = p.SuggestID(path)
			} else if p.Sources[id] != nil {
				return nil, Errf(ExitUsage, "id-taken", fmt.Sprintf("pick another id, e.g. %s", p.SuggestID(id+".x")), "source id %q is already used for %s", id, p.Sources[id].Path)
			}
		}
		f, err := media.Probe(path, media.Options{Loudness: true})
		if err != nil {
			return nil, Errf(ExitExternal, "probe-failed", "make sure it is a video or audio file ffmpeg can read", "%s: %v", t.path, err)
		}
		st, _ := os.Stat(path)
		p.Sources[id] = &project.Source{ID: id, Path: p.RelPath(path), Size: st.Size(), ModTime: st.ModTime(), Facts: f}
		if existing != nil {
			refreshed = append(refreshed, id)
		} else {
			added = append(added, id)
		}
	}
	if err := p.Save(); err != nil {
		return nil, Errf(ExitInvalid, "save-failed", "check the project directory is writable", "%v", err)
	}
	head := "project created: wrote " + filepath.Join(absDir(c.Dir), project.File)
	if !created {
		head = "project updated: " + filepath.Join(absDir(c.Dir), project.File)
	}
	lines = append(lines, head)
	for _, id := range p.IDs() {
		s := p.Sources[id]
		tag := ""
		switch {
		case contains(added, id):
			tag = " (added)"
		case contains(refreshed, id):
			tag = " (re-inspected: file changed)"
		}
		lines = append(lines, fmt.Sprintf("  %-10s %s%s", id, oneLine(s.Facts, s.Path), tag))
		for _, w := range s.Facts.Warnings {
			lines = append(lines, "             "+warnLine(w))
		}
	}
	next := []string{"cutroom inspect " + p.IDs()[0] + "   (full facts)"}
	return &Result{Text: strings.Join(lines, "\n"), Next: next, Data: map[string]any{
		"project": absDir(c.Dir), "created": created, "added": nz(added), "reinspected": nz(refreshed), "unchanged": nz(unchanged), "sources": p.Sources}}, nil
}

func nz(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func runInspect(c *Ctx) (*Result, error) {
	type item struct {
		ID    string       `json:"id,omitempty"`
		Facts *media.Facts `json:"facts"`
	}
	var items []item
	opt := media.Options{Loudness: !c.Bool("fast")}
	p, perr := project.Load(c.Dir)
	if len(c.Args) == 0 {
		if perr != nil {
			return nil, Errf(ExitUsage, "no-project", "name a file (`cutroom inspect talk.mp4`) or run `cutroom init <file>` first", "no project here and no file given")
		}
		for _, id := range p.IDs() {
			s := p.Sources[id]
			f, err := refresh(p, s, opt)
			if err != nil {
				return nil, err
			}
			items = append(items, item{id, f})
		}
	}
	for _, a := range c.Args {
		if perr == nil && p.Sources[a] != nil {
			f, err := refresh(p, p.Sources[a], opt)
			if err != nil {
				return nil, err
			}
			items = append(items, item{a, f})
			continue
		}
		if _, err := os.Stat(a); err != nil {
			fix := "give a source id from `cutroom inspect` or a file path"
			if perr == nil {
				if s := suggest(a, p.IDs()); s != "" {
					fix = fmt.Sprintf("did you mean source %q?", s)
				} else {
					fix = "source ids: " + strings.Join(p.IDs(), ", ")
				}
			}
			return nil, Errf(ExitUsage, "unknown-source", fix, "%q is neither a source id nor a file", a)
		}
		f, err := media.Probe(a, opt)
		if err != nil {
			return nil, Errf(ExitExternal, "probe-failed", "make sure it is a video or audio file ffmpeg can read", "%s: %v", a, err)
		}
		items = append(items, item{"", f})
	}
	var b strings.Builder
	for i, it := range items {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(describe(it.ID, it.Facts))
	}
	var next []string
	if perr != nil && len(c.Args) > 0 {
		next = append(next, "cutroom init "+c.Args[0])
	}
	return &Result{Text: b.String(), Data: items, Next: next}, nil
}

// refresh re-probes a project source if its file changed (and saves), else returns stored facts.
func refresh(p *project.Project, s *project.Source, opt media.Options) (*media.Facts, error) {
	stale, err := p.Stale(s)
	if err != nil {
		return nil, Errf(ExitInvalid, "source-missing", "restore the file, or re-add it with `cutroom init --source "+s.ID+"=<new path>`", "source %q: %s is missing", s.ID, s.Path)
	}
	if !stale && s.Facts != nil {
		return s.Facts, nil
	}
	f, err := media.Probe(p.Abs(s), opt)
	if err != nil {
		return nil, Errf(ExitExternal, "probe-failed", "make sure it is a video or audio file ffmpeg can read", "%s: %v", s.Path, err)
	}
	st, _ := os.Stat(p.Abs(s))
	s.Facts, s.Size, s.ModTime = f, st.Size(), st.ModTime()
	_ = p.Save()
	return f, nil
}

func oneLine(f *media.Facts, path string) string {
	var parts []string
	if v := f.Video; v != nil {
		parts = append(parts, fmt.Sprintf("video %dx%d %s fps", v.Width, v.Height, fpsStr(v.FPS)))
	}
	if a := f.Audio; a != nil {
		parts = append(parts, fmt.Sprintf("audio %d Hz %s", a.SampleRate, chans(a.Channels)))
	}
	return fmt.Sprintf("%s  %s, %s", path, strings.Join(parts, " + "), dur(f.Duration))
}

func describe(id string, f *media.Facts) string {
	var b strings.Builder
	name := id
	if name == "" {
		name = filepath.Base(f.Path)
	}
	fmt.Fprintf(&b, "%s  %s\n", name, f.Path)
	fmt.Fprintf(&b, "  %s, %s, %s\n", f.Kind, dur(f.Duration), size(f.Size))
	if v := f.Video; v != nil {
		timing := "constant frame rate"
		if v.VFR {
			timing = fmt.Sprintf("VARIABLE frame rate (%d gaps, longest %.3f s, avg %.3f fps)", v.Gaps, v.MaxGap, v.AvgFPS)
		}
		color := "SDR"
		if v.HDR != "" {
			color = "HDR " + strings.ToUpper(v.HDR)
			if v.DolbyVision {
				color += " + Dolby Vision"
			}
		}
		if v.Primaries != "" || v.Transfer != "" {
			color += fmt.Sprintf(" (%s/%s)", orDash(v.Primaries), orDash(v.Transfer))
		}
		rot := ""
		if v.Rotation != 0 {
			rot = fmt.Sprintf(", rotated %d° by metadata", v.Rotation)
		}
		fmt.Fprintf(&b, "  video  %s %dx%d%s, %s fps nominal, %d frames, %s\n", v.Codec, v.Width, v.Height, rot, fpsStr(v.FPS), v.Frames, timing)
		fmt.Fprintf(&b, "         %s, %d-bit %s\n", color, v.BitDepth, v.PixFmt)
	}
	if a := f.Audio; a != nil {
		fmt.Fprintf(&b, "  audio  %s %d Hz %s", a.Codec, a.SampleRate, chans(a.Channels))
		if f.Video != nil {
			fmt.Fprintf(&b, ", starts %+.3f s vs video", f.AVOffset)
		}
		if a.Measured {
			fmt.Fprintf(&b, ", %.1f LUFS, peak %.1f dBTP", a.LUFS, a.PeakDB)
		}
		b.WriteString("\n")
	}
	if len(f.Warnings) == 0 {
		b.WriteString("  no problems found\n")
	}
	for _, w := range f.Warnings {
		b.WriteString("  " + warnLine(w) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func warnLine(w media.Warning) string {
	tag := "ACTION "
	if w.Handled {
		tag = "handled"
	}
	s := fmt.Sprintf("%s %s: %s", tag, w.Code, w.Message)
	if w.Fix != "" && !w.Handled {
		s += " → " + w.Fix
	}
	return s
}

func fpsStr(s string) string {
	if strings.HasSuffix(s, "/1") {
		return strings.TrimSuffix(s, "/1")
	}
	switch s {
	case "30000/1001":
		return "29.97"
	case "60000/1001":
		return "59.94"
	case "24000/1001":
		return "23.976"
	}
	return s
}

func chans(n int) string {
	switch n {
	case 1:
		return "mono"
	case 2:
		return "stereo"
	}
	return fmt.Sprintf("%d ch", n)
}

func dur(s float64) string {
	if s >= 60 {
		m := math.Floor(s / 60)
		return fmt.Sprintf("%.0f:%06.3f (%.3f s)", m, s-m*60, s)
	}
	return fmt.Sprintf("%.3f s", s)
}

func size(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	}
	return fmt.Sprintf("%d KB", n>>10)
}

func orDash(s string) string {
	if s == "" || s == "unknown" {
		return "-"
	}
	return s
}
