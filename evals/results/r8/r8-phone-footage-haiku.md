# Friction report: r8-phone-footage-haiku

- **Outcome:** done. `out/final.mp4` (1280x720, 13.767 s) passes `cutroom check` with 0 warnings. The retake ("The first thing you need is, UM, the first thing.") and the cat remark were cut, and the pauses were tightened to 0.32 s. `ffprobe` shows no rotate tag and `check` reports display rotation 0°.

- **Commands run:** about 37 cutroom invocations across 10 shell calls, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide perceive`, `cutroom guide edit`
  3. `cutroom guide render`, `cutroom guide verify`, `cutroom guide pitfalls`, `cutroom doctor`
  4. `cutroom init talk-rot.mov`, `cutroom inspect`, `cutroom transcribe`, `cutroom transcript`
  5. `cutroom transcript talk-rot --suspects`, `cutroom transcript talk-rot --words --from 6 --to 12`, `cutroom edit tighten`, `cutroom timeline`, `cutroom timeline --joins`
  6. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`, `cutroom edit cut --from "By the way" --to "keyboard." --dry-run`
  7. `cutroom edit cut ...` (both, applied), `cutroom timeline`, `cutroom timeline --joins`, `cutroom listen 9.0-10.6`, `cutroom listen 10.6-12.2`, `cutroom listen 16.3-17.0`
  8. `cutroom render --preview --dry-run`, `cutroom render --preview`, `cutroom look out/preview.mp4 2.0`, `cutroom check out/preview.mp4`
  9. `cutroom render --dry-run`, `cutroom render`, `cutroom check`, `ffprobe` (stream tags and side data), `ls out/`, then a failed `cutroom look talk-rot 3.0`
  10. `cutroom look 3.0`, `cutroom look out/final.mp4 2.0`, `ffprobe` (side data only)

- **Errors hit:**
  - `error[bad-args]: "talk-rot" is not a file and look takes one time or range` / `fix: one time or range, optionally after a file: cutroom look out/final.mp4 3-5`. The fix line was enough. I dropped the source id and used `cutroom look 3.0`, which worked.

- **Guessed wrong:**
  - I assumed `cutroom look <source-id> <time>` would work the way other commands take ids. It does not: `look` takes only a time, or a file followed by a time. The guide does say this, but I did not read the `look` section before using it.
  - I assumed the `inspect` line "rotated 180° by metadata" meant a phone-style 90° rotation. I did not check it further, so I cannot say whether 180° is the real orientation.

- **Confusing output:**
  - `cutroom listen 9.0-10.6` printed `! the isolated clip heard different words than the transcript`. The clip covers a span that was already cut, so the mismatch is expected. The warning fires on content that is no longer in the output, which made it hard to tell whether anything was wrong.
  - The `retake-kept` WARN from `tighten` gives a ready `edit cut` command. Its `--near 6.31` is a source time, while the dry run reported the removed span as `0:06.200–0:09.433`. Both are correct, but the difference is easy to misread.
  - `cutroom timeline` and `--joins` print output times in the first column and source ranges after, which is handy but needs a second look when you are checking which take survived.

- **Missing:**
  - A side-by-side view of source frame and output frame at matching content, so orientation can be compared without opening two PNGs.
  - A way for `look` to show the source frame as stored (before rotation), so I could confirm the 180° rotation was really what the file needed.
  - A listen or look shortcut for the join at the cat-remark cut (output 0:08.13). I checked the timeline joins text but did not listen to that join.

- **Not sure it's right:**
  - The picture is a SMPTE-style test pattern, not a talking head. `check` says the picture matches its source frames and is upright, but I could not judge the speaker or framing by eye.
  - I checked the orientation only by comparing one source frame (`look 3.0`, rotation applied) with one output frame (`look out/final.mp4 2.0`). I did not confirm the rotation against the phone's intent.
  - The word check passed, but it uses base.en, which the guide says confirms the cut and not the recogniser.
  - I did not listen to the join at output 0:08.13 (after the cat remark was removed). `timeline --joins` shows only a 0.33 s pause there, and `check` passed, so it is probably fine.

- **Best part:** `cutroom edit cut ... --dry-run` plus `cutroom timeline --joins`. The dry run showed the removed span and the new join before anything changed, and the joins list read like a viewer would hear it. The `check` output also stated plainly that the rotation was baked into the pixels with no rotation metadata, which answered the playback requirement directly.

- **One change:** Make `cutroom look` accept a source id, or print a clear hint that says "use a time, not a source id" with the exact command for the current project. That would have avoided the one failed call.
