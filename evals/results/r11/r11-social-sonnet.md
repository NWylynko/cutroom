# Friction report

- **Outcome:** done. out/final.mp4 (burned captions, 13.8 s) and out/final.srt produced; `cutroom check` PASS with 0 failures/warnings.
- **Commands run:** ~12: guide; guide workflow/edit/captions; init talk.mp4; transcribe; transcript; transcript --suspects; edit tighten; edit cut (retake); edit cut (aside); edit cut 16.9-17.4; timeline; timeline --joins; captions --burn --output out/final.srt; render --output out/final.mp4; check; look --output 0.9.
- **Errors hit:** `error[nothing-cut]: nothing to cut in talk 16.9-17.4` - the stray sound was already removed by tighten (its output said so); fix line was accurate, harmless.
- **Guessed wrong:** nothing significant. I assumed `render --output` exists; it worked.
- **Confusing output:** timeline segment ids skip (s1,s3,s4,s6,s7). The captions --burn output says "5 cues" for the SRT but 14 burned groups - fine but worth a note that the two differ by design.
- **Missing:** nothing essential.
- **Not sure it's right:** I could not eyeball the caption look myself (check samples frames and confirms captions on screen, but I didn't view the image). Transcript is base.en; word check only confirms against the same model.
- **Best part:** `captions --burn --output out/final.srt` does both deliverables in one call; `check` covers captions and SRT.
- **One change:** have `tighten` print a note when it also removes the retake/aside suspects' sibling sounds, and have `look` output an image path inline - minor. Mostly: nothing blocking.
