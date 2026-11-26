package timeline

import "testing"

func TestStyleSetValidatesAndUnsets(t *testing.T) {
	var s Style
	for _, c := range [][2]string{{"color", "#ffd400"}, {"box", "none"}, {"size", "0.08"}, {"words", "2"}, {"case", "upper"}, {"y", "0"}} {
		if err := s.Set(c[0], c[1]); err != nil {
			t.Fatalf("%s=%s: %v", c[0], c[1], err)
		}
	}
	if s.Y == nil || *s.Y != 0 {
		t.Fatalf("y 0 must be kept (the top of the frame), got %v", s.Y)
	}
	for _, c := range [][2]string{{"color", "#zzz"}, {"size", "2"}, {"words", "0"}, {"case", "title"}, {"nope", "1"}} {
		if err := s.Set(c[0], c[1]); err == nil {
			t.Errorf("%s=%s accepted", c[0], c[1])
		}
	}
	if err := s.Set("color", "default"); err != nil || s.Color != "" {
		t.Errorf("default must unset color: %q %v", s.Color, err)
	}
}

func TestStyleLayers(t *testing.T) {
	base := Style{Font: "bold", Color: "#ffffff", Box: "#000000aa", Words: 3}
	proj := &Style{Color: "#ffd400", Box: "none"}
	got := proj.Over(base)
	if got.Color != "#ffd400" || got.Box != "none" || got.Font != "bold" || got.Words != 3 {
		t.Errorf("project over defaults: %+v", got)
	}
	var none *Style
	if none.Over(base) != base || !none.Empty() {
		t.Errorf("nil style must leave the base")
	}
}

func TestParseColor(t *testing.T) {
	for in, want := range map[string][4]uint8{"#fff": {255, 255, 255, 255}, "#00000099": {0, 0, 0, 0x99}, "Yellow": {250, 204, 21, 255}, "none": {}} {
		if c, err := ParseColor(in); err != nil || c != want {
			t.Errorf("%s: %v %v", in, c, err)
		}
	}
}
