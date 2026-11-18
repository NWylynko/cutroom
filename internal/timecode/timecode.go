// Package timecode holds cutroom's time model. Times are exact rationals internally; every cut
// resolves to an integer frame and an integer sample before it reaches ffmpeg (lesson L01:
// float-second trims dropped frames and drifted A/V by up to 113 ms over 14 cuts).
package timecode

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// Rate is a frame rate as a reduced fraction, e.g. 30000/1001.
type Rate struct{ Num, Den int64 }

// ParseRate accepts "30", "29.97", "30000/1001".
func ParseRate(s string) (Rate, error) {
	s = strings.TrimSpace(s)
	if n, d, ok := strings.Cut(s, "/"); ok {
		a, err1 := strconv.ParseInt(n, 10, 64)
		b, err2 := strconv.ParseInt(d, 10, 64)
		if err1 != nil || err2 != nil || a <= 0 || b <= 0 {
			return Rate{}, fmt.Errorf("invalid frame rate %q (want e.g. 30, 29.97 or 30000/1001)", s)
		}
		return reduce(a, b), nil
	}
	switch s {
	case "29.97":
		return Rate{30000, 1001}, nil
	case "59.94":
		return Rate{60000, 1001}, nil
	case "23.976", "23.98":
		return Rate{24000, 1001}, nil
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok || r.Sign() <= 0 {
		return Rate{}, fmt.Errorf("invalid frame rate %q (want e.g. 30, 29.97 or 30000/1001)", s)
	}
	return reduce(r.Num().Int64(), r.Denom().Int64()), nil
}

func reduce(a, b int64) Rate {
	g := gcd(a, b)
	return Rate{a / g, b / g}
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}

func (r Rate) String() string {
	if r.Den == 1 {
		return strconv.FormatInt(r.Num, 10)
	}
	return fmt.Sprintf("%d/%d", r.Num, r.Den)
}

// FPS is the rate as a float, for display only.
func (r Rate) FPS() float64 { return float64(r.Num) / float64(r.Den) }

// Time is an exact time in seconds.
type Time struct{ r *big.Rat }

func FromRat(r *big.Rat) Time { return Time{new(big.Rat).Set(r)} }

// Seconds builds a Time from a decimal string exactly ("12.345" is 12345/1000, not a float).
func Seconds(s string) (Time, error) {
	r, ok := new(big.Rat).SetString(strings.TrimSpace(s))
	if !ok {
		return Time{}, fmt.Errorf("not a number: %q", s)
	}
	return Time{r}, nil
}

// FromFloat is for values that are already floats (e.g. from a transcriber). Rounded to 1 µs.
func FromFloat(f float64) Time {
	return Time{big.NewRat(int64(math.Round(f*1e6)), 1e6)}
}

func (t Time) rat() *big.Rat {
	if t.r == nil {
		return new(big.Rat)
	}
	return t.r
}

func (t Time) Float() float64 { f, _ := t.rat().Float64(); return f }

func (t Time) Add(u Time) Time { return Time{new(big.Rat).Add(t.rat(), u.rat())} }
func (t Time) Sub(u Time) Time { return Time{new(big.Rat).Sub(t.rat(), u.rat())} }
func (t Time) Cmp(u Time) int  { return t.rat().Cmp(u.rat()) }

// Frame returns the index of the frame slot containing t on a CFR grid (round to nearest
// frame start: a cut at t keeps frames whose start is ≥ t rounded to the nearest frame).
func (t Time) Frame(r Rate) int64 {
	x := new(big.Rat).Mul(t.rat(), big.NewRat(r.Num, r.Den))
	return roundRat(x)
}

// Sample returns the nearest audio sample index at the given sample rate.
func (t Time) Sample(rate int) int64 {
	x := new(big.Rat).Mul(t.rat(), big.NewRat(int64(rate), 1))
	return roundRat(x)
}

// FrameTime is the exact start time of frame n.
func FrameTime(n int64, r Rate) Time {
	return Time{new(big.Rat).Mul(big.NewRat(n, 1), big.NewRat(r.Den, r.Num))}
}

func roundRat(x *big.Rat) int64 {
	// round half away from zero
	num := new(big.Int).Set(x.Num())
	den := x.Denom()
	neg := num.Sign() < 0
	num.Abs(num)
	q, m := new(big.Int).QuoRem(num, den, new(big.Int))
	if new(big.Int).Mul(m, big.NewInt(2)).Cmp(den) >= 0 {
		q.Add(q, big.NewInt(1))
	}
	if neg {
		q.Neg(q)
	}
	return q.Int64()
}

// Parse accepts "12.5", "1:02.5", "1:00:02.5", "12.5s", or a frame "f372" (needs rate).
func Parse(s string, r Rate) (Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Time{}, fmt.Errorf("empty time")
	}
	if strings.HasPrefix(s, "f") {
		n, err := strconv.ParseInt(s[1:], 10, 64)
		if err != nil || n < 0 {
			return Time{}, fmt.Errorf("invalid frame %q (want f<integer>, e.g. f372)", s)
		}
		if r.Num == 0 {
			return Time{}, fmt.Errorf("frame time %q needs a frame rate", s)
		}
		return FrameTime(n, r), nil
	}
	s = strings.TrimSuffix(s, "s")
	parts := strings.Split(s, ":")
	if len(parts) > 3 {
		return Time{}, fmt.Errorf("invalid time %q (want 12.5, 1:02.5, 1:00:02.5 or f372)", s)
	}
	total := new(big.Rat)
	for i, p := range parts {
		v, ok := new(big.Rat).SetString(p)
		if !ok || v.Sign() < 0 {
			return Time{}, fmt.Errorf("invalid time %q (want 12.5, 1:02.5, 1:00:02.5 or f372)", s)
		}
		if i < len(parts)-1 && !v.IsInt() {
			return Time{}, fmt.Errorf("invalid time %q: only the seconds part may have decimals", s)
		}
		total.Mul(total, big.NewRat(60, 1))
		total.Add(total, v)
	}
	return Time{total}, nil
}

// String formats as seconds with 3 decimals ("62.500").
func (t Time) String() string { return strconv.FormatFloat(t.Float(), 'f', 3, 64) }

// Clock formats as m:ss.mmm (or h:mm:ss.mmm), for humans reading agent output.
func (t Time) Clock() string {
	ms := int64(math.Round(t.Float() * 1000))
	sign := ""
	if ms < 0 {
		sign, ms = "-", -ms
	}
	h, m, s, f := ms/3600000, ms/60000%60, ms/1000%60, ms%1000
	if h > 0 {
		return fmt.Sprintf("%s%d:%02d:%02d.%03d", sign, h, m, s, f)
	}
	return fmt.Sprintf("%s%d:%02d.%03d", sign, m, s, f)
}
