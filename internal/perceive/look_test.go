package perceive

import (
	"strings"
	"testing"
	"time"

	"github.com/NWylynko/cutroom/internal/raster"
)

// fit once looped forever: it cut 2 bytes and appended a 3-byte ellipsis.
func TestFitTerminatesAndFits(t *testing.T) {
	face := raster.Face("mono", 13)
	cases := []struct {
		s string
		w int
	}{{"keyboard.", 8}, {"a very long word indeed", 60}, {"ok", 100}, {"ünïcødé wörds", 40}}
	for _, c := range cases {
		done := make(chan string, 1)
		go func() { done <- fit(c.s, c.w) }()
		select {
		case got := <-done:
			if raster.Measure(face, got) > c.w {
				t.Errorf("fit(%q, %d) = %q is %d px wide", c.s, c.w, got, raster.Measure(face, got))
			}
			if raster.Measure(face, c.s) <= c.w && got != c.s {
				t.Errorf("fit(%q) changed a string that fits: %q", c.s, got)
			}
			if got != c.s && got != "" && !strings.HasSuffix(got, "…") {
				t.Errorf("truncated string must end with an ellipsis: %q", got)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("fit(%q, %d) did not terminate", c.s, c.w)
		}
	}
}
