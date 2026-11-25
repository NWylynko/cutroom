# Friction report

- **Outcome:** done — `out/final.mp4` (13.767 s) passes `cutroom check` with 0 failures and 0 warnings, and its transcript reads as the intended script; the one judgment call (sentence flow) was checked by reading, not by the tool.

- **Commands run:** 17 cutroom invocations plus one image read, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom edit history`
  4. `cutroom timeline`
  5. `cutroom transcript`
  6. `cutroom guide edit`
  7. `cutroom transcript talk --suspects`
  8. `cutroom timeline --joins`
  9. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`
  10. `cutroom guide perceive`
  11. `cutroom timeline --apply retake-kept@4.09`
  12. `cutroom timeline`
  13. `cutroom timeline --joins`
  14. `cutroom status`
  15. `cutroom render --preview --range 0:00-0:14`
  16. `cutroom render`
  17. `cutroom check out/final.mp4`
  18. `cutroom look --file out/final.mp4 --cuts` (then read the PNG)
  19. `cutroom transcript --file out/final.mp4`
  Plus `cutroom status` and `cutroom init`-free inspection via `cat` of `cutroom.json` and `timeline.json` (shell, not cutroom).

- **Errors hit:** none. Nothing failed, so no `fix:` line was needed.

- **Guessed wrong:** nothing material. I expected `--apply` to need the exact warning text; `retake-kept@4.09` worked from the warning's own time. The `--dry-run` on the suggested cut matched the applied result.

- **Confusing output:**
  - The `WARN[retake-kept]` line is very long: it carries the diagnosis, the full command, the removed text and a `--near` note in one line. Usable, but I had to re-read it to find the actionable part.
  - `cutroom transcript --file out/final.mp4` prints the whole render as one paragraph (`Hi, today I'm ... The first thing you need is an old laptop, then install Linux on it. ...`). The 0.30 s pauses between phrases are not shown, so it is hard to tell from this output where the sentence breaks really fall. The source transcript is phrase-by-phrase and clear; the render transcript is not.
  - `cutroom status` at the start said `out/preview.mp4 current, not checked`, but the timeline had already changed since that preview was made in an earlier session. After my edit it became `STALE`. Labelling confused me until I compared the timestamps.

- **Missing:**
  - A way to judge sentence completeness from the tool. "Every sentence complete" is a content judgment; the transcript gives no sentence or phrase boundaries in the render, so I read it by hand.
  - A command to delete or mark stale renders. `out/` now holds `preview.mp4` (stale), `preview-range.mp4` (my range preview) and `final.mp4`. Nothing tells me which of these is the deliverable besides the filename.
  - A way to list the source recogniser's confidence for the render itself. The check's `words` line says 44 of 44 heard, which confirms the cut but not the wording.

- **Not sure it's right:**
  - Sentence flow: "The first thing you need is an old laptop, then install Linux on it." The source had two sentences ("...old laptop." and "Then install Linux on it.") with a 0.30 s pause. The cut is clean, but the remaining join reads as one run-on sentence in the transcript. I kept it because the pause is clean and the brief was no false starts or asides, not a rewrite. A human should listen to the 0:06 join.
  - The earlier editor's `tighten --min-cut 0.4` and the cat-aside cut were inherited, not checked by me beyond reading the timeline and joins. `check` passing covers sync, words and loudness, not whether the aside cut was the right boundary.
  - `cutroom listen` was not run on any join; the joins were judged from the pause lengths and words only.

- **Best part:** `cutroom status` plus the `WARN[...]` line with its own `--apply` command made resuming fast: the one open problem was named, with its fix, and `--dry-run` confirmed the join before I applied it. `cutroom check` then gave a clear per-line PASS with the exact words checked.

- **One change:** have `cutroom transcript --file <render>` print one line per phrase with its pause (like the source transcript does), so the final sentence breaks can be read directly from the deliverable instead of being inferred from a merged paragraph.
