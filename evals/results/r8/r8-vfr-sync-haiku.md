# Friction report: r8-vfr-sync-haiku

- **Outcome:** done, with one open question. out/final.mp4 (17.233 s, 1280x720) passes every `cutroom check` line, but the source's audio starts +0.064 s after its picture, and the tool has no way to correct that for the camera's own audio, so I could not confirm picture-to-speech alignment.

- **Commands run:** 24 `cutroom` invocations (plus `cd`/`ls`):
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom guide edit`
  4. `cutroom guide render`
  5. `cutroom guide verify`
  6. `cutroom guide pitfalls`
  7. `cutroom doctor`
  8. `cutroom init talk-vfr.mp4`
  9. `cutroom inspect`
  10. `cutroom transcribe`
  11. `cutroom transcript`
  12. `cutroom transcript --suspects`
  13. `cutroom edit tighten`
  14. `cutroom timeline`
  15. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`
  16. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."`
  17. `cutroom timeline`
  18. `cutroom timeline --joins`
  19. `cutroom render --preview --dry-run`
  20. `cutroom render --preview`
  21. `cutroom check` (on out/preview.mp4)
  22. `cutroom render --dry-run`
  23. `cutroom render`
  24. `cutroom check out/final.mp4`
  Also `cutroom sync --help` and `cutroom edit audio --help` to look for an offset option.

- **Errors hit:** None from cutroom. The only error was a hook message from my own task-list tool (`PreToolUse:mcp__ccd_turn__update_plan hook error: ... you are a subagent`), which is not a cutroom error. I skipped it.

- **Guessed wrong:** Nothing. The guide's `edit`, `render` and `verify` sections described the commands accurately. I did not need to guess a flag.

- **Confusing output:**
  - `cutroom transcript` prints a `[repeat]` tag on the first-take line and the `--suspects` view lists the same retake again. Two views of one problem, and I had to reconcile them.
  - `cutroom init` says `handled av-offset: audio starts +0.064 s relative to video`. The render then says `kept as recorded, so lip sync is as the camera captured it`, and `check` prints `ok sync ... median -0 ms, worst -0 ms (limit 45)`. A reader could take that `ok` as "lip sync is good". The sync check only compares the render's sound with the source's sound, so a constant 64 ms camera offset (over the 45 ms limit it names) passes silently.
  - `cutroom edit cut --dry-run` reports the cut as 6.230–9.430, while the WARN's `--near 6.31` suggested 6.314–9.214 in source time. The tool's padding explains it, but it was not stated, and I had to work it out.

- **Missing:**
  - A way to correct a constant audio offset on the camera's own track (the 64 ms lead). `cutroom sync` only handles a separate recorder, and `cutroom edit audio` only switches tracks. For this file, the only option is to leave it.
  - A picture-to-speech check. `check` compares the render with the source's sound, not the mouth in the picture. So "lip sync" can only mean "as recorded".
  - A pointer in `inspect` or `check` saying "this offset is over the 45 ms limit: the source itself may be out of sync". The 64 ms figure is printed but not judged.

- **Not sure it's right:**
  - Whether the +0.064 s camera offset is actually out of lip sync. I can't watch or hear the video, and the tool gives no picture-to-sound measurement, so I could not verify it.
  - The word check (`all 54 expected words heard`) used base.en, the same model that made the transcript. It confirms the cut, not the transcription, as its own text says.
  - I removed the false start and the 6 long pauses only. I kept the "my cat just walked across the keyboard" aside, since you asked for pauses and the false start, not off-topic removal. I also kept the 0.33 s sound at 0:16.974 out of the ~0.32 s pause that tighten removed with the pause.
  - Tightened pauses are 0.32 s each. That is the tool's default and I did not change it.

- **Best part:** `cutroom edit cut --from … --near … --to …` with a `--dry-run` preview, then `cutroom timeline --joins`, then `cutroom check`. The loop from pause-tightening to a verified render took about 20 commands with no surprises, and `check` caught the sync and word tests in one pass.

- **One change:** Add a constant-offset line to `inspect` and `check` that says whether the camera's own audio is off the picture, with the figure against the 45 ms limit and a fix option (for example `cutroom sync --offset` or `edit audio --shift`). That would have settled the lip-sync question for this run instead of leaving it as "kept as recorded".
