# Friction report: r8-fix-edit-haiku

- **Outcome:** done. Final render passes `cutroom check` (0 fail, 0 warn) and the render transcript reads as one clean explanation; I did not listen to the joins by ear.

- **Commands run:** 10 shell calls (about 25 cutroom invocations), in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`; `cutroom guide edit`
  3. `cutroom status`; `cutroom timeline`; `cutroom timeline --joins`; `cutroom transcript`
  4. `cutroom transcript talk --suspects`; `cutroom edit history`; `cutroom edit --help`
  5. `cutroom edit undo`; `cutroom timeline`; `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`
  6. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."`; `cutroom edit cut --from "By the way" --to "keyboard." --dry-run`
  7. `cutroom edit cut --from "By the way" --to "keyboard."`; `cutroom timeline`; `cutroom timeline --joins`; `cutroom transcript talk --suspects`
  8. `cutroom render --help`; `cutroom render --preview`
  9. `cutroom render`; `cutroom check out/final.mp4`
  10. `cutroom transcript --file out/final.mp4`; `cutroom timeline --joins`; `cutroom status`

- **Errors hit:** none. No command failed.

- **Guessed wrong:** nothing failed, but I had to work out that `cutroom edit undo` removes only the last edit (here `edit cut talk 1.9-2.9`) and keeps the earlier `tighten`. `edit history` shows the order, but `undo` gives no preview of what it will restore. I also assumed the `timeline` warning's `--apply retake-kept@<time>` example in the guide would match the printed time. It did not: the guide shows `@4.09`, while the warning printed at `0:03.093` before the undo and `0:04.093` after, so I used the suggested `edit cut` line directly.

- **Confusing output:**
  - `timeline` warnings are long single lines with the suggested command inline. I read them several times.
  - The `retake-kept` suggestion uses `--near 6.31` (a source time) while the timeline prints output times, so the reader has to keep two time bases apart.
  - `timeline` shows the source range for each segment but not the words of the whole output. I had to infer the final explanation by combining segment lines.
  - The render transcript writes "is an old laptop, then install Linux on it." with a comma where the source has a sentence break. That is recogniser punctuation, not a real join problem, but it is easy to misread as a run-on.

- **Missing:**
  - No way to see the whole edited script as prose before rendering. A `timeline --text` or a `--dry-run` that prints the resulting transcript would have let me check the explanation without a render.
  - No detection of off-topic asides. `transcript --suspects` found the retake and the stray sound but not "By the way, my cat just walked across the keyboard." I cut it because the guide says to cut asides, but that was my judgment from reading the transcript.
  - No preview for `edit undo`, so I could not see what it would restore before running it.
  - No check on the original `out/final.mp4` before I edited, so I could not compare its `check` result with the fixed one.

- **Not sure it's right:**
  - Cutting the cat aside was a judgment call. The task said "every sentence complete", and I read the aside as off-topic. If the author wanted it, the timeline would be 3.4 s longer.
  - The 0.30 s pauses at the joins pass `check` (its `sync` and `words` checks only cover the picture-to-sound match and the word list), but I did not `cutroom listen` any join, so I cannot say the joins sound natural.
  - `check` confirmed all 44 expected words are heard in order, but the transcript reads them from the render with `base.en`, the same model the edit used. It confirms the cut, not the recogniser.

- **Best part:** `cutroom timeline --joins` shows each cut with the words either side and the pause length, and `cutroom transcript --file out/final.mp4` read the render back. Together with `check`'s per-line PASS, I could confirm the result without watching the video. The `--dry-run` on each cut showed the join and the new total before I committed it.

- **One change:** a preview of the whole edited transcript before rendering, for example `cutroom timeline --text` or `cutroom edit ... --dry-run` printing the resulting script with segment boundaries marked. That would have let me check the explanation's sentences and asides in one read instead of combining segment lines by hand.
