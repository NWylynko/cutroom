# Friction report: r9-pickup-haiku

- **Outcome:** done. `out/final.mp4` (13.8 s, 1280x720) has the talk with the pickup's "Then install Linux on it." in place of the original, and the false start, the cat aside and the long pauses removed. `cutroom check` passes on the final, and the final transcript reads correctly end to end.

- **Commands run:** 36 `cutroom` invocations, plus one `ls` and one image read of the look PNG. In order:
  1. `cutroom guide`, then `guide workflow`, `guide edit`, `guide perceive`, `guide render`, `guide verify`, `guide pitfalls`
  2. `cutroom doctor`
  3. `cutroom init talk.mp4 pickup.mp4`
  4. `cutroom inspect`
  5. `cutroom transcribe`
  6. `cutroom transcript`, `cutroom transcript --suspects`
  7. `cutroom transcript pickup`, `cutroom transcript pickup --words`, `cutroom transcript talk --words` (piped to head)
  8. `cutroom transcript talk --words --from 38 --to 63` (error, see below)
  9. `cutroom edit replace --help`, `cutroom edit --help`
  10. `cutroom transcript talk --words --from 18 --to 30.3`
  11. `cutroom edit tighten --dry-run`, `cutroom edit replace --text "Then install Linux on it." --with pickup --source talk --dry-run`
  12. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --source talk --reason "false start"`
  13. `cutroom edit cut --from "By the way" --to "keyboard." --source talk --reason "off-topic aside"`
  14. `cutroom edit replace --text "Then install Linux on it." --with pickup --source talk --reason "speaker re-recorded line"`
  15. `cutroom edit tighten`
  16. `cutroom timeline`, `cutroom timeline --joins`
  17. `cutroom render --preview --dry-run`, `cutroom render --preview`, `cutroom check`
  18. `cutroom render`, `cutroom check`
  19. `cutroom transcript --file out/final.mp4`
  20. `cutroom look --output 0:06.5`
  21. `cutroom timeline --help`

- **Errors hit:**
  - `error[out-of-range]: --from 38 is past the end of talk` / `fix: talk is 30.261 s long; --from/--to are times in seconds (or m:ss), not word numbers`. The fix line got me unstuck at once. I had guessed `--words --from` took word indices.
  - No other errors. Every edit printed its own warnings and a ready-made suggested command.

- **Guessed wrong:**
  - I assumed `transcript --words --from/--to` took word numbers (see above). It takes times.
  - I assumed `edit replace` needed `--source talk` to avoid matching the pickup's copy of the same words. I added it without testing whether it was needed.
  - I ran the replace dry-run before the aside cut, so its preview showed a cut to 18.500 s. Once the aside was cut first, the real replace cut ended at 15.667 s. This was my ordering, not a tool bug, but a dry run after earlier edits would have matched the real result more closely.

- **Confusing output:**
  - `transcript --words` with no range printed 40 lines then "… 24 more lines". I had to re-run with `--from/--to`, so I did not see the full word list in one go.
  - After `edit replace`, the timeline lists segments by output time with ids out of order (`s1, s2, s4, s3, s5`). The ids are correct but easy to misread on a first pass.
  - `audio-mixed` INFO printed on every edit that touched the pickup, repeating the same "shouldn't be heard" advice each time.
  - `timeline` header says "sound from talk" even though `s4` is pickup audio. The per-segment detail is only in `check`'s sync line, which does say "(pickup, talk)".

- **Missing:**
  - A way to see the source transcripts without the 40-line cap by default, or a flag for "words in one line per phrase" for a single source.
  - A way to name or label a segment (e.g. "pickup line") so the timeline and `check` output say which take each line comes from.
  - Nothing else; the rest of the workflow had what I needed.

- **Not sure it's right:**
  - The pickup's version of the line is in the output, but I could not hear it. The `words` check and the final transcript both use base.en, which confirms the words match the timeline, not that the audio is the pickup's. The `sync` line says the pickup segment is measured against pickup's own sound, which is my best evidence.
  - `cutroom look` frames are a synthetic test pattern, so I could not judge the picture of the pickup line against the talk by eye. The `picture` check passes.
  - The `aside` cut also removed a sound with no words at 16.970 s, so I did not need to address the `kept-sound` warning separately. It is gone from the final `check`, but I did not `listen` to it.

- **Best part:** The `guide workflow` section "Several takes (re-recorded lines)" gave me `edit replace --text … --with pickup` in one line, and the dry-run plus `check` made the result easy to trust. Each warning also printed the exact `edit cut` command to fix it, which I could run as-is.

- **One change:** Let `transcript --words` (and `--search`) take a `--source` filter and a `--phrase` view, so I can quote exact word spans from either take in one call. Right now I had to guess time ranges from the phrase view and then re-run with `--from/--to`.
