# Friction report

- **Outcome:** done. out/final.mp4 is 13.8 s, `cutroom check` PASS with 0 failures and 0 warnings. Both cards were looked at in the overlays image.
- **Commands run:** 13. guide; guide workflow/edit/overlays/verify (one call); init; transcribe, transcript, transcript --suspects (one call); edit tighten; edit cut (retake); edit cut (aside); overlay add (title); overlay add (lower-third); timeline; timeline --joins; overlay set o2 --duration 2.1; look --overlays; render; check. Several were chained in a single shell call.
- **Errors hit:** none.
- **Guessed wrong:** nothing material. I chose `--to` anchors for the overlays; the lower-third came out 1.43 s and warned `overlay-short`, which `overlay set --duration 2.1` fixed.
- **Confusing output:** `tighten` printed the retake-kept and aside-kept WARNs before I had cut, which is fine. Segment ids skip numbers (s1, s3, s4, s6, s7) after cuts, which looks odd.
- **Missing:** nothing. I would have liked a one-line summary of what the final video says, in order.
- **Not sure it's right:** the transcript is from the small base.en model, so the check only confirms the cut against that model. I did not listen to the audio. The title card is on screen 0.1 s to 3.7 s, which covers the first sentence.
- **Best part:** quoting words for cuts and anchors, and `look --overlays` showing the cards as rendered.
- **One change:** have `overlay add` default the duration to the reading time, so the overlay-short warning does not appear when `--to` is used.
