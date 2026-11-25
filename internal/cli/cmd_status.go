package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/NWylynko/cutroom/internal/render"
	"github.com/NWylynko/cutroom/internal/speech"
	"github.com/NWylynko/cutroom/internal/timeline"
	"github.com/NWylynko/cutroom/internal/verify"
)

func init() {
	Register(&Command{
		Name:     "status",
		Category: "start",
		Summary:  "Where the project stands: sources, timeline, renders (current? checked?), the next step",
		Help: `For an agent with no memory of earlier work: reads cutroom.json, timeline.json, .cutroom/ and
out/ and says what exists, what is stale and what to do next. Changes nothing.
  sources   kind, size, length; transcribed?; changed on disk since inspected?; recorder sync
  timeline  segments, length, sound source, edits so far, problems (as ` + "`cutroom timeline`" + ` reports them)
  renders   every file cutroom rendered in out/: current or made from an older timeline, and the
            last ` + "`cutroom check`" + ` result on it (if the file hasn't changed since)`,
		Run: runStatus,
	})
}

type statusSource struct {
	ID          string `json:"id"`
	Path        string `json:"path"`
	Summary     string `json:"summary"`
	Transcribed bool   `json:"transcribed"`
	Changed     bool   `json:"changed_on_disk"`
	Sync        string `json:"sync,omitempty"`
}

type statusTimeline struct {
	Segments int              `json:"segments"`
	Duration float64          `json:"duration"`
	Audio    string           `json:"audio_source"`
	Edits    int              `json:"edits"`
	LastEdit string           `json:"last_edit,omitempty"`
	Errors   int              `json:"errors"`
	Warnings int              `json:"warnings"`
	Issues   []timeline.Issue `json:"issues,omitempty"`
	Hash     string           `json:"hash"`
	history  []timeline.HistoryEntry
}

type statusRender struct {
	Path    string    `json:"path"`
	Preview bool      `json:"preview"`
	Partial bool      `json:"partial"`
	Current bool      `json:"current"` // made from the timeline as it is now
	Check   *checkRec `json:"check,omitempty"`
	mtime   time.Time
}

type statusData struct {
	Project  string           `json:"project"`
	Sources  []statusSource   `json:"sources"`
	Timeline *statusTimeline  `json:"timeline,omitempty"`
	Renders  []statusRender   `json:"renders"`
	Captions []statusCaptions `json:"captions,omitempty"`
}

type statusCaptions struct {
	Path    string `json:"path"`
	Current bool   `json:"current"`
	Cues    int    `json:"cues"`
}

// captionFilesIn lists .srt/.vtt files in dir.
func captionFilesIn(dir string) []string {
	var out []string
	ents, _ := os.ReadDir(dir)
	for _, e := range ents {
		if n := strings.ToLower(e.Name()); strings.HasSuffix(n, ".srt") || strings.HasSuffix(n, ".vtt") {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	return out
}

func runStatus(c *Ctx) (*Result, error) {
	p, err := needProject(c.Dir)
	if err != nil {
		return nil, err
	}
	d := statusData{Project: absDir(p.Dir)}
	var lines, next []string
	lines = append(lines, "project: "+absDir(p.Dir))

	// sources
	untranscribed := ""
	lines = append(lines, fmt.Sprintf("sources (%d):", len(p.IDs())))
	for _, id := range p.IDs() {
		s := p.Sources[id]
		ss := statusSource{ID: id, Path: s.Path}
		if f := s.Facts; f != nil {
			if v := f.Video; v != nil {
				ss.Summary = fmt.Sprintf("video %dx%d %s fps", v.Width, v.Height, v.FPS)
				if f.Audio == nil {
					ss.Summary += ", no audio"
				}
			} else {
				ss.Summary = "audio"
			}
			ss.Summary += ", " + secs(f.Duration)
		}
		ss.Changed, _ = p.Stale(s)
		if _, err := speech.Load(cacheDir(p), id); err == nil {
			ss.Transcribed = true
		}
		if s.Sync != nil {
			ss.Sync = fmt.Sprintf("synced to %s (%+.3f s, %+.1f ppm)", s.Sync.To, s.Sync.Offset, s.Sync.PPM)
		}
		state := []string{}
		if ss.Transcribed {
			state = append(state, "transcribed")
		} else if s.Facts != nil && s.Facts.Audio != nil && s.Sync == nil {
			state = append(state, "not transcribed")
			if untranscribed == "" {
				untranscribed = id
			}
		}
		if ss.Sync != "" {
			state = append(state, ss.Sync)
		}
		if ss.Changed {
			state = append(state, "CHANGED on disk since inspected")
			next = append(next, "cutroom init")
		}
		lines = append(lines, fmt.Sprintf("  %-10s %s  %s  %s", id, s.Path, ss.Summary, strings.Join(state, ", ")))
		d.Sources = append(d.Sources, ss)
	}

	// timeline
	e, err := loadEditEnv(c, false)
	if err == nil {
		tl := &statusTimeline{Segments: len(e.tl.Segments), Duration: e.tl.Duration(), Audio: e.tl.Audio, Hash: render.TimelineHash(e.tl)}
		tl.history = timeline.History(cacheDir(p))
		tl.Edits = len(tl.history)
		if tl.Edits > 0 {
			tl.LastEdit = tl.history[tl.Edits-1].Op
		}
		var src float64
		for _, id := range p.IDs() {
			if p.Sources[id].Facts != nil && p.Sources[id].Facts.Video != nil && src == 0 {
				src = p.Sources[id].Facts.Duration
			}
		}
		for _, is := range e.validate() {
			switch is.Level {
			case "error":
				tl.Errors++
			case "warn":
				tl.Warnings++
			}
			tl.Issues = append(tl.Issues, is)
		}
		l := fmt.Sprintf("timeline: %d segment(s), %s", tl.Segments, secs(tl.Duration))
		if src > 0 {
			l += fmt.Sprintf(" (sources %s)", secs(src))
		}
		if tl.Audio != "" {
			l += ", sound from " + tl.Audio
		}
		l += fmt.Sprintf(", %d edit(s)", tl.Edits)
		if tl.LastEdit != "" {
			l += " (last: " + tl.LastEdit + "; `cutroom edit history` lists them all)"
		}
		lines = append(lines, l)
		if tl.Errors+tl.Warnings > 0 {
			var codes []string
			for _, is := range tl.Issues {
				if is.Level == "error" || is.Level == "warn" {
					codes = append(codes, fmt.Sprintf("%s at %s", is.Code, clock(is.At)))
				}
			}
			lines = append(lines, fmt.Sprintf("  %d error(s), %d warning(s): %s — `cutroom timeline` shows them with fixes", tl.Errors, tl.Warnings, strings.Join(truncList(codes, 4), "; ")))
		}
		d.Timeline = tl
	} else if ce, ok := err.(*Error); ok && ce.Code == "no-timeline" {
		lines = append(lines, "timeline: none yet (the first edit creates it)")
	} else {
		return nil, err
	}

	// renders
	checks := loadChecks(p.Dir)
	for _, f := range rendersIn(filepath.Join(p.Dir, "out")) {
		m, err := render.LoadManifest(f)
		if err != nil {
			continue
		}
		r := statusRender{Path: relToCwd(f), Preview: m.Preview, Partial: m.Partial}
		if d.Timeline != nil {
			r.Current = m.TimelineHash == d.Timeline.Hash
		}
		if st, err := os.Stat(f); err == nil {
			r.mtime = st.ModTime()
			if ck, ok := checks[checkKey(p.Dir, f)]; ok && ck.Size == st.Size() && ck.ModTime.Equal(st.ModTime()) {
				r.Check = &ck
			}
		}
		d.Renders = append(d.Renders, r)
	}
	sort.SliceStable(d.Renders, func(i, j int) bool { // full renders first, newest first
		if d.Renders[i].Preview != d.Renders[j].Preview {
			return !d.Renders[i].Preview
		}
		return d.Renders[i].mtime.After(d.Renders[j].mtime)
	})
	if len(d.Renders) == 0 {
		lines = append(lines, "renders: none")
	} else {
		lines = append(lines, fmt.Sprintf("renders (%d):", len(d.Renders)))
		for _, r := range d.Renders {
			kind := "full"
			if r.Preview {
				kind = "preview"
			}
			if r.Partial {
				kind += ", partial"
			}
			state := "current"
			if !r.Current {
				state = "STALE (timeline changed since)"
			}
			ck := "not checked"
			if r.Check != nil {
				ck = fmt.Sprintf("check %s (%d fail, %d warn)", map[bool]string{true: "PASS", false: "FAIL"}[r.Check.Pass], r.Check.Fails, r.Check.Warnings)
				if r.Check.First != "" {
					ck += ": " + r.Check.First
				}
			}
			lines = append(lines, fmt.Sprintf("  %-22s %-16s %s, %s", r.Path, kind, state, ck))
		}
	}

	// shorts made from this project (cutroom short): each is a project of its own
	if ms, _ := filepath.Glob(filepath.Join(p.Dir, "shorts", "*", "timeline.json")); len(ms) > 0 {
		var ss []string
		for _, m := range ms {
			if t, err := timeline.Load(filepath.Dir(m)); err == nil {
				ss = append(ss, fmt.Sprintf("%s (%s)", relToCwd(filepath.Dir(m)), secs(t.Duration())))
			}
		}
		lines = append(lines, fmt.Sprintf("shorts: %s — work on one with --project <dir>", strings.Join(ss, ", ")))
	}
	// caption files: current when written from this timeline
	for _, f := range captionFilesIn(filepath.Join(p.Dir, "out")) {
		b, err := os.ReadFile(captionsMetaPath(f))
		var m captionsMeta
		if err != nil || json.Unmarshal(b, &m) != nil {
			continue
		}
		cs := statusCaptions{Path: relToCwd(f), Current: d.Timeline != nil && m.TimelineHash == render.CutHash(e.tl), Cues: m.Cues}
		d.Captions = append(d.Captions, cs)
		state := "current"
		if !cs.Current {
			state = "STALE (timeline changed since): `cutroom captions` writes it again"
		}
		lines = append(lines, fmt.Sprintf("captions: %s (%d cues) %s", cs.Path, cs.Cues, state))
	}

	// next step: the first thing not done yet
	if len(next) == 0 {
		next = statusNext(d, untranscribed)
	}
	if k := len(next) - 1; k >= 0 && strings.HasPrefix(next[k], "done") {
		lines = append(lines, next[k])
		next = next[:k]
	}
	return &Result{Text: strings.Join(lines, "\n"), Data: d, Next: next}, nil
}

func statusNext(d statusData, untranscribed string) []string {
	if untranscribed != "" {
		return []string{"cutroom transcribe " + untranscribed}
	}
	if d.Timeline == nil {
		return []string{"cutroom transcript --suspects", "cutroom edit tighten"}
	}
	if d.Timeline.Errors > 0 {
		return []string{"cutroom timeline"}
	}
	var pre []string
	if d.Timeline.Warnings > 0 { // open warnings come first: fix each or decide it is fine
		pre = []string{fmt.Sprintf("cutroom timeline   (%d open warning(s): fix each — `--apply <code>@<time>` — or keep it — `--accept … --reason`)", d.Timeline.Warnings)}
		for _, is := range d.Timeline.Issues {
			if is.Level != "warn" {
				continue
			}
			if m := fixCmd.FindStringSubmatch(is.Fix); m != nil && !strings.Contains(m[1], "<") {
				pre = append(pre, fmt.Sprintf("cutroom timeline --apply %s@%.2f   (runs: %s)", is.Code, is.At, m[1]))
			}
			break
		}
	}
	return append(pre, statusRenderNext(d)...)
}

func statusRenderNext(d statusData) []string {
	var full, preview *statusRender
	for i := range d.Renders {
		r := &d.Renders[i]
		if !r.Current || r.Partial {
			continue
		}
		if !r.Preview && (full == nil || strings.HasSuffix(r.Path, "out/final.mp4")) {
			full = r // out/final.mp4 is the deliverable when it is current
		}
		if r.Preview && preview == nil {
			preview = r
		}
	}
	stale := ""
	for _, r := range d.Renders {
		if !r.Current && !r.Partial && !r.Preview {
			stale = r.Path
		}
	}
	switch {
	case full == nil && preview == nil && stale != "":
		return []string{"cutroom timeline --joins", fmt.Sprintf("cutroom render   (%s is stale: the edit changed after it was rendered)", stale)}
	case full == nil && preview == nil:
		return []string{"cutroom timeline --joins", "cutroom render --preview"}
	case full == nil:
		if preview.Check == nil {
			return []string{"cutroom check " + preview.Path}
		}
		if !preview.Check.Pass {
			return []string{"cutroom check " + preview.Path + " (fix its FAIL lines)"}
		}
		return []string{"cutroom render"}
	case full.Check == nil:
		return []string{"cutroom check " + full.Path}
	case !full.Check.Pass:
		return []string{"cutroom check " + full.Path + " (fix its FAIL lines, render again)"}
	}
	return []string{fmt.Sprintf("done: %s is current and passed check", full.Path)}
}

// --- check records -------------------------------------------------------------------------------

// checkRec is the last `cutroom check` verdict on a file, valid while the file is unchanged.
type checkRec struct {
	Pass     bool      `json:"pass"`
	Fails    int       `json:"fails"`
	Warnings int       `json:"warnings"`
	First    string    `json:"first_problem,omitempty"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"mtime"`
	At       string    `json:"checked_at"`
}

func checksPath(dir string) string { return filepath.Join(dir, ".cutroom", "checks.json") }

func checkKey(dir, file string) string {
	if r, err := filepath.Rel(dir, file); err == nil {
		return filepath.ToSlash(r)
	}
	return file
}

func loadChecks(dir string) map[string]checkRec {
	m := map[string]checkRec{}
	if b, err := os.ReadFile(checksPath(dir)); err == nil {
		json.Unmarshal(b, &m)
	}
	return m
}

func saveCheck(dir, file string, rep *verify.Report) {
	st, err := os.Stat(file)
	if err != nil {
		return
	}
	rec := checkRec{Pass: rep.Pass, Fails: rep.Fails, Warnings: rep.Warnings, Size: st.Size(), ModTime: st.ModTime(), At: time.Now().UTC().Format(time.RFC3339)}
	for _, it := range rep.Items {
		if it.Status == "fail" || (it.Status == "warn" && rec.First == "") {
			rec.First = it.Check
			if it.Status == "fail" {
				break
			}
		}
	}
	m := loadChecks(dir)
	m[checkKey(dir, file)] = rec
	b, _ := json.MarshalIndent(m, "", "  ")
	os.MkdirAll(filepath.Dir(checksPath(dir)), 0o755)
	os.WriteFile(checksPath(dir), b, 0o644)
}

func truncList(xs []string, n int) []string {
	if len(xs) <= n {
		return xs
	}
	return append(xs[:n:n], fmt.Sprintf("… %d more", len(xs)-n))
}
