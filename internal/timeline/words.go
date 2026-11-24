package timeline

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/cutroom/cutroom/internal/speech"
)

// Transcripts maps source id → transcript (only sources that have one).
type Transcripts map[string]*speech.Transcript

// keptWord reports whether a word is (mostly) in the timeline.
func (t *Timeline) keptWord(src string, w speech.Word) bool {
	return t.Kept(src, (w.Start+w.End)/2)
}

// WordCutResult describes a cut made by words.
type WordCutResult struct {
	Source    string   `json:"source"`
	From      float64  `json:"from"`
	To        float64  `json:"to"`
	Removed   float64  `json:"removed_s"`
	Text      string   `json:"text"`
	Before    string   `json:"before"`       // kept words just before the join
	After     string   `json:"after"`        // kept words just after the join
	JoinPause float64  `json:"join_pause_s"` // pause at the new join, -1 if not between words
	Warnings  []string `json:"warnings"`
}

// CutWords removes words i..j (inclusive) of a source, cutting inside the surrounding pauses and
// leaving PadAfter after the previous kept word and PadBefore before the next one (L32).
func (t *Timeline) CutWords(src string, words []speech.Word, i, j int, reason string) WordCutResult {
	wi, wj := words[i], words[j]
	segI, _ := t.segAt(src, wi.Start, wi.End)
	segJ, _ := t.segAt(src, wj.Start, wj.End)
	r := WordCutResult{Source: src, Text: speech.Text(words, i, j), JoinPause: -1, Warnings: []string{}}
	prev, next := -1, -1
	for k := i - 1; k >= 0; k-- {
		if t.keptWord(src, words[k]) {
			prev = k
			break
		}
	}
	for k := j + 1; k < len(words); k++ {
		if t.keptWord(src, words[k]) {
			next = k
			break
		}
	}
	var a, b float64
	if prev >= 0 && segI != nil && words[prev].End >= segI.In {
		p := words[prev]
		a = clamp(p.End+t.PadAfter, p.End, wi.Start)
		if wi.Start-p.End < 0.12 {
			r.Warnings = append(r.Warnings, fmt.Sprintf("the cut starts inside continuous speech right after %q — `cutroom listen` the join", p.Text))
		}
	} else if segI != nil {
		a = segI.In
	} else {
		a = wi.Start - t.PadBefore
	}
	if next >= 0 && segJ != nil && words[next].Start < segJ.Out {
		n := words[next]
		b = clamp(n.Start-t.PadBefore, wj.End, n.Start)
		if n.Start-wj.End < 0.12 {
			r.Warnings = append(r.Warnings, fmt.Sprintf("the cut ends inside continuous speech right before %q — `cutroom listen` the join", n.Text))
		}
	} else if segJ != nil {
		b = segJ.Out
	} else {
		b = wj.End + t.PadAfter
	}
	r.From, r.To = t.Snap(a), t.Snap(b)
	r.Removed = t.Cut(src, a, b, reason, true)
	if prev >= 0 {
		r.Before = speech.Text(words, max(0, prev-3), prev)
	}
	if next >= 0 {
		r.After = speech.Text(words, next, min(len(words)-1, next+3))
	}
	if prev >= 0 && next >= 0 {
		r.JoinPause = t.KeptBetween(src, words[prev].End, words[next].Start)
	}
	return r
}

// KeptBetween is how much of source range [a, b) the timeline keeps (e.g. the pause at a join).
func (t *Timeline) KeptBetween(src string, a, b float64) float64 {
	k := 0.0
	for _, g := range t.Segments {
		if g.Source == src {
			k += math.Max(0, math.Min(b, g.Out)-math.Max(a, g.In))
		}
	}
	return round6(k)
}

func (t *Timeline) segAt(src string, a, b float64) (*Segment, int) {
	m := (a + b) / 2
	for i := range t.Segments {
		g := &t.Segments[i]
		if g.Source == src && m >= g.In && m < g.Out {
			return g, i
		}
	}
	return nil, -1
}

func clamp(x, lo, hi float64) float64 {
	if hi < lo {
		return (lo + hi) / 2
	}
	return math.Max(lo, math.Min(hi, x))
}

// TightenResult summarises a tighten pass.
type TightenResult struct {
	Pauses  int     `json:"pauses_shortened"`
	Edges   int     `json:"edges_trimmed"`
	Removed float64 `json:"removed_s"`
	Skipped int     `json:"pinned_edges_kept"`
	// Shortened lists each pause that was cut, by source time, with its length before and after.
	Shortened []PauseChange `json:"shortened"`
}

// PauseChange is one pause shortened by tighten.
type PauseChange struct {
	Source string  `json:"source"`
	At     float64 `json:"source_time"` // end of the word before the pause
	Before float64 `json:"before_s"`
	After  float64 `json:"after_s"`
}

// Tighten shortens every pause between kept words whose removable part (pause − pads) is at least
// minCut, and trims segment edges to the pads. Pinned edges never move (L39). [L38]
func (t *Timeline) Tighten(tr Transcripts, minCut float64) TightenResult {
	var res TightenResult
	before := t.Duration()
	type cut struct {
		src       string
		a, b, at  float64
		gap, left float64
	}
	var cuts []cut
	for _, g := range append([]Segment(nil), t.Segments...) {
		tx := tr[g.Source]
		if tx == nil {
			continue
		}
		var ws []speech.Word
		for _, w := range tx.Words {
			m := (w.Start + w.End) / 2
			if m >= g.In && m < g.Out {
				ws = append(ws, w)
			}
		}
		if len(ws) == 0 {
			continue
		}
		for k := 0; k+1 < len(ws); k++ {
			gap := ws[k+1].Start - ws[k].End
			if gap-t.PadAfter-t.PadBefore >= minCut {
				a, b := ws[k].End+t.PadAfter, ws[k+1].Start-t.PadBefore
				cuts = append(cuts, cut{g.Source, a, b, ws[k].End, gap, gap - (b - a)})
			}
		}
		// segment edges
		if in := ws[0].Start - t.PadBefore; in-g.In >= math.Min(minCut, 0.2) {
			if t.SetEdge(g.ID, true, in) {
				res.Edges++
			} else {
				res.Skipped++
			}
		}
		if out := ws[len(ws)-1].End + t.PadAfter; g.Out-out >= math.Min(minCut, 0.2) {
			if t.SetEdge(g.ID, false, out) {
				res.Edges++
			} else {
				res.Skipped++
			}
		}
	}
	for _, c := range cuts {
		if t.Cut(c.src, c.a, c.b, "", false) > 0 {
			res.Pauses++
			res.Shortened = append(res.Shortened, PauseChange{c.src, round6(c.at), round6(c.gap), round6(c.left)})
		}
	}
	res.Removed = round6(before - t.Duration())
	return res
}

// Issue is a validation finding. Level: "error" (render would fail or be wrong), "warn" (likely a
// mistake), "info".
type Issue struct {
	Level   string  `json:"level"`
	Code    string  `json:"code"`
	Segment string  `json:"segment,omitempty"`
	At      float64 `json:"output_time"`
	Message string  `json:"message"`
	Fix     string  `json:"fix,omitempty"`
}

// Validate checks the timeline against source durations and transcripts.
func (t *Timeline) Validate(durations map[string]float64, tr Transcripts) []Issue {
	var is []Issue
	add := func(l, code string, g *Segment, at float64, fix, f string, a ...any) {
		id := ""
		if g != nil {
			id = g.ID
		}
		is = append(is, Issue{Level: l, Code: code, Segment: id, At: round6(at), Message: fmt.Sprintf(f, a...), Fix: fix})
	}
	if len(t.Segments) == 0 {
		add("error", "empty", nil, 0, "`cutroom edit undo`, or `cutroom timeline reset`", "the timeline is empty")
		return is
	}
	used := map[string][][2]float64{}
	for i := range t.Segments {
		g := &t.Segments[i]
		d, ok := durations[g.Source]
		if !ok {
			add("error", "unknown-source", g, g.OutStart, "re-add the source with `cutroom init`", "segment %s uses unknown source %q", g.ID, g.Source)
			continue
		}
		if g.In < -1e-6 || g.Out > d+1/t.fpsNum+1e-6 || g.Out <= g.In {
			add("error", "out-of-source", g, g.OutStart, "`cutroom edit undo`", "segment %s range %.3f–%.3f is outside %s (0–%.3f)", g.ID, g.In, g.Out, g.Source, d)
		}
		if g.Out-g.In < 0.3 {
			add("warn", "short-segment", g, g.OutStart, "check it is intended: `cutroom timeline`", "segment %s is only %.2f s", g.ID, g.Out-g.In)
		}
		for _, r := range used[g.Source] {
			if g.In < r[1] && g.Out > r[0] {
				add("warn", "repeated-material", g, g.OutStart, "remove one of them", "segment %s repeats source material already used earlier (%.3f–%.3f)", g.ID, math.Max(g.In, r[0]), math.Min(g.Out, r[1]))
			}
		}
		used[g.Source] = append(used[g.Source], [2]float64{g.In, g.Out})
		tx := tr[g.Source]
		if tx == nil {
			continue
		}
		for _, w := range tx.Words {
			for _, edge := range []struct {
				at  float64
				nm  string
				out float64
			}{{g.In, "starts", g.OutStart}, {g.Out, "ends", g.OutStart + g.Out - g.In}} {
				if edge.at > w.Start+0.04 && edge.at < w.End-0.04 {
					add("warn", "edge-in-word", g, edge.out, "move the edge into a pause (cut by words with `cutroom edit cut --from/--to`), or `cutroom listen` it",
						"segment %s %s inside the word %q (%.3f–%.3f)", g.ID, edge.nm, w.Text, w.Start, w.End)
				}
			}
		}
	}
	// cuts that break a sentence: words removed from the middle of a sentence, not a retake
	for i := 0; i+1 < len(t.Segments); i++ {
		l, r := t.Segments[i], t.Segments[i+1]
		tx := tr[l.Source]
		if tx == nil || l.Source != r.Source || r.In <= l.Out {
			continue
		}
		var lw, fw *speech.Word
		var removed []string
		for k := range tx.Words {
			w := &tx.Words[k]
			m := (w.Start + w.End) / 2
			switch {
			case m >= l.In && m < l.Out:
				lw = w
			case m >= l.Out && m < r.In:
				removed = append(removed, w.Text)
			case m >= r.In && m < r.Out && fw == nil:
				fw = w
			}
		}
		if lw == nil || fw == nil || len(removed) == 0 {
			continue
		}
		retake := false
		for _, s := range tx.Suspects {
			if s.Kind == "repeat" && s.Start < r.In && s.End > l.Out {
				retake = true
			}
		}
		midLeft := !endsSentence(lw.Text)
		midRight := startsLower(fw.Text)
		if retake || (!midLeft && !midRight) {
			continue
		}
		add("warn", "broken-sentence", &t.Segments[i+1], r.OutStart,
			fmt.Sprintf("if unintended: `cutroom edit undo`, or cut whole sentences with --from/--to (check with `cutroom transcript --file` after rendering)"),
			"the cut at %.3f removes %q from inside a sentence: …%s | %s…", r.OutStart, strings.Join(removed, " "), lw.Text, fw.Text)
	}
	// retakes still in the cut: repeats whose attempts interleave (each first take comes before the
	// other's last take) are one retake — one warning, one cut
	for src, tx := range tr {
		type rep struct {
			s           speech.Suspect
			first, last float64
		}
		var reps []rep
		for _, s := range tx.Suspects {
			if s.Kind != "repeat" {
				continue
			}
			var takes []float64
			for _, at := range s.Related {
				if t.Kept(src, at+0.05) {
					takes = append(takes, at)
				}
			}
			if len(takes) >= 2 {
				reps = append(reps, rep{s, takes[0], takes[len(takes)-1]})
			}
		}
		sort.Slice(reps, func(a, b int) bool { return reps[a].first < reps[b].first })
		for len(reps) > 0 {
			g := []rep{reps[0]}
			var rest []rep
			for _, r := range reps[1:] {
				in := false
				for _, x := range g {
					if r.first < x.last && x.first < r.last {
						in = true
					}
				}
				if in {
					g = append(g, r)
				} else {
					rest = append(rest, r)
				}
			}
			reps = rest
			var sus []speech.Suspect
			for _, r := range g {
				sus = append(sus, r.s)
			}
			o, _ := t.ToOutput(src, g[0].first+0.05)
			fix := "cut all but one take: `cutroom transcript --search \"" + g[0].s.Text + "\"`"
			if cmd, gone := t.retakeCut(tx, src, sus); cmd != "" {
				if len(tr) > 1 {
					cmd += " --source " + src
				}
				fix = "keep the last take: `" + cmd + "` removes the earlier take, " + gone + " (--near is a source time; add --dry-run to check the join first)"
			}
			msg := fmt.Sprintf("%q is still said %d× in the timeline (a retake?)", g[0].s.Text, len(g[0].s.Related))
			if len(g) > 1 {
				msg = fmt.Sprintf("%q and %d more phrase(s) are said again later (one retake?)", g[0].s.Text, len(g)-1)
			}
			add("warn", "retake-kept", nil, o, fix, "%s", msg)
		}
	}
	// other suspects still in the cut
	for src, tx := range tr {
		for _, s := range tx.Suspects {
			switch s.Kind {
			case "repeat":
				// grouped into retakes below
			case "sound":
				if o, ok := t.ToOutput(src, (s.Start+s.End)/2); ok {
					add("warn", "kept-sound", nil, o, fmt.Sprintf("`cutroom listen %.2f-%.2f`; cut it with `cutroom edit cut %.2f-%.2f` if unwanted", s.Start-0.5, s.End+0.5, s.Start, s.End),
						"a sound with no words is kept at source %.3f–%.3f (cough, bump, or missed words?)", s.Start, s.End)
				}
			}
		}
	}
	return is
}

// Join is the boundary between two consecutive segments, as an agent should check it.
type Join struct {
	At     float64 `json:"output_time"`
	Before string  `json:"before"`
	After  string  `json:"after"`
	Pause  float64 `json:"pause_s"` // silence across the join, -1 if unknown
	Kind   string  `json:"kind"`    // "cut", or "zoom" where the picture continues and only the zoom changes
	Left   string  `json:"left_segment"`
	Right  string  `json:"right_segment"`
}

// Joins lists every cut point with the words on each side.
func (t *Timeline) Joins(tr Transcripts) []Join {
	var js []Join
	for i := 0; i+1 < len(t.Segments); i++ {
		l, r := t.Segments[i], t.Segments[i+1]
		j := Join{At: r.OutStart, Left: l.ID, Right: r.ID, Pause: -1, Kind: "cut"}
		if l.Source == r.Source && math.Abs(r.In-l.Out) < 1e-6 {
			j.Kind = "zoom" // continuous: only the zoom changes here, nothing is cut
		}
		var lastEnd, firstStart float64 = -1, -1
		if tx := tr[l.Source]; tx != nil {
			var ws []string
			for _, w := range tx.Words {
				if w.End <= l.Out+0.04 && w.Start >= l.In-0.04 {
					ws = append(ws, w.Text)
					lastEnd = w.End
				}
			}
			j.Before = tail(ws, 5)
		}
		if tx := tr[r.Source]; tx != nil {
			var ws []string
			for _, w := range tx.Words {
				if w.Start >= r.In-0.04 && w.End <= r.Out+0.04 {
					if firstStart < 0 {
						firstStart = w.Start
					}
					ws = append(ws, w.Text)
					if len(ws) == 5 {
						break
					}
				}
			}
			j.After = strings.Join(ws, " ")
		}
		if lastEnd >= 0 && firstStart >= 0 {
			j.Pause = round6(math.Max(0, l.Out-lastEnd) + math.Max(0, firstStart-r.In))
		}
		js = append(js, j)
	}
	return js
}

func endsSentence(w string) bool {
	w = strings.TrimRight(w, "\"')”’")
	return strings.HasSuffix(w, ".") || strings.HasSuffix(w, "?") || strings.HasSuffix(w, "!")
}

func startsLower(w string) bool {
	for _, r := range w {
		return r >= 'a' && r <= 'z'
	}
	return false
}

func tail(ws []string, n int) string {
	if len(ws) > n {
		ws = ws[len(ws)-n:]
	}
	return strings.Join(ws, " ")
}

// retakeCut builds a ready-to-run `edit cut` that removes the first kept take of a repeat, from its
// first words up to the last word before the next kept take, in the --from/--near/--to form the CLI
// resolves. The --to phrase is checked to resolve to exactly that word; "" if none does.
func (t *Timeline) retakeCut(tx *speech.Transcript, src string, group []speech.Suspect) (cmd, removes string) {
	ws := tx.Words
	kept := func(k int) bool { return t.Kept(src, (ws[k].Start+ws[k].End)/2) }
	// a kept occurrence of phrase nearest time at (as the CLI picks with --near)
	nearest := func(phrase string, at float64) (int, int, float64) {
		fi, fl, bd := -1, -1, math.Inf(1)
		for _, m := range speech.Search(ws, phrase) {
			ok := true
			for k := m.First; k <= m.Last; k++ {
				ok = ok && kept(k)
			}
			if d := math.Abs(m.Start - at); ok && d < bd {
				fi, fl, bd = m.First, m.Last, d
			}
		}
		return fi, fl, bd
	}
	type occ struct{ fi, fl, ni int }
	var os []occ
	for _, s := range group {
		var takes []float64
		for _, at := range s.Related {
			if t.Kept(src, at+0.05) {
				takes = append(takes, at)
			}
		}
		if len(takes) < 2 {
			continue
		}
		fi, fl, bd := nearest(s.Text, takes[0])
		if fi < 0 || bd > 0.5 {
			continue
		}
		ni, _, bd2 := nearest(s.Text, takes[1])
		if bd2 > 0.5 || ni <= fl {
			ni = -1
		}
		os = append(os, occ{fi, fl, ni})
	}
	if len(os) == 0 {
		return "", ""
	}
	// An editor cuts the whole first attempt: from the start of the sentence that holds its first
	// repeated phrase to the start of the sentence where the next attempt begins (T3: real retakes
	// repeat a sentence's tail, and one retake repeats several phrases).
	from, fromLast, firstLast, next := len(ws), -1, -1, len(ws)
	for _, o := range os {
		from = min(from, sentenceStart(ws, o.fi, kept))
		firstLast = max(firstLast, o.fl)
		if o.ni >= 0 {
			next = min(next, sentenceStart(ws, o.ni, kept))
		}
	}
	// …but never past the end of the first attempt: words after it may exist only in the next one
	j := min(next-1, sentenceEnd(ws, firstLast, kept))
	for j > firstLast && !kept(j) {
		j--
	}
	if next == len(ws) || j < firstLast || next <= from {
		// no clean sentence boundary: cut from the first repeated phrase to just before the next take
		o := os[0]
		from, j = o.fi, -1
		firstLast = o.fl
		at := group[0].Related
		for k := o.fl + 1; k < len(ws) && len(at) > 1; k++ {
			if ws[k].Start >= at[len(at)-1]-0.05 {
				break
			}
			if kept(k) {
				j = k
			}
		}
		if j < 0 {
			return "", ""
		}
	}
	fromLast = min(from+4, j)
	fromText := speech.Text(ws, from, fromLast)
	for n := 1; n <= 6 && j-n+1 > from; n++ {
		phrase := speech.Text(ws, j-n+1, j)
		for _, m := range speech.Search(ws, phrase) {
			if m.First < from+1 || m.Last < fromLast {
				continue
			}
			all := true
			for k := m.First; k <= m.Last; k++ {
				all = all && kept(k)
			}
			if !all {
				continue
			}
			if m.Last == j && len(strings.Fields(phrase)) >= 2 || m.Last == j && n == j-from {
				return fmt.Sprintf("cutroom edit cut --from %q --near %.2f --to %q", fromText, ws[from].Start, phrase),
					fmt.Sprintf("source %s–%s %q", mss(ws[from].Start), mss(ws[j].End), speech.Text(ws, from, j))
			}
			break // first match after --from is the wrong one: use more words
		}
	}
	return "", ""
}

// sentenceEnd walks forward from word k to the last word of its sentence (same stops as sentenceStart).
func sentenceEnd(ws []speech.Word, k int, kept func(int) bool) int {
	for k+1 < len(ws) && kept(k+1) && !endsSentence(ws[k].Text) && ws[k+1].Start-ws[k].End < 1.0 {
		k++
	}
	return k
}

// sentenceStart walks back from word k to the first word of its sentence: stop after a word that
// ends a sentence, at a pause of a second or more, or at a word the timeline doesn't keep.
func sentenceStart(ws []speech.Word, k int, kept func(int) bool) int {
	for k > 0 && kept(k-1) && !endsSentence(ws[k-1].Text) && ws[k].Start-ws[k-1].End < 1.0 {
		k--
	}
	return k
}

// mss formats seconds as m:ss.mmm.
func mss(x float64) string {
	ms := int64(math.Round(x * 1000))
	return fmt.Sprintf("%d:%02d.%03d", ms/60000, ms/1000%60, ms%1000)
}
