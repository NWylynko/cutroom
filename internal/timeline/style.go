package timeline

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Style is the look of burned-in captions or text cards. Every field is optional: an empty value
// means "cutroom's default", so a project only records what the user chose. Styles layer: cutroom's
// defaults ← the project's caption or card style ← one overlay's own style.
//
// Sizes are fractions of the frame's short side (the same look on landscape and vertical); widths,
// paddings and radii are fractions of the main font size; X/Y place the centre of the box as a
// fraction of the frame (0–1 across, down), like `edit frame --anchor`.
type Style struct {
	Font         string   `json:"font,omitempty"`          // bold | regular | mono | a font file | an installed font's name
	SubFont      string   `json:"sub_font,omitempty"`      // cards: the second line's font
	Size         float64  `json:"size,omitempty"`          // main text height, fraction of the short side
	SubSize      float64  `json:"sub_size,omitempty"`      // cards: second line
	Color        string   `json:"color,omitempty"`         // text colour
	SubColor     string   `json:"sub_color,omitempty"`     // cards: second line colour
	Highlight    string   `json:"highlight,omitempty"`     // captions: colour of the word being said ("none" = off)
	HighlightBox string   `json:"highlight_box,omitempty"` // captions: a rounded box behind the word being said ("none" = off)
	Box          string   `json:"box,omitempty"`           // background colour, or none
	Accent       string   `json:"accent,omitempty"`        // cards: bar on the left of lower thirds, or none
	Outline      string   `json:"outline,omitempty"`       // text outline colour, or none
	OutlineWidth float64  `json:"outline_width,omitempty"` // fraction of the font size
	Shadow       string   `json:"shadow,omitempty"`        // drop-shadow colour under the letters, or none
	BoxShadow    string   `json:"box_shadow,omitempty"`    // soft shadow under the box, or none
	Radius       float64  `json:"radius,omitempty"`        // box corner radius, fraction of the font size
	Pad          float64  `json:"pad,omitempty"`           // box padding, fraction of the font size
	Case         string   `json:"case,omitempty"`          // upper | upper-keep-names | lower | as-is
	Align        string   `json:"align,omitempty"`         // left | center | right (lines inside the box)
	X            *float64 `json:"x,omitempty"`             // centre of the box, 0–1 across
	Y            *float64 `json:"y,omitempty"`             // centre of the box, 0–1 down
	Width        float64  `json:"width,omitempty"`         // widest line, fraction of the frame width
	Animate      string   `json:"animate,omitempty"`       // how it comes in and goes out: none | fade | pop | slide-up | slide-down | slide-left | slide-right
	AnimateTime  float64  `json:"animate_time,omitempty"`  // seconds the animation takes
	Words        int      `json:"words,omitempty"`         // captions: most words shown at once
	Chars        int      `json:"chars,omitempty"`         // captions: most characters shown at once
}

// StyleField describes one settable field: its flag name, kind and help.
type StyleField struct {
	Name, Kind, Help, For string // For: "captions", "cards" or "" (both)
}

// StyleFields lists every field in flag order (the `style` command's flags and its printout).
var StyleFields = []StyleField{
	{"font", "font", "bold | regular | mono | a .ttf/.otf/.ttc file | an installed font's name (e.g. \"Arial Black\"); name#2 picks a face in a .ttc", ""},
	{"size", "num", "text height as a fraction of the frame's short side (captions default 0.068)", ""},
	{"color", "color", "text colour: #rrggbb, #rrggbbaa or a name (white, black, yellow, …)", ""},
	{"highlight", "color", "colour of the word being said, karaoke style (none = off)", "captions"},
	{"highlight-box", "color", "a rounded box in this colour behind the word being said (none = off)", "captions"},
	{"box", "color", "background box colour (with alpha: #00000099), or none", ""},
	{"outline", "color", "outline round the letters, or none", ""},
	{"outline-width", "num", "outline thickness as a fraction of the font size (default 0.08 when an outline is set)", ""},
	{"shadow", "color", "drop shadow under the letters, or none", ""},
	{"box-shadow", "color", "soft shadow under the box (e.g. #00000066), or none", ""},
	{"radius", "num", "box corner radius as a fraction of the font size (0 = square)", ""},
	{"pad", "num", "box padding as a fraction of the font size", ""},
	{"case", "enum:upper,upper-keep-names,lower,as-is", "letter case (upper-keep-names: capitals, but iCloud, OAuth, GitHub stay as written)", ""},
	{"align", "enum:left,center,right", "line alignment inside the box", ""},
	{"x", "num", "centre of the box, 0–1 across (default: centred; cards: from --layout)", ""},
	{"y", "num", "centre of the box, 0–1 down (default: the caption band / the layout's place)", ""},
	{"width", "num", "widest line as a fraction of the frame width (longer text wraps)", ""},
	{"animate", "enum:none,fade,pop,slide-up,slide-down,slide-left,slide-right", "how it comes in (and cards go out): captions animate each group as it appears; cards also keep their --fade", ""},
	{"animate-time", "num", "seconds the animation takes (default 0.15 captions, 0.3 cards)", ""},
	{"words", "int", "most words shown at once (default 3)", "captions"},
	{"chars", "int", "most characters shown at once (default 20)", "captions"},
	{"sub-font", "font", "the second line's font", "cards"},
	{"sub-size", "num", "the second line's size", "cards"},
	{"sub-color", "color", "the second line's colour", "cards"},
	{"accent", "color", "the bar on the left of a lower third, or none", "cards"},
}

func (f StyleField) goName() string {
	parts := strings.Split(f.Name, "-")
	for i, p := range parts {
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

func (s *Style) field(name string) reflect.Value {
	return reflect.ValueOf(s).Elem().FieldByName(StyleField{Name: name}.goName())
}

// Set sets one field from its flag text; "default" clears it. It checks the value.
func (s *Style) Set(name, val string) error {
	var fd StyleField
	for _, f := range StyleFields {
		if f.Name == name {
			fd = f
		}
	}
	if fd.Name == "" {
		return fmt.Errorf("no style field %q", name)
	}
	v := s.field(name)
	if val == "default" || val == "" {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	switch {
	case fd.Kind == "color":
		if _, err := ParseColor(val); err != nil {
			return fmt.Errorf("--%s: %w", name, err)
		}
		v.SetString(strings.ToLower(val))
	case fd.Kind == "font":
		v.SetString(val)
	case strings.HasPrefix(fd.Kind, "enum:"):
		opts := strings.Split(strings.TrimPrefix(fd.Kind, "enum:"), ",")
		ok := false
		for _, o := range opts {
			ok = ok || o == val
		}
		if !ok {
			return fmt.Errorf("--%s takes %s", name, strings.Join(opts, ", "))
		}
		v.SetString(val)
	case fd.Kind == "int":
		n, err := strconv.Atoi(val)
		if err != nil || n < 1 || n > 40 {
			return fmt.Errorf("--%s takes a whole number 1–40", name)
		}
		v.SetInt(int64(n))
	case fd.Kind == "num":
		f, err := strconv.ParseFloat(val, 64)
		lo, hi := 0.0, 1.0
		switch name {
		case "outline-width", "radius", "pad":
			hi = 3
		}
		if err != nil || f < lo || f > hi {
			return fmt.Errorf("--%s takes a number %g–%g", name, lo, hi)
		}
		if v.Kind() == reflect.Pointer {
			v.Set(reflect.ValueOf(&f))
		} else {
			v.SetFloat(f)
		}
	}
	return nil
}

// Get is a field's value as text ("" when unset).
func (s *Style) Get(name string) string {
	if s == nil {
		return ""
	}
	v := s.field(name)
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int:
		if v.Int() == 0 {
			return ""
		}
		return strconv.Itoa(int(v.Int()))
	case reflect.Float64:
		if v.Float() == 0 {
			return ""
		}
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Pointer:
		if v.IsNil() {
			return ""
		}
		return strconv.FormatFloat(v.Elem().Float(), 'f', -1, 64)
	}
	return ""
}

// Empty reports whether nothing is set.
func (s *Style) Empty() bool {
	if s == nil {
		return true
	}
	for _, f := range StyleFields {
		if s.Get(f.Name) != "" {
			return false
		}
	}
	return true
}

// Over returns s with every field it leaves unset taken from base (s wins).
func (s *Style) Over(base Style) Style {
	out := base
	if s == nil {
		return out
	}
	for _, f := range StyleFields {
		if v := s.field(f.Name); !v.IsZero() {
			out.field(f.Name).Set(v)
		}
	}
	return out
}

var namedColors = map[string][4]uint8{
	"white": {255, 255, 255, 255}, "black": {0, 0, 0, 255}, "red": {239, 68, 68, 255},
	"green": {34, 197, 94, 255}, "blue": {59, 130, 246, 255}, "yellow": {250, 204, 21, 255},
	"orange": {249, 115, 22, 255}, "purple": {168, 85, 247, 255}, "pink": {236, 72, 153, 255},
	"cyan": {34, 211, 238, 255}, "grey": {150, 150, 150, 255}, "gray": {150, 150, 150, 255},
	"transparent": {0, 0, 0, 0}, "none": {0, 0, 0, 0},
}

// ParseColor reads #rgb, #rrggbb, #rrggbbaa or a colour name into RGBA (not premultiplied).
func ParseColor(s string) ([4]uint8, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if c, ok := namedColors[s]; ok {
		return c, nil
	}
	h := strings.TrimPrefix(s, "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	if len(h) == 6 {
		h += "ff"
	}
	if len(h) != 8 {
		return [4]uint8{}, fmt.Errorf("%q is not a colour: use #rrggbb, #rrggbbaa (aa = opacity) or a name like white", s)
	}
	n, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return [4]uint8{}, fmt.Errorf("%q is not a colour: use #rrggbb, #rrggbbaa (aa = opacity) or a name like white", s)
	}
	return [4]uint8{uint8(n >> 24), uint8(n >> 16), uint8(n >> 8), uint8(n)}, nil
}
