# Friction report

- **Outcome:** done. out/final.mp4 renders from recorder audio (mic synced +1.700 s, -42.7 ppm drift), pauses tightened (30.2 s -> 20.4 s); `check` PASS with 0 warnings.
- **Commands run:** 9 (plus guide topics): guide; guide workflow; guide pitfalls; init talk.mp4 --source mic=recorder.wav; sync mic; edit audio mic; transcribe; transcript talk (+ --suspects); edit tighten; render --output out/final.mp4 (failed); render --help; render; check.
- **Errors hit:** `error[unknown-flag]: render has no flag --output` (fix: run render --help; worked, got unstuck). Also check on a missing file: `no-render`.
- **Guessed wrong:** assumed `render --output`; the output path is a positional arg or defaults to out/final.mp4 (the task's path happened to be the default).
- **Confusing output:** `edit tighten` leaves WARN[retake-kept] for the "first thing you need is" repeat; I did not cut it since the task only asked for pause removal. Unclear whether to resolve it.
- **Missing:** nothing important.
- **Not sure it's right:** I did not listen to the final or look at frames; I trust check. Retake and the cat aside remain in the video by design.
- **Best part:** the guide's "Separate audio recorder" section gave the exact 3 commands; sync reported drift and check verified sync per segment.
- **One change:** make `render` accept `--output` (or have the unknown-flag fix mention the positional path).
