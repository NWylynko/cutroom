package render

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"

	"github.com/cutroom/cutroom/internal/raster"
	"github.com/cutroom/cutroom/internal/timeline"
)

// CardVersion changes whenever card drawing changes, so cached cards are redrawn.
const CardVersion = 1

// Placement is one overlay as rendered: a full-frame RGBA PNG shown over output [Start, End),
// fading in and out over Fade seconds. Box is the drawn area as fractions of the frame
// (x0, y0, x1, y1), for checks.
type Placement struct {
	ID    string     `json:"id"`
	Kind  string     `json:"kind"`
	PNG   string     `json:"png"`
	Key   string     `json:"key"`
	Start float64    `json:"start"`
	End   float64    `json:"end"`
	Fade  float64    `json:"fade"`
	Box   [4]float64 `json:"box"`
	Name  string     `json:"name"`
}

// Placements resolves every overlay of tl that shows to a card PNG at w×h (cached in cacheDir).
// Overlays whose anchor was cut are skipped (the timeline warns about them).
func Placements(tl *timeline.Timeline, projDir, cacheDir string, w, h int) ([]Placement, error) {
	var ps []Placement
	for _, o := range tl.Overlays {
		a, b, ok, _ := tl.Window(o)
		if !ok {
			continue
		}
		key, err := cardKey(o, projDir, w, h)
		if err != nil {
			return nil, err
		}
		png := filepath.Join(cacheDir, "overlays", key+".png")
		box, err := cardBox(o, projDir, w, h)
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(png); err != nil {
			img, _, err := Card(o, projDir, w, h)
			if err != nil {
				return nil, err
			}
			if err := ensureDir(png); err != nil {
				return nil, err
			}
			if err := raster.SavePNG(png, img); err != nil {
				return nil, err
			}
		}
		name := o.Text
		if o.Kind == "image" {
			name = o.Path
		}
		ps = append(ps, Placement{ID: o.ID, Kind: o.Kind, PNG: png, Key: key, Start: a, End: b, Fade: o.Fade, Box: box, Name: name})
	}
	return ps, nil
}

// cardKey hashes everything a card depends on, including the bytes of a referenced image (L19).
func cardKey(o timeline.Overlay, projDir string, w, h int) (string, error) {
	parts := []any{CardVersion, w, h, o.Kind, o.Text, o.Sub, o.Layout, o.Size, o.Path}
	if o.Kind == "image" {
		b, err := os.ReadFile(absIn(projDir, o.Path))
		if err != nil {
			return "", fmt.Errorf("overlay %s: %w", o.ID, err)
		}
		parts = append(parts, fmt.Sprintf("%x", sha256.Sum256(b)))
	}
	return hash(parts), nil
}

func absIn(dir, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(dir, p)
}

func cardBox(o timeline.Overlay, projDir string, w, h int) ([4]float64, error) {
	_, r, err := Card(o, projDir, w, h)
	if err != nil {
		return [4]float64{}, err
	}
	return [4]float64{float64(r.Min.X) / float64(w), float64(r.Min.Y) / float64(h), float64(r.Max.X) / float64(w), float64(r.Max.Y) / float64(h)}, nil
}

var (
	cardBG     = color.NRGBA{12, 12, 16, 190}
	cardAccent = color.NRGBA{245, 158, 11, 255}
)

// Card draws an overlay into a transparent w×h image; r is the area it covers.
func Card(o timeline.Overlay, projDir string, w, h int) (img *image.NRGBA, r image.Rectangle, err error) {
	img = image.NewNRGBA(image.Rect(0, 0, w, h))
	H := float64(h)
	margin := int(0.06 * float64(w))
	switch o.Kind {
	case "image":
		f, err := os.Open(absIn(projDir, o.Path))
		if err != nil {
			return nil, r, fmt.Errorf("overlay %s: %w", o.ID, err)
		}
		defer f.Close()
		src, _, err := image.Decode(f)
		if err != nil {
			return nil, r, fmt.Errorf("overlay %s: %s is not a PNG or JPEG: %w", o.ID, o.Path, err)
		}
		size := o.Size
		if size <= 0 {
			size = 0.25
		}
		iw := int(size * float64(w))
		ih := iw * src.Bounds().Dy() / max(1, src.Bounds().Dx())
		if ih > h-2*margin {
			ih = h - 2*margin
			iw = ih * src.Bounds().Dx() / max(1, src.Bounds().Dy())
		}
		x, y := place(o.Layout, w, h, iw, ih, margin)
		r = image.Rect(x, y, x+iw, y+ih)
		xdraw.CatmullRom.Scale(img, r, src, src.Bounds(), draw.Over, nil)
		return img, r, nil
	case "text":
		big, small := 0.055, 0.036
		maxW := 0.62
		switch o.Layout {
		case "title", "center":
			big, small, maxW = 0.085, 0.045, 0.84
		case "top":
			big, small, maxW = 0.05, 0.034, 0.84
		case "corner":
			big, small, maxW = 0.04, 0.028, 0.4
		}
		bf := raster.Face("bold", math.Round(big*H))
		sf := raster.Face("regular", math.Round(small*H))
		mw := int(maxW * float64(w))
		pad := int(0.022 * H)
		var lines []line
		for _, l := range wrapText(bf, o.Text, mw) {
			lines = append(lines, line{l, bf, color.NRGBA{255, 255, 255, 255}})
		}
		if o.Sub != "" {
			for _, l := range wrapText(sf, o.Sub, mw) {
				lines = append(lines, line{l, sf, color.NRGBA{220, 220, 225, 255}})
			}
		}
		tw, th := 0, 0
		for _, l := range lines {
			tw = max(tw, raster.Measure(l.face, l.s))
			th += lineH(l.face)
		}
		accent := 0
		if o.Layout == "lower-third" || o.Layout == "" {
			accent = int(math.Max(4, 0.006*float64(w)))
		}
		bw, bh := tw+2*pad+accent, th+2*pad
		x, y := place(o.Layout, w, h, bw, bh, margin)
		r = image.Rect(x, y, x+bw, y+bh)
		draw.Draw(img, r, image.NewUniform(cardBG), image.Point{}, draw.Over)
		if accent > 0 {
			draw.Draw(img, image.Rect(x, y, x+accent, y+bh), image.NewUniform(cardAccent), image.Point{}, draw.Over)
		}
		cy := y + pad
		for _, l := range lines {
			lx := x + accent + pad
			if o.Layout == "title" || o.Layout == "center" || o.Layout == "top" {
				lx = x + (bw-raster.Measure(l.face, l.s))/2
			}
			raster.Text(img, l.face, lx, cy+l.face.Metrics().Ascent.Round(), l.c, l.s)
			cy += lineH(l.face)
		}
		return img, r, nil
	}
	return nil, r, fmt.Errorf("overlay %s: unknown kind %q", o.ID, o.Kind)
}

type line struct {
	s    string
	face font.Face
	c    color.Color
}

func lineH(f font.Face) int {
	m := f.Metrics()
	return int(float64((m.Ascent + m.Descent).Round()) * 1.12)
}

// place puts a w×h box in the frame for a layout.
func place(layout string, W, H, w, h, margin int) (int, int) {
	switch layout {
	case "title", "center":
		return (W - w) / 2, (H - h) / 2
	case "top":
		return (W - w) / 2, int(0.08 * float64(H))
	case "corner":
		return W - margin - w, int(0.06 * float64(H))
	default: // lower-third
		return margin, H - int(0.12*float64(H)) - h
	}
}

// wrapText breaks s into lines no wider than mw (explicit "\n" kept).
func wrapText(f font.Face, s string, mw int) []string {
	var out []string
	for _, para := range strings.Split(s, "\n") {
		cur := ""
		for _, wd := range strings.Fields(para) {
			try := strings.TrimSpace(cur + " " + wd)
			if cur != "" && raster.Measure(f, try) > mw {
				out = append(out, cur)
				cur = wd
			} else {
				cur = try
			}
		}
		if cur != "" {
			out = append(out, cur)
		}
	}
	return out
}

// Layouts are the overlay positions.
var Layouts = []string{"lower-third", "title", "top", "corner", "center"}
