# Friction report

- **Outcome:** done. out/final.mp4 uses the recorder audio, synced (offset +1.700 s, drift -42.8 ppm), pauses tightened, and `cutroom check` passed with 0 failures and 0 warnings.
- **Commands run:** 4 shell calls, 9 cutroom commands in order: guide, guide workflow, guide edit, init talk.mp4 --source mic=recorder.wav, sync mic, transcribe, edit audio mic, edit tighten, timeline, render, check.
- **Errors hit:** none.
- **Guessed wrong:** nothing. The workflow guide's recorder section gave the exact steps.
- **Confusing output:** timeline printed retake-kept and aside-kept WARNs that I left open. The task asked only for pause removal, so I did not cut the retake or the cat aside, and I did not run `--accept`. The WARNs stay "open" and I had to decide myself that they were out of scope.
- **Missing:** nothing important. A single `--audio mic` shortcut on tighten or render is not needed.
- **Not sure it's right:** the transcript used the small base.en model, so word checks confirm the cuts, not the recogniser. I did not listen to the joins. Also unsure whether the user wanted the retake and aside removed.
- **Best part:** the workflow guide's recorder section and `check` reporting sync against both mic and camera audio.
- **One change:** have `init` with a `--source` audio file print the full three-step plan (sync, edit audio, tighten) as next lines.
