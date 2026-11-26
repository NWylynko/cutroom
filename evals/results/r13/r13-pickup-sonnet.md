# Friction report

- **Outcome:** done. out/final.mp4 (13.8 s) has the pickup swapped in, pauses tightened, false start and cat aside cut; `check` PASS with 0 failures and 0 warnings.
- **Commands run:** 13 (about 14 counting the shell `ls`): guide; guide workflow + guide edit; init talk.mp4 pickup.mp4; transcribe; transcript; transcript --suspects; transcript pickup; edit tighten; edit cut (false start); edit cut (aside); edit replace --text ... --with pickup; timeline; timeline --joins; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant. I did not know how the pickup would be transcribed, so I checked `transcript pickup` first.
- **Confusing output:** `tighten` printed the retake-kept and aside-kept WARNs with ready-made cuts, so I used my own cuts, which was fine. The `--joins` output repeats the INFO audio-mixed line each time. Source times differ slightly between transcript (0:06.310) and the cut ranges (0:06.200); that is explained but surprising.
- **Missing:** nothing that blocked me. A quick listen or look at the swapped join is not offered automatically.
- **Not sure it's right:** I did not run `listen` across the pickup join (output 4.9-7.9 s). `check` words/sync pass, but naturalness of the join and the retake choice (I cut the whole "first thing you need is, UM, the first thing." take and kept the second complete one) rest on the transcript only. base.en is a small model.
- **Best part:** `edit replace --text ... --with pickup` did the swap in one step. `guide workflow` documented the pickup case directly.
- **One change:** have `check` or `render` print a one-line summary of the final kept text, so I can confirm the edit reads right without running `transcript --file`.
