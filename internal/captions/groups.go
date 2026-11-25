package captions

import (
	"strings"
	"unicode/utf8"
)

// Group is a burned-in caption: 1–3 words shown together (a 4th only to avoid a weak ending; L48).
type Group struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
	First int     `json:"first_word"`
	Last  int     `json:"last_word"`
}

// GroupOptions: MaxChars per group (20, L48), MaxWords (3), Pause that makes a good break (0.3 s;
// ≥ 1 s always splits),
// Hold after the last word when nothing follows soon (0.4 s).
type GroupOptions struct {
	MaxChars, MaxWords int
	Pause, Hold        float64
}

var DefaultGroups = GroupOptions{MaxChars: 20, MaxWords: 3, Pause: 0.3, Hold: 0.4}

// Groups splits the edit's words into short burned-in caption groups: at most MaxWords words and
// MaxChars characters, split at sentence ends and pauses, never ending on a weak word ("the", "to")
// and never leaving a short sentence tail on its own (L48). Each group shows until the next starts
// (or its last word ends + Hold, when the next is further away).
func Groups(ws []Word, o GroupOptions, duration float64) []Group {
	var gs []Group
	// runs: sentences, split again at pauses; each is grouped on its own
	a := 0
	for k := range ws {
		if k+1 == len(ws) || endsSentence(ws[k].Text) || ws[k+1].Start-ws[k].End >= 1.0 {
			for _, r := range bestGroups(ws, a, k+1, o) {
				gs = append(gs, Group{Start: ws[r[0]].Start, End: ws[r[1]-1].End, Text: groupText(ws[r[0]:r[1]]), First: r[0], Last: r[1] - 1})
			}
			a = k + 1
		}
	}
	for k := range gs {
		end := gs[k].End + o.Hold
		if k+1 < len(gs) {
			if gs[k+1].Start-gs[k].End < 0.5 {
				end = gs[k+1].Start // continuous speech: no flicker between groups
			} else {
				end = min(end, gs[k+1].Start)
			}
		}
		if duration > 0 {
			end = min(end, duration)
		}
		gs[k].Start, gs[k].End = r3(gs[k].Start), r3(end)
	}
	return gs
}

// bestGroups partitions ws[a:b] into groups [s, e) with the lowest total cost: weak endings and
// lone words cost most, a break after a comma is cheap, and groups near 15 characters read best.
func bestGroups(ws []Word, a, b int, o GroupOptions) [][2]int {
	n := b - a
	best := make([]float64, n+1)
	from := make([]int, n+1)
	for k := 1; k <= n; k++ {
		best[k] = 1e18
		for s := k - 1; s >= 0 && k-s <= o.MaxWords+1; s-- {
			chars := utf8.RuneCountInString(groupText(ws[a+s : a+k]))
			if (chars > o.MaxChars || k-s > o.MaxWords) && k-s > 1 {
				if k-s > o.MaxWords+1 || chars > o.MaxChars {
					continue
				}
			}
			c := 5.0 + float64((chars-15)*(chars-15))/40
			last := a + k - 1
			if weakAt(ws, last) && (k < n || !endsSentence(ws[last].Text)) {
				c += 10
				if determiners[bare(ws[last].Text)] {
					c += 7 // "the |", "my |": the worst place to break
				}
			}
			if k < n && ws[last+1].Start-ws[last].End >= o.Pause {
				c -= 5 // the speaker pauses here: a natural break
			}
			if k-s == 1 {
				c += 6
				if utf8.RuneCountInString(ws[last].Text) <= 5 {
					c += 6
				}
			}
			if k-s > o.MaxWords {
				c += 5 // a fourth word only to avoid something worse (a weak ending)
			}
			if p := trailingPunct(ws[last].Text); k < n && (p == "," || p == ";" || p == ":") {
				c -= 4
			}
			if best[s]+c < best[k] {
				best[k], from[k] = best[s]+c, s
			}
		}
	}
	var rs [][2]int
	for k := n; k > 0; k = from[k] {
		rs = append([][2]int{{a + from[k], a + k}}, rs...)
	}
	return rs
}

// groupText joins words and drops the trailing comma/full stop a short caption doesn't need.
func groupText(ws []Word) string {
	return strings.TrimRight(join(ws), ".,;:")
}

// determiners lean hardest on the next word.
var determiners = set(`a an the my your our their his her its this that these those to`)
