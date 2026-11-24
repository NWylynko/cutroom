# Friction report

- **Outcome:** done. The earlier editor had tightened and cut the cat aside; I cut the remaining retake. out/final.mp4 is 13.767 s and `check` passed with 0 failures and 0 warnings.
- **Commands run:** 8 (`guide`; `status` + `guide workflow` + cat timeline.json; `transcript` + `--suspects` + `timeline`; `edit cut`, `timeline --joins`, `render --preview`, `check preview`; `render`, `check final`).
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** none. The `retake-kept` warning gave a ready-made `edit cut` command that worked as written. The `--suspects` output marking the sound as "(already cut)" helped.
- **Missing:** no way to see the edit history (what the previous person did and why). Status only shows the last edit. I did not verify the first take's "UM" fragment by listening.
- **Not sure it's right:** the check cannot confirm that every sentence is complete in meaning. I judged that from the transcript. The last segment ends at 29.5 s, and the source runs to 30.26 s.
- **Best part:** `status` showed where the work stood, and the warning's fix line was copy-paste.
- **One change:** add an `edit log` command that lists previous edits.
