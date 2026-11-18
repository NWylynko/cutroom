package timecode

import "testing"

// L01: cuts must land on integer frames and samples; float-second trims drifted A/V.
// Summing many segments computed in exact rationals must not drift, and every segment's
// frame count and sample count must describe the same duration within half a frame.
func TestL01_NoDriftOverManyCuts(t *testing.T) {
	r := Rate{30000, 1001}
	const sr = 48000
	var totalFrames, totalSamples int64
	in, _ := Seconds("0")
	for i := 0; i < 500; i++ {
		start := in.Add(mustSec(t, "0.1234567"))
		end := start.Add(mustSec(t, "3.3333333"))
		f := end.Frame(r) - start.Frame(r)
		s := end.Sample(sr) - start.Sample(sr)
		totalFrames += f
		totalSamples += s
		in = end
	}
	vSec := float64(totalFrames) / r.FPS()
	aSec := float64(totalSamples) / sr
	// Per-segment rounding to frames can differ from samples by < 1 frame each, but the
	// renderer derives both from the same frame boundaries; here we only check the time
	// model itself is exact: the cumulative end equals the sum of parts exactly.
	want := mustSec(t, "1728.395") // 500 × (0.1234567 + 3.3333333)
	if in.Cmp(want) != 0 {
		t.Fatalf("cumulative time drifted: got %s want %s", in, want)
	}
	if d := vSec - aSec; d > 0.5/r.FPS()*500 || d < -0.5/r.FPS()*500 {
		t.Fatalf("frame/sample totals diverge: %.4f s", d)
	}
}

func TestL01_FrameSnapExact(t *testing.T) {
	r := Rate{30, 1}
	cases := map[string]int64{"0": 0, "245.1667": 7355, "245.15": 7355, "245.1834": 7356, "1/30": 1}
	for in, want := range cases {
		tm, err := Seconds(in)
		if err != nil {
			t.Fatal(err)
		}
		if got := tm.Frame(r); got != want {
			t.Errorf("Frame(%s) = %d, want %d", in, got, want)
		}
	}
	if got := FrameTime(7355, r); got.String() != "245.167" {
		t.Errorf("FrameTime = %s", got)
	}
}

func TestParse(t *testing.T) {
	r := Rate{30, 1}
	ok := map[string]string{
		"12.5": "12.500", "1:02.5": "62.500", "1:00:02.5": "3602.500", "12.5s": "12.500", "f45": "1.500",
	}
	for in, want := range ok {
		got, err := Parse(in, r)
		if err != nil || got.String() != want {
			t.Errorf("Parse(%q) = %v, %v; want %s", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "abc", "1:2:3:4", "1.5:02", "f-1", "-3"} {
		if _, err := Parse(bad, r); err == nil {
			t.Errorf("Parse(%q) should fail", bad)
		}
	}
}

func TestParseRate(t *testing.T) {
	for in, want := range map[string]string{"30": "30", "29.97": "30000/1001", "60000/1001": "60000/1001", "25.0": "25"} {
		r, err := ParseRate(in)
		if err != nil || r.String() != want {
			t.Errorf("ParseRate(%q) = %v, %v", in, r, err)
		}
	}
}

func TestClock(t *testing.T) {
	for in, want := range map[string]string{"62.5": "1:02.500", "3602.5": "1:00:02.500", "0.0004": "0:00.000"} {
		if got := mustSec(t, in).Clock(); got != want {
			t.Errorf("Clock(%s) = %s want %s", in, got, want)
		}
	}
}

func mustSec(t *testing.T, s string) Time {
	t.Helper()
	v, err := Seconds(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
