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

func TestLowerThirdKeepsClearOfStyledCaptions(t *testing.T) {
	W, H := 1080, 1920
	o := timeline.Overlay{Kind: "text", Layout: "lower-third-raised"}
	x0, y0 := place("lower-third", W, H, 500, 150, 60)
	// captions moved up to the middle: the lower third keeps its own place
	o.CapBox = [4]float64{0.2, 0.55, 0.8, 0.62}
	if x, y := placeCard(o, W, H, 500, 150, 60); x != x0 || y != y0 {
		t.Errorf("captions out of the way: lower third moved to %d,%d (its place %d,%d)", x, y, x0, y0)
	}
	// captions where the lower third would be: it goes above them
	o.CapBox = [4]float64{0.2, 0.70, 0.8, 0.77}
	_, y := placeCard(o, W, H, 500, 150, 60)
	if y+150 > int(0.70*float64(H)) {
		t.Errorf("lower third at y %d–%d overlaps captions from %d", y, y+150, int(0.70*float64(H)))
	}
}

func TestAnimFrameScalesAroundCentre(t *testing.T) {
	l, _ := resolve(CaptionDefaults, "", 1080, 1920)
	img, r := captionCard([]string{"hello"}, -1, l, 1080, 1920)
	f := animFrame(img, r, 0.5, 0, 0, 1)
	// the corner of the settled box is now empty; the centre still drawn
	if a := f.NRGBAAt(r.Min.X+2, r.Min.Y+2).A; a != 0 {
		t.Errorf("scaled frame still covers the box corner (alpha %d)", a)
	}
	if a := f.NRGBAAt((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2).A; a == 0 {
		t.Errorf("scaled frame lost its centre")
	}
	for _, k := range []string{"pop", "slide-up", "fade", "slide-left"} {
		if s, dx, dy, a := animAt(k, 1); s != 1 || dx != 0 || dy != 0 || a != 1 {
			t.Errorf("%s at progress 1 must be settled: %g %g %g %g", k, s, dx, dy, a)
		}
	}
}
