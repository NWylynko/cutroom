# Friction report

- **Outcome:** done — `out/final.mp4` passes `cutroom check` with 0 warnings and the transcript reads as intended (pickup line in place, retake, aside and long pauses gone), but I could not listen to the pickup join, so the sound match is unverified by ear.

- **Commands run:** 13 shell calls containing about 36 cutroom invocations, in order:
  `cutroom guide`; `cutroom guide workflow`, `edit`, `perceive`; `cutroom guide render`, `verify`, `pitfalls`; `cutroom doctor`; `cutroom init talk.mp4 pickup.mp4`; `cutroom transcribe`; `cutroom transcript`; `cutroom transcript talk`, `talk --suspects`, `pickup --suspects`; `cutroom edit tighten`; `cutroom timeline`; `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --source talk --dry-run` (and the same for "By the way, my cat"); the two real cuts; `cutroom edit add --help`; `cutroom timeline`; `cutroom edit add --source pickup --from "Then install" --to "Linux on it." --before s4 --dry-run`; the real add; `cutroom edit cut --from "Then install" --to "Linux on it." --source talk`; `cutroom timeline`; `cutroom timeline --joins`; `cutroom render --preview --dry-run`; `cutroom render --preview`; `cutroom check out/preview.mp4`; `cutroom render`; `cutroom check out/final.mp4`; `cutroom transcript --file out/final.mp4`; `cutroom look --cuts --file out/final.mp4` (then opened the PNG).

- **Errors hit:** none. No command failed, so no `fix:` line was needed.

- **Guessed wrong:** Nothing failed, but I was unsure whether `edit add` would also remove the original line. It does not, so a second `edit cut --source talk` was needed. Replacing a line took three steps (add, then cut, then check the order), and I had to work out the ordering myself.

- **Confusing output:**
  - The `audio-mixed` WARN appeared on every timeline and joins call after the add. It stays after the render passes, and nothing in the output says whether it is harmless for a re-record on the same camera. The text says "otherwise this is fine", which I had to take as a judgement call.
  - `timeline` kept listing `retake-kept` as "still open" after the retake was cut, until I had removed the earlier take. That was a stale warning from before the cut, so I had to reason about which warnings were current.
  - `tighten` reports pauses in source time, while `timeline` reports output time. I had to convert between them by hand.

- **Missing:**
  - A single replace command, such as `cutroom edit replace --from "Then install" --to "Linux on it." --source pickup --over --from ... --to ...` (or `edit add --replace <segment>`), which would have made the swap one step.
  - A way to acknowledge `audio-mixed` for a re-record from the same camera, so `timeline` and the later output stay clean.
  - A `listen` or audio comparison between the two takes of the same line, to confirm they match in tone and level before render. I had no way to hear it.

- **Not sure it's right:**
  - The pickup line is spliced from `pickup.mp4` with the camera's own sound. The render's loudness is normalised to -16 LUFS, but I did not check that the pickup's tone or room sound matches the talk. I did not run `cutroom listen` on the join at output 6.4 s.
  - The source frames are test-pattern video, so I could only check picture continuity at cuts, not framing or faces.
  - The transcript was produced by `base.en`. The final transcript matches the expected 44 words, but a bigger model could have caught a misheard word the same way in both the expected and the render.
  - The `--cuts` image shows all four cuts land on clean frames.

- **Best part:** The sequence `transcript` → `transcript --suspects` → `edit tighten` → `timeline` → `edit cut --dry-run` → `check` was clear and quick. The dry-run showed the join and pause before anything was applied, and `check` printed every measured result with its limit.

- **One change:** Add a single `edit replace` command (or `edit add --replace <segment>`) that takes a span from another take and swaps it for the segment it names, with the original cut and the new one placed in the same step. This removes the add/cut/ordering dance and the manual check that the new line sits where the old one was.
