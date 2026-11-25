package cli

import (
	"fmt"
	"strings"

	"github.com/NWylynko/cutroom/internal/guide"
)

func init() {
	Register(&Command{
		Name:     "guide",
		Args:     "[topic]",
		MaxArgs:  1,
		Category: "start",
		Summary:  "The agent manual for this version. Read `cutroom guide` before anything else",
		Help: `With no topic: the start page plus a list of topics.
Examples:
  cutroom guide
  cutroom guide workflow
  cutroom guide --all`,
		Flags: []Flag{{Name: "all", Kind: "bool", Help: "Print every topic in reading order"}},
		Run:   runGuide,
	})
}

func runGuide(c *Ctx) (*Result, error) {
	all := guide.All()
	if c.Bool("all") {
		var parts []string
		for _, t := range all {
			parts = append(parts, t.Body)
		}
		return &Result{Text: strings.Join(parts, "\n\n---\n\n"), Data: all}, nil
	}
	name := "start"
	if len(c.Args) == 1 {
		name = c.Args[0]
	}
	t, ok := guide.Get(name)
	if !ok {
		fix := "topics: " + strings.Join(guide.Names(), ", ")
		if s := suggest(name, guide.Names()); s != "" {
			fix = fmt.Sprintf("did you mean `cutroom guide %s`? (%s)", s, fix)
		}
		return nil, Errf(ExitUsage, "unknown-topic", fix, "no guide topic %q", name)
	}
	text := t.Body
	var next []string
	if name == "start" {
		var b strings.Builder
		b.WriteString("\n\nTopics:\n")
		for _, x := range all {
			if x.Name != "start" {
				fmt.Fprintf(&b, "  %-10s %s\n", x.Name, x.Summary)
			}
		}
		text += strings.TrimRight(b.String(), "\n")
		next = []string{"cutroom doctor", "cutroom guide workflow"}
	}
	return &Result{Text: text, Data: map[string]any{"topic": t, "topics": summaries(all)}, Next: next}, nil
}

func summaries(ts []guide.Topic) []guide.Topic {
	out := make([]guide.Topic, len(ts))
	for i, t := range ts {
		out[i] = guide.Topic{Name: t.Name, Summary: t.Summary}
	}
	return out
}
