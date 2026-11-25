// Package perceive renders what an agent cannot watch or hear into images it can look at:
// frames, labelled contact sheets, and annotated spectrograms.
package perceive

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"strconv"
	"sync"

	"github.com/cutroom/cutroom/internal/deps"
	"github.com/cutroom/cutroom/internal/media"
	"github.com/cutroom/cutroom/internal/raster"
	"github.com/cutroom/cutroom/internal/speech"
	"github.com/cutroom/cutroom/internal/timecode"
	xdraw "golang.org/x/image/draw"
)

// MaxEdge is the default long edge for images meant for vision models.
const MaxEdge = 1568

// Frame decodes the frame shown at time t (source seconds), as displayed (rotation applied),
// scaled so its width is at most maxW (0 = native). crop (x,y,w,h in source pixels) is optional.
func Frame(src string, t float64, maxW int, crop *image.Rectangle) (image.Image, error) {
	return FrameVF(src, t, maxW, crop, "")
}

// FrameVF is Frame with a filter applied first (after any HDR conversion), e.g. a punch-in crop.
func FrameVF(src string, t float64, maxW int, crop *image.Rectangle, pre string) (image.Image, error) {
	vf := ""
	if pre != "" {
		vf = pre + ","
	}
	if crop != nil {
		vf = fmt.Sprintf("crop=%d:%d:%d:%d,", crop.Dx(), crop.Dy(), crop.Min.X, crop.Min.Y)
	}
	if maxW > 0 {
		vf += fmt.Sprintf("scale='min(%d,iw)':-2:flags=lanczos,", maxW)
	}
	vf += "format=rgb24"
	sdr := media.ReadSDR(src, 0, 0) // HDR sources shown as they render (L23)
	args := []string{"-hide_banner", "-loglevel", "error", "-ss", strconv.FormatFloat(t, 'f', 4, 64)}
	args = append(args, sdr.Pre...)
	args = append(args, "-i", src, "-frames:v", "1", "-vf", sdr.Chain(vf), "-f", "image2pipe", "-c:v", "png", "-")
	out, err := deps.Run("ffmpeg", args...)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no frame at %.3f s (past the end?)", t)
	}
	return png.Decode(bytes.NewReader(out))
}

// SheetFrame is one cell of a contact sheet.
type SheetFrame struct {
	Time   float64 `json:"time"`
	Frame  int64   `json:"frame"`
	Saying string  `json:"saying,omitempty"` // word(s) being spoken at that moment
}

// Cell is one frame on a sheet: which file, which time, how to label it.
type Cell struct {
	Path  string   `json:"path"`
	T     float64  `json:"time"`  // time in Path
	Label string   `json:"label"` // first line (white)
	Say   string   `json:"saying,omitempty"`
	Cut   bool     `json:"cut_before,omitempty"` // draw a cut marker on the left edge
	VF    string   `json:"-"`                    // filter applied to the frame first (e.g. a punch-in crop)
	Over  []string `json:"overlays,omitempty"`   // full-frame RGBA PNGs drawn over it (overlays)
}

// Sheet builds a labelled contact sheet of n frames evenly spaced over [a, b] of one file.
func Sheet(src string, a, b float64, n int, rate timecode.Rate, words []speech.Word) (image.Image, []SheetFrame, error) {
	if n < 1 {
		n = 1
	}
	var cells []Cell
	var meta []SheetFrame
	for i := 0; i < n; i++ {
		t := a
		if n > 1 {
			t = a + (b-a)*float64(i)/float64(n-1)
		}
		if i == n-1 && n > 1 {
			t = math.Max(a, b-0.5/rate.FPS()) // last frame inside the range
		}
		f := timecode.FromFloat(t).Frame(rate)
		say := sayingAt(words, t)
		cells = append(cells, Cell{Path: src, T: t, Label: fmt.Sprintf("%s f%d", timecode.FromFloat(t).Clock(), f), Say: say})
		meta = append(meta, SheetFrame{Time: math.Round(t*1000) / 1000, Frame: f, Saying: say})
	}
	img, err := Cells(cells, 0)
	return img, meta, err
}

// Cells lays out cells in a grid (cols 0 = automatic), each labelled, cut markers in red.
func Cells(cells []Cell, cols int) (image.Image, error) {
	n := len(cells)
	if n == 0 {
		return nil, fmt.Errorf("nothing to show")
	}
	if cols <= 0 {
		cols = 4
		switch {
		case n <= 2:
			cols = n
		case n <= 6:
			cols = 3
		case n > 16:
			cols = 6
		}
	}
	rows := (n + cols - 1) / cols
	thumbW := min(480, (MaxEdge-(cols+1)*8)/cols)
	imgs := make([]image.Image, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	sem := make(chan struct{}, 6)
	for i := range cells {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			imgs[i], errs[i] = FrameVF(cells[i].Path, cells[i].T, thumbW, nil, cells[i].VF)
		}(i)
	}
	wg.Wait()
	for i, e := range errs {
		if e != nil {
			return nil, fmt.Errorf("frame at %.3f of %s: %w", cells[i].T, cells[i].Path, e)
		}
	}
	th := imgs[0].Bounds().Dy() * thumbW / imgs[0].Bounds().Dx()
	labelH, pad := 40, 8
	W := pad + cols*(thumbW+pad)
	H := pad + rows*(th+labelH+pad)
	sheet := image.NewRGBA(image.Rect(0, 0, W, H))
	raster.Fill(sheet, sheet.Bounds(), raster.Panel)
	face := raster.Face("mono", 13)
	for i, im := range imgs {
		c := cells[i]
		x := pad + (i%cols)*(thumbW+pad)
		y := pad + (i/cols)*(th+labelH+pad)
		r := image.Rect(x, y, x+thumbW, y+th)
		xdraw.CatmullRom.Scale(sheet, r, im, im.Bounds(), draw.Src, nil)
		for _, p := range c.Over {
			if ov, err := raster.LoadPNG(p); err == nil {
				xdraw.CatmullRom.Scale(sheet, r, ov, ov.Bounds(), draw.Over, nil)
			}
		}
		if c.Cut {
			raster.Fill(sheet, image.Rect(x-pad+1, y, x-1, y+th), raster.Red)
		}
		raster.Text(sheet, face, x, y+th+16, raster.White, fit(c.Label, thumbW))
		if c.Say != "" {
			raster.Text(sheet, face, x, y+th+33, raster.Amber, fit("\""+c.Say+"\"", thumbW))
		}
	}
	return sheet, nil
}

// sayingAt returns the word spoken at t, or "" in a pause.
func sayingAt(words []speech.Word, t float64) string {
	for _, w := range words {
		if t >= w.Start && t < w.End {
			return w.Text
		}
	}
	return ""
}

// fit truncates s with an ellipsis so it renders within w pixels in the 13 px mono face.
func fit(s string, w int) string {
	f := raster.Face("mono", 13)
	if raster.Measure(f, s) <= w {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && raster.Measure(f, string(r)+"…") > w {
		r = r[:len(r)-1]
	}
	if len(r) == 0 {
		return ""
	}
	return string(r) + "…"
}

// Annotate draws a time label in the corner of a single frame.
func Annotate(img image.Image, label string) *image.RGBA {
	b := img.Bounds()
	out := image.NewRGBA(b)
	draw.Draw(out, b, img, b.Min, draw.Src)
	raster.Label(out, raster.Face("mono", 15), 8, b.Dy()-10, raster.White, color.RGBA{0, 0, 0, 170}, label)
	return out
}
