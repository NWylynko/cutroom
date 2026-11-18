# Friction report

- **Outcome:** done. Undid the bad mid-sentence cut, cut the retake and the cat aside, re-rendered. `cutroom check` PASS with 0 warnings. The render transcript reads as complete sentences.
- **Commands run:** 11: guide; guide --all | tail; transcribe; transcript; transcript --suspects; timeline; edit history; edit undo, edit cut (retake), edit cut (cat aside), edit tighten, timeline, timeline --joins (one chained call); render; check; transcript --file out/final.mp4.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant.
- **Confusing output:** `transcribe` was needed even though a project already existed, and it printed a "small model" note. Pauses in the original (1.03 s, 1.6 s) were already tightened to 0.3 s, so the "long pause" problem was not visible in the transcript at first glance.
- **Missing:** nothing important. A single "diagnose" command listing all timeline problems (bad cut, retake, aside) with proposed fixes would help.
- **Not sure it's right:** I never listened or looked at the cuts; I relied on check and the re-transcription (base.en). I did not verify the picture visually. I also treated the cat aside as off-topic and removed it, which was a judgment call.
- **Best part:** timeline warnings (broken-sentence, retake-kept) pointed straight at the problems; `edit history` + `undo` made the bad cut trivial to revert; `transcript --file` confirmed the final text.
- **One change:** have `timeline` suggest the exact fix command for each warning (e.g. the undo for broken-sentence) and flag off-topic asides.
