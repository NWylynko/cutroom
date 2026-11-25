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
