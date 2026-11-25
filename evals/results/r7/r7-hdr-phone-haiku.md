# Friction report: r7-hdr-phone-haiku

- **Outcome:** done. `out/final.mp4` passes `cutroom check` with 0 failures and 0 warnings, and the false start and cat remark are gone. The one thing I could not confirm is the look of the HDR-to-SDR conversion (see "Not sure it's right").

- **Commands run:** 29 cutroom commands, plus one `ffprobe` on the output.
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom guide perceive`
  4. `cutroom guide edit`
  5. `cutroom guide render`
  6. `cutroom guide verify`
  7. `cutroom guide pitfalls`
  8. `cutroom doctor`
  9. `cutroom init clip.mov`
  10. `cutroom inspect`
  11. `cutroom transcribe`
  12. `cutroom transcript`
  13. `cutroom transcript --suspects`
  14. `cutroom edit tighten`
  15. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`
  16. `cutroom edit cut --from "By the way, my cat" --to "across the keyboard." --dry-run`
  17. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing."`
  18. `cutroom edit cut --from "By the way, my cat" --to "across the keyboard."`
  19. `cutroom timeline`
  20. `cutroom timeline --joins`
  21. `cutroom render --preview --dry-run`
  22. `cutroom render --preview`
  23. `cutroom check out/preview.mp4`
  24. `cutroom look --output 0:02-0:03 --count 2`
  25. `cutroom look 5`
  26. `cutroom render --dry-run`
  27. `cutroom render`
  28. `cutroom check`
  29. `cutroom transcript --file out/final.mp4`

- **Errors hit:** None. Every command exited 0 on the first try.

- **Guessed wrong:** Nothing major. I expected `--from` text matching to be stricter about "UM," with the comma, and it accepted the quote as written. The `tighten` step also removed the sound at 0:16.970 without a separate cut, which the output did report ("this also removes the sound with no words"), so I did not need to act on it.

- **Confusing output:**
  - After `tighten`, the `retake-kept` WARN printed a ready-made `edit cut --from "The first thing you need" --near 6.31 --to "first thing."`. The text cut I used was equivalent, but I had to compare the two by hand to be sure the `--near` form was not needed.
  - `timeline` said "no problems found" after the cuts, yet the earlier `retake-kept` line had only been resolved implicitly. The `--dry-run` output for the second cut ("still open (unchanged by this edit): retake-kept") was the only place that made this clear. A one-line note on which warnings each edit resolved would help.
  - `check` reports `tone` as "saturation 87 vs source 87, brightness range 217 vs 217". I could not tell from this whether the picture is correct or just matches its own statistics, because the source is also tone-mapped in the same run. See below.

- **Missing:**
  - A way to see the HDR source without tone-mapping, or a reference SDR frame, so I could compare the two side by side with `cutroom look`. `cutroom look 5` on clip.mov showed a tone-mapped frame, and I had no way to tell that apart from a washed-out one.
  - A `listen` on the two joins I made (0:03.9, 0:11.5) before rendering. I read them from the transcript and `--joins`, but I did not listen to them, and the guide recommends listening to any doubtful join.
  - The render line "tone-mapped to SDR bt709 (videotoolbox)" does not say whether videotoolbox is a hardware path with different quality from the software path, or whether there is a flag to choose.

- **Not sure it's right:**
  - HDR look. `ffprobe` on `out/final.mp4` shows bt709 primaries, bt709 transfer, bt709 colour space and yuv420p, with no rotation tag. The frames I viewed (preview at 2.0 s and 3.0 s) are upright and saturated, and a source frame at 5 s looked the same. I could not verify that this is the correct SDR rendition instead of a washed-out or over-saturated one, because I had no SDR reference to compare against.
  - Joins. The joins are 0.30–0.33 s pauses in the transcript and the `--joins` output reads naturally. I have not listened to them.
  - Transcript check. The `words` check passed, but base.en is a small model, and `transcript --file out/final.mp4` read back the text correctly. I still have not heard the audio.

- **Best part:** The guide plus `timeline --joins` and `check` made it easy to finish. `cutroom edit cut --from ... --to ...` with `--dry-run` showed the exact join before applying it, and `check` gave a clear per-line pass/fail. The HDR and rotation handling was automatic and visible in `inspect` and `render` output, so I did not have to do anything special for the phone footage.

- **One change:** Add a flag or a sample that renders the HDR source as a plain SDR frame next to the tone-mapped render (for example `cutroom look 5 --compare`), so an agent can see whether the conversion looks right without guessing from statistics.
