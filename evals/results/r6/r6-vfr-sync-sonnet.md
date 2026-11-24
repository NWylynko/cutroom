# Friction report

- **Outcome:** done. out/final.mp4 rendered (17.233 s, 517 frames); `cutroom check` PASS with 0 failures and 0 warnings, including sync (worst 0 ms).
- **Commands run:** 11 (cutroom, each separately or chained): guide; guide workflow; guide edit; guide pitfalls; init talk-vfr.mp4; inspect; transcribe; transcript --suspects; transcript; edit tighten; timeline; edit cut --from "The first thing you need" --near 6.31 --to "first thing."; timeline --joins; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant. I left the cat aside ("By the way, my cat...") in because the task only asked for pauses and the false start. I judged that call myself.
- **Confusing output:** none. The VFR and audio-offset warnings were marked "handled", so I did nothing about them. The pasted `retake-kept` fix command worked as printed.
- **Missing:** nothing needed. I would have liked a one-line statement that VFR sync is handled automatically at render time. `check` does say this, but only after the render.
- **Not sure it's right:** I did not listen to the joins or look at frames. I relied on `check`. The transcript came from the small base.en model.
- **Best part:** the `retake-kept` warning printed a ready-to-run cut command. `check` measured sync per segment against the source.
- **One change:** make `inspect` say explicitly that VFR and audio offset will not cause drift, because the person asked about lip sync.
