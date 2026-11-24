# Friction report

- **Outcome:** done. out/final.mp4 is 13.77 s with the cat aside, the retake and the long pauses removed. cutroom check passed with 0 failures and 0 warnings, and the picture is upright with no rotation metadata.
- **Commands run:** 11 cutroom calls, plus 1 ffprobe: guide; guide --all; init talk-rot.mov; transcribe; transcript; transcript --suspects; edit tighten; edit cut (cat); edit cut (retake); timeline; timeline --joins; render --preview --range 0-5; render; check; look out/final.mp4 2.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant. I guessed `init` needed no flags for a single file, and that was right.
- **Confusing output:**
  - `--preview --range 0-5` rendered the whole 13.8 s chunk. It says so, but "test small first" didn't save any time here.
  - `edit tighten` printed a `retake-kept` WARN whose suggested fix used `--near 6.31 --to "first thing."`. I used my own `--from/--to` instead, and it worked.
  - The `guide --all` output was long, and I only used the tail.
- **Missing:** an explicit statement in the render output about which inspect facts (HDR, VFR) were handled. Only rotation was mentioned, which was fine here.
- **Not sure it's right:** I couldn't judge the audio by ear. I relied on check's "words" and "stray-sound" results. I viewed one frame and the digit was upright, which confirms the rotation was applied, but I didn't view every cut.
- **Best part:** `transcript` with its `[repeat]` flag and pause annotations, plus text-anchored cuts. Rotation was handled automatically, and check verified the display rotation was 0.
- **One change:** have `tighten` print the exact `edit cut` commands for the retake and for any aside, or let `--range` render only the requested slice when the timeline fits in one chunk.
