package media

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func fixture(t *testing.T, name string) string {
	t.Helper()
	p := filepath.Join("..", "..", "evals", "work", "fixtures", name)
	if _, err := os.Stat(p); err != nil {
		t.Skipf("fixture %s missing (run `make fixtures`)", name)
	}
	return p
}

func warned(f *Facts, code string) bool {
	for _, w := range f.Warnings {
		if w.Code == code {
			return true
		}
	}
	return false
}

func TestCleanSourceHasNoWarnings(t *testing.T) {
	f, err := Probe(fixture(t, "talk.mp4"), Options{Loudness: true})
	if err != nil {
		t.Fatal(err)
	}
	if f.Video.VFR || f.Video.Rotation != 0 || f.Video.Frames != 907 || len(f.Warnings) != 0 {
		t.Fatalf("clean fixture: vfr=%v rot=%d frames=%d warnings=%v", f.Video.VFR, f.Video.Rotation, f.Video.Frames, f.Warnings)
	}
	if !f.Audio.Measured || f.Audio.LUFS > -10 || f.Audio.LUFS < -40 {
		t.Fatalf("loudness not measured sensibly: %+v", f.Audio)
	}
}

func TestL02_DetectVFRAndAudioOffset(t *testing.T) {
	f, err := Probe(fixture(t, "talk-vfr.mp4"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	v := f.Video
	if !v.VFR || v.Gaps != 11 || v.Frames != 885 {
		t.Fatalf("vfr=%v gaps=%d frames=%d (want true, 11, 885)", v.VFR, v.Gaps, v.Frames)
	}
	// stream copy of AAC keeps the priming frame: first sample pts 0.063667 (content at 0.085)
	if math.Abs(f.AVOffset-0.063667) > 0.0005 || !warned(f, "vfr") || !warned(f, "av-offset") {
		t.Fatalf("av offset %.6f, warnings %v", f.AVOffset, f.Warnings)
	}
}

func TestL25_DetectRotation(t *testing.T) {
	f, err := Probe(fixture(t, "talk-rot.mov"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if f.Video.Rotation != 180 || !warned(f, "rotated") {
		t.Fatalf("rotation %d warnings %v", f.Video.Rotation, f.Warnings)
	}
}

func TestAudioOnlySource(t *testing.T) {
	f, err := Probe(fixture(t, "recorder.wav"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if f.Kind != "audio" || f.Video != nil || f.Audio.SampleRate != 48000 {
		t.Fatalf("%+v", f)
	}
}

// L04: a source whose first frame isn't at 0 (a joined take that lost its edit list) is reported:
// the picture's start and the offset to the sound, with no fix needed (cuts follow timestamps).
func TestL04_DetectLateFirstFrame(t *testing.T) {
	f, err := Probe(fixture(t, "talk-late.mp4"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(f.Video.Start-0.043) > 0.001 || math.Abs(f.AVOffset+0.043) > 0.001 || !warned(f, "av-offset") {
		t.Fatalf("video start %.4f, av offset %.4f, warnings %v", f.Video.Start, f.AVOffset, f.Warnings)
	}
}

// A container longer than its streams (a metadata track past the end): edits end where both picture
// and sound exist, and inspect says so (found on real phone footage in T3).
func TestFactsEndIsTheShortestStream(t *testing.T) {
	f := &Facts{Duration: 60.4, Video: &Video{Start: 0, Duration: 60.1667}, Audio: &Audio{Start: 0, Duration: 60.0093}}
	if math.Abs(f.End()-60.0093) > 1e-9 {
		t.Fatalf("end %.4f", f.End())
	}
	if !warned(&Facts{Duration: f.Duration, Video: f.Video, Audio: f.Audio, Warnings: warnings(f)}, "short-streams") {
		t.Fatal("no short-streams note")
	}
	audioOnly := &Facts{Duration: 90, Audio: &Audio{Duration: 89.5}}
	if audioOnly.End() != 90 {
		t.Fatalf("an audio-only source ends at its own duration, got %.3f", audioOnly.End())
	}
}
