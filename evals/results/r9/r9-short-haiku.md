# Friction report: r9-short-haiku

- **Outcome:** done. `out/short.mp4` (1080x1920, 7.267 s) passes `cutroom check` with 0 failures and 0 warnings, and it has the three steps, the cat aside cut, and burned-in captions. I could not confirm the speaker is framed, because the source is a colour test pattern with no person in it.

- **Commands run:** about 24 `cutroom` invocations, plus `ls`, `mkdir -p out`, and 3 image reads. In order: `cutroom guide`; `guide workflow`, `guide shorts`, `guide captions`; `doctor`; `init talk.mp4`; `inspect`; `transcribe`; `transcript`; `transcript --suspects`; `short --help`; `short --from ... --to ... --name steps`; `timeline --project shorts/steps`; `edit cut --text "By the way..." --project shorts/steps`; `timeline --joins`; `look 0:12 --project` (error); `look 0:11`; `look --output 0:03`; `look 0:11 --width 720`; `captions --burn --project shorts/steps`; `render --help`; `render --preview --range 0:00-0:04`; `look shorts/steps/out/preview-range.mp4 0-4` (error); `render --preview --project shorts/steps`; `look --file ... 0-7 --project`; `check shorts/steps/out/preview.mp4`; `render out/short.mp4 --project shorts/steps`; `check out/short.mp4`; `transcript --file out/short.mp4`; `look --file out/short.mp4 0-7` (error); `check out/short.mp4`; `timeline --project shorts/steps --json`; `look --file out/short.mp4 3.0` (error); `look --file out/short.mp4 3.0 --project shorts/steps`.

- **Errors hit:**
  - `error[out-of-range]: output 0:12.000 is past the end` / `fix: the edit is 7.267 s long`. The fix line was enough.
  - `error[frame-failed]: frame at 3.983 of shorts/steps/out/preview-range.mp4: no frame at 3.983 s (past the end?)` / `fix: check the time is inside the file`. This was not enough. The file is 4.000 s, so the end of a range is evidently sampled one frame past the last frame. I had to guess to use `--project` with a file range instead.
  - `error[no-timeline]: no timeline yet in .` / `fix: ... or work on a short: --project shorts/steps`. The fix line worked, but only after two wasted calls. `look --file out/short.mp4` run from the run folder silently used the root project, and nothing said that `--file` does not pick a project.

- **Guessed wrong:**
  - I assumed `look --file <render>` would find the short's project from the file path. It does not.
  - I assumed a range end is inclusive of the last frame. The error above shows it is not.
  - I expected `short` to leave its output at `out/` in the run folder. Its output lives in `shorts/steps/out/`. The `render out/short.mp4 --project shorts/steps` form is what puts it in the place the task asked for.
  - `render` also wrote `out/short.mp4.cutroom.json`, a sidecar I had not seen documented. I left it in place.

- **Confusing output:**
  - `short` printed `WARN the 1280x720 source is enlarged 2.7× to fill 1080x1920: it will look soft`. The guide says `--size 720x1280` halves the enlargement, but the task asks for 9:16 and I left the default. It is unclear whether 1080x1920 is still the right choice for a 720p source, and nothing tells you what the soft output would look like.
  - `transcript --suspects` flagged a 0.32 s sound at 16.97 s, "a cough, bump, breath, or words the recogniser skipped — `cutroom listen` it". That sound sits in the pause that the cut removed, so it no longer matters. I did not run `listen` to confirm what it was.
  - `timeline` lists segments as `s1`, `s2`, `s4`, with `s3` removed by the cut. That reads fine, but the gap in numbering is easy to misread as a missing segment.

- **Missing:**
  - A way to pick the project from `look --file`, or a note in the error saying `--file` ignores the project.
  - A way to set the end of a `--range` so it does not sample past the last frame.
  - Some sign of what the framing looks like on a real speaker. `look` and `check` confirm the crop and the safe zones, but not that the speaker is in frame.
  - A note that `captions --burn` also writes `shorts/steps/out/captions.srt`, not `out/short.srt` beside the video.

- **Not sure it's right:**
  - The words are right according to `base.en`, which is the same model that produced the transcript. `check` says "all 23 expected words heard in order", but it re-transcribed with that model, so it confirms the cut and not the recogniser. The transcript reads cleanly, but I could not hear the audio to verify it.
  - The crop is centred at 0.5,0.5 because the test pattern shows nothing to anchor on. If the real speaker is off-centre, the short will crop them out. I did not check this against a real speaker.
  - The `safe-zones` check passed. I checked the overlay boxes against the numbers it printed (y 260–1490, x ≤ 960) and they agree.
  - I did not run `cutroom render` at a second quality level or with `--loudness` changes. The default loudness gave a gain of +10.7 dB on a quiet source, which I accepted.

- **Best part:** `cutroom short --from ... --to ...` with the repeated "The first thing you need is" resolved to the right take automatically, and `timeline --joins` plus `check` gave a clear pass/fail on the cut itself. The `check` output was specific enough that I trusted it.

- **One change:** Make `look --file` and `transcript --file` take the project from the path when it is inside a shorts folder, or print the project it used. That would have saved three wasted calls and made the render-check loop work the first time.
