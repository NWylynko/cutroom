# BACKLOG

Top = next. Tags: `[friction]` from an eval (quote the agent), `[lesson Lxx]`, `[spec]`.
Move done items to the journal, not here.

## Needs human
- (none) — name `cutroom` and Apache-2.0 confirmed 2026-11-24; target is the full SPEC v1.

## NOW
- Golden 1: SPEC rows still missing: `edit split`, `edit pin`, `transcript --retakes`, `transcript
  --output` (the edit's words in output time), `look --overlays`, `look --crop`.
- export: validate the .otio with the real OpenTimelineIO reader once it can be installed (it isn't
  a declared dependency); today's guarantee is the round-trip tests.
- Golden 2: audit LESSONS.md — every INVARIANT has a TestLxx guard, every DETECT a fixture.
- Golden 4: T3 reviewer pass on the local real-footage excerpts.
- Golden 5: README + install; `go install` + `doctor` on a clean PATH; a fresh agent follows the README.
- M6: `export --otio --edl`, README + install, `doctor` on a clean PATH.
- Golden criterion 3 holds (sonnet clean in rounds 3–6); the remaining gap is spec coverage (M4–M6),
  T3 review, doctor/README.
## M0 — leftovers
- Fixtures on Linux: espeak-ng path untested.

## M1 — perceive
- `inspect`: multi-take concat pts offset (first frame pts ≠ 0) fixture. [L04]

## NEXT
- [friction r6] Orientation/HDR evidence: "print one explicit 'upright: yes' line backed by an image
  path I can look at" (phone-footage-sonnet; also phone-footage-haiku, hdr-phone-haiku).
- [friction r6] `listen`: mark low-confidence words in the isolated re-transcription (tighten-haiku:
  "a wrong isolated word looks as authoritative as a correct one").
- [friction r6] `check`: per-segment sync values and the offset/VFR handling applied (vfr-sync-haiku).
- [friction r6] timeline: show the words each cut removes (fix-edit-haiku).
- [friction r7] Inherited-edit diagnosis: per sentence kept whole / trimmed / retake / aside (fix-edit ×2).
- [friction r7] HDR: `look --compare` raw vs tone-mapped (hdr-haiku); orientation evidence independent
  of cutroom's own rotation handling (phone-haiku, r6 ×3).
- [friction r5] `edit cut --phrase N` using `cutroom transcript` phrase numbers (tighten-haiku).
- [friction r5] tighten: say which removed suspect sounds were never listened to (vfr-haiku).
- Fixture: add a per-frame marker (frame number box) to talk.mp4 so graders can place frames exactly
  through tone mapping (hdr-phone sync limit is 70 ms because of this).
- Warnings whose text depends on this machine (hdr: tone mapper available?) are frozen into
  cutroom.json at inspect time; recompute `handled`/fix at display time.
- Tone-map timing on 4K: measure render speed of a 60 s 4K HLG excerpt (VideoToolbox) for the
  estimator; zscale path is untested (no zscale in this ffmpeg).
- [friction r4] `edit history` could flag the edit that a timeline warning comes from (fix-edit-sonnet:
  "have `edit history` flag edits that trigger a timeline warning").
- Retake suggestion for 3+ takes: the fix removes only the first kept take; re-warns for the next. Fine,
  but could cut all but the last in one command once `edit cut` accepts several spans.

## M2 — decide/act
- `look`: untagged HD sources are converted with bt601 by ffmpeg but shown as bt709 by players; assume bt709 for untagged ≥720p (in_color_matrix) so `look` matches renders.
- Render: zoom/punch-in filters. Overlay inputs (M4).

## M3 — verify
- `check --frames` full frame-by-frame PSNR vs an overlay-free render once overlays exist (L20, L21).

## M4 — overlays & text
- burned captions: a group can still end on "and"/"is" when the 20-char limit leaves no other
  split ("authentication and | authorization checks"); word-level highlight (karaoke) is not v1.
- captions: line breaks can split an adjective from its noun ("fully | working" on deploy); a
  part-of-speech-free fix would be a small list of common adjective/noun pairs — low priority.

## M5 — multi-source & shorts
- Multi-take timelines (segments from several video sources; sync per take). HDR tone-map path. [L23 L24 L25]

## M6 — interchange & polish
- `export --otio --edl`. README, install, `doctor` on clean PATH. T3 real-footage evals.
