package raster

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

// FillRound fills r with c, corners rounded to radius px (anti-aliased; 0 = square).
func FillRound(dst draw.Image, r image.Rectangle, radius float64, c color.Color) {
	if radius <= 0.5 {
		Fill(dst, r, c)
		return
	}
	w, h := float32(r.Dx()), float32(r.Dy())
	rad := float32(math.Min(radius, float64(min(r.Dx(), r.Dy()))/2))
	z := vector.NewRasterizer(r.Dx(), r.Dy())
	const k = 0.5523 // cubic approximation of a quarter circle
	z.MoveTo(rad, 0)
	z.LineTo(w-rad, 0)
	z.CubeTo(w-rad+k*rad, 0, w, rad-k*rad, w, rad)
	z.LineTo(w, h-rad)
	z.CubeTo(w, h-rad+k*rad, w-rad+k*rad, h, w-rad, h)
	z.LineTo(rad, h)
	z.CubeTo(rad-k*rad, h, 0, h-rad+k*rad, 0, h-rad)
	z.LineTo(0, rad)
	z.CubeTo(0, rad-k*rad, rad-k*rad, 0, rad, 0)
	z.ClosePath()
	mask := image.NewAlpha(image.Rect(0, 0, r.Dx(), r.Dy()))
	z.Draw(mask, mask.Bounds(), image.Opaque, image.Point{})
	draw.DrawMask(dst, r, image.NewUniform(c), image.Point{}, mask, image.Point{}, draw.Over)
}

// Run is a piece of a line drawn in one colour.
type Run struct {
	S string
	C color.Color
}

// TextStyle is how StyledLine draws: an outline (Outline, OutlineW px) and a drop shadow (Shadow,
// offset ShadowOff px) under the letters. Nil colours draw nothing.
type TextStyle struct {
	Outline   color.Color
	OutlineW  int
	Shadow    color.Color
	ShadowOff int
}

// StyledLine draws runs one after another with the baseline at (x, y): shadow, then outline, then
// each run's letters in its colour. Returns the width.
func StyledLine(dst draw.Image, face font.Face, x, y int, runs []Run, st TextStyle) int {
	s := ""
	for _, r := range runs {
		s += r.S
	}
	w := Measure(face, s)
	m := face.Metrics()
	asc, desc := m.Ascent.Ceil(), m.Descent.Ceil()
	ow := max(0, st.OutlineW)
	if st.Outline == nil {
		ow = 0
	}
	if ow > 0 || st.Shadow != nil {
		// the line's letters as a mask, grown by the outline width
		pad := ow + 2
		mask := image.NewAlpha(image.Rect(0, 0, w+2*pad, asc+desc+2*pad))
		d := &font.Drawer{Dst: mask, Src: image.Opaque, Face: face, Dot: fixed.P(pad, pad+asc)}
		d.DrawString(s)
		grown := mask
		if ow > 0 {
			grown = dilate(mask, ow)
		}
		at := image.Pt(x-pad, y-asc-pad)
		if st.Shadow != nil {
			off := image.Pt(st.ShadowOff, st.ShadowOff)
			draw.DrawMask(dst, grown.Bounds().Add(at).Add(off), image.NewUniform(st.Shadow), image.Point{}, grown, image.Point{}, draw.Over)
		}
		if ow > 0 {
			draw.DrawMask(dst, grown.Bounds().Add(at), image.NewUniform(st.Outline), image.Point{}, grown, image.Point{}, draw.Over)
		}
	}
	cx := x
	for _, r := range runs {
		cx += Text(dst, face, cx, y, r.C, r.S)
	}
	return w
}

// dilate grows a mask by a disc of radius r (max filter), keeping anti-aliased edges.
func dilate(m *image.Alpha, r int) *image.Alpha {
	b := m.Bounds()
	out := image.NewAlpha(b)
	var offs []image.Point
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r+r {
				offs = append(offs, image.Pt(dx, dy))
			}
		}
	}
	W, H := b.Dx(), b.Dy()
	for y := range H {
		for x := range W {
			a := m.Pix[y*m.Stride+x]
			if a == 0 {
				continue
			}
			for _, o := range offs {
				xx, yy := x+o.X, y+o.Y
				if xx < 0 || yy < 0 || xx >= W || yy >= H {
					continue
				}
				if p := &out.Pix[yy*out.Stride+xx]; *p < a {
					*p = a
				}
			}
		}
	}
	return out
}

// SoftShadow draws a blurred rounded-rect shadow of r (radius px corners), moved down by off px and
// blurred by blur px, in c.
func SoftShadow(dst draw.Image, r image.Rectangle, radius float64, off, blur int, c color.Color) {
	pad := 2 * blur
	m := image.NewAlpha(image.Rect(0, 0, r.Dx()+2*pad, r.Dy()+2*pad))
	FillRound(m, image.Rect(pad, pad, pad+r.Dx(), pad+r.Dy()), radius, image.Opaque)
	for range 3 { // three box blurs ≈ a gaussian
		m = boxBlur(m, max(1, blur/2))
	}
	at := r.Min.Sub(image.Pt(pad, pad-off))
	draw.DrawMask(dst, m.Bounds().Add(at), image.NewUniform(c), image.Point{}, m, image.Point{}, draw.Over)
}

func boxBlur(m *image.Alpha, r int) *image.Alpha {
	W, H := m.Bounds().Dx(), m.Bounds().Dy()
	tmp := image.NewAlpha(m.Bounds())
	out := image.NewAlpha(m.Bounds())
	n := 2*r + 1
	for y := range H { // horizontal
		sum := 0
		for x := -r; x <= r; x++ {
			if x >= 0 && x < W {
				sum += int(m.Pix[y*m.Stride+x])
			}
		}
		for x := range W {
			tmp.Pix[y*tmp.Stride+x] = uint8(sum / n)
			if x-r >= 0 {
				sum -= int(m.Pix[y*m.Stride+x-r])
			}
			if x+r+1 < W {
				sum += int(m.Pix[y*m.Stride+x+r+1])
			}
		}
	}
	for x := range W { // vertical
		sum := 0
		for y := -r; y <= r; y++ {
			if y >= 0 && y < H {
				sum += int(tmp.Pix[y*tmp.Stride+x])
			}
		}
		for y := range H {
			out.Pix[y*out.Stride+x] = uint8(sum / n)
			if y-r >= 0 {
				sum -= int(tmp.Pix[(y-r)*tmp.Stride+x])
			}
			if y+r+1 < H {
				sum += int(tmp.Pix[(y+r+1)*tmp.Stride+x])
			}
		}
	}
	return out
}
