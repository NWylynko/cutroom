package speech

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cutroom/cutroom/internal/deps"
)

func TestParseWhisperMergesTokensAndPunctuation(t *testing.T) {
	js := `{"transcription":[{"tokens":[
	 {"text":"[_BEG_]","offsets":{"from":0,"to":0},"p":1,"t_dtw":-1},
	 {"text":" Hi","offsets":{"from":0,"to":180},"p":0.9,"t_dtw":82},
	 {"text":",","offsets":{"from":180,"to":360},"p":0.9,"t_dtw":140},
	 {"text":" I","offsets":{"from":810,"to":890},"p":0.8,"t_dtw":166},
	 {"text":"'m","offsets":{"from":890,"to":960},"p":0.6,"t_dtw":172},
	 {"text":" server","offsets":{"from":3690,"to":4100},"p":0.9,"t_dtw":388},
	 {"text":".","offsets":{"from":6340,"to":6380},"p":0.9,"t_dtw":566}]}]}`
	ws, err := parseWhisper([]byte(js))
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, w := range ws {
		got = append(got, w.Text)
	}
	if strings.Join(got, "|") != "Hi,|I'm|server." {
		t.Fatalf("words %v", got)
	}
	// punctuation must not stretch "server." over the following pause
	if ws[2].End != 4.1 || ws[2].Start != 3.88 || !ws[2].DTW {
		t.Fatalf("server timing %+v", ws[2])
	}
}

func TestParseWhisperDropsAnnotations(t *testing.T) {
	js := `{"transcription":[{"tokens":[
	 {"text":" ready","offsets":{"from":0,"to":100},"p":0.9,"t_dtw":5},
	 {"text":".","offsets":{"from":100,"to":110},"p":0.9,"t_dtw":10},
	 {"text":" [","offsets":{"from":200,"to":210},"p":0.9,"t_dtw":20},
	 {"text":"BL","offsets":{"from":210,"to":220},"p":0.9,"t_dtw":21},
	 {"text":"ANK_AUDIO]","offsets":{"from":220,"to":230},"p":0.9,"t_dtw":22},
	 {"text":" (music)","offsets":{"from":300,"to":310},"p":0.9,"t_dtw":30},
	 {"text":" Yes","offsets":{"from":400,"to":500},"p":0.9,"t_dtw":40}]}]}`
	ws, _ := parseWhisper([]byte(js))
	var got []string
	for _, w := range ws {
		got = append(got, w.Text)
	}
	if strings.Join(got, "|") != "ready.|Yes" {
		t.Fatalf("annotations leaked: %v", got)
	}
}

func TestL28_RefineStopsWordsAtPauses(t *testing.T) {
	regions := []Region{{Start: 0.6, End: 4.2}, {Start: 6.3, End: 9.2}}
	raw := []RawWord{{Text: "Hi", Start: 0.82}, {Text: "server.", Start: 3.88, End: 6.3}, {Text: "The", Start: 6.38}, {Text: "end.", Start: 8.9}}
	ws := refine(raw, regions)
	if ws[0].Start != 0.6 {
		t.Errorf("first word of a region should start at the region start, got %.2f", ws[0].Start)
	}
	if ws[1].End != 4.2 {
		t.Errorf("last word of a region must end at the region end, not stretch over the pause: %.2f", ws[1].End)
	}
	if ws[2].Start != 6.3 || ws[3].End != 9.2 {
		t.Errorf("second region words %+v %+v", ws[2], ws[3])
	}
}

func TestRepeatsFindRetake(t *testing.T) {
	var ws []Word
	t0 := 0.0
	for _, s := range strings.Fields("the first thing you need is um the first thing the first thing you need is an old laptop") {
		ws = append(ws, Word{Text: s, Start: t0, End: t0 + 0.3})
		t0 += 0.35
	}
	sus := repeats(ws)
	if len(sus) != 1 || sus[0].Text != "the first thing you need is" || len(sus[0].Related) != 2 {
		t.Fatalf("%+v", sus)
	}
}

// A long retake (> 8 words repeated) is seen through several overlapping n-gram windows; it must be
// reported once, spanning the whole repeated phrase (found on real footage, T3).
func TestL36_LongRetakeReportedOnce(t *testing.T) {
	var ws []Word
	t0 := 0.0
	take := "there is no point replacing a five a month subscription with a ten a month server"
	for _, s := range strings.Fields(take + " so " + take + " instead") {
		ws = append(ws, Word{Text: s, Start: t0, End: t0 + 0.3})
		t0 += 0.35
	}
	sus := repeats(ws)
	if len(sus) != 1 {
		t.Fatalf("want 1 repeat, got %d: %+v", len(sus), sus)
	}
	if sus[0].Text != take || len(sus[0].Related) != 2 {
		t.Fatalf("want the whole take %q ×2, got %q ×%d", take, sus[0].Text, len(sus[0].Related))
	}
}

func TestDiffPassesFindsStumble(t *testing.T) {
	mk := func(s string) []Word {
		var ws []Word
		for i, f := range strings.Fields(s) {
			ws = append(ws, Word{Text: f, Start: float64(i), End: float64(i) + 0.5})
		}
		return ws
	}
	d := DiffPasses(mk("ask your AI agent how to do it"), mk("ask your agent ask your AI agent how to do it"))
	if len(d) != 1 || !strings.Contains(d[0].Detail, "only the disfluency pass") {
		t.Fatalf("%+v", d)
	}
}

func TestSearchMatchesFillerWhenQuoted(t *testing.T) {
	ws := []Word{{Text: "The"}, {Text: "first"}, {Text: "is,"}, {Text: "UM,"}, {Text: "the"}, {Text: "first"}, {Text: "is"}, {Text: "an"}}
	if m := Search(ws, "first is, um"); len(m) != 1 || m[0].First != 1 || m[0].Last != 3 {
		t.Fatalf("quoted filler must disambiguate: %+v", m)
	}
	if m := Search(ws, "first is"); len(m) != 2 {
		t.Fatalf("unquoted: both takes: %+v", m)
	}
}

func TestSearchIgnoresCaseFillersPunctuation(t *testing.T) {
	ws := []Word{{Text: "The"}, {Text: "first"}, {Text: "um,"}, {Text: "thing,"}, {Text: "you"}}
	m := Search(ws, "first thing you")
	if len(m) != 1 || m[0].First != 1 || m[0].Last != 4 {
		t.Fatalf("%+v", m)
	}
}

// End to end on the talk fixture against its ground-truth manifest.
func TestL28_FixturePhrasesMatchGroundTruth(t *testing.T) {
	fix := filepath.Join("..", "..", "evals", "work", "fixtures")
	src := filepath.Join(fix, "talk.mp4")
	model, _ := deps.FindModel()
	if _, err := os.Stat(src); err != nil || model == "" {
		t.Skip("needs fixtures and a whisper model")
	}
	var man struct {
		Lines []struct {
			Say        string
			Tag        string
			Start, End float64
			Burst      bool
		}
	}
	b, _ := os.ReadFile(filepath.Join(fix, "talk.json"))
	json.Unmarshal(b, &man)
	tr, _, err := Transcribe("talk", src, 0, 30, Key(1, time.Now(), model, 2), Options{Model: model, Passes: 2, CacheDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	ph := Phrases(tr.Words, 0.5)
	var said []int
	for i, l := range man.Lines {
		if l.Say != "" {
			said = append(said, i)
		}
	}
	if len(ph) != len(said) {
		t.Fatalf("got %d phrases, want %d (one per spoken line)", len(ph), len(said))
	}
	for k, i := range said {
		l := man.Lines[i]
		if d := ph[k].Start - l.Start; math.Abs(d) > 0.08 {
			t.Errorf("phrase %d starts %.3f, truth %.3f (Δ %+.3f)", k, ph[k].Start, l.Start, d)
		}
		// the TTS file's end includes trailing silence, so speech ends before l.End
		if ph[k].End > l.End+0.05 || ph[k].End < l.End-0.4 {
			t.Errorf("phrase %d ends %.3f, truth (with tail) %.3f", k, ph[k].End, l.End)
		}
	}
	var sound, repeat bool
	for _, s := range tr.Suspects {
		for _, l := range man.Lines {
			if l.Burst && s.Kind == "sound" && s.Start >= l.Start && s.End <= l.End {
				sound = true
			}
		}
		if s.Kind == "repeat" && strings.Contains(s.Text, "first thing") {
			repeat = true
		}
	}
	if !sound || !repeat {
		t.Errorf("suspects missing (sound=%v repeat=%v): %+v", sound, repeat, tr.Suspects)
	}
}

// L28: breaths (short, well below speech) are not suspects; coughs, bumps and skipped words are.
// Levels from real footage (T3 deploy excerpt, speech −24 dB) and the fixture (speech −16 dB).
func TestL28_BreathIsNotASoundSuspect(t *testing.T) {
	for _, c := range []struct {
		peak, speech, dur float64
		breath            bool
	}{
		{-42, -24, 0.11, true},  // real breath, 18 dB below speech
		{-39, -24, 0.14, true},  // real breath, 15 dB below
		{-31, -24, 0.33, false}, // real: loud, right after a word — keep as a suspect
		{-35, -16, 0.33, true},  // fixture breath
		{-21, -16, 0.32, false}, // fixture bump
		{-42, -24, 1.2, false},  // long: not a breath, worth a listen
	} {
		if got := IsBreath(c.peak, c.speech, c.dur); got != c.breath {
			t.Errorf("IsBreath(%v, %v, %v) = %v, want %v", c.peak, c.speech, c.dur, got, c.breath)
		}
	}
}

func TestSentencesSplitTightEdits(t *testing.T) {
	ws := []Word{{Text: "Hi.", Start: 0, End: 0.3}, {Text: "Then", Start: 0.32, End: 0.5}, {Text: "go,", Start: 0.5, End: 0.8},
		{Text: "finally", Start: 1.0, End: 1.3}, {Text: "done.", Start: 1.3, End: 1.6}}
	if n := len(Phrases(ws, 0.5)); n != 1 {
		t.Fatalf("phrases at 0.5 s: %d, want 1 (tight edit)", n)
	}
	ps := Sentences(ws, 0.15)
	if len(ps) != 3 || ps[0].Last != 0 || ps[1].Last != 2 || ps[1].PauseAfter != 0.2 {
		t.Fatalf("sentences: %+v", ps)
	}
}
