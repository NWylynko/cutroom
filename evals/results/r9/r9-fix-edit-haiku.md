# Friction report: r9-fix-edit-haiku

- **Outcome:** done. The edit was reset and rebuilt; `cutroom check out/final.mp4` passes all checks with 0 warnings, and the render's transcript reads as complete sentences with no retake.

- **Commands run:** about 22 cutroom invocations across 8 shell calls, in order: `cutroom guide`; `cutroom guide workflow`; `cutroom guide edit`; `cutroom status`; `cutroom timeline`; `cutroom timeline --joins`; `cutroom transcript`; `cutroom transcript --suspects`; `cutroom edit history`; `cutroom edit --help`; `cutroom edit cut --help`; `cutroom edit tighten --help`; `cutroom edit reset`; `cutroom transcript --help`; `cutroom edit cut --from "The first thing you need is, UM" --near 6.31 --to "first thing." --dry-run`; `cutroom edit cut --from "By the way" --to "keyboard." --dry-run`; `cutroom edit cut --from ... --near 6.31 --to ... --reason ...`; `cutroom edit cut --from "By the way" --to "keyboard." --reason ...`; `cutroom edit tighten`; `cutroom timeline`; `cutroom timeline --joins`; `cutroom render --help`; `cutroom render`; `cutroom check out/final.mp4`; `cutroom transcript --file out/final.mp4`.

- **Errors hit:** none. No command failed.

- **Guessed wrong:** Nothing failed, but I read `cutroom edit reset` as a full wipe. It is one, but the output's `timeline reset to the whole of talk` line plus a leftover `kept-sound` warning was the only signal, so I had to re-check the timeline afterwards.

- **Confusing output:**
  - After `edit cut --from "The first thing you need is, UM" --near 6.31 --to "first thing." --dry-run`, the reported range was `talk 0:04.400–0:10.167 (5.767 s removed)`. That range starts at the end of the previous sentence, not at the retake's first word, so it was not obvious what would actually be removed until I read the join. The range includes the pause, which is worth stating.
  - The suggested `--from "The first thing you need"` in the `retake-kept` warning is ambiguous (it matches twice). The suggested `--near 6.31` fixed that, but I had to notice it.
  - In `timeline` output, segment ids are not in playback order (`s4` at 0:06.400 comes before `s3` at 0:08.133). This is correct but reads like a mistake.
  - `cutroom status` after the first edit said `out/final.mp4 ... current, not checked`, and it still listed the old render as current after I reset the timeline. It only became stale once I re-rendered.
  - `timeline` lists `retake-kept` and `edge-in-word` warnings, but the `broken-sentence` warning (cut through "going to show you how") is only in the warnings list and was not in the suggested fixes. I worked out the fix myself by resetting.

- **Missing:**
  - A way to see what an edit would do to the sentence (the full words removed, not just the time range) in `--dry-run`. The dry-run line gives the time range and the join, but not the words the cut would drop from inside the pause.
  - A `cutroom timeline` view of the edit's intent, so I could see which cut created a `broken-sentence` without reading the history.
  - I did not look at frames (`cutroom look`) before rendering. Nothing in the timeline or check tells you whether the picture is right at each cut, so the visual check was left to `check`.

- **Not sure it's right:**
  - I cut the off-topic aside "By the way, my cat just walked across the keyboard." This is a judgment call (it is not a retake or false start, and the task said to deliver the explanation). If the aside is meant to stay, it needs to go back in.
  - I did not verify the picture at each join visually. `check` samples 15 frames and says they match, and the `sync` and `words` checks pass, but I did not run `cutroom look` or `cutroom listen` on the joins.
  - The `words` check uses the base.en model, which is the same model that made the transcript, so it confirms the cut but not the recogniser (the guide says so).

- **Best part:** `cutroom timeline --joins` plus `cutroom transcript --file out/final.mp4` made the result easy to judge: the join list shows the words either side of every cut, and reading the render's own transcript confirmed the sentences are whole. `check` printed a clear PASS line with per-check detail.

- **One change:** `cutroom edit cut --dry-run` should show the words it would remove and the sentence it would leave broken (for example, "would remove: 'going to show you how'"), not only the time range and the join. That would have caught the bad cut at 2.9 in the original edit before it was rendered.
