# Friction report: r4-vfr-sync-haiku

- **Outcome:** done. `out/final.mp4` (17.233 s) passes `cutroom check` with 0 warnings: the long pauses are tightened, the false start is cut, and sync measured median and worst 0 ms.

- **Commands run:** 24 cutroom invocations across 8 shell calls, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `guide edit`, `guide render`, `guide verify`, `guide pitfalls`, `guide perceive`
  3. `cutroom doctor`, `cutroom init talk-vfr.mp4`, `cutroom inspect`
  4. `cutroom transcribe`, `cutroom transcript`, `cutroom transcript --suspects`
  5. `cutroom edit tighten`, `cutroom edit cut --from "is, UM, the first thing" --to "the first thing" --dry-run`, `cutroom timeline`
  6. `cutroom edit cut --from "the first thing you need is" --near 6.31 --to "first thing." --dry-run`, then the same without `--dry-run`, then `cutroom timeline --joins`
  7. `cutroom render --preview --dry-run`, `cutroom render --preview --range 0:03-0:12`, `cutroom render`, `cutroom check out/final.mp4`
  8. `cutroom look --cuts --file out/final.mp4` (then read the PNG)

- **Errors hit:** none. No command failed.

- **Guessed wrong:**
  - My first `--from "is, UM, the first thing" --to "the first thing"` dry-run would have cut 0:07.530–0:09.430. That left "The first thing you need is" in the output in front of the second take. I only caught it because I dry-ran it. The `retake-kept` warning's suggested command, with `--near 6.31`, was the correct one.
  - The `retake-kept` warning from `tighten` says the cut is "a retake?" and suggests "keep the last take". It does not say that the retake's first take is the one to remove. Reading the join output was needed to confirm it.
  - I expected `cutroom timeline` to show the true output-to-source mapping after a cut. It does, but the per-segment times mix output and source columns, and I had to re-read them.

- **Confusing output:**
  - `cutroom timeline` prints the `retake-kept` warning both before and after the cut, and the second copy was the same text. It was not clear whether it was stale.
  - `cutroom edit cut --dry-run` reports "would cut talk-vfr 0:06.230–0:09.430", which is a source range, but `timeline` shows output times. I had to keep the two straight by hand.
  - The `sync` line in `check` is long (about 230 characters). It is useful, but it is hard to scan next to the other lines.

- **Missing:**
  - A way to tell `tighten` to skip a named span, or to keep one specific aside. I left "By the way, my cat just walked across the keyboard" in the output. The task only said to remove long pauses and the false start, so I kept it, but I had no quick way to ask the tool whether that aside should go.
  - `cutroom check` does not say whether the audio/video offset from the source (+0.064 s) was applied. It only says "its audio start offset included" in the sync message. I trust that wording, but I could not check it.

- **Not sure it's right:**
  - Sync: `check` reports 0 ms median and worst across 6 segments, and I looked at the cut-pair contact sheet. I did not listen to any join (`cutroom listen`), so sync is verified by the tool's measurement, not by my own ear.
  - Transcript: the base.en model misheard nothing I could see, but its "UM" and the aside were read from its text only. I did not confirm the 0:16.974 sound, which `tighten` removed along with the long pause.
  - The 0.32 s pauses that `tighten` left look even in the cut-pair image, but I did not check the audio across them.

- **Best part:** `cutroom transcript --suspects` flagged the retake and the stray sound directly, and `cutroom check` gave a clear per-line PASS/WARN list with exact numbers. The `--dry-run` on `cut` was the single most useful feature for avoiding a bad edit.

- **One change:** Make `retake-kept` say which take to remove, not only which to keep. For example, "cut 0:06.230–0:09.430 (the first take, with `--near`)". That would have saved me the dry-run comparison.
