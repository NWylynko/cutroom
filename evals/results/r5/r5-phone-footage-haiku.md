# Friction report: r5-phone-footage-haiku

- **Outcome:** done. `out/final.mp4` (1280x720, 13.767 s) passes `cutroom check` with 0 failures and 0 warnings. The 180° display rotation is baked into the pixels, and ffprobe shows no rotation side data. The cat remark and the first take of the retake are removed, and the pauses are tightened to 0.32 s.

- **Commands run:** 31 `cutroom` commands, plus `ls -la`, one `ffprobe` and one `Read` of the output frame.
  1. `cutroom guide`
  2. `cutroom guide workflow`, `perceive`, `edit`, `render`, `verify`, `pitfalls` (six calls)
  3. `cutroom doctor`
  4. `cutroom init talk-rot.mov`
  5. `cutroom inspect`
  6. `cutroom transcribe`
  7. `cutroom transcript`
  8. `cutroom transcript --suspects`
  9. `cutroom listen 16.5-17.8`
  10. `cutroom edit tighten --dry-run`
  11. `cutroom edit tighten`
  12. `cutroom timeline`, `cutroom timeline --joins`
  13. `cutroom edit cut --from "the first thing you need is" --near 6.31 --to "first thing." --dry-run`
  14. `cutroom edit cut --from "By the way" --to "keyboard" --dry-run`
  15. the same two cuts without `--dry-run`
  16. `cutroom timeline`, `cutroom timeline --joins`
  17. `cutroom render --preview --dry-run`
  18. `cutroom render --preview`, `cutroom check out/preview.mp4`
  19. `cutroom render`, `cutroom check`
  20. `cutroom look out/final.mp4 2 --json`, `cutroom look out/final.mp4 2`

- **Errors hit:** none. No command failed, and no `fix:` line was needed.

- **Guessed wrong:** nothing failed, but I guessed at two things. I expected `cutroom guide <topic>` to print the whole topic and it did. I expected `edit cut --from` to accept a source-time anchor with `--near`, and it did, but I only learned that from the retake warning's suggested command.

- **Confusing output:**
  - The `WARN[retake-kept]` line printed after every `timeline` and `timeline --joins` call, even after the retake was already removed. This was wrong and noisy: the timeline I just read showed one "first thing" line, so the warning looked stale. I had to re-read it to confirm the timeline was correct.
  - The `tighten` dry-run prints the pause list, the pad arithmetic and the sound-removal note on separate lines. It was readable but dense.
  - `cutroom listen` printed `transcript: (nothing)` for a region with a 0.32 s sound. That is accurate, but I could not tell from the output whether that meant "no words" or "the recogniser skipped it".
  - `cutroom look out/final.mp4 2` gave a 1280x720 test-pattern frame with no obvious orientation cue. I could not confirm uprightness visually from it.

- **Missing:**
  - A way to silence or dedupe the `retake-kept` warning once the repeated take is already gone from the timeline.
  - A `look` comparison of source vs output at the same content (for example, `look talk-rot 2` next to `look out/final.mp4 2`) so I could judge orientation by eye, not only by `check`.
  - A short note in `timeline --joins` on which join is the one a retake warning refers to.

- **Not sure it's right:**
  - The retake choice: I kept the later take ("The first thing you need is an old laptop.") because the guide says the last take is usually the keeper. I did not listen to either take in isolation, so I did not verify that the kept one is the better delivery.
  - Word-level accuracy: the transcript is from `base.en`, so some wording may be misheard. `check` says all 44 expected words are heard in order, but that compares against the same recogniser's transcript, so it is not an independent check of what was actually said.
  - Uprightness: `check` reports `display rotation 0°` and a baked-in rotation, and `picture` matches the source frames. I did not view a source frame next to an output frame to confirm the orientation by eye.
  - I did not run `cutroom listen` on the two joins (`s1|s3` and `s6|s7`). The `--joins` output looked clean, and the pauses are 0.30 to 0.33 s.

- **Best part:** The workflow is clear. `tighten`, then `--suspects`, then cuts quoted from the transcript, then `timeline --joins`, then preview, then final, then `check`. `--dry-run` on every edit and render made each step cheap to verify. `check` gave a per-line PASS/FAIL that matched what I expected, so I could trust the result without re-reading all the output.

- **One change:** Make the `retake-kept` warning go away when its suggested cut is already applied, or when the repeated take is no longer in the timeline. That warning appeared on every later `timeline` call and made me re-check something I had already fixed.
