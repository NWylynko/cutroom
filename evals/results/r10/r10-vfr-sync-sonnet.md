# Friction report

- **Outcome:** done. out/final.mp4 rendered (17.233 s, down from 30.26 s). `cutroom check` passed with 0 failures and 0 warnings, including sync at -0 ms worst.
- **Commands run:** 12. guide; guide workflow + pitfalls; init; inspect (printed by init); transcribe; transcript --suspects; transcript; guide edit; edit tighten; edit cut --dry-run; edit cut; timeline --accept; render; check. (Some were chained in one shell call.)
- **Errors hit:** none.
- **Guessed wrong:** nothing. I left the "by the way" cat aside in because only pauses and the false start were requested, and I accepted the warning with a reason.
- **Confusing output:** the tighten output prints WARNs and a suggested cut for the retake, which I had to re-read to see it matched what I wanted. I typed a slightly different `--from`, which worked. The `--accept` time (8.26) was only guessed from "output 0:08.257"; it worked.
- **Missing:** nothing important. A one-line "this render has VFR, handled" summary exists in render output, which was good.
- **Not sure it's right:** the transcript comes from base.en, so the false-start wording ("UM") is a guess. The words check only confirms the cut against the same model. I did not listen to the joins.
- **Best part:** VFR and the audio offset are handled automatically, and check measures sync per segment. That answered the lip-sync worry directly.
- **One change:** have `render` or `check` print an explicit "lip sync verified" verdict line aimed at the user's concern, though the sync line already says most of this.
