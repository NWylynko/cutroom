package speech

import "testing"

// T3 r49 (mps): the recorder placed "and" 0.9 s early, in a noise burst before a pause; the camera
// microphone heard it right before "they". The recorder's word takes the camera's time and "they"
// starts after it.
func TestReconcileTimingFromCamera(t *testing.T) {
	rec := &Transcript{Words: []Word{
		{Text: "communication", Start: 8.973, End: 9.303},
		{Text: "and", Start: 9.933, End: 9.983, Flags: []string{"uncertain-timing"}},
		{Text: "they", Start: 10.833, End: 11.273},
		{Text: "find", Start: 11.273, End: 11.413},
	}}
	cam := &Transcript{Words: []Word{
		{Text: "community,", Start: 8.89, End: 9.28},
		{Text: "and", Start: 10.84, End: 11.08},
		{Text: "they", Start: 11.08, End: 11.30},
		{Text: "find", Start: 11.30, End: 11.44},
	}}
	if n := rec.ReconcileTiming(cam, 1.2); n != 1 {
		t.Fatalf("moved %d words, want 1", n)
	}
	and, they := rec.Words[1], rec.Words[2]
	if and.Start != 10.84 || and.End != 11.08 || they.Start != 11.08 {
		t.Fatalf("and %.3f–%.3f, they from %.3f; want 10.840–11.080, 11.080", and.Start, and.End, they.Start)
	}
	if len(and.Flags) != 1 || and.Flags[0] != "timing-from-camera" {
		t.Fatalf("flags %v", and.Flags)
	}
}

// A word with no agreeing neighbour in the reference (a different utterance), or too far away, stays.
func TestReconcileTimingNeedsAgreement(t *testing.T) {
	rec := &Transcript{Words: []Word{
		{Text: "so", Start: 1.0, End: 1.2},
		{Text: "and", Start: 2.0, End: 2.05, Flags: []string{"uncertain-timing"}},
		{Text: "then", Start: 3.0, End: 3.2},
	}}
	cam := &Transcript{Words: []Word{
		{Text: "but", Start: 2.3, End: 2.4},
		{Text: "and", Start: 2.5, End: 2.6},
		{Text: "now", Start: 2.6, End: 2.8},
		{Text: "so", Start: 5.0, End: 5.2},
		{Text: "and", Start: 5.3, End: 5.4},
	}}
	if n := rec.ReconcileTiming(cam, 1.2); n != 0 || rec.Words[1].Start != 2.0 {
		t.Fatalf("moved %d (and at %.3f); want it left alone", n, rec.Words[1].Start)
	}
}
