# Friction report: r8-recorder-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s, 1280x720) passes all `cutroom check` rows with 0 warnings, and the sync fit residual was 0.02 ms over 54 windows.

- **Commands run:** 32 cutroom invocations, in order:
  1. `cutroom guide` (then `guide workflow`, `guide perceive`, `guide edit`, `guide render`, `guide verify`, `guide pitfalls`)
  2. `cutroom doctor`
  3. `cutroom init talk.mp4 --source mic=recorder.wav`
  4. `cutroom sync --help`
  5. `cutroom sync mic`
  6. `cutroom transcribe`
  7. `cutroom transcript`
  8. `cutroom transcript --suspects`
  9. `cutroom transcript talk`
  10. `cutroom transcript talk --suspects`
  11. `cutroom timeline` (error: no-timeline)
  12. `cutroom edit audio mic`
  13. `cutroom edit tighten`
  14. `cutroom timeline`
  15. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --source talk --dry-run`
  16. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --source talk`
  17. `cutroom edit cut --from "By the way" --to "keyboard." --dry-run`
  18. `cutroom edit cut --from "By the way" --to "keyboard."`
  19. `cutroom timeline`
  20. `cutroom timeline --joins`
  21. `cutroom render --dry-run`
  22. `cutroom render --preview`
  23. `cutroom check out/preview.mp4`
  24. `cutroom render`
  25. `cutroom check`
  26. `cutroom look --cuts --file out/final.mp4`

- **Errors hit:**
  - `error[no-timeline]: no timeline yet` from `cutroom timeline`, before any edit. Its `fix:` line ("make the first edit (e.g. `cutroom edit tighten`)") was accurate and got me unstuck, though `edit audio` would also have created it.

- **Guessed wrong:**
  - I assumed `cutroom transcript` with no source would show the camera (`talk`). It showed `mic` (the recorder), with a one-line "also transcribed: talk" footer. The two clocks differ by 1.7 s, so I had to re-run with `transcript talk` to get the camera times I needed for the cuts.
  - I assumed `--near` took the same clock as the printed phrase times. The `--near 6.31` in the suggested fix is a source time, while the WARN line that carried it said `at output 0:06.360`. I worked out from the `--near` note that it was source time.

- **Confusing output:**
  - `sync mic` printed "mic started recording 1.700 s before talk" next to "mapping: mic time = +1.700 s + talk time x (1 -42.8 ppm)". Both are correct, but the first reads as the opposite sign to the second until you work through it.
  - `edit audio mic` said "created timeline from talk (whole source, 30.233 s)" and then "timeline: 1 segment, 30.233 s", but `inspect` gives the talk source as 30.261 s. The 28 ms difference is not explained anywhere.
  - WARN lines mix clocks in one line: `at output 0:17.130` next to a suggested `cutroom listen 16.47-17.79`, which is in source time. The guide explains this, but the line itself does not say which clock each number is in.
  - `cutroom look --cuts` printed the PNG path relative (`.cutroom/look/cuts_p1.png`). I had to resolve it against the run directory myself to open it.

- **Missing:**
  - A `--source` option on `transcript` (or a header naming the default source before the phrases), so the camera's transcript is one command rather than a second call.
  - No way to get the sync result or drift summary again without re-running `sync`. `cutroom status` may cover it, but I did not check.

- **Not sure it's right:**
  - I did not listen to the recorder against the camera audio. Sync was judged only by `check`'s sync and sync-camera rows (both about 0 ms median), which compare the render's sound with the mic and camera sound. The recorder's sound quality being "better" is taken from the task statement, not verified.
  - The retake choice: I kept "The first thing you need is an old laptop." and cut "The first thing you need is, UM, the first thing." on the transcript and suspect evidence alone, without `cutroom listen`. base.en may have misheard "UM" or the words around it.
  - The cat aside was cut on the transcript alone. The `check` `words` row confirms the render matches the timeline's words, not that the cut is the right editorial choice.
  - Stray sound at source 16.97 was removed by `tighten`; I did not `listen` to it. The `stray-sound` check passes, so nothing is kept there.

- **Best part:** `cutroom sync mic` measured both offset and clock drift and reported a residual, so alignment took one command. Dry-run cuts showed the join pause and the resulting timeline length before anything was changed. The suggested `edit cut` line in the retake WARN was directly runnable, which made the retake fix quick.

- **One change:** make `cutroom transcript` (no argument) show the camera (video) source by default, or put a clearly labelled header on each source's block, so output times can be trusted without re-reading the footer.
