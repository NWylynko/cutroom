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
			is = append(is, e.mixedIssue(g, tl.Audio))
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

// heardWords puts, for each video source whose sound in the edit comes from a synced recorder, the
// recorder's words (a close microphone: the better transcript, and the sound viewers hear) on the
// camera's clock in place of the camera's own transcript. Word matching, joins, retake warnings and
// check then follow what the render actually says.
func heardWords(p *project.Project, tl *timeline.Timeline, tr timeline.Transcripts, dur map[string]float64) {
	if tl == nil {
		return
	}
	rs := renderSources(p)
	for _, v := range p.IDs() {
		s := p.Sources[v]
		if s.Facts == nil || s.Facts.Video == nil {
			continue
		}
		a, _, err := render.AudioFor(tl, rs, v)
		if err != nil || a == "" || a == v || tr[a] == nil || p.Sources[a] == nil || p.Sources[a].Sync == nil || p.Sources[a].Sync.To != v {
			continue
		}
		sy := p.Sources[a].Sync
		cam := tr[v]
		tr[v] = tr[a].MapTo(v, func(r float64) float64 { return (r - sy.Offset) / (1 + sy.Drift) }, dur[v])
		tr[v].ReconcileTiming(cam, 1.2) // words the recorder couldn't place: the camera's own microphone heard them
	}
}

func (e *editEnv) validate() []timeline.Issue {
	is := append(append([]timeline.Issue{}, e.notes...), e.audioIssues()...)
	is = append(is, e.tl.Validate(e.dur, e.tr)...)
	is = append(is, e.safeZoneIssues()...)
	for i := range is { // decided by the editor: listed, not open
		if ok, _ := e.tl.IsAccepted(is[i]); ok && is[i].Level == "warn" {
			is[i].Level = "accepted"
		}
	}
	return is
}

// mixedIssue: a take using its own sound next to another source's. Its speech level and room tone
// (measured when transcribing) say whether the change will be heard: matching numbers make it a note.
func (e *editEnv) mixedIssue(g timeline.Segment, main string) timeline.Issue {
	is := timeline.Issue{Level: "warn", Code: "audio-mixed", Segment: g.ID, At: g.OutStart}
	recorderFix := fmt.Sprintf("if %s has its own recorder: `cutroom sync <recorder> --to %s` then `cutroom edit audio <recorder> --for %s`", g.Source, g.Source, g.Source)
	a, b := e.tr[g.Source], e.tr[main]
	if a == nil || b == nil || a.SpeechDB == 0 || b.SpeechDB == 0 {
		is.Message = fmt.Sprintf("segments from %s use %s's own sound while the rest use %s; if they were recorded on different microphones the change will be audible (transcribe both to compare their levels)", g.Source, g.Source, main)
		is.Fix = "same microphone (e.g. a pickup on the same camera): fine as it is; " + recorderFix
		return is
	}
	ds, dr := a.SpeechDB-b.SpeechDB, a.RoomTone-b.RoomTone
	nums := fmt.Sprintf("speech %.1f vs %.1f dB, room tone %.1f vs %.1f dB", a.SpeechDB, b.SpeechDB, a.RoomTone, b.RoomTone)
	if math.Abs(ds) < 3 && math.Abs(dr) < 4 {
		is.Level = "info"
		is.Message = fmt.Sprintf("segments from %s use %s's own sound next to %s; they match (%s %s / %s), so the change shouldn't be heard", g.Source, g.Source, main, nums, g.Source, main)
		is.Fix = fmt.Sprintf("to be sure, listen across the join: `cutroom listen --output %.2f-%.2f`", math.Max(0, g.OutStart-1.5), g.OutStart+1.5)
		return is
	}
	is.Message = fmt.Sprintf("segments from %s use %s's own sound next to %s, and they differ (%s %s / %s): the change will be audible", g.Source, g.Source, main, nums, g.Source, main)
	switch {
	case math.Abs(ds-dr) >= 4: // speech above the room changed: another microphone or room
		is.Fix = fmt.Sprintf("different room or microphone (speech stands %.0f dB above its room here, %.0f dB in %s): listen across the join (`cutroom listen --output %.2f-%.2f`); %s", a.SpeechDB-a.RoomTone, b.SpeechDB-b.RoomTone, main, math.Max(0, g.OutStart-1.5), g.OutStart+1.5, recorderFix)
	default: // same speech-to-room ratio: the same setup, recorded louder or quieter
		is.Fix = fmt.Sprintf("same setup, %.1f dB louder/quieter (speech and room tone moved together): the render's loudness pass evens the whole edit, not each take; listen across the join (`cutroom listen --output %.2f-%.2f`); %s", math.Abs(ds), math.Max(0, g.OutStart-1.5), g.OutStart+1.5, recorderFix)
	}
	return is
}

// safeZoneIssues: on a vertical frame, overlays must stay clear of the app's own title bar, buttons
// and caption area (L41): nothing above y 260, below y 1490 or in the right 120 px of 1080×1920.
func (e *editEnv) safeZoneIssues() []timeline.Issue {
	tl := e.tl
	if tl.Height <= tl.Width || len(tl.Overlays) == 0 {
		return nil
	}
	ps, err := render.Placements(tl, e.tr, e.p.Dir, cacheDir(e.p), tl.Width, tl.Height)
	if err != nil {
		return nil // the overlay's own problem is reported elsewhere (render says it)
	}
	var is []timeline.Issue
	for _, p := range ps {
		if p.Kind == "clip" && p.Full {
			continue // a full-frame cutaway is the picture itself
		}
		boxes := [][4]float64{p.Box}
		if p.Kind == "captions" {
			boxes = nil
			for _, g := range p.Groups {
				boxes = append(boxes, g.Box)
			}
		}
		if msg := unsafeBox(boxes, tl.Width, tl.Height); msg != "" {
			fix := fmt.Sprintf("`cutroom overlay set %s --layout <another>` or a smaller --size", p.ID)
			if p.Kind == "captions" {
				fix = "shorter caption groups can't be set yet: report it"
			}
			is = append(is, timeline.Issue{Level: "warn", Code: "overlay-unsafe", At: p.Start,
				Message: fmt.Sprintf("overlay %s (%s) %s: the app's own buttons and text cover that part of a vertical video (L41)", p.ID, p.Name, msg), Fix: fix})
		}
	}
	return is
}

// unsafeBox says how a box breaks the vertical safe zones ("" when it doesn't).
func unsafeBox(boxes [][4]float64, W, H int) string {
	top, bottom, right := render.SafeZone(W, H)
	for _, b := range boxes {
		x1, y0, y1 := int(b[2]*float64(W)), int(b[1]*float64(H)), int(b[3]*float64(H))
		switch {
		case y0 < top:
			return fmt.Sprintf("reaches up to y %d (safe from %d)", y0, top)
		case y1 > bottom:
			return fmt.Sprintf("reaches down to y %d (safe to %d)", y1, bottom)
		case x1 > right:
			return fmt.Sprintf("reaches right to x %d (safe to %d)", x1, right)
		}
	}
	return ""
}

// quietLevels measures, for the timeline's Quiet hook, the level (dB) of the sound the edit uses for
// a video source at given source times: its recorder's when synced (camera time → recorder time
// through the offset and drift), else its own. Each wav is read once, on first use.
func quietLevels(p *project.Project, tl *timeline.Timeline) func(string, []float64) []float64 {
	rs := renderSources(p)
	envs := map[string]*speech.Envelope{}
	return func(src string, ts []float64) []float64 {
		a, _, err := render.AudioFor(tl, rs, src)
		if err != nil || a == "" || p.Sources[a] == nil {
			return nil
		}
		toA := func(t float64) float64 { return t }
		if a != src {
			sy := p.Sources[a].Sync
			if sy == nil || sy.To != src {
				return nil
			}
			toA = func(t float64) float64 { return t*(1+sy.Drift) + sy.Offset }
		}
		env, ok := envs[a]
		if !ok {
			if x, err := srcWav(p, p.Sources[a]); err == nil {
				env = speech.NewEnvelope(x)
			}
			envs[a] = env
		}
		if env == nil {
			return nil
		}
		start := 0.0 // the wav begins at the first audio sample: source time = wav time + audio start
		if f := p.Sources[a].Facts; f != nil && f.Audio != nil {
			start = f.Audio.Start
		}
		out := make([]float64, len(ts))
		for i, t := range ts {
			r := toA(t) - start
			out[i] = env.Level(r-0.01, r+0.01)
		}
		return out
	}
}
