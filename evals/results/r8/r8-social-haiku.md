# Friction report (r8-social-haiku)

- **Outcome:** done — `out/final.mp4` (13.767 s, 1280x720, burned-in captions) and `out/final.srt` (5 cues) both rendered, and `cutroom check out/final.mp4` passes with 0 failures and 0 warnings. Caveat: I could not watch the picture; the source is a test-pattern video with a frame counter, so caption placement was judged from a contact sheet only.

- **Commands run:** 31 shell calls, in order:
  1. `cutroom guide` (then read it)
  2. `cutroom guide workflow`, `guide perceive`, `guide edit`
  3. `cutroom guide captions`, `guide render`, `guide verify`, `guide pitfalls`
  4. `cutroom doctor`
  5. `cutroom init talk.mp4`
  6. `cutroom inspect`
  7. `cutroom transcribe`
  8. `cutroom transcript talk`
  9. `cutroom transcript talk --suspects`
  10. `cutroom transcript talk --words --from 0 --to 31`
  11. `cutroom edit tighten`
  12. `cutroom edit cut --from "By the way" --to "keyboard" --dry-run`
  13. `cutroom edit cut --from "UM" --to "thing" --dry-run`
  14. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`
  15. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."`
  16. `cutroom edit cut --from "By the way" --to "keyboard"`
  17. `cutroom timeline`
  18. `cutroom timeline --joins`
  19. `cutroom captions --burn --output out/final.srt`
  20. `cutroom captions --check out/final.srt`
  21. `cutroom captions --burn --dry-run`
  22. `cutroom render --preview --dry-run`
  23. `cutroom render --preview`
  24. `cutroom check out/preview.mp4`
  25. `cutroom look --file out/preview.mp4 --count 12` (error)
  26. `cutroom look 0-13.7 --count 12 --file out/preview.mp4`
  27. Read the PNG it wrote
  28. `cutroom render`
  29. `cutroom check out/final.mp4`
  30. `cat out/final.srt`
  31. `cutroom status`

- **Errors hit:**
  - `error[bad-args]: look --file needs a time` — fix line: "give a time or range in the file, e.g. `cutroom look 3.9 --file out/final.mp4`". The fix line got me unstuck immediately. The error was the only one.

- **Guessed wrong:**
  - I assumed `look --file <path>` alone would give a contact sheet of the whole file. It needs a time or range.
  - I assumed `captions --burn` would write the SRT to the default `out/captions.srt`. It writes there only without `--output`; I passed `--output out/final.srt`, which worked. The guide says this, but I had to read it closely.
  - I expected `cutroom timeline`'s `retake-kept` suggestion to be the right cut. It removed the whole false start, which was correct, but I had to dry-run it to be sure the join was clean.

- **Confusing output:**
  - `transcript` tags both "The first thing you need is, UM…" and "The first thing you need is an old laptop." with `[repeat]`. Only the first is the false start, so you have to read the text to see which take is the keeper.
  - `captions --burn` prints the first four groups and then "… 4 more (--all)". A caption list is short enough that I'd have wanted the whole list in the default output.
  - `cutroom check` reports `words` as passing, but its note says the re-transcription used base.en, the same model that made the transcript. That check is therefore circular, and the text does not say how far to trust it when the transcript came from base.en.

- **Missing:**
  - No way to view the picture as a whole. `look` gives frames and contact sheets, but nothing like "does the caption sit clear of the subject" for a real face. I could not judge legibility beyond one 960x540 contact sheet.
  - No option to choose the output aspect ratio or a vertical crop for social. The request said "social media" and the guide has a `shorts` topic for vertical output, but I kept 16:9 since the brief did not ask for vertical and the source is 16:9. I did not check whether 16:9 is the right choice for the platform.
  - No single "which take did I keep" summary. I had to infer it from the timeline segments.

- **Not sure it's right:**
  - The false-start cut removes "The first thing you need is, UM, the first thing." (source 6.20–9.43) and keeps "The first thing you need is an old laptop." The join reads "…up a home server. | The first thing you…" with a 0.30 s pause. I did not listen to it with `cutroom listen`.
  - The cat aside was removed on the judgment that it is off-topic. The removal is correct for the brief, but the speaker's recording is a test pattern, so I could not confirm the picture around it.
  - Caption timing (e.g. the first cue shows from 0.113 s, before the first word at 0.467 s) follows the guide's rule, but I did not confirm it against audio beyond the `words` and `sync` checks.
  - The SRT and the burned captions come from the same cue list, and `captions --check` passes against the timeline. I did not compare the SRT to the rendered picture frame by frame.

- **Best part:** The timeline's `WARN[retake-kept]` gave a ready-to-run `edit cut` with exact quoted words and `--near`, and `--dry-run` showed the join before I applied it. Combined with `check` ending in a per-line PASS/FAIL list, the loop from edit to verified render took few steps.

- **One change:** Make `look --file <path>` with no time show the whole file as a contact sheet, or print the range in the error, so the common "show me the rendered output" case takes one call instead of two.
