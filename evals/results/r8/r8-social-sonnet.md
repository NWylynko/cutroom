# Friction report

- **Outcome:** done. out/final.mp4 (13.8 s, burned captions) and out/final.srt written; `check` PASS 0 failures 0 warnings; viewed a frame with the caption visible.
- **Commands run:** 11 substantive: guide, guide --all, doctor, init talk.mp4, transcribe, transcript, transcript --suspects, edit tighten, edit cut (retake), edit cut (aside), timeline, timeline --joins, captions --burn --output out/final.srt, render, check, captions --check out/final.srt, look --output 0.9 --file out/final.mp4.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant. I cut the retake with my own quoted span rather than the suggested warning command; the result was clean.
- **Confusing output:** the retake-kept WARN printed during tighten was long; the later cut resolved it silently (good). `look --output` on a path needed `--file`; I had to pick that from the guide.
- **Missing:** nothing blocking. No whole-run "verify all" that also checks that the SRT matches the burned groups (burned groups are 14, SRT cues are 5 — different segmentation, expected but not stated).
- **Not sure it's right:** transcript is base.en; I could not independently verify that the first take "The first thing you need is an old laptop" is the better delivery beyond it being the last. Only one burned-caption frame viewed by me; check sampled 6.
- **Best part:** `captions --burn --output` produced both deliverables in one call, and `check` covering captions, words and sync made verification easy.
- **One change:** have `check` also accept/validate the .srt next to the video in the same run.
