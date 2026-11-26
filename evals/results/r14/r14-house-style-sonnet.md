# Friction report

- **Outcome:** done. out/final.mp4 is 13.8 s with retake, cat aside and pauses removed, two left-layout cards in the house style, `cutroom check` PASS with 0 warnings.
- **Commands run:** about 12 (guide x6, init, transcribe, transcript, edit tighten, edit cut x2, timeline, overlay add x2, style cards, style --help, overlay set, look --overlays, render, check, style cards).
- **Errors hit:** none. Only warnings: `overlay-short` on o1 (fixed with `overlay set o1 --duration 1.6`, its fix line worked).
- **Guessed wrong:** nothing material. `--radius 0.3` was a guess for "rounded"; unsure of the target value.
- **Confusing output:** `overlay add` repeated the same overlay-short warning on later commands (noisy). The tighten message about the 3.16 s pause hid the cough, only noted in passing.
- **Missing:** no way to see eyebrow colour verified numerically; `style cards` does not show eyebrow/icon colour (they follow `accent`, which I only knew from the guide). Would like a per-card "what is drawn" dump.
- **Not sure it's right:** the "slightly see-through" box used #F3F6F7e6 (90%) - my choice. Did not view frames of the final render itself, only `look --overlays` pre-render (check confirmed visibility windows). Transcript is base.en, so word matching is only as good as that.
- **Best part:** quoted-text cuts and anchors (`--at "Then install" --to "on it."`), plus `look --overlays` showing the cards as rendered.
- **One change:** have `overlay add` auto-extend or suggest the duration upfront when the spoken window is shorter than the read time, instead of warning after.
