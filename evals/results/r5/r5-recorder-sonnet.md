# Friction report

- **Outcome:** done. out/final.mp4 renders with the recorder audio, and `cutroom check` passes with 0 failures and 0 warnings. I did not cut the retake.
- **Commands run:** 9 (`init`, `sync`, `edit audio`, `transcribe`, `transcript --suspects`, `edit tighten`, `timeline`, `render -o` (failed), `render`, `check`).
  In order: guide; guide workflow; guide pitfalls; init talk.mp4 --source mic=recorder.wav; sync mic; edit audio mic; transcribe; transcript talk --suspects; edit tighten; timeline; render -o out/final.mp4 (failed); render; check.
  The count is 13 including the guide calls and the failed render.
- **Errors hit:** `error[bad-flag]: render has no flag -o` with `fix: did you mean --output?`. The fix was clear, but I didn't need it because plain `cutroom render` writes out/final.mp4.
- **Guessed wrong:** I assumed `-o` existed. It doesn't (long flags only).
- **Confusing output:** `tighten` also prints a retake WARN. The task only asked to remove pauses, so I left the retake ("The first thing you need is, UM, the first thing.") in. It's unclear whether a hands-off request should also fix the retake.
- **Missing:** nothing important. A one-line "done, now run check" would be nice, but the `next:` lines already do that.
- **Not sure it's right:** I never listened to the audio. Sync is only verified by `check`, which measured sync against both the mic and the camera sound. It showed a worst case of 6 ms. I did not verify that the pause removal sounds natural.
- **Best part:** the "Separate audio recorder" section in `guide workflow` gave the exact 3-command sequence. `check` reports sync twice, against the aligned mic and against the camera, independently.
- **One change:** have `init --source` suggest `sync` as the next step. Also make the `-o` error mention that a plain `render` already writes to out/final.mp4.
