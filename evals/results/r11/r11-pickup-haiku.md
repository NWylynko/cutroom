# Friction report: 11-pickup-haiku

- **Outcome:** done. `out/final.mp4` (13.8 s, 1280x720) passes `cutroom check` with 0 failures and 0 warnings; the transcript reads "...The first thing you need is an old laptop, then install Linux on it. Finally..." with the pickup line in place and the false start, aside and long pauses gone. Not verified by ear: I did not `cutroom listen` the join at 4.90-7.90 that the pickup INFO line suggested.

- **Commands run:** 27 cutroom invocations in 10 shell calls, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide edit`, `cutroom guide verify`
  3. `cutroom doctor`, `cutroom init talk.mp4 pickup.mp4`, `cutroom status`
  4. `cutroom transcribe pickup`, `cutroom transcribe talk`, `cutroom transcript talk`, `cutroom transcript pickup`
  5. `cutroom transcript talk --suspects`, `cutroom transcript pickup --suspects`, `cutroom edit tighten --dry-run`, `cutroom edit tighten`, `cutroom timeline`
  6. `cutroom timeline --apply retake-kept@4.09`, `cutroom edit cut --from "By the way," --to "the keyboard." --source talk`, `cutroom edit replace --text "Then install Linux on it." --with pickup --dry-run`
  7. `cutroom edit replace --text "Then install Linux on it." --with pickup`, `cutroom timeline`, `cutroom timeline --joins`
  8. `cutroom render --preview`, `cutroom check out/preview.mp4`
  9. `cutroom render`, `cutroom check out/final.mp4`
  10. `cutroom transcript --file out/final.mp4`
  (Also one failed attempt to update a task list through a plan tool, unrelated to cutroom.)

- **Errors hit:** none from cutroom. No `error[code]:` lines and no non-zero exits.

- **Guessed wrong:** nothing significant. The guide's "Several takes" section named `edit replace --text "…" --with pickup` exactly, and it worked on the first try. The pickup line was the same words as the talk line, so `--text` matched without a `--source` hint.

- **Confusing output:**
  - `timeline` after `tighten` printed the `retake-kept` and `aside-kept` WARNs with long suggested commands. Those commands were useful, but `--apply` then ran a different cut from the one I wanted to see. I had to re-read to confirm what the retake fix would remove (source 6.200–9.433, the "UM" take).
  - After the retake fix, `aside-kept` was still listed as open, and the next edit's output dropped it without saying so. I only confirmed it cleared by running `timeline` again.
  - `INFO[audio-mixed]` printed in three separate outputs (`edit replace --dry-run`, `edit replace`, and `timeline --joins`) with the same text. It is noise after the first time.
  - `check` printed `median -0 ms, worst -0 ms`. The `-0` is odd.
  - The `replace` dry-run said the pickup line would be `0:00.380–0:02.130`, while `timeline` later showed `0:00.367–0:02.133`. The difference is about 13 ms; it was harmless but I had to check that the two were the same segment.

- **Missing:**
  - A way to see the pickup and original words side by side before replacing, with their timings. I had to read two transcripts and match them by hand.
  - A confirmation that `--apply` changed only what the warning named, without a second `timeline` run.
  - The `--near` value in the suggested retake command was a source time (6.31), not an output time. The guide says `--near` takes a source time, but the warning text does not say so.

- **Not sure it's right:**
  - The false start cut and the aside cut were chosen from base.en transcripts. The aside is off-topic by its words ("my cat just walked across the keyboard") and I cut it; I did not listen to the joins.
  - The `words` check confirms the render matches the expected words as re-transcribed by base.en. It does not confirm that the pickup line is the one the speaker wanted, because the re-recorded line was only identified by my reading of the transcript.
  - `sync` passed at 0 ms for each segment, but `check` does not say whether the pickup's own level and room tone sound like the talk. The INFO line claimed they match (-16.6 vs -16.5 dB), and I took that at face value.

- **Best part:** `timeline` followed by `timeline --joins` shows each cut with its neighbouring words and pause, so I could check the whole edit in two commands. `edit replace --dry-run` then showed exactly which talk span would go and which pickup span would replace it, before I changed anything.

- **One change:** have `edit replace` (and the `--apply` path for `retake-kept`) print a before/after table of the transcript words on each side of the swapped span, with the source times for both takes, so I could confirm the right take was chosen without a separate transcript read.
