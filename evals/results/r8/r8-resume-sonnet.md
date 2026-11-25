# Friction report
- **Outcome:** done. Previous editor had tightened and cut the cat aside; I applied the one open retake-kept warning, rendered, and `check` passed (0 failures, 0 warnings).
- **Commands run:** 8: `cutroom guide`; `cutroom guide workflow`; `cutroom status`; `cutroom edit history`; `cutroom transcript`; `cutroom timeline`; `cutroom timeline --apply retake-kept@4.09`; `cutroom timeline --joins`; `cutroom render`; `cutroom check` (plus ls/cat of project files).
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** none significant. Note history showed only 2 edits, but the stale preview existed; fine.
- **Missing:** nothing. Resuming was easy because `status` printed state and the next command.
- **Not sure it's right:** I did not look at frames or listen to joins; the 0.30 s pauses at joins and the "UM" removal rest on transcript and check only. The initial 0.47 s lead-in was kept as the earlier editor set it.
- **Best part:** `status` plus `timeline --apply <warning>` gave a one-command fix; `check` verifies words, sync and loudness.
- **One change:** none needed; maybe have `status` say which edits were by "someone else" vs pending decisions.
