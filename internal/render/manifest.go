package render

import (
	"encoding/json"
	"os"
	"time"

	"github.com/NWylynko/cutroom/internal/align"
	"github.com/NWylynko/cutroom/internal/timeline"
)

// Manifest records exactly what a render contains, so `check` can verify the file independently
// and detect when the timeline has changed since.
type Manifest struct {
	Version      int                `json:"version"`
	Output       string             `json:"output"`
	Project      string             `json:"project,omitempty"` // the project folder it was rendered from
	Preview      bool               `json:"preview"`
	FPS          string             `json:"fps"`
	Width        int                `json:"width"`
	Height       int                `json:"height"`
	Audio        string             `json:"audio_source"`
	Segments     []MSeg             `json:"segments"`
	Sources      map[string]MSource `json:"sources"`
	Loudness     float64            `json:"loudness_target"`
	TruePeak     float64            `json:"true_peak_ceiling"`
	Chain        string             `json:"audio_chain"`
	Frames       int64              `json:"frames"`
	Duration     float64            `json:"duration"`
	RangeStart   float64            `json:"range_start"`
	Partial      bool               `json:"partial"`
	Overlays     []Placement        `json:"overlays,omitempty"` // in file time
	Fill         bool               `json:"fill,omitempty"`
	FillX        float64            `json:"fill_x,omitempty"`
	FillY        float64            `json:"fill_y,omitempty"`
	Bars         string             `json:"bars,omitempty"`
	TimelineHash string             `json:"timeline_hash"`
	RenderedAt   time.Time          `json:"rendered_at"`
}

type MSeg struct {
	ID       string  `json:"id"`
	Source   string  `json:"source"`
	In       float64 `json:"in"`
	Out      float64 `json:"out"`
	OutStart float64 `json:"out_start"` // in the file (range renders start at 0)
	Frames   int64   `json:"frames"`
	Zoom     float64 `json:"zoom,omitempty"`
	AnchorX  float64 `json:"anchor_x,omitempty"`
	AnchorY  float64 `json:"anchor_y,omitempty"`
	FillX    float64 `json:"fill_x,omitempty"` // this segment's own crop centre (fill frames)
	Audio    string  `json:"audio,omitempty"`  // audio source this segment's sound came from (multi-take)
	// RangeIn/RangeOut: this edge is where a --range render starts/ends, not a cut in the edit
	RangeIn  bool `json:"range_in,omitempty"`
	RangeOut bool `json:"range_out,omitempty"`
}

// AudioOf is the audio source a segment's sound came from.
func (m *Manifest) AudioOf(g MSeg) string {
	if g.Audio != "" {
		return g.Audio
	}
	return m.Audio
}

type MSource struct {
	Path       string      `json:"path"`
	Size       int64       `json:"size"`
	ModTime    time.Time   `json:"mtime"`
	AudioStart float64     `json:"audio_start"`
	Sync       *align.Sync `json:"sync,omitempty"`
}

func ManifestPath(out string) string { return out + ".cutroom.json" }

// TimelineHash identifies the edit a render was made from.
func TimelineHash(tl *timeline.Timeline) string {
	var segs []any
	for _, g := range tl.Segments {
		segs = append(segs, []any{g.Source, g.In, g.Out, g.Zoom, g.AnchorX, g.AnchorY, g.FillAt})
	}
	parts := []any{tl.FPS, tl.Width, tl.Height, tl.Audio, tl.Grid, segs, tl.AudioFor}
	if len(tl.Overlays) > 0 { // unchanged for edits without overlays
		parts = append(parts, tl.Overlays)
	}
	if tl.Fill {
		parts = append(parts, []any{"fill", tl.FillX, tl.FillY})
	}
	if tl.Bars != "" && !tl.Fill {
		parts = append(parts, []any{"bars", tl.Bars})
	}
	return hash(parts)
}

// CutHash identifies the cut alone (what is said when): captions depend on it, not on overlays.
func CutHash(tl *timeline.Timeline) string {
	c := *tl
	c.Overlays = nil
	return TimelineHash(&c)
}

// writeManifest records chunks (planned on vtl) as rendered from the edit tl.
func writeManifest(tl, vtl *timeline.Timeline, srcs map[string]Source, chunks []Chunk, res *Result, o Options) error {
	m := Manifest{Version: Version, Output: res.Output, Project: o.ProjDir, Preview: o.Preview, FPS: tl.FPS, Width: res.Width, Height: res.Height,
		Audio: tl.Audio, Sources: map[string]MSource{}, Loudness: o.Loudness, TruePeak: o.TruePeak, Chain: o.Chain,
		Frames: res.Frames, Duration: res.Duration, RangeStart: res.RangeStart, Partial: o.Range[1] > 0, Overlays: res.Overlays,
		Fill: tl.Fill, FillX: tl.FillX, FillY: tl.FillY, Bars: tl.Bars,
		TimelineHash: TimelineHash(tl), RenderedAt: time.Now().UTC().Truncate(time.Second)}
	for _, c := range chunks {
		for _, k := range c.Segments {
			g := vtl.Segments[k]
			a, _, _ := AudioFor(tl, srcs, g.Source)
			ms := MSeg{ID: g.ID, Source: g.Source, In: g.In, Out: g.Out,
				OutStart: g.OutStart, Frames: tl.Frames(g), Zoom: g.Zoom, AnchorX: g.AnchorX, AnchorY: g.AnchorY, FillX: g.FillAt, Audio: a}
			if full := tl.Find(g.ID); full != nil && vtl != tl {
				ms.RangeIn, ms.RangeOut = g.In > full.In, g.Out < full.Out
			}
			m.Segments = append(m.Segments, ms)
			for _, id := range []string{g.Source, a} {
				if s, ok := srcs[id]; ok {
					m.Sources[id] = MSource{s.Path, s.Size, s.ModTime, s.AudioStart, s.Sync}
				}
			}
		}
	}
	if s, ok := srcs[tl.Audio]; ok {
		m.Sources[tl.Audio] = MSource{s.Path, s.Size, s.ModTime, s.AudioStart, s.Sync}
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	return os.WriteFile(ManifestPath(res.Output), b, 0o644)
}

func LoadManifest(out string) (*Manifest, error) {
	b, err := os.ReadFile(ManifestPath(out))
	if err != nil {
		return nil, err
	}
	m := &Manifest{}
	return m, json.Unmarshal(b, m)
}

// SegVF is the PictureVF for a manifest segment (what look/check must apply to source frames).
func (m *Manifest) SegVF(g MSeg) string {
	ax, ay := g.Anchor()
	fx := m.FillX
	if g.FillX > 0 {
		fx = g.FillX
	}
	return PictureVF(g.Zoom, ax, ay, m.Fill, fx, m.FillY, m.Bars, m.Width, m.Height)
}
