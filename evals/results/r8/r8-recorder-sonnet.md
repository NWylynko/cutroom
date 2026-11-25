# Friction report

- **Outcome:** done — out/final.mp4 renders with mic audio synced (offset +1.700 s, drift -42.8 ppm), 6 long pauses shortened; `check` PASS with 0 failures/warnings.
- **Commands run:** 8 — `cutroom guide`; `guide workflow` + `guide pitfalls`; `init talk.mp4 --source mic=recorder.wav`; `sync mic`; `edit audio mic`; `transcribe`; `edit tighten`; `timeline --joins`; `render`; `check`. (Init/sync/audio were chained in one shell call.)
- **Errors hit:** none.
- **Guessed wrong:** nothing; the workflow guide's "Separate audio recorder" section gave the exact sequence.
- **Confusing output:** `tighten` reports a retake-kept WARN ("the first thing you need is" said 2x) with a suggested cut. I left it since the task only asked to remove pauses; unclear if the task intended it. The warning stayed "open" and check didn't mention it.
- **Missing:** nothing essential. A single "recorder" recipe command would save 3 steps.
- **Not sure it's right:** I can't hear the result; sync was verified numerically (worst -6 ms) only. The transcription used base.en (small model); `words` check only confirms against that model. The retake and the "UM" were left in.
- **Best part:** the workflow guide's recorder section, and `check` measuring sync against both the mic alignment and the camera audio independently.
- **One change:** have `tighten` (or `render`) say explicitly whether leaving open retake warnings is acceptable, or offer a `--fix-retakes` option.
