package perceive

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cutroom/cutroom/internal/deps"
	"github.com/cutroom/cutroom/internal/raster"
	"github.com/cutroom/cutroom/internal/speech"
	"github.com/cutroom/cutroom/internal/timecode"
)

// Listening is what `listen` reports about a stretch of audio.
type Listening struct {
	Start      float64         `json:"start"`
	End        float64         `json:"end"`
	Image      string          `json:"image"`
	Heard      string          `json:"heard_isolated"`   // whisper on the isolated clip (no prompt, L30)
	Transcript string          `json:"transcript_words"` // words of the main transcript in range
	RoomTone   float64         `json:"room_tone_db"`
	PeakDB     float64         `json:"peak_db"`
	Sounds     []speech.Region `json:"sound_regions"`        // source seconds
	Unworded   []speech.Region `json:"sounds_without_words"` // regions no transcript word overlaps
}

const (
	lW      = 1400 // image width
	specH   = 260
	envH    = 110
	wordsH  = 64
	axisH   = 26
	headerH = 30
)

// Listen analyses [a,b] of src. words/roomTone/speechDB come from the transcript when available.
// cuts are times (in src) to mark as cut points on the image.
func Listen(src string, a, b float64, words []speech.Word, cuts []float64, roomTone, speechDB float64, model, outPNG, tmp string) (*Listening, error) {
	d := b - a
	wav := filepath.Join(tmp, "clip.wav")
	if _, err := deps.Run("ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-ss", f3(a), "-t", f3(d), "-i", src,
		"-vn", "-ac", "1", "-ar", "16000", "-c:a", "pcm_s16le", wav); err != nil {
		return nil, err
	}
	x, err := speech.ReadWav(wav)
	if err != nil || len(x) == 0 {
		return nil, fmt.Errorf("no audio in %.3f–%.3f", a, b)
	}
	env := speech.NewEnvelope(x)
	if !math.IsNaN(roomTone) && roomTone != 0 {
		env.Floor, env.Peak = roomTone, speechDB
		env.Thr = env.Floor + math.Max(10, 0.35*(env.Peak-env.Floor))
	}
	L := &Listening{Start: round3(a), End: round3(b), RoomTone: round3(env.Floor), Sounds: []speech.Region{}, Unworded: []speech.Region{}}
	L.PeakDB = math.Inf(-1)
	for _, v := range env.DB {
		L.PeakDB = math.Max(L.PeakDB, v)
	}
	L.PeakDB = round3(L.PeakDB)
	var inRange []speech.Word
	for _, w := range words {
		if w.End > a && w.Start < b {
			inRange = append(inRange, w)
		}
	}
	var tw []string
	for _, w := range inRange {
		tw = append(tw, w.Text)
	}
	L.Transcript = strings.Join(tw, " ")
	for _, r := range env.Regions() {
		sr := speech.Region{Start: round3(r.Start + a), End: round3(r.End + a), PeakDB: round3(r.PeakDB)}
		L.Sounds = append(L.Sounds, sr)
		covered := false
		for _, w := range inRange {
			if w.Start < sr.End && w.End > sr.Start {
				covered = true
			}
		}
		if !covered {
			L.Unworded = append(L.Unworded, sr)
		}
	}
	if model != "" {
		raw, err := speech.Whisper(model, wav, "", tmp)
		if err == nil {
			var h []string
			for _, w := range raw {
				h = append(h, w.Text)
			}
			L.Heard = strings.Join(h, " ")
		}
	}

	// spectrogram from ffmpeg (no legend; we draw our own axis)
	specOut, err := deps.Run("ffmpeg", "-hide_banner", "-loglevel", "error", "-i", wav, "-lavfi",
		fmt.Sprintf("showspectrumpic=s=%dx%d:legend=0:scale=log:color=intensity:stop=8000", lW, specH),
		"-frames:v", "1", "-f", "image2pipe", "-c:v", "png", "-")
	if err != nil {
		return nil, err
	}
	spec, err := png.Decode(bytes.NewReader(specOut))
	if err != nil {
		return nil, err
	}
	H := headerH + specH + envH + wordsH + axisH
	img := image.NewRGBA(image.Rect(0, 0, lW, H))
	raster.Fill(img, img.Bounds(), raster.Panel)
	mono := raster.Face("mono", 13)
	raster.Text(img, raster.Face("bold", 15), 8, 20, raster.White, fmt.Sprintf("listen %s – %s  (%.2f s)", clk(a), clk(b), d))
	raster.Text(img, mono, lW-470, 20, raster.Grey, "spectrogram 0–8 kHz · loudness dB · words · time")
	y0 := headerH
	draw.Draw(img, image.Rect(0, y0, lW, y0+specH), spec, spec.Bounds().Min, draw.Src)
	xOf := func(t float64) int { return int(math.Round((t - a) / d * float64(lW))) }

	// envelope
	ye := y0 + specH
	raster.Fill(img, image.Rect(0, ye, lW, ye+envH), color.RGBA{16, 16, 18, 255})
	lo, hi := env.Floor-6, math.Max(env.Peak+6, env.Floor+30)
	yOf := func(db float64) int {
		v := (db - lo) / (hi - lo)
		v = math.Max(0, math.Min(1, v))
		return ye + envH - 4 - int(v*float64(envH-8))
	}
	isUnworded := func(r speech.Region) bool {
		for _, u := range L.Unworded {
			if u == r {
				return true
			}
		}
		return false
	}
	for _, rg := range L.Sounds {
		c := color.RGBA{20, 83, 45, 255} // dark green: sound with words
		if isUnworded(rg) {
			c = color.RGBA{127, 29, 29, 255} // dark red: sound with no words
		}
		raster.Fill(img, image.Rect(xOf(rg.Start), ye, xOf(rg.End), ye+envH), c)
	}
	n := len(env.DB)
	for i := 1; i < n; i++ {
		x1 := int(float64(i-1) / float64(n) * float64(lW))
		x2 := int(float64(i) / float64(n) * float64(lW))
		yA, yB := yOf(env.DB[i-1]), yOf(env.DB[i])
		if yA > yB {
			yA, yB = yB, yA
		}
		raster.Fill(img, image.Rect(x1, yA, max(x2, x1+1), yB+1), raster.Cyan)
	}

	for _, line := range []struct {
		db float64
		c  color.RGBA
		s  string
	}{{env.Floor, raster.Grey, "room tone"}, {env.Thr, raster.Amber, "speech threshold"}} {
		y := yOf(line.db)
		for x := 0; x < lW; x += 6 {
			raster.Fill(img, image.Rect(x, y, x+3, y+1), line.c)
		}
		lbl := fmt.Sprintf("%s %.0f dB", line.s, line.db)
		raster.Label(img, raster.Face("mono", 11), lW-raster.Measure(raster.Face("mono", 11), lbl)-6, y-3, line.c, color.RGBA{16, 16, 18, 230}, lbl)
	}
	// words
	yw := ye + envH
	lastLabelEnd := -100
	for i, w := range inRange {
		x1, x2 := max(0, xOf(w.Start)), min(lW, xOf(w.End))
		c := color.RGBA{59, 130, 246, 255}
		if i%2 == 1 {
			c = color.RGBA{37, 99, 235, 255}
		}
		raster.Fill(img, image.Rect(x1, yw+6, max(x1+1, x2-1), yw+30), c)
		raster.Text(img, mono, x1+2, yw+23, raster.White, fit(w.Text, max(8, x2-x1-4)))
		if lbl := strconv.FormatFloat(w.Start, 'f', 2, 64); x1 >= lastLabelEnd+4 {
			raster.Text(img, raster.Face("mono", 11), x1+1, yw+46, raster.Grey, lbl)
			lastLabelEnd = x1 + 1 + raster.Measure(raster.Face("mono", 11), lbl)
		}
	}
	for _, rg := range L.Unworded {
		raster.Text(img, mono, xOf(rg.Start)+2, yw+60, raster.Red, "sound, no words")
	}

	// cut markers across spectrogram and loudness
	for _, ct := range cuts {
		if ct > a && ct < b {
			x := xOf(ct)
			raster.Fill(img, image.Rect(x-1, headerH, x+1, yw), raster.Red)
			raster.Label(img, raster.Face("mono", 11), x+4, headerH+14, raster.White, raster.Red, "cut")
		}
	}
	// axis
	ya := yw + wordsH
	step := niceStep(d)
	for t := math.Ceil(a/step) * step; t <= b+1e-9; t += step {
		x := xOf(t)
		raster.Fill(img, image.Rect(x, ya, x+1, ya+6), raster.Grey)
		raster.Text(img, mono, x+3, ya+18, raster.Grey, clk(t))
	}
	if err := raster.SavePNG(outPNG, img); err != nil {
		return nil, err
	}
	L.Image = outPNG
	return L, nil
}

func niceStep(d float64) float64 {
	for _, s := range []float64{0.1, 0.25, 0.5, 1, 2, 5, 10} {
		if d/s <= 12 {
			return s
		}
	}
	return 10
}

func clk(t float64) string     { return timecode.FromFloat(t).Clock() }
func f3(t float64) string      { return strconv.FormatFloat(t, 'f', 3, 64) }
func round3(f float64) float64 { return math.Round(f*1000) / 1000 }
