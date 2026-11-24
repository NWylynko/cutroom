# Friction report: r6-tighten-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s, down from 30.261 s) passes `cutroom check` with 0 failures and 0 warnings. The retake, the cat aside and the stray sound are gone, and the transcript of the final file reads as the intended explanation.

- **Commands run:** 10 shell calls, 29 `cutroom` invocations, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide perceive`, `cutroom guide edit`
  3. `cutroom guide render`, `cutroom guide verify`, `cutroom guide pitfalls`
  4. `cutroom doctor`, `cutroom init talk.mp4`, `cutroom inspect`, `cutroom transcribe`
  5. `cutroom transcript talk`, `cutroom transcript talk --suspects`, `cutroom transcript talk --words`
  6. `cutroom edit tighten`, `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`, `cutroom edit cut --from "By the way" --to "keyboard." --dry-run`
  7. `cutroom edit cut` (same two, applied), `cutroom timeline`, `cutroom timeline --joins`
  8. `cutroom render --preview --dry-run`, `cutroom render --preview`, `cutroom check`
  9. `cutroom render`, `cutroom check`, `cutroom transcript --file out/final.mp4`
  10. `cutroom look --cuts --file out/final.mp4`, `cutroom listen --output 3.4-4.6 --file out/final.mp4`

- **Errors hit:** none. No command failed and no `fix:` line was needed.

- **Guessed wrong:** Nothing failed, but `cutroom transcript talk --words` with no range printed about 40 lines and cut off with "… 24 more lines". I had only wanted an overview, so the flag combination was a guess that gave more than I needed. `--words` without `--from`/`--to` is not very useful for a 30 s clip.

- **Confusing output:**
  - `cutroom listen --output 3.4-4.6 --file out/final.mp4` reported `heard (isolated clip): forever. The first thing`, while the words were planned as `home server. The first`. The full-file transcript of the final output says "home server." correctly. The isolated re-transcription was wrong, and nothing in the output explains how much to trust it. I resolved it only by checking the full transcript.
  - `cutroom edit cut --dry-run` prints `join: …up a home server. | The first thing you… (pause 0.30 s)`. That is clear. The join lines in `timeline --joins` were also clear.
  - The `retake-kept` warning after tighten suggested `--near 6.31` even though `--from` was unique. Both would work, but the suggested command is a little longer than needed.

- **Missing:** Nothing I needed was absent. I would have liked `cutroom check` to mention that a `sound` suspect inside a removed pause is dropped by tighten. The tighten output said this, but only as "this also removes the sound with no words", which was good.

- **Not sure it's right:**
  - Transcription used the `base.en` model. The suspects and word timings are approximate, so I checked the cuts by eye with `cutroom look --cuts` (all four joins show a clean, shot-to-shot cut on the test pattern) and by reading the final transcript. I did not listen to the audio joins myself. The `check` words, sync and stray-sound checks all passed.
  - I left one 0.30 s pause at each join, which seemed natural. I did not verify whether the pacing sounds right to a viewer.
  - I cut the first take of the retake and kept the second, which is "The first thing you need is an old laptop." The first take had "UM" and a restart, so the choice was clear.

- **Best part:** The sequence `tighten`, then `transcript --suspects`, then `edit cut --dry-run`, then apply, then `timeline --joins`, then `render --preview`, then `check` made each step's effect easy to see and verify. The dry-run join preview showed exactly what would be removed before I applied it.

- **One change:** Make `cutroom listen` state its confidence next to the isolated re-transcription. Right now a wrong isolated word ("forever." for "home server.") looks as authoritative as a correct one, and I had to cross-check the full transcript to see that it was wrong.
