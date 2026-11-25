# Friction report

- **Outcome:** done. Applied the pending retake-kept fix, rendered out/final.mp4 (13.767 s), `check` PASS with 0 failures and 0 warnings.
- **Commands run:** 7: `cutroom guide`; `cutroom guide workflow` and `cutroom status` (plus cat/ls); `cutroom transcript`, `timeline`, `edit history`; `timeline --apply retake-kept@4.09`; `timeline --joins`; `render`; `check`.
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** none. `status` showed the interrupted state (cat aside already cut, retake warning open) and its `next:` lines gave the exact fix.
- **Missing:** nothing needed.
- **Not sure it's right:** I did not listen to the joins or view frames. The checks passed, and the words check confirms every expected word.
- **Best part:** `status` plus `timeline` warnings with `--apply` made resuming without notes trivial.
- **One change:** have `status` also say outright "last render is stale or preview only; run render" after applying fixes. Minor.
