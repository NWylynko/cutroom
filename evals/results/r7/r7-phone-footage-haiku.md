# Friction report: r7-phone-footage-haiku

- **Outcome:** done. `out/final.mp4` passes `cutroom check` with 0 warnings: the retake, the cat aside and the long pauses are gone, and the file has no rotation metadata (confirmed with ffprobe: 1280x720, h264, no rotate tag or display matrix side data, 13.767 s).

- **Commands run:** 23 shell commands, plus one Read of a look PNG. In order:
  1. `mkdir -p … && ls -la && cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide perceive`, `cutroom guide edit`
  3. `cutroom doctor`, `cutroom guide render`, `cutroom guide verify`, `cutroom guide pitfalls`
  4. `cutroom init talk-rot.mov`, `cutroom inspect`
  5. `cutroom transcribe`, `cutroom transcript`
  6. `cutroom transcript talk-rot --suspects`, `cutroom edit tighten`, `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`
  7. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing."`, `cutroom edit cut --from "By the way" --to "keyboard." --dry-run`
  8. `cutroom edit cut --from "By the way" --to "keyboard."`, `cutroom timeline`, `cutroom timeline --joins`
  9. `cutroom render --preview --dry-run`, `cutroom render --preview`, `cutroom check out/preview.mp4`
  10. `cutroom look --file out/preview.mp4 2.0`, then Read on the PNG
  11. `cutroom render --dry-run`, `cutroom render`, `cutroom check`
  12. `ffprobe` on out/final.mp4 (stream size, rotation side data and tags; duration)

- **Errors hit:** none. No command exited non-zero.

- **Guessed wrong:** nothing failed, but I checked the `--to` semantics against the guide rather than the output. The `retake-kept` warning suggested `--from "The first thing you need" --near 6.31`; I used the full quoted phrase with `--to` instead and it resolved to one match without `--near`. I did not need `--occurrence` or `--near`.

- **Confusing output:**
  - `tighten` prints a long line: `pads 0.22+0.10 s + --min-cut 0.4 s`. I had to work out which pause values were removed. The per-pause list (`0:04.180 2.13→0.32 s …`) was the clearer part.
  - `transcript` tags the keeper of a retake with `[repeat]` as well as the take to drop. The tag alone does not say which take to keep. The warning text does ("the last take is usually the keeper"), but only in `--suspects`.
  - Segment ids skip (s2 and s5 disappear after the cuts). This is expected but unexplained.
  - The `look` frame at 2.0 s is a colour test pattern, so I could not judge orientation from the picture. The "upright" claim in `render` and `check` rests on the tool's own rotation handling, not on anything I could see.

- **Missing:**
  - A way to check orientation independently of cutroom's rotation logic. `check` says "upright, right moment", but it compares the render against the source as the tool itself reads it. If the tool had the rotation direction wrong, `check` would still pass. A rotation-independent check, such as a face or text anchor, or a `look` on the raw stored pixels next to the rendered one, would catch that.
  - `render` does not say which output file a plain `render` will write until the dry run. I used `--dry-run` first, which is fine, but a one-line "will write out/final.mp4" in the normal output would help.

- **Not sure it's right:**
  - Orientation: the 180° rotation is baked into the pixels, and ffprobe shows no rotation metadata. Whether the upright direction is correct is not something I could verify visually from this footage (see above).
  - Edit quality: the transcript came from `base.en`, which mishears words. The `words` check passed (44 of 44), but that compares the render against the same recogniser's transcript, so it cannot catch a mis-heard word that was cut wrongly. I did not run `cutroom listen` on the joins. The join pauses are all about 0.30 to 0.33 s, and the cat aside sat between two clean sentences.
  - I did not inspect the picture of the cut joins with `look --cuts`.

- **Best part:** the `--dry-run` on `edit cut` printed the exact join and the resulting length before I committed, and `render --preview` then `check` gave a clear PASS/FAIL table per check. The loop was easy to follow with no guesswork.

- **One change:** have `check` verify orientation against something other than its own rotation handling, for example by comparing the rendered frame with the raw stored frame rotated by the metadata, plus a `look` that shows raw and rendered side by side. That would let me confirm "upright" rather than trust it.
