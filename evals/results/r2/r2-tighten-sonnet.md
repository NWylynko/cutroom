# Friction report
- **Outcome:** done — out/final.mp4 (13.77 s from 30.26 s) rendered, `cutroom check` PASS with 0 failures/warnings.
- **Commands run:** 10: guide; guide workflow; guide edit; doctor; init talk.mp4; transcribe; transcript; transcript --suspects; edit tighten; edit cut (retake); edit cut (cat aside); timeline; timeline --joins; render --help; render --preview --range 0-14; render -o (error); render; check. (approx 17 invocations, several chained with &&)
- **Errors hit:** `error[bad-flag]: short flags are not supported: -o` / `fix: use the long form, e.g. --o` — the fix was wrong/unhelpful (render takes a positional output path, default out/final.mp4); I just ran plain `cutroom render`.
- **Guessed wrong:** assumed `-o` output flag; the output is positional.
- **Confusing output:** the fix line for -o suggests `--o`, which doesn't exist. Preview range render wrote out/preview-range.mp4, a stray file.
- **Missing:** nothing significant. Base.en transcript was accurate here. No way to see the final video myself beyond check.
- **Not sure it's right:** the cut of the "um" take relied on transcript text; check confirms words but not that the kept sentence "The first thing you need is an old laptop" is the cleaner take (judged by suspects hint).
- **Best part:** `transcript --suspects` flagged the retake and the cough; text-anchored cuts; `check` covers words, sync, stray sounds.
- **One change:** make bad-flag fix lines context-aware (point to the positional output path).
