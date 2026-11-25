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
