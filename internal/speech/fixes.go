package speech

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// Fix corrects what the recogniser heard: the words From (matched like a quote) become To, in
// source Source, at every place they are said — or only the one nearest Near (a source time) when
// Near is set. Timing stays: the new words share the old words' span. The recording is untouched;
// captions, subtitle files and word checks use the corrected words.
type Fix struct {
	Source string  `json:"source"`
	From   string  `json:"from"`
	To     string  `json:"to"`
	Near   float64 `json:"near,omitempty"`
}

func fixesPath(cacheDir string) string { return filepath.Join(cacheDir, "transcripts", "fixes.json") }

// LoadFixes reads a project's word fixes (none when there is no file).
func LoadFixes(cacheDir string) []Fix {
	b, err := os.ReadFile(fixesPath(cacheDir))
	if err != nil {
		return nil
	}
	var fs []Fix
	_ = json.Unmarshal(b, &fs)
	return fs
}

// SaveFixes writes a project's word fixes.
func SaveFixes(cacheDir string, fs []Fix) error {
	if err := os.MkdirAll(filepath.Dir(fixesPath(cacheDir)), 0o755); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(fs, "", "  ")
	return os.WriteFile(fixesPath(cacheDir), append(b, '\n'), 0o644)
}

// CopyFixes copies a project's word fixes into another (a short made from it).
func CopyFixes(fromCache, toCache string) {
	if fs := LoadFixes(fromCache); len(fs) > 0 {
		_ = SaveFixes(toCache, fs)
	}
}

// ApplyFixes corrects t's words with every fix for its source; it returns how many places changed.
func (t *Transcript) ApplyFixes(fs []Fix) int {
	n := 0
	for _, f := range fs {
		if f.Source != t.Source || strings.TrimSpace(f.To) == "" {
			continue
		}
		ms := Search(t.Words, f.From)
		if f.Near > 0 && len(ms) > 1 {
			best := ms[0]
			for _, m := range ms {
				if math.Abs(t.Words[m.First].Start-f.Near) < math.Abs(t.Words[best.First].Start-f.Near) {
					best = m
				}
			}
			ms = []Match{best}
		}
		for k := len(ms) - 1; k >= 0; k-- { // from the end, so earlier indexes stay valid
			t.replaceWords(ms[k].First, ms[k].Last, f.To)
			n++
		}
	}
	if n > 0 {
		for i := range t.Words {
			t.Words[i].I = i
		}
	}
	return n
}

// replaceWords puts the words of to in place of words a..b, sharing their time span by length; the
// last old word's trailing punctuation stays when to has none.
func (t *Transcript) replaceWords(a, b int, to string) {
	start, end := t.Words[a].Start, t.Words[b].End
	last := strings.TrimSpace(t.Words[b].Text)
	punct := last[len(strings.TrimRight(last, ".,?!;:")):]
	toks := strings.Fields(to)
	if punct != "" && strings.TrimRight(toks[len(toks)-1], ".,?!;:") == toks[len(toks)-1] {
		toks[len(toks)-1] += punct
	}
	total := 0
	for _, s := range toks {
		total += utf8.RuneCountInString(s)
	}
	var nw []Word
	at := start
	for _, s := range toks {
		d := (end - start) * float64(utf8.RuneCountInString(s)) / float64(max(1, total))
		nw = append(nw, Word{Text: s, Start: round3(at), End: round3(at + d), Prob: 1, Flags: []string{"fixed"}})
		at += d
	}
	t.Words = append(t.Words[:a], append(nw, t.Words[b+1:]...)...)
}
