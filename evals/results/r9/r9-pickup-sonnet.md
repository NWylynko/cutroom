# Friction report

- **Outcome:** done. out/final.mp4 rendered (13.8 s) and `cutroom check` passed with 0 failures and 0 warnings.
- **Commands run:** 8 in total: `guide`, `guide workflow`, `guide edit`, `init talk.mp4 pickup.mp4`, `transcribe`, `transcript --suspects`, `transcript`, `transcript pickup`, `edit tighten`, `edit cut` (false start), `edit cut` (aside), `edit replace --text ... --with pickup`, `timeline --joins`, `render`, `check`.
  (The count is approximate. Several of these ran chained in one shell call.)
- **Errors hit:** none.
- **Guessed wrong:** nothing. The workflow guide covered the pickup swap with an exact command.
- **Confusing output:** `tighten` printed a `retake-kept` warning that suggested a cut. I made my own equivalent cut afterwards, and the warning was gone. It was unclear whether I should have used the suggested command. I did not run `cutroom listen` on the pickup join and relied on the audio-mixed INFO line.
- **Missing:** nothing.
- **Not sure it's right:** I could not hear the audio or the pickup join. `check` re-transcribed with the same small base.en model, so a misheard word would not show up.
- **Best part:** `edit replace --text --with pickup` did the swap in one step. `check` summarised the result clearly.
- **One change:** have `transcript --suspects` also say when a pickup source repeats a line already in the talk, and suggest the `edit replace` command.
