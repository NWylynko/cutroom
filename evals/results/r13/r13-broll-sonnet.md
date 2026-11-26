# Friction report
- **Outcome:** done. out/final.mp4 is 13.77 s with retake and aside cut, pauses tightened, terminal.mp4 full-screen over "open the terminal ... update command." with talk audio; `cutroom check` PASS, 0 warnings.
- **Commands run:** 8 (after reading guides): guide; guide workflow/edit/overlays/verify; init talk.mp4 terminal.mp4; transcribe; transcript; transcript --suspects; edit tighten; edit cut (retake); edit cut --text (aside); timeline; overlay add clip; overlay list; timeline --joins; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant.
- **Confusing output:** `edit cut` output of the first cut left a stale-looking aside-kept warning (expected). Retake cut kept the later "first thing you need is an old laptop" correctly.
- **Missing:** nothing; I did not view frames (`look --overlays`) so the visual was verified only by check's overlay line.
- **Not sure it's right:** base.en words check only confirms the cut, not the recogniser; I did not listen to joins or look at the cutaway frame.
- **Best part:** `overlay add clip --at/--to` by words, and the thorough `check`.
- **One change:** have `timeline` suggest overlay clips / print a look command post-render so visual verification is prompted.
