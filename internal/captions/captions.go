// Package captions turns the edit's words (in output time) into caption cues, writes and reads
// SRT/VTT, and checks cue files: reading limits, sentence seams at cuts (L49), words, timing.
package captions

import (
	"fmt"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/NWylynko/cutroom/internal/speech"
	"github.com/NWylynko/cutroom/internal/timeline"
)

// Word is a word of the edit in output time.
type Word struct {
	Text       string  `json:"text"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	AfterCut   bool    `json:"after_cut,omitempty"`   // first word after a cut in the edit
	Segment    string  `json:"segment"`               // segment id
	SourceTime float64 `json:"source_time,omitempty"` // the word's start in its source
}

// Cue is one caption.
type Cue struct {
	Start float64  `json:"start"`
	End   float64  `json:"end"`
	Lines []string `json:"lines"`
	First int      `json:"first_word"` // index into the words it was built from (-1 when parsed)
	Last  int      `json:"last_word"`
}

func (c Cue) Text() string { return strings.Join(c.Lines, " ") }

// Options are the reading limits.
type Options struct {
	MaxChars int     // per line
	MaxLines int     // per cue
	MaxDur   float64 // s
	MinDur   float64 // s, when there is room
	Hold     float64 // s after the last word
	Gap      float64 // s between cues
	Pause    float64 // a pause this long always starts a new cue
	MaxCPS   float64 // reading speed limit (characters per second)
}

var Default = Options{MaxChars: 42, MaxLines: 2, MaxDur: 7, MinDur: 1, Hold: 0.3, Gap: 0.04, Pause: 0.7, MaxCPS: 20}

// OutputWords lists the words the timeline keeps, in output time. A word is kept when its middle is
// inside a segment; its times are clamped to the segment. Fillers (um, uh) are dropped, and
// punctuation they carried ends up on the word before ("is, UM, the" → "is, the"). The first word of
// a segment that doesn't continue the previous one (a cut) is marked AfterCut. missing lists video
// sources in the edit that have no transcript.
func OutputWords(tl *timeline.Timeline, tr timeline.Transcripts) (ws []Word, missing []string) {
	tl.Recompute()
	fps := tl.Rate()
	for k, g := range tl.Segments {
		t := tr[g.Source]
		if t == nil {
			if !contains(missing, g.Source) {
				missing = append(missing, g.Source)
			}
			continue
		}
		cont := k > 0 && tl.Segments[k-1].Source == g.Source && math.Abs(tl.Segments[k-1].Out-g.In) < 0.5/fps
		end := g.OutStart + float64(tl.Frames(g))/fps
		first := true
		for _, w := range t.Words {
			mid := (w.Start + w.End) / 2
			if mid < g.In || mid >= g.Out {
				continue
			}
			text := strings.TrimSpace(w.Text)
			if speech.Norm(text) == "" {
				// a filler: keep a comma or full stop it carried, on the word before (same segment)
				if p := trailingPunct(text); p != "" && len(ws) > 0 && !first {
					prev := &ws[len(ws)-1]
					if trailingPunct(prev.Text) == "" {
						prev.Text += p
					}
				}
				continue
			}
			a := g.OutStart + math.Max(w.Start, g.In) - g.In
			b := math.Min(g.OutStart+w.End-g.In, end)
			ws = append(ws, Word{Text: text, Start: r3(a), End: r3(b), AfterCut: first && k > 0 && !cont, Segment: g.ID, SourceTime: w.Start})
			first = false
		}
	}
	return ws, missing
}

// Build groups words into cues.
func Build(ws []Word, o Options, duration float64) []Cue {
	var cues []Cue
	i := 0
	for i < len(ws) {
		j := extent(ws, i, o) // cue = ws[i..j]
		cues = append(cues, Cue{Start: r3(ws[i].Start), End: r3(ws[j].End), Lines: wrap(ws[i:j+1], o), First: i, Last: j})
		i = j + 1
	}
	// timing: hold after the last word, at least MinDur when there is room, never overlapping
	for k := range cues {
		c := &cues[k]
		limit := duration
		if k+1 < len(cues) {
			limit = cues[k+1].Start - o.Gap
		}
		end := math.Max(c.End+o.Hold, c.Start+o.MinDur)
		c.End = r3(math.Max(c.End, math.Min(end, limit)))
		if duration > 0 && c.End > duration {
			c.End = r3(duration)
		}
	}
	return cues
}

// extent returns the last word index of the cue starting at i.
func extent(ws []Word, i int, o Options) int {
	maxLen := o.MaxChars * o.MaxLines
	j := i
	for j+1 < len(ws) {
		n := j + 1
		if fits := textLen(ws[i:n+1]) <= maxLen && ws[n].End-ws[i].Start <= o.MaxDur && canWrap(ws[i:n+1], o); !fits {
			break
		}
		if ws[n].Start-ws[j].End >= o.Pause {
			return j
		}
		// a new sentence starts a new cue when this one can end here (and isn't a lone word)
		if endsSentence(ws[j].Text) && (j > i || len(ws[j].Text) > 6) {
			return j
		}
		// a cut where the sentence before doesn't continue: start fresh
		if ws[n].AfterCut && startsSentence(ws[n].Text) {
			return j
		}
		j = n
	}
	if j+1 >= len(ws) {
		return j
	}
	// forced break before j+1: end at a comma/dash in the cue's second half when there is one …
	for k := j; k >= i+max(1, (j-i)/2); k-- {
		if p := trailingPunct(ws[k].Text); p == "," || p == ";" || p == ":" || strings.HasSuffix(ws[k].Text, "--") {
			j = k
			break
		}
	}
	// … don't strand a 1–2 word sentence tail (L49) …
	if e := sentenceEndFrom(ws, j+1); e >= 0 && e-(j+1)+1 <= 2 && !ws[j+1].AfterCut {
		need := 3 - (e - (j + 1) + 1)
		if j-need >= i+1 && !endsSentence(ws[j-need].Text) {
			j -= need
		}
	}
	// … and don't end on a weak word
	for j > i+1 && weakAt(ws, j) {
		j--
	}
	return j
}

// sentenceEndFrom: index of the first word at or after k that ends a sentence (-1 if none within 4).
func sentenceEndFrom(ws []Word, k int) int {
	for m := k; m < len(ws) && m < k+4; m++ {
		if endsSentence(ws[m].Text) {
			return m
		}
		if m > k && ws[m].AfterCut {
			return -1
		}
	}
	return -1
}

// wrap splits a cue's words into ≤ MaxLines balanced lines, preferring breaks after punctuation and
// never ending a line on a weak word when there is another choice.
func wrap(ws []Word, o Options) []string {
	all := join(ws)
	if utf8.RuneCountInString(all) <= o.MaxChars || o.MaxLines < 2 || len(ws) < 2 {
		return []string{all}
	}
	best, bestScore := -1, math.Inf(1)
	for k := 1; k < len(ws); k++ { // line 1 = ws[:k]
		a, b := utf8.RuneCountInString(join(ws[:k])), utf8.RuneCountInString(join(ws[k:]))
		if a > o.MaxChars || b > o.MaxChars {
			continue
		}
		s := math.Abs(float64(a - b))
		if trailingPunct(ws[k-1].Text) != "" {
			s -= 12
		}
		if weakAt(ws, k-1) {
			s += 25
		}
		if s < bestScore {
			best, bestScore = k, s
		}
	}
	if best < 0 {
		return []string{all}
	}
	return []string{join(ws[:best]), join(ws[best:])}
}

func canWrap(ws []Word, o Options) bool {
	if utf8.RuneCountInString(join(ws)) <= o.MaxChars {
		return true
	}
	if o.MaxLines < 2 {
		return false
	}
	for k := 1; k < len(ws); k++ {
		if utf8.RuneCountInString(join(ws[:k])) <= o.MaxChars && utf8.RuneCountInString(join(ws[k:])) <= o.MaxChars {
			return true
		}
	}
	return false
}

func join(ws []Word) string {
	var p []string
	for _, w := range ws {
		p = append(p, w.Text)
	}
	return strings.Join(p, " ")
}

func textLen(ws []Word) int { return utf8.RuneCountInString(join(ws)) }

// weakWords shouldn't end a line or cue: articles, prepositions, conjunctions, auxiliaries,
// possessives and question words all lean on the word after them.
var weakWords = set(`a an the to of and or but nor in on at for with from by into onto than as if so
is are was were be been am will would can could should shall may might must do does did have has had
not my your our their his her its this that these those which who whom whose what when where how why
because while although though about like very just really it's i'm i've i'll we're you're they're`)

// subjects are weak after a conjunction ("…, if you | don't know"); as objects they're fine.
var subjects = set(`i you we they he she it`)

func set(s string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(s) {
		m[w] = true
	}
	return m
}

func bare(w string) string {
	return strings.ToLower(strings.Trim(strings.TrimSpace(w), `"'“”‘’()`))
}

func weak(w string) bool {
	return trailingPunct(w) == "" && !strings.HasSuffix(w, "--") && weakWords[bare(w)]
}

// weakAt: ws[k] shouldn't end a line or cue.
func weakAt(ws []Word, k int) bool {
	if weak(ws[k].Text) {
		return true
	}
	return trailingPunct(ws[k].Text) == "" && subjects[bare(ws[k].Text)] && k > 0 &&
		(weak(ws[k-1].Text) || trailingPunct(ws[k-1].Text) != "")
}

func endsSentence(w string) bool {
	w = strings.TrimRight(w, `"')]”’`)
	return strings.HasSuffix(w, ".") || strings.HasSuffix(w, "?") || strings.HasSuffix(w, "!")
}

func startsSentence(w string) bool {
	r, _ := utf8.DecodeRuneInString(strings.TrimLeft(w, `"'(“‘`))
	return unicode.IsUpper(r) && !strings.HasPrefix(w, "I ") && w != "I" && !strings.HasPrefix(w, "I'")
}

func trailingPunct(w string) string {
	t := strings.TrimRightFunc(w, func(r rune) bool { return strings.ContainsRune(".,!?;:", r) })
	return w[len(t):]
}

func contains(s []string, x string) bool {
	for _, y := range s {
		if y == x {
			return true
		}
	}
	return false
}

func r3(f float64) float64 { return math.Round(f*1000) / 1000 }

// Stamp formats seconds as SRT (00:00:01,234) or VTT (00:00:01.234) time.
func Stamp(s float64, vtt bool) string {
	ms := int64(math.Round(s * 1000))
	sep := ","
	if vtt {
		sep = "."
	}
	return fmt.Sprintf("%02d:%02d:%02d%s%03d", ms/3600000, ms/60000%60, ms/1000%60, sep, ms%1000)
}
