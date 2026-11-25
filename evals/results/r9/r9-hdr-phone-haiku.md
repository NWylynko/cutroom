# Friction report: r9-hdr-phone-haiku

- **Outcome:** done — out/final.mp4 is 13.8 s, passes every `cutroom check` line, and is tagged bt709 SDR, with the false start, the cat aside and the pauses removed. The only unverified part is the orientation and colour judgement by eye, since the clip is a synthetic test pattern.

- **Commands run:** about 30 `cutroom` invocations across 13 shell calls, plus one Read of a PNG and one ffprobe. In order: `cutroom guide`; `guide workflow`; `guide perceive`; `guide edit`; `guide render`; `guide verify`; `guide pitfalls`; `doctor`; `init clip.mov`; `inspect`; `transcribe`; `transcript`; `transcript --suspects`; `edit tighten`; `timeline`; `transcript --search "first thing you need"`; `transcript --search "cat"`; `edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`; `edit cut --from "By the way" --to "keyboard." --dry-run`; the same two cuts without `--dry-run`; `timeline`; `timeline --joins`; `listen 9.8-10.6`; `render --preview --dry-run`; `render --preview`; `look --output 0:02`; `check out/preview.mp4`; `render`; `check`; `transcript --file out/final.mp4`.

- **Errors hit:** none. No command failed and no `fix:` line was needed.

- **Guessed wrong:** nothing material. I expected `edit cut` to need a separate `--near` flag for the retake, and the suggested command from the `retake-kept` WARN worked as printed. I did not check whether `--dry-run` on `render` would show the HDR tone-map path; it printed only chunks and an estimate.

- **Confusing output:**
  - The `inspect` output says the clip is "rotated 180° by metadata" and the render says "baked into the pixels". Both are true, but the reader has to work out that they describe the same fix.
  - `timeline` labels the segments `s1, s3, s4, s6, s7` after the first cut, with no note that the numbering skips. It is harmless but reads as if segments went missing.
  - `listen` on a join printed "transcript: The first / heard: The first one." and a "!" line about different words. I had to read the note about cut-off words to decide it was fine.
  - `look` output is a PNG with a 2-line caption. The caption says `out 0:02.000 = clip 2.47`, which is useful, but I had to open the image to see the frame at all.

- **Missing:**
  - A way to check orientation and colour by eye that doesn't depend on the content. The clip is a colour-bar test pattern, so I could not tell from the picture alone whether it was upright; I relied on `check`'s `picture` and `format` lines and the rotation note.
  - A side-by-side view of the source frame and the rendered frame. The `tone` check gives numbers (saturation 87 vs 87, brightness range 217 vs 217), but I could not see the two frames together.
  - A note on how to judge a synthetic test pattern. I had no guidance on whether `check` treats it as a weaker test than real footage.

- **Not sure it's right:**
  - Orientation: `check` says "upright, right moment" and the format line says display rotation 0°, but I could not confirm by eye that the test pattern is the right way up.
  - Tone: `check` says the HLG→SDR conversion matches the source's saturation and brightness, but these are the tool's own measurements. I did not view the HDR source next to the SDR output.
  - Transcript: the words check re-transcribes with base.en, the same small model as the transcript. It confirms the cut, not the recognition. The listen result at 10.25 s shows base.en mishearing "The first one" for "The first" at a clip edge, so the word list itself may be off in places.

- **Best part:** `transcript --suspects` plus `timeline` and `timeline --joins` made the edit clear. The retake, the cat aside and the stray sound at 16.97 s were all flagged, and the suggested `--near` command worked as printed. `--dry-run` before each cut showed the join, so I could check the result before committing.

- **One change:** have `check` (or `inspect`) say in one line whether the picture is a test pattern or has enough detail to judge orientation, and offer a `look --side-by-side` of source and render frames for HDR sources. That would have let me confirm the colour and orientation fix without taking the tool's word for it.
