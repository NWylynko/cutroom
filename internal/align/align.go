// Package align measures how a separately recorded audio track lines up with a camera's audio:
// a constant offset plus clock drift (L05). Coarse: FFT cross-correlation of 10 ms loudness
// envelopes over the whole recordings. Fine: windowed cross-correlation of the waveforms around
// the coarse offset at many points, then a robust line fit (outliers dropped), because a plain
// least-squares fit was pulled off by windows in silence.
package align

import (
	"errors"
	"fmt"
	"math"
	"math/cmplx"
	"sort"

	"github.com/cutroom/cutroom/internal/speech"
)

// Sync maps reference time t (seconds, on the reference source's clock) to the other source:
// other = Offset + t × (1 + Drift).
type Sync struct {
	To       string     `json:"to"`          // reference source id (the camera)
	Offset   float64    `json:"offset"`      // seconds
	Drift    float64    `json:"drift"`       // dimensionless (ppm × 1e-6)
	PPM      float64    `json:"drift_ppm"`   // for display
	Residual float64    `json:"residual_ms"` // median absolute deviation of window offsets after the fit
	Windows  int        `json:"windows"`     // windows used in the fit
	Rejected int        `json:"rejected"`    // windows dropped as outliers or low confidence
	Overlap  [2]float64 `json:"overlap"`     // reference-time span both recordings cover
	Coarse   float64    `json:"coarse_offset"`
}

// Map converts a reference time to the other source's time.
func (s Sync) Map(t float64) float64 { return s.Offset + t*(1+s.Drift) }

const sr = speech.SampleRate

var ErrNoMatch = errors.New("not enough shared sound to align")

// Measure aligns other (x) to ref (y). Both are 16 kHz mono samples whose sample 0 is at
// refStart / otherStart seconds on their own source clocks.
func Measure(ref []float64, refStart float64, other []float64, otherStart float64) (Sync, error) {
	er, eo := envelope(ref), envelope(other)
	if len(er) < 200 || len(eo) < 200 {
		return Sync{}, fmt.Errorf("%w: recordings shorter than 2 s", ErrNoMatch)
	}
	lagFrames, corr := xcorrFFT(er, eo)
	if corr < 0.3 {
		return Sync{}, fmt.Errorf("%w: best envelope correlation %.2f (are these the same event?)", ErrNoMatch, corr)
	}
	// other_wav_time = ref_wav_time + coarse (wav clocks)
	coarse := float64(lagFrames) * 0.01

	// fine stage on the shared speech band only: two mics rarely agree outside 300–3000 Hz
	ref, other = bandpass(ref), bandpass(other)
	// fine windows: 2 s of ref speech, searched ±60 ms around the coarse offset
	win := 2 * sr
	maxLag := int(0.06 * sr)
	type pt struct{ t, off, c float64 }
	var pts []pt
	rejected := 0
	refDur := float64(len(ref)) / sr
	n := 40
	for k := 0; k < n; k++ {
		t := (float64(k) + 0.5) / float64(n) * refDur // ref wav time (window centre)
		ri := int(t*sr) - win/2
		oi := int((t+coarse)*sr) - win/2
		if ri < 0 || ri+win > len(ref) || oi-maxLag < 0 || oi+win+maxLag > len(other) {
			continue
		}
		seg := ref[ri : ri+win]
		if rms(seg) < 0.003 {
			continue // silence: nothing to correlate (L05: silent windows pulled a plain fit off)
		}
		lag, c := xcorr(seg, other[oi-maxLag:oi+win+maxLag], maxLag)
		if c < 0.5 {
			rejected++
			continue
		}
		pts = append(pts, pt{t, coarse + float64(lag)/sr, c})
	}
	if len(pts) < 4 {
		return Sync{}, fmt.Errorf("%w: only %d confident windows", ErrNoMatch, len(pts))
	}
	// robust line fit: least squares, drop > 3×MAD (min 2 ms), refit
	fit := func(ps []pt) (a, b float64) {
		var st, so, stt, sto float64
		for _, p := range ps {
			st += p.t
			so += p.off
			stt += p.t * p.t
			sto += p.t * p.off
		}
		nn := float64(len(ps))
		den := nn*stt - st*st
		if math.Abs(den) < 1e-9 {
			return so / nn, 0
		}
		b = (nn*sto - st*so) / den
		a = (so - b*st) / nn
		return
	}
	a, b := fit(pts)
	for iter := 0; iter < 3; iter++ {
		res := make([]float64, len(pts))
		for i, p := range pts {
			res[i] = math.Abs(p.off - (a + b*p.t))
		}
		mad := median(res)
		lim := math.Max(3*mad, 0.002)
		var keep []pt
		for i, p := range pts {
			if res[i] <= lim {
				keep = append(keep, p)
			} else {
				rejected++
			}
		}
		if len(keep) == len(pts) || len(keep) < 4 {
			break
		}
		pts = keep
		a, b = fit(pts)
	}
	res := make([]float64, len(pts))
	for i, p := range pts {
		res[i] = math.Abs(p.off - (a + b*p.t))
	}
	// convert wav clocks to source clocks:
	// other_src = otherStart + other_wav = otherStart + ref_wav + a + b·ref_wav,  ref_wav = t − refStart
	//           = (otherStart + a − (1+b)·refStart) + (1+b)·t
	s := Sync{Drift: b, PPM: round(b*1e6, 1), Residual: round(median(res)*1000, 2), Windows: len(pts), Rejected: rejected,
		Coarse: round(coarse+otherStart-refStart, 4)}
	s.Offset = round(otherStart+a-(1+b)*refStart, 6)
	// overlap in ref time
	lo := math.Max(refStart, (otherStart-s.Offset)/(1+b))
	hi := math.Min(refStart+refDur, (otherStart+float64(len(other))/sr-s.Offset)/(1+b))
	s.Overlap = [2]float64{round(lo, 3), round(hi, 3)}
	return s, nil
}

// envelope: 10 ms RMS, log-compressed, mean removed (robust to level differences between mics).
func envelope(x []float64) []float64 {
	n := sr / 100
	var e []float64
	for i := 0; i+n <= len(x); i += n {
		s := 0.0
		for _, v := range x[i : i+n] {
			s += v * v
		}
		e = append(e, math.Log10(s/float64(n)+1e-9))
	}
	m := median(append([]float64(nil), e...))
	for i := range e {
		e[i] = math.Max(e[i]-m, 0) // only energy above the median (speech), not the noise floor shape
	}
	return e
}

// xcorrFFT returns the lag L (in frames) maximising Σ ref[i]·other[i+L], i.e. other ≈ ref shifted by L.
func xcorrFFT(ref, other []float64) (int, float64) {
	n := 1
	for n < len(ref)+len(other) {
		n <<= 1
	}
	a := make([]complex128, n)
	b := make([]complex128, n)
	for i, v := range ref {
		a[i] = complex(v, 0)
	}
	for i, v := range other {
		b[i] = complex(v, 0)
	}
	fft(a, false)
	fft(b, false)
	for i := range a {
		a[i] = cmplx.Conj(a[i]) * b[i]
	}
	fft(a, true)
	er, eo := energy(ref), energy(other)
	best, bl := math.Inf(-1), 0
	for i := 0; i < n; i++ {
		lag := i
		if i > n/2 {
			lag = i - n
		}
		if lag <= -len(ref) || lag >= len(other) {
			continue
		}
		if v := real(a[i]); v > best {
			best, bl = v, lag
		}
	}
	return bl, best / math.Sqrt(er*eo+1e-12)
}

func energy(x []float64) float64 {
	s := 0.0
	for _, v := range x {
		s += v * v
	}
	return s
}

// fft is an in-place radix-2 Cooley–Tukey transform; len(a) must be a power of two.
func fft(a []complex128, inverse bool) {
	n := len(a)
	for i, j := 1, 0; i < n; i++ {
		bit := n >> 1
		for ; j&bit != 0; bit >>= 1 {
			j ^= bit
		}
		j ^= bit
		if i < j {
			a[i], a[j] = a[j], a[i]
		}
	}
	for size := 2; size <= n; size <<= 1 {
		ang := 2 * math.Pi / float64(size)
		if !inverse {
			ang = -ang
		}
		wl := cmplx.Exp(complex(0, ang))
		for i := 0; i < n; i += size {
			w := complex(1, 0)
			for k := 0; k < size/2; k++ {
				u, v := a[i+k], a[i+k+size/2]*w
				a[i+k], a[i+k+size/2] = u+v, u-v
				w *= wl
			}
		}
	}
	if inverse {
		for i := range a {
			a[i] /= complex(float64(n), 0)
		}
	}
}

// xcorr: lag in [-maxLag, maxLag] where seg best matches ref (ref = seg + 2·maxLag, centred);
// positive = the content appears later in ref.
func xcorr(seg, ref []float64, maxLag int) (int, float64) {
	es := energy(seg)
	score := func(lag int) float64 {
		off := maxLag + lag
		s, er := 0.0, 0.0
		for i, v := range seg {
			x := ref[off+i]
			s += v * x
			er += x * x
		}
		return s / math.Sqrt(es*er+1e-12)
	}
	best, bl := -2.0, 0
	for lag := -maxLag; lag <= maxLag; lag += 8 {
		if c := score(lag); c > best {
			best, bl = c, lag
		}
	}
	for lag := bl - 8; lag <= bl+8; lag++ {
		if lag >= -maxLag && lag <= maxLag {
			if c := score(lag); c > best {
				best, bl = c, lag
			}
		}
	}
	return bl, best
}

// Bandpass keeps the speech band (300–3000 Hz), so two microphones with different frequency
// responses correlate on what they share.
func Bandpass(x []float64) []float64 { return bandpass(x) }

// bandpass: 2nd-order high-pass at 300 Hz and low-pass at 3000 Hz (cascaded one-pole sections).
func bandpass(x []float64) []float64 {
	y := append([]float64(nil), x...)
	dt := 1.0 / sr
	for pass := 0; pass < 2; pass++ {
		rc := 1 / (2 * math.Pi * 300)
		a := rc / (rc + dt)
		prevX, prevY := y[0], 0.0
		for i := range y {
			v := a * (prevY + y[i] - prevX)
			prevX, prevY = y[i], v
			y[i] = v
		}
		rc = 1 / (2 * math.Pi * 3000)
		b := dt / (rc + dt)
		acc := 0.0
		for i := range y {
			acc += b * (y[i] - acc)
			y[i] = acc
		}
	}
	return y
}

func rms(x []float64) float64 { return math.Sqrt(energy(x) / float64(len(x))) }

func median(x []float64) float64 {
	if len(x) == 0 {
		return 0
	}
	s := append([]float64(nil), x...)
	sort.Float64s(s)
	return s[len(s)/2]
}

func round(f float64, d int) float64 {
	p := math.Pow(10, float64(d))
	return math.Round(f*p) / p
}
