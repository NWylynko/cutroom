# Friction report: r3-vfr-sync-haiku

- **Outcome:** done, with one caveat. `out/final.mp4` rendered and passed all 8 checks (sync median and worst 0 ms), but the sync check compares render audio to the source's own audio, so it cannot show that the source's +0.064 s audio offset was handled correctly against the picture (see Not sure it's right).

- **Commands run:** 22 cutroom invocations in 21 shell calls, plus one `ls`. In order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide edit`, `cutroom guide verify`, `cutroom guide pitfalls`
  3. `cutroom doctor`
  4. `cutroom init talk-vfr.mp4`
  5. `cutroom inspect`
  6. `cutroom transcribe`
  7. `cutroom transcript`
  8. `cutroom transcript talk-vfr --suspects`
  9. `cutroom edit tighten`
  10. `cutroom timeline`
  11. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`
  12. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing."`
  13. `cutroom timeline`
  14. `cutroom timeline --joins`
  15. `cutroom guide render`
  16. `cutroom render --preview --range 0:00-0:12`
  17. `cutroom check out/preview-range.mp4`
  18. `cutroom render`
  19. `cutroom check`

- **Errors hit:** none. No command failed and no `fix:` line was needed.

- **Guessed wrong:** I expected `render --preview --range 0:00-0:12` to render only the first 12 seconds. It rendered the full 17.233 s timeline ("range render: starts at output 0:00.000"). It was one 45 s chunk, so the range seems to be applied at chunk granularity, not as a trim. The output did not say that. I also had to work out that `--to` matches the first occurrence after `--from`, and that the `--from` text had to be unique, before the cut worked.

- **Confusing output:**
  - `cutroom transcript --suspects` printed the repeat at `0:06.314–0:11.504`, while the cut I needed started at `0:06.230`. The two times are for different things, and I had to reconcile them by hand.
  - The render line `range render: starts at output 0:00.000` followed by a full-length duration read as a contradiction.
  - The sync check line `render audio vs talk-vfr at the same moment` is ambiguous. It reads as "vs picture" to me, but it compares to the source audio.

- **Missing:**
  - A way to see the source's audio-to-picture offset after the render, so I could confirm the +0.064 s was corrected or preserved. `inspect` reports it only for the source. Nothing reports it for `out/final.mp4`.
  - A way to make `--range` trim output. Or, if it can't, a warning when the range covers more than it renders.
  - Nothing in `timeline` or `check` says whether the edited cut point between "is," and "the first thing" lands on a sound that a viewer would notice as a jump. I had to infer that from the join pauses (0.31 s).

- **Not sure it's right:**
  - **Lip sync (the main risk).** `inspect` reported `audio starts +0.064 s vs video` and marked it `handled av-offset`. The render's sync check passed with 0 ms against `talk-vfr` audio. That check compares render audio to the source audio, so a 64 ms error that exists in the source would pass too. I did not listen to the output or measure the audio against the picture. I cannot confirm from the tool's output that lip sync is right.
  - I kept the "By the way, my cat just walked across the keyboard" aside, because the request named only long pauses and the false start. If that aside should go, it is one more cut.
  - Word transcription uses `base.en`. The guide says to expect misheard words. The `words` check passed (54 of 54), but that check compares against the same transcript model.
  - I did not run `cutroom listen` on the join points or on the 0.33 s sound at `0:16.974`. The checks reported no stray sound, but I did not verify by ear.

- **Best part:** The flow was linear and the checks were easy to read. `cutroom timeline` and `cutroom timeline --joins` showed the edit with the words on each side of each cut. `cutroom check` printed one clear `ok` or `FAIL` line per check, so I could tell at a glance what passed.

- **One change:** Add an audio-to-picture offset line to `cutroom check` for the final render, measured against the picture, and report it next to the source's `av-offset`. This would close the lip-sync gap that the current sync check leaves open.
