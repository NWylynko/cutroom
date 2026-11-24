# BACKLOG

Top = next. Tags: `[friction]` from an eval (quote the agent), `[lesson Lxx]`, `[spec]`.
Move done items to the journal, not here.

## Needs human
- (none) — name `cutroom` and Apache-2.0 confirmed 2026-11-24; target is the full SPEC v1.

## NOW
- T3 multi-take: securing-personal-software take-1/take-2 (HEVC 4K HLG) + audio-1/audio-2 excerpts:
  sync each recorder to its take, add a line from take 2, render, check. [L24 L05]
- Golden criterion 3 met in rounds 3+4; the remaining gap is spec coverage (M2–M6), T3, doctor/README.
## M0 — leftovers
- Fixtures on Linux: espeak-ng path untested.

## M1 — perceive
- `inspect`: multi-take concat pts offset (first frame pts ≠ 0) fixture. [L04]

## NEXT
- [friction r5] Acknowledge a warning you decided to keep (`cutroom timeline --accept retake-kept@4.09`
  or similar), so status/check show it as decided (recorder-haiku, phone-haiku).
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
- Built-in text rasteriser (embedded font), text/image/clip overlays anchored to text; caption clip. [L47 L48 L16 L19]
- `captions` (SRT) with seam checks. [L49]

## M5 — multi-source & shorts
- Multi-take timelines (segments from several video sources; sync per take). HDR tone-map path. [L23 L24 L25]
- `short` preset + safe-zone check. [L41]

## M6 — interchange & polish
- `export --otio --edl`. README, install, `doctor` on clean PATH. T3 real-footage evals.
