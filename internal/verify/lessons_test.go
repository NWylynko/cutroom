package verify

import (
	"os"
	"path/filepath"
	"testing"
)

// L03: frame reads for checking keep a VFR source's real frames: across a dropped-frame gap the
// timestamps jump (no phantom duplicates filling it), and each frame comes once.
func TestL03_FrameReadsKeepVFRGaps(t *testing.T) {
	p := filepath.Join("..", "..", "evals", "work", "fixtures", "talk-vfr.mp4")
	if _, err := os.Stat(p); err != nil {
		t.Skip("fixture missing (run `make fixtures`)")
	}
	// frames 40 and 41 of the 30 fps original are dropped (≈ 1.333–1.400 s)
	fr, pts := greyBurst(p, "", 1.0, 20)
	if len(fr) != 20 || len(pts) != 20 {
		t.Fatalf("read %d frames, %d times", len(fr), len(pts))
	}
	gap := false
	for i := 1; i < len(pts); i++ {
		d := pts[i] - pts[i-1]
		if d <= 0.001 {
			t.Fatalf("frame %d repeats time %.4f: a duplicate filled a gap", i, pts[i])
		}
		gap = gap || d > 1.5/30
	}
	if !gap {
		t.Fatalf("no gap seen across the dropped frames: %v", pts)
	}
}
