# Friction report: run 11-fix-edit-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s) passes `cutroom check` with 0 failures and 0 warnings, and the render transcript reads as complete sentences with no retake, aside, or pause over 0.32 s.
- **Commands run:** about 26 `cutroom` invocations over 8 shell calls, plus `ls`, `cat` and an update_plan call that was blocked for subagents. In order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide edit`, `cutroom guide verify`
  3. `cutroom status`, `cutroom timeline`, `cutroom transcript`, `cutroom transcript --suspects`, `cutroom check` (on the broken edit)
  4. `cutroom edit history`, `cutroom transcript --words --from 0:00 --to 0:05`, `cutroom edit --help`
  5. `cutroom edit undo` (reverted `edit cut talk 1.9-2.9`), `cutroom timeline`, `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`, `cutroom edit cut --from "By the way," --to "the keyboard." --dry-run`
  6. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."`, `cutroom edit cut --from "By the way," --to "the keyboard."`, `cutroom timeline`, `cutroom timeline --joins`
  7. `cutroom render --preview`, `cutroom render`, `cutroom check out/final.mp4`
  8. `cutroom transcript --file out/final.mp4`, `cutroom transcript --suspects`, `cutroom status`
- **Errors hit:** none from cutroom. The only error was from the harness, not cutroom: the update_plan hook refused the call because I am a subagent.
- **Guessed wrong:** I first assumed the inherited edit was a word cut. `cutroom edit history` and the timeline `reason` field showed it was a range cut, `cut talk 1.9-2.9`, which removed "going to show you how" from inside the sentence. Nothing in `cutroom edit history` says the range was a sentence-level mistake. I also assumed `status` "0 error(s), 3 warning(s)" meant the edit was sound. It was not: the `broken-sentence` warning was a WARN, and the sentence was broken.
- **Confusing output:**
  - The timeline warnings printed `broken-sentence` with the fix "if unintended: `cutroom edit undo`, or cut whole sentences with --from/--to". That is two different options in one line, and the undo option is the only one that gets you back to a clean state.
  - Output times shift after each edit. `retake-kept` was at 0:03.093 before the undo and 0:04.093 after it. I had to re-read the timeline each time to get the right time for `--apply` or `--accept`.
  - `transcript --suspects` still lists the retake and aside after they are cut, marked "(resolved)" or "(already cut)". That is accurate but long to re-read.
- **Missing:**
  - A way to see the output's spoken text as one paragraph before rendering. `cutroom timeline` shows segment starts and "Hi, today I'm" style fragments, so I had to mentally join them. A `cutroom timeline --text` or `transcript --timeline` would have shown the result as a reader hears it.
  - A way to tell an accidental range cut from a deliberate one. The reason string is "cut after: range 1.9-2.9" and nothing marks it as a mistake.
- **Not sure it's right:**
  - I did not `cutroom listen` any join. The join pauses are 0.30 to 0.33 s, and the word check passes, but I did not hear the joins.
  - I cut the cat aside ("By the way, my cat just walked across the keyboard.") because the warning suggested it. Whether it is a digression you want gone is a judgement call. I treated it as off-topic.
  - The check says "all 44 expected words heard in order", but base.en is also the model that built the transcript, so it only confirms the cut, not the transcription.
- **Best part:** `cutroom timeline` with the joins view and the warnings' exact suggested commands. The `--dry-run` on each cut showed the resulting join and the new total length before I applied anything, so I did not have to guess at the result.
- **One change:** when `cutroom timeline` reports a `broken-sentence` WARN, print the sentence as it now reads (for example, `Hi, today I'm | to set up a home server.`) and suggest the specific `--from/--to` cut that keeps the whole sentence, not just "undo or cut whole sentences".
