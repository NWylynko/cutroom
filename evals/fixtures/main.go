// Command fixtures generates synthetic test media with planted, known problems and writes a
// ground-truth manifest for each. Deterministic for a given machine (speech from macOS `say`;
// set FIXTURE_TTS=espeak on Linux).
//
//	go run ./evals/fixtures -out evals/work/fixtures
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// A line of the script: speech, or a silence (optionally with a loud noise burst in it).
type line struct {
	Say    string  `json:"say,omitempty"`
	Pause  float64 `json:"pause,omitempty"`
	Burst  bool    `json:"burst,omitempty"`  // a loud 0.25 s noise burst in the middle of the pause
	Breath bool    `json:"breath,omitempty"` // a soft 0.3 s inhale ending 0.25 s before the next line
	Tag    string  `json:"tag,omitempty"`    // ground-truth label: retake-bad, retake-good, keep, aside
	Start  float64 `json:"start"`
	End    float64 `json:"end"`
	Detail string  `json:"detail,omitempty"`
}

type manifest struct {
	Name        string            `json:"name"`
	File        string            `json:"file"`
	Duration    float64           `json:"duration"`
	Lines       []line            `json:"lines"`
	Facts       map[string]any    `json:"facts"`
	Expect      map[string]string `json:"expect"`
	Description string            `json:"description"`
}

// The talk script. Tags are the ground truth an editor should act on.
var talk = []line{
	{Pause: 0.6},
	{Say: "Hi. Today I'm going to show you how to set up a home server.", Tag: "keep"},
	{Pause: 2.0, Detail: "long pause"},
	{Say: "The first thing you need is, um, the first thing", Tag: "retake-bad", Detail: "abandoned start"},
	{Pause: 0.9},
	{Say: "The first thing you need is an old laptop.", Tag: "retake-good"},
	{Pause: 1.4, Breath: true, Detail: "pause with a quiet breath before the next line (normal, not a suspect)"},
	{Say: "Then install Linux on it.", Tag: "keep"},
	{Pause: 3.0, Burst: true, Detail: "loud noise burst in a pause (e.g. a cough or a bump)"},
	{Say: "By the way, my cat just walked across the keyboard.", Tag: "aside"},
	{Pause: 1.0},
	{Say: "Finally, open the terminal and run the update command.", Tag: "keep"},
	{Pause: 1.2},
	{Say: "And that's it. Your server is ready.", Tag: "keep"},
	{Pause: 0.8},
}

var out string

func main() {
	flag.StringVar(&out, "out", "evals/work/fixtures", "output directory")
	flag.Parse()
	must(os.MkdirAll(out, 0o755))
	tmp := filepath.Join(out, "tmp")
	must(os.MkdirAll(tmp, 0o755))

	lines, wav := speech(tmp, talk)
	dur := probeDur(wav)

	base := manifest{Lines: lines, Duration: dur, Expect: map[string]string{
		"must_contain":     "show you how to set up a home server|The first thing you need is an old laptop|Then install Linux on it|open the terminal and run the update command|Your server is ready",
		"must_not_contain": "um, the first thing|my cat",
	}}

	// 1. talk.mp4 — the clean baseline: CFR 30, 1280x720 testsrc (frame counter visible), AAC 48k.
	m := base
	m.Name, m.File = "talk", "talk.mp4"
	m.Description = "Talking clip: long pause, abandoned retake, aside, loud burst in a pause. CFR 30 fps."
	m.Facts = map[string]any{"fps": "30", "vfr": false, "rotation": 0, "hdr": false, "audio_start": 0.0}
	ff("-f", "lavfi", "-i", fmt.Sprintf("testsrc=size=1280x720:rate=30:duration=%.3f", dur), "-i", wav,
		"-map", "0:v", "-map", "1:a", "-c:v", "libx264", "-preset", "veryfast", "-crf", "23", "-pix_fmt", "yuv420p",
		// camera mic: band-limited and noisier than the separate recorder (lets graders tell them apart)
		"-af", "highpass=f=180,lowpass=f=3500,volume=0.8", "-c:a", "aac", "-b:a", "128k", "-ar", "48000", "-shortest", path("talk.mp4"))
	writeManifest(m)

	// 2. talk-vfr.mp4 — dropped frames (VFR) and audio starting 0.085 s after video. [L02]
	m = base
	m.Name, m.File = "talk-vfr", "talk-vfr.mp4"
	m.Description = "Same talk, VFR: 11 dropped-frame gaps; audio starts 0.085 s after the first frame."
	m.Facts = map[string]any{"fps": "30", "vfr": true, "rotation": 0, "hdr": false, "audio_start": 0.085, "dropped_gaps": 11}
	drops := []string{}
	for i := 0; i < 11; i++ {
		n := 40 + i*53
		drops = append(drops, fmt.Sprintf("eq(n,%d)+eq(n,%d)", n, n+1))
	}
	// Audio starts 0.085 s after the first frame but stays in sync by timestamps (as on a real
	// camera): drop its first 0.085 s of content, then stamp it to start at 0.085.
	ff("-i", path("talk.mp4"), "-vn", "-af", "atrim=start=0.085,asetpts=PTS-STARTPTS", "-c:a", "aac", "-ar", "48000", path("tmp/late.m4a"))
	ff("-i", path("talk.mp4"), "-itsoffset", "0.085", "-i", path("tmp/late.m4a"), "-map", "0:v", "-map", "1:a",
		"-vf", "select='not("+strings.Join(drops, "+")+")'", "-fps_mode", "vfr",
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "23", "-c:a", "copy",
		"-video_track_timescale", "90000", path("talk-vfr.mp4"))
	m.Facts["audio_start_stream"] = probeStart(path("talk-vfr.mp4"), "a:0")
	m.Facts["note"] = "audio starts 0.085 s after video, in sync by timestamps; the stream start_time reads lower by the AAC priming delay (1024 samples)"
	writeManifest(m)

	// 3. talk-rot.mov — display matrix says rotate 180 (phone mounted upside down). [L25]
	m = base
	m.Name, m.File = "talk-rot", "talk-rot.mov"
	m.Description = "Same talk, pixels stored upside down with a 180° display rotation in metadata."
	m.Facts = map[string]any{"fps": "30", "vfr": false, "rotation": 180, "hdr": false, "audio_start": 0.0}
	ff("-i", path("talk.mp4"), "-vf", "hflip,vflip", "-c:v", "libx264", "-preset", "veryfast", "-crf", "23", "-c:a", "copy", path("tmp/flipped.mov"))
	ff("-display_rotation", "180", "-i", path("tmp/flipped.mov"), "-c", "copy", path("talk-rot.mov"))
	writeManifest(m)

	// 4. recorder.wav — a separate audio recorder: starts 1.7 s before the camera, clock +50 ppm. [L05]
	m = base
	m.Name, m.File = "recorder", "recorder.wav"
	m.Description = "Separate recorder for talk.mp4: started 1.700 s earlier, clock runs +50 ppm fast."
	m.Facts = map[string]any{"offset": 1.7, "drift_ppm": 50.0, "pairs_with": "talk.mp4"}
	ff("-f", "lavfi", "-i", "anoisesrc=color=pink:amplitude=0.002:duration=1.7:sample_rate=48000",
		"-i", wav, "-filter_complex",
		"[1:a]aresample=48000,asetrate=48000*1.00005,aresample=48000[v];[0:a][v]concat=n=2:v=0:a=1,pan=mono|c0=c0",
		"-ar", "48000", "-c:a", "pcm_s16le", path("recorder.wav"))
	writeManifest(m)

	// 6. pickup.mp4 — one line re-recorded after the talk ("Then install Linux on it."), with 0.5 s
	// either side; picture negated so a grader can tell which take a frame came from.
	m = base
	m.Name, m.File = "pickup", "pickup.mp4"
	m.Description = "A re-recorded line (pickup) of talk.mp4's 'Then install Linux on it.', picture negated."
	var pl line
	for _, l := range lines {
		if strings.HasPrefix(l.Say, "Then install Linux") {
			pl = l
		}
	}
	m.Facts = map[string]any{"fps": "30", "line": pl.Say, "talk_start": pl.Start - 0.5}
	ff("-ss", fmt.Sprintf("%.3f", pl.Start-0.5), "-t", fmt.Sprintf("%.3f", pl.End-pl.Start+1.0), "-i", path("talk.mp4"),
		"-vf", "negate", "-c:v", "libx264", "-preset", "veryfast", "-crf", "20", "-c:a", "aac", "-b:a", "128k", path("pickup.mp4"))
	writeManifest(m)

	// 5. talk-hdr.mov — iPhone-style HDR: HEVC 10-bit tagged HLG/bt2020, stored upside down with a
	// 180° display rotation (real iPhone footage in T3 is exactly this). [L23 L25]
	m = base
	m.Name, m.File = "talk-hdr", "talk-hdr.mov"
	m.Description = "Same talk as HEVC 10-bit HLG (bt2020) with a 180° display rotation, like iPhone HDR video."
	m.Facts = map[string]any{"fps": "30", "vfr": false, "rotation": 180, "hdr": "hlg", "audio_start": 0.0}
	ff("-i", path("talk.mp4"), "-vf", "hflip,vflip,format=yuv420p10le", "-c:v", "libx265", "-preset", "ultrafast", "-crf", "24",
		"-x265-params", "log-level=error:colorprim=bt2020:transfer=arib-std-b67:colormatrix=bt2020nc",
		"-color_primaries", "bt2020", "-color_trc", "arib-std-b67", "-colorspace", "bt2020nc", "-tag:v", "hvc1",
		"-c:a", "copy", path("tmp/hdr.mov"))
	ff("-display_rotation", "180", "-i", path("tmp/hdr.mov"), "-c", "copy", path("talk-hdr.mov"))
	writeManifest(m)

	// 7. terminal.mp4 — B-roll: 6 s of a different, moving picture (every frame unique) with its own
	// sound (a 6 kHz tone that must not reach a render: B-roll is muted).
	m = base
	m.Name, m.File = "terminal", "terminal.mp4"
	m.Description = "B-roll clip: 6 s, 1280x720 30 fps, mandelbrot zoom (unique frames), a 6 kHz tone as its sound (the talk is band-limited below 3.5 kHz)."
	m.Lines = nil
	m.Facts = map[string]any{"fps": "30", "duration": 6.0}
	ff("-f", "lavfi", "-i", "mandelbrot=s=1280x720:rate=30", "-f", "lavfi", "-i", "sine=frequency=6000:sample_rate=48000",
		"-t", "6", "-af", "volume=-12dB", "-c:v", "libx264", "-preset", "veryfast", "-crf", "20", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", "128k", path("terminal.mp4"))
	writeManifest(m)

	// 8. talk-late.mp4 — the picture starts 43 ms after the sound, as when camera takes are joined by
	// stream copy and the HEVC reorder delay's edit list is dropped. In sync by timestamps. [L04]
	m = base
	m.Name, m.File = "talk-late", "talk-late.mp4"
	m.Description = "Same talk; the first video frame's pts is 0.043 s, audio starts at 0 (in sync by timestamps)."
	m.Facts = map[string]any{"fps": "30", "video_start": 0.043, "audio_start": 0.0}
	ff("-itsoffset", "0.043", "-i", path("talk.mp4"), "-i", path("talk.mp4"), "-map", "0:v", "-map", "1:a",
		"-c", "copy", "-video_track_timescale", "90000", path("talk-late.mp4"))
	m.Facts["video_start_stream"] = probeStart(path("talk-late.mp4"), "v:0")
	writeManifest(m)

	os.RemoveAll(tmp)
	fmt.Println("fixtures in", out)
}

// speech renders each line to 48 kHz mono, concatenates them over a low room-tone bed, and
// returns the lines with measured start/end times.
func speech(tmp string, script []line) ([]line, string) {
	var parts []string
	t := 0.0
	res := make([]line, len(script))
	for i, l := range script {
		p := filepath.Join(tmp, fmt.Sprintf("l%02d.wav", i))
		if l.Say != "" {
			aiff := filepath.Join(tmp, fmt.Sprintf("l%02d.aiff", i))
			tts(l.Say, aiff)
			ff("-i", aiff, "-ar", "48000", "-ac", "1", "-c:a", "pcm_s16le", p)
		} else if l.Burst {
			half := (l.Pause - 0.25) / 2
			ff("-f", "lavfi", "-i", fmt.Sprintf("anullsrc=r=48000:cl=mono:d=%.3f", half),
				"-f", "lavfi", "-i", "anoisesrc=color=brown:amplitude=0.6:duration=0.25:sample_rate=48000",
				"-filter_complex", "[0:a][1:a][0:a]concat=n=3:v=0:a=1", "-ac", "1", "-c:a", "pcm_s16le", p)
		} else if l.Breath {
			// breath: band-limited soft noise ~20 dB below the speech peaks, with a quick fade in/out
			pre := l.Pause - 0.3 - 0.25
			ff("-f", "lavfi", "-i", fmt.Sprintf("anullsrc=r=48000:cl=mono:d=%.3f", pre),
				"-f", "lavfi", "-i", "anoisesrc=color=pink:amplitude=0.25:duration=0.3:sample_rate=48000",
				"-f", "lavfi", "-i", "anullsrc=r=48000:cl=mono:d=0.25",
				"-filter_complex", "[1:a]highpass=f=500,lowpass=f=5000,afade=t=in:d=0.1,afade=t=out:st=0.2:d=0.1[b];[0:a][b][2:a]concat=n=3:v=0:a=1",
				"-ac", "1", "-c:a", "pcm_s16le", p)
		} else {
			ff("-f", "lavfi", "-i", fmt.Sprintf("anullsrc=r=48000:cl=mono:d=%.3f", l.Pause), "-c:a", "pcm_s16le", p)
		}
		d := probeDur(p)
		l.Start, l.End = round3(t), round3(t+d)
		t += d
		res[i] = l
		parts = append(parts, p)
	}
	list := filepath.Join(tmp, "list.txt")
	var b strings.Builder
	for _, p := range parts {
		abs, _ := filepath.Abs(p)
		fmt.Fprintf(&b, "file '%s'\n", abs)
	}
	must(os.WriteFile(list, []byte(b.String()), 0o644))
	joined := filepath.Join(tmp, "joined.wav")
	ff("-f", "concat", "-safe", "0", "-i", list, "-c", "copy", joined)
	wav := filepath.Join(tmp, "voice.wav")
	ff("-i", joined, "-f", "lavfi", "-i", fmt.Sprintf("anoisesrc=color=pink:amplitude=0.003:duration=%.3f:sample_rate=48000:seed=7", t),
		"-filter_complex", "[0:a][1:a]amix=inputs=2:normalize=0:duration=first", "-ar", "48000", "-ac", "1", "-c:a", "pcm_s16le", wav)
	return res, wav
}

func tts(text, out string) {
	if os.Getenv("FIXTURE_TTS") == "espeak" {
		run("espeak-ng", "-w", strings.TrimSuffix(out, ".aiff")+".wav", text)
		ff("-i", strings.TrimSuffix(out, ".aiff")+".wav", out)
		return
	}
	run("say", "-v", "Samantha", "-r", "175", "-o", out, text)
}

func path(name string) string { return filepath.Join(out, name) }

func ff(args ...string) {
	run("ffmpeg", append([]string{"-hide_banner", "-loglevel", "error", "-y"}, args...)...)
}

func run(name string, args ...string) {
	cmd := exec.Command(name, args...)
	if b, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n%s\n", name, strings.Join(args, " "), b)
		os.Exit(1)
	}
}

func probeDur(p string) float64 {
	b, err := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "csv=p=0", p).Output()
	must(err)
	f, err := strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
	must(err)
	return f
}

func probeStart(p, stream string) float64 {
	b, err := exec.Command("ffprobe", "-v", "error", "-select_streams", stream, "-show_entries", "stream=start_time", "-of", "csv=p=0", p).Output()
	must(err)
	f, err := strconv.ParseFloat(strings.TrimSpace(string(b)), 64)
	must(err)
	return f
}

func writeManifest(m manifest) {
	b, _ := json.MarshalIndent(m, "", "  ")
	must(os.WriteFile(path(m.Name+".json"), b, 0o644))
}

func round3(f float64) float64 { return float64(int64(f*1000+0.5)) / 1000 }

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
