package raster

import (
	"encoding/xml"
	"fmt"
	"image"
	"image/draw"
	"io"
	"math"
	"strconv"
	"strings"

	"golang.org/x/image/vector"
)

// An SVG icon reduced to what icon sets like Lucide use: stroked shapes (path, circle, ellipse,
// line, rect, polyline, polygon) on a viewBox, with round caps and joins. Fills are ignored —
// stroke icons have none — and so are transforms, gradients and text.
type svgIcon struct {
	vx, vy, vw, vh float64
	width          float64
	lines          [][]pt // flattened strokes, each a polyline (closed ones repeat their start)
}

type pt struct{ x, y float64 }

// SVGMask renders an SVG icon's strokes as an alpha mask size×size (the viewBox scaled to fit).
func SVGMask(data []byte, size int) (*image.Alpha, error) {
	ic, err := parseSVG(data)
	if err != nil {
		return nil, err
	}
	if len(ic.lines) == 0 {
		return nil, fmt.Errorf("no strokes found (only stroked path/circle/ellipse/line/rect/polyline/polygon are drawn)")
	}
	s := float64(size) / math.Max(ic.vw, ic.vh)
	half := math.Max(0.5, ic.width*s/2)
	z := vector.NewRasterizer(size, size)
	tr := func(p pt) pt { return pt{(p.x - ic.vx) * s, (p.y - ic.vy) * s} }
	for _, l := range ic.lines {
		for i, p := range l {
			q := tr(p)
			addPoly(z, circlePts(q, half))
			if i > 0 {
				a := tr(l[i-1])
				dx, dy := q.x-a.x, q.y-a.y
				d := math.Hypot(dx, dy)
				if d < 1e-9 {
					continue
				}
				nx, ny := -dy/d*half, dx/d*half
				addPoly(z, []pt{{a.x + nx, a.y + ny}, {q.x + nx, q.y + ny}, {q.x - nx, q.y - ny}, {a.x - nx, a.y - ny}})
			}
		}
	}
	m := image.NewAlpha(image.Rect(0, 0, size, size))
	z.Draw(m, m.Bounds(), image.Opaque, image.Point{})
	return m, nil
}

// addPoly adds a closed polygon, always wound the same way so overlapping shapes add up (a union)
// instead of cancelling.
func addPoly(z *vector.Rasterizer, p []pt) {
	a := 0.0
	for i := range p {
		j := (i + 1) % len(p)
		a += p[i].x*p[j].y - p[j].x*p[i].y
	}
	if a < 0 {
		for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
			p[i], p[j] = p[j], p[i]
		}
	}
	z.MoveTo(float32(p[0].x), float32(p[0].y))
	for _, q := range p[1:] {
		z.LineTo(float32(q.x), float32(q.y))
	}
	z.ClosePath()
}

func circlePts(c pt, r float64) []pt {
	n := max(8, int(r*2))
	out := make([]pt, n)
	for i := range out {
		t := 2 * math.Pi * float64(i) / float64(n)
		out[i] = pt{c.x + r*math.Cos(t), c.y + r*math.Sin(t)}
	}
	return out
}

func parseSVG(data []byte) (*svgIcon, error) {
	ic := &svgIcon{vw: 24, vh: 24, width: 2}
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	root := true
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("not an SVG this can read: %w", err)
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		at := map[string]string{}
		for _, a := range se.Attr {
			at[a.Name.Local] = a.Value
		}
		f := func(k string) float64 { v, _ := strconv.ParseFloat(strings.TrimSuffix(at[k], "px"), 64); return v }
		if se.Name.Local == "svg" && root {
			root = false
			if vb := strings.Fields(strings.ReplaceAll(at["viewBox"], ",", " ")); len(vb) == 4 {
				ic.vx, _ = strconv.ParseFloat(vb[0], 64)
				ic.vy, _ = strconv.ParseFloat(vb[1], 64)
				ic.vw, _ = strconv.ParseFloat(vb[2], 64)
				ic.vh, _ = strconv.ParseFloat(vb[3], 64)
			} else if at["width"] != "" {
				ic.vw, ic.vh = f("width"), f("height")
			}
			if at["stroke-width"] != "" {
				ic.width = f("stroke-width")
			}
			continue
		}
		switch se.Name.Local {
		case "path":
			ic.lines = append(ic.lines, pathLines(at["d"])...)
		case "circle":
			ic.lines = append(ic.lines, ellipseLine(f("cx"), f("cy"), f("r"), f("r")))
		case "ellipse":
			ic.lines = append(ic.lines, ellipseLine(f("cx"), f("cy"), f("rx"), f("ry")))
		case "line":
			ic.lines = append(ic.lines, []pt{{f("x1"), f("y1")}, {f("x2"), f("y2")}})
		case "rect":
			x, y, w, h, rx := f("x"), f("y"), f("width"), f("height"), f("rx")
			if rx == 0 {
				rx = f("ry")
			}
			ic.lines = append(ic.lines, pathLines(rectPath(x, y, w, h, rx))...)
		case "polyline", "polygon":
			ns := nums(at["points"])
			var l []pt
			for i := 0; i+1 < len(ns); i += 2 {
				l = append(l, pt{ns[i], ns[i+1]})
			}
			if se.Name.Local == "polygon" && len(l) > 0 {
				l = append(l, l[0])
			}
			ic.lines = append(ic.lines, l)
		}
	}
	return ic, nil
}

func rectPath(x, y, w, h, r float64) string {
	r = math.Min(r, math.Min(w, h)/2)
	if r <= 0 {
		return fmt.Sprintf("M%g %gh%gv%gh%gz", x, y, w, h, -w)
	}
	return fmt.Sprintf("M%g %gh%ga%g %g 0 0 1 %g %gv%ga%g %g 0 0 1 %g %gh%ga%g %g 0 0 1 %g %gv%ga%g %g 0 0 1 %g %gz",
		x+r, y, w-2*r, r, r, r, r, h-2*r, r, r, -r, r, -(w - 2*r), r, r, -r, -r, -(h - 2*r), r, r, r, -r)
}

func ellipseLine(cx, cy, rx, ry float64) []pt {
	var l []pt
	for i := 0; i <= 48; i++ {
		t := 2 * math.Pi * float64(i) / 48
		l = append(l, pt{cx + rx*math.Cos(t), cy + ry*math.Sin(t)})
	}
	return l
}

// nums reads the numbers of an SVG number list ("1.5-2 .3e1,4").
func nums(s string) []float64 {
	var out []float64
	for _, t := range numTokens(s, nil) {
		if v, err := strconv.ParseFloat(t, 64); err == nil {
			out = append(out, v)
		}
	}
	return out
}

// numTokens splits a run of SVG numbers; flagAt(i) says the i-th number is an arc flag (a single
// digit that may be written with no separator: "0 01").
func numTokens(s string, flagAt func(int) bool) []string {
	var out []string
	i := 0
	for i < len(s) {
		c := s[i]
		if c == ' ' || c == ',' || c == '\t' || c == '\n' || c == '\r' {
			i++
			continue
		}
		if flagAt != nil && flagAt(len(out)) && (c == '0' || c == '1') {
			out = append(out, string(c))
			i++
			continue
		}
		j := i
		if s[j] == '-' || s[j] == '+' {
			j++
		}
		dot, exp := false, false
		for j < len(s) {
			d := s[j]
			switch {
			case d >= '0' && d <= '9':
			case d == '.' && !dot && !exp:
				dot = true
			case (d == 'e' || d == 'E') && !exp:
				exp = true
				if j+1 < len(s) && (s[j+1] == '-' || s[j+1] == '+') {
					j++
				}
			default:
				goto done
			}
			j++
		}
	done:
		if j == i {
			i++
			continue
		}
		out = append(out, s[i:j])
		i = j
	}
	return out
}

// pathLines flattens path data into polylines (curves and arcs into short segments).
func pathLines(d string) [][]pt {
	var lines [][]pt
	var cur []pt
	var p, start, ctrl pt
	var last byte
	flush := func() {
		if len(cur) > 1 {
			lines = append(lines, cur)
		}
		cur = nil
	}
	add := func(q pt) {
		if len(cur) == 0 {
			cur = append(cur, p)
		}
		cur = append(cur, q)
		p = q
	}
	i := 0
	for i < len(d) {
		c := d[i]
		if !strings.ContainsRune("MmLlHhVvCcSsQqTtAaZz", rune(c)) {
			i++
			continue
		}
		j := i + 1
		for j < len(d) && !strings.ContainsRune("MmLlHhVvCcSsQqTtAaZz", rune(d[j])) {
			j++
		}
		cmd := c
		var a []float64
		if cmd == 'A' || cmd == 'a' {
			for _, t := range numTokens(d[i+1:j], func(k int) bool { return k%7 == 3 || k%7 == 4 }) {
				v, _ := strconv.ParseFloat(t, 64)
				a = append(a, v)
			}
		} else {
			a = nums(d[i+1 : j])
		}
		i = j
		rel := cmd >= 'a'
		up := cmd &^ 0x20
		off := func(x, y float64) pt {
			if rel {
				return pt{p.x + x, p.y + y}
			}
			return pt{x, y}
		}
		switch up {
		case 'Z':
			if len(cur) > 0 {
				add(start)
			}
			flush()
			p = start
		case 'M':
			for k := 0; k+1 < len(a); k += 2 {
				q := off(a[k], a[k+1])
				if k == 0 {
					flush()
					p, start = q, q
				} else {
					add(q)
				}
			}
		case 'L':
			for k := 0; k+1 < len(a); k += 2 {
				add(off(a[k], a[k+1]))
			}
		case 'H':
			for _, x := range a {
				if rel {
					add(pt{p.x + x, p.y})
				} else {
					add(pt{x, p.y})
				}
			}
		case 'V':
			for _, y := range a {
				if rel {
					add(pt{p.x, p.y + y})
				} else {
					add(pt{p.x, y})
				}
			}
		case 'C', 'S':
			step := 6
			if up == 'S' {
				step = 4
			}
			for k := 0; k+step-1 < len(a); k += step {
				var c1, c2, e pt
				if up == 'C' {
					c1, c2, e = off(a[k], a[k+1]), off(a[k+2], a[k+3]), off(a[k+4], a[k+5])
				} else {
					c1 = p
					if last == 'C' || last == 'S' {
						c1 = pt{2*p.x - ctrl.x, 2*p.y - ctrl.y}
					}
					c2, e = off(a[k], a[k+1]), off(a[k+2], a[k+3])
				}
				s := p
				for n := 1; n <= 16; n++ {
					t := float64(n) / 16
					u := 1 - t
					add(pt{u*u*u*s.x + 3*u*u*t*c1.x + 3*u*t*t*c2.x + t*t*t*e.x, u*u*u*s.y + 3*u*u*t*c1.y + 3*u*t*t*c2.y + t*t*t*e.y})
				}
				ctrl, last = c2, up
			}
			continue
		case 'Q', 'T':
			step := 4
			if up == 'T' {
				step = 2
			}
			for k := 0; k+step-1 < len(a); k += step {
				var c1, e pt
				if up == 'Q' {
					c1, e = off(a[k], a[k+1]), off(a[k+2], a[k+3])
				} else {
					c1 = p
					if last == 'Q' || last == 'T' {
						c1 = pt{2*p.x - ctrl.x, 2*p.y - ctrl.y}
					}
					e = off(a[k], a[k+1])
				}
				s := p
				for n := 1; n <= 12; n++ {
					t := float64(n) / 12
					u := 1 - t
					add(pt{u*u*s.x + 2*u*t*c1.x + t*t*e.x, u*u*s.y + 2*u*t*c1.y + t*t*e.y})
				}
				ctrl, last = c1, up
			}
			continue
		case 'A':
			for k := 0; k+6 < len(a); k += 7 {
				e := off(a[k+5], a[k+6])
				for _, q := range arcPts(p, e, a[k], a[k+1], a[k+2], a[k+3] != 0, a[k+4] != 0) {
					add(q)
				}
			}
		}
		last = up
	}
	flush()
	return lines
}

// arcPts flattens an SVG elliptical arc from p to e (endpoint parameterisation → centre form).
func arcPts(p, e pt, rx, ry, phiDeg float64, large, sweep bool) []pt {
	if rx == 0 || ry == 0 || (p.x == e.x && p.y == e.y) {
		return []pt{e}
	}
	rx, ry = math.Abs(rx), math.Abs(ry)
	phi := phiDeg * math.Pi / 180
	cs, sn := math.Cos(phi), math.Sin(phi)
	dx, dy := (p.x-e.x)/2, (p.y-e.y)/2
	x1 := cs*dx + sn*dy
	y1 := -sn*dx + cs*dy
	if l := x1*x1/(rx*rx) + y1*y1/(ry*ry); l > 1 {
		rx, ry = rx*math.Sqrt(l), ry*math.Sqrt(l)
	}
	num := rx*rx*ry*ry - rx*rx*y1*y1 - ry*ry*x1*x1
	den := rx*rx*y1*y1 + ry*ry*x1*x1
	co := math.Sqrt(math.Max(0, num/den))
	if large == sweep {
		co = -co
	}
	cx1, cy1 := co*rx*y1/ry, -co*ry*x1/rx
	cx := cs*cx1 - sn*cy1 + (p.x+e.x)/2
	cy := sn*cx1 + cs*cy1 + (p.y+e.y)/2
	ang := func(ux, uy, vx, vy float64) float64 {
		a := math.Atan2(ux*vy-uy*vx, ux*vx+uy*vy)
		return a
	}
	t1 := ang(1, 0, (x1-cx1)/rx, (y1-cy1)/ry)
	dt := ang((x1-cx1)/rx, (y1-cy1)/ry, (-x1-cx1)/rx, (-y1-cy1)/ry)
	if !sweep && dt > 0 {
		dt -= 2 * math.Pi
	} else if sweep && dt < 0 {
		dt += 2 * math.Pi
	}
	n := max(4, int(math.Abs(dt)/(math.Pi/16)))
	out := make([]pt, 0, n)
	for i := 1; i <= n; i++ {
		t := t1 + dt*float64(i)/float64(n)
		x, y := rx*math.Cos(t), ry*math.Sin(t)
		out = append(out, pt{cs*x - sn*y + cx, sn*x + cs*y + cy})
	}
	out[len(out)-1] = e
	return out
}

// isSVG says whether data looks like an SVG document.
func IsSVG(data []byte) bool {
	s := strings.TrimSpace(string(data[:min(len(data), 512)]))
	return strings.Contains(s, "<svg") || strings.HasPrefix(s, "<?xml")
}

var _ = draw.Over
