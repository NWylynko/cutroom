package render

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"golang.org/x/image/font"

	"github.com/NWylynko/cutroom/internal/captions"
	"github.com/NWylynko/cutroom/internal/raster"
	"github.com/NWylynko/cutroom/internal/timeline"
)

// CaptionDefaults is cutroom's caption look: bold white on a translucent black box (L48).
var CaptionDefaults = timeline.Style{Font: "bold", Size: 0.068, Color: "#ffffff", Box: "#000000aa", Pad: 0.3,
	Highlight: "none", Outline: "none", Shadow: "none", BoxShadow: "none", Case: "as-is", Align: "center", Words: 3, Chars: 20}

// CardDefaults is cutroom's text-card look for a layout: white on a dark translucent panel, an amber
// bar on lower thirds.
func CardDefaults(layout string) timeline.Style {
	s := timeline.Style{Font: "bold", SubFont: "regular", Size: 0.055, SubSize: 0.036, Color: "#ffffff", SubColor: "#dcdce1",
		Box: "#0c0c10be", Accent: "#f59e0b", Outline: "none", Shadow: "none", BoxShadow: "none", Pad: 0.4, Case: "as-is", Align: "left", Width: 0.62}
	switch layout {
	case "title", "center":
		s.Size, s.SubSize, s.Width, s.Align, s.Accent = 0.085, 0.045, 0.84, "center", "none"
	case "top":
		s.Size, s.SubSize, s.Width, s.Align, s.Accent = 0.05, 0.034, 0.84, "center", "none"
	case "corner":
		s.Size, s.SubSize, s.Width, s.Accent = 0.04, 0.028, 0.4, "none"
	}
	return s
}

// look is a style resolved for drawing at one frame size: faces, pixel sizes, colours (nil = off).
type look struct {
	face, subFace                                                   font.Face
	size                                                            float64
	color, subColor, highlight, box, accent, outln, shade, boxShade color.Color
	outlineW, shadowOff                                             int
	radius                                                          float64
	padX, padY                                                      int
	upper, lower                                                    bool
	align                                                           string
	x, y                                                            *float64
	width                                                           int
	key                                                             string // for caches: the style and its font files
}

func colorOf(s string) color.Color {
	c, err := timeline.ParseColor(s)
	if err != nil || c[3] == 0 {
		return nil
	}
	return color.NRGBA{c[0], c[1], c[2], c[3]}
}

// resolve turns a full style (defaults filled in) into a look for a w×h frame.
func resolve(st timeline.Style, dir string, w, h int) (look, error) {
	H := float64(min(w, h))
	fr, err := raster.ResolveFont(st.Font, dir)
	if err != nil {
		return look{}, err
	}
	sub := st.SubFont
	if sub == "" {
		sub = st.Font
	}
	sr, err := raster.ResolveFont(sub, dir)
	if err != nil {
		return look{}, err
	}
	size := math.Round(st.Size * H)
	subSize := math.Round(st.SubSize * H)
	if subSize == 0 {
		subSize = math.Round(0.65 * size)
	}
	l := look{face: raster.RefFace(fr, size), subFace: raster.RefFace(sr, subSize), size: size,
		color: colorOf(st.Color), subColor: colorOf(st.SubColor), highlight: colorOf(st.Highlight), box: colorOf(st.Box),
		accent: colorOf(st.Accent), outln: colorOf(st.Outline), shade: colorOf(st.Shadow), boxShade: colorOf(st.BoxShadow),
		radius: st.Radius * size, padX: int(st.Pad * size), padY: int(0.6 * st.Pad * size),
		upper: st.Case == "upper", lower: st.Case == "lower", align: st.Align, x: st.X, y: st.Y,
		width: int(st.Width * float64(w))}
	if l.color == nil {
		l.color = color.NRGBA{255, 255, 255, 255}
	}
	if l.subColor == nil {
		l.subColor = l.color
	}
	if l.outln != nil {
		ow := st.OutlineWidth
		if ow == 0 {
			ow = 0.08
		}
		l.outlineW = max(1, int(math.Round(ow*size)))
	}
	if l.shade != nil {
		l.shadowOff = max(1, int(math.Round(0.06*size)))
	}
	b, _ := json.Marshal(st)
	l.key = string(b) + "|" + fr.Key + "|" + sr.Key
	return l, nil
}

func (l look) text(s string) string {
	switch {
	case l.upper:
		return strings.ToUpper(s)
	case l.lower:
		return strings.ToLower(s)
	}
	return s
}

func (l look) textStyle() raster.TextStyle {
	return raster.TextStyle{Outline: l.outln, OutlineW: l.outlineW, Shadow: l.shade, ShadowOff: l.shadowOff}
}

// boxAt places a bw×bh box: centred on the style's x/y when given (kept inside the frame), else at
// the default top-left (dx, dy).
func (l look) boxAt(w, h, bw, bh, dx, dy int) image.Rectangle {
	x, y := dx, dy
	if l.x != nil {
		x = int(math.Round(*l.x*float64(w))) - bw/2
	}
	if l.y != nil {
		y = int(math.Round(*l.y*float64(h))) - bh/2
	}
	x = max(0, min(x, w-bw))
	y = max(0, min(y, h-bh))
	return image.Rect(x, y, x+bw, y+bh)
}

// CaptionStyle is the project's full caption style (defaults filled in).
func CaptionStyle(tl *timeline.Timeline) timeline.Style {
	return tl.CaptionStyle.Over(CaptionDefaults)
}

// CardStyle is a text overlay's full style: its own over the project's card style over the defaults.
func CardStyle(tl *timeline.Timeline, o timeline.Overlay) timeline.Style {
	base := tl.CardStyle.Over(CardDefaults(o.Layout))
	return o.Style.Over(base)
}

// GroupOptions is how caption words are grouped under a style.
func GroupOptions(st timeline.Style) captions.GroupOptions {
	g := captions.DefaultGroups
	if st.Words > 0 {
		g.MaxWords = st.Words
	}
	if st.Chars > 0 {
		g.MaxChars = st.Chars
	}
	return g
}

// CheckStyle resolves a style's fonts and colours, so a bad one is refused when it is set, not at
// render time.
func CheckStyle(st timeline.Style, dir string) error {
	for _, f := range []string{st.Font, st.SubFont} {
		if f != "" {
			if _, err := raster.ResolveFont(f, dir); err != nil {
				return err
			}
		}
	}
	return nil
}

// FontName is the face a font spec resolves to (for printing), or the error.
func FontName(spec, dir string) string {
	r, err := raster.ResolveFont(spec, dir)
	if err != nil {
		return fmt.Sprintf("? (%v)", err)
	}
	if r.Path != "" {
		return fmt.Sprintf("%s (%s)", r.Name, r.Path)
	}
	return r.Name
}

// drawBox draws the style's box shadow and box at r.
func (l look) drawBox(img *image.NRGBA, r image.Rectangle) {
	if l.boxShade != nil {
		raster.SoftShadow(img, r, l.radius, int(0.15*l.size), int(0.35*l.size), l.boxShade)
	}
	if l.box != nil {
		raster.FillRound(img, r, l.radius, l.box)
	}
}
