# Friction report

- **Outcome:** done. The old edit had cut "going to show you how" (broken sentence), kept the retake and the cat aside, and left a clipped word. I reset and redid it. `check` now passes with 0 warnings.
- **Commands run:** 8: `guide`; `guide workflow` + `guide edit` + `check`; `transcript` + `--suspects`; `edit reset`; `edit tighten`; `edit cut` (retake); `edit cut` (aside); `timeline --joins`; `timeline`; `render`; `check`.
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** the first `check` passed with two WARNs, so a PASS can still hide the broken sentence. Only reading the transcript showed it.
- **Missing:** nothing important. `check` could flag a segment pair cut mid-sentence.
- **Not sure it's right:** I did not listen to the joins. I relied on `check` (all 44 words heard) and the join pauses of 0.30 to 0.33 s.
- **Best part:** `edit reset` plus quoted-text cuts. `transcript` flagged the repeat and the aside.
- **One change:** have `check` or `timeline` warn when a pinned cut removes words in the middle of a sentence.
