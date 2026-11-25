# Friction report

- **Outcome:** done. out/final.mp4 (13.8 s, burned captions) and out/final.srt delivered; `cutroom check` PASS with 0 failures and 0 warnings.
- **Commands run:** 14 — guide; guide workflow/captions/edit (one call); init; transcribe; transcript; transcript --suspects; edit tighten; edit cut (retake); edit cut 16.9-17.4; edit cut (aside); timeline; timeline --joins; captions --burn --output out/final.srt; render --output out/final.mp4; check; look --output 0.9.
- **Errors hit:** `error[nothing-cut]: nothing to cut in talk 16.9-17.4` / fix: range not in the timeline (already cut?). The fix was right: tighten had already removed the stray sound, so I did not need that cut.
- **Guessed wrong:** I tried to cut the stray sound by hand before checking tighten's output. It had already said "this also removes the sound".
- **Confusing output:** none significant. Timeline segment ids skip (s1, s3, s4, s6, s7) after cuts, which looks odd at first.
- **Missing:** nothing needed. A one-shot "tighten plus apply the suggested retake and aside cuts" would save steps.
- **Not sure it's right:** the base.en transcript could have misheard words, though `check` re-transcribes with the same model. I viewed only one caption frame, and it was a test-pattern fixture. The SRT is sentence-level while the burned captions are 1-3 words, which is expected.
- **Best part:** the `check` output, with a per-line verdict covering the captions, the subtitle file, sync and words. Transcript tags such as [repeat] and [aside] made the edit decisions obvious.
- **One change:** make `captions --burn` also say which render command to run next. It does print `next: cutroom render --preview`, but a final-render hint would be better.
