# Friction report: r10-hdr-phone-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s) passes `cutroom check` with 0 failures and 0 warnings. The false start, the cat aside and the pauses are gone. The HDR-to-SDR colour is only confirmed by cutroom's own self-comparison plus one visual look (see Not sure it's right).

- **Commands run:** 13 shell invocations, about 22 cutroom commands, in this order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `guide perceive`, `guide edit`
  3. `cutroom guide render`, `guide verify`, `guide pitfalls`
  4. `cutroom doctor`, `cutroom init clip.mov`, `cutroom inspect`
  5. `cutroom transcribe`, `cutroom transcript`, `cutroom transcript --suspects`
  6. `cutroom edit tighten`, `cutroom edit cut --from "The first thing you need is um" --to "the first thing." --dry-run`
  7. `cutroom edit cut --from "The first thing you need is um" --to "the first thing."`, `cutroom edit cut --from "By the way," --to "the keyboard." --dry-run`
  8. `cutroom edit cut --from "By the way," --to "the keyboard."`, `cutroom timeline`, `cutroom timeline --joins`
  9. `cutroom render --preview --dry-run`, `cutroom render --preview`, `cutroom check out/preview.mp4`
  10. `cutroom look --file out/preview.mp4 2`, `cutroom look 0:01 --count 1`
  11. Two image reads of the PNGs (not cutroom)
  12. `cutroom render --dry-run`, `cutroom render`, `cutroom check`
  13. `cutroom transcript --file out/final.mp4`, `cutroom look --file out/final.mp4 1`, one image read

- **Errors hit:** none. No command failed and no `fix:` line was needed.

- **Guessed wrong:** nothing significant. I expected `--from` to need an exact transcript match, and it accepted "um" in "is um" because the transcript had "UM,". The `--near` flag in the tighten retake suggestion was not needed, since the unique `--from` text worked.

- **Confusing output:**
  - `cutroom timeline` after `tighten` prints the retake and aside warnings as long single lines with embedded commands. They are useful but dense; a short line per warning with the command on its own line would be easier to scan.
  - The render-time `picture:` lines ("tone-mapped to SDR bt709 (videotoolbox)") are informative, but nothing says whether the tone map is right beyond the check's self-referential comparison.
  - The `--dry-run` time estimate (~4 s for the preview) was about half the real 8.2 s. Minor.
  - `cutroom look` writes its PNG into a hidden `.cutroom/look/` folder inside the run directory. The printed path is correct, but it is easy to miss.

- **Missing:**
  - A way to judge HDR-to-SDR colour against a reference without relying on cutroom's own tone map. The `tone` check compares the render to the source using the same pipeline, so a consistently wrong tone map (too dark, too flat) would pass. I had nothing independent to compare against except one frame visually.
  - A one-line summary of what `check` did not verify (for example, that the transcript words are right, not just heard in order).

- **Not sure it's right:**
  - Colour: `tone` reports saturation 87 vs 87 and brightness range 217 vs 217 for source vs render. Both sides are cutroom's own measurements, so this does not prove the SDR conversion is correct for a viewer. My one-frame check showed saturated, not washed-out, primaries and upright text ("show", "today"), so I believe it is right, but I did not verify it against a reference.
  - Transcript: the recogniser was `base.en` ("small model; expect misheard words"). I did not run a larger model. The `words` check confirms the render matches the timeline's words, so the content is only as good as that transcript.
  - Orientation: the source frame I looked at at 0:01 is a test pattern with no text, so it does not show the rotation. I judged orientation from the output frames instead, where text reads upright.

- **Best part:** the flow `transcript --suspects` → `edit cut --dry-run` → `edit cut` → `timeline --joins` → `render --preview` → `check` made the whole job straightforward. The suspects list pointed at exactly the false start and the aside, and the dry-runs showed the joins before I committed.

- **One change:** a reference-based colour check for HDR sources, for example a comparison against an independent ffmpeg tone-map or a stated expected saturation range, so an agent can tell a correct SDR conversion from a consistently wrong one.
