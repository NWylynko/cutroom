# Friction report

- **Outcome:** done. out/final.mp4 is 13.8 s, rotation baked in (check: display rotation 0°), all checks PASS.
- **Commands run:** 17 in total.
  1. `guide`
  2. `guide workflow; guide pitfalls`
  3. `init talk-rot.mov` + `inspect` + `transcribe` + `transcript` (one chained shell call)
  4. `guide edit`
  5. `edit tighten`
  6. `edit cut --from "By the way" --to "keyboard"`
  7. `edit cut --from "The first thing you need is, UM" --to "the first thing."`
  8. `edit cut 16.9-17.1` (errored)
  9. `timeline --joins`
  10. `render --help`
  11. `render --output out/final.mp4` (errored)
  12. `render`
  13. `check`
  14. `look --help`
  15. `look out/final.mp4` (errored)
  16. `look 2 --file out/final.mp4`
  17. Read of the PNG
- **Errors hit:**
  - `error[nothing-cut]: nothing to cut in talk-rot 16.9-17.1`. I tried to cut the sound at 16.97. Its `fix:` line explained the sound was already gone, which got me unstuck.
  - `error[unknown-flag]: render has no flag --output`. The `fix:` line sent me to `--help`, but the default path is already out/final.mp4, so no flag was needed.
  - `error[bad-time]` from `look out/final.mp4`. The `fix:` line was fine, but it did not hint that `--file` is the flag I wanted.
- **Guessed wrong:** `render --output`; `look <file>` taking a path positionally; I assumed the stray sound needed a manual cut (tighten had already dropped it).
- **Confusing output:** The first edit printed a `retake-kept` WARN right after tighten. The same WARN repeated after the cat cut and cleared only after the retake cut. That is correct, but it was noisy. Also, the transcript shows "(contains a sound at 0:16.970)" in a pause, and nothing said that tighten removed it. I only inferred that from the nothing-cut error.
- **Missing:** A "rotation baked in / plays upright" line in the render output itself (check shows `display rotation 0°`, which is enough, but the render output is silent about it). A `check`-style note that the stray sound was removed would also help.
- **Not sure it's right:** I judged the picture orientation from one frame with `look --file`. The check's "display rotation 0°" is the real proof for players that ignore rotation metadata. I could not test an actual player.
- **Best part:** The transcript's `[repeat]` tags and the quoted-text `edit cut --from/--to` made the retake and the cat aside trivial. `check` covered everything in one command.
- **One change:** Make `look` accept a rendered file path positionally, or have the bad-time fix line mention `--file`.
