# Friction report
- **Outcome:** done. out/final.mp4 is 13.767 s, display rotation 0 (rotation baked in), `cutroom check` PASS with 0 warnings.
- **Commands run:** 9 (plus `ls`): guide; guide workflow/edit/pitfalls (one call); init; inspect; transcribe; transcript; edit tighten + 2x edit cut + timeline --joins (one call); timeline + render + check (one call).
- **Errors hit:** none.
- **Guessed wrong:** nothing. I also cut the "UM, the first thing" retake, which the task didn't mention (the transcript flagged it [repeat]).
- **Confusing output:** `tighten` output printed a retake warning before my cuts, which is expected, but the edit-cut commands re-print it each time until resolved. Fine.
- **Missing:** nothing needed. I did not view frames to confirm upright orientation; I relied on `check` "display rotation 0°".
- **Not sure it's right:** whether the picture is actually upright (check compares to source frames and reports rotation 0, but no explicit "upright" test); the cat cut was judged from base.en transcript only; I did not listen to joins.
- **Best part:** quoted-text cuts and the `handled` rotation notice; check output was complete and unambiguous.
- **One change:** have `check` (or render) include a frame image/orientation confirmation so rotation correctness is verified visually, not just by metadata.
