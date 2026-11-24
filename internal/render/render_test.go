package render

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cutroom/cutroom/internal/timeline"
)

func srcs() map[string]Source {
	return map[string]Source{"cam": {Path: "/x/cam.mp4", HasAudio: true, AudioStart: 0.0637, Size: 1, ModTime: time.Unix(1, 0)}}
}

// L07: never -itsoffset on a seeked read. L08: restamp after concat. L09: early seek + spare frames.
func TestL07_L08_ChunkCommandInvariants(t *testing.T) {
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
	total, err := Assemble(src, out, []Piece{{0, 4800}, {24000, 28800}}, 20)
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
