// Package timeline is the edit: an ordered list of kept source ranges plus output settings.
// Edges are snapped to the output frame grid, anchored on the source's audio start (L01, L02),
// so every duration shown is the duration the render will have.
package timeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const File = "timeline.json"

type Segment struct {
	ID       string  `json:"id"`
	Source   string  `json:"source"`
	In       float64 `json:"in"`  // source seconds, on the frame grid
	Out      float64 `json:"out"` // source seconds, on the frame grid
	Reason   string  `json:"reason,omitempty"`
	PinIn    bool    `json:"pin_in,omitempty"`  // set by explicit edits; tighten never moves it (L39)
	PinOut   bool    `json:"pin_out,omitempty"` //
	Zoom     float64 `json:"zoom,omitempty"`    // 1 = none
	AnchorX  float64 `json:"anchor_x,omitempty"`
	AnchorY  float64 `json:"anchor_y,omitempty"`
	OutStart float64 `json:"-"` // computed: output time of In
}

type Timeline struct {
	Version    int    `json:"version"`
	FPS        string `json:"fps"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	SampleRate int    `json:"sample_rate"`
	Video      string `json:"video"` // default video source
	Audio      string `json:"audio"` // source whose audio is used
	// AudioFor maps a video source to the audio source its segments use (multi-take: each take's own
	// recorder). Unmapped sources use Audio when Audio is them or is synced to them, else their own sound.
	AudioFor  map[string]string `json:"audio_for,omitempty"`
	PadBefore float64           `json:"pad_before"`  // kept before a word at a cut (s)
	PadAfter  float64           `json:"pad_after"`   // kept after a word at a cut (s)
	Grid      float64           `json:"grid_origin"` // frame grid origin in source time (audio start)
	Segments  []Segment         `json:"segments"`
	NextID    int               `json:"next_id"`
	Modified  time.Time         `json:"modified"`
	fpsNum    float64
}

var ErrNone = errors.New("no timeline yet")

func Path(dir string) string { return filepath.Join(dir, File) }

func Load(dir string) (*Timeline, error) {
	b, err := os.ReadFile(Path(dir))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNone
	}
	if err != nil {
		return nil, err
	}
	t := &Timeline{}
	if err := json.Unmarshal(b, t); err != nil {
		return nil, fmt.Errorf("%s is not valid JSON: %w", File, err)
	}
	t.init()
	t.Recompute()
	return t, nil
}

func (t *Timeline) init() {
	t.fpsNum = rate(t.FPS)
	if t.fpsNum <= 0 {
		t.fpsNum = 30
	}
	if t.PadAfter == 0 && t.PadBefore == 0 {
		t.PadBefore, t.PadAfter = 0.10, 0.22
	}
}

// New creates a timeline holding one whole source.
func New(fps string, w, h int, video, audio string, grid, duration float64) *Timeline {
	t := &Timeline{Version: 1, FPS: fps, Width: w, Height: h, SampleRate: 48000, Video: video, Audio: audio,
		PadBefore: 0.10, PadAfter: 0.22, Grid: grid, NextID: 1}
	t.init()
	end := t.snapDown(duration)
	t.Segments = []Segment{{ID: t.newID(), Source: video, In: t.Snap(grid), Out: end, Reason: "whole source"}}
	t.Recompute()
	return t
}

func (t *Timeline) newID() string {
	id := "s" + strconv.Itoa(t.NextID)
	t.NextID++
	return id
}

func (t *Timeline) Rate() float64 { return t.fpsNum }

// Snap rounds a source time to the nearest frame boundary of the grid.
func (t *Timeline) Snap(s float64) float64 {
	k := math.Round((s - t.Grid) * t.fpsNum)
	return round6(t.Grid + k/t.fpsNum)
}

func (t *Timeline) snapDown(s float64) float64 {
	k := math.Floor((s-t.Grid)*t.fpsNum + 1e-6)
	return round6(t.Grid + k/t.fpsNum)
}

// Frames is the number of output frames of a segment.
func (t *Timeline) Frames(s Segment) int64 {
	return int64(math.Round((s.Out - s.In) * t.fpsNum))
}

// Duration is the output duration, computed from frames.
func (t *Timeline) Duration() float64 {
	var n int64
	for _, s := range t.Segments {
		n += t.Frames(s)
	}
	return round6(float64(n) / t.fpsNum)
}

// Recompute sets each segment's output start.
func (t *Timeline) Recompute() {
	var n int64
	for i := range t.Segments {
		t.Segments[i].OutStart = round6(float64(n) / t.fpsNum)
		n += t.Frames(t.Segments[i])
	}
}

// ToOutput maps a source time to output time; ok=false if that moment was cut.
func (t *Timeline) ToOutput(source string, s float64) (float64, bool) {
	for _, g := range t.Segments {
		if g.Source == source && s >= g.In && s < g.Out {
			return round6(g.OutStart + s - g.In), true
		}
	}
	return 0, false
}

// ToSource maps an output time to (segment, source time).
func (t *Timeline) ToSource(o float64) (*Segment, float64, bool) {
	for i := range t.Segments {
		g := &t.Segments[i]
		d := float64(t.Frames(*g)) / t.fpsNum
		if o >= g.OutStart && o < g.OutStart+d {
			return g, round6(g.In + o - g.OutStart), true
		}
	}
	return nil, 0, false
}

// Kept reports whether source time s of source is in the timeline.
func (t *Timeline) Kept(source string, s float64) bool {
	_, ok := t.ToOutput(source, s)
	return ok
}

// Cut removes [a, b) of a source from every segment. pin marks the new edges as pinned.
// Returns the source duration actually removed.
func (t *Timeline) Cut(source string, a, b float64, reason string, pin bool) float64 {
	a, b = t.Snap(a), t.Snap(b)
	if b <= a {
		return 0
	}
	removed := 0.0
	var out []Segment
	for _, g := range t.Segments {
		if g.Source != source || b <= g.In || a >= g.Out {
			out = append(out, g)
			continue
		}
		removed += math.Min(b, g.Out) - math.Max(a, g.In)
		if a > g.In {
			left := g
			left.Out = a
			left.PinOut = pin || (g.PinOut && a >= g.Out)
			if pin {
				left.Reason = joinReason(g.Reason, "cut after: "+reason)
			}
			out = append(out, left)
		}
		if b < g.Out {
			right := g
			right.In = b
			right.PinIn = pin
			if a > g.In {
				right.ID = t.newID()
			}
			if pin {
				right.Reason = joinReason(g.Reason, "cut before: "+reason)
			}
			out = append(out, right)
		}
	}
	t.Segments = dropTiny(out, t.fpsNum)
	t.Recompute()
	return round6(removed)
}

// Keep restricts the timeline to [a, b) of a source (everything else is removed).
func (t *Timeline) Keep(source string, a, b float64, reason string) {
	a, b = t.Snap(a), t.Snap(b)
	var out []Segment
	for _, g := range t.Segments {
		if g.Source != source || b <= g.In || a >= g.Out {
			continue
		}
		if a > g.In {
			g.In, g.PinIn = a, true
		}
		if b < g.Out {
			g.Out, g.PinOut = b, true
		}
		g.Reason = joinReason(g.Reason, "keep: "+reason)
		out = append(out, g)
	}
	t.Segments = dropTiny(out, t.fpsNum)
	t.Recompute()
}

// Zoom punches in on [a, b) of a source (L40): segments are split at a and b and the pieces inside
// get zoom z centred on (ax, ay) — normalised display coordinates; the crop stays inside the frame.
// z = 1 zooms back out. Split edges are pinned: they sit inside continuous speech, so tighten must
// never move them. Returns how many segments now carry the zoom.
func (t *Timeline) Zoom(source string, a, b, z, ax, ay float64) int {
	a, b = t.Snap(a), t.Snap(b)
	if b <= a {
		return 0
	}
	if z == 1 {
		ax, ay, z = 0, 0, 0
	}
	n := 0
	var out []Segment
	for _, g := range t.Segments {
		if g.Source != source || b <= g.In || a >= g.Out {
			out = append(out, g)
			continue
		}
		lo, hi := math.Max(a, g.In), math.Min(b, g.Out)
		m := g
		m.In, m.Out = lo, hi
		if lo > g.In {
			left := g
			left.Out, left.PinOut = lo, true
			out = append(out, left)
			m.ID, m.PinIn = t.newID(), true
		}
		m.Zoom, m.AnchorX, m.AnchorY = z, ax, ay
		if hi < g.Out {
			m.PinOut = true
		}
		out = append(out, m)
		n++
		if hi < g.Out {
			right := g
			right.ID, right.In, right.PinIn = t.newID(), hi, true
			out = append(out, right)
		}
	}
	t.Segments = dropTiny(out, t.fpsNum)
	t.Recompute()
	return n
}

// Insert adds [a, b) of a source as a new segment at position pos (0 = first, len = last) — a span
// of another take, or of an earlier part of the same one. Its edges are pinned (explicit). Returns
// the new segment's id.
func (t *Timeline) Insert(source string, a, b float64, pos int, reason string) string {
	g := Segment{ID: t.newID(), Source: source, In: t.Snap(a), Out: t.Snap(b), Reason: reason, PinIn: true, PinOut: true}
	pos = max(0, min(pos, len(t.Segments)))
	t.Segments = append(t.Segments[:pos], append([]Segment{g}, t.Segments[pos:]...)...)
	t.Recompute()
	return g.ID
}

// Move puts segment id at position pos of the timeline without it (0 = first). Returns false if
// there is no such segment.
func (t *Timeline) Move(id string, pos int) bool {
	i := t.Index(id)
	if i < 0 {
		return false
	}
	g := t.Segments[i]
	rest := append(append([]Segment{}, t.Segments[:i]...), t.Segments[i+1:]...)
	pos = max(0, min(pos, len(rest)))
	t.Segments = append(rest[:pos], append([]Segment{g}, rest[pos:]...)...)
	t.Recompute()
	return true
}

// Index is a segment's position, or -1.
func (t *Timeline) Index(id string) int {
	for i, g := range t.Segments {
		if g.ID == id {
			return i
		}
	}
	return -1
}

// ZoomOf returns a segment's zoom factor (1 when unset).
func (g Segment) ZoomOf() float64 {
	if g.Zoom <= 0 {
		return 1
	}
	return g.Zoom
}

// Anchor returns the zoom centre, defaulting to (0.5, 0.4): a face slightly above centre.
func (g Segment) Anchor() (float64, float64) {
	if g.AnchorX == 0 && g.AnchorY == 0 {
		return 0.5, 0.4
	}
	return g.AnchorX, g.AnchorY
}

func dropTiny(s []Segment, fps float64) []Segment {
	var out []Segment
	for _, g := range s {
		if math.Round((g.Out-g.In)*fps) >= 1 {
			out = append(out, g)
		}
	}
	return out
}

func joinReason(a, b string) string {
	if a == "" || a == "whole source" {
		return b
	}
	return a + "; " + b
}

// SetEdge moves a segment edge (used by tighten); refuses pinned edges.
func (t *Timeline) SetEdge(id string, in bool, s float64) bool {
	for i := range t.Segments {
		g := &t.Segments[i]
		if g.ID != id {
			continue
		}
		s = t.Snap(s)
		if in {
			if g.PinIn || s >= g.Out {
				return false
			}
			g.In = s
		} else {
			if g.PinOut || s <= g.In {
				return false
			}
			g.Out = s
		}
		t.Recompute()
		return true
	}
	return false
}

func (t *Timeline) Find(id string) *Segment {
	for i := range t.Segments {
		if t.Segments[i].ID == id {
			return &t.Segments[i]
		}
	}
	return nil
}

func (t *Timeline) IDs() []string {
	var ids []string
	for _, s := range t.Segments {
		ids = append(ids, s.ID)
	}
	return ids
}

// --- persistence and history ----------------------------------------------------------------

func (t *Timeline) Save(dir string) error {
	t.Modified = time.Now().UTC().Truncate(time.Second)
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	tmp := Path(dir) + ".tmp"
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, Path(dir))
}

type HistoryEntry struct {
	N        int     `json:"n"`
	Op       string  `json:"op"`
	Time     string  `json:"time"`
	Duration float64 `json:"duration_after"`
}

func histDir(cache string) string { return filepath.Join(cache, "history") }

// Checkpoint stores the current timeline file before an edit, labelled with the op.
func Checkpoint(dir, cache, op string) error {
	b, err := os.ReadFile(Path(dir))
	if errors.Is(err, os.ErrNotExist) {
		b = []byte("null")
	} else if err != nil {
		return err
	}
	if err := os.MkdirAll(histDir(cache), 0o755); err != nil {
		return err
	}
	n := len(History(cache)) + 1
	rec := map[string]any{"op": op, "time": time.Now().UTC().Format(time.RFC3339), "before": json.RawMessage(b)}
	out, _ := json.Marshal(rec)
	return os.WriteFile(filepath.Join(histDir(cache), fmt.Sprintf("%05d.json", n)), out, 0o644)
}

// History lists checkpoints, oldest first.
func History(cache string) []HistoryEntry {
	ents, _ := os.ReadDir(histDir(cache))
	var names []string
	for _, e := range ents {
		if strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	var out []HistoryEntry
	for i, n := range names {
		b, _ := os.ReadFile(filepath.Join(histDir(cache), n))
		var rec struct {
			Op   string `json:"op"`
			Time string `json:"time"`
		}
		json.Unmarshal(b, &rec)
		out = append(out, HistoryEntry{N: i + 1, Op: rec.Op, Time: rec.Time})
	}
	return out
}

// Undo restores the timeline from before the last edit. Returns the undone op.
func Undo(dir, cache string) (string, error) {
	ents, _ := os.ReadDir(histDir(cache))
	var names []string
	for _, e := range ents {
		if strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		return "", errors.New("nothing to undo")
	}
	sort.Strings(names)
	last := filepath.Join(histDir(cache), names[len(names)-1])
	b, err := os.ReadFile(last)
	if err != nil {
		return "", err
	}
	var rec struct {
		Op     string          `json:"op"`
		Before json.RawMessage `json:"before"`
	}
	if err := json.Unmarshal(b, &rec); err != nil {
		return "", err
	}
	if string(rec.Before) == "null" {
		os.Remove(Path(dir))
	} else if err := os.WriteFile(Path(dir), rec.Before, 0o644); err != nil {
		return "", err
	}
	return rec.Op, os.Remove(last)
}

func rate(s string) float64 {
	if n, d, ok := strings.Cut(s, "/"); ok {
		a, _ := strconv.ParseFloat(n, 64)
		b, _ := strconv.ParseFloat(d, 64)
		if b == 0 {
			return 0
		}
		return a / b
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func round6(f float64) float64 { return math.Round(f*1e6) / 1e6 }

// Trim returns a copy of the edit holding only output time [a, b), widened to whole frames: segments
// outside are dropped, the ones crossing an edge are clipped, and output time restarts at 0. start is
// the output time (in t) of the copy's first frame.
func (t *Timeline) Trim(a, b float64) (c *Timeline, start float64) {
	fa := int64(math.Floor(a*t.fpsNum + 1e-6))
	fb := int64(math.Ceil(b*t.fpsNum - 1e-6))
	cp := *t
	cp.Segments = nil
	var n int64
	for _, g := range t.Segments {
		f := t.Frames(g)
		s0 := n
		n += f
		lo, hi := max(s0, fa), min(s0+f, fb)
		if hi <= lo {
			continue
		}
		h := g
		h.In = round6(g.In + float64(lo-s0)/t.fpsNum)
		h.Out = round6(g.In + float64(hi-s0)/t.fpsNum)
		cp.Segments = append(cp.Segments, h)
	}
	cp.Recompute()
	return &cp, round6(float64(fa) / t.fpsNum)
}
