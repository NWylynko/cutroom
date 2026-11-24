# Friction report: r5-recorder-haiku

- **Outcome:** done. `out/final.mp4` (20.433 s, 1280x720) passes `cutroom check` with 0 failures and 0 warnings, including sync against the recorder and against the camera's own sound. It still contains the "UM" retake and the cat aside, because the brief asked only for long pauses to be removed.

- **Commands run:** 24 cutroom commands, plus a few `cd`/`ls` calls outside cutroom. In order:
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom guide edit`
  4. `cutroom guide render`
  5. `cutroom guide perceive`
  6. `cutroom guide verify`
  7. `cutroom guide pitfalls`
  8. `cutroom doctor`
  9. `cutroom init talk.mp4 --source mic=recorder.wav`
  10. `cutroom inspect`
  11. `cutroom sync mic`
  12. `cutroom transcribe`
  13. `cutroom transcript talk`
  14. `cutroom transcript talk --suspects`
  15. `cutroom edit audio mic`
  16. `cutroom edit tighten`
  17. `cutroom timeline`
  18. `cutroom timeline --joins`
  19. `cutroom render --preview --dry-run`
  20. `cutroom render --preview`
  21. `cutroom check out/preview.mp4`
  22. `cutroom render`
  23. `cutroom check out/final.mp4`
  24. `cutroom status`

- **Errors hit:** none. Every command exited 0, so I never needed a `fix:` line.

- **Guessed wrong:** nothing. The `workflow` guide's "Separate audio recorder" recipe (`init --source`, `sync`, `edit audio`) worked as written.

- **Confusing output:**
  - `sync` prints `mapping: mic time = +1.700 s + talk time × (1 -42.7 ppm)` and `offset +1.700 s`. Both are readable, but the sign convention (which clock is ahead, and which way the drift goes) is not stated, so I had to infer it. The `status` line `synced to talk (+1.700 s, -42.7 ppm)` has the same ambiguity.
  - `timeline` prints `timeline: 1 segments`. The pluralisation is wrong.
  - After `tighten`, the `retake-kept` warning repeats in `tighten`, `timeline`, `timeline --joins` and `status`. It is accurate, but it is noise when the brief only asks for pauses to be removed, and I had to check that it was not blocking anything.

- **Missing:**
  - A way to acknowledge or silence a warning I have decided to keep (for example `retake-kept` on a take I want), so the status line can show 0 warnings.
  - A quick way to see what `tighten` alone would do before it runs. `--dry-run` may exist, but I did not check it.
  - Confirmation of the drift sign convention in plain words.

- **Not sure it's right:**
  - I did not listen to the join points or the 1.7 s offset with `cutroom listen`. I relied on the `sync` fit (residual 0.02 ms) and on `check`'s sync and sync-camera results.
  - The transcript used `base.en`. The word check passed (63 of 63 words in order), but the transcript itself may have misheard words that both the source and the render share.
  - The retake ("the first thing you need is, UM, the first thing") and the cat aside ("By the way, my cat just walked across the keyboard") are still in the video. I did not cut them because the brief did not ask for it. The user should decide.
  - `cutroom check` reports `sync-camera` as 0 ms median and worst. I have no independent way to confirm that the result is not trivially zero.

- **Best part:** `cutroom guide workflow` has a "Separate audio recorder" section with three exact steps, and `check` reports the sync results per segment against two references. Together they made the sync and verification straightforward.

- **One change:** let warnings be acknowledged in the timeline (for example `cutroom timeline ack retake-kept` or a `--accept` flag). A kept retake or a deliberate sound would then stop reappearing in every command, and `status` could show a clean result without my having to explain it each time.
