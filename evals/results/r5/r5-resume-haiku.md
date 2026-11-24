# Friction report (r5-resume-haiku)

- **Outcome:** done — `out/final.mp4` (13.767 s, 1280x720) renders and passes `cutroom check` with 0 failures and 0 warnings; the false start "The first thing you need is, UM, the first thing." is removed and the "By the way, my cat..." aside (already cut by the earlier editor) stays out.

- **Commands run:** 17 cutroom calls, plus shell helpers (ls, cat of cutroom.json and timeline.json, find).
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom status`
  4. `cutroom timeline`
  5. `cutroom timeline --joins`
  6. `cutroom transcript`
  7. `cutroom transcript --suspects`
  8. `cutroom edit cut --from "the first thing you need is" --near 6.31 --to "first thing." --dry-run`
  9. `cutroom edit --help`
  10. `cutroom edit cut --from "the first thing you need is" --near 6.31 --to "first thing." --reason "..."`
  11. `cutroom timeline --joins`
  12. `cutroom status`
  13. `cutroom render --help`
  14. `cutroom render --preview`
  15. `cutroom check out/preview.mp4`
  16. `cutroom render`
  17. `cutroom check out/final.mp4`

- **Errors hit:** none. No command failed. The only problem was the timeline warning `WARN[retake-kept] at 0:04.093`, and its `fix:` line (the suggested `edit cut --from ... --near 6.31 --to "first thing."`) worked as printed.

- **Guessed wrong:** nothing failed, but I had to confirm that `--near` and `--reason` exist on `edit cut` (they do; `edit --help` lists them). I did not need `--occurrence`.

- **Confusing output:**
  - The retake warning says `at 0:04.093`, which is an output time (the start of segment s2). The fix line gives source times (6.31, 0:06.310–0:09.220). Nothing tells you which clock each number is on, so I had to work out that they referred to different timelines.
  - `cutroom transcript` reports source times, while `cutroom timeline` and `timeline --joins` report output times. Mixing them is easy to get wrong.
  - `cutroom status` after the first edits said the preview was "STALE ... not checked", which was clear. Its "next:" lines were useful.

- **Missing:**
  - A way to read what the rendered file actually says. `cutroom check out/final.mp4` says "all 44 expected words heard in order", but it does not print the expected words or a transcript of the output, so I cannot confirm from the check output itself that the false start is gone or that every sentence is complete. I had to trust that the expected-word list came from the timeline.
  - A tool to listen to the joins. I did not run `cutroom listen` on the joins because the check's sync and word checks passed. I could not hear them myself.

- **Not sure it's right:**
  - I accepted the earlier editor's `edit tighten --min-cut 0.4` and their cut of the cat aside without re-listening or re-transcribing the rendered output. Pauses now sit at 0.30 to 0.33 s, which equals the pads, so they look consistent with the design.
  - The render's word check passed, but I could not verify sentence completeness or the absence of audible artefacts at joins without listening to them.
  - Ending: "Finally, open the terminal and run the update command." followed by "And that's it, your server is ready." I kept these as they were. I did not judge whether the content is correct, only that the cut is clean.

- **Best part:** The warning printed the exact fix command with the source range, and `--dry-run` showed the join and the duration change before anything changed. `cutroom check` then reported every category, so the verify step was clear.

- **One change:** Have `cutroom check` (or a new `cutroom transcript <render>`) print the words it heard in the output file, so you can confirm the final text reads as intended, not just that the expected words appeared.
