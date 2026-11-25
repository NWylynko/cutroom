# Friction report: r9 phone footage (haiku)

- **Outcome:** done. `out/final.mp4` (13.767 s, 1280x720) passes `cutroom check` with 0 failures and 0 warnings. The retake ("The first thing you need is, UM, the first thing.") and the cat remark are cut, and the pauses are tightened to 0.32 s.

- **Commands run:** 19 cutroom edit/render/check commands and 7 `cutroom guide` topics (overview, workflow, perceive, edit, render, verify, pitfalls). Non-cutroom: `ls`, one `ffprobe` on out/final.mp4 to confirm no rotation side data or tag, and one Read of the cuts PNG. In order:
  1. `cutroom doctor`
  2. `cutroom init talk-rot.mov`
  3. `cutroom inspect`
  4. `cutroom transcribe`
  5. `cutroom transcript`
  6. `cutroom transcript --suspects`
  7. `cutroom edit tighten`
  8. `cutroom timeline`
  9. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`
  10. `cutroom edit cut --from "By the way" --to "keyboard." --dry-run`
  11. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."`
  12. `cutroom edit cut --from "By the way" --to "keyboard."`
  13. `cutroom timeline`
  14. `cutroom timeline --joins`
  15. `cutroom render --preview --dry-run`
  16. `cutroom render --preview`
  17. `cutroom look --cuts`
  18. `cutroom render`
  19. `cutroom check`

- **Errors hit:** None. No command failed and no `fix:` line was needed.

- **Guessed wrong:** Nothing I assumed was wrong. I did expect `timeline`'s `retake-kept` warning to need `--apply`, but I used the printed `edit cut` line directly, and that worked.

- **Confusing output:**
  - After the cuts, segment ids skip (s1, s3, s4, s6, s7). That is correct, but it reads like something went missing until you check the timeline.
  - `cutroom inspect` says "rotated 180° by metadata" and `check` says "display rotation 0°". Both are right, but a first-time reader has to work out that the render bakes the rotation into the pixels and writes no tag.
  - `check`'s `format` line reports the display rotation of the output, not the raw stream tags. I could only confirm "no rotation metadata" by running ffprobe myself.

- **Missing:**
  - A way to see the rotation tags of the output file from cutroom. I had to fall back to ffprobe.
  - A one-step way to remove a repeated take and an aside together. Two `edit cut` calls worked fine, so this is minor.

- **Not sure it's right:**
  - The transcript came from `base.en`, so I could not check word accuracy beyond what `check` reports. `check` re-transcribed with the same model, so it confirms the cut but not the recogniser.
  - I chose to keep the second take of "The first thing you need is…" and drop the first, which has the "UM". I judged this from the transcript text and did not listen to the audio around 6.2–9.4 s with `cutroom listen`.
  - I viewed the cuts contact sheet but did not view frames of the final render with `cutroom look --output`. The test-pattern picture in the source makes faces irrelevant, and `check`'s picture check passed (upright, right moment).
  - "Plays correctly in any player, including ones that ignore rotation": I confirmed there is no rotation metadata in out/final.mp4, and that the picture is upright in the stored pixels. I did not open it in a player that ignores rotation.

- **Best part:** The `timeline` WARN printed a ready-to-run `edit cut` command with its source range and join preview, and `--dry-run` showed the join before applying. `check` printed one line per check, so the PASS was easy to read.

- **One change:** Add the raw stream rotation tag to `check`'s `format` line (for example `rotation tag: none, display rotation 0°`), so an agent can confirm the "no rotation metadata" promise without running ffprobe.
