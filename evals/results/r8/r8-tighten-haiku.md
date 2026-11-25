# Friction report: r8-tighten-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s, 1280x720) passes `cutroom check` with 0 failures and 0 warnings; the retake, the cat aside and the long pauses are gone, and the final transcript reads as the intended explanation. Not verified by ear (see Not sure it's right).

- **Commands run:** 25 cutroom shell calls plus one image Read of `.cutroom/look/cuts_p1.png`, in this order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `guide perceive`, `guide edit`
  3. `cutroom guide render`, `guide verify`, `guide pitfalls`
  4. `cutroom doctor`
  5. `cutroom init talk.mp4`
  6. `cutroom inspect talk`
  7. `cutroom transcribe`
  8. `cutroom transcript`
  9. `cutroom transcript --suspects`
  10. `cutroom transcript --words --from 6 --to 12`
  11. `cutroom edit tighten`
  12. `cutroom timeline`
  13. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`
  14. `cutroom edit cut --from "By the way" --to "keyboard." --dry-run`
  15. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."`
  16. `cutroom edit cut --from "By the way" --to "keyboard."`
  17. `cutroom timeline`
  18. `cutroom timeline --joins`
  19. `cutroom render --preview --dry-run`
  20. `cutroom render --preview`
  21. `cutroom check out/preview.mp4`
  22. `cutroom render`
  23. `cutroom check out/final.mp4`
  24. `cutroom transcript --file out/final.mp4`
  25. `cutroom look --cuts`

- **Errors hit:** none. No command failed and no `fix:` line was needed.

- **Guessed wrong:** nothing failed, but `cutroom edit cut --from "The first thing you need"` needed `--near` to disambiguate. I had expected `--occurrence` to be the more natural fix. The suggested `--near 6.31` was correct, so the friction was small.

- **Confusing output:**
  - `timeline` reports output times, while `edit cut` and `transcript` use source times. I used both, and the help text explains the split, but it was easy to mix them up.
  - After the cuts the segment IDs jump (s1, s3, s4, s6, s7). That is correct but needs a second look.
  - `look --cuts` prints the "before" source time as 4.38 while `timeline` says 4.40. The difference is small and probably rounding, but I had to check it.
  - `timeline` lists "retake-kept" as a WARN after `tighten`. The suggested command is good, but the WARN appears before I have decided anything, and it cleared only after I applied the cut by hand. A short note on this would help.

- **Missing:**
  - A way to listen to the one sound suspect that `tighten` deleted, before `tighten` runs. I wanted to confirm it was not words, and I had to take the tool's word for it.
  - A transcript pass with a larger model for this run. `doctor` mentions one, but this run used base.en only. The words check confirms the cut, not the recogniser.

- **Not sure it's right:**
  - I did not run `cutroom listen` on any of the four joins (pause 0.30–0.33 s). I checked them only visually with `look --cuts`, which shows frames and not audio. A click or an abrupt stop at a join would not show in that image.
  - I did not run `cutroom listen` on the sound at 16.97 s. It was removed with the pause, and the tool said so, but I did not confirm it was a cough and not a mumbled word.
  - The transcript is from base.en, so a misheard word would show up as a "mismatch" in both the transcript and the check. The final transcript reads correctly to me, but this is my own reading of a small-model transcript.
  - The source is a test-pattern video with a counter. I did not judge picture quality beyond the cut frames and the format check.

- **Best part:** the flow from `transcript` → `--suspects` → `edit cut --dry-run` → apply → `timeline --joins` → `render --preview` → `check` was smooth and self-consistent. The dry runs showed the join and the new total before I committed, and the `next:` lines always pointed at the right following command.

- **One change:** have `cutroom listen` (or `timeline --joins`) show the audio around each join, not just the words, and show what a `tighten` would delete (with a `listen` command for each) before the tighten runs. That would have let me confirm the 16.97 s sound and the four joins by ear, which I could not do with these tools.
