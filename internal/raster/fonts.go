package raster

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

// FontRef is a resolved font: a built-in face or one face of a font file.
type FontRef struct {
	Builtin string // regular | bold | mono
	Path    string // font file
	Index   int    // face in a .ttc collection
	Name    string // the face's full name (for display)
	Key     string // changes when the file does (for caches)
}

// Builtins are the faces cutroom ships (Go fonts), always available.
var Builtins = []string{"bold", "regular", "mono"}

// FontDirs are where installed fonts are looked up by name.
func FontDirs() []string {
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "darwin" {
		return []string{filepath.Join(home, "Library/Fonts"), "/Library/Fonts", "/System/Library/Fonts", "/System/Library/Fonts/Supplemental"}
	}
	return []string{filepath.Join(home, ".fonts"), filepath.Join(home, ".local/share/fonts"), "/usr/local/share/fonts", "/usr/share/fonts"}
}

// ResolveFont finds a font: a built-in name, a font file (absolute, ~/…, or relative to dir; name#2
// picks a face of a collection), or an installed font by file name or face name ("Arial Black",
// "Avenir Next Bold"). The error lists close matches.
func ResolveFont(spec, dir string) (FontRef, error) {
	if spec == "" {
		spec = "bold"
	}
	for _, b := range Builtins {
		if spec == b {
			return FontRef{Builtin: b, Name: "Go " + b, Key: "builtin:" + b}, nil
		}
	}
	path, idx := spec, 0
	if i := strings.LastIndex(spec, "#"); i > 0 {
		if n, err := strconv.Atoi(spec[i+1:]); err == nil {
			path, idx = spec[:i], n
		}
	}
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}
	if !filepath.IsAbs(path) && dir != "" {
		if _, err := os.Stat(filepath.Join(dir, path)); err == nil {
			path = filepath.Join(dir, path)
		}
	}
	if st, err := os.Stat(path); err == nil && !st.IsDir() {
		return fileRef(path, idx)
	}
	if looksLikePath(spec) {
		return FontRef{}, fmt.Errorf("font file %s not found", spec)
	}
	want := normName(spec)
	for _, f := range InstalledFonts() {
		if normName(f.Name) == want || normName(strings.TrimSuffix(filepath.Base(f.Path), filepath.Ext(f.Path))) == want && f.Index == 0 {
			return fileRef(f.Path, f.Index)
		}
	}
	var close []string
	for _, f := range InstalledFonts() {
		if strings.Contains(normName(f.Name), want) || strings.Contains(want, normName(firstWord(f.Name))) && len(firstWord(f.Name)) > 3 {
			close = append(close, f.Name)
		}
	}
	sort.Strings(close)
	close = dedupe(close)
	if len(close) > 8 {
		close = close[:8]
	}
	msg := fmt.Sprintf("no font %q: give bold, regular, mono, a font file, or an installed font's name (`cutroom style fonts` lists them)", spec)
	if len(close) > 0 {
		msg += "; close: " + strings.Join(close, ", ")
	}
	return FontRef{}, fmt.Errorf("%s", msg)
}

func looksLikePath(s string) bool {
	ext := strings.ToLower(filepath.Ext(strings.SplitN(s, "#", 2)[0]))
	return strings.ContainsRune(s, '/') || ext == ".ttf" || ext == ".otf" || ext == ".ttc"
}

func firstWord(s string) string { return strings.Fields(s + " ")[0] }

func dedupe(ss []string) []string {
	var out []string
	for i, s := range ss {
		if i == 0 || s != ss[i-1] {
			out = append(out, s)
		}
	}
	return out
}

func normName(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '-' || r == '_' {
			return -1
		}
		return r
	}, strings.ToLower(s))
}

func fileRef(path string, idx int) (FontRef, error) {
	c, err := loadCollection(path)
	if err != nil {
		return FontRef{}, fmt.Errorf("font %s: %w", path, err)
	}
	if idx < 0 || idx >= c.NumFonts() {
		return FontRef{}, fmt.Errorf("font %s has faces #0–#%d", path, c.NumFonts()-1)
	}
	f, _ := c.Font(idx)
	st, _ := os.Stat(path)
	return FontRef{Path: path, Index: idx, Name: faceName(f, path),
		Key: fmt.Sprintf("%s#%d:%d:%d", path, idx, st.Size(), st.ModTime().UnixNano())}, nil
}

var (
	collMu sync.Mutex
	colls  = map[string]*opentype.Collection{}
)

func loadCollection(path string) (*opentype.Collection, error) {
	collMu.Lock()
	defer collMu.Unlock()
	if c, ok := colls[path]; ok {
		return c, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	c, err := opentype.ParseCollection(b)
	if err != nil {
		return nil, fmt.Errorf("not a font cutroom can read (TrueType/OpenType): %w", err)
	}
	colls[path] = c
	return c, nil
}

func faceName(f *opentype.Font, path string) string {
	var b sfnt.Buffer
	if n, err := f.Name(&b, sfnt.NameIDFull); err == nil && n != "" {
		return n
	}
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

// InstalledFont is one face found in the font folders.
type InstalledFont struct {
	Name  string
	Path  string
	Index int
}

var (
	installedOnce sync.Once
	installed     []InstalledFont
)

// InstalledFonts lists every face in FontDirs (scanned once per run).
func InstalledFonts() []InstalledFont {
	installedOnce.Do(func() {
		for _, d := range FontDirs() {
			_ = filepath.WalkDir(d, func(p string, e os.DirEntry, err error) error {
				if err != nil || e.IsDir() {
					return nil
				}
				switch strings.ToLower(filepath.Ext(p)) {
				case ".ttf", ".otf", ".ttc":
				default:
					return nil
				}
				b, err := os.ReadFile(p)
				if err != nil {
					return nil
				}
				c, err := opentype.ParseCollection(b)
				if err != nil {
					return nil
				}
				for i := range c.NumFonts() {
					if f, err := c.Font(i); err == nil {
						installed = append(installed, InstalledFont{faceName(f, p), p, i})
					}
				}
				return nil
			})
		}
		sort.Slice(installed, func(i, j int) bool { return installed[i].Name < installed[j].Name })
	})
	return installed
}

var (
	refFaceMu sync.Mutex
	refFaces  = map[string]font.Face{}
)

// RefFace is the face of a resolved font at size px (cached).
func RefFace(r FontRef, size float64) font.Face {
	if r.Builtin != "" {
		return Face(r.Builtin, size)
	}
	k := fmt.Sprintf("%s@%g", r.Key, size)
	refFaceMu.Lock()
	defer refFaceMu.Unlock()
	if f, ok := refFaces[k]; ok {
		return f
	}
	c, err := loadCollection(r.Path)
	if err != nil {
		return Face("bold", size) // resolved earlier, so this only happens if the file vanished
	}
	f, err := c.Font(r.Index)
	if err != nil {
		return Face("bold", size)
	}
	ff, err := opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingNone})
	if err != nil {
		return Face("bold", size)
	}
	refFaces[k] = ff
	return ff
}
