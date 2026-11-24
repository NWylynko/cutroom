# Friction report: r5-hdr-phone-haiku

- **Outcome:** done — `out/final.mp4` (1280x720, H.264 bt709, 13.767 s) passes `cutroom check` with 0 failures and 0 warnings, and the false start ("The first thing you need is, UM, the first thing."), the cat remark and the long pauses are gone. Colour is only verified by eye on a synthetic test pattern (see Not sure).
- **Commands run:** about 33 cutroom invocations, plus one `ffprobe`, one `ls`, and three image reads. In order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `edit`, `render`, `verify`, `pitfalls`, `perceive` (one bash call)
  3. `cutroom doctor`
  4. `cutroom init clip.mov`
  5. `cutroom inspect`
  6. `cutroom transcribe`, `cutroom transcript`, `cutroom transcript --suspects`
  7. `cutroom edit tighten`, `cutroom transcript --search "first thing"`, `cutroom listen 16.5-17.8`
  8. `cutroom edit cut --from "The first thing you need is" --occurrence 1 --to "the first thing" --dry-run` (this does not match: see Guessed wrong)
  9. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing."`, then the cat cut with `--dry-run`, then applied
  10. `cutroom timeline`, `cutroom timeline --joins`
  11. `cutroom render --preview --dry-run`, `cutroom render --preview`, `cutroom check out/preview.mp4`
  12. `cutroom look 2`, `cutroom look --file out/preview.mp4 3`, `cutroom look --file clip.mov 2`
  13. `cutroom render`, `cutroom check out/final.mp4`, `ffprobe` on out/final.mp4
  14. `cutroom look --file out/final.mp4 1`, `cutroom look --cuts`, `cutroom status`
- **Errors hit:** none. No command failed, and no `fix:` line was needed.
- **Guessed wrong:**
  - I assumed `--occurrence` would resolve the repeated "The first thing you need is" in one go. My dry-run used it, and the cut that actually went in used a longer, unique `--from` instead. The guide does say `--from` must be unique, so this was a guess and not a bug.
  - `cutroom timeline` said `retake-kept` and printed a ready-made `edit cut --from "the first thing you need is" --near 6.31 --to "first thing."`. I did not run it. I cut by my own quotes, because I wanted the cut to remove the whole false start, including "The first thing you need is, UM, the first thing.", which the printed command would have left at the front. The suggestion is sound, but I did not verify that it would do what I wanted before choosing my own cut.
- **Confusing output:**
  - Tighten reported "this also removes the sound with no words at source 0:16.970–0:17.290". That is accurate, but it happened as a side effect, so I had to check that the stray sound was really gone (I checked with `check`, which passed).
  - `cutroom look --cuts` labels the join "before (clip 4.38)" while `timeline` lists the segment as ending at clip 4.400. Both are plausible, but I had to think about why they differ.
  - `cutroom render` and `check` say nothing about HDR. Their picture lines cover rotation only ("clip's 180° display rotation is baked into the pixels"). The HDR-to-SDR conversion is reported by `inspect` ("handled hdr: ... a plain SDR conversion looks washed out") and then never mentioned again, so I could not tell from render or check output that it had been applied.
- **Missing:**
  - A check or render line that says HDR was converted to SDR (for example: "HLG bt2020 10-bit tone-mapped to bt709 SDR; mean luma X vs source Y"). Without one I have to trust that it happened.
  - A quantitative colour test. `check` tests frames, sync, words, loudness and format. It does not test colour, so "not washed out" can only be judged by eye.
  - A short flag or line that says which `--from` matches are repeated, before I have to run a dry-run to see them.
- **Not sure it's right:**
  - Colour and the HDR conversion. The source turned out to be a synthetic SMPTE-style colour-bar test pattern with a moving bar and a checker box, not a real scene. I checked the source frame at 2 s, the preview at 3 s, and the final at 1 s, and the bars looked saturated and correct, with the box on the right in every case. I did not compare skin tones or highlights, and a test pattern cannot show a subtle washed-out or crushed look. Treat the colour as plausible, not proven.
  - Orientation. The source metadata says 180°, and the frame checks show the box on the right in source, preview and final, which matches the 180° rotation being applied. I verified this only by comparing frames, not against a reference.
  - Transcript words. Transcription used `base.en`, and `check` compares words against the same model's transcript, so the `words` PASS shows the render matches the timeline, not that the transcript is correct. I did not `listen` to the false start or to the joins.
- **Best part:** The `guide` topics were short and specific, and `transcript --suspects` pointed straight at the repeat and the stray sound. Each cut has `--dry-run`, and `timeline --joins` plus `check` made the verify loop quick.
- **One change:** Add an HDR line to `render` and `check`, for example "HDR HLG bt2020 → bt709 SDR tone-mapped; luma compared with source", so an agent can confirm the conversion happened and is not washed out.
