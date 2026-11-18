# Friction report

- **Outcome:** done. out/final.mp4 is 13.77 s (from 30.26 s); `cutroom check` PASS with 0 failures and 0 warnings.
- **Commands run:** 11 — `guide`, `guide --all`, `doctor`, `init talk.mp4`, `transcribe`, `transcript`, `transcript --suspects`, `edit tighten`, `edit cut --from "The first thing you need is, UM, the first thing." --to "the first thing."`, `edit cut --from "By the way" --to "keyboard."`, `timeline`, `timeline --joins`, `render`, `check`.
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** `tighten` warned `retake-kept` for the "first thing you need is" retake, which I then cut. Expected, but the warning has no "will fix later" hint. The first cut's `--to "the first thing."` worked but I wasn't sure it would match the first occurrence.
- **Missing:** nothing. I did not look at frames or listen to the cat/cough areas; check's stray-sound test passed so I trusted it.
- **Not sure it's right:** base.en transcript could have missed words, but check's "words" test matched all 44 expected words. I didn't view the joins visually (`look --cuts`).
- **Best part:** text-anchored cuts, plus a check that measures the render itself.
- **One change:** have `guide` say that the tighten-time `retake-kept` warning is expected before cuts.
