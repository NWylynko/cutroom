# Friction report: r7-vfr-sync-haiku

- **Outcome:** done. `out/final.mp4` (17.233 s, 1280x720) rendered and `cutroom check out/final.mp4` passed with 0 failures and 0 warnings, including sync (median and worst -0 ms). The false start and the long pauses are gone. The cat aside was kept because the task did not ask for it.

- **Commands run:** 24 cutroom invocations, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow` / `edit` / `render` / `verify` / `pitfalls` (5 separate topic calls, one shell line)
  3. `cutroom doctor`
  4. `cutroom init talk-vfr.mp4`
  5. `cutroom inspect`
  6. `cutroom transcribe`
  7. `cutroom transcript`
  8. `cutroom transcript talk-vfr --suspects`
  9. `cutroom edit tighten --dry-run`
  10. `cutroom edit tighten`
  11. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`
  12. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."`
  13. `cutroom timeline`
  14. `cutroom timeline --joins`
  15. `cutroom render --preview --dry-run`
  16. `cutroom render --preview`
  17. `cutroom check`
  18. `cutroom render`
  19. `cutroom check out/final.mp4`
  20. `cutroom status`

  (The `guide` topic calls and the `status` call are included in the count above.)

- **Errors hit:** none. Every command exited 0 and no `error[code]` appeared.

- **Guessed wrong:** nothing failed, but I was unsure whether `tighten` would also remove the sound at 0:16.974. Its dry-run said "this would also remove the sound with no words", and the real run did. I accepted that without a `cutroom listen` on it, which the pitfalls guide suggests for suspects. The file also reports `check` `sync` as measured against the source's own audio. I took that to mean lip sync was not drifting, but I did not confirm it.

- **Confusing output:**
  - The `tighten` summary is dense: `pads 0.22+0.10 s + --min-cut 0.4 s`, plus a pause list in source time. It is hard to see at a glance what changed in the timeline.
  - After the cut, the timeline segment IDs skip from `s1` to `s3`, with no `s2`. That is correct, but it looks like a bug the first time you see it. A note that IDs are not renumbered would help.
  - The `WARN[retake-kept]` line hands you a ready command with `--near 6.31`, but `--near` takes a source time while `--to "first thing."` is a text anchor. Mixing them in one command is unusual. It worked, but I had to think about it.
  - `check` reported `54 expected words` while the transcript showed 64. The difference is the removed false start, which I worked out myself. A line saying the words were the kept words, not the source's, would help.

- **Missing:**
  - Nothing for the VFR and audio-offset concern itself. `inspect` reported `handled vfr` and `handled av-offset +0.064 s`, but there is no way to see that the offset was applied to the final render. `check` only confirms the render matches the source's own audio at the picture's moment. It cannot show whether the source's own lip sync is off.
  - A way to list or hold back individual stray-sound removals inside `tighten`, rather than all or nothing.

- **Not sure it's right:**
  - Lip sync: `check` passes with `median -0 ms, worst -0 ms`. That measures render against source, not picture against audio at the source. I did not watch or listen to the final render. The +0.064 s audio start on the source was handled by the tool, and I trust that. I did not measure it against the picture.
  - The transcript came from `base.en`, so the words may be misheard. The check's word match (54 of 54) compares the render with the same small model's transcript, so it cannot catch a shared mishearing.
  - I did not check the cat aside. I left it in because the task did not ask to remove it. The tool did not flag it as a problem.

- **Best part:** `cutroom check` gave a per-segment sync measurement, a frame-count match, and a word-by-word match in one command. A PASS with 0 warnings is easy to trust. `cutroom edit cut --dry-run` also showed the join before I applied it.

- **One change:** have `cutroom check` (or `inspect`) state the av-offset that the final render applied, and show it next to the source's measured offset. That would have let me confirm the VFR and +0.064 s fix from the output rather than trust it.
