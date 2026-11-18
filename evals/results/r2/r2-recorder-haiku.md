# Friction report (fill this in as the last thing you do)

- **Outcome:** done. `out/final.mp4` (13.767 s, 1280x720) uses the recorder's audio, synced to the camera (sync residual 0.02 ms over 37 windows), with the long pauses tightened, one retake and one aside cut, and `cutroom check` passes on both the preview and the final render.

- **Commands run:** 14 Bash calls plus 1 Read of a PNG (and 1 blocked `update_plan` call, see Errors).
  1. `cutroom guide`
  2. `cutroom guide workflow; guide edit; guide render; guide verify; guide pitfalls; guide perceive` (one shell call)
  3. `ls -la` and `cutroom doctor`
  4. `cutroom init talk.mp4 --source mic=recorder.wav` then `cutroom inspect`
  5. `cutroom sync mic`
  6. `cutroom edit audio mic` then `cutroom transcribe`
  7. `cutroom transcript talk` then `cutroom transcript talk --suspects`
  8. `cutroom edit tighten --dry-run` then `cutroom edit tighten`
  9. `cutroom edit cut --from "The first thing you need is, UM" --to "UM, the first thing" --dry-run` and the same for `--from "By the way, my cat" --to "across the keyboard"`
  10. Both cuts applied for real, then `cutroom timeline` and `cutroom timeline --joins`
  11. `cutroom render --preview --dry-run` then `cutroom render --preview`
  12. `cutroom check out/preview.mp4`
  13. `cutroom render` then `cutroom check out/final.mp4`
  14. `cutroom look --cuts --file out/final.mp4`

- **Errors hit:**
  - The `update_plan` tool call was rejected by a PreToolUse hook: "you are a subagent, and the app shows only the main conversation's task list... Do not call update_plan or wrap_up again". Not a cutroom error, and harmless; I skipped it after that.
  - No cutroom command failed. Nothing needed a `fix:` line.

- **Guessed wrong:**
  - `cutroom doctor` ends with `next: cutroom init --source cam=<video file>`, which does not match the guide's syntax (`cutroom init <video> --source mic=<file>`). The guide form worked, so I used it. The doctor suggestion looks stale or wrong.
  - I expected `sync` to need the mic source to be transcribed first. It did not; `sync` worked straight after `init`.
  - I assumed `edit audio mic` would also transcribe the mic. It did not, and `transcribe` then ran over both sources. That was fine, but the guide's "Separate audio recorder" section does not mention transcribing.

- **Confusing output:**
  - `cutroom sync` printed `mic time = +1.700 s + talk time × (1 -42.7 ppm)`. The sign convention is only clear if you already know the direction; a one-line "recorder sample = camera time + offset" would help.
  - `cutroom check`'s `sync` line says "audio vs source median -0 ms". I could not tell which reference "source" means (the camera's own audio, or the aligned mic). The result is correct, but the wording is unclear.
  - `timeline` printed `talk 0:00.467–0:04.400` (source ranges) next to output times. That is documented, but the mix of two timelines on one line was something I had to re-read.
  - The `WARN[retake-kept]` from `tighten` was expected but is printed on every later command (`cut --dry-run`, `tighten`) until the retake is cut. It stays noisy after the fix. (It was gone by the `timeline` call.)

- **Missing:**
  - A way to listen to the final audio around a join via `cutroom listen --output`. I did not do it because `check` passed, but I wanted it.
  - Something in `cutroom check` that states which audio it compared against for the separate-recorder case.

- **Not sure it's right:**
  - I did not listen to the recorder against the camera by ear. The `sync` fit (residual 0.02 ms, 1 window rejected) and `check` sync (-0 ms median, -6 ms worst) are my only evidence the audio lines up.
  - The stray sound at 0:16.97 (source) sits inside the cat aside I cut (18.5–21.93). I assumed it is part of the aside, and `check`'s stray-sound test passes, but I did not `listen` to it.
  - The transcript used `base.en`, which the tool itself warns about. The `words` check compares against that same small-model transcript, so a misheard word could hide a problem.
  - The retake choice (keeping "The first thing you need is an old laptop") was read from the transcript, not verified on the picture.

- **Best part:** The separate-recorder flow is three commands (`init --source`, `sync`, `edit audio`), and the timeline, joins and check output are short and enough to act on. `cutroom check` gave a clear per-line PASS list.

- **One change:** Make `cutroom doctor` and `cutroom init --help` use the same syntax as the guide, and have `cutroom check` name the audio reference it compares the recorder against for separate-audio projects.
