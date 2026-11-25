# Friction report
- **Outcome:** done. out/final.mp4 is 13.77 s with the retake, the aside, the cough and the long pauses removed. terminal.mp4 covers 9.0-11.3 s with the talk's voice, and `check` passed with 0 warnings.
- **Commands run:** 14: guide; guide workflow/edit/verify (one call); doctor; init; transcribe; transcript; transcript --suspects; guide overlays; edit tighten; edit cut (retake); edit cut (aside); timeline; timeline --joins; overlay add clip; look --overlays; render --preview; check; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant.
- **Confusing output:** none. The tighten step printed retake-kept warnings, and cutting by quoted words cleared them.
- **Missing:** nothing. I did not look at the image from `look --overlays` (I only ran it).
- **Not sure it's right:** the words check uses the same small base.en model, so it confirms the cut but not the recogniser. I did not view the full-screen frames myself, though `check` confirmed the overlay frames match the clip.
- **Best part:** quoted-word cuts, the overlay anchored to words, and `check` verifying the overlay window and sync.
- **One change:** have `tighten` offer to apply the retake and aside cuts in one command.
