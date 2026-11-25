# Friction report: 11-hdr-phone-haiku

- **Outcome:** done. `out/final.mp4` (1280x720, 13.767 s) passes `cutroom check` with 0 warnings, and the retake, the cat aside and the long pauses are all gone.

- **Commands run:** 27 cutroom invocations, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom guide edit`
  4. `cutroom guide verify`
  5. `cutroom doctor`
  6. `cutroom init clip.mov`
  7. `cutroom inspect`
  8. `cutroom transcribe`
  9. `cutroom transcript clip`
  10. `cutroom transcript clip --suspects`
  11. `cutroom edit --help`
  12. `cutroom edit cut --from "The first thing you need is" --to "the first thing." --near 7.5 --dry-run`
  13. `cutroom edit cut --from "By the way" --to "keyboard." --dry-run`
  14. `cutroom edit cut --from "The first thing you need is" --to "the first thing." --near 7.5`
  15. `cutroom edit cut --from "By the way" --to "keyboard."`
  16. `cutroom edit tighten`
  17. `cutroom timeline`
  18. `cutroom timeline --joins`
  19. `cutroom render --preview --range 0:00-0:12`
  20. `cutroom check out/preview-range.mp4`
  21. `cutroom look --help`
  22. `cutroom look out/preview-range.mp4 5`
  23. `cutroom look --source clip 0:01`
  24. `cutroom render`
  25. `cutroom check out/final.mp4`
  26. `cutroom transcript --file out/final.mp4`
  27. `cutroom look out/final.mp4 3`
  Also used the Read tool on three PNGs that `look` wrote, and once tried a non-cutroom plan tool (below).

- **Errors hit:** none from cutroom. Every cutroom command exited 0, so no `fix:` line was needed. The only error was from a non-cutroom tool: `PreToolUse:mcp__ccd_turn__update_plan hook error: Not shown: you are a subagent ...`. It blocked a task-list call and had nothing to do with the edit.

- **Guessed wrong:** nothing badly. I guessed that `--near 7.5` would pick the first take of the repeated phrase "The first thing you need is", and it did. I expected the aside cut to remove only "By the way ... keyboard.", but the dry-run showed it also removes 0.22 s of pause on each side and the stray sound at 16.97 s. That is documented in the guide, but I did not expect the removed range to be so much wider than the quoted words.

- **Confusing output:** The `timeline` and `timeline --joins` output was clear. The WARN lines came from the timeline before I made the second cut. They listed `kept-sound` and `aside-kept`, which I then removed. A reader might think a warning was left over, so it would help if the output said which WARNs the current edit has already resolved. The `retake-kept` dry-run suggestion printed `--near 6.31` for the same retake I had already cut by a different span; I had to work out that my cut had already done it.

- **Missing:**
  - There is no way to see the output frame and the source frame side by side, or a colour comparison in the output. For HDR-to-SDR, the only evidence that the output is right is the `check` line `tone ... saturation 87 vs source 87, brightness range 217 vs 217`. A test-card frame does not let me judge "washed out" by eye.
  - `inspect` reports "a plain SDR conversion looks washed out" as "handled" but does not state the output colour path (tone-mapping from HLG to bt709) until a render runs. Saying so in `inspect` would help.

- **Not sure it's right:**
  - The SDR result is verified only by `check` and by eye on a test card. I could not visually confirm that the colours in the final file are not washed out.
  - The transcript comes from `base.en`, the small model. The words are the ones I expected, and the render's own re-transcription says all 44 words are heard in order. That check uses the same small model, so it confirms the cut but not the recogniser's accuracy.
  - The `--near` and `--from` quotes were chosen by me from the transcript. I did not listen to the joins; `timeline --joins` shows the words and pauses, and the pauses are 0.30 to 0.33 s.

- **Best part:** The sequence `transcript --suspects` → `edit cut --dry-run` → `edit cut` → `timeline` → `render --preview --range` → `check` is clear. The suspects list named each problem with its time and a suggestion, and `check` confirmed the render automatically, including the HDR tone check and the rotation.

- **One change:** Have `timeline` or `check` print the output's colour path in one line (source HLG bt2020, tone-mapped to bt709 SDR, method videotoolbox) with the source and output saturation and brightness numbers, so an agent can confirm an HDR-to-SDR result without guessing from a test card.
