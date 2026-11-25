package captions

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/NWylynko/cutroom/internal/speech"
	"github.com/NWylynko/cutroom/internal/timeline"
)

// words makes evenly spaced words (0.3 s each, 0.05 s apart) from a sentence; "|" marks a cut
// before the next word, "/" a 1 s pause.
func words(s string) []Word {
	var ws []Word
	t, cut := 0.0, false
	for _, f := range strings.Fields(s) {
		switch f {
		case "|":
			cut = true
			continue
		case "/":
			t += 1
			continue
		}
		ws = append(ws, Word{Text: f, Start: t, End: t + 0.3, AfterCut: cut, Segment: "s1"})
		t += 0.35
		cut = false
	}
	return ws
}

func status(is []Issue, check string) []Issue {
	var o []Issue
	for _, i := range is {
		if i.Check == check && i.Status != "pass" {
			o = append(o, i)
		}
	}
	return o
}

func TestL48_CuesReadLikeAPersonMadeThem(t *testing.T) {
	ws := words("Hi, today I'm going to show you how to set up a home server on an old laptop that you have lying around in a drawer. The first thing you need is an old laptop. Then install Linux on it.")
	cues := Build(ws, Default, 100)
	for k, c := range cues {
		if len(c.Lines) > 2 {
			t.Errorf("cue %d has %d lines", k+1, len(c.Lines))
		}
		for _, l := range c.Lines {
			if utf8.RuneCountInString(l) > 42 {
				t.Errorf("cue %d line %q > 42", k+1, l)
			}
			f := strings.Fields(l)
			if weak(f[len(f)-1]) {
				t.Errorf("cue %d line ends on a weak word: %q", k+1, l)
			}
		}
		if k > 0 && c.Start < cues[k-1].End {
			t.Errorf("cue %d overlaps", k+1)
		}
	}
	// each sentence starts a cue
	starts := map[string]bool{}
	for _, c := range cues {
		starts[strings.Fields(c.Text())[0]] = true
	}
	if !starts["The"] || !starts["Then"] {
		t.Errorf("sentences don't start cues: %v", texts(cues))
	}
	if is := Check(cues, ws, Default, 100); len(status(is, "reading"))+len(status(is, "seams"))+len(status(is, "words"))+len(status(is, "timing")) > 0 {
		t.Errorf("own cues fail checks: %+v", is)
	}
}

func TestL49_NoStrandedSentenceTail(t *testing.T) {
	// 84+ chars before "server." forces a break right before the last word of the sentence
	ws := words("Today we are going to configure the small computer under my desk into a really reliable home server. Next step.")
	cues := Build(ws, Default, 100)
	if len(cues) < 3 {
		t.Fatalf("the sentence should need two cues: %v", texts(cues))
	}
	for k := 1; k < len(cues); k++ {
		f := strings.Fields(cues[k].Text())
		prevEnds := endsSentence(cues[k-1].Lines[len(cues[k-1].Lines)-1])
		if !prevEnds && (endsSentence(f[0]) || (len(f) > 1 && endsSentence(f[1]))) {
			t.Errorf("cue %d opens on a stranded tail: %v", k+1, texts(cues))
		}
	}
	// the check catches a hand-made file with the tail stranded
	bad := []Cue{{Start: 0, End: 4, Lines: []string{"Today we are going to configure the small", "computer under my desk into a really reliable home"}},
		{Start: 4.1, End: 5, Lines: []string{"server. Next step."}}}
	if s := status(Check(bad, ws, Default, 100), "seams"); len(s) != 1 || !strings.Contains(s[0].Detail, `stranded sentence tail "server."`) {
		t.Errorf("stranded tail not caught: %+v", s)
	}
}

func TestL49_SeamsAtCuts(t *testing.T) {
	// "you need is | Then install": an unfinished sentence before a cut
	ws := words("The first thing you need is | Then install Linux on it. And that's it.")
	if s := status(Check(Build(ws, Default, 100), ws, Default, 100), "seams"); len(s) != 1 || !strings.Contains(s[0].Detail, "stops without ending") {
		t.Errorf("unended sentence not caught: %+v", s)
	}
	// "laptop. | install Linux": the next sentence's start was cut
	ws = words("You need an old laptop. | install Linux on it.")
	if s := status(Check(Build(ws, Default, 100), ws, Default, 100), "seams"); len(s) != 1 || !strings.Contains(s[0].Detail, "opens mid-sentence") {
		t.Errorf("headless sentence not caught: %+v", s)
	}
	// a cut inside a sentence that still reads (filler/aside removed): fine
	ws = words("The first thing you need is | an old laptop.")
	if s := status(Check(Build(ws, Default, 100), ws, Default, 100), "seams"); len(s) != 0 {
		t.Errorf("clean mid-sentence join flagged: %+v", s)
	}
}

func TestPauseStartsCueAndTimingHolds(t *testing.T) {
	ws := words("Hello there / general")
	cues := Build(ws, Default, 3)
	if len(cues) != 2 {
		t.Fatalf("pause didn't split: %v", texts(cues))
	}
	if cues[0].End != r3(max(ws[1].End+Default.Hold, ws[0].Start+Default.MinDur)) {
		t.Errorf("hold: %v", cues[0].End)
	}
	if cues[1].End > 3 {
		t.Errorf("cue past the video: %v", cues[1].End)
	}
}

func TestSRTAndVTTRoundTrip(t *testing.T) {
	ws := words("Hi, today I'm going to show you how to set up a home server. The first thing you need is an old laptop.")
	cues := Build(ws, Default, 100)
	for _, vtt := range []bool{false, true} {
		back, err := Parse(Format(cues, vtt))
		if err != nil {
			t.Fatal(err)
		}
		if len(back) != len(cues) {
			t.Fatalf("vtt=%v: %d cues back, want %d", vtt, len(back), len(cues))
		}
		for k := range cues {
			if back[k].Start != cues[k].Start || back[k].End != cues[k].End || back[k].Text() != cues[k].Text() {
				t.Errorf("vtt=%v cue %d: %+v vs %+v", vtt, k, back[k], cues[k])
			}
		}
		if s := status(Check(back, ws, Default, 100), "words"); len(s) != 0 {
			t.Errorf("words differ after round trip: %+v", s)
		}
	}
	if Stamp(3725.5, false) != "01:02:05,500" || Stamp(0.04, true) != "00:00:00.040" {
		t.Errorf("stamps %s %s", Stamp(3725.5, false), Stamp(0.04, true))
	}
	// a stale file (a sentence the edit no longer has) fails the words check
	stale := append([]Cue{{Start: 0, End: 1, Lines: []string{"By the way, my cat just walked across the keyboard."}}}, cues...)
	for k := range stale[1:] {
		stale[k+1].Start += 1
		stale[k+1].End += 1
	}
	if s := status(Check(stale, ws, Default, 100), "words"); len(s) != 1 || s[0].Status != "fail" {
		t.Errorf("stale file not failed: %+v", s)
	}
}

func TestOutputWordsDropFillersAndMarkCuts(t *testing.T) {
	tl := timeline.New("30", 1280, 720, "v", "v", 0, 10)
	tl.Cut("v", 2.0, 4.0, "x", true)
	tr := timeline.Transcripts{"v": {Source: "v", Words: []speech.Word{
		{Text: "The", Start: 0.1, End: 0.3}, {Text: "thing", Start: 0.4, End: 0.7}, {Text: "is,", Start: 0.8, End: 1.0},
		{Text: "UM,", Start: 1.1, End: 1.4}, {Text: "gone.", Start: 2.5, End: 3}, {Text: "an", Start: 4.2, End: 4.4},
		{Text: "old", Start: 4.5, End: 4.8}, {Text: "laptop.", Start: 4.9, End: 5.4}}}}
	ws, missing := OutputWords(tl, tr)
	if len(missing) != 0 {
		t.Fatal(missing)
	}
	var got []string
	for _, w := range ws {
		got = append(got, w.Text)
	}
	if strings.Join(got, " ") != "The thing is, an old laptop." {
		t.Errorf("words %q", got)
	}
	if !ws[3].AfterCut || ws[2].AfterCut || ws[3].Start != 2.2 {
		t.Errorf("cut marks/times: %+v", ws)
	}
}

func texts(cs []Cue) []string {
	var o []string
	for _, c := range cs {
		o = append(o, c.Text())
	}
	return o
}
