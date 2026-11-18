# BACKLOG

Top = next. Tags: `[friction]` from an eval (quote the agent), `[lesson Lxx]`, `[spec]`.
Move done items to the journal, not here.

## Needs human
- Final name (working name: cutroom) and license (proposed: Apache-2.0).

## NOW
- **Eval round 4** (all 5 tasks × haiku/sonnet). Round 3 was clean (10/10, no medium friction); golden
  criterion 3 needs round 4 clean too. Then T3 real footage and the other golden criteria.

## M0 — leftovers
- Fixtures on Linux: espeak-ng path untested.

## M1 — perceive
- T3 sanity: run `transcribe` on a 2-min excerpt of real footage (local only) and compare phrase edges / suspects with what the published edit cut; tune envelope threshold for breaths and room tone (synthetic audio is too clean).
- HDR fixture (HEVC 10-bit HLG tags via libx265) so `inspect` hdr warning has a test. [L23]
- `inspect`: multi-take concat pts offset (first frame pts ≠ 0) fixture. [L04]
- `status`.

## NEXT
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
