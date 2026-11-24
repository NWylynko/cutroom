# Friction report

- **Outcome:** done. out/final.mp4 (13.8 s) has the pickup line in place of the original, with pauses tightened and the false start and cat aside cut; `cutroom check` PASS, 0 warnings.
- **Commands run:** 12. guide; guide --all; init talk.mp4 pickup.mp4; transcribe; transcript; transcript --suspects; transcript talk; transcript talk --suspects; edit tighten + edit cut (false start) + edit cut (aside) + timeline; edit cut --help; edit add --source pickup ... --after s4 + edit cut --text (original line) + timeline + timeline --joins; render + check.
- **Errors hit:** none.
- **Guessed wrong:** `transcript` with several sources showed only the first (pickup), so I needed `transcript talk`. Not an error, but the first call looked like the whole answer.
- **Confusing output:** `guide` only prints the topic index, so I had to run `guide --all` (long). The `audio-mixed` warning shows for the pickup, which has no separate recorder; the text says it is fine in that case, but it stays listed as "still open".
- **Missing:** a one-step "replace this span with that take" command. I had to add the pickup and then cut the original. I also couldn't hear whether the pickup matches the room tone and level of talk. The render's loudness normalisation helps, but I didn't audit it.
- **Not sure it's right:** I did not look at frames or listen to the pickup join (s3|s8, s8|s6). Check passes, but the pickup's picture (lighting, framing) vs talk is unverified. The "first thing" false start was cut as the first take; I assumed the aside was just the cat line.
- **Best part:** quoted-text cuts, the `retake-kept` hint with a ready command, and `check` verifying per-segment sync across sources.
- **One change:** a `edit replace --source pickup --from ... --to ...` that swaps the matching original span for the pickup in one command.
