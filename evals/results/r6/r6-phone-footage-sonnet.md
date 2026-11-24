# Friction report

- **Outcome:** done. out/final.mp4 rendered, cat remark and long pauses cut, the retake removed, rotation baked in, and `check` passed with 0 failures and 0 warnings.
- **Commands run:** 9 (not counting `ls`): `guide`, `guide workflow`, `guide pitfalls`, `guide edit`, `init talk-rot.mov`, `inspect`, `transcribe`, `transcript`, `edit tighten`, `edit cut --from "By the way" --to "keyboard"`, `timeline`, `edit cut --from "The first thing you need" --near 6.31 --to "first thing."`, `render`, `check`.
- **Errors hit:** none.
- **Guessed wrong:** nothing. I did not view a frame with `cutroom look` to confirm the picture is upright. I relied on the render output line and `check` saying the rotation is baked in and the display rotation is 0°.
- **Confusing output:** `tighten` and `cut` print a lot of text back to back. Output after `tighten` showed a `retake-kept` WARN that had a ready-made fix, which was helpful.
- **Missing:** nothing important. A one-shot "render for any player" confirmation was effectively there in the render line.
- **Not sure it's right:** I never looked at a frame of the output to confirm the picture is upright (180° case). `check` says format display rotation 0° and the picture matches source frames, but "matches source frames" is measured through the rotation handling, so I trust it but did not see it. I also did not listen to the joins.
- **Best part:** the `retake-kept` warning handing me the exact `edit cut` command with `--near`, and the render line stating rotation is baked in with no metadata.
- **One change:** have `check` (or render) print one explicit "upright: yes" line backed by an image path I can look at, so rotation can be verified visually.
