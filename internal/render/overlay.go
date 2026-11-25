package render

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"

	"github.com/cutroom/cutroom/internal/captions"
	"github.com/cutroom/cutroom/internal/deps"
	"github.com/cutroom/cutroom/internal/media"
	"github.com/cutroom/cutroom/internal/raster"
	"github.com/cutroom/cutroom/internal/timeline"
)

// CardVersion changes whenever card drawing changes, so cached cards are redrawn.
const CardVersion = 1

// Placement is one overlay as rendered: a full-frame RGBA PNG shown over output [Start, End),
// fading in and out over Fade seconds. Box is the drawn area as fractions of the frame
// (x0, y0, x1, y1), for checks.
type Placement struct {
	ID    string     `json:"id"`
	Kind  string     `json:"kind"`
	PNG   string     `json:"png"`
	Key   string     `json:"key"`
	Start float64    `json:"start"`
	End   float64    `json:"end"`
	Fade  float64    `json:"fade"`
	Box   [4]float64 `json:"box"`
	Name  string     `json:"name"`
	// Groups: burned-in captions, each a card shown over [Start, End)
	Groups []CapGroup `json:"groups,omitempty"`
	// clip: the video (absolute path) and where in it the overlay starts
	Clip    string  `json:"clip,omitempty"`
	ClipIn  float64 `json:"clip_in,omitempty"`
	ClipDur float64 `json:"clip_duration,omitempty"`
	Full    bool    `json:"full,omitempty"` // covers the whole frame
}

// CapGroup is one burned-in caption group as rendered.
type CapGroup struct {
	Start float64    `json:"start"`
	End   float64    `json:"end"`
	Text  string     `json:"text"`
	PNG   string     `json:"png"`
	Key   string     `json:"key"`
	Box   [4]float64 `json:"box"`
}

// Placements resolves every overlay of tl that shows to a card PNG at w×h (cached in cacheDir).
// Overlays whose anchor was cut are skipped (the timeline warns about them).
func Placements(tl *timeline.Timeline, tr timeline.Transcripts, projDir, cacheDir string, w, h int) ([]Placement, error) {
	var ps []Placement
	burned := tl.Captions() != nil
	for _, o := range tl.Overlays {
		a, b, ok, _ := tl.Window(o)
		if !ok {
			continue
		}
		if o.Kind == "captions" {
			p, err := captionPlacement(tl, tr, o, cacheDir, w, h)
			if err != nil {
				return nil, err
			}
			ps = append(ps, p)
			continue
		}
		if o.Kind == "clip" {
			abs := absIn(projDir, o.Path)
			st, err := os.Stat(abs)
			if err != nil {
				return nil, fmt.Errorf("overlay %s: %w", o.ID, err)
			}
			r := clipBox(o, w, h)
			ps = append(ps, Placement{ID: o.ID, Kind: "clip", Start: a, End: b, Fade: o.Fade, Name: o.Path, Clip: abs, ClipIn: o.ClipIn, ClipDur: o.ClipDur, Full: o.Layout == "full",
				Key: hash([]any{CardVersion, "clip", abs, st.Size(), st.ModTime().UnixNano(), o.ClipIn, o.Layout, o.Size, w, h}),
				Box: [4]float64{float64(r.Min.X) / float64(w), float64(r.Min.Y) / float64(h), float64(r.Max.X) / float64(w), float64(r.Max.Y) / float64(h)}})
			continue
		}
		if burned && (o.Layout == "lower-third" || o.Layout == "") && o.Kind != "captions" {
			o.Layout = "lower-third-raised" // keep clear of the caption band
		}
		key, err := cardKey(o, projDir, w, h)
		if err != nil {
			return nil, err
		}
		png := filepath.Join(cacheDir, "overlays", key+".png")
		box, err := cardBox(o, projDir, w, h)
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(png); err != nil {
			img, _, err := Card(o, projDir, w, h)
			if err != nil {
				return nil, err
			}
			if err := ensureDir(png); err != nil {
				return nil, err
			}
			if err := raster.SavePNG(png, img); err != nil {
				return nil, err
			}
		}
		name := o.Text
		if o.Kind == "image" {
			name = o.Path
		}
		ps = append(ps, Placement{ID: o.ID, Kind: o.Kind, PNG: png, Key: key, Start: a, End: b, Fade: o.Fade, Box: box, Name: name})
	}
	return ps, nil
}

// cardKey hashes everything a card depends on, including the bytes of a referenced image (L19).
func cardKey(o timeline.Overlay, projDir string, w, h int) (string, error) {
	parts := []any{CardVersion, w, h, o.Kind, o.Text, o.Sub, o.Layout, o.Size, o.Path}
	if o.Kind == "image" {
		b, err := os.ReadFile(absIn(projDir, o.Path))
		if err != nil {
			return "", fmt.Errorf("overlay %s: %w", o.ID, err)
		}
		parts = append(parts, fmt.Sprintf("%x", sha256.Sum256(b)))
	}
	return hash(parts), nil
}

func absIn(dir, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(dir, p)
}

func cardBox(o timeline.Overlay, projDir string, w, h int) ([4]float64, error) {
	_, r, err := Card(o, projDir, w, h)
	if err != nil {
		return [4]float64{}, err
	}
	return [4]float64{float64(r.Min.X) / float64(w), float64(r.Min.Y) / float64(h), float64(r.Max.X) / float64(w), float64(r.Max.Y) / float64(h)}, nil
}

var (
	cardBG     = color.NRGBA{12, 12, 16, 190}
	cardAccent = color.NRGBA{245, 158, 11, 255}
)

// Card draws an overlay into a transparent w×h image; r is the area it covers.
func Card(o timeline.Overlay, projDir string, w, h int) (img *image.NRGBA, r image.Rectangle, err error) {
	img = image.NewNRGBA(image.Rect(0, 0, w, h))
	H := float64(min(w, h)) // text scales with the short side: the same look on landscape and vertical
	margin := int(0.06 * float64(w))
	switch o.Kind {
	case "image":
		f, err := os.Open(absIn(projDir, o.Path))
		if err != nil {
			return nil, r, fmt.Errorf("overlay %s: %w", o.ID, err)
		}
		defer f.Close()
		src, _, err := image.Decode(f)
		if err != nil {
			return nil, r, fmt.Errorf("overlay %s: %s is not a PNG or JPEG: %w", o.ID, o.Path, err)
		}
		size := o.Size
		if size <= 0 {
			size = 0.25
		}
		iw := int(size * float64(w))
		ih := iw * src.Bounds().Dy() / max(1, src.Bounds().Dx())
		if ih > h-2*margin {
			ih = h - 2*margin
			iw = ih * src.Bounds().Dx() / max(1, src.Bounds().Dy())
		}
		x, y := place(o.Layout, w, h, iw, ih, margin)
		r = image.Rect(x, y, x+iw, y+ih)
		xdraw.CatmullRom.Scale(img, r, src, src.Bounds(), draw.Over, nil)
		return img, r, nil
	case "text":
		big, small := 0.055, 0.036
		maxW := 0.62
		switch o.Layout {
		case "title", "center":
			big, small, maxW = 0.085, 0.045, 0.84
		case "top":
			big, small, maxW = 0.05, 0.034, 0.84
		case "corner":
			big, small, maxW = 0.04, 0.028, 0.4
		}
		bf := raster.Face("bold", math.Round(big*H))
		sf := raster.Face("regular", math.Round(small*H))
		mw := int(maxW * float64(w))
		if h > w { // vertical: stay left of the right-hand buttons
			_, _, right := SafeZone(w, h)
			mw = min(mw, right-2*margin)
		}
		pad := int(0.022 * H)
		var lines []line
		for _, l := range wrapText(bf, o.Text, mw) {
			lines = append(lines, line{l, bf, color.NRGBA{255, 255, 255, 255}})
		}
		if o.Sub != "" {
			for _, l := range wrapText(sf, o.Sub, mw) {
				lines = append(lines, line{l, sf, color.NRGBA{220, 220, 225, 255}})
			}
		}
		tw, th := 0, 0
		for _, l := range lines {
			tw = max(tw, raster.Measure(l.face, l.s))
			th += lineH(l.face)
		}
		accent := 0
		if o.Layout == "lower-third" || o.Layout == "" || o.Layout == "lower-third-raised" {
			accent = int(math.Max(4, 0.006*float64(w)))
		}
		bw, bh := tw+2*pad+accent, th+2*pad
		x, y := place(o.Layout, w, h, bw, bh, margin)
		r = image.Rect(x, y, x+bw, y+bh)
		draw.Draw(img, r, image.NewUniform(cardBG), image.Point{}, draw.Over)
		if accent > 0 {
			draw.Draw(img, image.Rect(x, y, x+accent, y+bh), image.NewUniform(cardAccent), image.Point{}, draw.Over)
		}
		cy := y + pad
		for _, l := range lines {
			lx := x + accent + pad
			if o.Layout == "title" || o.Layout == "center" || o.Layout == "top" {
				lx = x + (bw-raster.Measure(l.face, l.s))/2
			}
			raster.Text(img, l.face, lx, cy+l.face.Metrics().Ascent.Round(), l.c, l.s)
			cy += lineH(l.face)
		}
		return img, r, nil
	}
	return nil, r, fmt.Errorf("overlay %s: unknown kind %q", o.ID, o.Kind)
}

type line struct {
	s    string
	face font.Face
	c    color.Color
}

func lineH(f font.Face) int {
	m := f.Metrics()
	return int(float64((m.Ascent + m.Descent).Round()) * 1.12)
}

// Vertical safe zones (L41, for 1080×1920; scaled to any vertical frame): the app's own title bar
// covers the top, its caption, buttons and description cover the bottom and the right edge.
func SafeZone(W, H int) (top, bottom, right int) {
	if H <= W {
		return 0, H, W
	}
	return int(math.Ceil(260 * float64(H) / 1920)), int(1490 * float64(H) / 1920), W - int(math.Ceil(120*float64(W)/1080))
}

// place puts a w×h box in the frame for a layout (inside the safe zones on vertical frames).
func place(layout string, W, H, w, h, margin int) (int, int) {
	if H > W {
		top, bottom, right := SafeZone(W, H)
		switch layout {
		case "title", "center": // centred in the safe area
			return max(0, (right-w)/2), max(top, top+(bottom-top-h)/2)
		case "top":
			return max(0, (right-w)/2), top
		case "corner":
			return right - w, top
		case "lower-third-raised":
			return margin, int(captionTop(W, H)) - int(0.02*float64(H)) - h
		default: // lower-third: above the bottom zone
			return margin, bottom - int(0.02*float64(H)) - h
		}
	}
	switch layout {
	case "title", "center":
		return (W - w) / 2, (H - h) / 2
	case "top":
		return (W - w) / 2, int(0.08 * float64(H))
	case "corner":
		return W - margin - w, int(0.06 * float64(H))
	case "lower-third-raised": // above burned-in captions
		return margin, int(captionTop(W, H)) - int(0.03*float64(H)) - h
	default: // lower-third
		return margin, H - int(0.12*float64(H)) - h
	}
}

// wrapText breaks s into lines no wider than mw (explicit "\n" kept).
func wrapText(f font.Face, s string, mw int) []string {
	var out []string
	for _, para := range strings.Split(s, "\n") {
		cur := ""
		for _, wd := range strings.Fields(para) {
			try := strings.TrimSpace(cur + " " + wd)
			if cur != "" && raster.Measure(f, try) > mw {
				out = append(out, cur)
				cur = wd
			} else {
				cur = try
			}
		}
		if cur != "" {
			out = append(out, cur)
		}
	}
	return out
}

// Layouts are the overlay positions.
var Layouts = []string{"lower-third", "title", "top", "corner", "center"}

// --- burned-in captions (L48) ----------------------------------------------------------------

// captionBottom is where the caption box ends: low on landscape frames; on vertical frames above
// the platform's buttons and description (safe zone, L41: nothing below y 1490 of 1920).
func captionBottom(W, H int) float64 {
	if H > W {
		return 0.77 * float64(H)
	}
	return 0.92 * float64(H)
}

func captionSize(W, H int) float64 { return math.Round(0.068 * float64(min(W, H))) }

// captionTop is the top of the caption band (for raising lower thirds).
func captionTop(W, H int) float64 {
	return captionBottom(W, H) - 1.6*captionSize(W, H)
}

func captionPlacement(tl *timeline.Timeline, tr timeline.Transcripts, o timeline.Overlay, cacheDir string, w, h int) (Placement, error) {
	ws, missing := captions.OutputWords(tl, tr)
	if len(missing) > 0 {
		return Placement{}, fmt.Errorf("burned-in captions need a transcript of %s: run `cutroom transcribe %s`", missing[0], missing[0])
	}
	p := Placement{ID: o.ID, Kind: "captions", Start: 0, End: tl.Duration(), Name: "burned-in captions", Key: "captions"}
	var keys []any
	for _, g := range captions.Groups(ws, captions.DefaultGroups, tl.Duration()) {
		key := hash([]any{CardVersion, "caption", w, h, g.Text})
		png := filepath.Join(cacheDir, "overlays", key+".png")
		img, r := CaptionCard(g.Text, w, h)
		if _, err := os.Stat(png); err != nil {
			if err := ensureDir(png); err != nil {
				return p, err
			}
			if err := raster.SavePNG(png, img); err != nil {
				return p, err
			}
		}
		box := [4]float64{float64(r.Min.X) / float64(w), float64(r.Min.Y) / float64(h), float64(r.Max.X) / float64(w), float64(r.Max.Y) / float64(h)}
		p.Groups = append(p.Groups, CapGroup{Start: g.Start, End: g.End, Text: g.Text, PNG: png, Key: key, Box: box})
		keys = append(keys, []any{key, g.Start, g.End})
		for i := range 4 { // the placement's box covers every group's box
			if p.Box == ([4]float64{}) {
				p.Box = box
			}
			if i < 2 {
				p.Box[i] = math.Min(p.Box[i], box[i])
			} else {
				p.Box[i] = math.Max(p.Box[i], box[i])
			}
		}
	}
	p.Key = hash(keys)
	return p, nil
}

// CaptionCard draws one caption group: bold white on a dark box, centred in the caption band.
func CaptionCard(text string, w, h int) (*image.NRGBA, image.Rectangle) {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	f := raster.Face("bold", captionSize(w, h))
	m := f.Metrics()
	tw := raster.Measure(f, text)
	padX, padY := int(0.3*captionSize(w, h)), int(0.18*captionSize(w, h))
	bw, bh := tw+2*padX, (m.Ascent+m.Descent).Round()+2*padY
	x, y := (w-bw)/2, int(captionBottom(w, h))-bh
	if _, _, right := SafeZone(w, h); h > w {
		x = max(0, (right-bw)/2) // vertical: centred left of the right-hand buttons (L41)
	}
	r := image.Rect(x, y, x+bw, y+bh)
	draw.Draw(img, r, image.NewUniform(color.NRGBA{0, 0, 0, 170}), image.Point{}, draw.Over)
	raster.Text(img, f, x+padX, y+padY+m.Ascent.Round(), color.NRGBA{255, 255, 255, 255}, text)
	return img, r
}

// blankPNG is a fully transparent frame (gaps between caption groups).
func blankPNG(cacheDir string, w, h int) (string, error) {
	p := filepath.Join(cacheDir, "overlays", fmt.Sprintf("blank-%dx%d.png", w, h))
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}
	if err := ensureDir(p); err != nil {
		return "", err
	}
	return p, raster.SavePNG(p, image.NewNRGBA(image.Rect(0, 0, w, h)))
}

// clipBox is where a clip overlay sits: the whole frame (full), or a box of --size × the frame
// width (default 0.35) with the clip's aspect, placed by layout. Sizes are even (yuv420).
func clipBox(o timeline.Overlay, w, h int) image.Rectangle {
	if o.Layout == "full" {
		return image.Rect(0, 0, w, h)
	}
	size := o.Size
	if size <= 0 {
		size = 0.35
	}
	bw := even(size * float64(w))
	bh := even(float64(bw) * float64(max(1, o.ClipH)) / float64(max(1, o.ClipW)))
	margin := int(0.06 * float64(w))
	if bh > h-2*margin {
		bh = even(float64(h - 2*margin))
		bw = even(float64(bh) * float64(max(1, o.ClipW)) / float64(max(1, o.ClipH)))
	}
	x, y := place(o.Layout, w, h, bw, bh, margin)
	return image.Rect(x, y, x+bw, y+bh)
}

// ClipCard draws the frame a clip overlay shows at output time t into a transparent w×h image (for
// `look --output`), cached in cacheDir.
func ClipCard(p Placement, t float64, cacheDir string, w, h int) (string, error) {
	ct := p.ClipIn + (t - p.Start)
	if p.ClipDur > 0 { // past its end the render holds the last frame
		ct = math.Min(ct, p.ClipDur-0.1)
	}
	png := filepath.Join(cacheDir, "overlays", hash([]any{p.Key, round3(ct), w, h})+".png")
	if _, err := os.Stat(png); err == nil {
		return png, nil
	}
	bx, by := int(math.Round(p.Box[0]*float64(w))), int(math.Round(p.Box[1]*float64(h)))
	bw, bh := even(p.Box[2]*float64(w)-float64(bx)), even(p.Box[3]*float64(h)-float64(by))
	sdr := media.ReadSDR(p.Clip, 0, 0)
	frame := png + ".frame.png"
	args := append([]string{"-hide_banner", "-loglevel", "error", "-y", "-ss", f6(math.Max(0, ct))}, sdr.Pre...)
	args = append(args, "-i", p.Clip, "-vf", sdr.Chain(fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=increase,crop=%d:%d", bw, bh, bw, bh)), "-frames:v", "1", frame)
	if err := ensureDir(png); err != nil {
		return "", err
	}
	if _, err := deps.Run("ffmpeg", args...); err != nil {
		return "", err
	}
	defer os.Remove(frame)
	src, err := raster.LoadPNG(frame)
	if err != nil {
		return "", err
	}
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, image.Rect(bx, by, bx+bw, by+bh), src, src.Bounds().Min, draw.Src)
	return png, raster.SavePNG(png, img)
}
