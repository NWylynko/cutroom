# Friction report

- **Outcome:** done. `out/final.mp4` (1280x720, 13.767 s, 413 frames) passes `cutroom check` with 0 warnings; the cat remark and the retake are cut, pauses are tightened to 0.30-0.33 s joins, and the picture is upright in the pixels with no rotation tag. Verified by comparing `cutroom look` frames of source and final, and ffprobe showing no rotate side data.

- **Commands run:** about 12 Bash calls with roughly 30 cutroom invocations, in order: `ls -la && cutroom guide`; `guide workflow`, `guide perceive`, `guide edit`; `guide render`, `guide verify`, `guide pitfalls`; `doctor`; `init talk-rot.mov`; `inspect`; `transcribe`; `transcript`; `transcript --suspects`; `edit tighten`; `edit cut --from "By the way" --to "keyboard" --dry-run`; `edit cut --from "is, UM" --to "the first thing." --dry-run`; `edit cut --from "The first thing you need is" --near 6.3 --to "the first thing." --dry-run`; `listen 6.0-9.6`; `edit cut ...` (applied, retake); `edit cut --from "By the way" --to "keyboard."` (applied); `listen 16.5-17.6`; `timeline`; `timeline --joins`; `render --preview --dry-run`; `render --preview`; `check`; `look 3 --file out/preview.mp4`; `render --dry-run`; `render`; `check out/final.mp4`; `look 3 --file out/final.mp4`; `ffprobe` on final; `look 3.0`. Plus two image views.

- **Errors hit:** none from cutroom. The only non-zero exit was an update_plan hook error (not a cutroom issue), which I ignored. No `fix:` line was needed.

- **Guessed wrong:** I assumed `--from "is, UM"` would be the natural anchor for the retake. The dry-run showed a join warning ("cut starts inside continuous speech right after 'need'"). Switching to `--from "The first thing you need is" --near 6.3` fixed it, with a 0.30 s pause at the join. I did not know `--near` could disambiguate a phrase that repeats until I tried it. The `edit` guide mentions `--near` only in the error path.

- **Confusing output:** `timeline` said `no problems found` after I cut the retake, while the earlier `tighten` had printed `WARN[retake-kept]` for the repeated phrase. That is correct, but it means the warning cleared without a visible message. The `WARN[retake-kept]` line also points at `transcript --search` rather than giving the cut command that would keep the second take. The `look` label "during the word 'to'" for 3.0 s was clear. Output lengths were fine.

- **Missing:**
  - A way to see stream-level metadata of the rendered file (rotation side data, codec tags) in cutroom. I used ffprobe. `check`'s "display rotation 0°" line covers it, but I wanted the raw side data too.
  - A `--cuts` look at the final render. I ran `look --cuts` only in the guide, not on this edit, so the join frames were not inspected visually.
  - A suggested cut command for a `repeat` suspect, e.g. `edit cut --from ... --near ... --to ...` that keeps the last take.

- **Not sure it's right:**
  - Transcription used `base.en`. The retake and the cat sentence were heard correctly in isolated listens, but I did not check whether the small model missed any other off-topic phrase or any word inside the tightened pauses.
  - The stray sound at 0:16.97 fell in the removed span (15.67-22.83 in source), so it is not in the output. I confirmed this from the timeline, not by listening to the final render around that point.
  - The picture check compared 15 sampled frames. Orientation was verified visually on one frame, against a source frame rendered with rotation applied.
  - I did not listen to the join at 0:06.40 or 0:08.13 in the final render.

- **Best part:** `transcript --suspects` plus `cut --dry-run` made the retake and the cat remark easy to find, check and apply, and the `check` output lists each measured property with its numbers.

- **One change:** when `tighten` or `timeline` reports `retake-kept` for a repeated phrase, print a ready-to-run cut command that keeps the last take (for example `cutroom edit cut --from "The first thing you need is" --near 6.3 --to "the first thing."`). That would have saved two dry-runs and a listen.
