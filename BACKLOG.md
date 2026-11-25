# BACKLOG

Top = next. Tags: `[friction]` from an eval (quote the agent), `[lesson Lxx]`, `[spec]`.
Move done items to the journal, not here.

## Needs human
- (none) — name `cutroom` and Apache-2.0 confirmed 2026-11-24; target is the full SPEC v1.

## NOW
- export: validate the .otio with the real OpenTimelineIO reader once it can be installed (it isn't
  a declared dependency); today's guarantee is the round-trip tests.
- Golden 4: T3 reviewer pass on the local real-footage excerpts.
- Golden: final eval round 11 after T3 — all five criteria must hold in the same round (rounds 9–10
  predate iterations 45–47). Rate friction low/medium/high in the round summary (criterion 3).
- Golden 1 is code-complete (iteration 45); confirm with an eval round that exercises the new
  commands (split/pin/--retakes/--output/--overlays/validate) before calling it met.
## M0 — leftovers
- Fixtures on Linux: espeak-ng path untested.

## M1 — perceive

## NEXT
- [friction T3 r50] sec: "the `timeline` header still says 'sound from sec-take1'" after `edit audio mic1/mic2
  --for` — name each take's sound source in the header.
- [friction T3 r50] sec: "camera vs mic word times are in different clocks; I converted by hand" — the
  recorder's transcript view should say that after `edit audio`, `cutroom transcript <camera>` shows
  these same words on the camera's clock (quote those in edits).
- [friction T3 r50] mps: "make `ends-mid-speech` suggest moving a pickup that completes the sentence
  when one exists" (the editor found the earlier take that continues it and moved it to the end).
- [friction T3 r50] started: `transcript --words --from/--to` still capped at 40 lines; a range should
  print in full.
- [friction T3 r50] deploy: "$5 a month" transcribed as "a$5" — normalise currency so check doesn't
  warn on it (seen in r48 and r50).
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
