# Friction report

- **Outcome:** done. 13.8 s edit with burned-in captions at out/final.mp4 and out/final.srt; `cutroom check` PASS, 0 warnings.
- **Commands run:** ~13: guide; guide workflow/captions/edit; init; transcribe; transcript; transcript --suspects; edit tighten; edit cut (retake); edit cut (cat aside); edit cut 16.9-17.4; timeline; timeline --joins; captions --burn --output out/final.srt; render; check; ls/cat srt; captions --check; look --output 0.90.
- **Errors hit:** `error[nothing-cut]: nothing to cut in talk 16.9-17.4` (tighten had already removed that sound; the fix line explained it, harmless).
- **Guessed wrong:** cut the cough sound manually before checking whether tighten had removed it (it prints that it did).
- **Confusing output:** the first cut-by-quote worked even with "UM" in the quote; nothing confusing. Timeline segment ids skip (s1,s3,s4,s6,s7), slightly odd.
- **Missing:** nothing essential. A way to view a few burned-caption frames in one go (look takes one time, I used one).
- **Not sure it's right:** transcript is base.en, so only checked words against itself; caption wording is as the model heard it. I only viewed one frame, not the whole caption sequence.
- **Best part:** `captions --burn --output out/final.srt` made both deliverables in one step; `check` covers captions, sync, words.
- **One change:** have `tighten` retake suggestions auto-offered as one command (the retake was cut by me manually, fine, but an `--apply` for suspects would save a step).
