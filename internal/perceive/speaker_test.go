package perceive

import (
	"math"
	"testing"
)

// a skin-coloured "face" that moves a little around x = 0.7 on a grey wall, with a static
// skin-coloured poster on the left (it doesn't move, so it weighs less)
func synthFrames(faceX float64, n int) [][]byte {
	var fs [][]byte
	for k := range n {
		f := make([]byte, spW*spH*3)
		for i := range spW * spH {
			f[3*i], f[3*i+1], f[3*i+2] = 120, 120, 125
		}
		put := func(cx, cy, r int) {
			for y := cy - r; y <= cy+r; y++ {
				for x := cx - r; x <= cx+r; x++ {
					if x >= 0 && y >= 0 && x < spW && y < spH && (x-cx)*(x-cx)+(y-cy)*(y-cy) <= r*r {
						i := y*spW + x
						f[3*i], f[3*i+1], f[3*i+2] = 220, 160, 130
					}
				}
			}
		}
		put(int(faceX*spW)+int(3*math.Sin(float64(k))), 30+k%3, 14)
		put(20, 40, 10) // poster: still
		fs = append(fs, f)
	}
	return fs
}

func TestSpeakerFoundWhereTheFaceMoves(t *testing.T) {
	sp := speakerIn(synthFrames(0.7, 16))
	if !sp.Found || math.Abs(sp.X-0.7) > 0.05 {
		t.Errorf("speaker at %+v, want x≈0.70", sp)
	}
}

func TestNoSpeakerInAPattern(t *testing.T) {
	var fs [][]byte
	for range 16 {
		f := make([]byte, spW*spH*3)
		for i := range spW * spH {
			f[3*i], f[3*i+1], f[3*i+2] = 30, byte(i%200), 200
		}
		fs = append(fs, f)
	}
	if sp := speakerIn(fs); sp.Found || sp.X != 0.5 {
		t.Errorf("a pattern with no skin: %+v, want not found at 0.5", sp)
	}
}

func TestSpeakerFaceBox(t *testing.T) {
	sp := speakerIn(synthFrames(0.7, 16))
	f := sp.Face
	if f[0] > 0.66 || f[2] < 0.74 || f[1] > 0.25 || f[3] < 0.4 {
		t.Errorf("face box %v should cover the moving skin blob around x 0.70, y 0.18–0.5", f)
	}
}
