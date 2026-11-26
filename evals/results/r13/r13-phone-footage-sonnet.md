# Friction report

- **Outcome:** done. out/final.mp4 rendered, cat aside and false-start retake and long pauses cut, 180 degree rotation baked into pixels with no rotation tag; `cutroom check` passed with 0 failures and 0 warnings.
- **Commands run:** 11 — guide; guide workflow/edit/pitfalls (one call); init; inspect; transcribe; transcript (one chained call); edit tighten; edit cut (aside); edit cut (retake); timeline; timeline --joins; render; check; ffprobe (cross-check).
- **Errors hit:** none.
- **Guessed wrong:** nothing. I expected to need a rotation flag; none was needed.
- **Confusing output:** the tighten output was long (all pause lists). The timeline printed WARNs before I had cut, which is fine but redundant with the later "no problems found".
- **Missing:** nothing essential. A way to view the upright frame in the same step (`look`) would be a nice nudge in the render output; I did not run it.
- **Not sure it's right:** I did not view frames myself (no `cutroom look`); relied on check's "picture ... upright" line and ffprobe showing no rotation side data. Word check uses the same base.en model as the transcript.
- **Best part:** render automatically bakes rotation, and `check` explicitly states "no rotation tag: upright in any player".
- **One change:** have `init`/`inspect` say outright "render will bake this rotation into pixels", so an agent needn't wonder whether a flag is needed.
