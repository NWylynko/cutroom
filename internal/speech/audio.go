// Package speech turns a source's audio into a word-level transcript an agent can trust more than
// a raw speech-to-text pass: word onsets from whisper.cpp DTW, word ends and pause edges snapped
// to measured speech energy, and suspects flagged (LESSONS L28–L36).
package speech

import (
	"encoding/binary"
	"math"
	"os"
	"sort"

	"github.com/NWylynko/cutroom/internal/deps"
)

const (
	SampleRate = 16000
	FrameSec   = 0.01 // envelope resolution
)

// ExtractWav writes 16 kHz mono PCM of the first audio stream. Sample 0 is the stream's first
// sample, i.e. source time = wav time + audio start.
func ExtractWav(src, dst string) error {
	_, err := deps.Run("ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-i", src, "-vn", "-map", "0:a:0",
		"-ac", "1", "-ar", "16000", "-c:a", "pcm_s16le", dst)
	return err
}

// ReadWav reads a 16-bit mono WAV written by ExtractWav.
func ReadWav(p string) ([]float64, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	// find "data" chunk
	off := 12
	for off+8 <= len(b) {
		id := string(b[off : off+4])
		n := int(binary.LittleEndian.Uint32(b[off+4:]))
		if id == "data" {
			off += 8
			if off+n > len(b) {
				n = len(b) - off
			}
			x := make([]float64, n/2)
			for i := range x {
				x[i] = float64(int16(binary.LittleEndian.Uint16(b[off+2*i:]))) / 32768
			}
			return x, nil
		}
		off += 8 + n + n%2
	}
	return nil, os.ErrInvalid
}

// Envelope is RMS level in dBFS per 10 ms frame, after an 80 Hz high-pass (L46: rumble hides pauses).
type Envelope struct {
	DB    []float64
	Floor float64 // room tone (10th percentile)
	Peak  float64 // speech level (95th percentile)
	Thr   float64 // speech threshold
}

func NewEnvelope(x []float64) *Envelope {
	// first-order high-pass at 80 Hz
	rc := 1 / (2 * math.Pi * 80)
	dt := 1.0 / SampleRate
	a := rc / (rc + dt)
	y := make([]float64, len(x))
	for i := 1; i < len(x); i++ {
		y[i] = a * (y[i-1] + x[i] - x[i-1])
	}
	n := int(FrameSec * SampleRate)
	e := &Envelope{}
	for i := 0; i+n <= len(y); i += n {
		s := 0.0
		for _, v := range y[i : i+n] {
			s += v * v
		}
		e.DB = append(e.DB, 10*math.Log10(s/float64(n)+1e-12))
	}
	if len(e.DB) == 0 {
		return e
	}
	sorted := append([]float64(nil), e.DB...)
	sort.Float64s(sorted)
	e.Floor = sorted[len(sorted)/10]
	e.Peak = sorted[len(sorted)*95/100]
	e.Thr = e.Floor + math.Max(10, 0.35*(e.Peak-e.Floor))
	return e
}

// Region is a stretch of sound above the speech threshold, in wav seconds.
type Region struct{ Start, End, PeakDB float64 }

// Regions returns sound regions: frames above threshold, gaps < 80 ms closed, regions < 40 ms
// dropped, 30 ms hangover on each side (soft onsets and trailing sibilants).
func (e *Envelope) Regions() []Region {
	var rs []Region
	in := false
	var cur Region
	for i, d := range e.DB {
		t := float64(i) * FrameSec
		if d >= e.Thr {
			if !in {
				cur = Region{Start: t, PeakDB: d}
				in = true
			}
			cur.End = t + FrameSec
			cur.PeakDB = math.Max(cur.PeakDB, d)
		} else if in && t-cur.End >= 0.08 {
			rs = append(rs, cur)
			in = false
		}
	}
	if in {
		rs = append(rs, cur)
	}
	var out []Region
	for _, r := range rs {
		if r.End-r.Start < 0.04 {
			continue
		}
		r.Start = math.Max(0, r.Start-0.03)
		r.End += 0.03
		if len(out) > 0 && r.Start <= out[len(out)-1].End {
			out[len(out)-1].End = r.End
			out[len(out)-1].PeakDB = math.Max(out[len(out)-1].PeakDB, r.PeakDB)
			continue
		}
		out = append(out, r)
	}
	return out
}

// Level returns the max dB in [a, b) (wav seconds).
func (e *Envelope) Level(a, b float64) float64 {
	i, j := int(a/FrameSec), int(math.Ceil(b/FrameSec))
	m := math.Inf(-1)
	for k := max(0, i); k < min(len(e.DB), j); k++ {
		m = math.Max(m, e.DB[k])
	}
	return m
}
