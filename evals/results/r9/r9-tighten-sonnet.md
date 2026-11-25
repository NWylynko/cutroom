# Friction report

- **Outcome:** done. Tightened pauses, cut the "UM" retake and the cat aside, rendered out/final.mp4 (13.767 s), and `check` passed with 0 failures and 0 warnings.
- **Commands run:** 9: guide; guide workflow/edit/perceive (1 shell call); init; transcribe; transcript; transcript --suspects; edit tighten; edit cut x2 (--dry-run first); timeline; edit cut x2; timeline --joins; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** none. The retake-kept warning printed a ready command, which was useful.
- **Missing:** nothing needed.
- **Not sure it's right:** the transcript comes from base.en, so a misheard word is possible. check re-transcribes with the same model. I did not open frames or listen to joins.
- **Best part:** `tighten` followed by the retake warning's suggested cut, and `check` listing every verification line.
- **One change:** have `check` also say that the "first thing" partial-sentence retake was removed as a whole span (a summary of what was dropped).
