# Friction report: r6-resume-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s, 1280x720) renders and passes `cutroom check` with 0 failures and 0 warnings. The retake was removed and the aside was already cut by the earlier session. I did not listen to the joins with `cutroom listen`, so the audio at the cut points is only checked by the automated checks.

- **Commands run:** 8 shell calls with about 21 `cutroom` invocations, in this order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide edit`
  3. `cutroom status`, `cutroom edit history`, `cutroom timeline`, `cutroom transcript`
  4. `cutroom transcript talk --suspects`, `cutroom guide perceive`, `cutroom timeline --joins`
  5. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`, then the same without `--dry-run`, then `cutroom timeline`, `cutroom timeline --joins`
  6. `cutroom guide render`, `cutroom render --preview --range 0:00-0:08`, `cutroom render --preview`
  7. `cutroom check out/preview.mp4`, `cutroom guide verify`
  8. `cutroom render`, `cutroom check out/final.mp4`, `cutroom transcript --file out/final.mp4`

- **Errors hit:** none. No command failed, so no `fix:` line was needed.

- **Guessed wrong:** nothing significant. The `--near 6.31` value came straight from the `retake-kept` warning, so I did not have to work it out.

- **Confusing output:**
  - `cutroom render --preview --range 0:00-0:08` rendered the whole 13.767 s timeline into `out/preview-range.mp4`, not 0 to 8 s. The explanation is in the output, but only after a line that says the range "renders the whole chunk(s) covering it". I read it twice before accepting it.
  - `cutroom transcript --file out/final.mp4` shows "an old laptop, then install Linux on it." The recogniser joined two sentences with a comma. The source has a 0.30 s pause between them, and the transcript has no punctuation to show a sentence boundary. I could not tell from the tool output whether the render actually has two sentences.
  - `cutroom edit history` shows only the op and time. The "before" snapshots are in `.cutroom/history/*.json`, and the human output does not mention them.

- **Missing:**
  - A way to check sentence boundaries in the render, such as the pause length between two phrases, not just the words. The `--joins` output gives this for the timeline, but it does not say whether a join is a sentence boundary.
  - A quick way to see the state of a half-finished job: `status` said "1 open warning" and `timeline` listed the open retake. Together they were enough, but `status` could name which edits the earlier session made and which warnings remain.

- **Not sure it's right:**
  - The retake cut was checked by the `retake-kept` warning clearing and by `check`'s word test passing. I did not confirm the join sounds natural with `cutroom listen`. The join at 0:06.4 is between "is an old laptop." and "Then install Linux on it." at a 0.30 s pause.
  - The earlier session's cut of "By the way, my cat just walked across the keyboard." and the tighten at `--min-cut 0.4` were accepted as correct without re-checking them against the full transcript.
  - The "keeps every sentence complete" requirement was judged from the transcript text, not heard.

- **Best part:** `cutroom timeline` printed the exact fix command for the retake, with `--dry-run` shown on the next line. Running the dry run first, then the real cut, then `timeline --joins` took three calls and left no warnings.

- **One change:** have `render --range` either cut to the requested range or refuse with a clear message, instead of quietly rendering the whole timeline into a file named `-range.mp4`.
