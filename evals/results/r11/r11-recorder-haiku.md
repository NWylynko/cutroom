# Friction report: 11-recorder-haiku

- **Outcome:** done. out/final.mp4 (20.433 s, 1280x720) uses recorder.wav aligned to talk (offset +1.700 s, drift -42.8 ppm), with the long pauses tightened; `cutroom check out/final.mp4` passes with 0 warnings. Retake and aside were left in because the task only asked for pause removal.

- **Commands run:** 21 cutroom invocations over 17 shell calls (plus one `ls`), in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide edit`, `cutroom guide verify`
  3. `cutroom doctor`
  4. `cutroom init talk.mp4 --source mic=recorder.wav`
  5. `cutroom inspect`
  6. `cutroom sync mic`
  7. `cutroom sync --help`
  8. `cutroom edit audio mic`
  9. `cutroom transcribe`
  10. `cutroom transcript`
  11. `cutroom transcript talk --suspects`
  12. `cutroom edit tighten --dry-run`
  13. `cutroom edit tighten`
  14. `cutroom timeline`
  15. `cutroom timeline --joins`
  16. `cutroom render --preview`
  17. `cutroom check out/preview.mp4`
  18. `cutroom render`
  19. `cutroom check`

- **Errors hit:** none. No command exited non-zero.

- **Guessed wrong:** Nothing that failed. I did not use `--json` anywhere, so I can't say whether the JSON output differs in a way that matters.

- **Confusing output:**
  - Three different durations for the same talk: `inspect` says the video is 30.261 s, `transcript` says 0:30.233, and `edit audio` says the timeline was created as "whole source, 30.200 s". I could not tell which one the edit actually covers.
  - `edit audio mic` printed `timeline: 1 segment, 30.200 s (was 30.200 s)`. The "was" wording is odd for a timeline that was just created.
  - `sync` printed `mapping: mic time = +1.700 s + talk time × (1 -42.8 ppm)`. The sign convention for the drift term isn't stated, so I could not check it from the output alone.
  - `transcript talk` says the words come "from recorder mic", while `transcript` also says "also transcribed: mic". Both are reasonable, but reading them together took a second look.
  - `timeline` gives the retake and aside warnings with a ready-made `edit cut` command each, and the `tighten` output lists the pauses. Both were clear. The output was long but readable.

- **Missing:**
  - A way to show the sync offset and drift as an end-to-end check over the whole video, not just the 54 windows summarised by `sync`. The check reports a worst case of -6 ms, but the only way I had to see this was `cutroom check`.
  - A preview size flag. `render --preview` produced 960x540 silently, and I only noticed it from the file. The final render is 1280x720, so this only matters for how the preview looks.

- **Not sure it's right:**
  - The sync check says worst -6 ms over 7 segments. The `words` check re-transcribes with base.en, so it confirms the cuts but cannot confirm the recogniser was right.
  - I did not listen to any join. I relied on `timeline --joins` and the `check` results. The sound-without-words at 16.971 s was removed by tighten, which the dry run predicted.
  - I did not open or watch out/final.mp4. Every claim about the result comes from `check` and from the tool's own text.

- **Best part:** The `cutroom guide edit` and `cutroom guide workflow` sections include a "Separate audio recorder" recipe: `init --source`, `sync`, `edit audio`. Following it, the whole job took about 19 shell calls with no errors, and `check` gave a clear per-check PASS table.

- **One change:** Make `edit audio` print one duration that matches `transcript` and `inspect`, and drop the "was" wording when it creates the timeline.
