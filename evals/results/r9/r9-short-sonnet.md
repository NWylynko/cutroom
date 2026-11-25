# Friction report
- **Outcome:** done. out/short.mp4 (720x1280, 7.27 s) rendered; `check` PASS with 0 failures/warnings; looked at a captioned frame.
- **Commands run:** 13: guide; guide shorts + guide workflow; init; transcribe; transcript; short --from/--to --name steps; edit cut (cat aside); timeline; edit frame --size 720x1280; look --output 1; captions --burn; guide captions; render out/short.mp4; check; look (file) x2.
- **Errors hit:** `error[no-timeline]: no timeline yet in .` when I ran `look --file` without `--project shorts/steps`; the fix line pointed at --project, which solved it.
- **Guessed wrong:** `look --output 1.5 --file out/short.mp4` (wrong; the form is `look out/short.mp4 1.5`).
- **Confusing output:** the short's timeline shows segment ids s1,s2,s4 (gaps after the cut) - harmless. Source is a synthetic test pattern so framing cannot be judged by content.
- **Missing:** nothing important. A note that --project is needed for every short command, even look on a render, could be more prominent.
- **Not sure it's right:** base.en transcript could mishear words; check re-transcribed with same model. Caption legibility/placement judged from one frame only.
- **Best part:** `short` + `--from/--to` quoted text, `captions --burn` auto-placing in safe zones, and the thorough `check`.
- **One change:** have `short` print a suggested `edit cut` for detected asides (the cat sentence) or flag off-topic sentences inside the span.
