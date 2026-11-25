# Friction report: r8-broll-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s) renders from a timeline with the retake and the cat aside removed, the pauses tightened, and `terminal.mp4` as full-frame B-roll over "open the terminal ... update command."; `cutroom check` passes with 0 warnings, and the final transcript reads correctly.

- **Commands run:** about 32 `cutroom` invocations, in this order: `guide`; `guide workflow`, `guide edit`, `guide overlays`; `guide render`, `guide verify`, `guide perceive`, `guide pitfalls`; `doctor`; `init talk.mp4 terminal.mp4`; `inspect`; `transcribe`; `transcript talk`; `transcript talk --suspects`; `transcript talk --words --from 0 --to 31`; `edit tighten`; `edit cut --from "is, UM" --to "first thing." --dry-run`; `edit cut --from "By the way" --to "keyboard." --dry-run`; `edit cut --from "The first thing you need" --near 6.31 --to "first thing." --source talk --dry-run`; `edit cut 6.2-9.3 --dry-run`; `timeline`; `edit cut --from "The first thing you need" --near 6.31 --to "first thing." --source talk`; `edit cut --from "By the way" --to "keyboard."`; `timeline`; `timeline --joins`; `overlay add clip terminal.mp4 --at "open the terminal" --to "update command."`; `overlay list`; `timeline`; `look --output 0:09`; `render --preview --dry-run`; `render --preview`; `check`; `look --file out/preview.mp4 --output 8.5-11.8 --count 6`; `listen --file out/preview.mp4 --output 8-12`; `render`; `check`; `transcript --file out/final.mp4`.

- **Errors hit:** None. No command failed and no `fix:` line was needed.

- **Guessed wrong:** Nothing failed, but the retake WARN text was slightly off. It said the suggested cut removes "source 0:06.310–0:09.220", but the dry-run and the real cut covered 6.200–9.433 (pads included). The only way to see the real range was to run the suggested command with `--dry-run`. I also first tried `--from "is, UM"`, which left "The first thing you need" in place. That was my mistake, not the tool's, but the dry-run output showed it clearly.

- **Confusing output:** The `retake-kept` warning's suggested command and the range it describes don't match exactly (see above). `timeline` prints segment starts in output time and source ranges in a separate column, so I had to keep the two straight by hand. `overlay add` printed "appears on ... (talk 0:23.700)", which is useful, but the output time of the overlay (9.000) is only in the next `overlay list` line.

- **Missing:**
  - No explicit check that a clip's own audio is muted in the render. `check` verifies that the render's sound matches the talk's sound (sync), and I relied on `guide overlays` for "Clips are muted". A `check` line such as "clip audio: muted" would have settled it.
  - No way to make `talk`'s retake choice from the tool itself. I chose the second take by reading the transcript, which is expected.
  - `timeline --joins` with `listen` on each join was not something I did. I read the joins and the pause lengths (0.30–0.33 s) and trusted `check`'s "words" result.

- **Not sure it's right:**
  - The transcript used the base.en model, which heard "Terminal" and "Update" with capitals. The `check` "words" result also used base.en, so it confirms the cut, not the recogniser. I did not run a larger model.
  - I did not listen to the three remaining joins, so the judgement that they sound natural rests on the pause lengths and the transcript.
  - I did not view the final render's frames, only the preview's. The final was rendered from the same timeline and passed the same checks, and the overlay checks reported the same frame range.

- **Best part:** The `timeline` and `timeline --joins` pair, followed by `check`, showed each segment, each join, and each overlay's frame range, so verifying the edit took only a few reads.

- **One change:** Have `check` report the audio of each overlay clip explicitly (for example "clip o1: audio muted, voice from talk") and make the retake warning's suggested range match what its `--dry-run` shows.
