# Friction report: r10-fix-edit-haiku

- **Outcome:** done — out/final.mp4 (13.767 s) passes `cutroom check` with 0 warnings and its transcript reads as the intended explanation; I did not listen to the audio, so the joins are checked by transcript and `timeline --joins` only.

- **Commands run:** about 33 `cutroom` calls, plus a few shell `ls`/`cat` reads. In order: `guide`; `guide workflow`; `guide edit`; `status`; `edit history`; `transcript`; `timeline`; `transcript talk --suspects`; `timeline --joins`; `edit undo --dry-run` (undid the edit, unintended); `edit undo` (undid tighten too, unintended); `timeline` (error); `edit --help`; `edit redo --help`; `edit tighten --help`; `edit tighten --min-cut 0.4`; `timeline`; `edit cut --from "The first thing you need is, UM" --to "first thing." --near 6.31 --dry-run`; `timeline | head -3`; `edit cut --from "The first thing you need is, UM" --to "first thing." --near 6.31`; `edit cut --from "By the way," --to "the keyboard." --dry-run`; `edit cut --from "By the way," --to "the keyboard."`; `timeline`; `timeline --joins`; `edit history`; `guide render`; `guide verify`; `render --preview`; `render --dry-run`; `render`; `check out/final.mp4`; `transcript --file out/final.mp4`; `status`.

- **Errors hit:**
  - `error[no-timeline]: no timeline yet in .` after my own `edit undo` removed the only edit. The `fix:` line ("make the first edit (e.g. `cutroom edit tighten`), which creates it") got me unstuck, but I had to re-type tighten's flags from `edit history`'s memory and the old `timeline.json`.
  - Exit code 2 on the batch above, which was just the no-timeline error.

- **Guessed wrong:**
  - I assumed `cutroom edit undo --dry-run` would preview the undo. It did not: it undid the last edit ("undid: edit cut talk 1.9-2.9"). The `--dry-run` line in `edit --help` ("Use --dry-run to preview without saving") appears to cover every `edit` subcommand, but `undo` ignores it.
  - I assumed there was a redo. There is none (`edit redo --help` printed the general edit help, not a redo command), so an undo is one-way unless you re-do the edit by hand.
  - I assumed `edit cut --dry-run` saves nothing. It does, and it printed "DRY RUN: nothing was changed" correctly. That assumption held.

- **Confusing output:**
  - The `edit undo --dry-run` output was "undid: edit cut ..." with no "dry run" marker. It looked like a preview, so I did not realise it had applied until the timeline changed.
  - After re-running tighten, warning times moved (retake-kept from 0:03.093 to 0:04.093, aside-kept from 0:10.527 to 0:11.527). The warnings' own suggested commands still worked, but I had to re-read them to confirm the source range matched.
  - `transcript --suspects` printed "(already cut)" on the sound at 0:16.970, which was useful but easy to miss in a long line.

- **Missing:**
  - A `redo`, or a way to see what undo would revert before running it.
  - A way to listen to a join from the CLI without an extra `listen` call. I did not use `cutroom listen`, so the audio join at 0:06.400 and 0:08.133 is unverified by ear.
  - A short summary after each edit of the full sentence list in output order, since I rebuilt that from `timeline` lines each time.

- **Not sure it's right:**
  - The render transcript shows "an old laptop," with a comma where the source had a period. That is most likely the recogniser's punctuation, not a real problem, but I could not confirm it by listening.
  - I cut the aside "By the way, my cat just walked across the keyboard." as off-topic. That is a judgement call on the content, and the tool only flagged it as a probable digression.
  - The 0.30 s pauses after tighten are the tool's default pads, not something I tuned. `check` passed and the words are all present, but I did not listen for whether the pacing feels natural.

- **Best part:** The `timeline` warnings each came with an exact suggested command, and `edit cut --dry-run` showed the join and the remaining open warnings before applying. Combined with `render` plus `check` plus `transcript --file`, it was easy to confirm the final file.

- **One change:** Make `edit undo --dry-run` preview without applying the undo (or add `edit redo`). The undo applying silently under `--dry-run` cost me the tighten edit and a re-run.
