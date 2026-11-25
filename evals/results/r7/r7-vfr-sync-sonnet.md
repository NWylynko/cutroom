# Friction report

- **Outcome:** done. out/final.mp4 rendered (17.233 s) and `cutroom check` PASS, 0 failures/warnings, sync median -0 ms across 6 of 6 segments.
- **Commands run:** 12 — guide; guide workflow; guide pitfalls; guide edit; doctor; init talk-vfr.mp4; inspect; transcribe; transcript; transcript --suspects; edit tighten; edit cut --dry-run; edit cut; timeline --joins; render; check (some chained in one shell call).
- **Errors hit:** none.
- **Guessed wrong:** nothing. I wasn't sure whether the VFR source or +0.064 s audio offset needed action; inspect marks both "handled", and check's sync line confirmed it.
- **Confusing output:** the cut "--from 'The first thing you need is, um'" needed no tuning, but the retake-kept hint used a different (--near) form than I used; both are fine. Transcript marks the false start as "[repeat]" on two phrases, slightly ambiguous about which to drop.
- **Missing:** nothing essential. I would have liked an explicit statement from `render` that VFR was normalised to constant frame rate.
- **Not sure it's right:** I kept the "cat walked across the keyboard" aside since only pauses and the false start were requested; the user might want it removed. Did not listen to joins; relied on check.
- **Best part:** check's sync line, which measures the render against the source including the audio offset, resolved the user's lip-sync concern. The retake warning with a ready cut command was also useful.
- **One change:** make `init`/`render` say outright "VFR handled: output is constant 30 fps, sync verified by check" so the lip-sync question is answered without needing to run check.
