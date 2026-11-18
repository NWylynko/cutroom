# Friction report

- **Outcome:** done. out/final.mp4 rendered (17.233 s, from 30.261 s); `cutroom check` PASS with 0 failures and 0 warnings, including sync (worst 0 ms) and picture on a VFR source.
- **Commands run:** 9 cutroom calls: guide; guide workflow; guide edit; guide pitfalls; init talk-vfr.mp4; inspect; transcribe; transcript; transcript --suspects; edit tighten; edit cut --from "The first thing you need is, um" --to "the first thing."; timeline --joins; render; check. (Several were chained in one shell call.)
- **Errors hit:** none.
- **Guessed wrong:** nothing significant. I did not need a separate sync step: VFR and the +0.064 s audio offset were marked `handled`.
- **Confusing output:** `edit tighten` warned retake-kept ("the first thing you need is" 2x) before I made the cut; expected, and the warning went away after the cut. I left the cat aside in on purpose (task only asked for pauses and the false start), and cutroom did not flag it as an aside.
- **Missing:** nothing needed. A one-line statement that VFR sources are re-timed to constant fps (and that sync is therefore safe) would reassure; the sync check line covered it.
- **Not sure it's right:** the transcript is from base.en, so some words may be misheard, but the false-start text was clear. I did not listen to the joins; relied on check's word/sync/stray-sound results.
- **Best part:** `inspect` marking vfr/av-offset as "handled", and `check` proving sync per segment against the source.
- **One change:** have `edit tighten` suggest/preview the retake cut from its own suspects so the retake-kept warning comes with a ready command.
