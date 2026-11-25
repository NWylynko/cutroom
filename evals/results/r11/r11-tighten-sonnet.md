# Friction report

- **Outcome:** done. out/final.mp4 is 13.77 s with the retake, the "um" take and the cat aside cut; `cutroom check` passed with 0 warnings.
- **Commands run:** 9 in total: guide, guide workflow/edit/verify, init, transcribe, transcript, transcript --suspects, edit tighten, edit cut (retake), edit cut (aside), timeline, timeline --joins, render, check.
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** none. The first-take cut's text ended at "the first thing." and the tool cut cleanly.
- **Missing:** nothing.
- **Not sure it's right:** the check used base.en, a small model, so a mishearing could hide a clipped word. I did not run `cutroom listen` on the joins.
- **Best part:** the transcript suspects (repeat/aside/sound) and the quoted-word cuts made this a few commands.
- **One change:** have `tighten` suggest the retake and aside cuts in one step.
