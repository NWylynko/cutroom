package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/NWylynko/cutroom/internal/raster"
	"github.com/NWylynko/cutroom/internal/timeline"
)

// Look is a saved caption + card style: a channel's look, applied to any project in one call.
type Look struct {
	Captions *timeline.Style `json:"captions,omitempty"`
	Cards    *timeline.Style `json:"cards,omitempty"`
}

// looksDir is where named looks live: $CUTROOM_LOOKS, else <user config dir>/cutroom/looks.
func looksDir() string {
	if d := os.Getenv("CUTROOM_LOOKS"); d != "" {
		return d
	}
	d, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		d = filepath.Join(home, ".config")
	}
	return filepath.Join(d, "cutroom", "looks")
}

// lookPath: a name is a saved look; anything with a / or ending in .json is a file (to keep a look
// in a repo and share it).
func lookPath(name string) string {
	if strings.ContainsRune(name, '/') || strings.HasSuffix(name, ".json") {
		return userPath(name)
	}
	return filepath.Join(looksDir(), name+".json")
}

func saveLook(name string, tl *timeline.Timeline, projDir string) (string, error) {
	if tl.CaptionStyle.Empty() && tl.CardStyle.Empty() {
		return "", Errf(ExitUsage, "nothing-to-save", "set a look first: `cutroom style captions …` / `cutroom style cards …`", "this project uses cutroom's default look: nothing to save")
	}
	lk := Look{Captions: absFonts(tl.CaptionStyle, projDir), Cards: absFonts(tl.CardStyle, projDir)}
	p := lookPath(name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	b, _ := json.MarshalIndent(lk, "", "  ")
	return p, os.WriteFile(p, append(b, '\n'), 0o644)
}

// absFonts copies a style with font files given relative to the project made absolute, so the
// look works from any project.
func absFonts(s *timeline.Style, dir string) *timeline.Style {
	if s.Empty() {
		return nil
	}
	c := *s
	for _, f := range []*string{&c.Font, &c.SubFont} {
		if *f == "" {
			continue
		}
		if r, err := raster.ResolveFont(*f, dir); err == nil && r.Path != "" && looksLikeFontFile(*f) {
			*f = r.Path
			if r.Index > 0 {
				*f += fmt.Sprintf("#%d", r.Index)
			}
		}
	}
	return &c
}

func looksLikeFontFile(s string) bool {
	ext := strings.ToLower(filepath.Ext(strings.SplitN(s, "#", 2)[0]))
	return strings.ContainsRune(s, '/') || ext == ".ttf" || ext == ".otf" || ext == ".ttc"
}

func loadLook(name string) (Look, string, error) {
	p := lookPath(name)
	b, err := os.ReadFile(p)
	if err != nil {
		fix := "`cutroom style looks` lists saved looks; save one with `cutroom style save <name>`"
		if names := lookNames(); len(names) > 0 {
			if s := suggest(name, names); s != "" {
				fix = fmt.Sprintf("did you mean %q? (%s)", s, fix)
			}
		}
		return Look{}, p, Errf(ExitUsage, "no-look", fix, "no look %q (%s)", name, p)
	}
	var lk Look
	if err := json.Unmarshal(b, &lk); err != nil {
		return Look{}, p, Errf(ExitInvalid, "bad-look", "fix the JSON or save it again with `cutroom style save`", "%s: %v", p, err)
	}
	return lk, p, nil
}

func lookNames() []string {
	ents, _ := os.ReadDir(looksDir())
	var out []string
	for _, e := range ents {
		if strings.HasSuffix(e.Name(), ".json") {
			out = append(out, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	sort.Strings(out)
	return out
}

// applyLook replaces the project's caption and card style with the look's (cards' own styles stay).
func applyLook(tl *timeline.Timeline, lk Look) {
	tl.CaptionStyle, tl.CardStyle = nil, nil
	if !lk.Captions.Empty() {
		c := *lk.Captions
		tl.CaptionStyle = &c
	}
	if !lk.Cards.Empty() {
		c := *lk.Cards
		tl.CardStyle = &c
	}
}

func lookSummary(lk Look) string {
	var parts []string
	if !lk.Captions.Empty() {
		parts = append(parts, "captions: "+setSummary(lk.Captions))
	}
	if !lk.Cards.Empty() {
		parts = append(parts, "cards: "+setSummary(lk.Cards))
	}
	return strings.Join(parts, "; ")
}
