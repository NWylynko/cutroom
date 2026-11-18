# Friction report (fill this in as the last thing you do)

- **Outcome:** done — `out/final.mp4` (1280x720, 13.767 s, 413 frames) passes `cutroom check` with 0 warnings; the retake and the cat remark are cut, pauses tightened, and the file has no rotation tag. Not fully verified: whether the picture is upright in a player, since I could not compare against a reference frame by eye (see Not sure).

- **Commands run:** 14 Bash calls (plus 3 image reads of PNGs cutroom wrote).
  1. `cutroom guide`
  2. `cutroom guide workflow && cutroom guide edit && cutroom guide render`
  3. `ls -la` and `cutroom doctor`
  4. `cutroom init talk-rot.mov` and `cutroom inspect`
  5. `cutroom guide perceive && cutroom guide verify | head -80`
  6. `cutroom transcribe` and `cutroom transcript`
  7. `cutroom guide pitfalls | head -60`
  8. `cutroom transcript talk-rot --suspects` and `cutroom edit tighten --dry-run`
  9. `cutroom listen 16.5-17.8` and `cutroom transcript --search "the first thing you need is"`
  10. `cutroom edit tighten` plus two `cutroom edit cut ... --dry-run`
  11. Two real `cutroom edit cut`, then `cutroom timeline` and `cutroom timeline --joins`
  12. `cutroom render --preview` and `cutroom check out/preview.mp4`
  13. `cutroom look --file out/preview.mp4 2` and `cutroom look --cuts`
  14. `cutroom render`, `cutroom check out/final.mp4`, and `ffprobe` for the rotate tag

- **Errors hit:** None from cutroom. The only warnings were `WARN[retake-kept]` after tighten, which was expected because the retake had not been cut yet, and it cleared after the cut. The `fix:` lines were not needed.

- **Guessed wrong:**
  - `cutroom edit tighten --dry-run` printed `created timeline from talk-rot (whole source, 30.233 s)` before the dry-run note. So a dry run seems to write `timeline.json`, which contradicts `DRY RUN: nothing was changed`. The real tighten then printed the same "created timeline" line, so I could not tell whether the dry run had already created it.
  - I assumed `--to` in `edit cut` would match the end of the same phrase. It does (the first match after `--from`), but the `--from` phrase "The first thing you need is, UM" also contains "the first thing" for `--to`. I only found out it worked by checking the dry-run output.

- **Confusing output:**
  - `cutroom transcript` tags the retake as `[repeat]` on both takes. It is clear enough once you read it, but it does not say which take to keep.
  - The `tighten` summary mixes "pads" and "edges trimmed" in one line, so I could not tell how much each change affected the cut points.
  - `cutroom look` prints the PNG path and then only "view the image". I had to open each PNG to see anything; the text alone says nothing about orientation.

- **Missing:**
  - A way to confirm orientation. `check` reports `display rotation 0°` and `picture ... worst difference 0.0`, but that compares the render to the source frames as cutroom decodes them. It does not show me whether the picture is upright. I wanted a side-by-side of the source frame and the render frame, or a reference such as a text overlay in the test footage.
  - A way to cut a whole phrase by its transcript line number, instead of quoting the words.
  - A way to show what a dry run would change to `timeline.json` without writing it.

- **Not sure it's right:**
  - Upright orientation. The source is marked as rotated 180° by metadata. I trusted the renderer's claim that the pixels are turned upright and the `picture` check passed. I did not confirm this against a reference frame, and the test pattern has no text or face to judge by.
  - Word accuracy. The transcript uses the `base.en` model, which cutroom itself warns about. The `words` check compares the render to the timeline's words, using the same small model, so a misheard word would pass both sides. I did not check words with a larger model.
  - The cat-remark cut (0:18.500–0:21.933) starts and ends in pauses. The join reads "install Linux on it. | Finally, open the terminal" with a 0.33 s pause. I did not listen to the join.

- **Best part:** `cutroom transcript` with `[repeat]` tags and `--suspects`, plus `cutroom edit cut --from/--to` by quoted words, found and removed the retake and the cat remark in two commands. `cutroom check` listing each check with its result made it easy to confirm the final file.

- **One change:** Make `edit ... --dry-run` truly read-only (no `timeline.json` written, no "created timeline" line), so that dry runs can be used safely before any edit.
