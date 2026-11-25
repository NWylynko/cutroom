# Friction report

- **Outcome:** done. out/final.mp4 is 13.77 s, upright, no rotation metadata, and `cutroom check` passed with 0 failures and 0 warnings.
- **Commands run:** 11 (including `ls`). In order: `cutroom guide`; `guide workflow`; `guide pitfalls`; `guide edit`; `guide render`; `init talk-rot.mov`; `inspect`; `transcribe`; `transcript`; `edit tighten`; `edit cut --from "By the way" --to "keyboard"`; `timeline`; `timeline --apply retake-kept@4.09`; `render`; `check`. (Several were chained.)
- **Errors hit:** none.
- **Guessed wrong:** nothing. The retake ("The first thing you need is, UM...") was not in the task, but the tool flagged it and I removed it.
- **Confusing output:** `edit tighten` says it "also removes the sound with no words at 0:16.970" (the long-pause sound), which was easy to miss in a long line. The transcript's `[repeat]` marks are terse.
- **Missing:** nothing important. I would have liked `render` to print the rotation fix in the same place it prints the `check` result.
- **Not sure it's right:** I never looked at a frame to confirm "upright" myself. I relied on `check`'s "15 sampled frames match as a player shows them". I did not listen to the joins. The transcript is base.en, so "cat" wording was trusted.
- **Best part:** `timeline --apply` for the retake warning, and `check` explicitly verifying rotation 0° and upright frames.
- **One change:** have `render` run `check` automatically, or print "run check" more prominently, since the rules require it anyway.
