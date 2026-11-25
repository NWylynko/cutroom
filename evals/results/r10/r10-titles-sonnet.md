# Friction report

- **Outcome:** done. out/final.mp4 rendered (13.77 s); `cutroom check` PASS, 0 failures/warnings; I viewed both overlay frames.
- **Commands run:** ~14: guide; guide workflow/edit/overlays; init; transcribe; transcript; transcript --suspects; edit tighten; edit cut (retake); edit cut (aside); edit cut 16.9-17.4; timeline; overlay add x2; overlay list; look x2; overlay set o2 --duration 2.1; render; check.
- **Errors hit:** `error[nothing-cut]: nothing to cut in talk 16.9-17.4` (tighten had already removed that sound; fix line explained it, fine; my cut was redundant).
- **Guessed wrong:** nothing significant. Used `--to` for card end instead of --duration.
- **Confusing output:** overlay-short warning on o2 appeared as soon as I used `--to "on it."` (1.43 s); I extended to 2.1 s, so the card slightly outlasts the phrase into the next segment.
- **Missing:** nothing; a one-line note that the card then overlaps the next sentence would be nice.
- **Not sure it's right:** word check uses base.en only; I did not listen to joins or verify the 2.1 s card runs into "Finally" speech (acceptable).
- **Best part:** text anchors for cuts/overlays, `look` drawing overlays pre-render, and a thorough `check`.
- **One change:** have `overlay add` auto-extend (or offer) duration when anchored span is under the readable minimum.
