# Friction report

- **Outcome:** done. out/final.mp4 is 13.8 s, with the retake and the cat aside removed. `cutroom check` passed with 0 failures and 0 warnings.
- **Commands run:** 8 in total: guide, guide workflow + guide edit, init, transcribe, transcript, transcript --suspects, edit tighten, edit cut (retake), edit cut (aside), timeline, timeline --joins, render, check. (Several were chained in one shell call.)
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** `tighten` printed a retake-kept warning with a ready command. I used my own --from/--to instead and it worked. The cut times were reported on a padded range (0:06.200-0:09.433), which looked odd but was fine.
- **Missing:** nothing needed.
- **Not sure it's right:** I did not run `transcript --file out/final.mp4` to read the render. The transcript came from the small base.en model, so a misheard word could in theory hide a problem. I also did not listen to the joins.
- **Best part:** the transcript suspects and the tighten warning pointed straight at the problems, and `check` verified words, sync and loudness.
- **One change:** have the cut-confirmation line show the kept words on both sides of the join, which it does partly. Otherwise nothing.
