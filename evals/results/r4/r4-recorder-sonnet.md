# Friction report

- **Outcome:** done. out/final.mp4 made with recorder audio (synced, +1.700 s, -42.7 ppm drift), pauses tightened, `cutroom check` PASS with 0 failures and 0 warnings.
- **Commands run:** 8. `guide`, `guide workflow; guide pitfalls`, `init talk.mp4 --source mic=recorder.wav && sync mic && edit audio mic && transcribe`, `transcript talk; guide edit; edit tighten`, `render --help; render --output out/final.mp4` (failed), `render && check`. Plus a final FRICTION write.
- **Errors hit:** `error[unknown-flag]: render has no flag --output` (fix line was right: positional path, default out/final.mp4). It unstuck me at once.
- **Guessed wrong:** that render takes `--output`. The guide's render topic wasn't read first.
- **Confusing output:** the tighten output had a WARN[retake-kept] for "the first thing you need is" (a real retake) and I left it, since the task asked only for pauses removed. I wasn't sure whether to cut the retake and the cat aside. The task didn't say, so the retake and aside remain in the video.
- **Missing:** nothing needed. A hint in the task-level guide that "render [path]" is positional would help.
- **Not sure it's right:** I did not listen to the joins or look at frames beyond what check reported. Retake and cat aside left in on purpose.
- **Best part:** the workflow guide has an exact recipe for separate recorders (init --source, sync, edit audio), and check's sync line verifies the recorder alignment.
- **One change:** list `render [output.mp4]` as positional in `guide workflow`.
