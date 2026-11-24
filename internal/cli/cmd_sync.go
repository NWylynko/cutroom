package cli

import (
	"errors"
	"fmt"
	"github.com/cutroom/cutroom/internal/render"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/cutroom/cutroom/internal/align"
	"github.com/cutroom/cutroom/internal/project"
	"github.com/cutroom/cutroom/internal/speech"
	"github.com/cutroom/cutroom/internal/timeline"
)

func init() {
	Register(&Command{
		Name:     "sync",
		Args:     "<audio-source>",
		MinArgs:  1,
		MaxArgs:  1,
		Category: "edit",
		Summary:  "Line up a separately recorded audio track (e.g. a mic recorder) with a camera: offset + clock drift",
		Help: `Measures how <audio-source> lines up with --to (default: the first video source) by comparing their
sound: a coarse match over the whole recordings, then ~40 precise matches, then a robust fit of
offset and clock drift. The result is stored in cutroom.json; renders then use it automatically.
Next, make the timeline use that audio: ` + "`cutroom edit audio <audio-source>`" + `.
Both sources need to have recorded the same event (the camera's own mic is enough for matching).
Examples:
  cutroom init --source mic=recorder.wav
  cutroom sync mic
  cutroom edit audio mic`,
		Flags: []Flag{{Name: "to", Kind: "string", Help: "Reference (camera) source id; default: the first video source"}},
		Run:   runSync,
	})
}

func srcWav(p *project.Project, s *project.Source) ([]float64, error) {
	w := speech.WavPath(cacheDir(p), s.ID)
	// compare with the file as it is on disk now, not as recorded at inspect time
	src, _ := os.Stat(p.Abs(s))
	if st, err := os.Stat(w); err != nil || src == nil || st.ModTime().Before(src.ModTime()) {
		os.MkdirAll(filepath.Dir(w), 0o755)
		if err := speech.ExtractWav(p.Abs(s), w); err != nil {
			return nil, err
		}
	}
	return speech.ReadWav(w)
}

func runSync(c *Ctx) (*Result, error) {
	p, err := needProject(c.Dir)
	if err != nil {
		return nil, err
	}
	other, err := sourceByID(p, c.Args[0])
	if err != nil {
		return nil, err
	}
	var ref *project.Source
	if c.Has("to") {
		if ref, err = sourceByID(p, c.Str("to")); err != nil {
			return nil, err
		}
	} else {
		for _, id := range p.IDs() {
			if s := p.Sources[id]; s.Facts.Video != nil && id != other.ID {
				ref = s
				break
			}
		}
		if ref == nil {
			return nil, Errf(ExitUsage, "no-reference", "name the camera source: --to <id>", "no other video source to sync %q to", other.ID)
		}
	}
	if ref.ID == other.ID {
		return nil, Errf(ExitUsage, "same-source", "sync a separate audio source to the camera, e.g. `cutroom sync mic --to cam`", "a source can't be synced to itself")
	}
	for _, s := range []*project.Source{ref, other} {
		if s.Facts.Audio == nil {
			return nil, Errf(ExitUsage, "no-audio", "both sources need sound to be matched (the camera's own mic is enough)", "source %q has no audio", s.ID)
		}
	}
	for _, src := range []*project.Source{ref, other} {
		if err := notChanged(p, src); err != nil {
			return nil, err
		}
	}
	rx, err := srcWav(p, ref)
	if err != nil {
		return nil, Errf(ExitExternal, "decode-failed", "check the file plays", "%s: %v", ref.ID, err)
	}
	ox, err := srcWav(p, other)
	if err != nil {
		return nil, Errf(ExitExternal, "decode-failed", "check the file plays", "%s: %v", other.ID, err)
	}
	s, err := align.Measure(rx, ref.Facts.Audio.Start, ox, other.Facts.Audio.Start)
	if errors.Is(err, align.ErrNoMatch) {
		return nil, Errf(ExitCheckFail, "no-match", "make sure both files recorded the same event; `cutroom transcript` of each should show the same words", "%v", err)
	}
	if err != nil {
		return nil, err
	}
	s.To = ref.ID
	other.Sync = &s
	if err := p.Save(); err != nil {
		return nil, err
	}
	dur := ref.Facts.Duration
	driftMS := math.Abs(s.Drift) * dur * 1000
	first := fmt.Sprintf("%s started recording %.3f s before %s", other.ID, s.Offset, ref.ID)
	if s.Offset < 0 {
		first = fmt.Sprintf("%s started recording %.3f s after %s", other.ID, -s.Offset, ref.ID)
	}
	lines := []string{
		fmt.Sprintf("synced %s to %s: %s", other.ID, ref.ID, first),
		fmt.Sprintf("  mapping: %s time = %+.3f s + %s time × (1 %+.1f ppm)", other.ID, s.Offset, ref.ID, s.PPM),
		fmt.Sprintf("  offset %+.3f s; clock drift %+.1f ppm (%.1f ms over the %s of %s)", s.Offset, s.PPM, driftMS, clock(dur), ref.ID),
		fmt.Sprintf("  fit: %d windows, %d rejected, residual %.2f ms", s.Windows, s.Rejected, s.Residual),
	}
	var warn []string
	if s.Overlap[0] > 0.05 || s.Overlap[1] < dur-0.05 {
		warn = append(warn, fmt.Sprintf("WARN %s only covers %s %s–%s; parts of %s outside that have no %s audio", other.ID, ref.ID, clock(s.Overlap[0]), clock(s.Overlap[1]), ref.ID, other.ID))
	}
	if s.Residual > 5 {
		warn = append(warn, fmt.Sprintf("WARN the fit is loose (%.1f ms residual): listen to a few joins after rendering", s.Residual))
	}
	lines = append(lines, warn...)
	return &Result{Text: strings.Join(lines, "\n"), Data: s, Next: []string{"cutroom edit audio " + other.ID + "   (use it in the timeline)"}}, nil
}

// audioIssues: the timeline's audio must exist and be aligned to every video source it plays under.
func (e *editEnv) audioIssues() []timeline.Issue {
	tl := e.tl
	if tl.Audio == "" {
		return []timeline.Issue{{Level: "error", Code: "no-audio", Message: "the timeline has no audio source", Fix: "`cutroom edit audio <source>`"}}
	}
	if e.p.Sources[tl.Audio] == nil {
		return []timeline.Issue{{Level: "error", Code: "unknown-audio", Message: fmt.Sprintf("audio source %q is not in the project", tl.Audio), Fix: "`cutroom edit audio <source>`"}}
	}
	srcs := renderSources(e.p)
	var is []timeline.Issue
	seen := map[string]bool{}
	for _, g := range tl.Segments {
		if seen[g.Source] {
			continue
		}
		seen[g.Source] = true
		a, own, err := render.AudioFor(tl, srcs, g.Source)
		if err != nil {
			fix := fmt.Sprintf("`cutroom sync %s --to %s`", tl.Audio, g.Source)
			if m := tl.AudioFor[g.Source]; m != "" {
				fix = fmt.Sprintf("`cutroom sync %s --to %s`, or `cutroom edit audio %s --for %s` to use its own sound", m, g.Source, g.Source, g.Source)
			}
			is = append(is, timeline.Issue{Level: "error", Code: "audio-not-synced", Segment: g.ID, At: g.OutStart, Message: err.Error(), Fix: fix})
			continue
		}
		if own {
			is = append(is, timeline.Issue{Level: "warn", Code: "audio-mixed", Segment: g.ID, At: g.OutStart,
				Message: fmt.Sprintf("segments from %s use %s's own sound while the rest use %s; if they were recorded on different microphones the change will be audible", g.Source, g.Source, tl.Audio),
				Fix:     fmt.Sprintf("same microphone (e.g. a pickup on the same camera): fine as it is. If %s has its own recorder: `cutroom sync <recorder> --to %s` then `cutroom edit audio <recorder> --for %s`", g.Source, g.Source, g.Source)})
			continue
		}
		if a == g.Source {
			continue
		}
		rec := e.p.Sources[a]
		for _, h := range tl.Segments {
			if h.Source == g.Source && rec.Sync != nil && (h.In < rec.Sync.Overlap[0]-0.05 || h.Out > rec.Sync.Overlap[1]+0.05) {
				is = append(is, timeline.Issue{Level: "warn", Code: "audio-gap", Segment: h.ID, At: h.OutStart,
					Message: fmt.Sprintf("segment %s (%s %.2f–%.2f) is outside what %s recorded; it will be silent there", h.ID, h.Source, h.In, h.Out, a)})
			}
		}
	}
	return is
}

// renderSources describes the project's sources for render (and for resolving each take's audio).
func renderSources(p *project.Project) map[string]render.Source {
	srcs := map[string]render.Source{}
	for id, s := range p.Sources {
		rs := render.Source{Path: p.Abs(s), HasAudio: s.Facts != nil && s.Facts.Audio != nil, Size: s.Size, ModTime: s.ModTime, Sync: s.Sync}
		if rs.HasAudio {
			rs.AudioStart = s.Facts.Audio.Start
		}
		srcs[id] = rs
	}
	return srcs
}

func (e *editEnv) validate() []timeline.Issue {
	return append(e.audioIssues(), e.tl.Validate(e.dur, e.tr)...)
}
