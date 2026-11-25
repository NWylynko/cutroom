package timeline

import (
	"fmt"
	"math"
)

// Overlay is a text card, image or clip shown over the edit. It is anchored to a moment of a source
// (the first word of a quoted phrase), not to an output time, so re-cutting never moves it off its
// words; its output window is derived (Window).
type Overlay struct {
	ID     string  `json:"id"`
	Kind   string  `json:"kind"`   // text | image
	Source string  `json:"source"` // anchor source
	At     float64 `json:"at"`     // anchor: source time where it appears
	Until  float64 `json:"until,omitempty"`
	// Until > 0: it disappears at this source time (the end of a quoted phrase); else after Duration
	Duration   float64 `json:"duration,omitempty"`    // output seconds (when Until is 0)
	AnchorText string  `json:"anchor_text,omitempty"` // the quoted words, for display
	Text       string  `json:"text,omitempty"`        // text: first line
	Sub        string  `json:"sub,omitempty"`         // text: smaller second line
	Path       string  `json:"path,omitempty"`        // image: file (relative to the project)
	Layout     string  `json:"layout"`                // lower-third | title | top | corner | center
	Size       float64 `json:"size,omitempty"`        // image: width as a fraction of the frame
	Fade       float64 `json:"fade"`                  // s, in and out
}

// Window is the overlay's output time span, on the frame grid. ok is false (with why) when its
// anchor moment is no longer in the edit.
func (t *Timeline) Window(o Overlay) (start, end float64, ok bool, why string) {
	t.Recompute()
	a, ok := t.ToOutput(o.Source, o.At+0.5/t.fpsNum)
	if !ok {
		// the anchor word's start may sit in a cut pad: accept the first kept moment of the phrase
		lim := o.At + 0.6
		if o.Until > 0 {
			lim = o.Until
		}
		for s := o.At; s < lim && !ok; s += 1 / t.fpsNum {
			a, ok = t.ToOutput(o.Source, s)
		}
		if !ok {
			return 0, 0, false, fmt.Sprintf("its anchor %q (%s %.2f) was cut from the edit", o.AnchorText, o.Source, o.At)
		}
	}
	a = math.Floor(a*t.fpsNum+1e-6) / t.fpsNum
	b := a + o.Duration
	if o.Until > 0 {
		if e, ok2 := t.ToOutput(o.Source, o.Until-0.5/t.fpsNum); ok2 && e > a {
			b = e + 0.5/t.fpsNum
		} else {
			b = a + math.Max(o.Duration, 1)
		}
	}
	b = math.Min(math.Round(b*t.fpsNum)/t.fpsNum, t.Duration())
	if b-a < 1/t.fpsNum {
		return 0, 0, false, "it would last less than a frame (its anchor is at the very end)"
	}
	return round6(a), round6(b), true, ""
}

// OverlayByID finds an overlay.
func (t *Timeline) OverlayByID(id string) int {
	for i, o := range t.Overlays {
		if o.ID == id {
			return i
		}
	}
	return -1
}

// AddOverlay stores o with a fresh id.
func (t *Timeline) AddOverlay(o Overlay) string {
	t.NextOverlay++
	o.ID = fmt.Sprintf("o%d", t.NextOverlay)
	t.Overlays = append(t.Overlays, o)
	return o.ID
}

// overlayIssues: overlays whose anchor was cut, and overlays in the same place at the same time.
func (t *Timeline) overlayIssues() []Issue {
	var is []Issue
	type win struct {
		o    Overlay
		a, b float64
	}
	var ws []win
	for _, o := range t.Overlays {
		a, b, ok, why := t.Window(o)
		if !ok {
			is = append(is, Issue{Level: "warn", Code: "overlay-orphaned", Message: fmt.Sprintf("overlay %s (%s %q) won't show: %s", o.ID, o.Kind, ovName(o), why),
				Fix: fmt.Sprintf("anchor it to words the edit keeps: `cutroom overlay rm %s` then `cutroom overlay add …`", o.ID)})
			continue
		}
		for _, w := range ws {
			if w.o.Layout == o.Layout && a < w.b && b > w.a {
				is = append(is, Issue{Level: "warn", Code: "overlay-overlap", At: round6(math.Max(a, w.a)),
					Message: fmt.Sprintf("overlays %s and %s are both %s at output %.2f–%.2f s: they draw over each other", w.o.ID, o.ID, o.Layout, math.Max(a, w.a), math.Min(b, w.b)),
					Fix:     fmt.Sprintf("shorten one (--duration), or give one another --layout (`cutroom overlay list`)")})
			}
		}
		ws = append(ws, win{o, a, b})
	}
	return is
}

func ovName(o Overlay) string {
	if o.Kind == "text" {
		return o.Text
	}
	return o.Path
}
