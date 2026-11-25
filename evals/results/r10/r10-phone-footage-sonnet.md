# Friction report

- **Outcome:** done. Cat aside, the "UM" false start and long pauses cut; final.mp4 is upright with no rotation tag; `check` PASS with 0 warnings.
- **Commands run:** 8: guide; guide workflow + guide pitfalls; init talk-rot.mov; transcribe + transcript; edit tighten + 2x edit cut + timeline --joins; render + check; ffprobe; this report.
- **Errors hit:** none.
- **Guessed wrong:** nothing. I had not expected to need to do anything for the 180 degree rotation, and I didn't.
- **Confusing output:** `edit tighten` printed WARNs (retake-kept, aside-kept) before my cuts, which made them read as current. Later output noted "still open (unchanged by this edit)", which was clear enough.
- **Missing:** nothing important. I'd have liked `check` to say which source words were intentionally cut.
- **Not sure it's right:** I could not see the picture. check says frames match as a player shows them, which I trust. I also cut the "UM" first take as a retake, which is a judgement call.
- **Best part:** rotation handled automatically and check explicitly verifies "upright in any player"; the retake and aside warnings gave ready-made cut commands.
- **One change:** have `init` or `transcribe` suggest the retake and aside cuts up front, so I need not run tighten first to learn them.
