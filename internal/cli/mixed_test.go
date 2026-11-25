package cli

import (
	"strings"
	"testing"

	"github.com/cutroom/cutroom/internal/speech"
	"github.com/cutroom/cutroom/internal/timeline"
)

// A take's own sound next to the main one: a note when levels and room match, a warning that tells
// a gain difference (speech and room moved together) from another mic or room (their gap changed).
func TestAudioMixedClassifies(t *testing.T) {
	main := &speech.Transcript{SpeechDB: -16.5, RoomTone: -72.7}
	g := timeline.Segment{ID: "s3", Source: "take2", OutStart: 10}
	for _, c := range []struct {
		speech, room float64
		level, want  string
	}{
		{-16.6, -72.7, "info", "they match"},
		{-28.5, -84.7, "warn", "same setup, 12.0 dB louder/quieter"},
		{-17.0, -55.0, "warn", "different room or microphone"},
	} {
		e := &editEnv{tr: timeline.Transcripts{"talk": main, "take2": {SpeechDB: c.speech, RoomTone: c.room}}}
		is := e.mixedIssue(g, "talk")
		if is.Level != c.level || !strings.Contains(is.Message+" "+is.Fix, c.want) {
			t.Errorf("speech %.1f room %.1f: %s %q / %q, want %s %q", c.speech, c.room, is.Level, is.Message, is.Fix, c.level, c.want)
		}
	}
}
