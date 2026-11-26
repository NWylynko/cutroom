# Friction report

- **Outcome:** done. out/final.mp4 rendered (SDR bt709, upright, 13.8 s); `cutroom check` PASS with 0 failures/warnings.
- **Commands run:** 10 — guide; guide workflow; guide edit; init clip.mov; inspect; transcribe; transcript; edit tighten + edit cut (x2) + timeline; render + check (chained).
- **Errors hit:** none.
- **Guessed wrong:** nothing. HDR and rotation were handled automatically; no flags needed.
- **Confusing output:** `tighten` prints a long block with two WARNs; slightly dense but the suggested cuts were clear. Note the retake suggestion cuts the earlier take, which was what I wanted.
- **Missing:** nothing needed. I did not view frames myself (`cutroom look`); relied on check's picture/tone lines.
- **Not sure it's right:** I never visually looked at a frame to confirm upright/not washed out; I trusted check's "picture" and "tone" comparisons against the source. Also base.en transcript only; cut boundaries not listened to.
- **Best part:** init/inspect flagged HDR and rotation as "handled", and render auto-tone-mapped and baked in rotation; check verified both explicitly.
- **One change:** have `render` print a suggested `cutroom look` frame of the output so the agent can view it before claiming done.
