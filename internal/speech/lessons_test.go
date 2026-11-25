package speech

import (
	"math"
	"math/rand"
	"testing"
)

// L46: pause detection high-passes first. Rumble (brown noise, louder than the voice) runs under speech –
// pause – speech; the pause must still be found.
func TestL46_PauseFoundUnderRumble(t *testing.T) {
	n := 3 * SampleRate
	x := make([]float64, n)
	// rumble: brown noise (handling, wind, traffic): its energy is almost all below 80 Hz
	r := rand.New(rand.NewSource(1))
	b, sum := 0.0, 0.0
	for i := range x {
		b = 0.999*b + r.NormFloat64()*0.01
		x[i] = b
		sum += b * b
	}
	g := 0.3 / math.Sqrt(sum/float64(n)) // rumble RMS 0.3, the voice's 0.21: louder than the voice
	for i := range x {
		s := float64(i) / SampleRate
		x[i] *= g
		if s < 1 || s >= 2 {
			x[i] += 0.3 * math.Sin(2*math.Pi*300*s)
		}
	}
	rs := NewEnvelope(x).Regions()
	if len(rs) != 2 {
		t.Fatalf("want speech | pause | speech as 2 regions, got %+v", rs)
	}
	if math.Abs(rs[0].End-1) > 0.1 || math.Abs(rs[1].Start-2) > 0.1 {
		t.Fatalf("pause found at %.2f–%.2f, want 1–2", rs[0].End, rs[1].Start)
	}
}

// A recorder's words re-timed onto the camera's clock keep their order and drop what the camera
// didn't record.
func TestMapToCameraClock(t *testing.T) {
	rec := &Transcript{Source: "mic", Words: []Word{{Text: "before", Start: 1, End: 1.5}, {Text: "hello", Start: 16, End: 16.5}, {Text: "world.", Start: 17, End: 17.4}}}
	m := rec.MapTo("cam", func(r float64) float64 { return r - 15.17 }, 60)
	if m.Source != "cam" || m.From != "mic" || len(m.Words) != 2 || m.Words[0].Text != "hello" || m.Words[0].I != 0 || math.Abs(m.Words[0].Start-0.83) > 1e-9 {
		t.Fatalf("%+v", m)
	}
}
