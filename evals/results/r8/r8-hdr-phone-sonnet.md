# Friction report

- **Outcome:** done. out/final.mp4 is 13.8 s, 1280x720. HDR was tone-mapped to SDR bt709, the 180° rotation was baked in, and `cutroom check` passed with 0 failures and 0 warnings.
- **Commands run:** 9 (init, inspect, transcribe, transcript, guide edit, edit cut x2, edit tighten, timeline, timeline --joins, render, check). I also ran `guide`, `guide workflow` and `guide pitfalls`.
  Order: guide; guide workflow; guide pitfalls; init; inspect; transcribe; transcript; guide edit; edit cut (false start); edit cut (cat); edit tighten; timeline; timeline --joins; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** `init` already printed the HDR and rotation notes, so the separate `inspect` was redundant. Segment ids in `timeline` came out of order (s2, s4, s3), which looked odd but is harmless.
- **Missing:** nothing needed. I had no way to look at a frame myself (`cutroom look`) to confirm the picture is upright, but I did not run it. `check` reported the sampled frames as upright and matching.
- **Not sure it's right:** the transcript came from the small base.en model. The check re-transcribed with the same model, so it confirms the cut but not the recogniser. I did not listen to the joins, though the join pauses of about 0.3 s looked fine.
- **Best part:** HDR and rotation were handled automatically and the `check` output states them explicitly (tone, picture, format lines). The cut-by-quoted-words approach was easy.
- **One change:** have `init` print the next step as "transcribe, then cut", so `inspect` is not needed as a separate step.
