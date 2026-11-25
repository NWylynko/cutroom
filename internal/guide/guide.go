// Package guide embeds the agent manual. Topics are markdown files in topics/; the first line
// of each is "<!-- summary: … -->". The manual is versioned with the binary (SPEC rule 10).
package guide

import (
	"embed"
	"sort"
	"strings"
)

//go:embed topics/*.md
var files embed.FS

type Topic struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Body    string `json:"body,omitempty"`
}

// Order is the reading order; topics not listed sort after, alphabetically.
var Order = []string{"start", "workflow", "perceive", "timeline", "edit", "overlays", "render", "verify", "captions", "shorts", "pitfalls"}

func All() []Topic {
	ents, _ := files.ReadDir("topics")
	var out []Topic
	for _, e := range ents {
		name := strings.TrimSuffix(e.Name(), ".md")
		t, _ := Get(name)
		out = append(out, t)
	}
	idx := func(n string) int {
		for i, o := range Order {
			if o == n {
				return i
			}
		}
		return len(Order)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if idx(out[i].Name) != idx(out[j].Name) {
			return idx(out[i].Name) < idx(out[j].Name)
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func Get(name string) (Topic, bool) {
	b, err := files.ReadFile("topics/" + name + ".md")
	if err != nil {
		return Topic{}, false
	}
	body := string(b)
	summary := ""
	if first, rest, ok := strings.Cut(body, "\n"); ok && strings.HasPrefix(first, "<!-- summary:") {
		summary = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(first, "<!-- summary:"), "-->"))
		body = strings.TrimLeft(rest, "\n")
	}
	return Topic{Name: name, Summary: summary, Body: strings.TrimRight(body, "\n")}, true
}

func Names() []string {
	var n []string
	for _, t := range All() {
		n = append(n, t.Name)
	}
	return n
}
