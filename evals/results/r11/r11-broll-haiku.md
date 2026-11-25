# Friction report: 11-broll-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s, 1280x720) renders and passes `cutroom check` with 0 failures and 0 warnings. The retake, the cat aside and the terminal cutaway are all in place, and the rendered transcript reads cleanly. I did not listen to the audio myself.

- **Commands run:** about 33 `cutroom` invocations, plus one image Read and a few `ls`/`cd` shell calls. In order: `cutroom guide`; `guide workflow`, `guide edit`, `guide verify`; `guide overlays`, `guide render`, `guide pitfalls`; `cutroom doctor`; `cutroom init talk.mp4`; `cutroom inspect`; `cutroom transcribe`; `cutroom transcript`; `transcript --suspects`; `transcript --search "terminal"`; `cutroom listen --help`; `cutroom edit tighten`; `edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`; `edit cut --from "By the way" --to "keyboard." --dry-run`; `cutroom timeline`; `edit cut` (retake, applied); `edit cut --from "By the way," --to "the keyboard."` (aside, applied); `timeline --joins`; `timeline`; `overlay add clip terminal.mp4 --at "open the terminal" --to "update command." --dry-run`; `overlay add clip ...` (applied); `overlay list`; `render --preview --dry-run`; `render --preview`; `check out/preview.mp4`; `look --overlays`; `render`; `check out/final.mp4`; `transcript --file out/final.mp4`.

- **Errors hit:** none. No command failed, and no `fix:` line was needed.

- **Guessed wrong:** nothing failed, but I had to work out that `--from` must match the transcript's exact words, including the comma after "UM". I guessed that from the guide's advice to quote the transcript exactly and it worked on the first try. I also checked `cutroom listen --help` for the flags and it matched my expectations.

- **Confusing output:**
  - The `timeline` WARN lines are long. Each one bundles the code, the reason, the suggested command, and an alternative `--accept` command on one line, so they are hard to scan. The `retake-kept` suggestion also uses `--near 6.31` as a source time next to `--from` text, which is easy to misread.
  - The `retake-kept` WARN suggested `--from "The first thing you need" --near 6.31 --to "first thing."`. I used the transcript's own quoted words instead. Both would work, but the suggestion does not say which is preferred.
  - `overlay add ... --dry-run` prints `added o1:` even though nothing was added. That reads like a real change and only the trailing `DRY RUN` line says otherwise.
  - `cutroom look --overlays` writes a PNG and says "view the image". The frame it showed is labelled `"run"` and looks like a fractal. I could not tell from that whether the clip is a terminal.

- **Missing:**
  - I could not find a way to look at a frame of the B-roll clip itself (`terminal.mp4`) before it is an overlay. I only saw it through `look --overlays`, so I have not confirmed the cutaway shows a terminal.
  - I wanted a way to see how long the clip is compared with the window it fills. `overlay add` shows the window (2.3 s) but not the clip's length or whether it was trimmed.
  - Nothing says whether `cutroom check` compares the clip's content with the intended content, or only that the overlay is visible at the right frames.

- **Not sure it's right:**
  - I did not listen to the audio. The 0.33 s sound at source 16.97 was removed by `tighten` (it said so), but I did not confirm it was a bump rather than a word.
  - The transcripts come from `base.en`, so a misheard word could hide a problem. `check` confirmed the 44 expected words in order, but it uses the same small model, so it confirms the cuts rather than the words.
  - I cannot confirm from the tool that the full-frame terminal cutaway shows the right thing. The check passed on timing and visibility only.
  - The retake cut keeps "The first thing you need is an old laptop." and drops the stumbled take. The aside cut removes the cat sentence. I judged both from the transcript text, not by listening.

- **Best part:** `cutroom guide` gave a clear sequence. `tighten`, then `timeline` with its WARN lines, then `cut`, then `overlay add clip --at/--to`, then `render --preview`, then `check`. The `--dry-run` options, the `timeline` warnings that name the next command, and `check` passing with per-item lines made each step easy to confirm.

- **One change:** let `cutroom look` take a clip file (for example `cutroom look terminal.mp4 --at 2.0`), or have `overlay add clip` print the clip's length next to the window. That would have let me confirm the B-roll shows a terminal before the final render, without guessing from one overlay frame.
