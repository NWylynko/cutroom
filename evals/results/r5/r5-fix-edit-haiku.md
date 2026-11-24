# Friction report: r5-fix-edit-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s) passes `cutroom check` with 0 failures and 0 warnings, all 44 words heard in order, and the transcript of the render has no suspects. I did not watch or listen to the result myself, so the judgment rests on the tool's checks and transcript.

- **Commands run:** 25 `cutroom` invocations, plus a few shell `ls`/`cat` calls.
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom guide edit`
  4. `cutroom status`
  5. `cutroom edit history`
  6. `cutroom transcript`
  7. `cutroom timeline`
  8. `cutroom timeline --joins`
  9. `cutroom transcript talk --suspects`
  10. `cutroom transcript talk --help`
  11. `cutroom edit --help`
  12. `cutroom edit reset`
  13. `cutroom edit tighten`
  14. `cutroom edit cut --from "By the way" --to "keyboard." --dry-run`
  15. `cutroom edit cut --from "The first thing you need is" --near 6.31 --to "first thing." --dry-run`
  16. `cutroom edit cut --from "The first thing you need is" --near 6.31 --to "first thing."`
  17. `cutroom edit cut --from "By the way" --to "keyboard."`
  18. `cutroom timeline`
  19. `cutroom timeline --joins`
  20. `cutroom transcript --suspects`
  21. `cutroom render --help`
  22. `cutroom render`
  23. `cutroom check out/final.mp4`
  24. `cutroom transcript --file out/final.mp4`
  25. `cutroom transcript --file out/final.mp4 --suspects`

- **Errors hit:** none. No command failed. The one `fix:` line I acted on, the `retake-kept` suggestion `cutroom edit cut --from "the first thing you need is" --near 6.31 --to "first thing."`, worked verbatim (I used the same command in the source's original casing).

- **Guessed wrong:**
  - I ran `cutroom transcript talk --help` to look up flags. It printed the flag list, but the usage line shows `[source-id]` as optional and I had passed it anyway. No real trap, just a wasted call.
  - I did not read `cutroom guide render` or `cutroom guide verify` before rendering and checking. I relied on the `next:` lines and `render --help`. This was a gap on my side, not a tool error.
  - I did not run `cutroom doctor`, which `guide` suggested as a first step. Nothing I used needed it.

- **Confusing output:**
  - The original edit had a `cut after: range 1.9-2.9` reason, a raw time-range cut that removed "going to show you how" from inside a sentence. `cutroom status` only said `3 warning(s)` and `next: cutroom check out/final.mp4`. The specific problems were only visible in `cutroom timeline`.
  - Times switch between output time and source time within the same output. The `timeline` warnings use output times (`at 0:01.433`), while the suggested fix uses source times (`--near 6.31`). `tighten` prints pauses in source time. I had to keep track of which was which.
  - `cutroom timeline --joins` repeats the same three WARN lines that `cutroom timeline` printed. Not harmful, but it doubles the output to re-read.

- **Missing:**
  - No way to see the longest pause or join gap in the render itself. `timeline --joins` covers the timeline, not the rendered file, so I relied on the joins at 0.30–0.33 s being fine.
  - No frame or visual check that I used. I ran `check`, which reports sampled frames match, but did not run `cutroom look`. I skipped the visual look and listen steps.
  - No single command that says "undo the last N edits" or "restore to a given history step". I used `edit reset` and then re-ran `tighten` and both cuts. That worked, but it was more work than an undo to step 0.

- **Not sure it's right:**
  - Whether the join `…up a home server. | The first thing…` and the cut from `talk 0:18.500–0:21.933` (the cat aside) sound natural. The dry-run and `--joins` show 0.30–0.33 s pauses, but I did not listen to them.
  - The `tighten` default `--min-cut 0.4` removed the stray sound at source 16.970–17.290 as a side effect. The tool reported this, and the final `check` reports `stray-sound ok`, but I did not hear the result.
  - The render was made at final quality with the default loudness target. `check` reports -16.0 LUFS and -4.5 dBTP, which passes.

- **Best part:** the `--dry-run` on each cut plus `timeline --joins` let me check each change before applying it. The retake and aside fix lines were specific enough to copy and run.

- **One change:** `cutroom status` should say `next: cutroom timeline` (fix warnings first) instead of `next: cutroom check out/final.mp4` when the timeline has open warnings and the render is older than the timeline. As written, the next-step line pointed at checking a render built from a broken timeline.
