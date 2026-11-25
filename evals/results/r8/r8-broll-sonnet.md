# Friction report

- **Outcome:** done. out/final.mp4 (13.767 s) rendered; `cutroom check` PASS with 0 failures and 0 warnings.
- **Commands run:** 11, in order: guide; guide workflow; guide edit; guide overlays; init talk.mp4 terminal.mp4; transcribe; transcript; transcript --suspects; edit tighten; edit cut (retake); edit cut (cat aside); overlay add clip; timeline; timeline --joins; render; check; look --output 10.15. (Some were chained in one shell call.)
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** the retake-kept warning right after tighten printed a command with --near; I used --from/--to instead and it was fine. The overlay began on "open the terminal", so the clip starts after "Finally," which is correct but I had to infer it from times.
- **Missing:** nothing required. A way to see the clip's own picture is only the look frame, which was enough. The "terminal" fixture is a fractal, so I could not confirm it is a terminal.
- **Not sure it's right:** the words check uses base.en, which confirms the cut, not the recogniser. I did not listen to joins (only read them).
- **Best part:** quoting transcript words for cuts and the anchored clip overlay; `check` verified overlay frames and sync directly.
- **One change:** have `edit tighten` offer to apply the retake-kept fix (or the cuts) in one step.
