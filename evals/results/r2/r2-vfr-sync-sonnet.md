# Friction report
- **Outcome:** partly done / done with one unresolved check. out/final.mp4 rendered (17.233 s, 517 frames = timeline). `cutroom check` fails 1 item (`picture`); I judged it a false positive but could not prove it with the tool.
- **Commands run:** 13 (cutroom): guide, guide workflow, guide pitfalls, init, inspect, transcribe, transcript --suspects, transcript, guide edit, edit tighten, edit cut --from "The first thing you need is, um" --to "the first thing.", timeline --joins, render, check (+ check --json, guide verify, look x4, in a few combined calls).
- **Errors hit:** `look --render` -> error[unknown-flag], fix line (`look --help`) got me to `--file`. Fine.
- **Guessed wrong:** `look` flag for a rendered file (guessed --render; it is --file).
- **Confusing output:** `FAIL picture ... offset -2 frames, difference 0.0` -- "difference 0.0" suggests frames are identical, so the offset is unmeasurable, yet it is reported as FAIL. Its fix line says `cutroom look 10.20 --source talk-vfr` but gives no equivalent command for the render frame. Source frame f306 and render frame look identical (static test pattern); other 17 samples and per-segment audio sync (worst 0 ms) pass.
- **Missing:** a way to see source-vs-render frame side by side for the failing sample; a statement whether a 0.0-difference mismatch means "ambiguous/static content". Also the exit code 4 made my chained `&&` render+check look like a render error.
- **Not sure it's right:** the picture FAIL (VFR source with 11 dropped-frame gaps; offset -2 near source 10.2 s). Lip-sync itself measured OK. I left the cat aside in (task only asked for pauses + false start).
- **Best part:** transcript --suspects flagged the retake; `edit cut --from/--to` by quoted words; init reported VFR and A/V offset as "handled".
- **One change:** make `check` picture treat zero-difference (static/identical) frames as inconclusive rather than FAIL, or say how to confirm.
