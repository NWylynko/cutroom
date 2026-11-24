package align

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/cutroom/cutroom/internal/deps"
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

// L05 on a far, reverberant camera mic (T3: securing-personal-software, an iPhone across a room vs a
// lapel recorder correlated only 0.35–0.55 at the right lag). The camera's sound gets echoes, band
// limits and room noise: the old 0.5 window threshold must fail, the current one must land the offset.
func TestL05_ReverberantCamera(t *testing.T) {
	fix := filepath.Join("..", "..", "evals", "work", "fixtures")
	if _, err := os.Stat(filepath.Join(fix, "recorder.wav")); err != nil {
		t.Skip("fixtures missing")
	}
	d := t.TempDir()
	cam, rec := filepath.Join(d, "cam.wav"), filepath.Join(d, "rec.wav")
	if _, err := deps.Run("ffmpeg", "-v", "error", "-y", "-i", filepath.Join(fix, "talk.mp4"),
		"-f", "lavfi", "-i", "anoisesrc=color=brown:amplitude=0.02:duration=40:sample_rate=48000",
		"-filter_complex", "[0:a]aecho=0.8:0.9:31|73|149|241:0.55|0.45|0.4|0.35,highpass=f=200,lowpass=f=4000[r];[r][1:a]amix=inputs=2:duration=first:normalize=0",
		"-ac", "1", "-ar", "16000", "-c:a", "pcm_s16le", cam); err != nil {
		t.Fatal(err)
	}
	if err := speech.ExtractWav(filepath.Join(fix, "recorder.wav"), rec); err != nil {
		t.Fatal(err)
	}
	cx, _ := speech.ReadWav(cam)
	rx, _ := speech.ReadWav(rec)
	defer func(v float64) { minWindowCorr = v }(minWindowCorr)
	minWindowCorr = 0.5 // the old threshold: too strict for this camera
	if s, err := Measure(cx, 0, rx, 0); err == nil {
		t.Errorf("expected the 0.5 window threshold to fail on a reverberant camera (it measured %+v): the test lost its teeth", s)
	}
	minWindowCorr = 0.35
	s, err := Measure(cx, 0, rx, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", s)
	// echoes pull the correlation peak by a millisecond or two (real footage: 2 ms); lip sync needs
	// tens of ms, so 2.5 ms is the honest bound here, not the 0.5 ms a clean pair achieves
	if math.Abs(s.Offset-1.7) > 0.0025 {
		t.Errorf("offset %.5f, want 1.700 ± 2.5 ms", s.Offset)
	}
}
