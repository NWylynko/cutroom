# Friction report: r13-recorder-haiku

- **Outcome:** done — `out/final.mp4` (13.767 s, 1280x720) uses the recorder's audio, aligned to the camera with a +1.700 s offset and -42.8 ppm drift, with the long pauses tightened and two cuts applied; `cutroom check` passed with 0 warnings.

- **Commands run:** about 23 cutroom invocations across 10 shell calls, in order: `cutroom guide`; `cutroom guide workflow`, `guide edit`, `guide verify`; `cutroom doctor`; `cutroom init talk.mp4 --source mic=recorder.wav`; `cutroom inspect`; `cutroom sync mic`; `cutroom sync --help`; `cutroom edit audio mic`; `cutroom transcribe`; `cutroom edit tighten`; `cutroom timeline`; `cutroom edit cut ... --dry-run` (x2); `cutroom edit cut ...` (x2); `cutroom timeline`; `cutroom timeline --joins`; `cutroom render --help`; `cutroom render --preview --range 0:00-0:08`; `cutroom render`; `cutroom check`.

- **Errors hit:** none. Every command exited 0 on the first try.

- **Guessed wrong:** nothing material. I did not have to guess a flag. The only thing I took on trust was that `--near 6.31` in the suggested `retake-kept` command was a source time (it was, and the dry run confirmed it).

- **Confusing output:**
  - `sync` says `mic started recording 1.700 s before talk` and then `offset +1.700 s`. The sign convention is only clear after reading the mapping line (`mic time = +1.700 s + talk time × ...`). A one-line statement of which direction to shift would help.
  - The `timeline` warnings list both the original warning and a suggested command with a `--near` time. That is useful, but the `--near` value is a source time, and nothing in the warning says so.
  - `transcribe` printed `talk 64 words` and `mic 64 words`, with 3 suspects for talk and 4 for mic. I could not tell which suspects matter for a recorder-audio render, since the timeline now plays mic audio.

- **Missing:**
  - A way to see the sync result against the sound at specific points (for example, a per-segment offset table). `check` shows the median and worst case, which is enough for a pass, but I could not inspect the per-segment numbers before rendering.
  - A way to tell whether the transcript used for `retake-kept` came from talk or mic. The timeline output says `talk` for the source, but the mic transcript had its own word list and I did not know which one the warnings were using.

- **Not sure it's right:**
  - I did not run `cutroom listen` on any join, and I did not re-transcribe with a larger model. The transcript is base.en, which the tool itself flags as expecting misheard words. The `words` check passing only proves the render matches the base.en transcript.
  - The "whole video in sync" requirement is backed by `sync` (residual 0.02 ms over 54 windows, 1 rejected) and `check` (sync median -0 ms, worst -6 ms over 5 segments). I did not confirm sync at the very end of the 30 s recording, where the drift accumulates. The `sync-camera` check also passed, which covers that.
  - I did not listen to the output, so I cannot say whether the cat-aside cut and the false-start cut sound natural. The dry-run join texts looked right.

- **Best part:** the separate-recorder workflow is short and works. `init --source mic=...`, then `sync mic`, then `edit audio mic`, then `render` and `check` is four commands, and `check` reports `sync-camera` independently of `sync`'s own offset. The `--dry-run` on each cut showed the join and the new timeline length before I applied it.

- **One change:** have `sync` print the shift direction in one plain sentence (for example, "shift mic 1.700 s later to line it up with talk") and print the per-segment sync residuals, so I could confirm alignment before rendering rather than only after `check`.
