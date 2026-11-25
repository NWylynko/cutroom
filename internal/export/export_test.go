package export

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/cutroom/cutroom/internal/align"
	"github.com/cutroom/cutroom/internal/timeline"
)

func fixture() (*timeline.Timeline, map[string]Source) {
	tl := timeline.New("30", 1280, 720, "cam", "cam", 0, 30)
	tl.Cut("cam", 4, 6, "x", true)
	tl.Cut("cam", 15, 18.5, "x", true)
	return tl, map[string]Source{
		"cam": {Path: "/media/my talk.mp4"},
		"mic": {Path: "/media/mic.wav", Sync: &align.Sync{To: "cam", Offset: 1.7, Drift: -42.8e-6}},
	}
}

// The OTIO export lands on the timeline's frames when read back: V1 clips start at each segment's
// source frame, their durations add up to the edit, A1 takes recorder sound at the recorder's clock.
func TestOTIORoundTrip(t *testing.T) {
	tl, srcs := fixture()
	es := Events(tl, func(string) string { return "mic" }, srcs)
	// listed out of time order, as overlays are when added out of order
	b, err := OTIO("talk", tl, es, srcs, []Card{{Name: "o2", Path: "/c/o2.png", Start: 6, End: 7, Track: "overlays"}, {Name: "o1", Path: "/c/o1.png", Start: 2, End: 4, Track: "overlays"}})
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Schema string `json:"OTIO_SCHEMA"`
		Tracks struct {
			Children []struct {
				Name, Kind string
				Children   []struct {
					Schema      string `json:"OTIO_SCHEMA"`
					SourceRange struct {
						Start struct{ Rate, Value float64 } `json:"start_time"`
						Dur   struct{ Rate, Value float64 } `json:"duration"`
					} `json:"source_range"`
					Media map[string]struct {
						URL string `json:"target_url"`
					} `json:"media_references"`
				}
			}
		}
	}
	if err := json.Unmarshal(b, &doc); err != nil || doc.Schema != "Timeline.1" {
		t.Fatalf("not an OTIO timeline: %v %s", err, doc.Schema)
	}
	byName := map[string]int{}
	for i, tr := range doc.Tracks.Children {
		byName[tr.Name] = i
	}
	v1, a1 := doc.Tracks.Children[byName["V1"]], doc.Tracks.Children[byName["A1"]]
	var total float64
	for i, c := range v1.Children {
		g := tl.Segments[i]
		if c.SourceRange.Start.Value != float64(es[i].SrcIn) || c.SourceRange.Start.Rate != 30 || c.SourceRange.Dur.Value != float64(tl.Frames(g)) {
			t.Errorf("V1 clip %d: %+v vs segment %+v", i, c.SourceRange, g)
		}
		if c.Media["DEFAULT_MEDIA"].URL != "file:///media/my%20talk.mp4" {
			t.Errorf("url %q", c.Media["DEFAULT_MEDIA"].URL)
		}
		total += c.SourceRange.Dur.Value
	}
	if total != tl.Duration()*30 {
		t.Errorf("V1 frames %v, edit %v", total, tl.Duration()*30)
	}
	want := (1.7 + tl.Segments[1].In*(1-42.8e-6)) * 30
	if got := a1.Children[1].SourceRange.Start.Value; got != float64(int64(want+0.5)) {
		t.Errorf("A1 mapped to the recorder clock: %v, want %.0f", got, want)
	}
	ov := doc.Tracks.Children[byName["V2 overlays"]]
	if len(ov.Children) != 4 || ov.Children[0].Schema != "Gap.1" || ov.Children[0].SourceRange.Dur.Value != 60 || ov.Children[1].SourceRange.Dur.Value != 60 || ov.Children[2].SourceRange.Dur.Value != 60 || ov.Children[3].SourceRange.Dur.Value != 30 {
		t.Errorf("overlay track: %+v", ov.Children)
	}
}

var edlEvent = regexp.MustCompile(`^(\d{3})  (\S+)\s+(B|V|A)\s+C\s+(\S+) (\S+) (\S+) (\S+)$`)

func tc(s string) int64 {
	p := strings.Split(s, ":")
	h, _ := strconv.Atoi(p[0])
	m, _ := strconv.Atoi(p[1])
	sec, _ := strconv.Atoi(p[2])
	f, _ := strconv.Atoi(p[3])
	return int64((h*3600+m*60+sec)*30 + f)
}

// The EDL's events cover the edit without gaps from 01:00:00:00, each as long as its segment, with
// sound from the recorder as A events at its own clock.
func TestEDLRoundTrip(t *testing.T) {
	tl, srcs := fixture()
	for _, audio := range []string{"cam", "mic"} {
		es := Events(tl, func(string) string { return audio }, srcs)
		edl := EDL("talk", tl, es, srcs, nil)
		if !strings.HasPrefix(edl, "TITLE: talk\nFCM: NON-DROP FRAME\n") {
			t.Fatalf("header: %q", edl[:40])
		}
		rec := int64(108000)
		n := 0
		for _, l := range strings.Split(edl, "\n") {
			m := edlEvent.FindStringSubmatch(l)
			if m == nil {
				continue
			}
			n++
			src0, src1, rec0, rec1 := tc(m[4]), tc(m[5]), tc(m[6]), tc(m[7])
			if src1-src0 != rec1-rec0 {
				t.Errorf("event %s: source %d frames, record %d", m[1], src1-src0, rec1-rec0)
			}
			if m[3] != "A" { // picture events are contiguous
				if rec0 != rec {
					t.Errorf("event %s starts at %d, want %d", m[1], rec0, rec)
				}
				rec = rec1
			}
			if audio == "mic" && m[3] == "A" && m[2] != "MIC" {
				t.Errorf("sound reel %s", m[2])
			}
		}
		want := len(tl.Segments)
		if audio == "mic" {
			want *= 2
		}
		if n != want || rec-108000 != int64(tl.Duration()*30+0.5) {
			t.Errorf("audio %s: %d events (want %d), %d frames (want %.0f)", audio, n, want, rec-108000, tl.Duration()*30)
		}
	}
	if Timecode(108000+1799, 30) != "01:00:59:29" || Reel("talk-take_2") != "TALKTAKE" {
		t.Errorf("timecode/reel: %s %s", Timecode(108000+1799, 30), Reel("talk-take_2"))
	}
}
