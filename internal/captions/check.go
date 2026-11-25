package captions

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/NWylynko/cutroom/internal/speech"
)

// Issue is one check result.
type Issue struct {
	Check  string  `json:"check"`
	Status string  `json:"status"` // pass, warn, fail
	At     float64 `json:"at,omitempty"`
	Cue    int     `json:"cue,omitempty"` // 1-based
	Detail string  `json:"detail"`
	Fix    string  `json:"fix,omitempty"`
}

// Check verifies cues against the edit's words (ws, output time) and the video duration.
func Check(cues []Cue, ws []Word, o Options, duration float64) []Issue {
	var out []Issue
	out = append(out, reading(cues, o)...)
	out = append(out, timing(cues, duration)...)
	out = append(out, wordsCheck(cues, ws)...)
	out = append(out, seams(cues, ws)...)
	return out
}

func reading(cues []Cue, o Options) []Issue {
	var out []Issue
	longest, fastest, fastCue := 0, 0.0, 0
	for k, c := range cues {
		if len(c.Lines) > o.MaxLines {
			out = append(out, Issue{Check: "reading", Status: "warn", At: c.Start, Cue: k + 1,
				Detail: fmt.Sprintf("cue %d has %d lines (limit %d): %q", k+1, len(c.Lines), o.MaxLines, c.Text())})
		}
		for _, l := range c.Lines {
			n := utf8.RuneCountInString(l)
			longest = max(longest, n)
			if n > o.MaxChars {
				out = append(out, Issue{Check: "reading", Status: "warn", At: c.Start, Cue: k + 1,
					Detail: fmt.Sprintf("cue %d line is %d chars (limit %d): %q", k+1, n, o.MaxChars, l)})
			}
		}
		if d := c.End - c.Start; d > 0 {
			cps := float64(utf8.RuneCountInString(strings.ReplaceAll(c.Text(), " ", ""))) / d
			if cps > fastest {
				fastest, fastCue = cps, k+1
			}
			if cps > o.MaxCPS {
				out = append(out, Issue{Check: "reading", Status: "warn", At: c.Start, Cue: k + 1,
					Detail: fmt.Sprintf("cue %d reads at %.0f chars/s (limit %.0f): %q — the speech is fast here", k+1, cps, o.MaxCPS, c.Text()),
					Fix:    "usually fine for fast speech; a shorter cue limit doesn't slow the speaker down"})
			}
		}
	}
	if len(out) == 0 {
		out = append(out, Issue{Check: "reading", Status: "pass",
			Detail: fmt.Sprintf("longest line %d chars (limit %d), fastest cue %d at %.0f chars/s (limit %.0f)", longest, o.MaxChars, fastCue, fastest, o.MaxCPS)})
	}
	return out
}

func timing(cues []Cue, duration float64) []Issue {
	var out []Issue
	for k, c := range cues {
		switch {
		case c.End <= c.Start:
			out = append(out, Issue{Check: "timing", Status: "fail", At: c.Start, Cue: k + 1, Detail: fmt.Sprintf("cue %d ends before it starts (%s → %s)", k+1, Clock(c.Start), Clock(c.End))})
		case k > 0 && c.Start < cues[k-1].End-0.0005:
			out = append(out, Issue{Check: "timing", Status: "fail", At: c.Start, Cue: k + 1, Detail: fmt.Sprintf("cue %d starts at %s, before cue %d ends (%s)", k+1, Clock(c.Start), k, Clock(cues[k-1].End))})
		case duration > 0 && c.End > duration+0.0005:
			out = append(out, Issue{Check: "timing", Status: "fail", At: c.Start, Cue: k + 1, Detail: fmt.Sprintf("cue %d ends at %s, after the video (%s)", k+1, Clock(c.End), Clock(duration)),
				Fix: "the file is for a longer edit: `cutroom captions` writes one for this timeline"})
		}
	}
	if len(out) == 0 {
		out = append(out, Issue{Check: "timing", Status: "pass", Detail: fmt.Sprintf("%d cues in order, no overlaps, all inside the %.3f s video", len(cues), duration)})
	}
	return out
}

// wordsCheck: the file's words are the edit's words, in order, at about the right time.
func wordsCheck(cues []Cue, ws []Word) []Issue {
	type tw struct {
		w string
		t float64
	}
	var got []tw
	for _, c := range cues {
		for _, f := range strings.Fields(c.Text()) {
			if n := speech.Norm(f); n != "" {
				got = append(got, tw{n, c.Start})
			}
		}
	}
	var want []tw
	for _, w := range ws {
		if n := speech.Norm(w.Text); n != "" {
			want = append(want, tw{n, w.Start})
		}
	}
	n, m := len(want), len(got)
	dp := make([][]int32, n+1)
	for i := range dp {
		dp[i] = make([]int32, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if want[i].w == got[j].w {
				dp[i][j] = dp[i+1][j+1] + 1
			} else {
				dp[i][j] = max(dp[i+1][j], dp[i][j+1])
			}
		}
	}
	var missing, extra []string
	at := -1.0
	i, j := 0, 0
	for i < n || j < m {
		switch {
		case i < n && j < m && want[i].w == got[j].w:
			i, j = i+1, j+1
		case j >= m || (i < n && dp[i+1][j] >= dp[i][j+1]):
			if at < 0 {
				at = want[i].t
			}
			missing = append(missing, want[i].w)
			i++
		default:
			if at < 0 {
				at = got[j].t
			}
			extra = append(extra, got[j].w)
			j++
		}
	}
	match := int(dp[0][0])
	if len(missing)+len(extra) == 0 {
		return []Issue{{Check: "words", Status: "pass", Detail: fmt.Sprintf("all %d words of the edit, in order", n)}}
	}
	st := "warn"
	if n == 0 || float64(match) < 0.9*float64(n) || len(extra) >= 3 {
		st = "fail"
	}
	d := fmt.Sprintf("%d of %d words of the edit in order", match, n)
	if len(missing) > 0 {
		d += fmt.Sprintf("; not in the file: %q", clip(missing))
	}
	if len(extra) > 0 {
		d += fmt.Sprintf("; in the file but not the edit: %q", clip(extra))
	}
	return []Issue{{Check: "words", Status: st, At: at, Detail: d,
		Fix: "a file made before the last edit is stale: `cutroom captions` writes it again from the timeline"}}
}

func clip(s []string) string {
	if len(s) > 8 {
		return strings.Join(s[:8], " ") + fmt.Sprintf(" … (+%d)", len(s)-8)
	}
	return strings.Join(s, " ")
}

// seams (L49): sentences broken by cuts, and cues that open on a stranded sentence tail.
func seams(cues []Cue, ws []Word) []Issue {
	var out []Issue
	cuts := 0
	for k := 1; k < len(ws); k++ {
		if !ws[k].AfterCut {
			continue
		}
		cuts++
		prev, next := ws[k-1], ws[k]
		switch {
		case !endsSentence(prev.Text) && startsSentence(next.Text) && !strings.HasSuffix(prev.Text, ","):
			out = append(out, Issue{Check: "seams", Status: "warn", At: next.Start,
				Detail: fmt.Sprintf("the sentence before the cut at %s stops without ending: %q | %q", Clock(next.Start), tailText(ws, k-1, 5), headText(ws, k, 4)),
				Fix:    fmt.Sprintf("cut the unfinished sentence too, or keep its end: `cutroom timeline --joins` (segment %s)", next.Segment)})
		case endsSentence(prev.Text) && !startsSentence(next.Text) && !isI(next.Text):
			out = append(out, Issue{Check: "seams", Status: "warn", At: next.Start,
				Detail: fmt.Sprintf("after the cut at %s the text opens mid-sentence (its start was cut): %q", Clock(next.Start), headText(ws, k, 5)),
				Fix:    fmt.Sprintf("keep the sentence's start or cut its tail too: `cutroom timeline --joins` (segment %s)", next.Segment)})
		}
	}
	// stranded tails in the file: a cue opening with 1–2 words that end the previous cue's sentence
	for k := 1; k < len(cues); k++ {
		f := strings.Fields(cues[k].Text())
		if endsSentence(cues[k-1].Lines[len(cues[k-1].Lines)-1]) {
			continue
		}
		for n := 0; n < len(f) && n < 2; n++ {
			if endsSentence(f[n]) {
				out = append(out, Issue{Check: "seams", Status: "warn", At: cues[k].Start, Cue: k + 1,
					Detail: fmt.Sprintf("cue %d opens on a stranded sentence tail %q (the sentence starts in cue %d)", k+1, strings.Join(f[:n+1], " "), k),
					Fix:    "move those words to the end of the cue before (or split the sentence earlier)"})
				break
			}
		}
	}
	if len(out) == 0 {
		out = append(out, Issue{Check: "seams", Status: "pass", Detail: fmt.Sprintf("%d cut(s): no cue opens on a sentence tail, every sentence before a cut ends", cuts)})
	}
	return out
}

func isI(w string) bool { return w == "I" || strings.HasPrefix(w, "I'") }

func tailText(ws []Word, k, n int) string {
	var p []string
	for i := max(0, k-n+1); i <= k; i++ {
		p = append(p, ws[i].Text)
	}
	return strings.Join(p, " ")
}

func headText(ws []Word, k, n int) string {
	var p []string
	for i := k; i < len(ws) && i < k+n; i++ {
		p = append(p, ws[i].Text)
	}
	return strings.Join(p, " ")
}

// Clock formats seconds as m:ss.mmm, like the rest of cutroom's output.
func Clock(s float64) string {
	ms := int64(math.Round(s * 1000))
	return fmt.Sprintf("%d:%02d.%03d", ms/60000, ms/1000%60, ms%1000)
}
