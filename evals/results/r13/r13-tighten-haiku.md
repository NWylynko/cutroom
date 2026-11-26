# Friction report: r13-tighten-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s, 1280x720) passes `cutroom check` with 0 failures and 0 warnings, and its re-read transcript shows only the kept lines.

- **Commands run:** 23 cutroom invocations across 9 shell calls, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide edit`, `cutroom guide verify`
  3. `cutroom doctor`, `cutroom init talk.mp4`, `cutroom inspect`, `cutroom transcribe`, `cutroom transcript`
  4. `cutroom transcript talk --suspects`, `cutroom edit tighten`, `cutroom timeline`
  5. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`, `cutroom edit cut --from "By the way," --to "the keyboard." --dry-run`
  6. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."`, `cutroom edit cut --from "By the way," --to "the keyboard."`, `cutroom timeline --joins`, `cutroom timeline`
  7. `cutroom render --preview`, `cutroom check`
  8. `cutroom render`, `cutroom check out/final.mp4`
  9. `cutroom transcript --file out/final.mp4`

  One extra non-cutroom call (a terminal read) returned nothing useful.

- **Errors hit:** none. No command failed.

- **Guessed wrong:** nothing major. I assumed `cutroom render` writes `out/final.mp4` by default and that `cutroom check` with no argument checks the final; both held. I did not need to read the source.

- **Confusing output:**
  - The `retake-kept` and `aside-kept` warnings each print a ready `edit cut` command. Following them worked on the first try, but the `--near 6.31` value is a source time while `--from`/`--to` are words. That is only clear if you read the guide.
  - `timeline` lists every warning twice: once after the tighten output and again after the timeline table. A short line saying which is current would help.
  - `transcript` printed `[repeat]` on the good take as well as the false start. It is only a hint, but the label reads like an instruction to cut.

- **Missing:**
  - A way to listen to a join without a separate `cutroom listen` call. I skipped `listen` and relied on the join text, the pause lengths and the re-transcript. A `--listen` flag on `timeline --joins` would make this more reliable.
  - A way to see the render's own words in the final transcript with timings compared against the timeline. The output was readable, but I had to check the timings by hand.

- **Not sure it's right:**
  - The re-transcript uses `base.en`, which the check says only confirms the cut, not the recognition. The transcript itself may contain misheard words. I read the final transcript and it matched the intended script, but I did not listen to the audio.
  - The joins at 0:03.93, 0:06.40, 0:08.13 and 0:11.50 have pauses of 0.30 to 0.33 s. I judged them natural from the text only and did not run `cutroom listen` on any of them.
  - The stray sound at source 0:16.97 was removed by `tighten`. I did not verify that it was a non-speech bump rather than a word the recogniser missed. The check's stray-sound test passes, but that only confirms no sound remains.

- **Best part:** the suspects list, the retake warning with its ready fix, and `--dry-run` on each cut. The whole edit took two dry runs and two applies, and the check output was clear enough to read in one pass.

- **One change:** `timeline --joins` (or `cutroom check`) should report the actual listened level and transition at each join, so a join can be judged without a separate `listen` run for every cut.
