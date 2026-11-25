package cli

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/NWylynko/cutroom/internal/timecode"
	"github.com/NWylynko/cutroom/internal/timeline"
)

// pickIssue finds the warning a `code[@time]` names.
func pickIssue(issues []timeline.Issue, spec string, wantLevel ...string) (timeline.Issue, error) {
	code, at, hasAt := spec, 0.0, false
	if i := strings.LastIndex(spec, "@"); i > 0 {
		code = spec[:i]
		v, err := timecode.Parse(spec[i+1:], timecode.Rate{Num: 30, Den: 1})
		if err != nil {
			return timeline.Issue{}, Errf(ExitUsage, "bad-time", "give the output time `cutroom timeline` prints, e.g. retake-kept@4.09 or @0:04.093", "%v", err)
		}
		at, hasAt = v.Float(), true
	}
	var cands []timeline.Issue
	var codes []string
	for _, is := range issues {
		if !contains(codes, is.Code) {
			codes = append(codes, is.Code)
		}
		if is.Code == code && (len(wantLevel) == 0 || contains(wantLevel, is.Level)) {
			cands = append(cands, is)
		}
	}
	if len(cands) == 0 {
		fix := "`cutroom timeline` lists the warnings"
		if s := suggest(code, codes); s != "" {
			fix = fmt.Sprintf("did you mean %s? (%s)", s, fix)
		}
		return timeline.Issue{}, Errf(ExitUsage, "no-such-warning", fix, "no open %q warning in the timeline", code)
	}
	if !hasAt {
		if len(cands) > 1 {
			var ats []string
			for _, x := range cands {
				ats = append(ats, fmt.Sprintf("%s@%.2f", code, x.At))
			}
			return timeline.Issue{}, Errf(ExitUsage, "ambiguous-warning", "name one: "+strings.Join(ats, ", "), "%d %q warnings", len(cands), code)
		}
		return cands[0], nil
	}
	best, bd := cands[0], math.Inf(1)
	for _, x := range cands {
		if d := math.Abs(x.At - at); d < bd {
			best, bd = x, d
		}
	}
	if bd > 0.5 {
		if len(cands) == 1 { // the only one of its kind: an earlier edit moved it — it's this one
			return best, nil
		}
		var ats []string
		for _, x := range cands {
			ats = append(ats, fmt.Sprintf("%s@%.2f", code, x.At))
		}
		return timeline.Issue{}, Errf(ExitUsage, "no-such-warning", "edits move warnings in output time; they are now: "+strings.Join(ats, ", "), "no %q warning near output %s", code, clock(at))
	}
	return best, nil
}

var fixCmd = regexp.MustCompile("`(cutroom (?:edit|overlay) [^`]+)`")

// splitArgs splits a command line with double/single quotes (no escapes inside quotes needed here).
func splitArgs(s string) []string {
	var out []string
	var cur strings.Builder
	q, in := rune(0), false
	for _, r := range s {
		switch {
		case q != 0 && r == q:
			q = 0
		case q == 0 && (r == '"' || r == '\''):
			q, in = r, true
		case q == 0 && r == ' ':
			if in {
				out = append(out, cur.String())
				cur.Reset()
				in = false
			}
		default:
			cur.WriteRune(r)
			in = true
		}
	}
	if in {
		out = append(out, cur.String())
	}
	return out
}

func decideIssue(c *Ctx, e *editEnv) (*Result, error) {
	issues := e.validate()
	if c.Has("apply") {
		is, err := pickIssue(issues, c.Str("apply"), "warn", "info", "accepted")
		if err != nil {
			return nil, err
		}
		m := fixCmd.FindStringSubmatch(is.Fix)
		if m == nil || strings.Contains(m[1], "<") {
			return nil, Errf(ExitUsage, "no-runnable-fix", "this one needs your judgement: "+orStr(is.Fix, "see `cutroom timeline`")+"; or keep it: `cutroom timeline --accept "+c.Str("apply")+" --reason \"…\"`",
				"the %s warning has no single edit to run", is.Code)
		}
		args := splitArgs(m[1])[1:]
		if c.Bool("dry-run") {
			args = append(args, "--dry-run")
		}
		if c.Dir != "" {
			args = append(args, "--project", c.Dir)
		}
		res, err := dispatch(args, c.Out, c.ErrW, c.Stdin, c.JSON)
		if err != nil {
			return nil, err
		}
		res.Text = fmt.Sprintf("applying the %s fix: %s\n%s", is.Code, m[1], res.Text)
		return res, nil
	}
	is, err := pickIssue(issues, c.Str("accept"), "warn", "info")
	if err != nil {
		for _, x := range issues {
			if x.Level == "error" && strings.HasPrefix(c.Str("accept"), x.Code) {
				return nil, Errf(ExitUsage, "cannot-accept", orStr(x.Fix, "`cutroom timeline`"), "%s is an error: the render can't be made until it is fixed", x.Code)
			}
		}
		return nil, err
	}
	if !c.Has("reason") {
		return nil, Errf(ExitUsage, "missing-flag", fmt.Sprintf(`say why it's fine: cutroom timeline --accept %s --reason "…"`, c.Str("accept")), "--accept needs --reason")
	}
	e.tl.Accepted = append(e.tl.Accepted, timeline.Accepted{Code: is.Code, Key: timeline.IssueKey(is), Reason: c.Str("reason")})
	if err := timeline.Checkpoint(e.p.Dir, cacheDir(e.p), fmt.Sprintf("timeline --accept %s", c.Str("accept"))); err != nil {
		return nil, err
	}
	if err := e.tl.Save(e.p.Dir); err != nil {
		return nil, err
	}
	open := 0
	for _, x := range e.validate() {
		if x.Level == "warn" || x.Level == "error" {
			open++
		}
	}
	return &Result{Text: fmt.Sprintf("accepted %s at output %s (%q): listed by `cutroom timeline`, no longer an open warning; %d still open\n(undo: `cutroom edit undo`)", is.Code, clock(is.At), c.Str("reason"), open),
		Data: map[string]any{"accepted": is, "open": open}, Next: []string{"cutroom status"}}, nil
}
