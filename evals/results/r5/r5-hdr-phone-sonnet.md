# Friction report

- **Outcome:** done. out/final.mp4 (13.77 s, 1280x720, bt709, upright) rendered; `cutroom check` PASS with 0 failures and 0 warnings. A frame was viewed: upright, normal colours.
- **Commands run:** about 13. guide; guide workflow + guide edit; init clip.mov; inspect; transcribe; transcript; transcript --suspects; edit tighten; edit cut (false start); edit cut (cat); edit cut 16.9-17.4; timeline; timeline --joins; guide render; render --preview; check; render; check; look out/final.mp4 2.
- **Errors hit:** `error[nothing-cut]: nothing to cut in clip 16.9-17.4`, fix: "that range is not in the timeline (already cut?)". It was clear; tighten had already removed the sound. No real obstacle.
- **Guessed wrong:** nothing significant. I cut the stray sound range before checking whether tighten had removed it. The tighten output did say so.
- **Confusing output:** the `retake-kept` WARN after tighten came with a ready-made cut command. I used my own `--from/--to` cut instead, which also worked. The cut-by-words output reports a source time range (0:06.200-0:09.433) that differs from the transcript times; that is fine but slightly surprising.
- **Missing:** nothing needed. HDR and rotation were handled automatically; the only evidence was the "handled" lines and the look frame. I would like `check` to state explicitly that tone mapping was applied (it only says "colour bt709").
- **Not sure it's right:** HDR-to-SDR tone-mapping quality. I judged it by one test-pattern frame only, and `check` has no explicit HDR/brightness line. The transcript came from the base.en model, so words may be misheard, though the checks passed.
- **Best part:** `init` and `inspect` flagged rotation and HDR as handled. Quoted-word cuts and the one-shot `check` made it quick.
- **One change:** add a "tone-mapped HLG->SDR, luminance compared to source" line to `check`/render output so HDR correctness is verified rather than inferred.
