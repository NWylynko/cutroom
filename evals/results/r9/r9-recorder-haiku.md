# Friction report (r9-recorder-haiku)

- **Outcome:** done — `out/final.mp4` passes `cutroom check` (0 failures, 0 warnings): recorder audio aligned (+1.700 s, -42.8 ppm), all six long pauses cut to 0.32 s, 30.3 s became 20.4 s. The retake was deliberately kept (the brief asked only for pause removal), so the viewer still hears "The first thing you need is, UM, the first thing." followed by the full take.

- **Commands run:** 28 cutroom commands (plus one `ls` and one image `Read` of the look PNG), in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom guide edit`
  4. `cutroom guide perceive`
  5. `cutroom guide render`
  6. `cutroom guide verify`
  7. `cutroom guide pitfalls`
  8. `cutroom doctor`
  9. `cutroom init talk.mp4 --source mic=recorder.wav`
  10. `cutroom sync mic`
  11. `cutroom sync mic --help`
  12. `cutroom transcribe`
  13. `cutroom transcript`
  14. `cutroom transcript talk`
  15. `cutroom transcript talk --suspects`
  16. `cutroom transcript mic --suspects`
  17. `cutroom edit audio mic`
  18. `cutroom edit tighten`
  19. `cutroom timeline`
  20. `cutroom timeline --joins`
  21. `cutroom timeline --accept retake-kept@4.09 --reason "..."`
  22. `cutroom render --preview --dry-run`
  23. `cutroom render --preview`
  24. `cutroom check` (on out/preview.mp4)
  25. `cutroom render`
  26. `cutroom check`
  27. `cutroom look --file out/final.mp4 --output 0:05`
  28. `cutroom transcript --file out/final.mp4`

- **Errors hit:** None. No command exited non-zero.

- **Guessed wrong:**
  - I assumed `cutroom transcript` with no source would show the camera's transcript. It showed the **mic** transcript (`mic — 64 words, 0:31.960`) and only noted `also transcribed: talk` on one line. Times were in the recorder's clock, 1.7 s off from the camera, so I had to re-run `transcript talk` to get the times I wanted to cut on.
  - I expected `retake-kept`'s suggested fix to use the warning's own output time. The warning line says `at output 0:06.360` but its suggested command uses `--near 6.31`, which is a **source** time. Only the guide says `--near` is a source time.
  - The render guide says `--preview` renders 540p. The output said `960x540`, which is 540 lines high, so this is only a wording mismatch, not a bug.

- **Confusing output:**
  - `timeline` printed the retake warning twice in different units: once before tighten at `output 0:06.360`, then after tighten at `output 0:04.093`. Both are correct, but I had to work out they were the same warning.
  - The transcript says 64 words, but `cutroom check` reports `all 63 expected words`. I could not tell from the output which word was dropped or why the counts differ (probably "UM", but nothing says so).
  - `transcript --suspects` for talk printed `(1 breath(s) not listed: quiet, normal; --all shows them)`. That was useful, but a slightly longer blurb would make it clearer why it is hidden.

- **Missing:**
  - A way to remove only pauses (`tighten`) and leave retakes alone without first accepting a `retake-kept` warning. I wanted to say "pauses only" and did not find a clean way to state that. `--accept` with a reason did the job.
  - A `transcript` default that matches the edit's clock. With no argument, `transcript` showed the recorder (`mic`) first, so its times were 1.7 s off from the camera times that `edit cut --from/--to` and `timeline` use. Either the default should be the camera source, or the output should say up front which source's clock the times are in.
  - A one-line summary that the tightened timeline has no open warnings other than the one I accepted.

- **Not sure it's right:**
  - I did not listen to the final audio. The `sync` and `sync-camera` checks say the recorder sound lines up with the camera's own sound to within 6 ms worst case, which I trust, but I did not verify by ear that the recorder's audio is the one in the file.
  - The retake was deliberately left in. If the user wanted it removed, the `--apply` suggestion from the warning would do that.
  - I did not check for missed words beyond the 63-vs-64 count.
  - `cutroom look` produced a test-pattern frame (`"you"` label) — this is the fixture's picture. I did not judge the image further.

- **Best part:** The separate-recorder workflow in `cutroom guide workflow` was exact: `init --source mic=…`, `sync mic`, `edit audio mic`, then transcribe and cut with the camera. The `sync` output (offset, drift, residual) and the `check` `sync-camera` line gave me real confidence the alignment held for the whole video.

- **One change:** Make `cutroom transcript` (and `timeline`) print times in the timeline's own clock and default to the source the timeline's sound comes from — or at least add one line at the top saying which source's clock every time is in. That single confusion cost me the most re-reading.
