# Friction report

- **Outcome:** done. Pauses tightened, false start cut, out/final.mp4 rendered; `check` PASS with 0 warnings (sync worst 0 ms on a VFR source with +0.064 s audio offset).
- **Commands run:** 9: guide; guide --all; init talk-vfr.mp4 + inspect + transcribe + transcript + transcript --suspects (one chained shell call); edit tighten; edit cut --dry-run; edit cut; timeline; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing. I left the "cat walked across the keyboard" aside in because only pauses and the false start were requested; this is a judgement call.
- **Confusing output:** the tighten output printed a retake-kept warning that suggested a cut slightly different from mine (it only removes the first take), which I had to compare with the --from/--to I chose. Minor.
- **Missing:** nothing needed. A VFR-specific line in `check` confirming the source was VFR would reassure; the sync line did say "audio start offset included".
- **Not sure it's right:** I can't hear the audio, so the 0.31 s joins and the cat aside rely on the transcript. I didn't open `look --cuts`.
- **Best part:** VFR and audio offset were handled automatically, flagged by inspect as "handled", and verified by check's sync line.
- **One change:** have `tighten` print the exact suggested `edit cut` for a false start with "um" (the retake suggestion existed, but its --to differed from the transcript text).
