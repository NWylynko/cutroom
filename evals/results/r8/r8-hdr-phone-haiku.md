# Friction report: r8-hdr-phone-haiku

- **Outcome:** done. `out/final.mp4` passes `cutroom check` with 0 failures and 0 warnings, and ffprobe shows H.264, yuv420p, bt709 tags, 1280x720, no rotation metadata. The false start, the cat remark and the pauses are gone, leaving 13.767 s of video.

- **Commands run:** 29 cutroom commands, plus `ls`, `ffprobe`, and two image reads of the `look` PNGs.
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom guide perceive`
  4. `cutroom guide pitfalls`
  5. `cutroom doctor`
  6. `cutroom init clip.mov`
  7. `cutroom inspect`
  8. `cutroom guide edit`
  9. `cutroom guide render`
  10. `cutroom guide verify`
  11. `cutroom transcribe`
  12. `cutroom transcript`
  13. `cutroom transcript --suspects`
  14. `cutroom edit tighten`
  15. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`
  16. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing."`
  17. `cutroom edit cut --from "By the way, my cat" --to "across the keyboard." --dry-run`
  18. `cutroom edit cut --from "By the way, my cat" --to "across the keyboard."`
  19. `cutroom timeline`
  20. `cutroom timeline --joins`
  21. `cutroom render --preview --dry-run`
  22. `cutroom render --preview`
  23. `cutroom check out/preview.mp4`
  24. `cutroom look out/preview.mp4 6.0`
  25. `cutroom look clip.mov 5.0`
  26. `cutroom render --dry-run`
  27. `cutroom render`
  28. `cutroom check`
  29. `cutroom transcript --file out/final.mp4`

- **Errors hit:** None. No command failed, and no `fix:` line was needed.

- **Guessed wrong:** Nothing outright. The `retake-kept` warning suggested `--from "The first thing you need" --near 6.31 --to "first thing."`. I used the exact transcript phrase instead, and its dry run showed the same removal (0:06.200–0:09.433), so the suggestion was not needed.

- **Confusing output:**
  - `cutroom look` labels its output "file time" for a file path and "output time" for `--output`. I had to think about which time base I was in, and I had to open the PNG to confirm the rotation was applied. The text line alone does not say the picture is upright.
  - `cutroom edit cut --dry-run` prints the padded range (0:06.200–0:09.433) and not the exact words' range. That is correct, but it reads as if more was cut than the phrase you named.
  - Every render prints the same HDR and rotation `picture:` lines, even on the second render. They are useful the first time and noise after that.
  - The `check` tone line ("saturation 87 vs source 87, brightness range 217 vs 217") is precise but hard to interpret without a reference. I could not tell from it whether the result looks right to a person.

- **Missing:**
  - A before/after image for HDR. `cutroom look` shows one frame. I wanted the source frame and the SDR render side by side, to judge the tone mapping by eye.
  - A way to listen to the 0.32 s sound at 0:16.97 before tighten removed it. `transcript --suspects` said "`cutroom listen` it", but I did not run it. `tighten` removed the sound silently, and the output only mentioned it in the middle of the tighten summary.
  - A larger transcription model. The guide said `cutroom doctor` shows how to get one, but I did not check. All transcript words and the "words" check used `base.en`.

- **Not sure it's right:**
  - SDR appearance is checked only by ffprobe tags, `check`'s tone statistics, and my own look at two frames. The frames were a synthetic test pattern with no skin tones, so I did not verify how faces or colour would look in YouTube or another player.
  - I did not run `cutroom listen` on any join or on the 0:16.97 sound. The joins read cleanly in `timeline --joins`, and the `words` and `stray-sound` checks passed, but I did not listen.
  - The word check used `base.en`, the same model that made the transcript. It confirms the cut, not the transcript's accuracy.

- **Best part:** The workflow from `guide workflow` to `check` mapped directly onto the task. The `--dry-run` on every cut showed the join and the new total before I applied it, and `check` passed on the first render with no fixes needed.

- **One change:** Add a side-by-side HDR check to `cutroom look`, for example `cutroom look --compare clip.mov out/final.mp4 5.0`. It would put the source and the SDR render in one image, labelled, so an agent could judge tone mapping by eye and not only from a saturation number.
