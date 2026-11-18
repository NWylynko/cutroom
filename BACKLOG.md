# BACKLOG

Top = next. Tags: `[friction]` from an eval (quote the agent), `[lesson Lxx]`, `[spec]`.
Move done items to the journal, not here.

## Needs human
- Final name (working name: cutroom) and license (proposed: Apache-2.0).

## M0 — leftovers
- Fixtures on Linux: espeak-ng path untested.

## M1 — perceive
- T3 sanity: run `transcribe` on a 2-min excerpt of real footage (local only) and compare phrase edges / suspects with what the published edit cut; tune envelope threshold for breaths and room tone (synthetic audio is too clean).
- HDR fixture (HEVC 10-bit HLG tags via libx265) so `inspect` hdr warning has a test. [L23]
- `inspect`: multi-take concat pts offset (first frame pts ≠ 0) fixture. [L04]
- `status`.

## M2 — decide/act
- **First T2 eval round right after `render` works** (moved from iteration 5: no task is completable before rendering).
- `look --cuts` / `look --output` once a timeline exists (frames either side of every cut).
- `edit zoom` (punch-ins with anchor; warn > 1.2× on 1080p) once render exists. [L40]
- `edit move` / reorder and multi-source (takes) segments. 
- `look`: untagged HD sources are converted with bt601 by ffmpeg but shown as bt709 by players; assume bt709 for untagged ≥720p (in_color_matrix) so `look` matches renders.
- Render: segments from a source other than the audio source (multi-take); per-source audio mapping.
- Render: zoom/punch-in filters. Overlay inputs (M4).

## M3 — verify
- `check`: duration, re-transcribe vs timeline text, per-segment A/V drift, frame match, tail pads, loudness. [L06 L20 L33 L37]
- First T2 eval round.

## M4 — overlays & text
- Built-in text rasteriser (embedded font), text/image/clip overlays anchored to text; caption clip. [L47 L48 L16 L19]
- `captions` (SRT) with seam checks. [L49]

## M5 — multi-source & shorts
- `sync` (offset + drift). [L05] Multi-take timelines. HDR tone-map path. [L23 L24 L25]
- `short` preset + safe-zone check. [L41]

## M6 — interchange & polish
- `export --otio --edl`. README, install, `doctor` on clean PATH. T3 real-footage evals.
