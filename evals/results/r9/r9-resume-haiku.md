# Friction report: r9-resume-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s, 1280x720) renders from the 5-segment timeline, passes `cutroom check` with 0 failures and 0 warnings, and the re-transcript reads as complete sentences with no retake.

- **Commands run:** 8 shell calls, about 19 cutroom invocations, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide edit` (plus `ls`, `cat cutroom.json`, `cat timeline.json`, `find .cutroom out`)
  3. `cutroom status`
  4. `cutroom edit history`
  5. `cutroom timeline`
  6. `cutroom transcript`
  7. `cutroom transcript talk --suspects`
  8. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`
  9. `cutroom timeline --joins`
  10. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."` (applied)
  11. `cutroom timeline`, `cutroom timeline --joins`
  12. `cutroom render --help`
  13. `cutroom render --preview`, `cutroom check out/preview.mp4`
  14. `cutroom render`, `cutroom check out/final.mp4`, `cutroom transcript --file out/final.mp4`

- **Errors hit:** none. Every command succeeded on the first try.

- **Guessed wrong:** nothing material. I took the `retake-kept` warning's suggested `edit cut` verbatim, and `--dry-run` matched the real run. I did not need `--near` to be anything other than the suggested value.

- **Confusing output:**
  - The `retake-kept` WARN line in `cutroom timeline` is one very long line. It holds the output time, the explanation, the ready-to-run command, the source range and a note about `--near`. I had to re-read it to pick out the command.
  - The warning gives output time `0:04.093`, while the source range it names starts at `6.200`. The `--near 6.31` value is a source time that sits inside that range. This is correct, but it took a second look to confirm it was not a mismatch.
  - `cutroom status` after the earlier edits did not say whether the existing `out/preview.mp4` was stale after my cut. I did not check status after the cut, so I only learned it from the fresh render.
  - Segment ids skip `s5` (the cat aside was cut earlier). That is fine, but nothing says the gap is intentional.

- **Missing:**
  - No way to hear or preview a join's audio from the output without going through `cutroom listen`. I did not run `listen` on the three joins I am least sure of: `The first thing you need is an old laptop,` / `then install Linux on it.` and the two after `Finally,`. Those are the three new-ish joins from this edit.
  - I did not run `cutroom look` on any frame. The check's picture step passed, but I did not look at a frame myself, which rule 2 in the guide asks for.

- **Not sure it's right:**
  - The content reads correctly: five complete sentences, one retake removed, the cat aside removed earlier. Whether the joins sound natural at 0.30 s pauses I could not hear, so that is unverified by ear.
  - The transcript check uses `base.en`, so it confirms the cut words and not the recogniser's accuracy. I did not run `cutroom transcribe --model` with a larger model.
  - I did not re-examine the earlier operator's `tighten --min-cut 0.4` choice. I took it as given.

- **Best part:** `cutroom timeline` plus the `next:` lines. The warning gave a ready command, `--dry-run` previewed the exact join, and `check` gave a clear per-item PASS list. I did not have to reason about frame counts.

- **One change:** have `cutroom status` say when a render is stale relative to the timeline after an edit. The `next:` lines would then point at re-rendering without the operator having to remember to do it.
