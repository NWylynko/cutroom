package speech

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/cutroom/cutroom/internal/deps"
)

// DisfluencyPrompt nudges whisper to transcribe false starts and fillers it otherwise smooths over (L29).
const DisfluencyPrompt = "Umm, so, uh, I was like, um, thinking, uh, about it."

// RawWord is a word as whisper reported it (wav seconds).
type RawWord struct {
	Text  string
	Start float64 // DTW onset when available, else token start
	End   float64 // token end (unreliable next to pauses)
	Prob  float64
	DTW   bool
}

var dtwPresets = map[string]bool{"tiny": true, "tiny.en": true, "base": true, "base.en": true, "small": true, "small.en": true,
	"medium": true, "medium.en": true, "large-v1": true, "large-v2": true, "large-v3": true, "large-v3-turbo": true}

// ModelName returns "base.en" for ".../ggml-base.en.bin".
func ModelName(path string) string {
	return strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), "ggml-"), ".bin")
}

// Whisper runs one pass of whisper-cli over wav and returns words.
func Whisper(model, wav, prompt, tmpDir string) ([]RawWord, error) {
	of := filepath.Join(tmpDir, "pass")
	args := []string{"-m", model, "-f", wav, "-ojf", "-of", of, "-np", "-l", "en"}
	name := ModelName(model)
	if dtwPresets[name] {
		args = append(args, "-nfa", "-dtw", name) // DTW needs flash attention off
	}
	if prompt != "" {
		args = append(args, "--prompt", prompt)
	}
	if _, err := deps.Run("whisper-cli", args...); err != nil {
		return nil, err
	}
	b, err := os.ReadFile(of + ".json")
	if err != nil {
		return nil, err
	}
	defer os.Remove(of + ".json")
	return parseWhisper(b)
}

type wjson struct {
	Transcription []struct {
		Tokens []struct {
			Text    string `json:"text"`
			Offsets struct {
				From int `json:"from"`
				To   int `json:"to"`
			} `json:"offsets"`
			P    float64 `json:"p"`
			TDTW int     `json:"t_dtw"`
		} `json:"tokens"`
	} `json:"transcription"`
}

// parseWhisper merges sub-word tokens into words. A token starting with a space starts a word;
// punctuation-only tokens attach to the previous word's text but never extend its timing (they
// carry the pause that follows a sentence).
func parseWhisper(b []byte) ([]RawWord, error) {
	var d wjson
	if err := json.Unmarshal(b, &d); err != nil {
		return nil, err
	}
	var words []RawWord
	var probs []float64
	flush := func() {
		if len(words) > 0 && len(probs) > 0 {
			s := 0.0
			for _, p := range probs {
				s += p
			}
			words[len(words)-1].Prob = s / float64(len(probs))
		}
		probs = nil
	}
	depth := 0 // inside [..] or (..) annotations such as [BLANK_AUDIO] or (music)
	for _, seg := range d.Transcription {
		for _, t := range seg.Tokens {
			if o := strings.Count(t.Text, "[") + strings.Count(t.Text, "("); o > 0 || depth > 0 {
				depth += o - strings.Count(t.Text, "]") - strings.Count(t.Text, ")")
				if depth < 0 {
					depth = 0
				}
				continue
			}
			if strings.HasPrefix(t.Text, "[_") || strings.HasPrefix(t.Text, "<|") || t.Text == "" {
				continue
			}
			start := float64(t.Offsets.From) / 1000
			end := float64(t.Offsets.To) / 1000
			dtw := t.TDTW >= 0
			if dtw {
				start = float64(t.TDTW) / 100
			}
			trim := strings.TrimSpace(t.Text)
			if trim == "" {
				continue
			}
			if isPunct(trim) {
				if len(words) > 0 {
					words[len(words)-1].Text += trim
				}
				continue
			}
			if strings.HasPrefix(t.Text, " ") || len(words) == 0 {
				flush()
				words = append(words, RawWord{Text: trim, Start: start, End: end, DTW: dtw})
			} else {
				w := &words[len(words)-1]
				w.Text += trim
				w.End = end
			}
			probs = append(probs, t.P)
		}
		flush()
	}
	return words, nil
}

func isPunct(s string) bool {
	for _, r := range s {
		if !unicode.IsPunct(r) && !unicode.IsSymbol(r) {
			return false
		}
	}
	return true
}
