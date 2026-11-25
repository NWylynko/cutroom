// Package project loads and saves cutroom.json: the sources (id → file + probed facts) and settings.
// Source files are only ever read (L55).
package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/NWylynko/cutroom/internal/align"
	"github.com/NWylynko/cutroom/internal/media"
)

const File = "cutroom.json"

type Source struct {
	ID      string       `json:"id"`
	Path    string       `json:"path"` // relative to the project dir when inside it, else absolute
	Size    int64        `json:"size_bytes"`
	ModTime time.Time    `json:"mtime"`
	Facts   *media.Facts `json:"facts"`
	Sync    *align.Sync  `json:"sync,omitempty"` // how this source's audio lines up with another source
}

type Project struct {
	Version int                `json:"version"`
	Sources map[string]*Source `json:"sources"`
	Dir     string             `json:"-"`
}

var ErrNoProject = errors.New("no cutroom project here")

func Load(dir string) (*Project, error) {
	b, err := os.ReadFile(filepath.Join(dir, File))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNoProject
	}
	if err != nil {
		return nil, err
	}
	p := &Project{}
	if err := json.Unmarshal(b, p); err != nil {
		return nil, fmt.Errorf("%s is not valid JSON: %w", File, err)
	}
	p.Dir = absDir(dir)
	if p.Sources == nil {
		p.Sources = map[string]*Source{}
	}
	return p, nil
}

func New(dir string) *Project {
	return &Project{Version: 1, Sources: map[string]*Source{}, Dir: absDir(dir)}
}

func absDir(d string) string {
	a, err := filepath.Abs(d)
	if err != nil {
		return d
	}
	return a
}

func (p *Project) Save() error {
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(p.Dir, File+".tmp")
	if err := os.WriteFile(tmp, append(b, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(p.Dir, File))
}

// Abs returns the absolute path of a source.
func (p *Project) Abs(s *Source) string {
	if filepath.IsAbs(s.Path) {
		return s.Path
	}
	return filepath.Join(p.Dir, s.Path)
}

// RelPath stores a path relative to the project dir when it lives inside it.
func (p *Project) RelPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	dir, _ := filepath.Abs(p.Dir)
	if r, err := filepath.EvalSymlinks(dir); err == nil {
		dir = r
	}
	if r, err := filepath.EvalSymlinks(abs); err == nil {
		abs = r
	}
	if rel, err := filepath.Rel(dir, abs); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return abs
}

// IDs returns source ids sorted (videos first, then audio, then by id).
func (p *Project) IDs() []string {
	var ids []string
	for id := range p.Sources {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := p.Sources[ids[i]], p.Sources[ids[j]]
		ka, kb := kind(a), kind(b)
		if ka != kb {
			return ka < kb
		}
		return ids[i] < ids[j]
	})
	return ids
}

func kind(s *Source) string {
	if s.Facts != nil && s.Facts.Kind == "video" {
		return "0"
	}
	return "1"
}

// FindByPath returns the source that is the same file (symlinks and relative paths resolved).
func (p *Project) FindByPath(path string) *Source {
	want, err := os.Stat(path)
	if err != nil {
		return nil
	}
	for _, s := range p.Sources {
		if st, err := os.Stat(p.Abs(s)); err == nil && os.SameFile(want, st) {
			return s
		}
	}
	return nil
}

var idRe = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)

func ValidID(id string) bool { return idRe.MatchString(id) }

// SuggestID derives an id from a filename: lowercase stem, non-alnum → '-', unique in the project.
func (p *Project) SuggestID(path string) string {
	stem := strings.ToLower(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
	stem = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(stem, "-")
	stem = strings.Trim(stem, "-")
	if stem == "" || !(stem[0] >= 'a' && stem[0] <= 'z') {
		stem = "src-" + stem
	}
	if len(stem) > 24 {
		stem = strings.TrimRight(stem[:24], "-")
	}
	id := stem
	for n := 2; p.Sources[id] != nil; n++ {
		id = fmt.Sprintf("%s-%d", stem, n)
	}
	return id
}

// Stale reports whether a source file changed since it was probed (or is missing).
func (p *Project) Stale(s *Source) (bool, error) {
	st, err := os.Stat(p.Abs(s))
	if err != nil {
		return true, err
	}
	return st.Size() != s.Size || !st.ModTime().Equal(s.ModTime), nil
}
