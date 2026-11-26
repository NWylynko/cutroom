package speech

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Options struct {
	Model    string // path to ggml model
	Passes   int    // 1 or 2
	CacheDir string // <project>/.cutroom
	Force    bool
}

// Key identifies a transcript's inputs; a different key means re-transcribe.
func Key(srcSize int64, srcMod time.Time, model string, passes int) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%d|%d|%s|%d|%d", srcSize, srcMod.UnixNano(), ModelName(model), passes, Algorithm)))
	return fmt.Sprintf("%x", h[:8])
}

func TranscriptPath(cacheDir, id string) string {
	return filepath.Join(cacheDir, "transcripts", id+".json")
}

func WavPath(cacheDir, id string) string { return filepath.Join(cacheDir, "audio", id+".wav") }

func Load(cacheDir, id string) (*Transcript, error) {
	b, err := os.ReadFile(TranscriptPath(cacheDir, id))
	if err != nil {
		return nil, err
	}
	t := &Transcript{}
	if err := json.Unmarshal(b, t); err != nil {
		return t, err
	}
	t.ApplyFixes(LoadFixes(cacheDir)) // the user's corrections of what was heard
	// repeats (retakes) are found again from the words, so older transcripts get today's rules
	kept := t.Suspects[:0]
	for _, x := range t.Suspects {
		if x.Kind != "repeat" {
			kept = append(kept, x)
		}
	}
	t.Suspects = append(kept, repeats(t.Words)...)
	// asides are found from the words on load, so transcripts made before them get them too
	for _, a := range Asides(t.Words) {
		dup := false
		for _, x := range t.Suspects {
			dup = dup || (x.Kind == "aside" && x.Start == a.Start)
		}
		if !dup {
			t.Suspects = append(t.Suspects, a)
		}
	}
	sortSuspects(t.Suspects)
	return t, nil
}

// Transcribe produces (or returns the cached) transcript of a source. cached reports a cache hit.
func Transcribe(id, src string, audioStart, duration float64, key string, o Options) (t *Transcript, cached bool, err error) {
	if !o.Force {
		if old, err := Load(o.CacheDir, id); err == nil && old.Key == key {
			return old, true, nil
		}
	}
	if err := os.MkdirAll(filepath.Join(o.CacheDir, "audio"), 0o755); err != nil {
		return nil, false, err
	}
	if err := os.MkdirAll(filepath.Join(o.CacheDir, "transcripts"), 0o755); err != nil {
		return nil, false, err
	}
	wav := WavPath(o.CacheDir, id)
	if err := ExtractWav(src, wav); err != nil {
		return nil, false, err
	}
	x, err := ReadWav(wav)
	if err != nil {
		return nil, false, err
	}
	env := NewEnvelope(x)
	tmp, err := os.MkdirTemp(o.CacheDir, "whisper")
	if err != nil {
		return nil, false, err
	}
	defer os.RemoveAll(tmp)

	raw, err := Whisper(o.Model, wav, "", tmp)
	if err != nil {
		return nil, false, err
	}
	words, sus := Build(raw, env, audioStart, duration)
	if o.Passes >= 2 {
		raw2, err := Whisper(o.Model, wav, DisfluencyPrompt, tmp)
		if err != nil {
			return nil, false, err
		}
		words2, _ := Build(raw2, env, audioStart, duration)
		sus = append(sus, DiffPasses(words, words2)...)
		sortSuspects(sus)
	}
	t = &Transcript{Source: id, Model: ModelName(o.Model), Passes: max(1, o.Passes), Offset: audioStart, Duration: duration,
		RoomTone: round3(env.Floor), SpeechDB: round3(env.Peak), Words: words, Suspects: sus, Key: key, Algorithm: Algorithm}
	if t.Words == nil {
		t.Words = []Word{}
	}
	if t.Suspects == nil {
		t.Suspects = []Suspect{}
	}
	b, _ := json.MarshalIndent(t, "", " ")
	if err := os.WriteFile(TranscriptPath(o.CacheDir, id), b, 0o644); err != nil {
		return nil, false, err
	}
	return t, false, nil
}

func sortSuspects(s []Suspect) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j].Start < s[j-1].Start; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
