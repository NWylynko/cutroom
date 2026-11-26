package render

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/NWylynko/cutroom/internal/raster"
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

func TestAccentFollowsRoundedCorners(t *testing.T) {
	st := (&timeline.Style{Radius: 0.6, Accent: "#ff0000", Box: "#ffffff"}).Over(CardDefaults("lower-third"))
	o := timeline.Overlay{Kind: "text", Text: "Hi", Layout: "lower-third", Style: &st}
	img, r, err := Card(o, "", 1920, 1080)
	if err != nil {
		t.Fatal(err)
	}
	if a := img.NRGBAAt(r.Min.X, r.Min.Y).A; a != 0 {
		t.Errorf("the accent bar pokes out of the rounded corner (alpha %d)", a)
	}
	if c := img.NRGBAAt(r.Min.X+2, (r.Min.Y+r.Max.Y)/2); c != (color.NRGBA{255, 0, 0, 255}) {
		t.Errorf("accent bar missing at mid-height: %v", c)
	}
}

func TestCropSpanAndHolds(t *testing.T) {
	lo, hi := CropSpan(0.5, 1920, 1080, 1080, 1920)
	if lo < 0.34 || lo > 0.35 || hi < 0.65 || hi > 0.66 {
		t.Errorf("centre crop of 16:9 → 9:16 keeps %.3f–%.3f, want ≈0.342–0.658", lo, hi)
	}
	if lo, hi := CropSpan(0.95, 1920, 1080, 1080, 1920); hi != 1 || lo < 0.68 {
		t.Errorf("a crop near the edge stays inside the picture: %.3f–%.3f", lo, hi)
	}
	if CropHolds(0.5, 0.76, 1920, 1080, 1080, 1920) || !CropHolds(0.76, 0.76, 1920, 1080, 1080, 1920) {
		t.Errorf("a speaker at 0.76: outside a centred crop, inside one centred on them")
	}
}

func TestL57_GPUScaleKeepsTheSourceShape(t *testing.T) {
	// 4K landscape into a vertical 1080x1920 fill: scale to cover (3413x1920), never squash to 1080x1920
	if w, h := gpuSize(true, 3840, 2160, 1080, 1920, 1); w != 3414 && w != 3412 || h != 1920 {
		t.Errorf("fill: %dx%d, want ≈3413x1920", w, h)
	}
	if w, h := gpuSize(false, 3840, 2160, 1920, 1080, 1.2); w != 2304 || h != 1296 {
		t.Errorf("fit with zoom 1.2: %dx%d, want 2304x1296", w, h)
	}
	if w, h := gpuSize(true, 0, 0, 1080, 1920, 1); w != 0 || h != 0 {
		t.Errorf("unknown source size on a fill: no GPU scale, got %dx%d", w, h)
	}
}

func TestL58_CaptionCardsReadAtTheTimelineRate(t *testing.T) {
	l := capList([]CapRun{{"/a.png", 0, 7}, {"", 7, 20}}, "/blank.png", 60000.0/1001, "60000/1001")
	if strings.Count(l, "option framerate 60000/1001") != 3 {
		t.Errorf("every file (and the repeated last one) must be read at the timeline rate:\n%s", l)
	}
}

func TestUpperKeepNames(t *testing.T) {
	l := look{upper: true, keepNames: true}
	if got := l.text("using iCloud's OAuth on GitHub with Google API"); got != "USING iCloud's OAuth ON GitHub WITH GOOGLE API" {
		t.Errorf("got %q", got)
	}
}

func TestCaptionHighlightBox(t *testing.T) {
	st := (&timeline.Style{HighlightBox: "#00ff00", Highlight: "#ff0000", Box: "none", Outline: "black"}).Over(CaptionDefaults)
	l, _ := resolve(st, "", 1080, 1920)
	img, r := captionCard([]string{"one", "two"}, 1, l, 1080, 1920)
	if n := countColor(img, r, color.NRGBA{0, 255, 0, 255}); n < 500 {
		t.Errorf("highlight box: %d green px", n)
	}
	if n := countColor(img, r, color.NRGBA{255, 0, 0, 255}); n < 100 {
		t.Errorf("the boxed word's letters are drawn on top of the box: %d red px", n)
	}
}

func TestTrackingWidensText(t *testing.T) {
	l0, _ := resolve(CaptionDefaults, "", 1080, 1920)
	st := (&timeline.Style{Tracking: 0.1}).Over(CaptionDefaults)
	l1, _ := resolve(st, "", 1080, 1920)
	a, b := raster.Measure(l0.face, "HELLO"), raster.Measure(l1.face, "HELLO")
	want := a + int(5*0.1*l1.size) - 2
	if b < want {
		t.Errorf("tracking 0.1 on 5 letters: width %d, want ≥ %d (untracked %d)", b, want, a)
	}
}
