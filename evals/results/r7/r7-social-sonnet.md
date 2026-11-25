# Friction report

- **Outcome:** done. out/final.mp4 (13.77 s, burned captions, check PASS 0 fail/0 warn) and out/final.srt (captions --check ok) delivered.
- **Commands run:** 14: guide; guide --all; doctor; init talk.mp4; transcribe; transcript; transcript --suspects; edit tighten; edit cut (retake); edit cut (cat aside); timeline; timeline --joins; captions --burn; captions; render; check; cp captions.srt final.srt; captions --check out/final.srt; look --output 4.8; status.
- **Errors hit:** none.
- **Guessed wrong:** assumed `captions` could write to out/final.srt directly; it only writes out/captions.srt, so I copied it. Noticed the `--burn` and the SRT are separate invocations (had to run `captions --burn` and `captions`).
- **Confusing output:** `captions --burn` truncated groups ("4 more (--all)") - fine. The exit code of my combined command was 1 though everything passed (the final `cat` of a missing file), not a tool issue. Retake cut needed the "UM" in --from text; worked first time.
- **Missing:** an `--output path` flag on `captions` (to write out/final.srt); a single command that does burn + SRT.
- **Not sure it's right:** base.en transcript; burned caption only visually spot-checked at one frame (output 4.8). Source video is a test pattern so no face/lip check possible. Whether the SRT cues' times (offset 0.113 s start) are ideal I couldn't judge.
- **Best part:** `next:` lines, `retake-kept` warning with a ready cut command, and `check` covering captions, words and sync in one output.
- **One change:** let `captions` take `--output out/final.srt` (and write both burn and SRT in one call).
