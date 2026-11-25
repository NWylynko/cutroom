# Friction report

- **Outcome:** done — out/final.mp4 renders with the aside and the retake removed, `cutroom check` passes with 0 warnings, and the file has no rotation tag. Not independently confirmed as upright by eye (see Not sure).

- **Commands run:** 22 cutroom invocations, plus one `ls` and one image Read.
  1. `cutroom guide`
  2. `cutroom guide workflow` / `cutroom guide edit` / `cutroom guide verify`
  3. `cutroom doctor`
  4. `cutroom init talk-rot.mov`
  5. `cutroom inspect`
  6. `cutroom transcribe`
  7. `cutroom transcript`
  8. `cutroom transcript talk-rot --suspects`
  9. `cutroom edit tighten`
  10. `cutroom timeline`
  11. `cutroom edit cut --from "By the way," --to "the keyboard." --dry-run`
  12. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`
  13. `cutroom edit cut --from "By the way," --to "the keyboard."`
  14. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."`
  15. `cutroom timeline --joins`
  16. `cutroom timeline`
  17. `cutroom render --preview --range 0:00-0:08`
  18. `cutroom render`
  19. `cutroom check`
  20. `cutroom look --output 0:02`
  21. `cutroom transcript --file out/final.mp4`

- **Errors hit:** None. No command failed, so no `fix:` line was needed.

- **Guessed wrong:** Nothing failed, but I was unsure how `--near` combined with `--to` would resolve; the `--dry-run` output settled it. I also expected `render --preview --range` to take `0:00-0:08`, and it did.

- **Confusing output:**
  - `cutroom render --preview` wrote a 960x540 file while the final is 1280x720. The preview line does not say why the size differs, so it looked like a bug until I checked.
  - The `timeline` WARN lines are long (a full suggested command plus a quoted transcript line each). Useful, but they wrap badly in a terminal.
  - `cutroom look --output 0:02` printed `"show"` and the frame was a colour test pattern. It is not clear from the output what a rotated or upright frame should look like for this footage.

- **Missing:**
  - A way to confirm orientation by eye on a real frame. `look` gave me a test pattern with no orientation cue, so I could not see whether the picture is upright.
  - A second, independent transcription of the final. `check` `words` re-transcribes with base.en, the same model that made the transcript, so it confirms the cut but not the recogniser. `doctor` showed only base.en installed, and I did not find out how to get a larger model.

- **Not sure it's right:**
  - Upright orientation. `check` `format` and `picture` both say the rotation is baked in and the frames match "as a player shows them", but both use cutroom's own rotation logic. I did not confirm it visually because the look frame had no orientation cue.
  - Word-level accuracy. The `words` check passed, but both the source transcript and the check used base.en, so a misheard word would pass both.
  - The aside and retake removals. The dry-run joins read naturally ("…install Linux on it. | Finally, open the terminal…"), but I only read the text joins and did not `cutroom listen` to them.

- **Best part:** The `--dry-run` on `edit cut` printed the resulting join ("…install Linux on it. | Finally, open the terminal…", pause 0.33 s) before I committed. That made each cut a low-risk, reviewable step. `cutroom check` printing a per-check table ending in `PASS ... 0 warning(s)` made the final verdict easy to read.

- **One change:** Make `cutroom look` (or `check`) state the orientation of the rendered file in plain terms, e.g. "rendered file: 1280x720, no rotation tag; scene is upright (top of frame = top of subject)". I could not verify the upright requirement, and a short statement like that would have closed the gap.
