package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindByPathWithRelativeProjectDir(t *testing.T) {
	d := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	os.Chdir(d)
	os.WriteFile("a.mp4", []byte("x"), 0o644)
	p := New(".")
	p.Sources["a"] = &Source{ID: "a", Path: p.RelPath("a.mp4")}
	if p.Sources["a"].Path != "a.mp4" {
		t.Fatalf("path stored as %q, want relative", p.Sources["a"].Path)
	}
	if p.FindByPath("a.mp4") == nil || p.FindByPath(filepath.Join(d, "a.mp4")) == nil {
		t.Fatal("same file not found by relative or absolute path")
	}
	if err := p.Save(); err != nil {
		t.Fatal(err)
	}
	q, err := Load(".")
	if err != nil || q.FindByPath("a.mp4") == nil {
		t.Fatalf("after reload: %v", err)
	}
}

func TestSuggestID(t *testing.T) {
	p := New(t.TempDir())
	for in, want := range map[string]string{"talk.mp4": "talk", "IMG_0412.MOV": "img-0412", "2024 take.mov": "src-2024-take"} {
		if got := p.SuggestID(in); got != want {
			t.Errorf("SuggestID(%q)=%q want %q", in, got, want)
		}
	}
	p.Sources["talk"] = &Source{}
	if got := p.SuggestID("talk.wav"); got != "talk-2" {
		t.Errorf("collision: %q", got)
	}
}
