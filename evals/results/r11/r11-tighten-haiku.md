# Friction report: 11-tighten-haiku

- **Outcome:** done. out/final.mp4 (13.767 s, from 30.261 s) passed `cutroom check` with 0 fail and 0 warn, and its re-read transcript matches the intended edit (the false start, the repeated first take and the cat aside are gone).

- **Commands run:** 8 shell calls, 23 cutroom invocations, in order:
  1. `cutroom guide` (then `guide workflow`, `guide edit`, `guide verify` in the next shell call)
  2. `cutroom doctor`
  3. `cutroom init talk.mp4`
  4. `cutroom inspect`
  5. `cutroom transcribe`
  6. `cutroom transcript talk`
  7. `cutroom transcript talk --suspects`
  8. `cutroom edit tighten`
  9. `cutroom timeline`
  10. `cutroom edit cut --dry-run --from "The first thing you need is, UM" --to "the first thing."`
  11. `cutroom edit cut --dry-run --from "By the way" --to "keyboard."`
  12. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing."`
  13. `cutroom edit cut --from "By the way," --to "the keyboard."`
  14. `cutroom timeline`, then `cutroom timeline --joins`
  15. `cutroom render --preview --range 0:00-0:14`
  16. `cutroom render`
  17. `cutroom check`
  18. `cutroom transcript --file out/final.mp4`
  19. `cutroom status`
  (The count of 23 includes the four guide topics and the paired timeline calls. I also made one unrelated non-cutroom session call early on, which I did not need.)

- **Errors hit:** none. Every command exited 0.

- **Guessed wrong:** nothing material. The `--dry-run` on `edit cut` previewed the join correctly. I did not use the `--apply retake-kept@4.09` shortcut or the `--near 6.31` form the warning suggested; I cut by quoted words instead, and that worked.

- **Confusing output:**
  - `cutroom timeline` repeats each open WARN with a long suggested command. After the cuts were applied, the earlier `retake-kept` and `aside-kept` lines were gone, but I had to re-read the full output to confirm that.
  - `cutroom render --preview --range 0:00-0:14` wrote `out/preview-range.mp4`, and the `--range` value is not reflected in the timeline's output-time terms. That leftover file is in the run directory, and I had no obvious command to remove it.
  - `cutroom status` lists the stale preview as "not checked" with no suggestion to delete it. That is harmless but slightly noisy.

- **Missing:**
  - A way to delete or clean up a stale preview render from the project (`cutroom render --clean` or similar, or `status` offering to remove it).
  - A quick way to listen to a sound that `tighten` removes. The timeline said `this also removes the sound with no words at source 0:16.970–0:17.300`, and I could not confirm it was a cough and not a word the recogniser missed without running `cutroom listen` on it. I skipped that check, so it is one of the unverified points below.

- **Not sure it's right:**
  - I cannot hear the audio. The `check` words test uses the same base.en model that made the transcript, so it confirms the cut against itself, not against a stronger recogniser. The guide says a larger model via `cutroom transcribe --model` checks both, and I did not run it.
  - I did not `cutroom listen` the removed sound at 0:16.970, nor the joins (the pauses at 0:03.9, 0:06.4, 0:08.1 and 0:11.5 are all about 0.30 s, which reads naturally in the transcript but is unheard).
  - "UM" was cut as part of the false start; I did not check the audio to confirm the cut point is clean.
  - The removed cat aside is off-topic by the request, and I judged it so from the transcript text alone.

- **Best part:** `cutroom guide workflow`, `edit` and `verify` gave the full sequence up front. `cutroom transcript --suspects` flagged the retake, the sound and the aside together, so I could decide each one from a single list. The `--dry-run` on `edit cut` showed the join and the new length before I applied the cut, and `check` plus `transcript --file out/final.mp4` closed the loop with a clear PASS.

- **One change:** have `cutroom check` (or `status`) state explicitly that its word test uses the same recogniser as the transcript, and print the exact `transcribe --model` command for a stronger check, so the agent knows how much the PASS proves.
