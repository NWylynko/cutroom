# Friction report: r4-fix-edit-haiku

- **Outcome:** done. The corrected `out/final.mp4` (13.767 s) passes `cutroom check` with 0 failures and 0 warnings, and its transcript reads as one clean explanation.

- **Commands run:** 8 shell calls, in order:
  1. `cutroom guide` (with `ls -la` of the run folder)
  2. `cutroom guide workflow`, `cutroom guide edit`, `cutroom guide verify`, plus `cat` of cutroom.json and timeline.json
  3. `cutroom timeline`, `cutroom timeline --joins`, `cutroom transcript`, `cutroom transcript --suspects`, `cutroom check out/final.mp4`
  4. `cutroom edit history`, `cutroom transcript --words --from 0 --to 31`
  5. `cutroom edit undo`, `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`, same with `--from "By the way, my cat" --to "across the keyboard."`
  6. `cutroom edit cut` for both cuts above, then `cutroom timeline` and `cutroom timeline --joins`
  7. `cutroom render --preview`, `cutroom render`, `cutroom check out/final.mp4`
  8. `cutroom transcript --file out/final.mp4`, `cutroom transcript --suspects out/final.mp4`

- **Errors hit:**
  - `error[unknown-source]: no source "out/final.mp4"` / `fix: add it first: cutroom init out/final.mp4` from `cutroom transcript --suspects out/final.mp4`. I did not follow the fix. The `--suspects` flag works with `--file`, which the earlier `transcript --file` output suggested. The fix line is wrong for a render, since a render should not be registered as a source.

- **Guessed wrong:**
  - I assumed `cutroom transcript <file>` would read a rendered file. It printed nothing useful on the first try. `--file` is the form that works.
  - I assumed `--suspects` would accept a file path the same way. It does not; it needs `--file` first.

- **Confusing output:**
  - The initial `cutroom timeline` listed three WARN lines (`edge-in-word`, `broken-sentence`, `retake-kept`) that all came from one bad tighten cut (`edit cut talk 1.9-2.9`). The `broken-sentence` warning was the only one that named the actual problem ("going to show you how" removed). It took `edit history` and a word-level transcript to see that the tighten-era cut was the cause, and that `edit undo` was the right fix.
  - `retake-kept` suggested `--near 6.31` with `--to "first thing."`. That was correct, but I used a plain `--from "The first thing you need is, UM"` instead. The suggestion would have been clearer if it had named the take that gets removed.
  - `timeline --joins` was the most useful view. It showed each join with its pause, so the cat aside and the false start were easy to see.

- **Missing:**
  - Nothing stopped me. A `cutroom transcript --file <render> --suspects` form that errors clearly, instead of the misleading `init` fix, would have saved a round trip.
  - `cutroom timeline` does not flag an off-topic aside. I found the cat sentence only by reading the transcript. The guide says to cut asides, so a hint would help.

- **Not sure it's right:**
  - I checked audio and sync only through `cutroom check`, which passed. I did not `cutroom listen` at the joins, and I did not `cutroom look` at frames. Visual content was checked only by the picture check against the source frames.
  - I judged the "UM" take to be the false start and the cat sentence to be an aside. Both are judgments on the transcript, not on the video.
  - The pauses at the joins are about 0.30 s after the cuts. I read them as natural, but I did not listen to them.

- **Best part:** `cutroom timeline`, `timeline --joins`, and `transcript --words` together show exactly what is kept and where it joins. The `check` output lists every expected word and every WARN with its output time. Using `--dry-run` before each cut made the edits safe.

- **One change:** make the `unknown-source` fix line for a render path say "use `cutroom transcript --file <path>`" instead of "add it first: cutroom init". That single change would have avoided the only wrong turn.
