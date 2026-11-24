# Friction report

- **Outcome:** done. out/final.mp4 uses the recorder audio, `cutroom check` passes with 0 failures and 0 warnings (sync worst -6 ms, sync-camera worst 0 ms), and the 6 long pauses are shortened.
- **Commands run:** 7, in this order: `cutroom guide`; `cutroom guide --all`; `cutroom init talk.mp4 --source mic=recorder.wav`; `cutroom sync mic`; `cutroom edit audio mic`; `cutroom transcribe`; `cutroom edit tighten`; `cutroom timeline`; `cutroom render`; `cutroom check`. (The sync, edit audio, transcribe, tighten and timeline calls were chained in one shell command, so 10 invocations.)
- **Errors hit:** none.
- **Guessed wrong:** nothing. The guide's "Separate audio recorder" section gave the exact sequence.
- **Confusing output:** `tighten` left a `retake-kept` warning ("the first thing you need is", with a "UM" take). I did not cut it because the task only asked for pause removal. It is not obvious whether "long pauses removed" should also cover retakes or the cat aside.
- **Missing:** nothing needed. A one-line "recorder task: init, sync, edit audio, tighten, render, check" recipe would be a small shortcut.
- **Not sure it's right:** I did not listen to or look at the output. I relied on `check` alone. Retake and the "cat walked across the keyboard" aside remain in the video, which is a judgment call.
- **Best part:** the guide's recorder recipe, and `sync` reporting offset (+1.700 s) and drift. `check` has an independent sync-camera test.
- **One change:** have `tighten` or `timeline` say whether the remaining warnings (retake-kept) are in scope for a pause-only job.
