package verify

import (
	"math/rand"
	"testing"
)

// Sign convention: render audio late by d samples → +d; early → −d. (It was inverted once:
// a planted 100 ms delay read as −100.)
func TestXcorrSignLateIsPositive(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	src := make([]float64, 20000)
	for i := range src {
		src[i] = rng.NormFloat64()
	}
	maxLag, win, centre := 2400, 8000, 10000
	ref := src[centre-win/2-maxLag : centre+win/2+maxLag]
	for _, d := range []int{800, -800, 0} {
		// the render plays, at the moment mapped to source `centre`, source content from centre-d
		seg := src[centre-d-win/2 : centre-d+win/2]
		if lag, c := xcorr(seg, ref, maxLag); lag != d || c < 0.999 {
			t.Errorf("render late by %d samples: got lag %d (corr %.3f)", d, lag, c)
		}
	}
}
