# Friction report
- **Outcome:** done. out/final.mp4 rendered (13.8 s) and `cutroom check` PASS, 0 warnings.
- **Commands run:** 9 — guide; guide workflow + guide edit; init talk.mp4 pickup.mp4; transcribe + transcript --suspects + transcript; transcript pickup; edit tighten + cut (um false start) + cut (cat aside) + replace --text "Then install Linux on it" --with pickup + timeline + timeline --joins; render + check; (this report).
- **Errors hit:** none.
- **Guessed wrong:** nothing; the workflow guide's pickup recipe worked as written.
- **Confusing output:** the audio-mixed warning repeats on every timeline call, and I could not tell whether it mattered. Check showed sync fine. Retake-kept warning text after tighten disappeared once I cut; fine.
- **Missing:** nothing essential. I would like a way to hear/verify that pickup and talk loudness match (check only reports overall loudness).
- **Not sure it's right:** the audio-mixed note (different mic?) can't be verified; the base.en transcript could misname words, though check heard all 44 expected words.
- **Best part:** `edit replace --text ... --with pickup` in one step, and quoted-text cuts with printed joins.
- **One change:** let the audio-mixed warning say whether the loudness/room tone of the two sources actually differ, so I know if it is ignorable.
