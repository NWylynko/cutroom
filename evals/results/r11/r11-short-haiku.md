# Friction report: run 11-short-haiku

- **Outcome:** done. `out/short.mp4` (1080x1920, 7.267 s, 3 segments, captions burned in) passes `cutroom check` with 0 failures and 0 warnings. Caveat: the source picture looked like a colour test pattern, not a speaker, so framing could not be judged by eye.

- **Commands run:** 27 cutroom invocations, plus an `ls` and two image reads. In order: `guide`, `guide workflow`, `guide shorts`, `guide edit`, `guide captions`, `doctor`, `init talk.mp4`, `inspect`, `transcribe`, `transcript`, `transcript --suspects`, `short --from "The first thing you need is an" --to "update command." --name steps`, `timeline --project shorts/steps`, `timeline --joins --project shorts/steps`, `edit cut --from "By the way," --to "the keyboard." --project shorts/steps --dry-run`, `edit cut ... --project shorts/steps`, `timeline --project`, `timeline --joins --project`, `captions --burn --project shorts/steps`, `render --preview --project shorts/steps`, `look --output 0.35 --project shorts/steps`, `look --output 3.6 --project shorts/steps`, `overlay ls --project shorts/steps` (failed), `render out/short.mp4 --project shorts/steps`, `check out/short.mp4 --project shorts/steps`, `transcript --file out/short.mp4`, `status --project shorts/steps`.

- **Errors hit:**
  - `error[unknown-op]: no overlay op "ls"` with `fix: did you mean cutroom overlay list?` The fix line was correct and clear. I did not re-run it, since I did not need the overlay list.

- **Guessed wrong:**
  - I guessed `overlay ls`. The real verb is `list`, and the error named it.
  - I assumed the shorts project inherited the run folder's transcript. It did, with no re-transcribe needed, and that was right.

- **Confusing output:**
  - The `short` warning "the 1280x720 source is enlarged 2.7x to fill 1080x1920 ... --size 720x1280 halves the enlargement" does not say what the trade-off is. A 720x1280 output still enlarges 1.5x, and the guide does not say whether a smaller short is preferable to a 2.7x soft one.
  - `look` frames are labelled only by time and the crop anchor. With a test-pattern source there was no way to see whether the anchor 0.50,0.50 lands on anything meaningful.
  - `overlay` docs in `captions --burn` output say "lower thirds move up above them", but there were no lower thirds here, so I could not confirm that behaviour.

- **Missing:**
  - A whole-source contact sheet, so I could see what the picture contains before choosing a crop anchor. `look` only takes one frame at a time.
  - A way to listen to the joins from the short command without a separate `cutroom listen` step. I did not run `listen` on the two joins (0:02.367 and 0:04.000), so audio seams rest only on the `--joins` text and the `check` results.

- **Not sure it's right:**
  - Framing: the picture is a test pattern, so I cannot confirm the speaker is in the crop. `check` passes `picture` by matching frames to the source, which is not the same as the speaker being in frame.
  - Audio seams after the cat aside was cut: I relied on `timeline` reporting "no problems found" and `check` passing `sync`, `words` and `stray-sound`. I did not listen.
  - Transcription used `base.en`, which the tool itself warns can mishear words. `check`'s word pass uses the same model, so it confirms the cut, not the wording. The on-screen words match what I read in the transcript, but I did not verify against a larger model.
  - Upscale: the picture is 2.7x enlarged and will look soft. I left the default 1080x1920 rather than test 720x1280.

- **Best part:** `short --from/--to` plus the `aside-kept` warning, which printed the exact `edit cut --from ... --to ...` to remove the cat aside. Running it with `--dry-run` first showed the join and the new length, so I applied it with no surprises. `check` then reported every category with a clear verdict.

- **One change:** have `short` take `--without "phrase"` (or a `--cut` list) so asides can be dropped at creation time, with the join shown in the same output. That would have saved the separate timeline round-trip, and the warning already prints the exact command.
