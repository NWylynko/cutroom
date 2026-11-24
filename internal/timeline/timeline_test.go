package timeline

import (
	"math"
	"strings"
	"testing"

	"github.com/cutroom/cutroom/internal/speech"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

// L01/L02: every edge sits on the frame grid anchored at the audio start.
func TestL01_EdgesOnFrameGrid(t *testing.T) {
	tl := New("30", 1280, 720, "cam", "cam", 0.0637, 30.26)
	tl.Cut("cam", 6.31, 9.22, "x", true)
	for _, g := range tl.Segments {
		for _, e := range []float64{g.In, g.Out} {
			k := (e - 0.0637) * 30
			if math.Abs(k-math.Round(k)) > 1e-4 {
				t.Fatalf("edge %.6f is off the grid (k=%.4f)", e, k)
			}
		}
	}
	// output duration is computed from whole frames
	var frames int64
	for _, g := range tl.Segments {
		frames += tl.Frames(g)
	}
	if !near(tl.Duration(), float64(frames)/30) {
		t.Fatalf("duration %.6f vs frames %d", tl.Duration(), frames)
	}
}

func TestCutSplitsAndMaps(t *testing.T) {
	tl := New("30", 1280, 720, "cam", "cam", 0, 30)
	tl.Cut("cam", 10, 12, "x", true)
	if len(tl.Segments) != 2 || !near(tl.Duration(), 28) {
		t.Fatalf("%+v %.3f", tl.Segments, tl.Duration())
	}
	if o, ok := tl.ToOutput("cam", 13); !ok || !near(o, 11) {
		t.Fatalf("ToOutput(13) = %v %v", o, ok)
	}
	if _, ok := tl.ToOutput("cam", 11); ok {
		t.Fatal("11 s was cut")
	}
	if g, s, ok := tl.ToSource(11); !ok || g.ID != tl.Segments[1].ID || !near(s, 13) {
		t.Fatalf("ToSource(11) = %v %v %v", g, s, ok)
	}
	if !tl.Segments[0].PinOut || !tl.Segments[1].PinIn {
		t.Fatal("explicit cut edges must be pinned")
	}
}

func words(spec ...float64) []speech.Word {
	var ws []speech.Word
	for i := 0; i+1 < len(spec); i += 2 {
		ws = append(ws, speech.Word{I: len(ws), Text: "w", Start: spec[i], End: spec[i+1]})
	}
	return ws
}

// L39: tighten never moves an edge you cut explicitly.
func TestL39_TightenKeepsPinnedEdges(t *testing.T) {
	ws := words(1, 2, 2.1, 3, 6, 7, 7.1, 8)
	tr := Transcripts{"cam": {Words: ws}}
	tl := New("30", 1280, 720, "cam", "cam", 0, 10)
	tl.Cut("cam", 3.5, 5.0, "explicit", true) // leaves 0.5 s after w2 and 1.0 s before w3, pinned
	r := tl.Tighten(tr, 0.4)
	if !near(tl.Segments[0].Out, 3.5) || !near(tl.Segments[1].In, 5.0) {
		t.Fatalf("pinned edges moved: %+v (result %+v)", tl.Segments, r)
	}
	if !near(tl.Segments[0].In, 0.9) || !near(tl.Segments[1].Out, 8.233333) {
		t.Fatalf("unpinned outer edges should be trimmed to the pads: %+v", tl.Segments)
	}
}

// L32/L38: tighten leaves pad-after + pad-before in every shortened pause.
func TestL38_TightenLeavesPads(t *testing.T) {
	ws := words(0.5, 1.5, 4.0, 5.0)
	tr := Transcripts{"cam": {Words: ws}}
	tl := New("30", 1280, 720, "cam", "cam", 0, 6)
	tl.Tighten(tr, 0.4)
	if got := tl.KeptBetween("cam", 1.5, 4.0); math.Abs(got-0.32) > 1.0/30 {
		t.Fatalf("kept pause %.3f, want ~0.32", got)
	}
}

func TestCutWordsJoinPauseAfterTighten(t *testing.T) {
	ws := words(0.5, 1.5, 4.0, 5.0, 8.0, 9.0)
	tr := Transcripts{"cam": {Words: ws}}
	tl := New("30", 1280, 720, "cam", "cam", 0, 10)
	tl.Tighten(tr, 0.4)
	r := tl.CutWords("cam", ws, 1, 1, "drop middle")
	if math.Abs(r.JoinPause-0.32) > 2.0/30 {
		t.Fatalf("join pause %.3f, want ~0.32 (what is actually kept)", r.JoinPause)
	}
}

func TestValidateFlagsEdgeInWordAndKeptRetake(t *testing.T) {
	ws := words(1, 2, 2.1, 3)
	tr := Transcripts{"cam": {Words: ws, Suspects: []speech.Suspect{{Kind: "repeat", Text: "x y", Related: []float64{1, 2.1}}}}}
	tl := New("30", 1280, 720, "cam", "cam", 0, 4)
	tl.Cut("cam", 2.5, 2.6, "bad", true)
	codes := map[string]bool{}
	for _, is := range tl.Validate(map[string]float64{"cam": 4}, tr) {
		codes[is.Code] = true
	}
	if !codes["edge-in-word"] || !codes["retake-kept"] {
		t.Fatalf("got %v", codes)
	}
}

func TestBrokenSentenceWarned(t *testing.T) {
	mk := func(spec []string) []speech.Word {
		var ws []speech.Word
		for i, x := range spec {
			ws = append(ws, speech.Word{I: i, Text: x, Start: float64(i), End: float64(i) + 0.8})
		}
		return ws
	}
	ws := mk([]string{"Hi,", "today", "I'm", "going", "to", "show", "you."})
	tl := New("30", 1280, 720, "cam", "cam", 0, 8)
	tl.Cut("cam", 2.9, 4.9, "x", true) // removes "going to"
	found := false
	for _, is := range tl.Validate(map[string]float64{"cam": 8}, Transcripts{"cam": {Words: ws}}) {
		if is.Code == "broken-sentence" {
			found = true
		}
	}
	if !found {
		t.Fatal("cut inside a sentence not flagged")
	}
	// cutting a whole sentence is fine
	ws2 := mk([]string{"One.", "Two", "three.", "Four."})
	tl2 := New("30", 1280, 720, "cam", "cam", 0, 5)
	tl2.Cut("cam", 0.9, 2.9, "x", true)
	for _, is := range tl2.Validate(map[string]float64{"cam": 5}, Transcripts{"cam": {Words: ws2}}) {
		if is.Code == "broken-sentence" {
			t.Fatalf("whole-sentence cut flagged: %+v", is)
		}
	}
}

// L40: a punch-in splits segments at its edges, pins those edges (they sit inside continuous
// speech: tighten must not open a gap there), keeps the picture continuous, and zoom 1 undoes it.
func TestL40_ZoomSplitsPinsAndStaysContinuous(t *testing.T) {
	ws := words(1, 2, 2.1, 3, 6, 7, 7.1, 8)
	tr := Transcripts{"cam": {Words: ws}}
	tl := New("30", 1280, 720, "cam", "cam", 0, 10)
	before := tl.Duration()
	if n := tl.Zoom("cam", 2.0, 2.5, 1.2, 0.5, 0.3); n != 1 || len(tl.Segments) != 3 {
		t.Fatalf("want 1 zoomed of 3 segments, got %d of %+v", n, tl.Segments)
	}
	m := tl.Segments[1]
	if m.ZoomOf() != 1.2 || !m.PinIn || !m.PinOut || !tl.Segments[0].PinOut || !tl.Segments[2].PinIn {
		t.Fatalf("zoomed piece or its neighbours not pinned: %+v", tl.Segments)
	}
	if !near(tl.Segments[0].Out, m.In) || !near(m.Out, tl.Segments[2].In) || !near(tl.Duration(), before) {
		t.Fatalf("zoom must not change the picture's continuity or length: %+v", tl.Segments)
	}
	tl.Tighten(tr, 0.4)
	for i := 0; i+1 < len(tl.Segments); i++ {
		l, r := tl.Segments[i], tl.Segments[i+1]
		if l.ZoomOf() != r.ZoomOf() && !near(l.Out, r.In) {
			t.Fatalf("tighten opened a gap at a zoom edge: %+v", tl.Segments)
		}
	}
	js := tl.Joins(tr)
	kinds := map[string]int{}
	for _, j := range js {
		kinds[j.Kind]++
	}
	if kinds["zoom"] != 2 {
		t.Fatalf("want 2 zoom joins, got %+v", js)
	}
	tl.Zoom("cam", 0, 10, 1, 0, 0)
	for _, g := range tl.Segments {
		if g.ZoomOf() != 1 {
			t.Fatalf("zoom 1 must clear the punch-in: %+v", g)
		}
	}
}

// L36 on real speech (T3, securing-personal-software take 1): a retake repeats several phrases of a
// sentence, not its start. One warning per retake, cutting the whole first attempt at sentence
// boundaries — never into words that exist only in the next attempt ("My solution is…").
func TestL36_RetakeCutAtSentences(t *testing.T) {
	text := "Instead of making sure every endpoint is secure, I'm going to wrap the website in security. " +
		"My solution is to hug the software, instead of making sure every endpoint is secure, I'm going to wrap the website in security. Done."
	var ws []speech.Word
	at := 1.0
	for _, w := range strings.Fields(text) {
		ws = append(ws, speech.Word{I: len(ws), Text: w, Start: at, End: at + 0.25})
		at += 0.3
		if strings.HasSuffix(w, ".") {
			at += 0.6
		}
	}
	find := func(from int, phrase string) float64 {
		ms := speech.Search(ws, phrase)
		for _, m := range ms {
			if m.First >= from {
				return m.Start
			}
		}
		t.Fatalf("no %q", phrase)
		return 0
	}
	p1, p2 := "instead of making sure every endpoint", "to wrap the website in security"
	sus := []speech.Suspect{
		{Kind: "repeat", Text: p1, Related: []float64{find(0, p1), find(10, p1)}},
		{Kind: "repeat", Text: p2, Related: []float64{find(0, p2), find(20, p2)}},
	}
	tr := Transcripts{"cam": {Words: ws, Suspects: sus}}
	tl := New("30", 1280, 720, "cam", "cam", 0, at+1)
	var rk []Issue
	for _, is := range tl.Validate(map[string]float64{"cam": at + 1}, tr) {
		if is.Code == "retake-kept" {
			rk = append(rk, is)
		}
	}
	if len(rk) != 1 {
		t.Fatalf("want one retake warning for one retake, got %d: %+v", len(rk), rk)
	}
	f := rk[0].Fix
	if !strings.Contains(f, `--from "Instead of making sure every"`) || !strings.Contains(f, `"Instead of making sure every endpoint is secure, I'm going to wrap the website in security." (`) {
		t.Fatalf("cut should run from the first attempt's sentence start to its end: %s", f)
	}
	if strings.Contains(f, "My solution") {
		t.Fatalf("cut must not remove words only the next attempt has: %s", f)
	}
}
