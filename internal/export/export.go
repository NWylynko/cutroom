// Package export writes the edit for other editors: OpenTimelineIO (JSON, Timeline.1) and CMX 3600
// EDL. Times are whole frames of the timeline's rate, so a re-import lands on the same frames.
package export

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/NWylynko/cutroom/internal/align"
	"github.com/NWylynko/cutroom/internal/timeline"
)

// Source is what export needs to know about a source.
type Source struct {
	Path string      // absolute
	Sync *align.Sync // audio aligned to another source (a recorder)
}

// Card is an overlay as exported: a picture file shown over [Start, End) of the output.
type Card struct {
	Name       string
	Path       string // absolute: a card PNG, an image, or a clip
	Start, End float64
	ClipIn     float64 // clips: where in the clip it starts
	Track      string  // "overlays" or "captions"
}

// Event is one piece of the edit as both formats describe it.
type Event struct {
	Seg           string
	Source        string // video source id
	Audio         string // audio source id
	SrcIn, Frames int64  // video: source start frame, length
	AudIn         int64  // audio: start frame in the audio source's own clock
	RecIn         int64  // output start frame
}

// Events lists the edit's pieces in output order, in frames at the timeline's rate.
func Events(tl *timeline.Timeline, audioFor func(string) string, srcs map[string]Source) []Event {
	tl.Recompute()
	r := tl.Rate()
	var es []Event
	for _, g := range tl.Segments {
		a := audioFor(g.Source)
		e := Event{Seg: g.ID, Source: g.Source, Audio: a, SrcIn: int64(math.Round(g.In * r)), Frames: tl.Frames(g), RecIn: int64(math.Round(g.OutStart * r))}
		e.AudIn = e.SrcIn
		if a != g.Source {
			if s := srcs[a]; s.Sync != nil {
				e.AudIn = int64(math.Round(s.Sync.Map(g.In) * r))
			}
		}
		es = append(es, e)
	}
	return es
}

// --- OpenTimelineIO --------------------------------------------------------------------------

type obj = map[string]any

func rt(v int64, rate float64) obj {
	return obj{"OTIO_SCHEMA": "RationalTime.1", "rate": rate, "value": float64(v)}
}

func trange(start, dur int64, rate float64) obj {
	return obj{"OTIO_SCHEMA": "TimeRange.1", "start_time": rt(start, rate), "duration": rt(dur, rate)}
}

func fileURL(p string) string { return (&url.URL{Scheme: "file", Path: filepath.ToSlash(p)}).String() }

func clip(name, path string, start, dur int64, rate float64, meta obj) obj {
	if meta == nil {
		meta = obj{}
	}
	return obj{"OTIO_SCHEMA": "Clip.2", "name": name, "metadata": meta, "enabled": true,
		"source_range": trange(start, dur, rate), "effects": []any{}, "markers": []any{},
		"media_references": obj{"DEFAULT_MEDIA": obj{"OTIO_SCHEMA": "ExternalReference.1", "name": filepath.Base(path),
			"target_url": fileURL(path), "available_range": nil, "available_image_bounds": nil, "metadata": obj{}}},
		"active_media_reference_key": "DEFAULT_MEDIA"}
}

func gap(dur int64, rate float64) obj {
	return obj{"OTIO_SCHEMA": "Gap.1", "name": "", "metadata": obj{}, "enabled": true,
		"source_range": trange(0, dur, rate), "effects": []any{}, "markers": []any{}}
}

func track(name, kind string, children []any) obj {
	return obj{"OTIO_SCHEMA": "Track.1", "name": name, "kind": kind, "metadata": obj{}, "enabled": true,
		"source_range": nil, "effects": []any{}, "markers": []any{}, "children": children}
}

// OTIO is the edit as an OpenTimelineIO timeline: V1 the picture, V2/V3 overlays and burned
// captions (as the files they are drawn from), A1 the sound (each piece from the source it is taken
// from, recorders at their own clock).
func OTIO(name string, tl *timeline.Timeline, es []Event, srcs map[string]Source, cards []Card) ([]byte, error) {
	r := tl.Rate()
	var v1, a1 []any
	for _, e := range es {
		v1 = append(v1, clip(e.Seg, srcs[e.Source].Path, e.SrcIn, e.Frames, r, obj{"cutroom": obj{"segment": e.Seg, "source": e.Source}}))
		a1 = append(a1, clip(e.Seg, srcs[e.Audio].Path, e.AudIn, e.Frames, r, obj{"cutroom": obj{"segment": e.Seg, "source": e.Audio}}))
	}
	tracks := []any{track("V1", "Video", v1)}
	cards = append([]Card(nil), cards...)
	sort.SliceStable(cards, func(i, j int) bool { return cards[i].Start < cards[j].Start }) // tracks run in time order
	for _, tn := range []string{"overlays", "captions"} {
		var ch []any
		at := int64(0)
		for _, c := range cards {
			if c.Track != tn {
				continue
			}
			a, b := int64(math.Round(c.Start*r)), int64(math.Round(c.End*r))
			if a < at || b <= a {
				continue // overlapping cards can't share a track; the first wins (render draws both)
			}
			if a > at {
				ch = append(ch, gap(a-at, r))
			}
			ch = append(ch, clip(c.Name, c.Path, int64(math.Round(c.ClipIn*r)), b-a, r, obj{"cutroom": obj{"overlay": c.Name}}))
			at = b
		}
		if len(ch) > 0 {
			label := map[string]string{"overlays": "V2 overlays", "captions": "V3 captions"}[tn]
			tracks = append(tracks, track(label, "Video", ch))
		}
	}
	tracks = append(tracks, track("A1", "Audio", a1))
	t := obj{"OTIO_SCHEMA": "Timeline.1", "name": name, "metadata": obj{"cutroom": obj{"fps": tl.FPS, "width": tl.Width, "height": tl.Height}},
		"global_start_time": nil,
		"tracks": obj{"OTIO_SCHEMA": "Stack.1", "name": "tracks", "metadata": obj{}, "enabled": true, "source_range": nil,
			"effects": []any{}, "markers": []any{}, "children": tracks}}
	return json.MarshalIndent(t, "", "  ")
}

// --- CMX 3600 EDL ------------------------------------------------------------------------------

var reelChars = regexp.MustCompile(`[^A-Z0-9]`)

// Reel is a source id as an EDL reel name (≤ 8 characters, A–Z 0–9).
func Reel(id string) string {
	r := reelChars.ReplaceAllString(strings.ToUpper(id), "")
	if r == "" {
		r = "AX"
	}
	if len(r) > 8 {
		r = r[:8]
	}
	return r
}

// Timecode formats a frame count as HH:MM:SS:FF at an integer frame base (non-drop).
func Timecode(f int64, base int) string {
	b := int64(base)
	return fmt.Sprintf("%02d:%02d:%02d:%02d", f/(3600*b), f/(60*b)%60, f/b%60, f%b)
}

// EDL is the edit as a CMX 3600 list: one event per segment — "B" (picture and sound) when the
// sound is the source's own, else a V event and an A event from the audio source at its own clock.
// Record time starts at 01:00:00:00. Overlays aren't representable and are listed as comments.
func EDL(title string, tl *timeline.Timeline, es []Event, srcs map[string]Source, cards []Card) string {
	base := int(math.Round(tl.Rate()))
	var b strings.Builder
	fmt.Fprintf(&b, "TITLE: %s\nFCM: NON-DROP FRAME\n\n", title)
	rec0 := int64(3600 * base)
	n := 0
	line := func(reel, ch string, src, dur, rec int64) {
		n++
		fmt.Fprintf(&b, "%03d  %-8s %-5s C        %s %s %s %s\n", n, reel, ch,
			Timecode(src, base), Timecode(src+dur, base), Timecode(rec0+rec, base), Timecode(rec0+rec+dur, base))
	}
	for _, e := range es {
		if e.Audio == e.Source {
			line(Reel(e.Source), "B", e.SrcIn, e.Frames, e.RecIn)
			fmt.Fprintf(&b, "* FROM CLIP NAME: %s\n* CUTROOM SEGMENT: %s\n\n", filepath.Base(srcs[e.Source].Path), e.Seg)
			continue
		}
		line(Reel(e.Source), "V", e.SrcIn, e.Frames, e.RecIn)
		fmt.Fprintf(&b, "* FROM CLIP NAME: %s\n* CUTROOM SEGMENT: %s\n", filepath.Base(srcs[e.Source].Path), e.Seg)
		line(Reel(e.Audio), "A", e.AudIn, e.Frames, e.RecIn)
		fmt.Fprintf(&b, "* FROM CLIP NAME: %s\n* CUTROOM SEGMENT: %s (sound)\n\n", filepath.Base(srcs[e.Audio].Path), e.Seg)
	}
	if len(cards) > 0 {
		b.WriteString("* OVERLAYS (not representable in CMX 3600; see the .otio export):\n")
		for _, c := range cards {
			if c.Track == "captions" {
				continue
			}
			fmt.Fprintf(&b, "*   %s  %s–%s  %s\n", c.Name, Timecode(rec0+int64(math.Round(c.Start*tl.Rate())), base), Timecode(rec0+int64(math.Round(c.End*tl.Rate())), base), filepath.Base(c.Path))
		}
	}
	return b.String()
}
