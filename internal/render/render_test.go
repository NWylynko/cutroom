package render

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NWylynko/cutroom/internal/align"
	"github.com/NWylynko/cutroom/internal/timeline"
)

func srcs() map[string]Source {
	return map[string]Source{"cam": {Path: "/x/cam.mp4", HasAudio: true, AudioStart: 0.0637, Size: 1, ModTime: time.Unix(1, 0)}}
}

// L07: never -itsoffset on a seeked read. L08: restamp after concat. L09: early seek + spare frames.
func TestL07_L08_L09_ChunkCommandInvariants(t *testing.T) {
	tl := timeline.New("30", 1280, 720, "cam", "cam", 0.0637, 30)
	tl.Cut("cam", 6, 9, "x", true)
	c := Plan(tl, srcs(), Options{Target: 45, CacheDir: t.TempDir()})[0]
	args := strings.Join(chunkCmd(tl, srcs(), c, false, "out.mp4"), " ")
	if strings.Contains(args, "itsoffset") {
		t.Fatal("L07: chunk command uses -itsoffset")
	}
	if !strings.Contains(args, "setpts=N/(FRAME_RATE*TB)") {
		t.Fatal("L08: no restamp after concat")
	}
	if strings.Count(args, " -ss ") != 2 || !strings.Contains(args, "trim=end_frame=") {
		t.Fatalf("each segment needs its own seek and an exact frame trim: %s", args)
	}
	if !strings.Contains(args, "start_time=0.250000") {
		t.Fatalf("L09: expected a 0.25 s early seek with the grid starting after it: %s", args)
	}
}

func TestPlanChunksAtCutsAndKeys(t *testing.T) {
	tl := timeline.New("30", 1280, 720, "cam", "cam", 0, 200)
	for s := 10.0; s < 190; s += 10 {
		tl.Cut("cam", s, s+1, "x", true)
	}
	o := Options{Target: 45, CacheDir: t.TempDir()}
	cs := Plan(tl, srcs(), o)
	var frames int64
	for _, c := range cs {
		frames += c.Frames
		if c.Duration < 30 && c.Index != len(cs)-1 {
			t.Errorf("chunk %d only %.1f s", c.Index, c.Duration)
		}
	}
	var want int64
	for _, g := range tl.Segments {
		want += tl.Frames(g)
	}
	if frames != want {
		t.Fatalf("chunks hold %d frames, timeline %d", frames, want)
	}
	// a change in one segment changes only the key of its chunk
	tl2 := *tl
	tl2.Segments = append([]timeline.Segment(nil), tl.Segments...)
	tl2.Segments[len(tl2.Segments)-1].Out -= 1
	tl2.Recompute()
	cs2 := Plan(&tl2, srcs(), o)
	same := 0
	for i := range cs {
		if i < len(cs2) && cs[i].Key == cs2[i].Key {
			same++
		}
	}
	if same != len(cs)-1 {
		t.Fatalf("expected all but the last chunk cached; %d of %d keys equal", same, len(cs))
	}
}

// L01/L22: audio pieces are copied sample-exactly with short fades at each edge.
func TestL01_AssembleSampleExactWithFades(t *testing.T) {
	d := t.TempDir()
	src := filepath.Join(d, "src.wav")
	f, _ := os.Create(src)
	n := int64(SR) // 1 s of constant 0.5
	writeWavHeader(f, 1, n)
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, math.Float32bits(0.5))
	for i := int64(0); i < n; i++ {
		f.Write(b)
	}
	f.Close()
	out := filepath.Join(d, "out.wav")
	total, err := Assemble([]string{src}, out, []Piece{{0, 0, 4800}, {0, 24000, 28800}}, 20)
	if err != nil || total != 9600 {
		t.Fatalf("total %d err %v", total, err)
	}
	raw, _ := os.ReadFile(out)
	sample := func(i int) float32 { return math.Float32frombits(binary.LittleEndian.Uint32(raw[44+4*i:])) }
	if len(raw) != 44+9600*4 {
		t.Fatalf("size %d", len(raw))
	}
	if sample(0) != 0 || math.Abs(float64(sample(2400))-0.5) > 1e-6 || sample(4799) > 0.01 || sample(4800) != 0 {
		t.Fatalf("fades wrong: %v %v %v %v", sample(0), sample(2400), sample(4799), sample(4800))
	}
}

// L24: frames downloaded from the hardware tone mapper keep their bt2020 tags; every chunk is retagged
// bt709 in the graph and on the encoder, whatever the source.
func TestL24_OutputAlwaysTaggedBT709(t *testing.T) {
	tl := timeline.New("30", 1280, 720, "cam", "cam", 0, 30)
	tl.Cut("cam", 6, 9, "x", true)
	c := Plan(tl, srcs(), Options{Target: 45, CacheDir: t.TempDir()})[0]
	args := strings.Join(chunkCmd(tl, srcs(), c, false, "out.mp4"), " ")
	for _, want := range []string{"setparams=colorspace=bt709:color_primaries=bt709:color_trc=bt709:range=tv[out]",
		"-colorspace bt709", "-color_primaries bt709", "-color_trc bt709", "-color_range tv"} {
		if !strings.Contains(args, want) {
			t.Fatalf("missing %q in %s", want, args)
		}
	}
}

// L40: a zoomed segment is cropped around its anchor (kept inside the frame) before scaling;
// unzoomed segments are not cropped.
func TestL40_ZoomCropInChunk(t *testing.T) {
	tl := timeline.New("30", 1280, 720, "cam", "cam", 0, 30)
	tl.Zoom("cam", 10, 12, 1.2, 0.5, 0.3)
	c := Plan(tl, srcs(), Options{Target: 45, CacheDir: t.TempDir()})[0]
	args := strings.Join(chunkCmd(tl, srcs(), c, false, "out.mp4"), " ")
	if strings.Count(args, "crop=") != 1 || !strings.Contains(args, "trunc(iw/1.2/2)*2") || !strings.Contains(args, "clip(0.3*ih-oh/2,0,ih-oh)") {
		t.Fatalf("want one anchored crop: %s", args)
	}
}

// L05/L24 multi-take: each video source's sound is resolved explicitly; a recorder is used only where
// it is synced; other takes fall back to their own sound; a hand-set recorder synced to nothing is refused.
func TestAudioForMultiTake(t *testing.T) {
	sync := func(to string) *align.Sync { return &align.Sync{To: to, Offset: 1} }
	srcs := map[string]Source{
		"cam1": {HasAudio: true}, "cam2": {HasAudio: true}, "mute": {},
		"mic1": {HasAudio: true, Sync: sync("cam1")}, "mic2": {HasAudio: true, Sync: sync("cam2")}, "loose": {HasAudio: true},
	}
	tl := timeline.New("30", 1280, 720, "cam1", "mic1", 0, 30)
	tl.Insert("cam2", 1, 2, 1, "take 2")
	for _, c := range []struct {
		video, want string
		own, err    bool
	}{{"cam1", "mic1", false, false}, {"cam2", "cam2", true, false}} {
		a, own, err := AudioFor(tl, srcs, c.video)
		if a != c.want || own != c.own || (err != nil) != c.err {
			t.Errorf("%s: got %s own=%v err=%v", c.video, a, own, err)
		}
	}
	tl.AudioFor = map[string]string{"cam2": "mic2"}
	if a, own, err := AudioFor(tl, srcs, "cam2"); a != "mic2" || own || err != nil {
		t.Errorf("mapped recorder: %s %v %v", a, own, err)
	}
	tl.AudioFor = map[string]string{"cam2": "mic1"} // mic1 is synced to cam1, not cam2
	if _, _, err := AudioFor(tl, srcs, "cam2"); err == nil {
		t.Error("a recorder synced to another take must be refused")
	}
	tl.AudioFor, tl.Audio = nil, "loose"
	if _, _, err := AudioFor(tl, srcs, "cam1"); err == nil {
		t.Error("a hand-set recorder synced to nothing must be refused (L05)")
	}
}

// L16: an overlay is a looped still shown by frame number and repeated, never ended with
// eof_action=pass (which drops its last frame); across chunks each part fades only at its own ends,
// and chunks without overlays keep their cache keys.
func TestL16_OverlayAcrossChunksIsFrameExact(t *testing.T) {
	tl := timeline.New("30", 1280, 720, "cam", "cam", 0, 200)
	for s := 10.0; s < 190; s += 10 {
		tl.Cut("cam", s, s+1, "x", true)
	}
	o := Options{Target: 45, CacheDir: t.TempDir()}
	plain := Plan(tl, srcs(), o)
	b := plain[0].Start + plain[0].Duration // first chunk boundary
	o.Overlays = []Placement{{ID: "o1", PNG: "/x/o1.png", Key: "k1", Start: b - 2, End: b + 3, Fade: 0.2}}
	cs := Plan(tl, srcs(), o)
	if len(cs[0].Overlays) != 1 || len(cs[1].Overlays) != 1 {
		t.Fatalf("overlay should be in chunks 1 and 2: %+v / %+v", cs[0].Overlays, cs[1].Overlays)
	}
	a, c := cs[0].Overlays[0], cs[1].Overlays[0]
	if a.To != cs[0].Frames || a.From != cs[0].Frames-60 || !a.FadeIn || a.FadeOut {
		t.Errorf("first part: %+v (chunk %d frames)", a, cs[0].Frames)
	}
	if c.From != 0 || c.To != 90 || c.FadeIn || !c.FadeOut {
		t.Errorf("second part: %+v", c)
	}
	if cs[2].Key != plain[2].Key || cs[0].Key == plain[0].Key {
		t.Errorf("keys: chunks with the overlay must change, others not")
	}
	args := strings.Join(chunkCmd(tl, srcs(), cs[0], false, "out.mp4"), " ")
	for _, want := range []string{"-loop 1", "eof_action=repeat", fmt.Sprintf("between(n,%d,%d)", a.From, a.To-1), "fade=t=in", fmt.Sprintf("trim=end_frame=%d", cs[0].Frames)} {
		if !strings.Contains(args, want) {
			t.Errorf("chunk command lacks %q: %s", want, args)
		}
	}
	if strings.Contains(args, "eof_action=pass") || strings.Contains(args, "fade=t=out") {
		t.Errorf("L16 / wrong fade in first part: %s", args)
	}
}

// L48: burned-in captions are one alpha clip per chunk (an ffconcat of cards), not an overlay per
// group; its runs cover every frame of the chunk exactly once.
func TestL48_CaptionsOneClipPerChunk(t *testing.T) {
	tl := timeline.New("30", 1280, 720, "cam", "cam", 0, 200)
	for s := 10.0; s < 190; s += 10 {
		tl.Cut("cam", s, s+1, "x", true)
	}
	o := Options{Target: 45, CacheDir: t.TempDir()}
	plain := Plan(tl, srcs(), o)
	var gs []CapGroup
	for x := 0.5; x < tl.Duration()-1; x += 0.7 {
		gs = append(gs, CapGroup{Start: x, End: x + 0.6, Text: "w", PNG: fmt.Sprintf("/x/g%.1f.png", x)})
	}
	o.Overlays = []Placement{{ID: "o1", Kind: "captions", Start: 0, End: tl.Duration(), Groups: gs}}
	cs := Plan(tl, srcs(), o)
	for i, c := range cs {
		if len(c.Overlays) != 1 {
			t.Fatalf("chunk %d: %d caption overlays", i, len(c.Overlays))
		}
		at := int64(0)
		for _, r := range c.Overlays[0].Frames {
			if r.From != at || r.To <= r.From {
				t.Fatalf("chunk %d: runs not contiguous at %d: %+v", i, at, r)
			}
			at = r.To
		}
		if at != c.Frames {
			t.Fatalf("chunk %d: runs cover %d of %d frames", i, at, c.Frames)
		}
		if c.Key == plain[i].Key {
			t.Errorf("chunk %d key unchanged by captions", i)
		}
		args := strings.Join(chunkCmd(tl, srcs(), c, false, "out.mp4"), " ")
		if strings.Count(args, "-f concat") != 1 || strings.Contains(args, "-loop 1") || !strings.Contains(args, "eof_action=repeat") {
			t.Errorf("chunk %d: captions must be one concat clip: %s", i, args)
		}
	}
	list := capList(cs[0].Overlays[0].Frames, "/x/blank.png", 30)
	if !strings.HasPrefix(list, "ffconcat version 1.0\n") || strings.Count(list, "duration ") != len(cs[0].Overlays[0].Frames) {
		t.Errorf("concat list: %s", list)
	}
}

// A clip overlay is relabelled bt709 like the sources and kept in YUV: newer ffmpeg otherwise
// converts an untagged clip from bt601 to the base's bt709 and shifts its colours (round 8).
func TestClipKeepsColours(t *testing.T) {
	tl := timeline.New("30", 1280, 720, "cam", "cam", 0, 30)
	o := Options{Target: 45, CacheDir: t.TempDir(), Overlays: []Placement{{ID: "o1", Kind: "clip", Clip: "/x/b.mp4", Start: 2, End: 4, Full: true, Box: [4]float64{0, 0, 1, 1}, Key: "k"}}}
	c := Plan(tl, srcs(), o)[0]
	args := strings.Join(chunkCmd(tl, srcs(), c, false, "out.mp4"), " ")
	i := strings.Index(args, "-i /x/b.mp4")
	chain := args[strings.Index(args, "[1:v]"):]
	if i < 0 || !strings.Contains(chain, "setparams=colorspace=bt709") || !strings.Contains(chain, "format=yuva420p") || strings.Contains(chain[:strings.Index(chain, "[o0]")], "rgba") {
		t.Fatalf("clip chain must relabel bt709 and stay YUV: %s", chain)
	}
}
