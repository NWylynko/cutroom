# Friction report (fill this in as the last thing you do)

- **Outcome:** done. `out/final.mp4` (1280x720 h264, 13.767 s, matches the timeline) has the cat remark and the retake removed, pauses tightened, and no rotation metadata (ffprobe shows none). Not verified end to end: `cutroom check` does not exist in this version, so the checks were ffprobe, a contact sheet and one spectrogram.

- **Commands run:** 13 Bash calls (plus one failed `update_plan` call that was not a cutroom command, see Errors).
  1. `cutroom guide`
  2. `cutroom guide workflow` and `cutroom guide perceive`
  3. `ls -la`, then `cutroom guide edit`, `cutroom guide render`, `cutroom guide pitfalls`
  4. `cutroom doctor`, `cutroom init talk-rot.mov`, `cutroom inspect`
  5. `cutroom transcribe`, `cutroom transcript`, `cutroom transcript --suspects`
  6. `cutroom edit tighten`, `cutroom edit cut --from "By the way" --to "keyboard" --dry-run`, `cutroom edit cut --from "is, UM" --to "the first thing" --dry-run`
  7. `cutroom edit cut --from "By the way" --to "keyboard"`, `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`
  8. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing."`, `cutroom timeline`, `cutroom timeline --joins`
  9. `cutroom render --preview --dry-run`, `cutroom render --preview`
  10. `cutroom look --help`, `ls out/`
  11. `cutroom render --dry-run`, `cutroom render`, `ffprobe ... out/final.mp4` (stream and duration)
  12. `ffmpeg ... fps=1,scale=480:-1,tile=4x4 ... work/final_sheet.png`
  13. `cutroom listen 4.0-10.6`

- **Errors hit:**
  - The `update_plan` tool returned `PreToolUse ... hook error: Not shown: you are a subagent...`. It is not a cutroom error, and it was not a problem to ignore.
  - No cutroom command failed with an `error[code]`. Nothing needed a `fix:` line.
  - Warnings: `WARN the cut starts inside continuous speech right after "need"` on the retake dry run. The `cutroom listen` I ran afterward showed the take boundary was fine, and the joins after applying the cut showed no problems.
  - After the cat cut: `WARN[retake-kept] ... "the first thing you need is" is still said 2× in the timeline`. This was expected. It cleared after the retake cut, and `timeline` reported "no problems found".

- **Guessed wrong:**
  - I assumed `cutroom look` could show frames of the rendered `out/final.mp4`. It takes source times only ("one frame as a viewer sees it" of the source). So I used ffmpeg on the output to check frames, which is outside cutroom.
  - I assumed `cutroom check` existed because the guide's rule 1 says to run it. The workflow guide marks it "not available in this version yet", and the rule still says to run it, which contradicts itself.
  - `--from "is, UM"` was not enough to make a unique match in my first dry-run. It cut the right span only by luck, because `--to "the first thing"` also matched. I needed the longer phrase `"The first thing you need is, UM"` to be sure.

- **Confusing output:**
  - `cutroom transcript` tags the retake as `[repeat]` on both takes, and the 2-suspect list pointed at the sound at 0:16.970 inside the cat remark. That sound was removed with the cat cut, and nothing in the output says a suspect is inside a span you are already removing.
  - `cutroom edit cut --dry-run` output shows the join as `pause 0.30 s` but gives no indication whether it is in continuous speech, except through the separate WARN line.
  - The render line "13.767 s, 413 frames (timeline says 13.767 s)" was clear and useful.

- **Missing:**
  - Frame checks of the output timeline. `cutroom look` only takes source times, so checking output frames means leaving cutroom.
  - `cutroom check` (the guide tells me to run it before calling any edit done).
  - An explanation of what the source actually shows. The source is a synthetic colour-bar test pattern with a seconds counter, not a talking head. `cutroom inspect` and `cutroom look` do not flag that, so I could not judge faces or framing.
  - A way to verify rotation of the output with cutroom. I had to infer it from ffprobe plus the contact sheet.

- **Not sure it's right:**
  - The cut boundaries are right by transcript and timeline, but I did not listen to every join. I listened only to the retake join's spectrogram region (4.0–10.6 s source).
  - The transcript is from the `base.en` model, which the tool itself warns about ("expect misheard words"). The cat-cut and retake selection depend on that transcript, and I could not verify the words against a larger model.
  - I confirmed no rotation metadata in the output with ffprobe, and the contact sheet shows upright counter digits. I did not check the source's pixel orientation independently, so "upright" is judged from the digits, not from a known reference.
  - `render --preview` reports `960x540`, while the render guide says "540p". Both are fine, but the mismatch made me pause.

- **Best part:** The workflow is short and matches the task: `init`, `transcribe`, `transcript --suspects`, `edit tighten`, `edit cut --from/--to` with `--dry-run`, `timeline --joins`, then `render`. The `--dry-run` on every cut and the `timeline`/`--joins` output made each edit easy to check before I applied it. The render also printed the duration against the timeline, which I trusted.

- **One change:** Add `cutroom check` (or make `cutroom look` accept `out/final.mp4` times). Right now there is no cutroom way to look at the rendered output, and the guide's rule 1 says to run a command that doesn't exist yet.
