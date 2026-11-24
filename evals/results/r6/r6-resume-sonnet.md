# Friction report

- **Outcome:** done. Previous editor had tightened and cut the cat aside; I cut the remaining retake, rendered, and `check` passed with 0 warnings.
- **Commands run:** 6 — `cutroom guide`; `cutroom guide --all` (after cat of cutroom.json/timeline.json); `cutroom transcript` + `--suspects` + `timeline` + `edit history`; `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."`; `cutroom timeline --joins`; `cutroom render`; `cutroom check`.
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** none really. Resuming was easy because `timeline` and `edit history` showed prior work; the retake-kept warning printed the exact command to run.
- **Missing:** nothing needed. I did not listen to or look at the final render (only check + joins), so a `check` summary of "what the render says" is only a suggested next step.
- **Not sure it's right:** I did not read `cutroom transcript --file out/final.mp4` or look at cut frames; relied on check's word/stray-sound results. Transcript says the retake "UM" take was the earlier one, so keeping the later clean take seems right.
- **Best part:** the ready-made `edit cut` in the retake-kept warning, and `edit history` for resuming.
- **One change:** have `timeline` for a resumed project print a one-line "state: edited, N warnings, render stale/none" so the pickup point is obvious.
