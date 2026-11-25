# Friction report
- **Outcome:** done. final.mp4 is 13.77 s, and `check` passed with 0 failures and 0 warnings.
- **Commands run:** 11, in this order: guide, guide workflow/edit/overlays (one shell call), init, transcribe, transcript --suspects, transcript, edit tighten, edit cut (retake), edit cut (cat aside), overlay add clip, timeline --joins, render, check.
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** none. Quoting "um" in the first cut matched the transcript.
- **Missing:** nothing. I did not run `look` on the clip frame, because `check` already compared the overlay against the clip's own frames.
- **Not sure it's right:** the words check used base.en, so it confirms the cut but not the recogniser. I did not run `cutroom listen` on the joins.
- **Best part:** quoted-word cuts, the `--suspects` list and the clear `check` output.
- **One change:** have `tighten` suggest the aside cut too, or flag off-topic phrases.
