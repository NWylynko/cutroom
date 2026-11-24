# Friction report: r6-vfr-sync-haiku

- **Outcome:** done. `out/final.mp4` (17.233 s, 1280x720) rendered and `cutroom check` passed every line (sync median and worst -0 ms, words 54/54 in order, no stray sound). Caveat: I could not watch or hear the result myself, so the lip-sync verdict rests on the tool's check.

- **Commands run:** 20 cutroom invocations over 9 shell calls, in order:
  1. `cutroom guide`
  2. `cutroom guide --all`
  3. `cutroom doctor`
  4. `cutroom init talk-vfr.mp4`
  5. `cutroom inspect`
  6. `cutroom transcribe`
  7. `cutroom transcript`
  8. `cutroom transcript --suspects`
  9. `cutroom edit tighten`
  10. `cutroom timeline`
  11. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`
  12. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."`
  13. `cutroom timeline --joins`
  14. `cutroom listen 16.9-17.4`
  15. `cutroom render --preview --dry-run`
  16. `cutroom render --preview`
  17. `cutroom check out/preview.mp4`
  18. `cutroom render`
  19. `cutroom check out/final.mp4`
  20. `cutroom look --cuts` (then opened the PNG with the Read tool)

- **Errors hit:** none. Every command exited 0.

- **Guessed wrong:** nothing material. I expected `inspect` to say whether the av-offset was applied to the render. It only says `handled av-offset: audio starts +0.064 s`, so I had to infer the handling from `check`'s sync line ("its audio start offset included").

- **Confusing output:**
  - `check` prints `median -0 ms, worst -0 ms`. The negative zero reads as a bug, and I had to think about whether it meant "perfect" or a sign problem.
  - `guide --all` is about 10k tokens. I read it all, but most of it was not needed for this job. Topic-level guides would be easier to scan.
  - The `retake-kept` warning gave a ready `edit cut` line with `--near 6.31` as a source time while `--from`/`--to` are words. It worked, but the mix is easy to misread.

- **Missing:**
  - No way to tell `check` or `inspect` to show the per-segment sync measurements or the av-offset correction it applied. I wanted to confirm the VFR and +0.064 s audio offset were corrected, not just not reported.
  - No command to show what the tighten pass removed as a list of source ranges before applying it. I could only read the "pauses" lines after the fact.
  - No sync check against the original camera's picture. `check` compares the render's sound against the source's own sound, so if the source itself was already out of lip sync, it would pass. That is a real gap for a request about lip sync.

- **Not sure it's right:**
  - Lip sync: `check` passed on the preview and on the final, but it only proves the render matches the source's sound and picture. It does not prove the original recording was in sync. I could not watch the video.
  - Judgement call: I kept the "By the way, my cat just walked across the keyboard." aside, because the request only asked for long pauses and the false start to go. The user may want that removed too.
  - Transcript: base.en transcribed 64 words and check counted 54 expected words in the render. The gap is about the cut words and was not individually verified.
  - The cut-line PNG showed clean cuts, but I judged them from frames, not from listening to each join.

- **Best part:** `tighten` followed by the `retake-kept` warning with a ready `edit cut` command, plus `--dry-run` on the cut. The whole flow from transcript to verified final took very few steps, and `check` gave a clear per-line verdict.

- **One change:** `cutroom check` should report the av-offset and VFR handling it applied, with a per-segment table of measured lip-sync values, not only the median and worst. That would turn "PASS" into evidence the user can rely on for the lip-sync question.
