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
	"sync"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"

	"github.com/NWylynko/cutroom/internal/captions"
	"github.com/NWylynko/cutroom/internal/deps"
	"github.com/NWylynko/cutroom/internal/media"
	"github.com/NWylynko/cutroom/internal/raster"
	"github.com/NWylynko/cutroom/internal/timeline"
)

// CardVersion changes whenever card drawing changes, so cached cards are redrawn.
const CardVersion = 5 // 5: eyebrow tracking 0.14 by default; 4: 4: highlight box drawn over the outline; 3: clips relabelled bt709 like sources (newer ffmpeg converted untagged clips from bt601)

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
	// Anim: a card's motion in and out (pop, slide-…; "" or fade = none) over AnimTime seconds
	Anim     string  `json:"anim,omitempty"`
	AnimTime float64 `json:"anim_time,omitempty"`
}

// CapGroup is one burned-in caption group as rendered.
type CapGroup struct {
	Start float64    `json:"start"`
	End   float64    `json:"end"`
	Text  string     `json:"text"`
	PNG   string     `json:"png"`
	Key   string     `json:"key"`
	Box   [4]float64 `json:"box"`
	Anim  bool       `json:"anim,omitempty"` // one frame of a group's entrance animation
}

// Placements resolves every overlay of tl that shows to a card PNG at w×h (cached in cacheDir).
// Overlays whose anchor was cut are skipped (the timeline warns about them).
func Placements(tl *timeline.Timeline, tr timeline.Transcripts, projDir, cacheDir string, w, h int) ([]Placement, error) {
	var ps []Placement
	burned := tl.Captions() != nil
	var capBox [4]float64 // where the captions really are, for lower thirds to keep clear of
	if c := tl.Captions(); c != nil {
		if p, err := captionPlacement(tl, tr, *c, projDir, cacheDir, w, h); err == nil {
			capBox = p.Box
		}
	}
	for _, o := range tl.Overlays {
		a, b, ok, _ := tl.Window(o)
		if !ok {
			continue
		}
		if o.Kind == "captions" {
			p, err := captionPlacement(tl, tr, o, projDir, cacheDir, w, h)
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
			o.CapBox = capBox
		}
		if o.Kind == "text" {
			st := CardStyle(tl, o)
			o.Style = &st // Card draws the full style
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
		cs := CardStyle(tl, o)
		ps = append(ps, Placement{ID: o.ID, Kind: o.Kind, PNG: png, Key: key, Start: a, End: b, Fade: o.Fade, Box: box, Name: name, Anim: motion(cs.Animate), AnimTime: cs.AnimateTime})
	}
	return ps, nil
}

// cardKey hashes everything a card depends on, including the bytes of a referenced image (L19).
func cardKey(o timeline.Overlay, projDir string, w, h int) (string, error) {
	parts := []any{CardVersion, w, h, o.Kind, o.Text, o.Sub, o.Layout, o.Size, o.Path, o.CapBox, o.Eyebrow}
	if o.Kind == "text" {
		l, err := resolve(o.Style.Over(CardDefaults(o.Layout)), projDir, w, h)
		if err != nil {
			return "", fmt.Errorf("overlay %s: %w", o.ID, err)
		}
		parts = append(parts, l.key)
	}
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

// Card draws an overlay into a transparent w×h image; r is the area it covers.
func Card(o timeline.Overlay, projDir string, w, h int) (img *image.NRGBA, r image.Rectangle, err error) {
	img = image.NewNRGBA(image.Rect(0, 0, w, h))
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
		x, y := placeCard(o, w, h, iw, ih, margin)
		r = image.Rect(x, y, x+iw, y+ih)
		xdraw.CatmullRom.Scale(img, r, src, src.Bounds(), draw.Over, nil)
		return img, r, nil
	case "text":
		l, err := resolve(o.Style.Over(CardDefaults(o.Layout)), projDir, w, h)
		if err != nil {
			return nil, r, fmt.Errorf("overlay %s: %w", o.ID, err)
		}
		l.padY = l.padX
		mw := l.width
		if h > w && l.x == nil { // vertical: stay left of the right-hand buttons
			_, _, right := SafeZone(w, h)
			mw = min(mw, right-2*margin)
		}
		var lines []line
		if o.Eyebrow != "" { // a small label above the title, in the accent colour (or the sub-line's)
			c := l.accent
			if c == nil {
				c = l.subColor
			}
			es := math.Round(0.5 * l.size)
			ef := raster.Tracked(raster.RefFace(l.fontRef, es), l.eyebrowTrack*es)
			for _, t := range wrapText(ef, strings.ToUpper(o.Eyebrow), mw) {
				lines = append(lines, line{t, ef, c})
			}
		}
		for _, t := range wrapText(l.face, l.text(o.Text), mw) {
			lines = append(lines, line{t, l.face, l.color})
		}
		if o.Sub != "" {
			for _, t := range wrapText(l.subFace, o.Sub, mw) {
				lines = append(lines, line{t, l.subFace, l.subColor})
			}
		}
		tw, th := 0, 0
		for _, ln := range lines {
			tw = max(tw, raster.Measure(ln.face, ln.s))
			th += lineH(ln.face)
		}
		accent := 0
		if l.accent != nil && (o.Layout == "lower-third" || o.Layout == "" || o.Layout == "lower-third-raised") {
			accent = int(math.Max(4, 0.006*float64(w)))
			if l.radius > 0.5 { // wide enough to show inside the rounded corners
				accent = max(accent, int(math.Min(l.radius*0.5, 0.02*float64(w))))
			}
		}
		bw, bh := tw+2*l.padX+accent+2*l.outlineW, th+2*l.padY
		x, y := placeCard(o, w, h, bw, bh, margin)
		r = l.boxAt(w, h, bw, bh, x, y)
		l.drawBox(img, r)
		if accent > 0 {
			if l.radius > 0.5 { // a rounded panel: the bar follows its left corners
				raster.FillRoundPart(img, r, l.radius, l.accent, accent)
			} else {
				raster.Fill(img, image.Rect(r.Min.X, r.Min.Y, r.Min.X+accent, r.Max.Y), l.accent)
			}
		}
		cy := r.Min.Y + l.padY
		for _, ln := range lines {
			lw := raster.Measure(ln.face, ln.s)
			lx := r.Min.X + accent + l.padX + l.outlineW
			switch l.align {
			case "center":
				lx += (tw - lw) / 2
			case "right":
				lx += tw - lw
			}
			raster.StyledLine(img, ln.face, lx, cy+ln.face.Metrics().Ascent.Round(), []raster.Run{{S: ln.s, C: ln.c}}, l.textStyle())
			cy += lineH(ln.face)
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

// placeCard places an overlay's w×h box: by its layout, and for a lower third raised over burned-in
// captions, clear of where the captions really are (their style may move or enlarge them): it stays
// in its usual place if that doesn't touch them, else goes just above them (below, if above would
// leave the safe area and below fits).
func placeCard(o timeline.Overlay, W, H, w, h, margin int) (int, int) {
	if o.Layout != "lower-third-raised" || o.CapBox == ([4]float64{}) {
		return place(o.Layout, W, H, w, h, margin)
	}
	x, y := place("lower-third", W, H, w, h, margin)
	gap := int(0.02 * float64(H))
	ct, cb := int(o.CapBox[1]*float64(H)), int(math.Ceil(o.CapBox[3]*float64(H)))
	if y+h+gap <= ct || y >= cb+gap {
		return x, y
	}
	top, bottom := 0, H
	if H > W {
		top, bottom, _ = SafeZone(W, H)
	}
	if up := ct - gap - h; up >= top || cb+gap+h > bottom {
		return x, max(0, up)
	}
	return x, cb + gap
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
		case "left": // vertically centred in the safe area, at the left margin
			return margin, max(top, top+(bottom-top-h)/2)
		case "right":
			return max(0, right-w), max(top, top+(bottom-top-h)/2)
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
	case "left": // vertically centred, on the left (the empty side of the frame, say)
		return int(0.0375 * float64(W)), (H - h) / 2
	case "right":
		return W - int(0.0375*float64(W)) - w, (H - h) / 2
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
var Layouts = []string{"lower-third", "title", "top", "corner", "center", "left", "right"}

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

func captionPlacement(tl *timeline.Timeline, tr timeline.Transcripts, o timeline.Overlay, projDir, cacheDir string, w, h int) (Placement, error) {
	ws, missing := captions.OutputWords(tl, tr)
	if len(missing) > 0 {
		return Placement{}, fmt.Errorf("burned-in captions need a transcript of %s: run `cutroom transcribe %s`", missing[0], missing[0])
	}
	st := CaptionStyle(tl)
	l, err := resolve(st, projDir, w, h)
	if err != nil {
		return Placement{}, fmt.Errorf("caption style: %w", err)
	}
	p := Placement{ID: o.ID, Kind: "captions", Start: 0, End: tl.Duration(), Name: "burned-in captions", Key: "captions"}
	var keys []any
	boxSet := false
	fps := tl.Rate()
	nAnim := int(math.Round(st.AnimateTime * fps))
	add := func(g captions.Group, words []string, active int, a, b float64, animate bool) error {
		key := hash([]any{CardVersion, "caption", w, h, words, active, l.key})
		png := filepath.Join(cacheDir, "overlays", key+".png")
		capMu.Lock()
		r, known := capRects[key]
		capMu.Unlock()
		if _, err := os.Stat(png); err != nil || !known {
			var img *image.NRGBA
			img, r = captionCard(words, active, l, w, h)
			capMu.Lock()
			capRects[key] = r
			capMu.Unlock()
			if err != nil {
				if err := ensureDir(png); err != nil {
					return err
				}
				if err := raster.SavePNG(png, img); err != nil {
					return err
				}
			}
		}
		box := [4]float64{float64(r.Min.X) / float64(w), float64(r.Min.Y) / float64(h), float64(r.Max.X) / float64(w), float64(r.Max.Y) / float64(h)}
		f0 := math.Round(a * fps)
		if animate && st.Animate != "none" && nAnim > 0 && math.Round(b*fps)-f0 >= float64(nAnim+1) {
			// the entrance: nAnim frames drawn from the settled card, then the card itself
			var full *image.NRGBA
			for i := range nAnim {
				akey := hash([]any{key, "anim", st.Animate, i, nAnim})
				apng := filepath.Join(cacheDir, "overlays", akey+".png")
				if _, err := os.Stat(apng); err != nil {
					if full == nil {
						full, _ = captionCard(words, active, l, w, h)
					}
					sc, dx, dy, al := animAt(st.Animate, float64(i+1)/float64(nAnim+1))
					if err := ensureDir(apng); err != nil {
						return err
					}
					if err := raster.SavePNG(apng, animFrame(full, r, sc, dx, dy, al)); err != nil {
						return err
					}
				}
				fa, fb := (f0+float64(i))/fps, (f0+float64(i+1))/fps
				p.Groups = append(p.Groups, CapGroup{Start: fa, End: fb, Text: g.Text, PNG: apng, Key: akey, Box: box, Anim: true})
				keys = append(keys, []any{akey, fa, fb})
			}
			a = (f0 + float64(nAnim)) / fps
		}
		p.Groups = append(p.Groups, CapGroup{Start: a, End: b, Text: g.Text, PNG: png, Key: key, Box: box})
		keys = append(keys, []any{key, a, b})
		if !boxSet { // the placement's box covers every group's box
			p.Box, boxSet = box, true
		}
		p.Box = [4]float64{math.Min(p.Box[0], box[0]), math.Min(p.Box[1], box[1]), math.Max(p.Box[2], box[2]), math.Max(p.Box[3], box[3])}
		return nil
	}
	for _, g := range captions.Groups(ws, GroupOptions(st), tl.Duration()) {
		words := strings.Fields(g.Text)
		if (l.highlight == nil && l.hiBox == nil) || len(words) != g.Last-g.First+1 {
			if err := add(g, words, -1, g.Start, g.End, true); err != nil {
				return p, err
			}
			continue
		}
		// karaoke: one card per word, the word being said in the highlight colour
		for k := range words {
			a, b := g.Start, g.End
			if k > 0 {
				a = r3(ws[g.First+k].Start)
			}
			if k+1 < len(words) {
				b = r3(ws[g.First+k+1].Start)
			}
			if b <= a {
				continue
			}
			if err := add(g, words, k, a, b, k == 0); err != nil {
				return p, err
			}
		}
	}
	p.Key = hash(keys)
	return p, nil
}

func r3(f float64) float64 { return math.Round(f*1000) / 1000 }

// capRects remembers each caption card's box by key, so a card already on disk isn't drawn again
// just to learn where it is.
var (
	capMu    sync.Mutex
	capRects = map[string]image.Rectangle{}
)

// CaptionCard draws one caption group in cutroom's default caption style.
func CaptionCard(text string, w, h int) (*image.NRGBA, image.Rectangle) {
	l, _ := resolve(CaptionDefaults, "", w, h)
	return captionCard(strings.Fields(text), -1, l, w, h)
}

// captionCard draws one caption group: its words (wrapped to the style's width), the active one in
// the highlight colour, on the style's box; by default centred in the caption band.
func captionCard(words []string, active int, l look, w, h int) (*image.NRGBA, image.Rectangle) {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	right := w
	if h > w {
		_, _, right = SafeZone(w, h)
	}
	mw := l.width
	if mw <= 0 {
		// the safe width less a small margin and the box's own padding: wrap only when it must
		mw = right - int(0.06*float64(w)) - 2*(l.padX+l.outlineW+l.shadowOff)
	}
	words = append([]string(nil), words...)
	for i := range words {
		words[i] = l.text(words[i])
	}
	// wrap words into lines
	var lines [][]int
	cur, curS := []int{}, ""
	for i, wd := range words {
		try := strings.TrimSpace(curS + " " + wd)
		if len(cur) > 0 && raster.Measure(l.face, try) > mw {
			lines = append(lines, cur)
			cur, curS = nil, wd
		} else {
			curS = try
		}
		cur = append(cur, i)
	}
	if len(cur) > 0 {
		lines = append(lines, cur)
	}
	m := l.face.Metrics()
	lh := (m.Ascent + m.Descent).Round()
	gap := int(0.12 * l.size)
	tw := 0
	widths := make([]int, len(lines))
	for i, ln := range lines {
		var ss []string
		for _, k := range ln {
			ss = append(ss, words[k])
		}
		widths[i] = raster.Measure(l.face, strings.Join(ss, " "))
		tw = max(tw, widths[i])
	}
	ow := l.outlineW + l.shadowOff
	bw := tw + 2*l.padX + 2*ow
	bh := len(lines)*lh + (len(lines)-1)*gap + 2*l.padY + 2*ow
	dx := max(0, (right-bw)/2) // vertical: centred left of the right-hand buttons (L41)
	dy := int(captionBottom(w, h)) - bh
	r := l.boxAt(w, h, bw, bh, dx, dy)
	l.drawBox(img, r)
	y := r.Min.Y + l.padY + ow + m.Ascent.Round()
	for i, ln := range lines {
		x := r.Min.X + l.padX + ow
		switch l.align {
		case "center", "":
			x += (tw - widths[i]) / 2
		case "right":
			x += tw - widths[i]
		}
		var runs []raster.Run
		for j, k := range ln {
			c := l.color
			if k == active && l.highlight != nil {
				c = l.highlight
			}
			s := words[k]
			if j > 0 {
				runs = append(runs, raster.Run{S: " ", C: c})
			}
			runs = append(runs, raster.Run{S: s, C: c})
		}
		raster.StyledLine(img, l.face, x, y, runs, l.textStyle())
		if l.hiBox != nil { // a rounded box over the outline of the word being said, its letters on top
			pre := ""
			for j, k := range ln {
				if j > 0 {
					pre += " "
				}
				if k == active {
					wx := x + raster.Measure(l.face, pre)
					ww := raster.Measure(l.face, words[k])
					px, py := int(0.18*l.size), int(0.08*l.size)
					raster.FillRound(img, image.Rect(wx-px, y-m.Ascent.Round()-py, wx+ww+px, y+m.Descent.Round()+py), 0.22*l.size, l.hiBox)
					c := l.color
					if l.highlight != nil {
						c = l.highlight
					}
					raster.Text(img, l.face, wx, y, c, words[k])
					break
				}
				pre += words[k]
			}
		}
		y += lh + gap
	}
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

// animFrame is one frame of an entrance: the card's box r scaled by sc around its centre, moved by
// (dx, dy) × its size, at opacity al.
func animFrame(src *image.NRGBA, r image.Rectangle, sc, dx, dy, al float64) *image.NRGBA {
	b := src.Bounds()
	out := image.NewNRGBA(b)
	bw, bh := float64(r.Dx())*sc, float64(r.Dy())*sc
	cx := float64(r.Min.X+r.Max.X)/2 + dx*float64(r.Dx())
	cy := float64(r.Min.Y+r.Max.Y)/2 + dy*float64(r.Dy())
	d := image.Rect(int(math.Round(cx-bw/2)), int(math.Round(cy-bh/2)), int(math.Round(cx+bw/2)), int(math.Round(cy+bh/2)))
	xdraw.ApproxBiLinear.Scale(out, d, src, r, draw.Src, nil)
	if al < 1 {
		d = d.Intersect(b)
		for y := d.Min.Y; y < d.Max.Y; y++ {
			for x := d.Min.X; x < d.Max.X; x++ {
				i := out.PixOffset(x, y) + 3
				out.Pix[i] = uint8(math.Round(float64(out.Pix[i]) * al))
			}
		}
	}
	return out
}

// motion is an animation that moves or scales a card ("" for none or a plain fade).
func motion(kind string) string {
	if kind == "none" || kind == "fade" {
		return ""
	}
	return kind
}
