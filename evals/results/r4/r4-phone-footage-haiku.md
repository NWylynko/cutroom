# Friction report: r4-phone-footage-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s, 1280x720) passes `cutroom check` with 0 failures and 0 warnings. The retake and the cat remark are cut, all pauses are 0.32 s, and the picture is upright in the pixels with no rotation metadata.

- **Commands run:** about 22 shell calls in total, 14 of them `cutroom`, the rest verification with ffprobe/ffmpeg and Read on PNGs. The cutroom calls, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide pitfalls`, `cutroom doctor`
  3. `cutroom init talk-rot.mov`, `cutroom inspect`
  4. `cutroom guide perceive`, `cutroom transcribe`, `cutroom transcript`
  5. `cutroom transcript talk-rot --suspects`, `cutroom guide edit`
  6. `cutroom edit tighten`, `cutroom timeline`, `cutroom timeline --joins`
  7. `cutroom edit cut --from "By the way" --to "keyboard." --dry-run`, and the retake cut with `--near 6.31 --dry-run`
  8. Same two cuts applied, then `cutroom timeline`, `cutroom timeline --joins`
  9. `cutroom guide render`, `cutroom guide verify`
  10. `cutroom render --preview --range 0:00-0:08 --dry-run`, `cutroom render --preview --range 0:00-0:08`, `cutroom check out/preview.mp4` (failed)
  11. `cutroom check out/preview-range.mp4`
  12. `cutroom render`, `cutroom check out/final.mp4`
  13. `cutroom look --file out/final.mp4 2`
  14. `cutroom look 2`

- **Errors hit:**
  - `error[no-render]: /Users/nick/.../out/preview.mp4 not found` / `fix: check the path, or render first`. Cause: `render --preview --range` wrote `out/preview-range.mp4`, not `out/preview.mp4`. The `fix:` line did not say that, and I had to run `check` on the new name to continue. Once I used the right file, check passed. The shell also reported "Exit code 1" for the whole chain, which was confusing because the dry-run output printed first.

- **Guessed wrong:**
  - I assumed `--preview --range` would write `out/preview.mp4`, as the render guide's wording suggests. It writes `out/preview-range.mp4`. The guide does not mention this.
  - I assumed `--range` would cut the render down to the requested span. The dry-run said it "renders the whole chunk(s) covering it", which for a 13.8 s timeline is the whole video. That is correct but it makes `--range` pointless at this length, and the guide should say so.

- **Confusing output:**
  - `cutroom look` frames of this source are a static test pattern (colour bars, a checkered marker, a rainbow bar), not a talking head. The transcript is speech, but the picture shows no speaker. I could not tell from the frames whether the picture is upright, because the pattern is nearly symmetric to my eye. I settled it with ffmpeg (see Not sure below).
  - `inspect` says "handled rotated: display metadata rotates the picture 180°". That is clear, but the next step is unclear. Nothing tells you how to confirm the render is upright, except `check`'s "display rotation 0°" line, which only proves the metadata is gone, not that the picture was turned correctly.
  - The `timeline` output shows source times, and `timeline --joins` shows output times. The two columns use different time bases on the same screen and I had to re-read them.

- **Missing:**
  - A way to confirm orientation. Something like `cutroom look --orientation`, or a check that compares the render against the source as a viewer sees it, would have saved the ffmpeg detour. For this run I had to compute the PSNR of the final frame against the auto-rotated and raw source frames to confirm the render matched the auto-rotated source (24.9 dB vs 5.6 dB).
  - An option to preview a short span that renders only the needed chunks, with a real output file named as the render guide describes.

- **Not sure it's right:**
  - Orientation: verified only by ffmpeg PSNR against the auto-rotated source frame, not by eye. The final's `check` says display rotation 0° and no rotation side data is present in the file (ffprobe showed none).
  - The cuts at 0:18.5 (cat remark) and 0:06.2 (retake) were checked by `--joins` pause lengths and by `check`'s word match (all 44 expected words heard in order). I did not run `cutroom listen` on either join. The transcript is base.en, so the expected-word list may itself contain recogniser errors.
  - The two cut edges ("the first thing you need is" retake) follow the `retake-kept` warning's suggested command, which I used as printed. I did not confirm that the retained take is the better one; I kept the last take as the guide advises.

- **Best part:** `cutroom transcript --suspects` named the retake and the stray sound by time, and `edit cut --dry-run` showed the new join before I applied it. The retake warning printed the exact command to run, and it worked as printed.

- **One change:** make `render --preview` write `out/preview.mp4` also when `--range` is given (or make `check`'s no-render error list the files that do exist). That one change would have removed the only real error I hit.
