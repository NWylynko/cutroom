package render

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/NWylynko/cutroom/internal/timeline"
)

// Guards for LESSONS invariants about the render pipeline. The ones that need ffmpeg make their own
// tiny sources and skip when ffmpeg isn't installed.

func needFF(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not installed")
	}
}

// tinySource writes a 320x180 30 fps clip of dur seconds with a 440 Hz tone.
func tinySource(t *testing.T, dur float64) Source {
	t.Helper()
	p := filepath.Join(t.TempDir(), "cam.mp4")
	d := strconv.FormatFloat(dur, 'f', 3, 64)
	out, err := exec.Command("ffmpeg", "-v", "error", "-y", "-f", "lavfi", "-i", "testsrc2=s=320x180:r=30:d="+d,
		"-f", "lavfi", "-i", "sine=frequency=440:sample_rate=48000:d="+d, "-c:v", "libx264", "-preset", "ultrafast",
		"-pix_fmt", "yuv420p", "-c:a", "aac", "-shortest", p).CombinedOutput()
	if err != nil {
		t.Fatalf("make source: %v %s", err, out)
	}
	st, _ := os.Stat(p)
	return Source{Path: p, HasAudio: true, Size: st.Size(), ModTime: st.ModTime()}
}

func ovl(id string, a, b, fade float64) Placement {
	return Placement{ID: id, Kind: "text", PNG: "/x/" + id + ".png", Key: id, Start: a, End: b, Fade: fade, Box: [4]float64{0.1, 0.8, 0.5, 0.9}}
}

// L10: a chunk's graph holds only the overlays visible in it.
func TestL10_ChunksHoldOnlyVisibleOverlays(t *testing.T) {
	tl := timeline.New("30", 1280, 720, "cam", "cam", 0, 200)
	for s := 10.0; s < 190; s += 10 {
		tl.Cut("cam", s, s+1, "x", true)
	}
	o := Options{Target: 45, CacheDir: t.TempDir(), Overlays: []Placement{ovl("o1", 5, 8, 0), ovl("o2", 150, 153, 0)}}
	cs := Plan(tl, srcs(), o)
	if len(cs) < 3 {
		t.Fatalf("want several chunks, got %d", len(cs))
	}
	seen := 0
	for _, c := range cs {
		end := c.Start + c.Duration
		for _, co := range c.Overlays {
			seen++
			if co.P.End <= c.Start || co.P.Start >= end {
				t.Errorf("chunk %d (%.1f–%.1f) carries %s (%.1f–%.1f)", c.Index, c.Start, end, co.P.ID, co.P.Start, co.P.End)
			}
		}
		args := strings.Join(chunkCmd(tl, srcs(), c, false, "x.mp4"), " ")
		if n := strings.Count(args, "/x/o"); n != len(c.Overlays) {
			t.Errorf("chunk %d reads %d overlay inputs, holds %d", c.Index, n, len(c.Overlays))
		}
	}
	if seen != 2 {
		t.Errorf("each overlay should be in exactly one chunk here; seen %d", seen)
	}
}

// L11: the chunk key is the chunk's content: an overlay's file (its key), encoder settings and the
// frame change it; nothing else does.
func TestL11_ChunkKeyIsContent(t *testing.T) {
	tl := timeline.New("30", 1280, 720, "cam", "cam", 0, 30)
	key := func(o Options) string { o.CacheDir = t.TempDir(); o.Target = 45; return Plan(tl, srcs(), o)[0].Key }
	base := key(Options{Overlays: []Placement{ovl("o1", 2, 4, 0)}})
	if key(Options{Overlays: []Placement{ovl("o1", 2, 4, 0)}}) != base {
		t.Fatal("same content, different key")
	}
	changed := ovl("o1", 2, 4, 0)
	changed.Key = "other-bytes"
	for name, o := range map[string]Options{
		"overlay file": {Overlays: []Placement{changed}},
		"overlay time": {Overlays: []Placement{ovl("o1", 2, 5, 0)}},
		"encoder":      {Preview: true, Overlays: []Placement{ovl("o1", 2, 4, 0)}},
		"no overlay":   {},
	} {
		if key(o) == base {
			t.Errorf("%s change kept the key", name)
		}
	}
}

// L12: chunks are encoded from the sources (never from an earlier render) and joined by stream copy.
func TestL12_EncodeOnceJoinByCopy(t *testing.T) {
	tl := timeline.New("30", 1280, 720, "cam", "cam", 0, 30)
	tl.Cut("cam", 6, 9, "x", true)
	c := Plan(tl, srcs(), Options{Target: 45, CacheDir: t.TempDir()})[0]
	args := chunkCmd(tl, srcs(), c, false, "out.mp4")
	for i, a := range args {
		if a == "-i" && args[i+1] != "/x/cam.mp4" {
			t.Errorf("chunk input %s is not a source", args[i+1])
		}
	}
	j := strings.Join(joinArgs("list.txt", "a.m4a", "out.mp4"), " ")
	if !strings.Contains(j, "-c copy") || strings.Contains(j, "libx264") || strings.Contains(j, "-filter") {
		t.Fatalf("join must be a stream copy: %s", j)
	}
}

// L13: no chunk boundary falls inside an overlay's fade.
func TestL13_NoChunkBoundaryInsideAFade(t *testing.T) {
	tl := timeline.New("30", 1280, 720, "cam", "cam", 0, 120)
	tl.Cut("cam", 46, 47, "x", true) // segment boundary at output 46 s: where a 45 s chunk would end
	o := Options{Target: 45, CacheDir: t.TempDir()}
	plain := Plan(tl, srcs(), o)
	b := plain[0].Start + plain[0].Duration
	o.Overlays = []Placement{ovl("o1", b-0.2, b+3, 0.5)} // its fade-in straddles that boundary
	for _, c := range Plan(tl, srcs(), o) {
		if c.Index == 0 {
			continue
		}
		f := int64(c.Start*30 + 0.5)
		if inFade(o.Overlays, f, 30) {
			t.Fatalf("chunk %d starts at frame %d, inside o1's fade", c.Index, f)
		}
	}
	if len(Plan(tl, srcs(), o)) >= len(plain) && Plan(tl, srcs(), o)[0].Start+Plan(tl, srcs(), o)[0].Duration == b {
		t.Fatal("the boundary inside the fade was kept")
	}
}

// L14: a chunk with fewer frames than planned fails loudly instead of shortening the video.
func TestL14_ShortChunkFails(t *testing.T) {
	needFF(t)
	src := tinySource(t, 1)
	tl := timeline.New("30", 320, 180, "cam", "cam", 0, 2) // the timeline believes the source is 2 s
	ss := map[string]Source{"cam": src}
	c := Plan(tl, ss, Options{Target: 45, CacheDir: t.TempDir(), Preview: true})[0]
	err := renderChunk(tl, ss, &c, true)
	if err == nil || !strings.Contains(err.Error(), "L14") {
		t.Fatalf("want an L14 frame-count error, got %v", err)
	}
	if _, err := os.Stat(c.file); err == nil {
		t.Fatal("a short chunk was cached")
	}
}

// L15: every overlay has its own input; nothing branches off the main picture.
func TestL15_EveryOverlayHasItsOwnInput(t *testing.T) {
	tl := timeline.New("30", 1280, 720, "cam", "cam", 0, 30)
	tl.Cut("cam", 6, 9, "x", true)
	o := Options{Target: 45, CacheDir: t.TempDir(), Overlays: []Placement{ovl("o1", 1, 3, 0.3), ovl("o2", 10, 14, 0)}}
	c := Plan(tl, srcs(), o)[0]
	args := chunkCmd(tl, srcs(), c, false, "out.mp4")
	j := strings.Join(args, " ")
	if n := strings.Count(j, " -i "); n != len(c.Segments)+len(c.Overlays) {
		t.Fatalf("%d inputs for %d segments + %d overlays", n, len(c.Segments), len(c.Overlays))
	}
	fc := args[slicesIndex(args, "-filter_complex")+1]
	if strings.Contains(fc, "split") || strings.Contains(fc, "asplit") {
		t.Fatalf("graph branches a stream: %s", fc)
	}
}

func slicesIndex(a []string, s string) int {
	for i, x := range a {
		if x == s {
			return i
		}
	}
	return -1
}

// L17: the audio is rendered once and cached on its own inputs: a picture-only change (framing here;
// overlays and zoom are the same) re-encodes the picture and reuses the audio.
func TestL17_PictureChangeReusesAudio(t *testing.T) {
	needFF(t)
	src := tinySource(t, 3)
	tl := timeline.New("30", 320, 180, "cam", "cam", 0, 3)
	ss := map[string]Source{"cam": src}
	o := Options{Target: 45, CacheDir: t.TempDir(), Loudness: -16, TruePeak: -1.5, Preview: true}
	out := filepath.Join(t.TempDir(), "a.mp4")
	r1, err := Render(tl, ss, out, o)
	if err != nil {
		t.Fatal(err)
	}
	tl.Fill, tl.FillX, tl.FillY = true, 0.5, 0.5
	r2, err := Render(tl, ss, out, o)
	if err != nil {
		t.Fatal(err)
	}
	if r2.Chunks[0].Cached || r2.Chunks[0].Key == r1.Chunks[0].Key {
		t.Fatal("the picture change should re-render the chunk")
	}
	if !r2.Audio.Cached {
		t.Fatal("a picture-only change re-ran the audio (L17)")
	}
}

// L27 + L43: the audio pass is measure → gain → limit → re-measure, at 48 kHz, and reports what it hit.
func TestL27_L43_LoudnessMeasuredAnd48k(t *testing.T) {
	needFF(t)
	d := t.TempDir()
	wav := filepath.Join(d, "quiet.wav")
	if out, err := exec.Command("ffmpeg", "-v", "error", "-y", "-f", "lavfi", "-i", "sine=frequency=300:sample_rate=44100:d=6",
		"-af", "volume=-30dB", "-ac", "1", wav).CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, out)
	}
	lufs, _, err := Loudness(wav, "")
	if err != nil {
		t.Fatal(err)
	}
	m4a := filepath.Join(d, "out.m4a")
	r, err := FinishAudio(wav, m4a, "", -16, -1.5, lufs)
	if err != nil {
		t.Fatal(err)
	}
	if r.LUFS < -17 || r.LUFS > -15 || r.TruePeak > -1.0 {
		t.Fatalf("achieved %.2f LUFS, %.2f dBTP (target -16, ceiling -1.5)", r.LUFS, r.TruePeak)
	}
	sr, _ := exec.Command("ffprobe", "-v", "error", "-select_streams", "a:0", "-show_entries", "stream=sample_rate", "-of", "csv=p=0", m4a).Output()
	if strings.TrimSpace(string(sr)) != "48000" {
		t.Fatalf("output sample rate %q, want 48000", sr)
	}
}

// L47: text arrives as images cutroom draws itself: no ffmpeg text filters anywhere in a chunk graph.
func TestL47_NoFFmpegTextFilters(t *testing.T) {
	tl := timeline.New("30", 1280, 720, "cam", "cam", 0, 30)
	o := Options{Target: 45, CacheDir: t.TempDir(), Overlays: []Placement{ovl("o1", 1, 3, 0.3)}}
	j := strings.Join(chunkCmd(tl, srcs(), Plan(tl, srcs(), o)[0], false, "out.mp4"), " ")
	for _, f := range []string{"drawtext", "subtitles", "ass=", "libass"} {
		if strings.Contains(j, f) {
			t.Fatalf("chunk graph uses %s", f)
		}
	}
	img, r, err := Card(timeline.Overlay{ID: "o1", Kind: "text", Text: "Install Linux", Layout: "lower-third"}, t.TempDir(), 1280, 720)
	if err != nil || r.Empty() {
		t.Fatalf("card: %v %v", err, r)
	}
	opaque := 0
	for i := 3; i < len(img.Pix); i += 4 {
		if img.Pix[i] > 200 {
			opaque++
		}
	}
	if opaque < 1000 {
		t.Fatalf("card has %d opaque pixels: text not drawn", opaque)
	}
}
