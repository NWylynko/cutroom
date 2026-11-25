# Friction report: r10 phone footage (haiku)

- **Outcome:** done. `out/final.mp4` (1280x720, 13.767 s) passes `cutroom check` with 0 failures and 0 warnings. The retake ("The first thing you need is, UM, ..."), the cat aside, the 0.32 s sound and all long pauses are gone, and the render has no rotation tag.

- **Commands run:** 27 shell calls: 25 `cutroom` invocations, plus one `ls -la` and one image Read of the look PNG. In order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide perceive`, `cutroom guide edit`
  3. `cutroom guide render`, `cutroom guide pitfalls`, `cutroom doctor`, `cutroom init talk-rot.mov`, `cutroom inspect`
  4. `cutroom transcribe`, `cutroom transcript`, `cutroom transcript --suspects`
  5. `cutroom edit tighten`
  6. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`
  7. `cutroom edit cut --from "By the way, my cat" --to "walked across the keyboard." --dry-run`
  8. `cutroom listen 16.5-17.5`
  9. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing."`
  10. `cutroom edit cut --from "By the way, my cat" --to "walked across the keyboard."`
  11. `cutroom timeline`, `cutroom timeline --joins`
  12. `cutroom render --preview`, `cutroom check out/preview.mp4`, `cutroom look out/preview.mp4 2.0`
  13. `cutroom render`, `cutroom check out/final.mp4`

- **Errors hit:** None. No cutroom command failed, so no `fix:` line was needed.

- **Guessed wrong:** Nothing that broke. I did expect `transcript --suspects` to print the exact `edit cut` command for each suspect. It printed only the suspect and a `next:` pointer. The ready commands showed up only later, in `timeline` WARN lines after `tighten`.

- **Confusing output:**
  - `cutroom check` repeats the rotation fact in the render output ("picture: talk-rot's 180° display rotation is baked into...") and again in the `format` and `picture` check lines. Three near-identical lines is noise.
  - The `words` check line is long. It explains that the re-transcription uses base.en and "confirms the cut, not the recogniser", which is useful, but it should be shorter.
  - `timeline` lists the `aside-kept` warning as "still open" after the retake cut, mixed in with the joins output. It is correct but easy to misread.

- **Missing:**
  - A visual check of orientation. The fixture is a test card (colour bars with a text strip), so `cutroom look` shows nothing that tells me the picture is upright. I relied on the check's "upright in any player" line, which is not the same as seeing a face or text in the right way.
  - A way to confirm that `transcribe` with base.en actually heard the words, not just that the render matches its own transcript. The `words` check uses the same model, so it can't catch a misheard word. I would have liked the bigger model in one line, for example `cutroom transcribe --model small.en`, without digging through `doctor`.

- **Not sure it's right:**
  - The cuts rest on base.en word times. The retake boundary ("is, UM, the first thing.") and the aside start ("By the way,") are from the transcript, and the join listing reads fine, but I did not `cutroom listen` the two cut edges in the final, only the sound at 16.97 s.
  - The render is upright according to the tool. I did not confirm that visually on a real speaker.
  - Total duration dropped from 30.26 s to 13.77 s. Most of that is pause tightening, which I did not hand-check beyond the `tighten` summary.

- **Best part:** `cutroom edit cut ... --dry-run` showed the exact span removed, the join, and the pause before committing. Then `cutroom check` gave one line per property, including "upright in every player" and "all 44 expected words heard in order".

- **One change:** Have `cutroom transcript --suspects` print a ready `cutroom edit cut --from ... --to ...` line for each `repeat` and `aside` suspect, so the first cut does not need `tighten` and a timeline warning to surface the command.
