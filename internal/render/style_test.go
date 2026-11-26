package render

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/NWylynko/cutroom/internal/timeline"
)

func countColor(img *image.NRGBA, r image.Rectangle, want color.NRGBA) int {
	n := 0
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if img.NRGBAAt(x, y) == want {
				n++
			}
		}
	}
	return n
}

func TestCaptionStyleHighlightAndBox(t *testing.T) {
	st := (&timeline.Style{Highlight: "#ff0000", Box: "none", Outline: "black"}).Over(CaptionDefaults)
	l, err := resolve(st, "", 1080, 1920)
	if err != nil {
		t.Fatal(err)
	}
	img, r := captionCard([]string{"one", "two"}, 1, l, 1080, 1920)
	red := countColor(img, r, color.NRGBA{255, 0, 0, 255})
	if red < 200 {
		t.Errorf("the active word must be drawn in the highlight colour: %d red px", red)
	}
	// no box: the corners of the box area stay transparent
	if a := img.NRGBAAt(r.Min.X+1, r.Min.Y+1).A; a != 0 {
		t.Errorf("box none drew a box (alpha %d at its corner)", a)
	}
	img2, _ := captionCard([]string{"one", "two"}, -1, l, 1080, 1920)
	if countColor(img2, r, color.NRGBA{255, 0, 0, 255}) != 0 {
		t.Errorf("no active word: nothing highlighted")
	}
}

func TestStylePlacesBoxCentre(t *testing.T) {
	x, y := 0.3, 0.2
	st := (&timeline.Style{X: &x, Y: &y}).Over(CaptionDefaults)
	l, _ := resolve(st, "", 1920, 1080)
	_, r := captionCard([]string{"hello"}, -1, l, 1920, 1080)
	cx, cy := (r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2
	if abs(cx-576) > 2 || abs(cy-216) > 2 {
		t.Errorf("box centre %d,%d, want 576,216", cx, cy)
	}
}

func TestCaptionWordsWrapToWidth(t *testing.T) {
	w := 0.3
	st := (&timeline.Style{Width: w}).Over(CaptionDefaults)
	l, _ := resolve(st, "", 1920, 1080)
	_, r1 := captionCard([]string{"a", "much", "longer", "caption", "line"}, -1, l, 1920, 1080)
	_, r0 := captionCard([]string{"a"}, -1, l, 1920, 1080)
	if r1.Dy()*2 < 3*r0.Dy() || r1.Dx() > int(w*1920)+2*l.padX {
		t.Errorf("a line wider than --width must wrap: %v vs one line %v", r1, r0)
	}
}

func TestCardStyleLayers(t *testing.T) {
	tl := &timeline.Timeline{CardStyle: &timeline.Style{Box: "#ffffff", Color: "#101718"}}
	o := timeline.Overlay{Kind: "text", Text: "Hi", Layout: "title", Style: &timeline.Style{Color: "#307351"}}
	st := CardStyle(tl, o)
	if st.Box != "#ffffff" || st.Color != "#307351" || st.Align != "center" {
		t.Errorf("card style layers: %+v", st)
	}
	o.Style = &st
	img, r, err := Card(o, "", 1920, 1080)
	if err != nil {
		t.Fatal(err)
	}
	if c := img.NRGBAAt(r.Min.X+2, r.Min.Y+2); c != (color.NRGBA{255, 255, 255, 255}) {
		t.Errorf("box colour %v", c)
	}
}

func TestUnknownFontRefused(t *testing.T) {
	err := CheckStyle(timeline.Style{Font: "No Such Font Anywhere"}, "")
	if err == nil || !strings.Contains(err.Error(), "style fonts") {
		t.Errorf("unknown font: %v", err)
	}
	if err := CheckStyle(timeline.Style{Font: "missing/file.ttf"}, ""); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("missing file: %v", err)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
