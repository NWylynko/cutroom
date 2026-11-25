package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func run(args ...string) (string, string, int) {
	var o, e bytes.Buffer
	code := Run(args, &o, &e, strings.NewReader(""))
	return o.String(), e.String(), code
}

func TestOverviewPointsToGuide(t *testing.T) {
	out, _, code := run()
	if code != 0 || !strings.Contains(out, "cutroom guide") {
		t.Fatalf("overview must point agents to the guide; code=%d out=%s", code, out)
	}
}

func TestUnknownCommandSuggests(t *testing.T) {
	_, errOut, code := run("gude")
	if code != ExitUsage || !strings.Contains(errOut, "did you mean `cutroom guide`") {
		t.Fatalf("code=%d err=%s", code, errOut)
	}
	if !strings.HasPrefix(errOut, "error[unknown-command]:") || !strings.Contains(errOut, "\nfix: ") {
		t.Fatalf("errors must be error[code] + fix: lines; got %q", errOut)
	}
}

func TestUnknownFlagSuggests(t *testing.T) {
	_, errOut, code := run("guide", "--al")
	if code != ExitUsage || !strings.Contains(errOut, "did you mean --all?") {
		t.Fatalf("code=%d err=%s", code, errOut)
	}
}

func TestShortFlagsRejectedWithFix(t *testing.T) {
	_, errOut, code := run("guide", "-a")
	if code != ExitUsage || !strings.Contains(errOut, "--a") {
		t.Fatalf("code=%d err=%s", code, errOut)
	}
}

func TestJSONEnvelope(t *testing.T) {
	out, _, code := run("guide", "--json")
	var v map[string]any
	if err := json.Unmarshal([]byte(out), &v); err != nil || code != 0 {
		t.Fatalf("not JSON (%v): %s", err, out)
	}
	if v["ok"] != true || v["result"] == nil || v["next"] == nil {
		t.Fatalf("envelope must have ok/result/next: %v", v)
	}
	out, _, code = run("nope", "--json")
	if err := json.Unmarshal([]byte(out), &v); err != nil || code != ExitUsage {
		t.Fatalf("errors must be JSON too in --json mode: %s", out)
	}
	if e, _ := v["error"].(map[string]any); e["code"] != "unknown-command" || e["fix"] == "" {
		t.Fatalf("error object: %v", v)
	}
}

func TestGuideTopics(t *testing.T) {
	out, _, code := run("guide")
	if code != 0 || !strings.Contains(out, "Topics:") || !strings.Contains(out, "pitfalls") {
		t.Fatalf("guide start must list topics: %s", out)
	}
	_, errOut, code := run("guide", "pitfall")
	if code != ExitUsage || !strings.Contains(errOut, "did you mean `cutroom guide pitfalls`") {
		t.Fatalf("%d %s", code, errOut)
	}
}

func TestEveryCommandHasSummaryAndHelp(t *testing.T) {
	for _, c := range Commands() {
		if c.Summary == "" || c.Category == "" {
			t.Errorf("%s: needs Summary and Category", c.Name)
		}
		out, _, code := run(c.Name, "--help")
		if code != 0 || !strings.Contains(out, "cutroom "+c.Name) {
			t.Errorf("%s --help: %d %s", c.Name, code, out)
		}
	}
}

func TestSuggest(t *testing.T) {
	cands := []string{"transcribe", "transcript", "timeline", "render"}
	for in, want := range map[string]string{"transcibe": "transcribe", "rendr": "render", "zzzzzz": "", "timelin": "timeline"} {
		if got := suggest(in, cands); got != want {
			t.Errorf("suggest(%q)=%q want %q", in, got, want)
		}
	}
}

func TestParseSize(t *testing.T) {
	for _, c := range []struct {
		in         string
		cw, ch     int
		w, h       int
		shouldFail bool
	}{{"1080p", 3840, 2160, 1920, 1080, false}, {"1080p", 1080, 1920, 1080, 1920, false}, {"4K", 1920, 1080, 3840, 2160, false},
		{"720x1280", 1920, 1080, 720, 1280, false}, {"1081x3", 1920, 1080, 0, 0, true}} {
		w, h, err := parseSize(c.in, c.cw, c.ch)
		if (err != nil) != c.shouldFail || w != c.w || h != c.h {
			t.Errorf("%s from %dx%d: %dx%d %v", c.in, c.cw, c.ch, w, h, err)
		}
	}
}
