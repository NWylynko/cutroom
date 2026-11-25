# Friction report

- **Outcome:** done. out/final.mp4 is 13.77 s (from 30.26 s); retake, "UM" take, cat aside and the pauses are gone; both cards are on and `cutroom check` PASSed with 0 warnings.
- **Commands run:** about 12. guide; guide workflow/edit/overlays; init; transcribe; transcript; transcript --suspects; edit tighten; edit cut (x2); overlay add (x2); timeline; timeline --joins; overlay list; timeline --accept; look (x2); render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant. I used `--to "home server."` on the overlay (it worked).
- **Confusing output:** the o2 `overlay-short` warning. The label is anchored to "Then install Linux on it." (1.43 s), and the tool wants 2.1 s, which would run into the next sentence. I accepted it with a reason. The tool gives no hint on how to choose between extending and accepting.
- **Missing:** an overlay time-extension option that stays within the speech, such as an `--extend-into-pause` flag. A way to view a contact sheet of the overlay frames in one call would also help.
- **Not sure it's right:** text was checked only by looking at two frames; the check confirms an overlay is visible but not that the text is legible or correct. The base.en transcript may have misheard words (the check uses the same model).
- **Best part:** quoting transcript text for cuts and overlay anchors, plus `timeline` and `check` explaining everything.
- **One change:** have the overlay-short warning say whether extending would collide with the next speech.
