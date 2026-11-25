package render

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/cutroom/cutroom/internal/timeline"
)

func writePNG(t *testing.T, p string, c color.Color) {
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			img.Set(x, y, c)
		}
	}
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	png.Encode(f, img)
	f.Close()
}

// L19: card caches are keyed on the bytes of every referenced file, not just the spec.
func TestL19_CardKeyFollowsImageBytes(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "logo.png")
	writePNG(t, p, color.RGBA{255, 0, 0, 255})
	o := timeline.Overlay{ID: "o1", Kind: "image", Path: "logo.png", Layout: "corner", Size: 0.2}
	k1, err := cardKey(o, dir, 1280, 720)
	if err != nil {
		t.Fatal(err)
	}
	writePNG(t, p, color.RGBA{0, 0, 255, 255}) // same path, same size, new pixels
	k2, _ := cardKey(o, dir, 1280, 720)
	if k1 == k2 {
		t.Fatal("L19: changed image kept the same cache key")
	}
	if k3, _ := cardKey(o, dir, 960, 540); k3 == k2 {
		t.Fatal("output size must be part of the key")
	}
}

func TestCardsStayInFrame(t *testing.T) {
	dir := t.TempDir()
	writePNG(t, filepath.Join(dir, "logo.png"), color.White)
	long := "A rather long lower third title that has to wrap onto more than one line to fit"
	for _, l := range Layouts {
		for _, o := range []timeline.Overlay{
			{ID: "t", Kind: "text", Text: long, Sub: "and a subtitle", Layout: l},
			{ID: "i", Kind: "image", Path: "logo.png", Layout: l, Size: 0.3},
		} {
			img, r, err := Card(o, dir, 1280, 720)
			if err != nil {
				t.Fatal(err)
			}
			if !r.In(img.Bounds()) || r.Empty() {
				t.Errorf("%s %s: box %v outside the frame", o.Kind, l, r)
			}
		}
	}
}

func TestClipBoxes(t *testing.T) {
	o := timeline.Overlay{Kind: "clip", ClipW: 1920, ClipH: 1080, Layout: "full"}
	if r := clipBox(o, 1280, 720); r != image.Rect(0, 0, 1280, 720) {
		t.Fatalf("full: %v", r)
	}
	for _, l := range Layouts {
		for _, sz := range []float64{0, 0.2, 0.9} {
			o.Layout, o.Size = l, sz
			r := clipBox(o, 1280, 720)
			if !r.In(image.Rect(0, 0, 1280, 720)) || r.Dx()%2 != 0 || r.Dy()%2 != 0 || r.Empty() {
				t.Errorf("%s size %.1f: box %v (must be inside, even)", l, sz, r)
			}
		}
	}
	o.ClipW, o.ClipH, o.Layout, o.Size = 1080, 1920, "corner", 0.5 // a vertical clip is capped by height
	if r := clipBox(o, 1280, 720); r.Dy() > 720 {
		t.Errorf("vertical clip overflows: %v", r)
	}
}
