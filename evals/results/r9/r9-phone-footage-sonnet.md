# Friction report

- **Outcome:** done. out/final.mp4 (13.77 s) has the cat remark, the retake and long pauses removed; check PASS with 0 warnings; rotation baked in (no rotation metadata), a frame looked upright.
- **Commands run:** 9 cutroom + 1 ffprobe + 1 image read. Order: guide; guide workflow/edit/pitfalls; init talk-rot.mov; transcribe; transcript; transcript --suspects; edit tighten; edit cut (cat); edit cut (retake); timeline; timeline --joins; render; check; ffprobe; look --output 2.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant. Initially unsure whether rotation needed a flag; it was automatic.
- **Confusing output:** tighten printed a retake-kept WARN with a suggested cut that keeps the last take; I cut the earlier take with my own quoted words instead, and the warning cleared. Fine, but the numbers (--near) were extra noise.
- **Missing:** nothing needed. Slightly wanted `check` to state explicitly "no rotation metadata" in the format line (it says "display rotation 0°", fine).
- **Not sure it's right:** transcript is base.en; only 44 words verified by the same model. I did not listen to joins (pauses 0.30-0.33 s looked natural).
- **Best part:** the task's rotation trap was handled with zero action, and `check` confirmed picture upright vs source; tighten also removed the stray cough sound automatically.
- **One change:** have `tighten` also offer/apply the retake-kept fix in one step (or mention it in the same output) so the full sequence is a single command.
