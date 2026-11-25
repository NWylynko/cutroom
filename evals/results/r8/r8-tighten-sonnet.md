# Friction report
- **Outcome:** done. out/final.mp4 is 13.8 s. It dropped the pauses, the retake (with "um"), the cat aside and the cough. `check` passed with 0 failures and 0 warnings.
- **Commands run:** 9 (not counting file writes): guide; guide workflow + guide edit; init; transcribe; transcript + transcript --suspects; edit tighten; edit cut x2; timeline --joins; timeline; render; check (some chained).
- **Errors hit:** none.
- **Guessed wrong:** nothing. Quoting the transcript text worked first try.
- **Confusing output:** the cut message reports a source range wider than the words (0:06.200-0:09.433). That is the pad snapping, which is fine. The retake-kept warning appeared right after tighten and went away once I cut.
- **Missing:** nothing needed. I would have liked an automatic listen or look at the joins.
- **Not sure it's right:** the words check uses base.en, the same model as the transcript. It confirms the cut but not the recogniser. I did not listen to the joins or run a larger model. I also did not check whether the "first thing you need is an old laptop" take is clean.
- **Best part:** quote-based cuts, and `tighten` also removing the stray cough on its own.
- **One change:** have `timeline` or `check` print the final transcript of the render by default, so the result can be read without another command.
