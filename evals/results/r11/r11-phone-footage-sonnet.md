# Friction report

- **Outcome:** done. out/final.mp4 is 13.8 s with the cat aside, the retake and the long pauses removed. Rotation is baked in (no rotation tag), and `check` passes with 0 failures and 0 warnings.
- **Commands run:** 9 cutroom commands, plus 1 ffprobe. In order: guide; guide workflow; guide edit; init talk-rot.mov; inspect; transcribe; transcript; edit tighten; edit cut (cat aside); edit cut (UM retake); timeline --joins; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing. I expected to need a rotation flag, and none was needed because render handles it automatically.
- **Confusing output:** `edit tighten` printed its warnings (retake-kept, aside-kept) as ready-to-run cut commands. I ran my own equivalents instead, which was fine. I did not test how `--apply` compares.
- **Missing:** nothing important. I could not look at the frames myself, so I relied on check's picture line.
- **Not sure it's right:** I did not run `look` on the final frames. I trusted `check` ("upright, right moment").
- **Best part:** the `handled rotated` note in inspect, and check's explicit "display rotation 0° (no rotation tag)" line that answers the task requirement directly.
- **One change:** have `init` or `inspect` say, in the rotation line, that render bakes the rotation in automatically. I wasn't sure at that point and only learned it from the render output.
