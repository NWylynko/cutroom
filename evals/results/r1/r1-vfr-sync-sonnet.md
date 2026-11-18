# Friction report
- **Outcome:** done. out/final.mp4 is 17.233 s (from 30.261 s): pauses tightened, false start ("The first thing you need is, UM, the first thing.") cut. The cat aside was left in (not requested).
- **Commands run:** 11 cutroom (guide, guide workflow/edit/pitfalls, doctor, init, inspect, transcribe, transcript, transcript --suspects, edit tighten, edit cut --dry-run, edit cut, timeline --joins, render --preview --range, render, check) plus 1 ffprobe.
- **Errors hit:** `error[unknown-command]: unknown command "check"` (the guide says check is not yet available; the fix line just said run `cutroom`). Did not block.
- **Guessed wrong:** that `cutroom check` exists; guide rule 1 mandates it.
- **Confusing output:** guide rule says never finish without `check`, but check does not exist.
- **Missing:** `check` (A/V sync, duration vs timeline). I verified by ffprobe only: video and audio both start 0 and are 17.233 s, 517 frames. inspect marked VFR and +0.064 s audio offset as "handled", but I could not independently verify sync after the cuts.
- **Not sure it's right:** lip sync across the VFR source; whether the join after the cut sounds clean (did not run `listen`); base.en transcript accuracy.
- **Best part:** quoted-text cuts, --dry-run, and handled-warnings in inspect.
- **One change:** ship `cutroom check` (or remove the rule requiring it), including an audio/video sync measurement.
