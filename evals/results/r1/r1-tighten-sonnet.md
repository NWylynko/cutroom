# Friction report
- **Outcome:** done. out/final.mp4 is 13.767 s (from 30.261 s) and its re-transcript reads as one clean explanation.
- **Commands run:** 10 cutroom calls: guide, guide --all, doctor, init talk.mp4, transcribe, transcript, transcript --suspects, edit tighten, edit cut (retake), edit cut (cat aside), timeline, timeline --joins, render, check (failed). Plus init/transcribe/transcript on the render for verification (in a scratch dir, deleted).
- **Errors hit:** `error[unknown-command]: unknown command "check"` / `fix: run cutroom to list commands`. The guide says "Never call an edit done without running cutroom check", but it isn't implemented. The fix line was no help. I had to verify by hand.
- **Guessed wrong:** assumed `cutroom check` exists (guide mandates it).
- **Confusing output:** guide contradicts itself (workflow says "not available in this version yet", the rules section demands it). Nothing else.
- **Missing:** `check` command. Also a way to transcribe a render without creating a second project (I used a scratch project dir).
- **Not sure it's right:** verification was only a base.en re-transcript of the render plus duration match; no frame or join listening (did not run `listen`). The 0:16.97 cough sat in a pause that tighten/cut removed; I did not confirm that by listening.
- **Best part:** the transcript phrase view with [repeat] suspect, quoted-text cuts, and the timeline --joins view made the edit trivial.
- **One change:** implement `cutroom check` (render vs timeline words/duration/joins) so verification is one command.
