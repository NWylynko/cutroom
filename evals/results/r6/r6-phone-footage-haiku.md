# Friction report: r6-phone-footage-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s, 1280x720) passes `cutroom check` with 0 failures and 0 warnings, the retake and the cat aside are gone, and ffprobe shows no rotation metadata. I could not confirm orientation by eye (see Not sure).

- **Commands run:** 31 cutroom invocations, plus ffprobe and one Read of a PNG.
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom guide edit`
  4. `cutroom guide render`
  5. `cutroom guide verify`
  6. `cutroom guide pitfalls`
  7. `cutroom guide perceive`
  8. `cutroom doctor`
  9. `cutroom init talk-rot.mov`
  10. `cutroom inspect`
  11. `cutroom transcribe`
  12. `cutroom transcript`
  13. `cutroom transcript talk-rot --suspects`
  14. `cutroom edit tighten`
  15. `cutroom timeline`
  16. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`
  17. `cutroom edit cut --from "By the way, my cat" --to "keyboard." --dry-run`
  18. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."`
  19. `cutroom edit cut --from "By the way, my cat" --to "keyboard."`
  20. `cutroom timeline`
  21. `cutroom timeline --joins`
  22. `cutroom render --preview --dry-run`
  23. `cutroom render --preview`
  24. `cutroom check`
  25. `cutroom look 2 --output 0:02` (error)
  26. `cutroom look --file out/preview.mp4 2`
  27. `cutroom render --dry-run`
  28. `cutroom render`
  29. `cutroom check out/final.mp4`
  30. `cutroom transcript --file out/final.mp4`
  31. `cutroom transcript --file out/final.mp4 --suspects`

- **Errors hit:**
  - `error[bad-args]: "2" is not a file and look takes one time or range` with `fix: one time or range, optionally after a file: cutroom look out/final.mp4 3-5`. The fix line got me unstuck, but my mistake was mixing a positional time with `--output`. The guide's `look --output 0:03-0:06` example is a range, which made the form unclear for a single frame. Only one error in the whole run.

- **Guessed wrong:**
  - I assumed `look <time> --output <time>` would work for a single frame. It doesn't. The single-frame form is `look --file <video> <time>`, which the error message named.
  - I assumed the `--near` value on `edit cut` was a source time. The output said so, but it was not obvious from the flag name alone.

- **Confusing output:**
  - After `edit tighten`, the `WARN[retake-kept]` line contained a ready-made `edit cut` command. It was useful, but the `--from` quote was long and the command only worked because of `--near`. I was not sure whether `--to "first thing."` would match the right occurrence until I ran `--dry-run`.
  - The `transcript` output for the final file is one long line per phrase. It was readable here, but the 44 words came out as a single paragraph, so I couldn't see the phrase boundaries.
  - `edit tighten` said it "also removes the sound with no words at source 0:16.970". That is a useful line, but I only saw it because I read the full output.

- **Missing:**
  - A way to check picture orientation directly. `check` reports `display rotation 0°` and `picture ... worst difference 0.0`, which implies the picture is compared to the rotated source, but it never states that orientation matches what the viewer would see on the phone. A line like "upright: source displays 180° rotated, render is upright" would close this gap.
  - Tighten's pause target is a fixed 0.32 s. I had no option to say "keep pauses at 0.5 s" or "keep natural pauses for a talking head", and I had no way to see the result without applying it.

- **Not sure it's right:**
  - Orientation. The source is rotated 180° by metadata. The guide and the `check` output say the render is upright in the pixels. The one frame I looked at (`look --file out/preview.mp4 2`) was a colour-bar test pattern, so I could not judge orientation from it. I accepted the tool's word and ffprobe's lack of rotation tags.
  - Transcription used `base.en`, which the tool warns can mishear words. The `words` check passed (44 of 44 in order), and my own reading of the final transcript matches the intended script. I did not run a larger model.
  - I shortened pauses to 0.32 s using the default tighten. That is a judgement call about the user's style, not something I verified.

- **Best part:** `--dry-run` on every `edit cut`, which printed the exact removed text, the new join and the new total duration before I changed anything. Combined with `timeline` and `timeline --joins`, the whole edit was easy to verify.

- **One change:** Let `check` (or `inspect` on the render) state the displayed orientation of the output explicitly, for example `display orientation: upright (source metadata 180°, baked into pixels)`. That would have settled the one thing I could not verify by eye.
