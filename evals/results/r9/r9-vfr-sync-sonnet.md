# Friction report

- **Outcome:** done. out/final.mp4 rendered, `cutroom check` PASS (0 failures, 0 warnings), sync median -0 ms.
- **Commands run:** 11: guide; guide workflow/edit/pitfalls; init; inspect (output of init); transcribe; transcript --suspects; transcript; edit tighten; timeline; timeline --apply retake-kept@4.034; timeline --joins; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant. I kept the "cat walked across the keyboard" aside since the request only named pauses and the false start.
- **Confusing output:** the transcript's repeat suspect/retake wording is clear. `timeline --apply` needs the exact output time of the warning (4.034); I passed it from the printed text.
- **Missing:** nothing needed. A VFR-specific verify line in check (it reports sync well already) was enough.
- **Not sure it's right:** whether the user wanted the cat aside removed (ambiguous). base.en word check only confirms cuts, not recogniser.
- **Best part:** VFR and audio offset were handled automatically, and render/check explained exactly why sync is safe.
- **One change:** have `tighten` output state that the retake warning's fix is the false-start removal, so the task is one step shorter (it nearly does).
