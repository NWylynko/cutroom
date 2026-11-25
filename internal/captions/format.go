package captions

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Format renders cues as SRT, or WebVTT when vtt.
func Format(cues []Cue, vtt bool) string {
	var b strings.Builder
	if vtt {
		b.WriteString("WEBVTT\n\n")
	}
	for k, c := range cues {
		if !vtt {
			fmt.Fprintf(&b, "%d\n", k+1)
		}
		fmt.Fprintf(&b, "%s --> %s\n%s\n\n", Stamp(c.Start, vtt), Stamp(c.End, vtt), strings.Join(c.Lines, "\n"))
	}
	return b.String()
}

var stampRe = regexp.MustCompile(`^\s*(?:(\d+):)?(\d{1,2}):(\d{2})[,.](\d{3})\s*-->\s*(?:(\d+):)?(\d{1,2}):(\d{2})[,.](\d{3})`)

// Parse reads SRT or WebVTT. Cue settings after the times, numbering and VTT notes are ignored;
// tags like <i> are stripped.
func Parse(s string) ([]Cue, error) {
	s = strings.ReplaceAll(strings.TrimPrefix(s, "\ufeff"), "\r\n", "\n")
	var cues []Cue
	blocks := strings.Split(s, "\n\n")
	for _, blk := range blocks {
		lines := strings.Split(strings.Trim(blk, "\n"), "\n")
		at := -1
		for i, l := range lines {
			if strings.Contains(l, "-->") {
				at = i
				break
			}
		}
		if at < 0 {
			continue // WEBVTT header, NOTE, STYLE, or blank
		}
		m := stampRe.FindStringSubmatch(lines[at])
		if m == nil {
			return nil, fmt.Errorf("cue %d: can't read the times in %q", len(cues)+1, lines[at])
		}
		c := Cue{Start: hms(m[1], m[2], m[3], m[4]), End: hms(m[5], m[6], m[7], m[8]), First: -1, Last: -1}
		for _, l := range lines[at+1:] {
			if t := strings.TrimSpace(tagRe.ReplaceAllString(l, "")); t != "" {
				c.Lines = append(c.Lines, t)
			}
		}
		cues = append(cues, c)
	}
	if len(cues) == 0 {
		return nil, fmt.Errorf("no cues found (expected SRT or WebVTT: `00:00:01,000 --> 00:00:02,500` lines)")
	}
	return cues, nil
}

var tagRe = regexp.MustCompile(`<[^>]*>`)

func hms(h, m, s, ms string) float64 {
	a, _ := strconv.Atoi(h)
	b, _ := strconv.Atoi(m)
	c, _ := strconv.Atoi(s)
	d, _ := strconv.Atoi(ms)
	return float64(a*3600+b*60+c) + float64(d)/1000
}
