# Friction report: r6-hdr-phone-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s, 1280x720) passes `cutroom check` with 0 failures and 0 warnings. The retake, the cat aside and the pauses are gone, and the file is bt709 SDR with no rotation metadata.

- **Commands run:** about 33 `cutroom` invocations, plus `ffprobe` and `ls`. In order:
  `guide`, `guide workflow`, `guide edit`, `guide render`, `guide perceive`, `guide verify`, `guide pitfalls`, `doctor`, `init clip.mov`, `inspect`, `transcribe`, `transcript`, `transcript --suspects`, `edit tighten`, `edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`, `edit cut --from "By the way, my cat" --to "across the keyboard." --dry-run`, `edit cut` (same two, applied), `timeline`, `timeline --joins`, `render --preview --dry-run`, `render --preview`, `check out/preview.mp4`, `look --file out/preview.mp4 3`, `look --file out/preview.mp4 1`, `render --dry-run`, `render`, `check`, `look --file out/final.mp4 0:06` (failed, see below), `look --file out/final.mp4 0:01`, `look --output 0:03-0:06 --count 4`, `listen --output 3-5` (failed, see below).

- **Errors hit:**
  - `(eval):1: command not found: cutroom` on `look --file out/final.mp4 0:06` and `listen --output 3-5`. This was my mistake: I dropped the `export PATH=...` in that shell call. It printed no `fix:` line, which is fine for a PATH problem, but a shell that loses PATH between calls is easy to hit.
  - No cutroom errors otherwise. The one warning, `WARN[retake-kept]` after `tighten`, came with a ready-made `edit cut --near 6.31` command. I didn't use it: my text-quoted `--from "The first thing you need is, UM"` cut did the same job, and the warning cleared.

- **Guessed wrong:** Nothing was wrong. `--count` on `look --output` and `--dry-run` on `render --preview` both worked as I guessed. The guide's "Cuts" section (`--from` must be unique, `--to` is the first match after) matched the behaviour.

- **Confusing output:**
  - `tighten` reports pauses in source time, but `timeline` shows output time. I had to keep the two straight to read the join list.
  - `check` prints `picture: clip's 180° display rotation is baked into the pixels` on render, and `format` only checks that the file has rotation 0°. Neither tells me the pixels are upright. I had to confirm that by looking at the frames myself.
  - `check` `tone` prints saturation and brightness "87 vs source 87". That reads as a match, but I can't tell from it whether the file is washed out or correct.
  - The guide is long. I read all seven topics, which is a lot of tokens for a task this size. `guide workflow` and `guide edit` were enough.

- **Missing:**
  - A direct way to confirm orientation. Something like `check --upright` or a flag that reports whether the rotated pixels match the displayed orientation would have saved me from eyeballing a test-pattern frame.
  - A way to confirm the HDR-to-SDR conversion looks right, beyond `tone` numbers. `check` says the colour matches the source, but a measured comparison to a reference SDR conversion would help.
  - `timeline` said "no problems found" after the cuts, but it didn't show the retake warning had been resolved by my cut. I only knew because the warning disappeared.

- **Not sure it's right:**
  - Orientation: the frames show a seven-segment counter that reads left to right in the right direction, and the output frame matches the source timecode (out 4.000 = clip 10.23, which reads "10"). I'm fairly confident the picture is upright. I haven't checked it in an actual YouTube-style player.
  - HDR: `tone` says the file is SDR bt709 and the colours roughly match the source's saturation. I didn't compare against a reference conversion, so I can't say whether the highlights are right.
  - Word count: `check words` reports 44 expected words, which equals 64 source words minus the 20 I cut. That matches, but it is my own arithmetic, not a tool-reported figure.
  - The `sound` suspect at 0:16.97 was removed by `tighten`, and I didn't `listen` to it first. The render has no stray sound, so I think it was a breath or room noise.

- **Best part:** `transcribe` then `transcript --suspects` gave me the retake, the aside and the stray sound in one listing, with the right source times. Then `cut --from/--to --dry-run` showed each join before I applied it. That workflow was smooth, and the render and `check` results were clear.

- **One change:** Add an orientation and HDR check to `cutroom check` that compares the render with the source as a viewer sees it. It would report, for example, "displayed upright: yes (matches source rotated 180°)" and "SDR conversion: highlights within X% of reference." That is the one thing I couldn't verify from the tool's output and had to judge by eye.
