package perceive

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	"github.com/NWylynko/cutroom/internal/deps"
	"github.com/NWylynko/cutroom/internal/media"
)

// Speaker is where the person talking is, across a stretch of a video: X is the centre of their
// face and upper body (0–1 across the picture as shown). Found is false when nothing person-like
// moves there (a screen recording, a test pattern): then X is 0.5.
type Speaker struct {
	X     float64 `json:"x"`
	Found bool    `json:"found"`
	Skin  float64 `json:"skin"` // share of the frame that looks like skin, on average
	// Face: the face's box (x0, y0, x1, y1, 0–1 of the picture) — the top of the skin that moves
	// around X, about as tall as it is wide plus a third; zero when no speaker is found
	Face [4]float64 `json:"face"`
}

const spW, spH = 160, 90

// speakerVersion changes whenever FindSpeaker does, so cached results are found again.
const speakerVersion = 3

// FindSpeaker samples 16 frames of path over [start, start+dur) and finds the speaker: pixels that
// look like skin and move (a face talking, hands) weigh most, upper ones more than lower (the face
// over the hands); the centre is the weighted centroid inside the best window as wide as a 9:16
// crop of the picture. Tested on real talking-head footage (within ~0.04 of the face) — a face
// detector it is not, so it says how sure it is through Found.
func FindSpeaker(path string, start, dur float64) Speaker {
	sdr := media.ReadSDR(path, 0, 0)
	var frames [][]byte
	const n = 16
	for i := range n {
		t := start + dur*(float64(i)+0.5)/n
		args := append([]string{"-hide_banner", "-loglevel", "error", "-ss", fmt.Sprintf("%.3f", math.Max(0, t))}, sdr.Pre...)
		args = append(args, "-i", path, "-frames:v", "1", "-vf", sdr.Chain(fmt.Sprintf("scale=%d:%d", spW, spH)), "-pix_fmt", "rgb24", "-f", "rawvideo", "-")
		b, err := deps.Run("ffmpeg", args...)
		if err == nil && len(b) == spW*spH*3 {
			frames = append(frames, b)
		}
	}
	return speakerIn(frames)
}

func isSkin(r, g, b float64) bool {
	y := 0.299*r + 0.587*g + 0.114*b
	cb := 128 - 0.168736*r - 0.331264*g + 0.5*b
	cr := 128 + 0.5*r - 0.418688*g - 0.081312*b
	return y > 40 && cb > 77 && cb < 127 && cr > 135 && cr < 173 && r > g && r > b
}

func speakerIn(frames [][]byte) Speaker {
	sp := Speaker{X: 0.5}
	if len(frames) < 4 {
		return sp
	}
	n := float64(len(frames))
	skin := make([]float64, spW*spH)
	mean := make([]float64, spW*spH)
	for _, f := range frames {
		for i := range spW * spH {
			r, g, b := float64(f[3*i]), float64(f[3*i+1]), float64(f[3*i+2])
			if isSkin(r, g, b) {
				skin[i] += 1 / n
			}
			mean[i] += (0.299*r + 0.587*g + 0.114*b) / n
		}
	}
	mot := make([]float64, spW*spH)
	mm := 1e-9
	for _, f := range frames {
		for i := range spW * spH {
			y := 0.299*float64(f[3*i]) + 0.587*float64(f[3*i+1]) + 0.114*float64(f[3*i+2])
			mot[i] += math.Abs(y-mean[i]) / n
		}
	}
	for _, v := range mot {
		mm = math.Max(mm, v)
	}
	col := make([]float64, spW)
	area := 0.0
	for y := range spH {
		for x := range spW {
			i := y*spW + x
			area += skin[i]
			col[x] += skin[i] * (0.15 + mot[i]/mm) * (1.3 - float64(y)/spH)
		}
	}
	sp.Skin = area / (spW * spH)
	if sp.Skin < 0.02 {
		return sp
	}
	cw := spH * 9 / 16
	best, bx := -1.0, 0
	for x0 := 0; x0+cw <= spW; x0++ {
		s := 0.0
		for x := x0; x < x0+cw; x++ {
			s += col[x]
		}
		if s > best {
			best, bx = s, x0
		}
	}
	sx, sw := 0.0, 0.0
	for x := bx; x < bx+cw; x++ {
		sx += col[x] * float64(x)
		sw += col[x]
	}
	if sw <= 0 {
		return sp
	}
	sp.X = math.Round((sx/sw+0.5)/spW*100) / 100
	sp.Found = true
	// the face: from the first row with skin at the centre column (the forehead, under the hair),
	// its width the median run of skin-coloured pixels around the centre over the rows below that,
	// its height about its width (the run includes the ears)
	cx := int(sx / sw)
	isSk := func(x, y int) bool { return x >= 0 && x < spW && skin[y*spW+x] >= 0.5 }
	top := -1
	for y := range spH {
		n := 0
		for x := cx - 2; x <= cx+2; x++ {
			if isSk(x, y) {
				n++
			}
		}
		if n >= 3 {
			top = y
			break
		}
	}
	if top >= 0 {
		var runs []int
		var lefts []int
		for y := top + 2; y < min(spH, top+14); y++ {
			l, r := cx, cx
			if !isSk(cx, y) {
				continue
			}
			for l > 0 && (isSk(l-1, y) || isSk(l-2, y)) {
				l--
			}
			for r < spW-1 && (isSk(r+1, y) || isSk(r+2, y)) {
				r++
			}
			runs = append(runs, r-l+1)
			lefts = append(lefts, l)
		}
		if len(runs) > 0 {
			idx := make([]int, len(runs))
			for i := range idx {
				idx[i] = i
			}
			sort.Slice(idx, func(a, b int) bool { return runs[idx[a]] < runs[idx[b]] })
			m := idx[len(idx)/2]
			w, x0 := runs[m], lefts[m]
			sp.Face = [4]float64{r2(float64(x0) / spW), r2(float64(top) / spH), r2(float64(x0+w) / spW), r2(math.Min(1, (float64(top)+1.0*float64(w))/spH))}
		}
	}
	return sp
}

func r2(f float64) float64 { return math.Round(f*100) / 100 }

// FindSpeakerCached is FindSpeaker remembered in cacheDir per file and span (timeline warnings
// ask for it on every edit).
func FindSpeakerCached(cacheDir, path string, start, dur float64) Speaker {
	st, err := os.Stat(path)
	if err != nil {
		return FindSpeaker(path, start, dur)
	}
	key := fmt.Sprintf("v%d|%s|%d|%d|%.1f|%.1f", speakerVersion, path, st.Size(), st.ModTime().UnixNano(), start, dur)
	file := filepath.Join(cacheDir, "speakers.json")
	m := map[string]Speaker{}
	if b, err := os.ReadFile(file); err == nil {
		_ = json.Unmarshal(b, &m)
	}
	if sp, ok := m[key]; ok {
		return sp
	}
	sp := FindSpeaker(path, start, dur)
	m[key] = sp
	if b, err := json.Marshal(m); err == nil {
		_ = os.MkdirAll(cacheDir, 0o755)
		_ = os.WriteFile(file, b, 0o644)
	}
	return sp
}
