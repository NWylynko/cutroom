# Friction report: talk-vfr, pauses and false start, lip sync

- **Outcome:** done — out/final.mp4 is rendered (17.233 s, 6 segments), `cutroom check` PASSes with 0 warnings, and the pauses and the "The first thing you need is, UM…" false start are removed. Lip sync is verified only relative to the source file, not against the speaker's mouth (see Not sure).

- **Commands run:** 11 shell calls, about 27 `cutroom` invocations, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide edit`, `cutroom guide verify`
  3. `cutroom doctor`
  4. `cutroom init talk-vfr.mp4`, `cutroom inspect`
  5. `cutroom transcribe`, `cutroom guide perceive`, `cutroom transcript`
  6. `cutroom transcript --suspects`, `cutroom transcript --retakes`, `cutroom edit tighten`, `cutroom timeline`, `cutroom timeline --joins`
  7. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`
  8. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."`, `cutroom timeline`, `cutroom timeline --joins`, `cutroom timeline --accept aside-kept@11.46 --reason "…"`
  9. `cutroom render --help`, `cutroom render`
  10. `cutroom check out/final.mp4`
  11. `cutroom transcript --output`, `cutroom status`

- **Errors hit:** None. Every command exited 0 on the first try.

- **Guessed wrong:** Nothing failed, but two things were not obvious.
  - `init` reports `handled av-offset: the sound track starts 0.064 s after the picture track … not a lip-sync error`. That is a claim I could not check. The render keeps the 64 ms offset unchanged, so the output matches the source's sync, but the tool does not measure whether the mouth and the sound agree.
  - `--accept aside-kept@11.46` was given before the false-start cut, when the warning sat at 11.46. The warning moved to 8.26 after the cut. The accept still matched and reported `accepted aside-kept at output 0:08.257`, so the time in the key appears to be used loosely. I did not expect that.

- **Confusing output:**
  - `timeline` prints the same WARN twice in one run: once at the cut and again in `--joins`. Not a problem, but noisy.
  - `cutroom guide edit` is about 1,500 words. The parts I needed were the cut/`--near`/`--dry-run` lines and the WARN/accept rules, which are buried in the middle.
  - `init` says "handled" for the VFR and av-offset items. The word "handled" reads as "cutroom verified this," but for av-offset it only means "kept as recorded."

- **Missing:**
  - A way to measure mouth-to-sound sync, or at least a reading of the camera's own lip-sync offset. `check` compares the render against the source's sound, so it would not catch an offset that was already in the source. That is the exact failure the person described ("slowly went out of lip sync").
  - Confirmation of which offset is the true lip-sync offset. I could not tell from the output whether the 64 ms is a container start-time artifact or a real A/V lag.

- **Not sure it's right:**
  - Lip sync: `check` says `sync ... median -0 ms, worst -0 ms`, with the 64 ms offset included. That shows the render matches the source. It does not show the source itself is in sync with the speaker's face. I left the 64 ms offset as recorded because the tool says to, and because I could not measure mouth-to-sound sync myself. If the person's concern is a real 64 ms lag in the camera recording, this render keeps it.
  - Transcript: base.en mishears words. `check` confirmed the 54 words in order, but that check used base.en too, so it confirms the cut, not the recognition. I did not re-transcribe with a larger model.
  - Aside: I accepted the "by the way, my cat…" aside (`aside-kept`) and kept it in. The request asked only for pauses and the false start to go, so I did not cut the aside. It is a judgment call, and the person may want it removed.
  - Pauses: `tighten` cut six pauses down to 0.32 s and also removed the cough-like sound at 0:16.974. That was intended.

- **Best part:** The `edit cut --dry-run` step. It showed the removed span, the new join, and the new total before I changed anything, so the false-start cut was safe to apply. `check` with a per-segment sync line also made the render easy to trust.

- **One change:** Make `init` say what it verified for av-offset. Something like "kept as recorded: cutroom cannot measure lip sync against the face; if the mouth and the sound disagree, compare the camera's own clap or slate" would have told me exactly what remained unverified, instead of "not a lip-sync error."
