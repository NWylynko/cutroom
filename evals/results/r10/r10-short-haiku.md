# Friction report: r10-short-haiku

- **Outcome:** done. `out/short.mp4` (1080x1920, 7.267 s) passes `cutroom check` with 0 failures and 0 warnings, and a re-transcription reads "The first thing you need is an old laptop, then install Linux on it, finally, open the terminal and run the update command". One caveat: the source frames are a colour-bar test pattern with no visible person (see Not sure it's right).

- **Commands run:** 14 shell calls (plus image reads of the PNGs they wrote):
  1. `cutroom guide`, `ls -la`
  2. `cutroom guide shorts`, `guide workflow`, `guide captions`
  3. `cutroom guide perceive`, `guide overlays`, `guide render`
  4. `cutroom guide pitfalls`, `guide verify`, `cutroom doctor`
  5. `cutroom init talk.mp4`, `cutroom transcribe`, `cutroom transcript talk`
  6. `cutroom transcript talk --suspects`, `--words --from 9.5 --to 27`, `--search "update command"`, `--search "old laptop"`
  7. `cutroom short --from "The first thing you need is an" --to "update command." --name steps`
  8. `cutroom edit cut --from "By the way," --to "the keyboard." --project shorts/steps`, `cutroom timeline --project shorts/steps`, `timeline --joins --project shorts/steps`
  9. `cutroom look --output 0:01 --project shorts/steps`, `look --output 0:04.5 ...`, `look 11 --project shorts/steps`, `look --count 6` (failed)
  10. `cutroom captions --burn --project shorts/steps`, `cutroom overlay list --project shorts/steps`
  11. `cutroom look --output 0.35 ...`, `look --output 3.6 ...`, `cutroom captions --burn --dry-run ...`
  12. `cutroom render --preview --project shorts/steps`, `cutroom check shorts/steps/out/preview.mp4`
  13. `cutroom render out/short.mp4 --project shorts/steps`, `cutroom check out/short.mp4`
  14. `cutroom transcript --file out/short.mp4`, `look out/short.mp4 3.6`, `ffprobe`

- **Errors hit:**
  - `error[bad-args]: look needs a time` / `fix: give a time or range (cutroom look 12.5, cutroom look 6-12), or use --cuts`. My mistake (ran `look --count 6` with no time). The fix line got me unstuck immediately.
  - No other errors.

- **Guessed wrong:**
  - I assumed `cutroom look` with no time could show a contact sheet of the whole project. It needs a time or range.
  - Nothing else; the commands and flags I guessed from the guides worked.

- **Confusing output:**
  - The burn-in caption groups are described as "1–3 words (≤ 20 chars)" but the output contains 4-word groups: "is an old laptop", "the terminal and run". Either the guide is wrong or the grouping rule is not applied as written.
  - `short` printed the 720p-to-1080x1920 enlargement warning ("enlarged 2.7× ... will look soft") and then the aside warning. The aside warning's fix line was good, but I had no clear guidance on whether 720x1280 is the better trade for a 720p source. I kept 1080x1920 without measuring softness.
  - `look` labels read `out 0:01.000 = talk 11.20`; the mapping from output to source time was clear but easy to misread at a glance.
  - The `captions --burn` output lists groups but not which overlay ID they became until the following `overlay list` call. Fine, but I had to run two commands to see it.

- **Missing:**
  - A way to tell if the picture contains a person or a face. `look` showed colour bars with a black box and no speaker. Nothing in `check` or `inspect` flagged that, and the safe-zone and anchor advice assumes a speaker is there.
  - A way to listen to the final render's audio without writing a separate `listen` call per range.
  - A `captions --burn` option to set the maximum words per group, so the 1–3 word rule could be enforced or changed.

- **Not sure it's right:**
  - The picture is probably wrong for a human viewer: the frames I looked at are a synthetic colour-bar test pattern, not a person speaking. `check` passes because it compares the render to the source frames, not to what a viewer would expect. I did not confirm the source contents beyond the frames at 0:11 and the crop at 0:01, 0:03.6 and 0:04.5.
  - I never listened to the audio. The words are confirmed by base.en re-transcription only, which the tool itself says confirms the cut, not the recognition.
  - The 1080x1920 output from a 1280x720 source is enlarged 2.7×; `check` does not measure sharpness, so I have no evidence on softness.
  - The caption at 0:03.6 sits at about y 1370–1475 of 1920, inside the stated safe zone (260–1490) but close to its bottom edge.

- **Best part:** `cutroom short --from ... --to ... --name steps` built the whole vertical project in one call, with pauses tightened and the aside flagged as a WARN with the exact `edit cut` command to run. The render-then-`check` loop gave a full, readable PASS report, and `transcript --file` on the final render gave an independent word check.

- **One change:** have `cutroom look` (or `inspect`) report when the frames have no face or person, something like `WARN no face detected in 0:11.000 frame; the crop and anchor cannot be judged`. A short built on a non-speaker source would then fail at the first look instead of after a full render.
