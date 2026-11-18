# Friction report
- **Outcome:** done. Undid a bad cut that removed "going to show you how to", cut the retake and the cat aside, re-rendered; `cutroom check` PASS, 0 warnings.
- **Commands run:** 9: guide; guide --all (+ls); transcribe/transcript/--suspects/timeline; edit history; edit undo + 2x edit cut + tighten + timeline + --joins; render + check; FRICTION write.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant. Quoting "..., um, ..." text in --from/--to worked first time.
- **Confusing output:** `guide --all` is long; I tailed it. Timeline initially showed an old render state, so I had to infer which edit was the bad one from `edit history` (undo was clean).
- **Missing:** nothing needed. Did not listen/look at joins (could not view images); relied on check and transcript.
- **Not sure it's right:** the remaining pauses (~0.3 s) and the "take 2 vs take 1" choice were not auditioned; the small base.en model may mishear.
- **Best part:** timeline warnings (edge-in-word, retake-kept) pointed directly at the problems; `check` is a clear verdict.
- **One change:** have `timeline` flag cuts that remove a mid-sentence span (a cut dropping words with no retake/aside) as a "sentence broken" warning.
