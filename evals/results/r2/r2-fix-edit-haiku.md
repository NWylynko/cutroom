# Friction report (fill this in as the last thing you do)

- **Outcome:** done. out/final.mp4 is re-rendered and `cutroom check` passes with 0 warnings; the three edit problems I found (a broken cut that dropped words, a false start/retake, an off-topic aside) are fixed, and the sound at 0:16.97 was already cut.

- **Commands run:** 12 shell calls (about 25 cutroom invocations inside them), plus one image read. In order:
  1. `cutroom guide`; `ls -la`
  2. `cutroom guide workflow|edit|verify|render|perceive|pitfalls`
  3. `cat cutroom.json`; `cat timeline.json`; `find out .cutroom`
  4. `cutroom transcript`; `cutroom transcript --suspects`; `cutroom edit history`; `cutroom timeline`; `cutroom timeline --joins`
  5. `cutroom check out/final.mp4`
  6. `cutroom transcript --words --from 0 --to 5`; `--from 5.5 --to 16`; `--from 17 --to 30`; `cutroom edit history --help`
  7. `cutroom edit undo`; `cutroom timeline`; `cutroom edit cut --from "The first thing you need is" --to "the first thing." --near 6.5 --dry-run`
  8. `cutroom edit cut --from "The first thing you need is" --to "the first thing." --near 6.5`; `cutroom edit cut --from "By the way" --to "keyboard." --dry-run`
  9. `cutroom edit cut --from "By the way" --to "keyboard."`; `cutroom timeline`; `cutroom timeline --joins`; `cutroom transcript --suspects`
  10. `cutroom render --preview --dry-run`; `cutroom render --preview`; `cutroom check out/preview.mp4`
  11. `cutroom look --cuts --file out/preview.mp4`; `cutroom transcript --file out/preview.mp4` (failed)
  12. `cutroom render`; `cutroom check out/final.mp4`

- **Errors hit:**
  - `error[unknown-flag]: transcript has no flag --file` / `fix: run cutroom transcript --help to list flags`. The fix line was accurate but I did not run it; I just dropped the command. Its suggestion was the right route, though `transcript` has no file option at all, so I could not read the rendered speech directly.
  - The hook error on my plan tool (not cutroom) is noise for this task and is not part of the report.

- **Guessed wrong:**
  - I guessed `transcript --file` by analogy with `look --file`. It does not exist.
  - `cutroom edit history --help` printed the whole `edit` usage block, not anything about `history`. I expected history-specific help.
  - I assumed the earlier `edit cut 1.9-2.9` was a deliberate retake removal. It was not: it removed "going to show you how" (words at 1.96–2.92 s), and nothing in the transcript flagged those words as a retake. The tool's suspects did not warn me that the cut hit real speech.

- **Confusing output:**
  - `transcript --suspects` kept listing the repeat at 0:06.310–0:11.500 and the 0:16.970 sound after I had cut both. The sound line says "(already cut)" but the repeat line does not, so it looks like a live problem. I had to cross-check against `timeline` to tell.
  - `timeline` warnings were in two places: `timeline` prints them, and `timeline --joins` prints the same two again. The `retake-kept` warning was also printed after the earlier edit with the message "cut all but one take", which pointed at the wrong target: the kept take was fine, the problem was the false start before it.
  - `cutroom edit cut` reports the removed range and the new join, but not which words are left at each side in the full sentence; I had to reconstruct the sentence from `timeline` each time.

- **Missing:**
  - A way to read the rendered speech. `check`'s `words` line says "all 44 expected words heard in order", which proves the wording matches the timeline, not that the final text reads as complete sentences. I wanted something like `cutroom transcript --file out/final.mp4` or `cutroom transcribe out/final.mp4`.
  - A way to see what an edit removed and what it kept as a sentence (before/after text of the whole timeline). Reconstructing from the words listing is slow.
  - A way to listen to the joins in one step. I looked at the cut frames with `look --cuts` but did not run `listen` on the join audio.

- **Not sure it's right:**
  - Dropping "By the way, my cat just walked across the keyboard." is my judgement that it is an off-topic aside. The task did not say to drop asides, so this could be a choice the user did not want.
  - I did not listen to the audio at the four joins (`cutroom listen` not run), only checked the frames, the 0.30–0.33 s pauses and `check`'s sync/loudness. The joins sound right in the transcript, but that is not the same as hearing them.
  - The final text should read: "Hi, today I'm going to show you how to set up a home server. The first thing you need is an old laptop. Then install Linux on it. Finally, open the terminal and run the update command. And that's it, your server is ready." That is what the timeline says, but I did not confirm it by re-transcribing out/final.mp4.

- **Best part:** `cutroom timeline --joins` plus `cutroom check` made it easy to see each cut as a pair of words with its pause, and `edit undo` plus `--dry-run` made every change safe to try.

- **One change:** make `cutroom transcript` (or `transcribe`) accept an output file, so the rendered speech can be read and compared to the timeline text the way the sources can be.
