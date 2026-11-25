# Friction report: r7-recorder-haiku

- **Outcome:** done. `out/final.mp4` (1280x720, 20.433 s) uses the recorder's audio aligned to the camera (offset +1.700 s, drift -42.8 ppm), long pauses cut to 0.32 s, and `cutroom check out/final.mp4` passed every check (frames, picture, sync, sync-camera, words, stray-sound, loudness). Not verified by listening or looking at frames myself.

- **Commands run:** 20 cutroom invocations plus 2 shell commands (`pwd`, `ls`). In order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide edit`, `cutroom guide pitfalls`
  3. `cutroom --help`
  4. `cutroom doctor`
  5. `cutroom init talk.mp4 --source mic=recorder.wav`
  6. `cutroom inspect`
  7. `cutroom sync mic`
  8. `cutroom edit audio mic`
  9. `cutroom transcribe`
  10. `cutroom transcript`
  11. `cutroom edit tighten`
  12. `cutroom timeline`
  13. `cutroom timeline --joins`
  14. `cutroom render --preview`
  15. `cutroom check --help`
  16. `cutroom check out/preview.mp4 --all`
  17. `cutroom render`
  18. `cutroom check`

- **Errors hit:** none. No command failed and no `fix:` line was needed.

- **Guessed wrong:** Nothing I assumed was wrong. I did guess that `cutroom render` writes `out/final.mp4`; `cutroom check --help` confirmed it ("Default file: out/final.mp4"), but the render output also named it. I did not know until I ran it whether `sync` would need `--to` or a drift flag. It did not.

- **Confusing output:**
  - `edit audio mic` said `created timeline from talk (whole source, 30.233 s)`, while `inspect` said talk is 30.261 s. The 28 ms difference is never explained. Probably audio length versus container length, but I had to guess.
  - `sync mic` says `mic started recording 1.700 s before talk` and `mapping: mic time = +1.700 s + talk time × (1 -42.8 ppm)`. That is clear, but "started before" plus a "+" mapping needed a second read to confirm the sign.
  - After `tighten`, the `retake-kept` WARN repeats on every `timeline` and `timeline --joins` call with a long suggested `edit cut`. It is useful, but it is noise when the task was only "remove long pauses". I left the retake in, since it was not requested, and I am not sure that was the right call.
  - `transcript` prints `talk` times while the audio comes from `mic`. The note "times are each source's own" is correct but easy to miss.

- **Missing:**
  - A one-line summary of the sync result after `edit audio`, showing the offset is applied to the timeline's own output. I had to infer it from `timeline` ("sound from mic").
  - A way to say "remove pauses only, no retake or aside detection" on `tighten`, so the warnings stay out of a pause-only task.
  - In `check`, a clear line stating which source the `sync` check compares against. The `sync` line says "mic's sound" and `sync-camera` says "talk's own camera sound", which made it easy to confirm but I had to work it out.

- **Not sure it's right:**
  - `check` passing is strong evidence for sync, because `sync-camera` measures against the camera's own audio and gave median 0 ms and worst -0 ms. `sync` itself relies on the offset `sync` measured, so it is partly self-referential. I think `sync-camera` covers that.
  - The `words` check compares against a transcript from the `base.en` model, which the tool itself warns can mishear words. It passed, so the words the model hears are in order, but that is weak evidence that the speech is right.
  - I did not run `cutroom listen` or `cutroom look` on any join, so I have not heard or seen the result.
  - The final file keeps the "um" retake ("The first thing you need is, UM, the first thing.") and the cat aside. These were outside "remove long pauses", but the user may expect them gone.

- **Best part:** `cutroom sync mic` followed by `cutroom edit audio mic` was the whole recorder workflow in two commands, and the guide's "Separate audio recorder" section matched it exactly. `check` gave a per-measure PASS/FAIL, with the sync and sync-camera lines directly comparable, so I did not have to read the raw output.

- **One change:** Make `cutroom edit tighten` print a single summary line that says the audio source in use and the mapped offset (for example "sound from mic, offset +1.700 s, drift -42.8 ppm"), and have the `retake-kept` warning appear only once per run. That would have made the report and the final state much clearer with less re-reading.
