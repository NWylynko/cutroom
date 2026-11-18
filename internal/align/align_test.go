package align

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/cutroom/cutroom/internal/speech"
)

// L05 on the recorder fixture: started 1.700 s before the camera, clock +50 ppm
// (content compressed by 1/1.00005, i.e. other = 1.7 + t × (1 − 49.9975e-6)).
func TestL05_RecorderFixture(t *testing.T) {
	fix := filepath.Join("..", "..", "evals", "work", "fixtures")
	if _, err := os.Stat(filepath.Join(fix, "recorder.wav")); err != nil {
		t.Skip("fixtures missing")
	}
	d := t.TempDir()
	cam, rec := filepath.Join(d, "cam.wav"), filepath.Join(d, "rec.wav")
	if err := speech.ExtractWav(filepath.Join(fix, "talk.mp4"), cam); err != nil {
		t.Fatal(err)
	}
	if err := speech.ExtractWav(filepath.Join(fix, "recorder.wav"), rec); err != nil {
		t.Fatal(err)
	}
	cx, _ := speech.ReadWav(cam)
	rx, _ := speech.ReadWav(rec)
	s, err := Measure(cx, 0, rx, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", s)
	if math.Abs(s.Offset-1.7) > 0.0005 {
		t.Errorf("offset %.5f, want 1.700 ± 0.5 ms", s.Offset)
	}
	if math.Abs(s.PPM-(-50)) > 25 {
		t.Errorf("drift %.1f ppm, want about −50 (30 s is short for drift)", s.PPM)
	}
	// mapped times must agree within 2 ms over the clip
	for _, tt := range []float64{1, 15, 29} {
		want := 1.7 + tt/1.00005
		if math.Abs(s.Map(tt)-want) > 0.001 {
			t.Errorf("Map(%.0f) = %.4f, want %.4f", tt, s.Map(tt), want)
		}
	}
}

func TestFFTRoundTrip(t *testing.T) {
	a := []complex128{1, 2, 3, 4, 0, -1, 0.5, 2}
	b := append([]complex128(nil), a...)
	fft(b, false)
	fft(b, true)
	for i := range a {
		if math.Abs(real(a[i]-b[i])) > 1e-9 {
			t.Fatalf("%v vs %v", a, b)
		}
	}
}
