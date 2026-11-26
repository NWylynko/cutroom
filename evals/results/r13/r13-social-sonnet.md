# Friction report

- **Outcome:** done. out/final.mp4 (13.8 s, burned captions) and out/final.srt produced; `cutroom check` PASS with 0 failures/warnings.
- **Commands run:** ~12: guide; guide workflow/edit/captions; init talk.mp4; transcribe; transcript; transcript --suspects; edit tighten; edit cut (dry-run x2); edit cut (retake: "The first thing you need is, um" .. "the first thing"); edit cut (aside "By the way" .. "keyboard"); timeline --joins; captions --burn --output out/final.srt; render; check; look --output 1.0.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant.
- **Confusing output:** the retake-kept warning suggested a different cut than the one I used; both removed the same take, so no harm. Cue 1 starts at 0.113 s while transcript says first word 0.58 (tightened), fine.
- **Missing:** nothing; I did not need the source code.
- **Not sure it's right:** transcription used the small base.en model, so misheard words are possible; check re-transcribed with the same model. I did not review caption visuals beyond `look` output (cannot see video); check's caption sampling passed.
- **Best part:** suspects list + the `next:` lines + dry-run showing joins; --burn writing the SRT in the same call.
- **One change:** have `captions --burn` say plainly that it wrote both outputs to the requested paths (it did, but late in the output).
