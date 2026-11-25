// Package media reads facts about source files with ffprobe and turns the ones that matter for
// editing into warnings with fixes (LESSONS L02 L04 L23 L25 L27).
package media

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/cutroom/cutroom/internal/deps"
)

type Video struct {
	Codec       string  `json:"codec"`
	Profile     string  `json:"profile,omitempty"`
	Width       int     `json:"width"`  // as displayed (after rotation)
	Height      int     `json:"height"` // as displayed (after rotation)
	StoredSize  string  `json:"stored_size,omitempty"`
	FPS         string  `json:"fps"`       // nominal, e.g. "30" or "30000/1001"
	AvgFPS      float64 `json:"avg_fps"`   // frames / duration
	Frames      int     `json:"frames"`    // packets counted
	VFR         bool    `json:"vfr"`       // irregular frame timing
	Gaps        int     `json:"gaps"`      // frame intervals > 1.5× nominal (dropped frames)
	MaxGap      float64 `json:"max_gap_s"` // longest interval between frames
	Start       float64 `json:"start"`     // first frame pts (s)
	Rotation    int     `json:"rotation"`  // degrees the player rotates (display matrix), 0/90/180/270
	PixFmt      string  `json:"pix_fmt"`   // e.g. yuv420p, yuv420p10le
	BitDepth    int     `json:"bit_depth"` // 8, 10…
	Transfer    string  `json:"transfer"`  // color_transfer
	Primaries   string  `json:"primaries"` // color_primaries
	Matrix      string  `json:"matrix"`    // color_space
	HDR         string  `json:"hdr"`       // "", "hlg", "pq"
	DolbyVision bool    `json:"dolby_vision"`
	Duration    float64 `json:"duration"`
}

type Audio struct {
	Codec      string  `json:"codec"`
	SampleRate int     `json:"sample_rate"`
	Channels   int     `json:"channels"`
	Layout     string  `json:"layout,omitempty"`
	Start      float64 `json:"start"` // first sample pts (s)
	Duration   float64 `json:"duration"`
	LUFS       float64 `json:"lufs"`      // integrated loudness; NaN if not measured
	PeakDB     float64 `json:"true_peak"` // dBTP
	Streams    int     `json:"audio_streams"`
	Measured   bool    `json:"measured"`
	SilentHead float64 `json:"-"`
}

// End is where the source's usable content ends: the earliest end of its picture and its sound.
// A container can run past both (a data or metadata track); edits never reach past End.
func (f *Facts) End() float64 {
	end := f.Duration
	if v := f.Video; v != nil && v.Duration > 0 {
		end = math.Min(end, v.Start+v.Duration)
	}
	if a := f.Audio; a != nil && a.Duration > 0 && f.Video != nil {
		end = math.Min(end, a.Start+a.Duration)
	}
	return end
}

type Warning struct {
	Code    string `json:"code"`
	Lesson  string `json:"lesson,omitempty"`
	Message string `json:"message"`
	Handled bool   `json:"handled"` // cutroom deals with it automatically
	Fix     string `json:"fix,omitempty"`
}

type Facts struct {
	Path      string    `json:"path"`
	Kind      string    `json:"kind"` // "video" (with or without audio) or "audio"
	Container string    `json:"container"`
	Size      int64     `json:"size_bytes"`
	Duration  float64   `json:"duration"`
	Video     *Video    `json:"video,omitempty"`
	Audio     *Audio    `json:"audio,omitempty"`
	AVOffset  float64   `json:"av_offset"` // audio start − video start (s)
	Warnings  []Warning `json:"warnings"`
}

type probeOut struct {
	Format struct {
		FormatName string `json:"format_name"`
		Duration   string `json:"duration"`
		Size       string `json:"size"`
	} `json:"format"`
	Streams []struct {
		Index         int    `json:"index"`
		CodecType     string `json:"codec_type"`
		CodecName     string `json:"codec_name"`
		Profile       string `json:"profile"`
		Width         int    `json:"width"`
		Height        int    `json:"height"`
		RFrameRate    string `json:"r_frame_rate"`
		AvgFrameRate  string `json:"avg_frame_rate"`
		PixFmt        string `json:"pix_fmt"`
		ColorTransfer string `json:"color_transfer"`
		ColorPrim     string `json:"color_primaries"`
		ColorSpace    string `json:"color_space"`
		BitsRaw       string `json:"bits_per_raw_sample"`
		StartTime     string `json:"start_time"`
		Duration      string `json:"duration"`
		NbFrames      string `json:"nb_frames"`
		SampleRate    string `json:"sample_rate"`
		Channels      int    `json:"channels"`
		ChannelLayout string `json:"channel_layout"`
		Disposition   struct {
			AttachedPic int `json:"attached_pic"`
		} `json:"disposition"`
		SideData []struct {
			Type     string  `json:"side_data_type"`
			Rotation float64 `json:"rotation"`
		} `json:"side_data_list"`
	} `json:"streams"`
}

// Options control the slower measurements.
type Options struct {
	Loudness bool // measure integrated loudness (decodes all audio)
}

func Probe(path string, opt Options) (*Facts, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	out, err := deps.Run("ffprobe", "-v", "error", "-show_format", "-show_streams", "-of", "json", path)
	if err != nil {
		return nil, err
	}
	var p probeOut
	if err := json.Unmarshal(out, &p); err != nil {
		return nil, fmt.Errorf("ffprobe output: %w", err)
	}
	abs, _ := filepath.Abs(path)
	f := &Facts{Path: abs, Container: p.Format.FormatName, Size: st.Size(), Duration: num(p.Format.Duration)}
	nAudio := 0
	for _, s := range p.Streams {
		switch {
		case s.CodecType == "video" && s.Disposition.AttachedPic == 0 && f.Video == nil:
			v := &Video{Codec: s.CodecName, Profile: s.Profile, Width: s.Width, Height: s.Height, FPS: s.RFrameRate,
				PixFmt: s.PixFmt, Transfer: s.ColorTransfer, Primaries: s.ColorPrim, Matrix: s.ColorSpace,
				Start: num(s.StartTime), Duration: num(s.Duration)}
			if v.Duration == 0 {
				v.Duration = f.Duration
			}
			v.BitDepth = 8
			if b, _ := strconv.Atoi(s.BitsRaw); b > 0 {
				v.BitDepth = b
			} else if m := regexp.MustCompile(`p(\d+)(le|be)$`).FindStringSubmatch(s.PixFmt); m != nil {
				v.BitDepth, _ = strconv.Atoi(m[1])
			}
			for _, sd := range s.SideData {
				if sd.Type == "Display Matrix" {
					// ffprobe reports the counter-clockwise rotation to apply; normalise to 0..359 clockwise-from-player view.
					r := int(math.Round(-sd.Rotation))
					v.Rotation = ((r % 360) + 360) % 360
				}
				if strings.Contains(sd.Type, "DOVI") || strings.Contains(sd.Type, "Dolby Vision") {
					v.DolbyVision = true
				}
			}
			if v.Rotation == 90 || v.Rotation == 270 {
				v.StoredSize = fmt.Sprintf("%dx%d", v.Width, v.Height)
				v.Width, v.Height = v.Height, v.Width
			}
			switch v.Transfer {
			case "arib-std-b67":
				v.HDR = "hlg"
			case "smpte2084":
				v.HDR = "pq"
			}
			f.Video = v
			if err := frameTiming(path, s.Index, v); err != nil {
				return nil, err
			}
		case s.CodecType == "audio":
			nAudio++
			if f.Audio == nil {
				sr, _ := strconv.Atoi(s.SampleRate)
				f.Audio = &Audio{Codec: s.CodecName, SampleRate: sr, Channels: s.Channels, Layout: s.ChannelLayout,
					Start: num(s.StartTime), Duration: num(s.Duration), LUFS: math.NaN(), PeakDB: math.NaN()}
				if f.Audio.Duration == 0 {
					f.Audio.Duration = f.Duration
				}
			}
		}
	}
	if f.Audio != nil {
		f.Audio.Streams = nAudio
	}
	switch {
	case f.Video != nil:
		f.Kind = "video"
	case f.Audio != nil:
		f.Kind = "audio"
	default:
		return nil, fmt.Errorf("no audio or video streams")
	}
	if f.Video != nil && f.Audio != nil {
		f.AVOffset = round6(f.Audio.Start - f.Video.Start)
	}
	if opt.Loudness && f.Audio != nil {
		measureLoudness(path, f.Audio)
	}
	f.Warnings = warnings(f)
	return f, nil
}

// frameTiming reads packet timestamps (no decoding) to count frames and detect VFR/gaps. [L02]
func frameTiming(path string, index int, v *Video) error {
	out, err := deps.Run("ffprobe", "-v", "error", "-select_streams", strconv.Itoa(index),
		"-show_entries", "packet=pts_time", "-of", "csv=p=0", path)
	if err != nil {
		return err
	}
	var ts []float64
	for _, l := range strings.Split(string(out), "\n") {
		l = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(l), ","))
		if l == "" || l == "N/A" {
			continue
		}
		if t, err := strconv.ParseFloat(l, 64); err == nil {
			ts = append(ts, t)
		}
	}
	sort.Float64s(ts)
	v.Frames = len(ts)
	if len(ts) < 2 {
		return nil
	}
	nominal := rate(v.FPS)
	if nominal <= 0 {
		nominal = float64(len(ts)-1) / (ts[len(ts)-1] - ts[0])
	}
	period := 1 / nominal
	var irregular int
	for i := 1; i < len(ts); i++ {
		d := ts[i] - ts[i-1]
		if d > v.MaxGap {
			v.MaxGap = round6(d)
		}
		if d > 1.5*period {
			v.Gaps++
		}
		if math.Abs(d-period) > 0.25*period {
			irregular++
		}
	}
	span := ts[len(ts)-1] - ts[0] + period
	v.AvgFPS = math.Round(float64(len(ts))/span*1000) / 1000
	// VFR: any dropped-frame gaps, or > 1% irregular intervals, or average rate off nominal by > 0.5%.
	v.VFR = v.Gaps > 0 || float64(irregular) > 0.01*float64(len(ts)) || math.Abs(v.AvgFPS-nominal)/nominal > 0.005
	return nil
}

var ebuRe = regexp.MustCompile(`I:\s+(-?[\d.]+) LUFS`)
var peakRe = regexp.MustCompile(`Peak:\s+(-?[\d.]+|-inf) dBFS`)

func measureLoudness(path string, a *Audio) {
	text, err := deps.RunErr("ffmpeg", "-hide_banner", "-nostats", "-i", path, "-vn", "-af", "ebur128=peak=true", "-f", "null", "-")
	if err != nil {
		return
	}
	if m := ebuRe.FindAllStringSubmatch(text, -1); m != nil {
		a.LUFS, _ = strconv.ParseFloat(m[len(m)-1][1], 64)
		a.Measured = true
	}
	if m := peakRe.FindAllStringSubmatch(text, -1); m != nil {
		if m[len(m)-1][1] == "-inf" {
			a.PeakDB = math.Inf(-1)
		} else {
			a.PeakDB, _ = strconv.ParseFloat(m[len(m)-1][1], 64)
		}
	}
}

func warnings(f *Facts) []Warning {
	w := []Warning{}
	if v := f.Video; v != nil {
		if v.VFR {
			w = append(w, Warning{Code: "vfr", Lesson: "L02", Handled: true,
				Message: fmt.Sprintf("variable frame rate: %d dropped-frame gap(s), longest %.3f s, average %.3f fps", v.Gaps, v.MaxGap, v.AvgFPS),
				Fix:     "nothing to do: renders resample to a constant-rate grid anchored on the audio, which keeps lip sync"})
		}
		if v.Rotation != 0 {
			w = append(w, Warning{Code: "rotated", Lesson: "L25", Handled: true,
				Message: fmt.Sprintf("display metadata rotates the picture %d°; stored pixels are not upright", v.Rotation),
				Fix:     "nothing to do: renders apply the rotation to the pixels and write no rotation metadata. Judge orientation from `cutroom look`, which shows frames as rendered"})
		}
		if v.HDR != "" {
			msg := fmt.Sprintf("HDR source (%s, %s, %d-bit); a plain SDR conversion looks washed out", strings.ToUpper(v.HDR), v.Primaries, v.BitDepth)
			if tm := ToneMapper(); tm != "" {
				w = append(w, Warning{Code: "hdr", Lesson: "L23", Handled: true, Message: msg,
					Fix: "nothing to do: render, look and check tone-map it to SDR bt709 (" + tm + "); `cutroom look` shows frames as they will render"})
			} else {
				w = append(w, Warning{Code: "hdr", Lesson: "L23", Handled: false, Message: msg,
					Fix: "this ffmpeg can't tone-map (needs VideoToolbox on macOS, or the zscale filter: an ffmpeg built with libzimg); without it the output looks washed out. `cutroom doctor` shows how to install one"})
			}
		}
		if v.Width*v.Height > 1920*1080 {
			w = append(w, Warning{Code: "large", Handled: true,
				Message: fmt.Sprintf("%dx%d source; previews render at reduced size, finals at the timeline size (the source's unless you set it)", v.Width, v.Height),
				Fix:     "for a 1080p video: `cutroom edit frame --size 1080p` before the final render"})
		}
	}
	if a := f.Audio; a != nil {
		if f.Video != nil && math.Abs(f.AVOffset) >= 0.001 {
			w = append(w, Warning{Code: "av-offset", Lesson: "L02", Handled: true,
				Message: fmt.Sprintf("the sound track starts %.3f s %s the picture track (when the device began recording each); cutroom keeps that offset, so sound and picture stay as recorded — not a lip-sync error", math.Abs(f.AVOffset), map[bool]string{true: "after", false: "before"}[f.AVOffset > 0]),
				Fix:     "nothing to do: cuts are placed on the audio timeline and video is aligned by timestamps"})
		}
		if a.Streams > 1 {
			w = append(w, Warning{Code: "multi-audio", Message: fmt.Sprintf("%d audio streams; only the first is used", a.Streams)})
		}
		if a.Measured && a.LUFS < -35 {
			w = append(w, Warning{Code: "quiet", Message: fmt.Sprintf("very quiet audio (%.1f LUFS)", a.LUFS),
				Handled: true, Fix: "renders normalise loudness to the output target (default -16 LUFS)"})
		}
		if a.Measured && a.PeakDB > -0.1 {
			w = append(w, Warning{Code: "clipping", Message: fmt.Sprintf("audio peaks at %.1f dBTP; it may be clipped in the recording", a.PeakDB),
				Fix: "listen to the loudest passages with `cutroom listen`; clipping cannot be undone"})
		}
	}
	if end := f.End(); f.Duration-end > 0.1 {
		parts := []string{}
		if f.Video != nil {
			parts = append(parts, fmt.Sprintf("picture %.3f s", f.Video.Start+f.Video.Duration))
		}
		if f.Audio != nil {
			parts = append(parts, fmt.Sprintf("sound %.3f s", f.Audio.Start+f.Audio.Duration))
		}
		w = append(w, Warning{Code: "short-streams", Handled: true,
			Message: fmt.Sprintf("the file is %.3f s long but its streams end earlier (%s; the rest is another track, e.g. metadata)", f.Duration, strings.Join(parts, ", ")),
			Fix:     fmt.Sprintf("nothing to do: edits use 0–%.3f s, where both picture and sound exist", end)})
	}
	if f.Video != nil && f.Audio == nil {
		w = append(w, Warning{Code: "no-audio", Message: "no audio stream", Fix: "add audio from another source with `cutroom init --source id=<file>` and use it as the timeline's audio"})
	}
	return w
}

func rate(s string) float64 {
	if n, d, ok := strings.Cut(s, "/"); ok {
		a, _ := strconv.ParseFloat(n, 64)
		b, _ := strconv.ParseFloat(d, 64)
		if b == 0 {
			return 0
		}
		return a / b
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func num(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func round6(f float64) float64 { return math.Round(f*1e6) / 1e6 }
