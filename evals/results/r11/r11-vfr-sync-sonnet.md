# Friction report

- **Outcome:** done. out/final.mp4 rendered (17.233 s, from 30.2 s) and `cutroom check` passed with 0 failures and 0 warnings, including sync.
- **Commands run:** 13 (cutroom only), plus one `cat > FRICTION.md` to write this file: guide, guide workflow/edit/verify (one call), doctor, init, inspect (one call), transcribe, transcript, transcript --suspects (one call), edit tighten, edit cut --dry-run (one call), edit cut, timeline --accept, timeline --joins, render --preview, render, check (one call).
- **Errors hit:** none.
- **Guessed wrong:** nothing. I kept the "by the way" cat aside because the task only asked for pauses and the false start, and accepted its warning with a reason.
- **Confusing output:** `edit tighten` printed its warnings and "next" lines before the dry-run output in the same call, so the two ran together. The `--accept` time (8.26) is the output time after the cut, not the 11.46 shown earlier. I had to guess that it moved.
- **Missing:** nothing important. A one-line "VFR handled, lip sync safe" line in `check` would help. The render output already says it.
- **Not sure it's right:** I can't hear the joins. The base.en recogniser confirmed the words, and the sync check was -0 ms. I did not run `cutroom listen` on the joins. I did not look at frames.
- **Best part:** the VFR and audio-offset notes in `init` and `render` said what was handled, and `check` measured sync per segment, which answered the lip-sync worry directly.
- **One change:** have `timeline --accept` take the warning code alone, or print the current output time of each open warning after an edit, so the time is not stale.
