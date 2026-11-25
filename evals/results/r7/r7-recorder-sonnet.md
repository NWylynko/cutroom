# Friction report

- **Outcome:** done — out/final.mp4 made from the recorder audio, synced, pauses tightened; `check` PASS with 0 failures and 0 warnings.
- **Commands run:** 8 (`guide`, `guide workflow`, `guide edit`, `init talk.mp4 --source mic=recorder.wav`, `sync mic`, `edit audio mic`, `transcribe`, `edit tighten`, `render`, `check`, `timeline`, `transcript --suspects`).
- **Errors hit:** none.
- **Guessed wrong:** nothing; the workflow guide had the exact recipe.
- **Confusing output:** `tighten` reported a still-open `retake-kept` warning. I looked at it but did not cut it (the task asked only for long pauses removed).
- **Missing:** nothing important.
- **Not sure it's right:** the retake-kept item (a possible retake left in) was not resolved; the task did not ask for it.
- **Best part:** the "Separate audio recorder" section of `guide workflow`, plus `check` measuring sync against both mic and camera.
- **One change:** make `init` print the sync/audio next steps when a `--source` audio file is added.
