package media

import (
	"strings"
	"testing"
)

// L23: HDR sources are tone-mapped on the way in. On the VideoToolbox path the display matrix must be
// overridden with -display_rotation 0 (not -noautorotate, which copies it into the output so
// players rotate the picture twice), and the rotation applied by hand after hwdownload.
func TestL23_HardwareToneMapOverridesRotation(t *testing.T) {
	sdrMu.Lock()
	sdrCache["x.mov"] = SDR{HDR: "hlg", Rotation: 180}
	sdrMu.Unlock()
	toolOnce.Do(func() {})
	haveVT, haveZ = true, false
	s := ReadSDR("x.mov", 1920, 1080)
	pre := strings.Join(s.Pre, " ")
	if !strings.Contains(pre, "-display_rotation 0") || strings.Contains(pre, "noautorotate") {
		t.Fatalf("pre %q", pre)
	}
	if !strings.HasPrefix(s.Filter, "scale_vt=w=1920:h=1080:") || !strings.HasSuffix(s.Filter, "hflip,vflip") {
		t.Fatalf("filter %q", s.Filter)
	}
	if got := s.Chain("fps=30"); !strings.HasSuffix(got, ",hflip,vflip,fps=30") {
		t.Fatalf("chain %q", got)
	}
	// 90° sources are scaled in stored orientation
	sdrMu.Lock()
	sdrCache["y.mov"] = SDR{HDR: "pq", Rotation: 90}
	sdrMu.Unlock()
	if s := ReadSDR("y.mov", 1920, 1080); !strings.HasPrefix(s.Filter, "scale_vt=w=1080:h=1920:") || !strings.HasSuffix(s.Filter, "transpose=clock") {
		t.Fatalf("90° filter %q", s.Filter)
	}
	// SDR sources are read untouched
	sdrMu.Lock()
	sdrCache["z.mp4"] = SDR{}
	sdrMu.Unlock()
	if s := ReadSDR("z.mp4", 1920, 1080); len(s.Pre) != 0 || s.Chain("fps=30") != "fps=30" {
		t.Fatalf("sdr %+v", s)
	}
}
