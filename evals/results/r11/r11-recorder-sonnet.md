# Friction report

- **Outcome:** done. out/final.mp4 uses recorder audio, synced (worst 6 ms), long pauses tightened; check PASS with 0 failures and 0 warnings.
- **Commands run:** 8 (plus `guide`, `guide workflow`, `guide edit`): guide; guide workflow; guide edit; init talk.mp4 --source mic=recorder.wav; sync mic; edit audio mic; edit tighten (failed); transcribe; edit tighten; timeline; render; check.
- **Errors hit:** `error[no-transcript]: no transcripts` / `fix: run cutroom transcribe first`. The fix worked at once. I chained commands, so I should have read the workflow order more carefully.
- **Guessed wrong:** nothing significant.
- **Confusing output:** the timeline lists retake-kept and aside-kept WARNs that I left alone because the task only asked to remove long pauses. It is unclear whether "check PASS" is meant to coexist with open WARNs. The fix commands printed are very long.
- **Missing:** nothing. A one-line "recorder workflow does not need transcribing before tighten" hint in the sync section would help.
- **Not sure it's right:** the retake ("The first thing you need is, UM...") and the cat aside remain in the video. I deliberately did not cut them, as they were not requested. Transcription used the small base.en model.
- **Best part:** the recorder workflow in the guide is three commands, and check measures sync twice (against the recorder and against the camera sound).
- **One change:** have `edit tighten` say up front that it needs a transcript, or auto-run transcribe.
