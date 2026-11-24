package render

import (
	"encoding/json"
	"os"
	"time"

	"github.com/cutroom/cutroom/internal/align"
	"github.com/cutroom/cutroom/internal/timeline"
)

// Manifest records exactly what a render contains, so `check` can verify the file independently
// and detect when the timeline has changed since.
type Manifest struct {
	Version      int                `json:"version"`
	Output       string             `json:"output"`
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
		segs = append(segs, []any{g.Source, g.In, g.Out, g.Zoom, g.AnchorX, g.AnchorY})
	}
	return hash([]any{tl.FPS, tl.Width, tl.Height, tl.Audio, tl.Grid, segs})
}

func writeManifest(tl *timeline.Timeline, srcs map[string]Source, chunks []Chunk, res *Result, o Options) error {
	m := Manifest{Version: Version, Output: res.Output, Preview: o.Preview, FPS: tl.FPS, Width: res.Width, Height: res.Height,
		Audio: tl.Audio, Sources: map[string]MSource{}, Loudness: o.Loudness, TruePeak: o.TruePeak, Chain: o.Chain,
		Frames: res.Frames, Duration: res.Duration, RangeStart: res.RangeStart, Partial: o.Range[1] > 0,
		TimelineHash: TimelineHash(tl), RenderedAt: time.Now().UTC().Truncate(time.Second)}
	for _, c := range chunks {
		for _, k := range c.Segments {
			g := tl.Segments[k]
			m.Segments = append(m.Segments, MSeg{ID: g.ID, Source: g.Source, In: g.In, Out: g.Out,
				OutStart: round6(g.OutStart - res.RangeStart), Frames: tl.Frames(g), Zoom: g.Zoom, AnchorX: g.AnchorX, AnchorY: g.AnchorY})
			s := srcs[g.Source]
			m.Sources[g.Source] = MSource{s.Path, s.Size, s.ModTime, s.AudioStart, s.Sync}
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
