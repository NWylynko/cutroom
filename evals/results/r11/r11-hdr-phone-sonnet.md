# Friction report

- **Outcome:** done — out/final.mp4 rendered, SDR bt709, upright, `cutroom check` PASS with 0 warnings.
- **Commands run:** 9 cutroom commands (+ 1 ls): guide; guide workflow/edit/verify; init clip.mov; inspect; transcribe; transcript; edit tighten; edit cut (false start); edit cut (cat aside); timeline; timeline --joins; render --preview; render; check. (Several were chained in single shell calls.)
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** `edit tighten` printed a retake-kept warning whose suggested cut differs from the one I used (it said --near); harmless. Cut output said "0:06.200" etc. while transcript said 0:06.310 (pad explanation is documented, but briefly surprising).
- **Missing:** nothing. I did not visually look at a frame; relied on check's `tone` and `picture` lines.
- **Not sure it's right:** I never viewed a frame myself to confirm colour/orientation; trusting check. Base.en transcript was small-model; words check confirmed against same model only.
- **Best part:** init/inspect flagged HDR and rotation as `handled`, render handled both automatically, and check verifies tone and upright orientation explicitly.
- **One change:** have `init` or `render` suggest `cutroom look --output` on one frame by default for HDR/rotated sources so the agent can eyeball the result.
