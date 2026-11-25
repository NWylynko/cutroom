# Friction report

- **Outcome:** done. out/final.mp4 (13.767 s, 1280x720) renders and `cutroom check` passes all 8 checks; the one real problem left by the previous editor (the "The first thing you need is, UM, the first thing." false start) is cut, and the aside and pauses were already handled.

- **Commands run:** 18 cutroom commands plus 4 shell commands (ls, cat, find, head) for looking at the folder.
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom status`
  4. `cutroom timeline`
  5. `cutroom timeline --joins`
  6. `cutroom transcript`
  7. `cutroom transcript talk --suspects`
  8. `cutroom edit history`
  9. `cutroom edit --help`
  10. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`
  11. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --reason "..."`
  12. `cutroom timeline`
  13. `cutroom timeline --joins`
  14. `cutroom render --preview`
  15. `cutroom check out/preview.mp4`
  16. `cutroom render`
  17. `cutroom check out/final.mp4`
  18. `cutroom transcript --file out/final.mp4`

- **Errors hit:** none. Every command exited 0.

- **Guessed wrong:** nothing was wrong. The one thing I had to think about was that `--near 6.31` takes a source time, while the `WARN[retake-kept]` line gave the problem at output time 0:04.093. The warning text put the exact command in front of me, so I copied it.

- **Confusing output:**
  - `cutroom status` lists `next: cutroom timeline --apply retake-kept@4.09` and also offers a runnable `cutroom edit cut ...` in the warning text. I could not tell whether they do the same thing, so I used the explicit cut. A reader might not know which to pick.
  - `--apply retake-kept@4.09` uses an output time, but `--near` uses a source time. Both appear in the same warning line with no label, so mixing them up is easy.
  - `cutroom transcript talk` takes a source id, but `cutroom transcript --file out/final.mp4` takes a path. Both work, but the difference is not explained anywhere I looked.
  - The `--suspects` output marks some lines "(already cut)" and says "1 breath(s) not listed". Useful, but I had to read it twice to see that the aside and the sound at 0:16.97 were already handled by the previous editor's cut.
  - Everything else was short and clear. The `check` output (one line per check, `ok`/`PASS`) was the easiest part to read.

- **Missing:**
  - A way to listen to the joins. `cutroom listen` exists, but I did not use it. I only have the transcript and the `check` results, so I cannot say whether the 0.30 s pauses at the joins sound natural.
  - A "what the previous editor was doing" view. `edit history` shows two operations (`edit tighten --min-cut 0.4` and an aside cut) but not which segments were under review or what they meant to do next. I inferred the rest from the timeline and the transcript.

- **Not sure it's right:**
  - I did not listen to the output. The check confirms the words are heard in order and the loudness passes, but it does not judge whether the joins sound like a natural edit.
  - I judged "Then install Linux on it." complete from the transcript and treated it as fine. The transcript shows no false start there, but I have not heard it.
  - I treated the aside, the sound at 0:16.97, and the tighten pauses as already correct, based on the previous editor's history and `--suspects`. I did not re-check them independently.

- **Best part:** `cutroom status`, `cutroom timeline --joins`, and `cutroom check` together were enough to find the problem, cut it with a dry-run first, and verify the render without reading any source code. The `--dry-run` preview showed the removed text and the new join before I changed anything.

- **One change:** make every warning's `next:` line a single runnable command that names the time unit it uses (for example `--near-source 6.31`), so the output time and the source time can never be confused.
