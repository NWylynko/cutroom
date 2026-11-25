package timeline

import "testing"

// Split keeps every frame: the halves meet at the split, both edges pinned, the edit as long as before.
func TestSplitKeepsFrames(t *testing.T) {
	tl := New("30", 1280, 720, "cam", "cam", 0, 30)
	tl.Cut("cam", 4, 6, "x", true)
	d := tl.Duration()
	l, r, ok := tl.Split("cam", 12.01)
	if !ok || tl.Duration() != d || len(tl.Segments) != 3 {
		t.Fatalf("split: %v %v %v", ok, tl.Duration(), tl.Segments)
	}
	a, b := tl.Find(l), tl.Find(r)
	if a.Out != b.In || a.Out != 12 || !a.PinOut || !b.PinIn {
		t.Errorf("halves %+v %+v", a, b)
	}
	if _, _, ok := tl.Split("cam", 5); ok {
		t.Errorf("split inside a cut should fail")
	}
}
