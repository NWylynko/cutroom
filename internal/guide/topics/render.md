<!-- summary: rendering previews and finals; what is automatic; caching and time estimates -->
# Render

`cutroom render` turns the timeline into `out/final.mp4` (or `out/preview.mp4` with `--preview`).

- **Preview first.** `--preview` renders 540p with a fast encode. Check it, then render the final.
- **Ask before long work.** `--dry-run` shows the chunks, which are already cached, and a time estimate
  (it learns from this project's previous renders).
- **Output path.** `cutroom render out/x.mp4` or `--output out/x.mp4`; default out/final.mp4
  (out/preview.mp4 with --preview, *-range.mp4 with --range).
- **Partial renders.** `--range 0:20-0:40` renders only the chunks covering that *output* range.
- **Caching.** The video is rendered in ~45 s chunks split at cuts. Re-rendering after an edit only
  redoes the chunks that changed; with no changes a render takes about a second.

## Output size
The final renders at the timeline's size, which starts as the source's: a 4K phone video renders 4K.
For a 1080p video set it first — `cutroom edit frame --size 1080p` (or `render --size 1080p`, which
saves the same change; both are undoable). `720p`, `1440p`, `2160p` and `WxH` work too; a vertical
timeline gets 1080x1920 for `1080p`. Overlays, captions and zoom follow the size. `render --dry-run`
prints the size before you spend the time.

## Handled for you
- Cuts land on whole frames and on the exact matching audio sample; audio and video stay in sync
  even for variable-frame-rate phone footage and sources whose audio starts late.
- Rotated phone footage is turned upright in the pixels; no rotation metadata is written, so every
  player shows it the same way.
- Audio: 20 ms fades at every cut (no clicks), a speech chain (rumble cut, gentle compression, a
  shelf against harshness), then loudness measured and set to `--loudness` (default -16 LUFS) with
  a peak ceiling of `--true-peak` (-1.5 dBTP). Use `--audio-chain none` for music or already-mastered audio.
- Output: H.264 (CRF 18 final, 26 preview), AAC 192 kb/s stereo 48 kHz, bt709 colour tags, faststart.

The render prints its duration next to the timeline's. They must match; if they don't, it says so.

## Handing the edit to another editor
`cutroom export --otio --edl` writes `out/edit.otio` (OpenTimelineIO: picture, overlays and burned
captions as their image files, and the sound — a recorder's at its own clock) and `out/edit.edl`
(CMX 3600: a B event per segment, or V + A when the sound comes from a recorder; overlays only as
comments). Both are in whole frames of the timeline's rate, referencing sources where they are.
