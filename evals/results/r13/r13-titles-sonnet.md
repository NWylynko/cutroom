# Friction report

- **Outcome:** done. out/final.mp4 (13.77 s) passes `cutroom check` with 0 failures and 0 warnings. Retake, aside and pauses are cut; the two cards are anchored to the right words and I looked at them in the overlay image.
- **Commands run:** about 17 in order: guide; guide workflow/edit/overlays/verify; init; transcribe; transcript; transcript --suspects; edit tighten; edit cut (retake); edit cut (aside); edit cut 16.9-17.4; timeline; overlay add (title); overlay add (lower-third); timeline; look --overlays; render; check; overlay set o2 --duration 2.1; timeline; look --overlays; render; check.
- **Errors hit:** `error[nothing-cut]: nothing to cut in talk 16.9-17.4`. Its fix line explained it (already cut). Harmless: tighten had already removed the sound.
- **Guessed wrong:** I tried cutting the cough by time before noticing tighten reports it removed it ("this also removes the sound with no words").
- **Confusing output:** I used `--at ... --to` for the cards, and the first render also gave an overlay-short warning for o2 because its window was only the spoken words. I had to fix it afterwards. The cut command's `--from "The first thing you need is, UM"` worked.
- **Missing:** nothing important. An automatic "extend short overlay to the minimum" option on `overlay add` would save a step.
- **Not sure it's right:** the transcript is from base.en (small model); the check re-transcribes with the same model, so it confirms the cut, not the wording. I did not hear the audio. Cards were only checked as stills.
- **Best part:** text-anchored cuts and overlays, `timeline` warnings with ready-made fix commands, and `check` verifying each overlay's frame range.
- **One change:** make `overlay add` warn or auto-pad up front when the `--to` window is shorter than the readable minimum.
