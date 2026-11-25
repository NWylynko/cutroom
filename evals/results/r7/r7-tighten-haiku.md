# Friction report: r7-tighten-haiku

- **Outcome:** done. `out/final.mp4` (13.8 s, down from 30.3 s) keeps all five sentences that belong to the explanation, drops the retake and the cat aside, and passes `cutroom check` with 0 warnings; the final transcript reads correctly.

- **Commands run:** 29 cutroom commands, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom guide edit`
  4. `cutroom guide perceive`
  5. `cutroom guide render`
  6. `cutroom guide verify`
  7. `cutroom guide pitfalls`
  8. `cutroom doctor`
  9. `cutroom init talk.mp4`
  10. `cutroom inspect`
  11. `cutroom transcribe`
  12. `cutroom transcript`
  13. `cutroom transcript --suspects`
  14. `cutroom edit tighten`
  15. `cutroom timeline`
  16. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`
  17. `cutroom edit cut --from "By the way, my cat" --to "across the keyboard." --dry-run`
  18. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing."`
  19. `cutroom edit cut --from "By the way, my cat" --to "across the keyboard."`
  20. `cutroom timeline`
  21. `cutroom timeline --joins`
  22. `cutroom render --preview --dry-run`
  23. `cutroom render --preview`
  24. `cutroom check out/preview.mp4`
  25. `cutroom look --cuts --file out/preview.mp4`
  26. `cutroom listen --output 3-5 --file out/preview.mp4`
  27. `cutroom render --dry-run`
  28. `cutroom render`
  29. `cutroom check out/final.mp4`
  Plus `cutroom transcript --file out/final.mp4` to read the result. Also two image reads (the cut contact sheet and the join spectrogram).

- **Errors hit:** none. Every command succeeded on the first try.

- **Guessed wrong:** nothing material. I did not have to guess a flag. The `retake-kept` warning suggested `--near 6.31`, and I used `--from/--to` with the same text instead; that also worked because the `--from` text was unique.

- **Confusing output:** `cutroom transcript` tags the retake's second take as `[repeat]` but also tags the first as `[repeat]`; it's clear from context, but the marker does not say which take to keep. The `retake-kept` warning after `tighten` was helpful but long (one line with a full command and a source time range). `cutroom timeline` labels segments with output times and source ranges in the same row, which took a second read.

- **Missing:**
  - Nothing blocked me, but there is no way to ask for a stronger transcription model from inside the task flow. `doctor` only hints at it, and the transcript note says "expect misheard words" without saying how to get the larger model.
  - I'd have liked a single command that prints the final transcript and the cut join audio together. I had to call `listen` and `transcript --file` separately.

- **Not sure it's right:**
  - The `words` check in `cutroom check` re-transcribes the render with the same `base.en` model used to make the timeline. It can't catch a systematic mishearing, so a PASS there is weaker evidence than it looks. I checked the final transcript by eye: it reads correctly, but I cannot hear the audio.
  - I listened to only one join (`listen --output 3-5`, the cut at 0:03.933). I looked at the cut-frame contact sheet for all four cuts and the spectrogram for that one join. The other three joins have 0.30–0.33 s pauses and no stray sound, and `check` reports `stray-sound` as clean, but I did not listen to them.
  - The "Hi, today I'm going to show you how to set up a home server." intro is kept as-is. I judged it on-topic.

- **Best part:** `cutroom transcript --suspects` surfaced both the retake and the stray sound at 0:16.970 in one call, and `edit cut --dry-run` showed the exact join before applying each cut. Combined with `tighten` removing the stray sound automatically, the whole edit took a handful of commands.

- **One change:** have `cutroom check` (or the transcript step) flag when the verifying transcription used the same small model as the timeline, and suggest a larger model for a real check. That would make the PASS honest about what it proves.
