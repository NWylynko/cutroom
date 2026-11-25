package speech

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
)

// Word is a transcript word in source seconds.
type Word struct {
	I     int      `json:"i"`
	Text  string   `json:"text"`
	Start float64  `json:"start"`
	End   float64  `json:"end"`
	Prob  float64  `json:"p"`
	Flags []string `json:"flags,omitempty"` // "long", "uncertain-timing", "low-confidence"
}

// Suspect is something an agent should look at before cutting near it.
type Suspect struct {
	Kind    string    `json:"kind"` // repeat, sound, long-word, pass-diff, low-confidence
	Start   float64   `json:"start"`
	End     float64   `json:"end"`
	Text    string    `json:"text"`
	Detail  string    `json:"detail"`
	Lesson  string    `json:"lesson,omitempty"`
	Related []float64 `json:"related,omitempty"` // other times involved (e.g. the other takes of a repeat)
}

type Transcript struct {
	Source    string    `json:"source"`
	Model     string    `json:"model"`
	Passes    int       `json:"passes"`
	Offset    float64   `json:"audio_offset"` // added to wav times to get source times
	Duration  float64   `json:"duration"`
	RoomTone  float64   `json:"room_tone_db"`
	SpeechDB  float64   `json:"speech_db"`
	Words     []Word    `json:"words"`
	Suspects  []Suspect `json:"suspects"`
	Key       string    `json:"key"`
	Algorithm int       `json:"algorithm"`
	From      string    `json:"from,omitempty"` // set on a recorder's words mapped onto a camera's clock
}

// MapTo is this transcript (a recorder's) re-timed onto another source's clock (the camera it is
// synced to): toRef converts a time of this transcript to that source's time. Words and suspects
// outside 0–dur are dropped; word numbers are renumbered.
func (t *Transcript) MapTo(src string, toRef func(float64) float64, dur float64) *Transcript {
	m := *t
	m.Source, m.From, m.Duration, m.Words, m.Suspects = src, t.Source, dur, nil, nil
	for _, w := range t.Words {
		w.Start, w.End = toRef(w.Start), toRef(w.End)
		if w.End <= 0 || w.Start >= dur {
			continue
		}
		// a word the camera's recording cuts off is still there at its edge (T3 r49: the recorder's
		// "and" ran past the camera's last frame, was dropped, and 55 ms of it ended the video)
		w.Start, w.End = math.Max(0, w.Start), math.Min(dur, w.End)
		w.I = len(m.Words)
		m.Words = append(m.Words, w)
	}
	for _, s := range t.Suspects {
		s.Start, s.End = toRef(s.Start), toRef(s.End)
		if s.Start < 0 || s.End > dur {
			continue
		}
		rel := s.Related
		s.Related = nil
		for _, r := range rel {
			s.Related = append(s.Related, toRef(r))
		}
		m.Suspects = append(m.Suspects, s)
	}
	return &m
}

// Algorithm bumps invalidate cached transcripts.
const Algorithm = 3 // 2: overlapping repeat windows merged; 3: breaths told apart from sounds (T3, 2026-11-24)

// Build turns whisper words (wav seconds) plus the envelope into a refined transcript.
func Build(raw []RawWord, env *Envelope, offset, duration float64) ([]Word, []Suspect) {
	regions := env.Regions()
	words := refine(raw, regions)
	var sus []Suspect
	for i := range words {
		w := &words[i]
		w.I = i
		if w.End-w.Start > 0.75 {
			w.Flags = append(w.Flags, "long")
			sus = append(sus, Suspect{Kind: "long-word", Start: w.Start, End: w.End, Text: w.Text, Lesson: "L28",
				Detail: fmt.Sprintf("%q lasts %.2f s; a long word can hide a restart or a dropped take", w.Text, w.End-w.Start)})
		}
		if w.Prob > 0 && w.Prob < 0.35 {
			w.Flags = append(w.Flags, "low-confidence")
			sus = append(sus, Suspect{Kind: "low-confidence", Start: w.Start, End: w.End, Text: w.Text,
				Detail: fmt.Sprintf("recogniser confidence %.2f", w.Prob)})
		}
	}
	// Unworded sound: a region with no word in it.
	for _, r := range regions {
		has := false
		for _, w := range words {
			if w.Start < r.End && w.End > r.Start {
				has = true
				break
			}
		}
		if !has && r.End-r.Start >= 0.1 && IsBreath(r.PeakDB, env.Peak, r.End-r.Start) {
			sus = append(sus, Suspect{Kind: "breath", Start: r.Start, End: r.End, Lesson: "L28",
				Detail: fmt.Sprintf("breath or mouth noise, %.2f s, peak %.0f dB (%.0f dB below speech): normal, not a problem", r.End-r.Start, r.PeakDB, env.Peak-r.PeakDB)})
		} else if !has && r.End-r.Start >= 0.1 {
			sus = append(sus, Suspect{Kind: "sound", Start: r.Start, End: r.End, Lesson: "L28",
				Detail: fmt.Sprintf("sound with no words, %.2f s, peak %.0f dB (room tone %.0f dB): a cough, bump, breath, or words the recogniser skipped — `cutroom listen` it", r.End-r.Start, r.PeakDB, env.Floor)})
		}
	}
	sus = append(sus, repeats(words)...)
	for i := range words {
		words[i].Start = round3(words[i].Start + offset)
		words[i].End = round3(words[i].End + offset)
	}
	for i := range sus {
		sus[i].Start = round3(sus[i].Start + offset)
		sus[i].End = round3(sus[i].End + offset)
		for j := range sus[i].Related {
			sus[i].Related[j] = round3(sus[i].Related[j] + offset)
		}
	}
	sort.SliceStable(sus, func(i, j int) bool { return sus[i].Start < sus[j].Start })
	return words, sus
}

// refine sets word onsets/ends against measured sound regions.
func refine(raw []RawWord, regions []Region) []Word {
	words := make([]Word, len(raw))
	regionOf := func(t float64) int {
		for i, r := range regions {
			if t >= r.Start && t < r.End {
				return i
			}
		}
		return -1
	}
	nextRegion := func(t float64) int {
		for i, r := range regions {
			if r.Start >= t {
				return i
			}
		}
		return -1
	}
	ri := make([]int, len(raw))
	for i, w := range raw {
		words[i] = Word{Text: w.Text, Start: w.Start, End: w.End, Prob: w.Prob}
		r := regionOf(w.Start)
		if r < 0 {
			// onset in silence: snap forward to a region starting soon after, else back to one just ended
			if n := nextRegion(w.Start); n >= 0 && regions[n].Start-w.Start <= 0.35 {
				r = n
				words[i].Start = regions[n].Start
			} else if p := n - 1; n > 0 && w.Start-regions[p].End <= 0.15 {
				r = p
			} else if n < 0 && len(regions) > 0 && w.Start-regions[len(regions)-1].End <= 0.15 {
				r = len(regions) - 1
			}
		}
		ri[i] = r
	}
	// keep onsets monotonic
	for i := 1; i < len(words); i++ {
		if words[i].Start < words[i-1].Start {
			words[i].Start = words[i-1].Start
		}
	}
	for i := range words {
		r := ri[i]
		if r < 0 {
			words[i].Flags = append(words[i].Flags, "uncertain-timing")
			if i+1 < len(words) {
				words[i].End = math.Min(math.Max(words[i].End, words[i].Start+0.05), words[i+1].Start)
			}
			continue
		}
		// first word of its region: onset at the region start (DTW onsets run 40–300 ms late)
		if i == 0 || ri[i-1] != r {
			words[i].Start = regions[r].Start
		}
		if i+1 < len(words) && ri[i+1] == r {
			words[i].End = words[i+1].Start
		} else {
			words[i].End = regions[r].End
		}
		if words[i].End <= words[i].Start {
			words[i].End = words[i].Start + 0.05
		}
	}
	return words
}

var fillers = map[string]bool{"um": true, "uh": true, "umm": true, "uhh": true, "er": true, "erm": true, "hmm": true, "mm": true, "ah": true}
var nonAlnum = regexp.MustCompile(`[^a-z0-9']+`)

// Norm lowercases and strips punctuation; "" for fillers.
func Norm(s string) string {
	n := strings.Trim(nonAlnum.ReplaceAllString(strings.ToLower(s), ""), "'")
	if fillers[n] {
		return ""
	}
	return n
}

// cand is one repeated phrase: token index ranges [first, last] of each take, in time order.
type cand struct{ occ [][2]int }

// merge joins two repeats whose takes overlap pairwise (the same retake seen through two windows).
func (c cand) merge(d cand) (cand, bool) {
	if len(c.occ) != len(d.occ) {
		return c, false
	}
	m := cand{}
	for i := range c.occ {
		a, b := c.occ[i], d.occ[i]
		if a[0] > b[1]+1 || b[0] > a[1]+1 { // disjoint (adjacent counts as overlapping)
			return c, false
		}
		m.occ = append(m.occ, [2]int{min(a[0], b[0]), max(a[1], b[1])})
	}
	for i := 1; i < len(m.occ); i++ { // merged takes must not run into each other
		if m.occ[i][0] <= m.occ[i-1][1] {
			return c, false
		}
	}
	return m, true
}

// repeats finds n-grams (3–8 words) said more than once within 30 s: retakes (L36).
func repeats(words []Word) []Suspect {
	type tok struct {
		w string
		i int
	}
	var toks []tok
	for i, w := range words {
		if n := Norm(w.Text); n != "" {
			toks = append(toks, tok{n, i})
		}
	}
	covered := map[int]bool{}
	var out []Suspect
	var cands []cand
	for n := 8; n >= 3; n-- {
		pos := map[string][]int{}
		for k := 0; k+n <= len(toks); k++ {
			var parts []string
			for _, t := range toks[k : k+n] {
				parts = append(parts, t.w)
			}
			key := strings.Join(parts, " ")
			pos[key] = append(pos[key], k)
		}
		keys := make([]string, 0, len(pos))
		for k := range pos {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			ks := pos[key]
			if len(ks) < 2 {
				continue
			}
			// cluster occurrences within 30 s of the first, non-overlapping
			var group []int
			for _, k := range ks {
				if len(group) > 0 && (k < group[len(group)-1]+n || words[toks[k].i].Start-words[toks[group[0]].i].Start > 30) {
					continue
				}
				group = append(group, k)
			}
			if len(group) < 2 || covered[toks[group[0]].i] {
				continue
			}
			for _, k := range group {
				for j := k; j < k+n; j++ {
					covered[toks[j].i] = true
				}
			}
			c := cand{}
			for _, k := range group {
				c.occ = append(c.occ, [2]int{k, k + n - 1})
			}
			cands = append(cands, c)
		}
	}
	// sliding windows of one retake ("no point replacing a $5 a month…", "a $5 a month subscription
	// with…") overlap in every take: merge them so one retake is reported once.
	for merged := true; merged; {
		merged = false
		for a := 0; a < len(cands) && !merged; a++ {
			for b := a + 1; b < len(cands) && !merged; b++ {
				if m, ok := cands[a].merge(cands[b]); ok {
					cands[a] = m
					cands = append(cands[:b], cands[b+1:]...)
					merged = true
				}
			}
		}
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].occ[0][0] < cands[j].occ[0][0] })
	for _, c := range cands {
		var times []float64
		for _, o := range c.occ {
			times = append(times, words[toks[o[0]].i].Start)
		}
		var parts []string
		for _, t := range toks[c.occ[0][0] : c.occ[0][1]+1] {
			parts = append(parts, t.w)
		}
		key := strings.Join(parts, " ")
		first, last := words[toks[c.occ[0][0]].i], words[toks[c.occ[len(c.occ)-1][1]].i]
		out = append(out, Suspect{Kind: "repeat", Start: first.Start, End: last.End, Text: key, Lesson: "L36",
			Related: times,
			Detail:  fmt.Sprintf("%q is said %d× within %.0f s — likely a retake; the last take is usually the keeper", key, len(c.occ), times[len(times)-1]-times[0])})
	}
	return out
}

// DiffPasses compares two passes and returns spans the second pass heard differently (L29).
// Uses LCS within 60 s windows so memory stays small on long recordings.
func DiffPasses(a, b []Word) []Suspect {
	var out []Suspect
	if len(a) == 0 {
		return nil
	}
	end := a[len(a)-1].End
	for w0 := 0.0; w0 <= end; w0 += 60 {
		var x, y []Word
		for _, w := range a {
			if w.Start >= w0 && w.Start < w0+60 && Norm(w.Text) != "" {
				x = append(x, w)
			}
		}
		for _, w := range b {
			if w.Start >= w0 && w.Start < w0+60 && Norm(w.Text) != "" {
				y = append(y, w)
			}
		}
		out = append(out, lcsDiff(x, y)...)
	}
	return out
}

func lcsDiff(x, y []Word) []Suspect {
	n, m := len(x), len(y)
	dp := make([][]int32, n+1)
	for i := range dp {
		dp[i] = make([]int32, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if Norm(x[i].Text) == Norm(y[j].Text) {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}
	var out []Suspect
	var onlyA, onlyB []Word
	emit := func() {
		if len(onlyA) == 0 && len(onlyB) == 0 {
			return
		}
		var ta, tb []string
		for _, w := range onlyA {
			ta = append(ta, w.Text)
		}
		for _, w := range onlyB {
			tb = append(tb, w.Text)
		}
		s := Suspect{Kind: "pass-diff", Lesson: "L29"}
		all := append(append([]Word{}, onlyA...), onlyB...)
		s.Start, s.End = all[0].Start, all[0].End
		for _, w := range all {
			s.Start = math.Min(s.Start, w.Start)
			s.End = math.Max(s.End, w.End)
		}
		switch {
		case len(ta) == 0:
			s.Text = strings.Join(tb, " ")
			s.Detail = fmt.Sprintf("only the disfluency pass heard %q — possibly a stumble the main transcript smoothed over", s.Text)
		case len(tb) == 0:
			s.Text = strings.Join(ta, " ")
			s.Detail = fmt.Sprintf("only the main pass heard %q", s.Text)
		default:
			s.Text = strings.Join(ta, " ")
			s.Detail = fmt.Sprintf("passes disagree: %q vs %q", strings.Join(ta, " "), strings.Join(tb, " "))
		}
		out = append(out, s)
		onlyA, onlyB = nil, nil
	}
	i, j := 0, 0
	for i < n || j < m {
		switch {
		case i < n && j < m && Norm(x[i].Text) == Norm(y[j].Text):
			emit()
			i, j = i+1, j+1
		case j >= m || (i < n && dp[i+1][j] >= dp[i][j+1]):
			onlyA = append(onlyA, x[i])
			i++
		default:
			onlyB = append(onlyB, y[j])
			j++
		}
	}
	emit()
	return out
}

// Phrase is a run of words between pauses.
type Phrase struct {
	Start, End float64
	First      int // index of first word
	Last       int
	PauseAfter float64
}

// Phrases splits at pauses ≥ minGap.
func Phrases(words []Word, minGap float64) []Phrase { return phrases(words, minGap, false) }

// Sentences is Phrases that also splits after every sentence end — for reading a tight edit, where
// the pauses are too short to split on.
func Sentences(words []Word, minGap float64) []Phrase { return phrases(words, minGap, true) }

func phrases(words []Word, minGap float64, atSentences bool) []Phrase {
	var ps []Phrase
	for i, w := range words {
		ends := false
		if atSentences && i > 0 {
			t := strings.TrimRight(words[i-1].Text, `"')`)
			ends = strings.HasSuffix(t, ".") || strings.HasSuffix(t, "?") || strings.HasSuffix(t, "!")
		}
		if i == 0 || w.Start-words[i-1].End >= minGap || ends {
			if len(ps) > 0 {
				ps[len(ps)-1].PauseAfter = round3(w.Start - words[i-1].End)
			}
			ps = append(ps, Phrase{Start: w.Start, End: w.End, First: i, Last: i})
			continue
		}
		ps[len(ps)-1].End = w.End
		ps[len(ps)-1].Last = i
	}
	return ps
}

// Text joins words i..j.
func Text(words []Word, i, j int) string {
	var b strings.Builder
	for k := i; k <= j; k++ {
		if k > i {
			b.WriteByte(' ')
		}
		b.WriteString(words[k].Text)
	}
	return b.String()
}

// Match is a search hit: a word span.
type Match struct {
	First, Last int
	Start, End  float64
	Text        string
}

// Search finds every occurrence of the phrase. Case and punctuation are ignored. Fillers (um, uh…)
// in the transcript are skipped unless the query itself contains them, so quoting "is, um" can
// pick out the take with the filler.
func Search(words []Word, phrase string) []Match {
	var q []string
	for _, f := range strings.Fields(phrase) {
		if n := normRaw(f); n != "" {
			q = append(q, n)
		}
	}
	if len(q) == 0 {
		return nil
	}
	var out []Match
	for start := 0; start < len(words); start++ {
		if normRaw(words[start].Text) != q[0] {
			continue
		}
		k, i := 0, start
		for i < len(words) && k < len(q) {
			w := normRaw(words[i].Text)
			switch {
			case w == q[k]:
				k++
				i++
			case fillers[w] || w == "":
				i++ // skip a filler the query doesn't mention
			default:
				i = len(words) + 1
			}
		}
		if k == len(q) {
			out = append(out, Match{start, i - 1, words[start].Start, words[i-1].End, Text(words, start, i-1)})
		}
	}
	return out
}

func normRaw(s string) string {
	return strings.Trim(nonAlnum.ReplaceAllString(strings.ToLower(s), ""), "'")
}

func round3(f float64) float64 { return math.Round(f*1000) / 1000 }

// IsBreath tells a breath or mouth noise from a sound worth checking (cough, bump, skipped words):
// short and well below speech level. Measured on real footage (T3): breaths sit 15–19 dB below speech
// peaks; coughs, bumps and missed words within ~7 dB.
func IsBreath(peakDB, speechDB, dur float64) bool {
	return dur < 0.8 && peakDB < speechDB-12
}

// asideMarkers open a digression: a sentence that starts with one is probably off the topic.
var asideMarkers = [][]string{{"by", "the", "way"}, {"anyway"}, {"side", "note"}, {"off", "topic"},
	{"oh", "and"}, {"oops"}, {"sorry"}, {"hold", "on"}, {"random", "thought"}, {"fun", "fact"},
	{"unrelated"}, {"speaking", "of"}, {"before", "i", "forget"}, {"just", "a", "quick", "note"}}

// Asides flags sentences that open with a digression marker ("By the way, my cat…") as suspects of
// kind "aside": probably off-topic, worth deciding about (cut it, or keep it on purpose).
func Asides(words []Word) []Suspect {
	var out []Suspect
	for i := range words {
		if i > 0 && !endsSentenceWord(words[i-1].Text) {
			continue // only at the start of a sentence
		}
		for _, m := range asideMarkers {
			if i+len(m) > len(words) {
				continue
			}
			ok := true
			for k, w := range m {
				ok = ok && Norm(words[i+k].Text) == w
			}
			if !ok {
				continue
			}
			j := i
			for j < len(words)-1 && !endsSentenceWord(words[j].Text) {
				j++
			}
			out = append(out, Suspect{Kind: "aside", Start: words[i].Start, End: words[j].End, Text: Text(words, i, j),
				Detail: fmt.Sprintf("opens with %q: probably a digression — cut it if it is off-topic, or keep it", strings.Join(m, " "))})
			break
		}
	}
	return out
}

func endsSentenceWord(w string) bool {
	w = strings.TrimRight(w, `"')`)
	return strings.HasSuffix(w, ".") || strings.HasSuffix(w, "?") || strings.HasSuffix(w, "!")
}

// ReconcileTiming fixes words this transcript could not place in its own sound (flag
// "uncertain-timing") using a second transcript of the same speech on the same clock — a recorder's
// words mapped onto a camera, checked against the camera microphone's own words. A word is moved
// only when the reference has the same word, confidently placed within maxShift, with a neighbour
// that agrees too (so the two are the same utterance). The next word's onset is pushed past it when
// they now overlap. Moved words get the flag "timing-from-camera". Returns how many moved.
// (T3 r49: a recorder's "and" after a pause sat 0.9 s early; a cut "before they" left it audible.)
func (t *Transcript) ReconcileTiming(ref *Transcript, maxShift float64) int {
	if t == nil || ref == nil {
		return 0
	}
	has := func(w Word, f string) bool {
		for _, x := range w.Flags {
			if x == f {
				return true
			}
		}
		return false
	}
	n := 0
	for i := range t.Words {
		w := t.Words[i]
		if !has(w, "uncertain-timing") || Norm(w.Text) == "" {
			continue
		}
		best, bd := -1, maxShift
		for k, r := range ref.Words {
			if Norm(r.Text) != Norm(w.Text) || has(r, "uncertain-timing") {
				continue
			}
			d := math.Abs(r.Start - w.Start)
			if d > bd {
				continue
			}
			prevOK := i > 0 && k > 0 && Norm(t.Words[i-1].Text) == Norm(ref.Words[k-1].Text)
			nextOK := i+1 < len(t.Words) && k+1 < len(ref.Words) && Norm(t.Words[i+1].Text) == Norm(ref.Words[k+1].Text)
			if prevOK || nextOK {
				best, bd = k, d
			}
		}
		if best < 0 || math.Abs(ref.Words[best].Start-w.Start) < 0.08 {
			continue
		}
		r := ref.Words[best]
		lo, hi := math.Inf(-1), math.Inf(1)
		if i > 0 {
			lo = t.Words[i-1].End
		}
		if i+1 < len(t.Words) {
			hi = t.Words[i+1].End - 0.05
		}
		s, e := math.Max(r.Start, lo), math.Min(r.End, hi)
		if e-s < 0.03 {
			continue
		}
		t.Words[i].Start, t.Words[i].End = round3(s), round3(e)
		var fl []string
		for _, x := range w.Flags {
			if x != "uncertain-timing" {
				fl = append(fl, x)
			}
		}
		t.Words[i].Flags = append(fl, "timing-from-camera")
		if i+1 < len(t.Words) && t.Words[i+1].Start < t.Words[i].End {
			t.Words[i+1].Start = t.Words[i].End
		}
		n++
	}
	return n
}
