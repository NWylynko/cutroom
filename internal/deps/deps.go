// Package deps finds and describes the external tools cutroom shells out to.
package deps

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

type Status struct {
	Name     string   `json:"name"`
	OK       bool     `json:"ok"`
	Required bool     `json:"required"`
	Path     string   `json:"path,omitempty"`
	Version  string   `json:"version,omitempty"`
	Purpose  string   `json:"purpose"`
	Problem  string   `json:"problem,omitempty"`
	Fix      string   `json:"fix,omitempty"`
	Missing  []string `json:"missing_features,omitempty"`
}

// RequiredFilters are ffmpeg filters cutroom's core relies on.
var RequiredFilters = []string{"scdet", "ebur128", "showspectrumpic", "tile", "alimiter", "overlay", "fps", "trim", "atrim", "scale", "crop", "volumedetect", "silencedetect", "highpass", "aresample", "concat", "setpts"}

func install(pkg string) string {
	if runtime.GOOS == "darwin" {
		return "brew install " + pkg
	}
	switch pkg {
	case "whisper.cpp":
		return "build from https://github.com/ggml-org/whisper.cpp (whisper-cli) or install your distro's whisper.cpp package"
	}
	return "install " + pkg + " with your package manager (e.g. apt install " + pkg + ")"
}

func FFmpeg() Status {
	s := Status{Name: "ffmpeg", Required: true, Purpose: "decode, filter, encode"}
	p, err := exec.LookPath("ffmpeg")
	if err != nil {
		s.Problem, s.Fix = "not on PATH", install("ffmpeg")
		return s
	}
	s.Path = p
	out, _ := exec.Command(p, "-hide_banner", "-version").Output()
	if m := regexp.MustCompile(`ffmpeg version (\S+)`).FindSubmatch(out); m != nil {
		s.Version = string(m[1])
	}
	filters, _ := exec.Command(p, "-hide_banner", "-filters").Output()
	have := map[string]bool{}
	for _, line := range strings.Split(string(filters), "\n") {
		f := strings.Fields(line)
		if len(f) >= 2 {
			have[f[1]] = true
		}
	}
	for _, f := range RequiredFilters {
		if !have[f] {
			s.Missing = append(s.Missing, f)
		}
	}
	s.OK = len(s.Missing) == 0
	if !s.OK {
		s.Problem = "missing filters: " + strings.Join(s.Missing, ", ")
		s.Fix = "install a full ffmpeg build (" + install("ffmpeg") + ")"
	}
	return s
}

func FFprobe() Status {
	s := Status{Name: "ffprobe", Required: true, Purpose: "read media facts"}
	p, err := exec.LookPath("ffprobe")
	if err != nil {
		s.Problem, s.Fix = "not on PATH", install("ffmpeg")
		return s
	}
	s.Path, s.OK = p, true
	return s
}

func Whisper() Status {
	s := Status{Name: "whisper-cli", Required: true, Purpose: "speech to text with word timings (whisper.cpp)"}
	p, err := exec.LookPath("whisper-cli")
	if err != nil {
		s.Problem, s.Fix = "not on PATH", install("whisper.cpp")
		return s
	}
	s.Path, s.OK = p, true
	return s
}

// ModelDirs lists where whisper models are searched, in order.
func ModelDirs() []string {
	var dirs []string
	if d := os.Getenv("CUTROOM_MODELS"); d != "" {
		dirs = append(dirs, d)
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".cache", "cutroom", "models"))
	}
	return dirs
}

// ModelURL is the official download location of a whisper.cpp ggml model.
func ModelURL(name string) string {
	return "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-" + name + ".bin"
}

// FindModel returns the path of the preferred available model. Preference: the env override
// CUTROOM_WHISPER_MODEL (a path), then the largest known-good model found in ModelDirs.
func FindModel() (string, []string) {
	if p := os.Getenv("CUTROOM_WHISPER_MODEL"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, []string{p}
		}
	}
	pref := []string{"large-v3-turbo", "large-v3", "medium.en", "small.en", "base.en", "tiny.en"}
	var found []string
	for _, d := range ModelDirs() {
		m, _ := filepath.Glob(filepath.Join(d, "ggml-*.bin"))
		found = append(found, m...)
	}
	sort.Strings(found)
	for _, name := range pref {
		for _, f := range found {
			if filepath.Base(f) == "ggml-"+name+".bin" {
				return f, found
			}
		}
	}
	if len(found) > 0 {
		return found[0], found
	}
	return "", nil
}

func Model() Status {
	s := Status{Name: "whisper-model", Required: true, Purpose: "weights for whisper-cli"}
	p, _ := FindModel()
	if p == "" {
		dir := ModelDirs()[len(ModelDirs())-1]
		s.Problem = "no ggml model in " + strings.Join(ModelDirs(), ", ")
		s.Fix = "mkdir -p " + dir + " && curl -L -o " + filepath.Join(dir, "ggml-base.en.bin") + " " + ModelURL("base.en") +
			"   (148 MB; for best accuracy use large-v3-turbo, 1.6 GB)"
		return s
	}
	s.Path, s.OK = p, true
	s.Version = strings.TrimSuffix(strings.TrimPrefix(filepath.Base(p), "ggml-"), ".bin")
	return s
}

func All() []Status {
	return []Status{FFmpeg(), FFprobe(), Whisper(), Model()}
}

// Run executes a tool and returns stdout, with stderr folded into the error.
func Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > 1500 {
			msg = "…" + msg[len(msg)-1500:]
		}
		return out, &ToolError{Tool: name, Err: err, Stderr: msg}
	}
	return out, nil
}

// RunErr executes a tool whose useful output is on stderr (ffmpeg analysis filters).
func RunErr(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = nil
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > 1500 {
			msg = "…" + msg[len(msg)-1500:]
		}
		return stderr.String(), &ToolError{Tool: name, Err: err, Stderr: msg}
	}
	return stderr.String(), nil
}

type ToolError struct {
	Tool   string
	Err    error
	Stderr string
}

func (e *ToolError) Error() string {
	if e.Stderr != "" {
		return e.Tool + " failed: " + e.Stderr
	}
	return e.Tool + " failed: " + e.Err.Error()
}
