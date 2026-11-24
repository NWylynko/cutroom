# Friction report

- **Outcome:** done — `out/final.mp4` (13.767 s, down from 30.261 s) rendered and `cutroom check` passed with 0 warnings. Edits: tightened pauses, cut the false start "The first thing you need is, UM, the first thing.", and cut the off-topic "By the way, my cat just walked across the keyboard." Joins were checked with `timeline --joins` and `look --cuts` (images), not listened to one by one.

- **Commands run:** about 24 cutroom invocations, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide perceive`, `cutroom guide edit`
  3. `cutroom guide render`, `cutroom guide verify`, `cutroom guide pitfalls`
  4. `cutroom doctor`
  5. `cutroom init talk.mp4`
  6. `cutroom inspect`
  7. `cutroom transcribe`
  8. `cutroom transcript talk`
  9. `cutroom transcript talk --suspects`
  10. `cutroom transcript talk --words`
  11. `cutroom edit tighten`
  12. `cutroom edit cut --dry-run --from "The first thing you need is, UM" --to "first thing."`
  13. `cutroom edit cut --dry-run --from "By the way" --to "keyboard."`
  14. `cutroom listen 16.5-17.5`
  15. `cutroom edit cut --from "The first thing you need is, UM" --to "first thing."`
  16. `cutroom edit cut --from "By the way" --to "keyboard."`
  17. `cutroom timeline`
  18. `cutroom timeline --joins`
  19. `cutroom render --preview --dry-run`
  20. `cutroom render --preview`
  21. `cutroom check out/preview.mp4`
  22. `cutroom render`
  23. `cutroom check out/final.mp4`
  24. `cutroom look --cuts`

- **Errors hit:** none. No command failed.

- **Guessed wrong:** Nothing failed, but I did not use `--json` at all, so I never saw the structured output the guide describes. I also did not test the ambiguous `--from` path (numbered list plus `--occurrence`/`--near`), since my phrases were unique.

- **Confusing output:**
  - `cutroom transcript --words` cut itself off at 40 lines ("… 24 more lines"). I did not need the rest, but the truncation note sits in the middle of a table, which is easy to miss.
  - In `tighten`'s output, "pauses (source time)" lists the pauses with source times, while `timeline` columns are output times. The guide explains this, but I had to re-read it to keep the two straight.
  - The `retake-kept` warning gave a ready command with `--near 6.31`. That was useful, but I did not need `--near` because my phrase was unique. The output is long; the warning repeats on every subsequent command (`listen`, `cut --dry-run`), which is noisy.

- **Missing:**
  - No way to cut by transcript phrase number (for example, `cut --phrase 2`). Quoting words worked, but phrase numbers would be less error-prone for whole-sentence cuts.
  - Because the recogniser is `base.en`, I had no way to confirm the words without listening. `check` only compares the render against the timeline's own transcript, so it cannot catch a misheard word in the timeline.

- **Not sure it's right:**
  - I chose which take to keep from the transcript and the `repeat` suspect. I kept the second take ("The first thing you need is an old laptop.") and did not listen to either take in full.
  - I judged the "off-topic" aside from the transcript text alone. The intro, the laptop/Linux/terminal steps and the closing line are all kept, but I did not check the text against the actual speech beyond `check words`.
  - The three joins at 0.30–0.33 s pauses were checked visually (frames) and by transcript; I did not run `cutroom listen` on them.
  - The `sound` at 16.97 s (0.32 s, -22 dB) was removed by `tighten` as part of a pause. `listen` showed no words there, but I did not confirm what it was.

- **Best part:** `transcript --suspects` with the `repeat` finding, followed by `edit cut --dry-run` (which shows the join and the resulting timeline length before applying), then `render --preview` and `check` gave a clear loop with no surprises.

- **One change:** Let `edit cut` take transcript phrase numbers from `cutroom transcript` (for example, `cutroom edit cut --phrase 2`) so that whole-sentence cuts do not depend on quoting exact words and on uniqueness.
