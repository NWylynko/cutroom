# Friction report

- **Outcome:** done. out/final.mp4 (13.8 s) rendered; `cutroom check` PASS with 0 failures and 0 warnings.
- **Commands run:** 8 shell calls, 14 cutroom commands: guide, guide workflow, guide edit, init talk.mp4 pickup.mp4, transcribe, transcript, transcript --suspects, transcript pickup, edit tighten, edit cut (false start), edit cut (aside), edit replace --text ... --with pickup, timeline, timeline --joins, render, check.
- **Errors hit:** none.
- **Guessed wrong:** nothing. `edit replace` was named in the workflow guide.
- **Confusing output:** the first `edit cut` ran after tighten and the earlier retake-kept warning suggested a source range starting 6.200. Fine, but I had to confirm the cut removed the first "The first thing you need is" and the "um" take and kept the last. The timeline showed that.
- **Missing:** nothing needed. I did not listen to the pickup join (`cutroom listen`); the audio-mixed INFO said levels match.
- **Not sure it's right:** the transcript is from base.en, so the pickup text match relied on it hearing "Then install Linux on it" in both takes. I did not look at frames or listen to audio.
- **Best part:** `edit replace --text --with pickup` did the swap in one step, and the quoted-text cuts avoided any time typing.
- **One change:** `guide` could mention up front that `init` accepts several files and `edit replace` handles pickups, so the pickup case is found without reading the workflow topic.
