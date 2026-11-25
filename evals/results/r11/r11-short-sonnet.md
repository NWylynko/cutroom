# Friction report

- **Outcome:** done. out/short.mp4 is 720x1280, 7.3 s, three steps, aside cut, captions burned in; `check` passed with 0 failures and 0 warnings.
- **Commands run:** 14 cutroom commands: guide; guide shorts; guide workflow; init talk.mp4; transcribe; transcript --suspects; transcript; short --from --to --name steps; edit cut (By the way..keyboard.); timeline; edit frame --size 720x1280; look --output 0:01; guide captions (twice); captions --burn; timeline --joins; render out/short.mp4; check out/short.mp4; look --output 1.5.
- **Errors hit:** none.
- **Guessed wrong:** nothing material. I assumed `short` would drop the aside automatically; it only warned (aside-kept), as the guide says.
- **Confusing output:** `short` said "4 segments, 10.567 s" and the retake "The first thing you need is, UM" was dropped correctly, but the timeline ids skip s3 (s1, s2, s4), which looks odd. The caption "is an old laptop" starts mid-sentence because of 1-3 word grouping, which reads slightly odd but is fine.
- **Missing:** nothing blocking. A one-command way to cut all flagged asides would be nice. Soft-source warning (720p enlarged 2.7x) is shown but only fixed via a separate `edit frame --size` step; `short --size` default could pick 720x1280 for 720p sources.
- **Not sure it's right:** the footage is synthetic color bars, so I could not judge speaker crop framing. Transcript is from the base.en model; `check` words confirmation uses the same model.
- **Best part:** `short` + `captions --burn` + `check` (safe-zones, words, sync) made it quick and verifiable.
- **One change:** have `short` auto-apply the source-appropriate size (720x1280 for a 720p source) instead of warning about softness.
