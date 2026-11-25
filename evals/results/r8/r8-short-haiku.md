# Friction report: r8-short-haiku

- **Outcome:** done. `out/short.mp4` (1080x1920, 7.267 s) is the steps from "The first thing you need is an old laptop" to "run the update command", with the retake and the cat aside cut and captions burned in; `cutroom check` passes on the render with 0 failures and 0 warnings.

- **Commands run:** 9 Bash calls, about 30 cutroom invocations, plus 2 image reads. In order: `cutroom guide`; `cutroom guide workflow|shorts|captions|perceive`; `cutroom doctor`; `cutroom init talk.mp4`; `cutroom inspect`; `cutroom transcribe`; `cutroom transcript`; `cutroom transcript --suspects`; `cutroom short --from "The first thing you need is an" --to "update command." --name steps`; `cutroom timeline --project shorts/steps`; `cutroom edit cut --project shorts/steps --from "By the way" --to "keyboard."`; `cutroom timeline`; `cutroom captions --burn --project shorts/steps`; `cutroom timeline --joins`; `cutroom render --preview`; `cutroom look --output 0:01` and `0:05`; `cutroom render --project shorts/steps`; `cutroom check --project shorts/steps`; `cutroom captions --check`; `cutroom status`; `cutroom check out/short.mp4` (failed); `cutroom transcript --file out/short.mp4`; `cutroom captions --check shorts/steps/out/captions.srt --project shorts/steps` (failed); `ffprobe`; `cutroom captions --check out/captions.srt --project shorts/steps`.

- **Errors hit:**
  - `error[no-file]: can't read shorts/steps/shorts/steps/out/captions.srt` with `fix: cutroom captions writes out/captions.srt`. The path I passed with `--project` got resolved relative to the project and doubled. The fix line did not mention this, so I had to guess the rule; passing the path relative to the project worked.
  - `error[no-timeline]: no timeline yet` with `fix: make the first edit (e.g. cutroom edit tighten), which creates it`, from `cutroom check out/short.mp4` run in the root project. The timeline existed in `shorts/steps`. The fix line is wrong for this case; the real fix was adding `--project shorts/steps`, which the message never hints at.

- **Guessed wrong:**
  - I assumed `cutroom check <file>` would find the project's timeline from the file path. It needs `--project`.
  - I assumed `--check <file>` took a path from the cwd when `--project` is given. It takes a path from the project root.
  - `cutroom guide shorts` says `--project` works on the short. It does, but in the same command you need the project-relative path for every file argument, and that is not stated anywhere.

- **Confusing output:**
  - `cutroom status` lists the final render as "current, not checked" right after `cutroom render`, and the preview the same way. I did not know whether that meant a check was needed or had failed. It passed when run.
  - The `look` label `out 0:01.000 = talk 11.20 +cc` is clear once you know it, but the `+cc` tag is not explained at the point it appears.
  - The warning "the 1280x720 source is enlarged 2.7x ... it will look soft" appears in the middle of `short` output, and `--size 720x1280` is offered as the fix. I did not act on it, and the output does not say whether that is a quality cost worth accepting for this source.

- **Missing:**
  - A way to pass the picture to `look` or `check` from a short project without the `--project` juggling.
  - A short-form check that the on-screen text stays inside the safe zones. `check` does this for the burned captions, but it does not tell me anything about the picture's content, and the fixture is colour bars, so I could not judge framing from the frames.
  - A `status` or `check` summary of which file is the deliverable, so I did not have to infer that `shorts/steps/out/final.mp4` was it.

- **Not sure it's right:**
  - The transcript is `base.en`, so the words in the check are the recogniser's own words. The `words` check confirms the render matches the cut, not that the words are right.
  - The join at "install Linux on it. | Finally" has a 0.2 s pause. `timeline --joins` reports no problems, but I did not run `cutroom listen` on it.
  - The picture is a colour-bar fixture, so I could not confirm the speaker is framed well. The caption placement (about y 1280 of 1920) is inside the safe zone per `check`, but I did not see a speaker to judge against.
  - I judged 7.3 s as the right length for "just the steps". That is my call, not something the tool checked.

- **Best part:** `cutroom short --from ... --to ... --name steps` made the vertical project with safe zones and short-form pads in one command, and `cutroom check` then verified sync, words, stray sound, loudness and captions against the render with a clear PASS line.

- **One change:** Make file arguments resolve the same way with and without `--project` (or print the resolved path in the error), and have `no-timeline` name the project it looked in and suggest `--project` when a timeline exists elsewhere.
