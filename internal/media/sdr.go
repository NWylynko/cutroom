package media

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/NWylynko/cutroom/internal/deps"
)

// SDR describes how to read a source so its frames come out as SDR bt709 (L23): HDR sources (HLG,
// PQ) are tone-mapped on the way in; everything else is read as-is. Render, look and check all read
// sources through it, so they see the same pixels.
type SDR struct {
	HDR      string   // "", "hlg", "pq"
	Method   string   // "", "videotoolbox", "zscale", "none" (HDR but no tone mapper: washed out)
	Pre      []string // input options, placed before -i
	Filter   string   // filter chain to start the video graph with ("" = none); ends in a software frame
	Rotation int      // display rotation; on the hardware path it is applied inside Filter
}

var (
	sdrMu    sync.Mutex
	sdrCache = map[string]SDR{}
	toolOnce sync.Once
	haveVT   bool
	haveZ    bool
)

// ToneMapper reports which HDR→SDR method this ffmpeg supports.
func ToneMapper() string {
	toolOnce.Do(func() {
		hw, _ := exec.Command("ffmpeg", "-hide_banner", "-hwaccels").Output()
		fl, _ := exec.Command("ffmpeg", "-hide_banner", "-filters").Output()
		haveVT = strings.Contains(string(hw), "videotoolbox") && strings.Contains(string(fl), " scale_vt ")
		haveZ = strings.Contains(string(fl), " zscale ") && strings.Contains(string(fl), " tonemap ")
	})
	switch {
	case haveVT:
		return "videotoolbox"
	case haveZ:
		return "zscale"
	}
	return ""
}

// ReadSDR returns how to read path as SDR. w, h (display orientation, 0 = keep) let the
// hardware path scale while still on the GPU, which is much faster for 4K sources.
func ReadSDR(path string, w, h int) SDR {
	sdrMu.Lock()
	base, ok := sdrCache[path]
	sdrMu.Unlock()
	if !ok {
		base = probeSDR(path)
		sdrMu.Lock()
		sdrCache[path] = base
		sdrMu.Unlock()
	}
	if base.HDR == "" {
		return base
	}
	s := base
	switch ToneMapper() {
	case "videotoolbox":
		s.Method = "videotoolbox"
		// -display_rotation 0 overrides the input's display matrix: autorotate can't run on hardware
		// frames, and -noautorotate would copy the matrix into the output (players rotate it twice)
		s.Pre = []string{"-hwaccel", "videotoolbox", "-hwaccel_output_format", "videotoolbox_vld", "-display_rotation", "0"}
		size := ""
		if w > 0 && h > 0 {
			if s.Rotation == 90 || s.Rotation == 270 {
				w, h = h, w // scale_vt works in stored orientation
			}
			size = fmt.Sprintf("w=%d:h=%d:", w, h)
		}
		f := "scale_vt=" + size + "color_matrix=bt709:color_primaries=bt709:color_transfer=bt709,hwdownload,format=p010le," +
			"setparams=colorspace=bt709:color_primaries=bt709:color_trc=bt709:range=tv"
		if r := rotateFilter(s.Rotation); r != "" {
			f += "," + r
		}
		s.Filter = f
	case "zscale":
		s.Method = "zscale"
		s.Filter = "zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=tonemap=hable:desat=0," +
			"zscale=t=bt709:m=bt709:r=tv,format=yuv420p"
	default:
		s.Method = "none"
	}
	return s
}

func probeSDR(path string) SDR {
	out, err := deps.Run("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries",
		"stream=color_transfer:stream_side_data=rotation", "-of", "default=nw=1", path)
	if err != nil {
		return SDR{}
	}
	s := SDR{}
	for _, l := range strings.Split(string(out), "\n") {
		k, v, _ := strings.Cut(strings.TrimSpace(l), "=")
		switch k {
		case "color_transfer":
			s.HDR = hdrKind(v)
		case "rotation":
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				r := int(-f)
				s.Rotation = ((r % 360) + 360) % 360
			}
		}
	}
	return s
}

func hdrKind(transfer string) string {
	switch transfer {
	case "arib-std-b67":
		return "hlg"
	case "smpte2084":
		return "pq"
	}
	return ""
}

// rotateFilter applies a display rotation by hand, as ffmpeg's autorotate would (it can't run on
// hardware frames).
func rotateFilter(deg int) string {
	switch deg {
	case 90:
		return "transpose=clock"
	case 180:
		return "hflip,vflip"
	case 270:
		return "transpose=cclock"
	}
	return ""
}

// Chain prefixes a filter chain with the SDR conversion ("" stays "").
func (s SDR) Chain(rest string) string {
	switch {
	case s.Filter == "":
		return rest
	case rest == "":
		return s.Filter
	}
	return s.Filter + "," + rest
}
