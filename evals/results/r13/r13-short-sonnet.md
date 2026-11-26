# Friction report

- **Outcome:** done. out/short.mp4 (720x1280, 7.5 s) made, `cutroom check` PASS with 0 warnings.
- **Commands run:** 10 cutroom commands: guide; guide shorts + guide workflow; init talk.mp4; transcribe; transcript; short --from/--to --name steps; timeline; edit cut (aside); captions --burn; timeline; render out/short.mp4; check; look --output 1.5.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant. I did not know the short would be 720x1280 rather than 1080x1920 until `short` said so (explained and sensible).
- **Confusing output:** `short` output is long, with the size note and WARN lines. Source is a synthetic colour-bar test pattern, so "no speaker found" is expected; I left framing centred.
- **Missing:** nothing needed. A one-line "do the aside cut for you" flag on `short` would save a step.
- **Not sure it's right:** the words check uses the base.en model, so it confirms the cut, not recogniser accuracy. I could not hear the audio; sync/loudness measurements passed.
- **Best part:** `short` + the aside-kept warning with a ready-made cut command; `check` safe-zones and captions lines made verification easy.
- **One change:** have `short` flag/offer cutting the aside automatically (or list the cut command in next:).
