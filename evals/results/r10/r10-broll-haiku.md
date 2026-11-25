# Friction report: r10-broll-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s) renders, passes `cutroom check` with 0 failures and 0 warnings, and the frame contact sheets show the terminal clip covering exactly the "open the terminal ... update command." window. Not checked by ear.

- **Commands run:** about 34 cutroom invocations, in order: `guide`, `guide workflow`, `guide perceive`, `guide edit`, `guide overlays`, `guide render`, `guide verify`, `guide pitfalls`, `doctor`, `init talk.mp4 terminal.mp4`, `inspect`, `transcribe`, `transcript talk`, `transcript talk --suspects`, `transcript talk --words`, `edit tighten`, `edit cut --from "is, UM" --to "first thing" --dry-run`, `edit cut --from "By the way" --to "keyboard" --dry-run`, `edit cut --from "The first thing you need" --near 6.31 --to "first thing." --source talk --dry-run`, `edit cut` (retake, applied), `edit cut --from "By the way," --to "the keyboard." --source talk` (applied), `timeline`, `timeline --joins`, `overlay add clip terminal.mp4 --at "open the terminal" --to "update command."`, `overlay list`, `timeline`, `render --preview --range 0:06-0:13`, `timeline --joins`, `render --dry-run`, `render`, `check`, `look --output 0:08-0:12 --count 8`, `look --cuts`, `transcript --file out/final.mp4`. Plus two image Reads of the PNGs.

- **Errors hit:** none. No command failed.

- **Guessed wrong:** Nothing failed, but my first retake cut guess (`--from "is, UM" --to "first thing"`) would have kept the false start's opening "The first thing you need" and joined it to the second take. The `--dry-run` showed the join "need | The first" and the WARN, so I dropped it. The working cut was the one the `retake-kept` warning printed (`--from "The first thing you need" --near 6.31 --to "first thing."`). That was the right suggestion, but I only knew it was right because I dry-ran it.

- **Confusing output:**
  - `edit cut --dry-run` prints `join:` and `would cut` but does not show the words that would stay on each side of the cut at a glance. I had to read the join line carefully.
  - `overlay add` reports the anchor as `talk 0:23.700`, while `overlay list` shows output `0:09.000`. Both are right, but the anchor word's start is not obvious from the output.
  - `look --output` writes the PNG into `.cutroom/look/` and prints only the path. Fine, but I had no way to tell which frames were which without opening it.
  - `guide` topics repeat the same workflow steps. Reading all eight topics was about 8k tokens, most of it not needed for this task.

- **Missing:**
  - A way to see the full set of words kept after `transcript --words` without truncation. It showed 40 lines and said "24 more", which I did not need, but it was unclear whether `--from/--to` filtered by source time or output time.
  - A way to check that a clip's own audio is muted. `check` passes sync, but nothing reports the clip's sound level in the render.

- **Not sure it's right:**
  - The terminal clip's sound is supposed to be muted. I trusted the guide, and `check` passed sync and loudness, but I did not listen to `out/final.mp4` around 9.0–11.3 s.
  - The render's word check uses base.en, the same model the transcript used, so it confirms the cut but not the recogniser. The transcript itself came from base.en and may have misheard words; I read the talk once and it looked right.
  - The overlay boundaries (9.000–11.300) match the anchor words by time, not by audio. The contact sheet shows the clip starting between the "Finally," and "open" frames, which looks right.

- **Best part:** `timeline` after `tighten` printed a ready `edit cut` for the retake and the aside, and `--dry-run` previewed each join. The whole loop was quick once the transcript was read.

- **One change:** Make `edit cut --dry-run` print the last few words before and first few words after each cut, next to the join, so I can judge the retake cut without re-reading the transcript.
