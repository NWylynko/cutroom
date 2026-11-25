# BACKLOG

Top = next. Tags: `[friction]` from an eval (quote the agent), `[lesson Lxx]`, `[spec]`.
Move done items to the journal, not here.

## Needs human
- (none) — name `cutroom` and Apache-2.0 confirmed 2026-11-24; target is the full SPEC v1.

## NOW
- M4: clip overlays (a video inset / B-roll over the speaker), anchored like text/image overlays.
- Eval round 7 (due: iteration 31 was 5 after round 6's record): add overlay + captions tasks.
- Golden criterion 3 holds (sonnet clean in rounds 3–6); the remaining gap is spec coverage (M4–M6),
  T3 review, doctor/README.
## M0 — leftovers
- Fixtures on Linux: espeak-ng path untested.

## M1 — perceive
- `inspect`: multi-take concat pts offset (first frame pts ≠ 0) fixture. [L04]

## NEXT
- [friction r6] Orientation/HDR evidence: "print one explicit 'upright: yes' line backed by an image
  path I can look at" (phone-footage-sonnet; also phone-footage-haiku, hdr-phone-haiku).
- [friction r6] Off-topic aside suspect ("By the way …") (tighten-sonnet, fix-edit-sonnet).
- [friction r6] `listen`: mark low-confidence words in the isolated re-transcription (tighten-haiku:
  "a wrong isolated word looks as authoritative as a correct one").
- [friction r6] `check`: per-segment sync values and the offset/VFR handling applied (vfr-sync-haiku).
- [friction r6] timeline: show the words each cut removes (fix-edit-haiku).
- [friction r5, r6] Acknowledge a warning you decided to keep (`cutroom timeline --accept retake-kept@4.09`
  or similar), so status/check show it as decided (recorder-haiku, phone-haiku; r6 recorder ×2: "make
  the retake decision an explicit choice").
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
- Clip overlays (video B-roll/insets), anchored like text/image overlays. [L16]

## M5 — multi-source & shorts
- Multi-take timelines (segments from several video sources; sync per take). HDR tone-map path. [L23 L24 L25]
- `short` preset + safe-zone check. [L41]

## M6 — interchange & polish
- `export --otio --edl`. README, install, `doctor` on clean PATH. T3 real-footage evals.
