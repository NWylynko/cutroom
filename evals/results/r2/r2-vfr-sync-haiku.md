# Friction report (fill this in as the last thing you do)

- **Outcome:** partly done: `out/final.mp4` is rendered with the pauses and the false start removed, and audio sync measures 0 ms, but `cutroom check out/final.mp4` still exits 4 on a one-frame picture finding that I believe is a false positive from the variable frame rate (VFR) gap at the first cut, which I could not confirm numerically.
- **Commands run:** about 34 `cutroom` invocations across 18 Bash calls, plus 2 `ffprobe`, 3 `ffmpeg` (PSNR, failed), 1 `ls`, and 3 image reads. In order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide edit`, `cutroom guide pitfalls`
  3. `cutroom --help`
  4. `cutroom doctor`, `cutroom init talk-vfr.mp4`, `cutroom inspect`
  5. `cutroom transcribe`, `cutroom transcript`, `cutroom transcript --suspects`
  6. `cutroom edit tighten`, `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`
  7. `cutroom edit cut --from ... --to ...` (applied), `cutroom timeline`, `cutroom timeline --joins`
  8. `cutroom render --preview`, `cutroom check out/preview.mp4`
  9. `cutroom look 10.20 --source talk-vfr`, `cutroom look --help`, `cutroom look --cuts out/preview.mp4`
  10. `cutroom look 3.933 --file out/preview.mp4`, `cutroom look 10.164 --source talk-vfr`, `cutroom look --output 3.85-4.0 --count 6`
  11. `cutroom look 10.13 / 10.164 / 10.197 --source talk-vfr --crop 880,240,290,240 --width 400`, `cutroom look 3.933 --file out/preview.mp4 --crop 440,240,290,240 --width 400`
  12. `cutroom guide verify`, `cutroom guide render`
  13. `cutroom render --dry-run`, `cutroom render`, `cutroom check out/final.mp4`
- **Errors hit:** No cutroom errors. The `check` picture FAIL (exit 4) is the only failure. Its line read: `FAIL picture 1 of 18 sampled frames don't match their source frame; first at output 3.933 (segment s3): offset -2 frames, difference 0.0`. The `→ look at it: cutroom look 10.20 --source talk-vfr` hint pointed at the wrong frame. The `fix:` line did not apply; the check gives no fix for it.
- **Guessed wrong:**
  - I assumed `cutroom look --file X --crop ...` would crop the render. I never confirmed it did, and the output did not say.
  - I assumed the `handled vfr` line from `inspect` meant VFR could not cause picture mismatches. The check shows it can.
  - I assumed the `WARN[retake-kept]` from `tighten` (the 2× "the first thing you need is") was resolved by the cut. It was, since `timeline` then reported no problems.
- **Confusing output:**
  - `inspect` labels VFR and the audio offset as `handled`, yet `check` still flags a picture mismatch that traces to VFR. Nothing says which source frame a time maps to when the source drops frames.
  - Source frame 10.133 is followed by 10.233 (a 0.1 s dropped-frame gap). `look 10.164` reports "frame 305" by nominal 30 fps index, which is not the frame actually shown. The `frame N` labels are nominal, not real, and that is not stated.
  - The `check` offset of -2 frames reads as a sync bug. It is really the held frame during the gap.
  - `cutroom --help` printed the top-level command list only. It did not show `look`'s flags, so I needed `cutroom look --help` separately.
- **Missing:**
  - A way to see the real decoded frame timestamp (pts) at a given source time, so I could tell held frames from nominal ones.
  - A `check` option or note that accepts VFR holds, or a per-frame report that says whether the expected frame was held or dropped.
  - A `look --file` crop confirmation, and an option to compare two frames numerically from the CLI. My raw ffmpeg PSNR attempt printed nothing.
- **Not sure it's right:**
  - I believe the picture FAIL is a checker false positive caused by the dropped frame at source 10.133 to 10.233, because the render at output 3.933 matches the source frame 10.133 exactly (difference 0.0 at the reported offset). I did not confirm this with a working numeric comparison; my ffmpeg PSNR command returned nothing.
  - The final still fails `check`, so I am not claiming a clean pass.
  - Audio sync (0 ms median and worst across 6 segments) is reported by the tool, and I took it at face value. The source audio starts +0.064 s and was handled by the renderer. I did not measure it independently.
  - The transcript used the `base.en` model, which the tool warned may mishear words. The word check passed (54 of 54), but that depends on the same model.
  - The cat aside ("By the way, my cat just walked across the keyboard") is still in the video. The task said to remove pauses and the false start only, so I kept it. I did not ask.
- **Best part:** `tighten` followed by a dry-run of the cut, then `timeline --joins`, gave a clear picture of the edit in four commands. The `--dry-run` on the cut and the `timeline` join view made the false-start fix easy to check before applying it.
- **One change:** Make `check`'s picture test account for VFR dropped frames by mapping each output frame to the source frame actually shown (the held one), and print both the held and the nominal frame. That would make the FAIL either pass or point at a real problem, and it would remove the misleading `look 10.20` hint.
