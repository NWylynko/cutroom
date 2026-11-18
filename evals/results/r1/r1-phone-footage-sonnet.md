# Friction report
- **Outcome:** done. out/final.mp4 is 13.767 s (was 30.261 s), cat remark cut, pauses tightened, upright pixels, no rotation metadata. I also cut the "UM, the first thing" false start, which the task did not ask for.
- **Commands run:** 8 cutroom + 3 ffprobe/ffmpeg/Read: guide; guide --all; init; transcribe; transcript; edit tighten; edit cut (cat); timeline --joins; edit cut (retake); render; check (failed); look 2 (source frame); ffprobe; ffmpeg frame extract of final.mp4.
- **Errors hit:** `error[unknown-command]: unknown command "check"` / `fix: run cutroom to list commands`. Not helpful: the guide's rule 1 says to always run `cutroom check`, while the workflow section says "not available in this version yet". Dead end; no fix possible.
- **Guessed wrong:** assumed `cutroom check` exists (guide rule 1 demands it).
- **Confusing output:** render's next: line says "`cutroom check` is coming". The guide's rules contradict it. `cutroom look` only shows source frames, so it cannot verify the render's orientation.
- **Missing:** a way to verify the render (check, or `look` on out/final.mp4, plus rotation/side-data report). I verified orientation via ffprobe (no rotation tags) and ffmpeg frame extract plus viewing, outside cutroom.
- **Not sure it's right:** audio sync and cut joins in the render were not verified by any tool. The test-pattern digit looked upright, but I couldn't compare against the source frame the way a check would.
- **Best part:** text-anchored cuts (--from/--to quoting), the transcript with pauses and suspects, and the automatic rotation handling made it a 5-minute job. Render was 2 s.
- **One change:** ship `cutroom check` (or let `look` take the rendered file), and remove the guide's mandatory-check rule until it exists.
