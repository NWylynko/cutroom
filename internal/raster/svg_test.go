package raster

import "testing"

func TestSVGStrokesRender(t *testing.T) {
	svg := `<svg viewBox="0 0 24 24" stroke-width="2" fill="none"><circle cx="12" cy="12" r="8"/><path d="M4 12h16"/><path d="M12 4a8 8 0 0 1 0 16"/></svg>`
	m, err := SVGMask([]byte(svg), 48)
	if err != nil {
		t.Fatal(err)
	}
	at := func(x, y int) uint8 { return m.AlphaAt(x, y).A }
	if at(24, 24) < 200 { // the horizontal line through the centre
		t.Errorf("line not drawn at the centre: %d", at(24, 24))
	}
	if at(24, 8) < 200 { // top of the circle (r 8 → y 4 → 8 px)
		t.Errorf("circle not drawn at its top: %d", at(24, 8))
	}
	if at(16, 16) > 20 { // inside the circle, off the strokes: empty (no fill)
		t.Errorf("stroke icon was filled: %d", at(16, 16))
	}
	if _, err := SVGMask([]byte(`<svg viewBox="0 0 24 24"></svg>`), 24); err == nil {
		t.Errorf("an empty SVG should say it has no strokes")
	}
	if ns := numTokens("0 01-1.5.5", func(k int) bool { return k == 1 || k == 2 }); len(ns) != 5 {
		t.Errorf("arc flags written together: %q", ns)
	}
}
