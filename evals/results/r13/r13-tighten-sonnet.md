# Friction report

- **Outcome:** done. out/final.mp4 is 13.8 s (from 30.2 s) and `cutroom check` passed with 0 failures and 0 warnings.
- **Commands run:** about 10: guide; guide workflow/edit/verify; init talk.mp4; transcribe; transcript; transcript --suspects; edit tighten; edit cut (retake); edit cut (aside); timeline; timeline --joins; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** none. The retake warning's suggested cut dropped the earlier take (UM take), which was right.
- **Missing:** nothing. I could not hear the audio, only read transcripts.
- **Not sure it's right:** the "words" check uses the same small model (base.en) as the transcript, so it confirms the cut but not recogniser accuracy. I did not run `look` or `listen` on the joins.
- **Best part:** the guide, suspects list and ready-to-run cut suggestions made the task near-mechanical.
- **One change:** have `check` also summarise which source phrases were removed, as a final sanity line.
