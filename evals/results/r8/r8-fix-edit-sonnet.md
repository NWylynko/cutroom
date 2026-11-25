# Friction report

- **Outcome:** done. Reset the inherited edit, tightened, cut the retake and the cat aside, re-rendered; `check` PASS 0 failures/0 warnings.
- **Commands run:** 8 substantive: guide; guide workflow + guide edit + cat timeline.json + status; transcript + transcript --suspects; edit reset; edit tighten; edit cut (retake); edit cut (cat aside); timeline; timeline --joins; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant.
- **Confusing output:** none major. Inherited bug was a 1.9-2.9 cut mid-sentence ("going to show you how" lost), plus retake kept, cat aside kept, pauses long (tighten had not been run).
- **Missing:** nothing essential. A single "fix all warnings" command would be nice, but reset + 3 edits was quick.
- **Not sure it's right:** words check uses the same base.en model, so it confirms the cut, not recogniser accuracy; I did not listen to the audio.
- **Best part:** `edit reset` plus quoted-text cuts, and `check` verifying every word.
- **One change:** have `status`/`timeline` on an inherited edit point to `edit reset` when several problems exist.
