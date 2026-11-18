// Package raster draws text and simple shapes into images. ffmpeg builds often lack drawtext and
// libass (L47), so every label, caption and title cutroom shows is rasterised here.
package raster

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var (
	fontsOnce sync.Once
	fonts     map[string]*opentype.Font
	faceMu    sync.Mutex
	faces     = map[faceKey]font.Face{}
)

type faceKey struct {
	name string
	size float64
}

func loadFonts() {
	fonts = map[string]*opentype.Font{}
	for name, ttf := range map[string][]byte{"regular": goregular.TTF, "bold": gobold.TTF, "mono": gomono.TTF} {
		f, err := opentype.Parse(ttf)
		if err != nil {
			panic(err)
		}
		fonts[name] = f
	}
}

// Face returns a cached font face ("regular", "bold", "mono") at size px.
func Face(name string, size float64) font.Face {
	fontsOnce.Do(loadFonts)
	faceMu.Lock()
	defer faceMu.Unlock()
	k := faceKey{name, size}
	if f, ok := faces[k]; ok {
		return f
	}
	f, err := opentype.NewFace(fonts[name], &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		panic(err)
	}
	faces[k] = f
	return f
}

// Text draws s with its baseline at (x, y). Returns the advance width.
func Text(dst draw.Image, face font.Face, x, y int, c color.Color, s string) int {
	d := &font.Drawer{Dst: dst, Src: image.NewUniform(c), Face: face, Dot: fixed.P(x, y)}
	d.DrawString(s)
	return (d.Dot.X - fixed.I(x)).Round()
}

// Measure returns the advance width of s.
func Measure(face font.Face, s string) int {
	return font.MeasureString(face, s).Round()
}

// Label draws text on a filled box (for readable labels over video frames).
func Label(dst draw.Image, face font.Face, x, y int, fg, bg color.Color, s string) {
	m := face.Metrics()
	w := Measure(face, s)
	asc, desc := m.Ascent.Round(), m.Descent.Round()
	Fill(dst, image.Rect(x-3, y-asc-2, x+w+3, y+desc+2), bg)
	Text(dst, face, x, y, fg, s)
}

func Fill(dst draw.Image, r image.Rectangle, c color.Color) {
	draw.Draw(dst, r, image.NewUniform(c), image.Point{}, draw.Over)
}

func SavePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func LoadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

var (
	White = color.RGBA{255, 255, 255, 255}
	Black = color.RGBA{0, 0, 0, 255}
	Dim   = color.RGBA{0, 0, 0, 180}
	Grey  = color.RGBA{150, 150, 150, 255}
	Panel = color.RGBA{24, 24, 27, 255}
	Amber = color.RGBA{245, 158, 11, 255}
	Cyan  = color.RGBA{34, 211, 238, 255}
	Red   = color.RGBA{239, 68, 68, 255}
	Green = color.RGBA{34, 197, 94, 255}
)
