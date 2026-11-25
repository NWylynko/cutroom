# Friction report (r10-social-haiku)

- **Outcome:** done. `out/final.mp4` (13.767 s, burned-in captions) and `out/final.srt` (5 cues) both pass `cutroom check` / `cutroom captions --check` with 0 failures and 0 warnings, and the cut joins were looked at.

- **Commands run:** 37 cutroom invocations, in order, plus shell helpers (`ls`, `mkdir -p out`, `cat out/final.srt`) and 3 image reads.
  1. `cutroom guide`
  2. `cutroom guide workflow`, `perceive`, `edit`
  3. `cutroom guide captions`, `overlays`, `render`, `verify`, `pitfalls`
  4. `cutroom doctor`
  5. `cutroom init talk.mp4`
  6. `cutroom inspect`
  7. `cutroom transcribe`
  8. `cutroom transcript`
  9. `cutroom transcript --suspects`
  10. `cutroom transcript talk --words --from 6 --to 12.5`
  11. `cutroom listen 16.5-17.8`
  12. `cutroom edit tighten`
  13. `cutroom edit cut --from "The first thing you need is, UM" --to "first thing." --dry-run`
  14. `cutroom edit cut --from "By the way" --to "keyboard." --dry-run`
  15. `cutroom edit cut --from "The first thing you need is, UM" --to "first thing."`
  16. `cutroom edit cut --from "By the way" --to "keyboard."`
  17. `cutroom timeline`
  18. `cutroom timeline --joins`
  19. `cutroom captions --burn --output out/final.srt`
  20. `cutroom captions --burn --help`
  21. `cutroom render --preview --dry-run`
  22. `cutroom render --preview`
  23. `cutroom check out/preview.mp4`
  24. `cutroom look --file out/preview.mp4 --output 0:02.5`
  25. `cutroom look --file out/preview.mp4 --output 0:06.0`
  26. `cutroom render`
  27. `cutroom check out/final.mp4`
  28. `cutroom captions --check out/final.srt`
  29. `cutroom status`
  30. `cutroom look --cuts --file out/final.mp4`
  31. `cutroom edit history`

  (Numbering is approximate by one or two; the sequence is correct.)

- **Errors hit:** none. Every command exited 0. The only hard blocker was none; I never saw an `error[...]` line.

- **Guessed wrong:** nothing failed, but two things took a second look:
  - `--output` on `cutroom look` with `--file` reads as "output time" by name, yet the label said `(file time)`. It worked, but the flag name is overloaded (render's `--output` is a path, look's `--output` is a time).
  - I assumed `captions --burn` would write the 2-line SRT and the burned-in text would use a different segmentation (1–3 words). The guide says so, but the two outputs differ in cue boundaries, and that is easy to miss when you only read the SRT.

- **Confusing output:**
  - `cutroom transcript` marks the retake as `[repeat]` on both takes; I had to read `--words` to see that the first take is the one to drop. Fine, but the phrase-level view alone does not say which take is the keeper.
  - `cutroom timeline` after `tighten` printed a `retake-kept` WARN with a ready `--near` command; the `--from` version I used in dry-run gave the same range. Both work; the WARN's text is long.
  - `cutroom check` printed `words … re-transcribed with base.en, the model the transcript used: it confirms the cut, not the recogniser`. That is honest, but it means the words check cannot catch a misheard retake choice.

- **Missing:**
  - No way to see which models are available for a larger transcription pass from `doctor` (it only shows base.en). I wanted a second opinion on the retake choice and on the transcript wording; `cutroom transcribe --model` was in the check's text but I did not find how to list the options.
  - No single command that lists the SRT cue boundaries next to the burned-in groups, so the two outputs can be compared.

- **Not sure it's right:**
  - I chose the second take ("The first thing you need is an old laptop.") from the transcript and word timings only. I did not `cutroom listen` the first take's audio to confirm the stumble was in it. The `UM` and the repeat are clear in the transcript, so I am fairly confident.
  - I removed the cat aside on judgement (off-topic); the transcript flags it, and I agree.
  - The cough or sound at 16.97 s was removed by `tighten` (the tool said so); I did not listen to the final render at that point, but `check` reports `stray-sound` ok.
  - The source is a synthetic test-pattern video, so visual checks are only about caption placement and legibility, not the speaker.
  - `check`'s words pass uses the same base.en model, so it confirms the cut against the same recogniser; it does not independently verify the transcription of the speech.

- **Best part:** `cutroom transcript --suspects` plus `edit cut --dry-run` made the whole edit fast and safe. The dry-run join line (`join: … | …  (pause 0.30 s)`) showed exactly what each cut would sound like before it was applied, and `timeline --joins` plus `look --cuts` confirmed the result.

- **One change:** a single command or flag on `captions` that makes the subtitle file and the burned-in captions use the same cue boundaries (or prints both side by side), so an agent can see the sidecar and the picture agree before rendering.
