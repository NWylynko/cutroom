package captions

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestL48_CaptionGroups(t *testing.T) {
	ws := words("Hi, today I'm going to show you how to set up a home server. The first thing you need is an old laptop. Then install Linux on it.")
	gs := Groups(ws, DefaultGroups, 100)
	var all []string
	for k, g := range gs {
		f := strings.Fields(g.Text)
		if utf8.RuneCountInString(g.Text) > 24 || len(f) > 4 {
			t.Errorf("group %d too long: %q", k, g.Text)
		}
		if weak(f[len(f)-1]) && k+1 < len(gs) {
			t.Errorf("group %d ends on a weak word: %q", k, g.Text)
		}
		if len(f) == 1 && utf8.RuneCountInString(f[0]) <= 3 && k > 0 {
			t.Errorf("stranded short word: %q (groups %v)", g.Text, gtexts(gs))
		}
		if k > 0 && g.Start < gs[k-1].End {
			t.Errorf("group %d overlaps", k)
		}
		all = append(all, g.Text)
	}
	// every word appears once, in order (punctuation at group ends dropped)
	if got := strings.Join(all, " "); strings.ReplaceAll(strings.ReplaceAll(got, ",", ""), ".", "") !=
		strings.ReplaceAll(strings.ReplaceAll(join(ws), ",", ""), ".", "") {
		t.Errorf("words lost or reordered:\n%s", got)
	}
	// sentence ends split groups; continuous speech has no gaps between groups
	for k := 1; k < len(gs); k++ {
		if gs[k].Start != gs[k-1].End {
			t.Errorf("flicker gap before group %d", k)
		}
	}
	if !strings.HasPrefix(gtexts(gs)[len(gs)-2]+" "+gtexts(gs)[len(gs)-1], "Then") && !strings.Contains(strings.Join(gtexts(gs), "|"), "|Then") {
		t.Errorf("sentence didn't start a group: %v", gtexts(gs))
	}
}

func TestGroupsPauseAndHold(t *testing.T) {
	ws := words("Hello there / general")
	gs := Groups(ws, DefaultGroups, 10)
	if len(gs) != 2 || gs[0].Text != "Hello there" {
		t.Fatalf("%v", gtexts(gs))
	}
	if gs[0].End != r3(ws[1].End+0.4) {
		t.Errorf("hold after a pause: %v", gs[0].End)
	}
}

func gtexts(gs []Group) []string {
	var o []string
	for _, g := range gs {
		o = append(o, g.Text)
	}
	return o
}
