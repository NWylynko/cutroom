# Friction report: r3-recorder-haiku

- **Outcome:** done. `out/final.mp4` (17.2 s, 1280x720) uses the recorder's audio aligned to the camera, has the long pauses tightened, and the first retake of "The first thing you need is" is cut. `cutroom check out/final.mp4` passed with 0 failures and 0 warnings.

- **Commands run:** 25 cutroom invocations, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide edit`, `cutroom guide pitfalls`
  3. `cutroom guide perceive`, `cutroom guide render`, `cutroom guide verify`
  4. `cutroom doctor`
  5. `cutroom init talk.mp4 --source mic=recorder.wav`
  6. `cutroom inspect`
  7. `cutroom transcribe`
  8. `cutroom sync mic`
  9. `cutroom edit audio mic`
  10. `cutroom transcript talk`
  11. `cutroom transcript talk --suspects`
  12. `cutroom edit tighten`
  13. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`
  14. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing."`
  15. `cutroom timeline`
  16. `cutroom timeline --joins`
  17. `cutroom render --preview --dry-run`
  18. `cutroom render --preview`
  19. `cutroom check out/preview.mp4` (run as `cutroom check`, which picked the preview)
  20. `cutroom render`
  21. `cutroom check`

  Plus two plain shell `ls` calls.

- **Errors hit:** none. No command failed, so no `fix:` line was needed.

- **Guessed wrong:** nothing failed. I did not have to guess a flag. I assumed `--to` would match the first occurrence after `--from`, as the guide says, and the dry run confirmed it. `cutroom check` with no argument silently chose `out/preview.mp4` after the preview render. That was fine, but I had to read the output to see which file it checked.

- **Confusing output:**
  - After `edit audio mic`, the `retake-kept` warning said "still said 2×" and the `kept-sound` warning gave a sound at 16.97. After `edit tighten`, the sound was no longer reported. It had been removed by the tighten pass, but nothing said so. I had to compare `timeline` output to work it out.
  - The `retake-kept` warning suggests `cutroom transcript --search "the first thing you need is"`. That search returns every occurrence, so I still had to decide which take to keep and work out the `--from`/`--to` text myself. The `--from` text needed "UM" to be unique, which the guide mentions but the warning does not.
  - `check` prints the sync line as "render audio vs mic (aligned to talk by sync)". That is accurate but wordy. I could not tell at a glance what it measured.

- **Missing:**
  - A way to see the cut points' words in the preview before rendering the final. `timeline --joins` covers this, but a `--joins` view that also shows the sound's source (camera or mic) would help.
  - Nothing in `sync` or `check` shows the residual between the camera's own sound and the recorder after alignment. `sync` reports "fit: residual 0.02 ms" and `check` reports "worst -6 ms", but these measure different things.
  - No guidance on whether asides ("By the way, my cat just walked across the keyboard") should be cut. The guide mentions them under the edit sequence, but I was not told to cut them, so I kept it.

- **Not sure it's right:**
  - I did not hear or watch the output. The `check` results are measurements, and I read them as passing.
  - The sync check compares the render's audio with `mic`. The offset came from `sync mic`, which used the camera's own sound. So the check confirms the render matches the alignment, not that the picture matches the recorder's sound by an independent measure. I could not verify this with the tool.
  - I used `base.en` transcripts. The retake and the sound at 16.97 were judged from that transcript and the suspects list. I did not run `cutroom listen` on the sound at 16.97 because the tighten pass had already removed it from the timeline, so I did not check whether it was words.
  - Word count check: 54 expected words in the render after cutting the 10-word retake (64 - 10). That number is consistent but I did not compare it to a separate reading.
  - I kept the "By the way, my cat..." aside. The brief only asked for pauses to be removed, so I did not remove it. That is a judgement call.

- **Best part:** `cutroom sync mic` followed by `cutroom edit audio mic` made the separate-recorder case a three-command setup, and `render` plus `check` ran cleanly. The `--dry-run` on `edit cut` showed the exact span and the join before I applied it.

- **One change:** have the `retake-kept` warning print a ready-to-run `cutroom edit cut --from "…" --to "…"` line for the weaker take, with `--from` already made unique (for example by including "UM"), instead of a search command that lists every occurrence.
