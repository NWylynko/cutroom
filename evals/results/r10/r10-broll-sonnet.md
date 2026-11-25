# Friction report
- **Outcome:** done. out/final.mp4 is 13.77 s; `cutroom check` passed with 0 failures and 0 warnings.
- **Commands run:** 10 in all.
  1. guide
  2. guide workflow / edit / overlays
  3. init talk.mp4 terminal.mp4
  4. transcribe talk
  5. transcript --suspects
  6. transcript
  7. edit tighten
  8. edit cut (retake)
  9. edit cut (aside)
  10. overlay add clip
  11. timeline --joins
  12. look --output 10.15
  13. render
  14. check
  15. timeline
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** nothing serious. The retake text in the transcript is "The first thing you need is, UM, the first thing." Quoting it as `--from "The first thing you need is, um" --to "the first thing."` worked.
- **Missing:** nothing.
- **Not sure it's right:** I did not open the look PNG, and I did not hear the audio. I relied on the `check` lines (overlay frames, sync and words all ok). The B-roll starts at "open the terminal", about 0.7 s after "Finally," begins. The `check` words line used base.en, so it only confirms the cut, not the recogniser.
- **Best part:** `transcript --suspects`, `edit tighten`, and the `overlay add clip --at/--to` word anchors. `check` verifies the overlay window frame by frame.
- **One change:** have `overlay add` print the lead-in words before the anchor, or offer to look at its frame automatically.
