# Friction report: r5-vfr-sync-haiku

- **Outcome:** done. `out/final.mp4` (17.233 s, 1280x720) renders with all six sections of the timeline in place, the long pauses shortened, and the "The first thing you need is, UM, the first thing." false start cut. `cutroom check out/final.mp4` passes with 0 warnings, including sync (median and worst -0 ms, limit 45 ms). One caveat: I kept the cat aside, which the request did not mention.

- **Commands run:** 11 shell calls, 27 cutroom invocations, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide perceive`, `cutroom guide edit`
  3. `cutroom guide render`, `cutroom guide verify`, `cutroom guide pitfalls`
  4. `cutroom doctor`
  5. `cutroom init talk-vfr.mp4`
  6. `cutroom inspect`
  7. `cutroom transcribe`, `cutroom transcript`, `cutroom transcript --suspects`
  8. `cutroom edit tighten`, `cutroom timeline`, `cutroom timeline --joins`
  9. `cutroom edit cut --from "the first thing you need is" --near 6.31 --to "first thing." --dry-run`
  10. `cutroom edit cut --from "the first thing you need is" --near 6.31 --to "first thing."`, `cutroom timeline`, `cutroom timeline --joins`
  11. `cutroom render --preview --dry-run`, `cutroom render --preview`, `cutroom check out/preview.mp4`
  12. `cutroom render`, `cutroom check out/final.mp4`, `cutroom status`
  13. `cutroom look --cuts --file out/final.mp4`, then the PNG was opened and viewed

- **Errors hit:** None. No command failed and no `fix:` line was needed.

- **Guessed wrong:** Nothing I had to correct. I did not know whether `--near` would accept a source time in seconds. It did (`--near 6.31`), and the printed suggestion already used that form.

- **Confusing output:**
  - `cutroom inspect` repeats the `handled vfr` and `handled av-offset` lines that `init` had already printed. Nothing is wrong, but it is noise when you are checking lip sync specifically.
  - `cutroom edit tighten` removed the suspect sound at 0:16.974 (0.33 s, "a cough, bump, breath…") as a side effect. Its output says "this also removes the sound with no words", but it does not say it was never listened to. The `transcript --suspects` guidance says to `cutroom listen` a suspect before keeping or dropping it. I did not listen to it, so the guide and the tool's behaviour disagree.
  - The `WARN[retake-kept]` line from `tighten` was useful. It printed a ready `edit cut` command, and `--dry-run` showed the join.

- **Missing:**
  - A way to check that the audio-to-picture offset is handled in the source itself. `check`'s `sync` compares the render's sound against the source's own sound at the moment its picture shows. That proves the render keeps the source's sync, but not that the source was in sync to begin with. For a lip-sync complaint, the source's own sync is what matters, and I had no tool-level check for it.
  - A way to ask `tighten` to keep specific suspects, or to list what it would remove before running it.

- **Not sure it's right:**
  - Lip sync. The `sync` check passed with a median of -0 ms, and `inspect` reported `audio starts +0.064 s relative to video` as "handled". I did not verify the source's own lip sync independently. I could only confirm that the render matches the source's audio at each segment.
  - Whether the 0:16.974 sound was a cough, a bump, or a skipped word. I removed it by accident through `tighten` and never listened to it.
  - The `base.en` transcript. Its word list has 54 words in the render check, and `check` says they were all heard in order. I did not read the speech myself beyond the transcript.
  - The cat aside ("By the way, my cat just walked across the keyboard.") is still in the video. The request asked only for pauses and the false start, so I left it. If the author wants it gone, it is one `edit cut --from "By the way" --to "keyboard."`.

- **Best part:** The sequence from `init` to `check` was fast and needed few decisions. `tighten` finished most of the work, `timeline` and `timeline --joins` showed exactly where the joins were, and the `retake-kept` warning gave a copy-pastable fix. `check` then covered sync, words and loudness in one table.

- **One change:** Have `tighten` (and `cut`) print the suspects it removes and whether each one was listened to, or stop before removing any unlistened suspect sound. That would have caught my unlistened 0:16.974 removal.
