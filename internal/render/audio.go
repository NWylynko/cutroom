package render

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/NWylynko/cutroom/internal/deps"
)

// VoiceChain is the default audio chain from published videos (L44): rumble cut, gentle compression,
// a static de-harsh shelf. Loudness is applied after it by measurement (L43).
const VoiceChain = "highpass=f=80,acompressor=threshold=-24dB:ratio=2.5:attack=10:release=200,highshelf=f=5000:g=-3,equalizer=f=3500:t=q:w=1.2:g=-1.5"

const SR = 48000

// AudioWav decodes the first audio stream of src to 48 kHz float32 WAV (cached by key). Sample 0 is
// the stream's first sample (source time = audio start).
func AudioWav(src, dst string) (channels int, err error) { return AudioWavCh(src, dst, 0) }

// AudioWavCh is AudioWav with a fixed channel count (0 = the source's own); several sources joined
// in one timeline are read as mono so their pieces line up.
func AudioWavCh(src, dst string, chans int) (channels int, err error) {
	if ch, ok := wavChannels(dst); ok {
		return ch, nil
	}
	tmp := dst + ".tmp.wav"
	args := []string{"-hide_banner", "-loglevel", "error", "-y", "-i", src, "-vn", "-map", "0:a:0", "-ar", strconv.Itoa(SR)}
	if chans > 0 {
		args = append(args, "-ac", strconv.Itoa(chans))
	}
	if _, err := deps.Run("ffmpeg", append(args, "-c:a", "pcm_f32le", tmp)...); err != nil {
		return 0, err
	}
	if err := os.Rename(tmp, dst); err != nil {
		return 0, err
	}
	ch, ok := wavChannels(dst)
	if !ok {
		return 0, fmt.Errorf("bad wav %s", dst)
	}
	return ch, nil
}

type wavInfo struct {
	channels   int
	dataOff    int64
	dataLen    int64
	bytesPerSa int
}

func readWavInfo(p string) (wavInfo, error) {
	f, err := os.Open(p)
	if err != nil {
		return wavInfo{}, err
	}
	defer f.Close()
	hdr := make([]byte, 12)
	if _, err := io.ReadFull(f, hdr); err != nil || string(hdr[:4]) != "RIFF" || string(hdr[8:12]) != "WAVE" {
		return wavInfo{}, fmt.Errorf("not a wav: %s", p)
	}
	var wi wavInfo
	off := int64(12)
	for {
		ch := make([]byte, 8)
		if _, err := f.ReadAt(ch, off); err != nil {
			return wavInfo{}, fmt.Errorf("no data chunk in %s", p)
		}
		id, n := string(ch[:4]), int64(binary.LittleEndian.Uint32(ch[4:]))
		if id == "fmt " {
			b := make([]byte, 16)
			f.ReadAt(b, off+8)
			wi.channels = int(binary.LittleEndian.Uint16(b[2:]))
			wi.bytesPerSa = int(binary.LittleEndian.Uint16(b[14:])) / 8
		}
		if id == "data" {
			wi.dataOff = off + 8
			wi.dataLen = n
			if st, _ := f.Stat(); n == 0 || n == 0xFFFFFFFF || wi.dataOff+n > st.Size() {
				wi.dataLen = st.Size() - wi.dataOff
			}
			return wi, nil
		}
		off += 8 + n + n%2
	}
}

func wavChannels(p string) (int, bool) {
	wi, err := readWavInfo(p)
	if err != nil || wi.bytesPerSa != 4 {
		return 0, false
	}
	return wi.channels, true
}

// Piece is a range of source samples [From, To) to copy into the output.
// Piece is a sample range [From, To) of source file Src (index into Assemble's srcs).
type Piece struct {
	Src      int
	From, To int64
}

// Assemble writes the pieces of src (a float32 WAV) back to back into dst, with fadeMS raised-cosine
// fades at every piece edge (L22) so cuts don't click. Samples before 0 or past the end are silence.
func Assemble(srcs []string, dst string, pieces []Piece, fadeMS float64) (int64, error) {
	wis := make([]wavInfo, len(srcs))
	ins := make([]*os.File, len(srcs))
	for i, src := range srcs {
		wi, err := readWavInfo(src)
		if err != nil {
			return 0, err
		}
		if i > 0 && wi.channels != wis[0].channels {
			return 0, fmt.Errorf("audio sources differ in channel count (%d vs %d); read them with the same count", wis[0].channels, wi.channels)
		}
		in, err := os.Open(src)
		if err != nil {
			return 0, err
		}
		defer in.Close()
		wis[i], ins[i] = wi, in
	}
	wi := wis[0]
	out, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer out.Close()
	var total int64
	for _, p := range pieces {
		total += p.To - p.From
	}
	ch := wi.channels
	frameBytes := int64(4 * ch)
	writeWavHeader(out, ch, total)
	fade := int64(fadeMS / 1000 * SR)
	buf := make([]byte, 0, 1<<20)
	for _, p := range pieces {
		in, pw := ins[p.Src], wis[p.Src]
		nSrc := pw.dataLen / frameBytes
		n := p.To - p.From
		data := make([]byte, n*frameBytes)
		lo, hi := max(p.From, 0), min(p.To, nSrc)
		if hi > lo {
			if _, err := in.ReadAt(data[(lo-p.From)*frameBytes:(hi-p.From)*frameBytes], pw.dataOff+lo*frameBytes); err != nil && err != io.EOF {
				return 0, err
			}
		}
		f := min(fade, n/2)
		for i := int64(0); i < f; i++ {
			g := float32(0.5 - 0.5*math.Cos(math.Pi*float64(i)/float64(f)))
			for c := 0; c < ch; c++ {
				scale(data, (i*int64(ch)+int64(c))*4, g)
				scale(data, ((n-1-i)*int64(ch)+int64(c))*4, g)
			}
		}
		buf = append(buf[:0], data...)
		if _, err := out.Write(buf); err != nil {
			return 0, err
		}
	}
	return total, nil
}

func scale(b []byte, off int64, g float32) {
	v := math.Float32frombits(binary.LittleEndian.Uint32(b[off:]))
	binary.LittleEndian.PutUint32(b[off:], math.Float32bits(v*g))
}

func writeWavHeader(w io.Writer, ch int, frames int64) {
	dataLen := uint32(frames * int64(4*ch))
	h := make([]byte, 44)
	copy(h, "RIFF")
	binary.LittleEndian.PutUint32(h[4:], 36+dataLen)
	copy(h[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(h[16:], 16)
	binary.LittleEndian.PutUint16(h[20:], 3) // IEEE float
	binary.LittleEndian.PutUint16(h[22:], uint16(ch))
	binary.LittleEndian.PutUint32(h[24:], SR)
	binary.LittleEndian.PutUint32(h[28:], uint32(SR*4*ch))
	binary.LittleEndian.PutUint16(h[32:], uint16(4*ch))
	binary.LittleEndian.PutUint16(h[34:], 32)
	copy(h[36:], "data")
	binary.LittleEndian.PutUint32(h[40:], dataLen)
	w.Write(h)
}

var (
	ebuI    = regexp.MustCompile(`I:\s+(-?[\d.]+) LUFS`)
	ebuPeak = regexp.MustCompile(`Peak:\s+(-?[\d.]+|-inf) dBFS`)
)

// Loudness measures integrated loudness and true peak of in after the filter chain.
func Loudness(in, chain string) (lufs, peak float64, err error) {
	af := "ebur128=peak=true"
	if chain != "" {
		af = chain + "," + af
	}
	text, err := deps.RunErr("ffmpeg", "-hide_banner", "-nostats", "-i", in, "-af", af, "-f", "null", "-")
	if err != nil {
		return 0, 0, err
	}
	m := ebuI.FindAllStringSubmatch(text, -1)
	p := ebuPeak.FindAllStringSubmatch(text, -1)
	if m == nil || p == nil {
		return 0, 0, fmt.Errorf("could not measure loudness")
	}
	lufs, _ = strconv.ParseFloat(m[len(m)-1][1], 64)
	if p[len(p)-1][1] == "-inf" {
		peak = math.Inf(-1)
	} else {
		peak, _ = strconv.ParseFloat(p[len(p)-1][1], 64)
	}
	return lufs, peak, nil
}

// AudioResult reports what the audio pass did.
type AudioResult struct {
	File     string  `json:"file"`
	LUFS     float64 `json:"lufs"`
	TruePeak float64 `json:"true_peak_db"`
	GainDB   float64 `json:"gain_db"`
	Target   float64 `json:"target_lufs"`
	Samples  int64   `json:"samples"`
	Cached   bool    `json:"cached"`
}

// FinishAudio runs chain → measured gain → limiter (L43) and encodes AAC stereo 48 kHz (L27).
// lufs is the loudness of the whole timeline after the chain (measured by the caller), so a range
// render gets the same gain as the full one.
func FinishAudio(wav, out, chain string, target, tp, lufs float64) (AudioResult, error) {
	r := AudioResult{File: out, Target: target}
	gain := target - lufs
	if math.IsInf(lufs, -1) || math.IsNaN(lufs) || lufs < -70 {
		gain = 0 // silence: don't amplify noise
	}
	r.GainDB = math.Round(gain*100) / 100
	limit := math.Pow(10, tp/20)
	af := fmt.Sprintf("volume=%.2fdB,alimiter=limit=%.4f:level=0:latency=1:attack=5:release=50,aresample=%d", gain, limit, SR)
	if chain != "" {
		af = chain + "," + af
	}
	tmp := out + ".tmp.m4a"
	if _, err := deps.Run("ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-i", wav, "-af", af,
		"-ac", "2", "-ar", strconv.Itoa(SR), "-c:a", "aac", "-b:a", "192k", tmp); err != nil {
		return r, err
	}
	if err := os.Rename(tmp, out); err != nil {
		return r, err
	}
	r.LUFS, r.TruePeak, _ = Loudness(out, "")
	return r, nil
}

func ensureDir(p string) error { return os.MkdirAll(filepath.Dir(p), 0o755) }
