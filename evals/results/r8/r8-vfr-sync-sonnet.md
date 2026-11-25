# Friction report

- **Outcome:** done. out/final.mp4 rendered; `cutroom check` PASS, 0 failures/warnings, sync worst -0 ms.
- **Commands run:** 10: guide; guide workflow/edit/pitfalls (one shell call); init; inspect; transcribe; transcript --suspects; transcript; edit tighten; timeline; timeline --apply retake-kept@4.034; timeline --joins; render; check. (Some were chained in one shell call, so about 13 cutroom invocations.)
- **Errors hit:** none.
- **Guessed wrong:** nothing. I kept the "cat walked across keyboard" aside, since only pauses and the false start were requested.
- **Confusing output:** init already ran inspect, so my separate inspect repeated it. The tighten output is long but readable.
- **Missing:** nothing needed.
- **Not sure it's right:** the transcript used base.en, and the words check only confirms the cut, not the recogniser. I did not listen to the joins. The 0.064 s audio offset is kept as recorded and check reports sync as matching the source.
- **Best part:** render states how VFR and the audio offset were handled (constant-grid placement, "can't drift apart"), and check measures per-segment sync. That answered the lip-sync concern directly. `timeline --apply` made the false-start cut one step.
- **One change:** have `init` or `guide` state up front that VFR and offset are handled automatically, so an agent does not worry or hunt for flags.
