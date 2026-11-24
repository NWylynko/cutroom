# Friction report

- **Outcome:** done. Undid the broken edits, re-tightened, cut the retake and the cat aside, rendered, and `check` passed with 0 warnings.
- **Commands run:** 10 — guide; guide --all | tail; status; edit history/transcript/transcript --suspects/timeline (one shell call); edit undo x2, edit tighten, edit cut (retake), edit cut (aside), timeline, timeline --joins; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing. I assumed the old edit was not worth patching and undid it.
- **Confusing output:** `guide` alone is short, so I needed `guide --all`. `status` listed the render as "current, not checked", which was useful.
- **Missing:** nothing important. I could not listen to the final result myself; I relied on check.
- **Not sure it's right:** I never ran `look --cuts` or `listen`. The retake was chosen by transcript text only.
- **Best part:** the retake-kept warning with a ready cut command, and the single `check` that covers words, sync and stray sounds.
- **One change:** have `status` or `timeline` flag that `tighten --min-cut 0.4` left long pauses, and show the full list of problems in one place.
