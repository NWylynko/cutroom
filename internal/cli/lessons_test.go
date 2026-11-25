package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func listing(t *testing.T, dir string) string {
	t.Helper()
	var out []string
	filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		if err == nil {
			out = append(out, p+" "+fi.ModTime().String()+" "+fi.Mode().String())
		}
		return nil
	})
	sort.Strings(out)
	return strings.Join(out, "\n")
}

// L55: sources are read-only. A whole session (init, edit, render, look, export) leaves the source's
// folder exactly as it was, works when that folder is read-only, and refuses to render over a source.
func TestL55_SourcesAreNeverWritten(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
	media, proj := t.TempDir(), t.TempDir()
	src := filepath.Join(media, "cam.mp4")
	if out, err := exec.Command("ffmpeg", "-v", "error", "-y", "-f", "lavfi", "-i", "testsrc2=s=320x180:r=30:d=4",
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=48000:d=4", "-c:v", "libx264", "-preset", "ultrafast",
		"-pix_fmt", "yuv420p", "-c:a", "aac", "-shortest", src).CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, out)
	}
	before := listing(t, media)
	fi, _ := os.Stat(media)
	if err := os.Chmod(media, 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(media, fi.Mode().Perm())
	for _, args := range [][]string{
		{"init", src},
		{"edit", "cut", "1-1.5"},
		{"render", "--preview"},
		{"look", "0.5"},
		{"look", "--output", "1"},
		{"export", "--otio", "--edl"},
		{"check"},
	} {
		out, errOut, code := run(append(args, "--project", proj)...)
		if code != 0 && args[0] != "check" { // check may warn (no transcript); it must not write either
			t.Fatalf("%v: code %d\n%s%s", args, code, out, errOut)
		}
	}
	if _, errOut, code := run("render", src, "--project", proj); code == 0 || !strings.Contains(errOut, "would-overwrite-source") {
		t.Fatalf("rendering over the source was not refused: %d %s", code, errOut)
	}
	if _, errOut, code := run("export", "--otio", "--name", "../x", "--project", proj); code == 0 || !strings.Contains(errOut, "bad-name") {
		t.Fatalf("export --name with a path was not refused: %d %s", code, errOut)
	}
	os.Chmod(media, fi.Mode().Perm())
	if after := listing(t, media); after != before {
		t.Fatalf("the source folder changed:\nbefore\n%s\nafter\n%s", before, after)
	}
}
