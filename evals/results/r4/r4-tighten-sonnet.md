# Friction report
- **Outcome:** done. out/final.mp4 is 13.77 s and `cutroom check` PASSED with 0 failures and 0 warnings.
- **Commands run:** 10 in total: guide, guide workflow/edit/pitfalls (one call), init, transcribe, transcript, transcript --suspects, edit tighten, edit cut (retake), edit cut (aside), edit cut 16.9-17.4, timeline --joins, render, check.
- **Errors hit:** `error[nothing-cut]: nothing to cut in talk 16.9-17.4`. Its fix line was right: tighten had already removed that sound. I didn't need to do anything.
- **Guessed wrong:** I assumed the stray sound at 16.97 still needed a manual cut. Tighten had already removed it, and its output said so.
- **Confusing output:** none. I did not use the printed `retake-kept` command. I cut the retake from my own quoted words and the warning went away.
- **Missing:** nothing.
- **Not sure it's right:** I did not listen to the joins. The transcript is from the small base.en model, but check reported all 44 expected words heard.
- **Best part:** text-anchored cuts, plus a check that proves words, sync and stray sounds.
- **One change:** have `edit cut <time range>` say when the range was already removed by tighten, so the error is not a surprise.
