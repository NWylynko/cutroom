# Friction report

- **Outcome:** done. out/final.mp4 rendered (13.8 s, 1280x720, SDR bt709, upright) and `cutroom check` passed with 0 failures and 0 warnings.
- **Commands run:** 9 cutroom commands, in order: `guide`; `guide workflow` + `guide pitfalls`; `init clip.mov`; `inspect`; `transcribe`; `transcript`; `transcript --suspects`; `guide edit`; `edit tighten`; `edit cut --from "The first thing you need is, UM" --to "the first thing."`; `edit cut --from "By the way" --to "keyboard."`; `timeline --joins`; `render --help`; `render`; `check`. (The list is 15 invocations. `init` already inspected, so my extra `inspect` was redundant.)
- **Errors hit:** none.
- **Guessed wrong:** nothing. I assumed HDR and rotation needed flags, but both are automatic.
- **Confusing output:** `tighten` printed a `retake-kept` WARN before my cut. That is expected, and the guide says so. The cut message "cut clip 0:06.200–0:09.433" shows times that differ from the transcript's, because of the pads. That is fine but made me double-check.
- **Missing:** nothing important. I would like a built-in way to view a frame of the final next to the source. `look` exists, but I did not need it.
- **Not sure it's right:** I never looked at an actual frame of the final. I relied on check's `picture` and `tone` lines (upright, saturation 87 vs 87). I did not listen to the joins. The cut at "the first thing." relies on the transcript.
- **Best part:** `init` flagged HDR and rotation as `handled`. `render` printed what it did about them. `check` proved upright and tone match numerically.
- **One change:** have `edit tighten` print the suggested retake cut command inline as `next:`, or run the retake/aside cuts for you with `--auto`, so there is one fewer step.
