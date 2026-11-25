# Friction report
- **Outcome:** done. Inherited edit cut "going to show you how" mid-sentence, kept the "um" false start/retake, and kept the cat aside and the cough. Reset and rebuilt; `check` PASS.
- **Commands run:** ~8: ls; guide; guide workflow/edit; edit reset; transcript; transcript --suspects; edit tighten; edit cut x3; timeline --joins; render; check.
- **Errors hit:** `error[nothing-cut]` on `edit cut 16.97-17.29` (tighten had already removed it); fix line explained it, harmless.
- **Guessed wrong:** nothing significant.
- **Confusing output:** none; cat aside is a judgment call (I cut it as off-topic) which the tool can't decide.
- **Missing:** nothing; a one-shot "diagnose inherited edit" command would save reading timeline.json by hand.
- **Not sure it's right:** didn't read final transcript with `transcript --file`; check said all 44 words heard.
- **Best part:** `edit reset` + quoted-text cuts + check.
- **One change:** have `timeline` flag cuts that land mid-sentence in an inherited edit.
