package timeline

import (
	"strings"
	"testing"

	"github.com/cutroom/cutroom/internal/speech"
)

func sayWords(start float64, text string) []speech.Word {
	var ws []speech.Word
	for i, w := range strings.Fields(text) {
		ws = append(ws, speech.Word{I: i, Text: w, Start: start, End: start + 0.3})
		start += 0.35
	}
	return ws
}

// T3: an edit that ends where the recording stops mid-speech ("Those are much better." after "…at
// this point.") or opens on a sentence tail spoken as the camera starts is warned, with the cut.
func TestEdgesMidSpeech(t *testing.T) {
	tl := New("30", 1280, 720, "cam", "cam", 0, 10)
	ws := append(sayWords(0.05, "is going to serve you. Then we plan."), sayWords(6.5, "Not at this point. Those are much better.")...)
	tr := Transcripts{"cam": {Source: "cam", Words: ws}}
	var codes []string
	for _, is := range tl.Validate(map[string]float64{"cam": 9.4}, tr) {
		codes = append(codes, is.Code+": "+is.Fix)
	}
	all := strings.Join(codes, "\n")
	if !strings.Contains(all, `ends-mid-speech: end on the sentence before it: `+"`"+`cutroom edit cut --from "Those are much" --to "much better."`) {
		t.Errorf("no ends-mid-speech cut:\n%s", all)
	}
	if !strings.Contains(all, `starts-mid-speech: start on the next sentence: `+"`"+`cutroom edit cut --from "is going to" --to "serve you."`) {
		t.Errorf("no starts-mid-speech cut:\n%s", all)
	}
	// with a pause before the recording ends, or a capitalised opening, nothing is said
	clean := Transcripts{"cam": {Source: "cam", Words: append(sayWords(0.6, "Hi, today we plan."), sayWords(6.5, "Not at this point.")...)}}
	for _, is := range tl.Validate(map[string]float64{"cam": 10}, clean) {
		if strings.HasSuffix(is.Code, "-mid-speech") {
			t.Errorf("false %s: %s", is.Code, is.Message)
		}
	}
}

// T3 r49 (sec): "access, you" is continuous speech; the recogniser's boundary sat 50–80 ms after the
// real gap, and a cut on it left 40 ms of "you". With the sound's level known, the cut goes to the
// quietest frame boundary in the gap, and the edge isn't reported as inside "access,".
func TestCutInContinuousSpeechAtQuietestFrame(t *testing.T) {
	tl := New("30", 1280, 720, "cam", "cam", 0, 10)
	ws := []speech.Word{
		{I: 0, Text: "with", Start: 1.0, End: 1.3},
		{I: 1, Text: "access,", Start: 1.3, End: 1.84},
		{I: 2, Text: "you", Start: 1.84, End: 2.04},
		{I: 3, Text: "can.", Start: 2.04, End: 2.3},
		{I: 4, Text: "So", Start: 3.0, End: 3.2},
	}
	gap := 1.767 // the real gap: quiet from 1.75 to 1.78
	tl.Quiet = func(src string, ts []float64) []float64 {
		db := make([]float64, len(ts))
		for i, x := range ts {
			db[i] = -50
			if x > 1.745 && x < 1.785 {
				db[i] = -68
			}
		}
		return db
	}
	r := tl.CutWords("cam", ws, 2, 3, "cut")
	if d := r.From - gap; d < -0.02 || d > 0.02 {
		t.Fatalf("cut starts at %.3f, want the gap at %.3f (the recogniser said 1.84)", r.From, gap)
	}
	for _, is := range tl.Validate(map[string]float64{"cam": 10}, Transcripts{"cam": {Source: "cam", Words: ws}}) {
		if is.Code == "edge-in-word" {
			t.Errorf("edge in the measured gap reported: %s", is.Message)
		}
	}
	// without levels, the cut stays on the recogniser's boundary
	tl2 := New("30", 1280, 720, "cam", "cam", 0, 10)
	if r := tl2.CutWords("cam", ws, 2, 3, "cut"); r.From < 1.83 {
		t.Fatalf("no levels: cut at %.3f, want the transcript boundary", r.From)
	}
}

// T3 r49 (sec): the recorder's "and" ran past the camera's last frame; mapped onto the camera it is
// kept (clipped at the end), so the edge check sees the piece of it the last segment holds.
func TestMappedWordCutOffByCameraEndIsKept(t *testing.T) {
	rec := &speech.Transcript{Source: "mic", Words: []speech.Word{
		{Text: "enabled", Start: 4.2, End: 5.0},
		{Text: "and", Start: 5.2, End: 5.5},
	}}
	m := rec.MapTo("cam", func(r float64) float64 { return r - 0.1 }, 5.25)
	if len(m.Words) != 2 || m.Words[1].End != 5.25 {
		t.Fatalf("mapped words %+v; want \"and\" kept and clipped at 5.25", m.Words)
	}
	tl := New("30", 1280, 720, "cam", "cam", 0, 5.2) // the edit ends 50 ms before the camera does
	var found bool
	for _, is := range tl.Validate(map[string]float64{"cam": 5.25}, Transcripts{"cam": m}) {
		if is.Code == "edge-in-word" && strings.Contains(is.Fix, "cutroom edit cut 5.080-") {
			found = true
		}
	}
	if !found {
		t.Fatal("no edge-in-word with a runnable cut for the cut-off word")
	}
}
