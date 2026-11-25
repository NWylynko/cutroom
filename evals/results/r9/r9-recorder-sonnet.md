# Friction report
- **Outcome:** done. out/final.mp4 uses mic audio, synced (offset +1.7 s, drift -42.8 ppm), long pauses tightened; check PASS 0 failures.
- **Commands run:** 7 — guide; guide workflow + guide edit; init talk.mp4 --source mic=recorder.wav && sync mic && edit audio mic && edit tighten (failed at tighten); transcribe talk && edit tighten && timeline && render && check.
- **Errors hit:** error[no-transcript] on `edit tighten` — fix line said run `cutroom transcribe`; got me unstuck immediately.
- **Guessed wrong:** nothing; the workflow guide's recorder section was exact. I did not know tighten needs a transcript first.
- **Confusing output:** none major. The retake-kept WARN (and the removed "cat walked across" aside, UM retake) is outside the task so I left it open; unclear whether the open WARN is considered unfinished.
- **Missing:** nothing needed. A hint in the recorder section to transcribe before tighten would help.
- **Not sure it's right:** retake/aside content left in on purpose (task only asked for pauses). Transcript from small base.en model.
- **Best part:** sync + edit audio + check's two independent sync measurements (vs mic and vs camera) made verification easy.
- **One change:** have `edit tighten` offer/auto-run transcribe (or list it in the recorder workflow steps).
