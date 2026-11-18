# Friction report: r3-fix-edit-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s) passes `cutroom check` with 0 failures and 0 warnings, its own transcript reads "Hi, today I'm going to show you how to set up a home server. The first thing you need is an old laptop, then install Linux on it. Finally, open the terminal and run the update command. And that's it, your server is ready." with no suspects, and the join pauses are 0.30 to 0.33 s. I did not listen to the joins.

- **Commands run:** 28 cutroom commands, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom guide perceive`
  4. `cutroom guide edit`
  5. `cutroom transcript`
  6. `cutroom transcript --suspects`
  7. `cutroom timeline`
  8. `cutroom edit history`
  9. `cutroom edit undo`
  10. `cutroom edit undo`
  11. `cutroom edit history`
  12. `cutroom edit tighten --dry-run`
  13. `cutroom edit tighten`
  14. `cutroom timeline`
  15. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing" --dry-run`
  16. `cutroom edit cut --from "By the way, my cat" --to "across the keyboard" --dry-run`
  17. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing"`
  18. `cutroom edit cut --from "By the way, my cat" --to "across the keyboard"`
  19. `cutroom timeline`
  20. `cutroom timeline --joins`
  21. `cutroom guide render`
  22. `cutroom render --preview`
  23. `cutroom render`
  24. `cutroom check out/final.mp4`
  25. `cutroom guide verify`
  26. `cutroom transcript --file out/final.mp4`
  27. `cutroom transcript --file out/final.mp4 --suspects`
  28. `cutroom look --cuts`
  I also read the existing `out/final.mp4.cutroom.json` and the history files with cat, and viewed the `.cutroom/look/cuts_p1.png` image.

- **Errors hit:** none. No command failed.

- **Guessed wrong:**
  - The original editor ran `edit tighten --min-cut 0.4` (visible in `edit history`). I ran plain `edit tighten` without that flag and did not check the default `--min-cut` or whether it differs. The result looks right, but I did not confirm it matches what the flag does.
  - I assumed `edit undo` twice would return to the untouched source. It did, but the first undo deleted the timeline entirely ("no timeline now"), so the state was not a simple rewind of the tighten step.

- **Confusing output:**
  - `edit cut` prints the cut in source time (`cut talk 0:06.200–0:09.433`), while `timeline` prints output time in the first column and source range after the segment id. Two time bases in adjacent commands made me double-check which was which.
  - The first `timeline` after tightening printed `WARN[retake-kept] at 0:04.093` (output time), while the earlier `transcript --suspects` gave source time 0:06.310. I had to reason about the mapping to see they were the same retake.
  - `timeline.json` keeps segment ids out of order (`s1`, `s8`, `s2`...) and every segment's reason is "whole source" or "cut before/after". Nothing in the output says the pauses were removed by tighten; I inferred it from the 0.30 s joins.
  - `cutroom guide` ends with "next: cutroom doctor", which I did not run. Doctor was not needed here, but the next-line is unconditional.

- **Missing:**
  - tighten output does not report the pause lengths it removed or the min-cut it used. A line like "pauses over 0.40 s cut to 0.32 s, min-cut 0.40" would have saved the guess above.
  - A way to see the original transcript-to-output mapping for a single source range, without hand-converting times.
  - An explicit "this was the edit that made the render you have" line in `check` or `timeline` output, so I could confirm that `out/final.mp4` was from the broken timeline before I changed it.

- **Not sure it's right:**
  - I did not run `cutroom listen` on any join, so I have not heard the cuts. Check passes the sync and stray-sound checks and the transcript is clean, but clicks or a room-tone change at the joins were not listened to.
  - I did not check whether the tighten default removes the same pauses as `--min-cut 0.4` did in the original editor.
  - The render transcript was done with base.en. Word-level differences would not show a mis-heard word, but the output transcript matches the script I expected.

- **Best part:** `cutroom timeline` and `cutroom timeline --joins` pointed straight at the three real problems (`broken-sentence` from the bad 1.9–2.9 cut, `retake-kept`, `edge-in-word`), and the dry-run for each cut showed the join and the removed length before I applied it. `cutroom check` plus `cutroom transcript --file out/final.mp4` gave a clear pass result on the actual file.

- **One change:** have `edit tighten` print each pause it shortened with its old and new length (for example `pause 2.13 s -> 0.32 s at 0:04.4`), and print the min-cut value it used. That would have let me verify the tighten step directly instead of inferring it from join times, and would have made the undo-and-redo sequence easy to check.
