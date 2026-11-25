# Friction report

- **Outcome:** done — out/final.mp4 uses the recorder audio (offset +1.7 s, drift -42.8 ppm), long pauses tightened, `check` PASS with sync worst -6 ms.
- **Commands run:** 7: guide; guide workflow/edit/pitfalls; init talk.mp4 --source mic=recorder.wav; sync mic; edit audio mic; transcribe; edit tighten; timeline (chained in one call); timeline --accept x2; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing. The workflow guide's recorder section was exact.
- **Confusing output:** `tighten` also removed a stray-sound gap (0:16.97-17.29) and trimmed edges; fine but I only saw it by reading carefully. Two WARNs (retake, aside) appear though the task only asked for pause removal; I accepted them with a reason, unsure if that was the intended choice.
- **Missing:** nothing needed.
- **Not sure it's right:** whether leaving the "UM" retake and the cat aside in is what the requester wanted (I read the task literally).
- **Best part:** the recorder workflow (init/sync/edit audio) and check's sync-vs-camera lines proving sync independently.
- **One change:** have `render` hint when open WARNs were accepted without cuts, or say plainly in the task-level guide that warnings need not be fixed unless requested.
