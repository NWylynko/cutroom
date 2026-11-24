# BACKLOG

Top = next. Tags: `[friction]` from an eval (quote the agent), `[lesson Lxx]`, `[spec]`.
Move done items to the journal, not here.

## Needs human
- (none) — name `cutroom` and Apache-2.0 confirmed 2026-11-24; target is the full SPEC v1.

## NOW
- Golden criterion 3 met in rounds 3+4; the remaining gap is spec coverage (M2–M6), T3, doctor/README.
- HDR → SDR tone mapping for HLG/PQ sources (render, look, check picture). 2 of 4 real productions
  are iPhone HLG; today `inspect` says "output will look washed out". VideoToolbox on macOS, zscale
  tonemap fallback. Needs an HDR fixture (libx265 10-bit HLG tags). [L23 L24]
- Breaths: on real footage (deploy excerpt) 6 "sound" suspects are breaths (peaks −39…−43 dB over a
  −55 dB floor). Classify short low sounds next to speech as `breath` (not a suspect by default). [L28]

## M0 — leftovers
- Fixtures on Linux: espeak-ng path untested.

## M1 — perceive
- `inspect`: multi-take concat pts offset (first frame pts ≠ 0) fixture. [L04]

## NEXT
- [friction r4] `edit history` could flag the edit that a timeline warning comes from (fix-edit-sonnet:
  "have `edit history` flag edits that trigger a timeline warning").
- Retake suggestion for 3+ takes: the fix removes only the first kept take; re-warns for the next. Fine,
  but could cut all but the last in one command once `edit cut` accepts several spans.

## M2 — decide/act
- `edit zoom` (punch-ins with anchor; warn > 1.2× on 1080p) once render exists. [L40]
- `edit move` / reorder and multi-source (takes) segments. 
- `look`: untagged HD sources are converted with bt601 by ffmpeg but shown as bt709 by players; assume bt709 for untagged ≥720p (in_color_matrix) so `look` matches renders.
- Render: segments from a source other than the audio source (multi-take); per-source audio mapping.
- Render: zoom/punch-in filters. Overlay inputs (M4).

## M3 — verify
- `check` picture check skips zoomed segments; compare through the zoom once punch-ins exist.
- `check --frames` full frame-by-frame PSNR vs an overlay-free render once overlays exist (L20, L21).

## M4 — overlays & text
- Built-in text rasteriser (embedded font), text/image/clip overlays anchored to text; caption clip. [L47 L48 L16 L19]
- `captions` (SRT) with seam checks. [L49]

## M5 — multi-source & shorts
- Multi-take timelines (segments from several video sources; sync per take). HDR tone-map path. [L23 L24 L25]
- `short` preset + safe-zone check. [L41]

## M6 — interchange & polish
- `export --otio --edl`. README, install, `doctor` on clean PATH. T3 real-footage evals.
