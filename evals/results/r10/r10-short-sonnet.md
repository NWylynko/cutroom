# Friction report
- **Outcome:** done — out/short.mp4 (720x1280, 7.27 s) rendered; `check` PASS with 0 failures/warnings.
- **Commands run:** 11: guide; guide shorts + guide workflow; init talk.mp4; transcribe; transcript; short --from/--to --name steps; edit cut --from "By the way" --to "keyboard."; timeline; edit frame --size 720x1280; captions --burn; render out/short.mp4; check; look --output 1.5.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant. Note `short` kept the cat aside, so I cut it (the aside-kept WARN flagged it).
- **Confusing output:** `short` output says "talk 0:10.250" but the timeline starts at 10.200 (pads); minor. The `short` help example text differs from the actual span.
- **Missing:** nothing blocking. A one-shot `short --size` flag and `--captions` would save two commands.
- **Not sure it's right:** transcription is base.en; the video frame was a test pattern so I couldn't judge crop framing on a speaker. Captions only inspected at one frame.
- **Best part:** `check` covering captions, safe zones, words, sync and subtitles in one pass; guide shorts gave the exact recipe.
- **One change:** have `short` auto-cut flagged asides inside the span (or offer `--cut-asides`) and default size to 720x1280 for 720p sources.
