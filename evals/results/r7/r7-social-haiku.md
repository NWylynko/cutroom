# Friction report: r7-social-haiku

- **Outcome:** done. `out/final.mp4` (1280x720, 13.767 s, captions burned in) and `out/final.srt` (5 cues) both pass `cutroom check` / `cutroom captions --check`, with the retake and the cat aside removed. The only residual caveat is that the transcript came from `base.en`, so word checks are against the same recogniser (see "Not sure it's right").

- **Commands run:** about 40 `cutroom` invocations across 13 shell calls, in order: `cutroom guide`; `guide workflow`; `guide perceive`; `cutroom doctor`; `guide edit`; `guide captions`; `guide render`; `guide verify`; `guide pitfalls`; `cutroom init talk.mp4`; `cutroom inspect`; `cutroom transcribe`; `cutroom transcript`; `cutroom transcript --suspects`; `transcript --words --from 5.5 --to 13`; `transcript --words --from 17 --to 23`; `cutroom edit tighten`; `cutroom timeline`; `edit cut --from "The first thing you need is UM" --to "the first thing" --dry-run`; `edit cut --from "By the way" --to "keyboard" --dry-run`; `edit cut` (both, applied); `cutroom timeline`; `cutroom timeline --joins`; `captions --help`; `overlay --help`; `captions --output out/final.srt`; `captions --burn`; `overlay list`; `render --preview --dry-run`; `render --preview`; `look --file out/preview.mp4 --output 0:02`; `look --file ... 2.5`; `look --file ... 9.5`; `check out/preview.mp4`; `render`; `check out/final.mp4`; `transcript --file out/final.mp4`; `captions --check out/final.srt`; `look --cuts`.

- **Errors hit:** none. No command failed, and no `fix:` line was needed.

- **Guessed wrong:** nothing failed, but I did not know `captions --output` existed until I read `captions --help`; I had planned to rename the file by hand. I also assumed the burned captions would be a separate step from the SRT, which the output confirmed (`captions --burn` added overlay `o1`).

- **Confusing output:** `tighten` printed "this also removes the sound with no words at source 0:16.970" and the retake warning suggested a `--near 6.31` cut; my own `--from "The first thing you need is UM"` worked, so the suggestion was not needed. `captions --burn` prints groups for "the first 10" with `… 4 more (--all)`, which made it hard to see whether all 14 groups were sane without `--all`. The `look --cuts` output was a PNG only, so the text gave no per-cut verdict.

- **Missing:** a way to choose output names for `render` and captions in one step (`final.mp4` + `final.srt` together). A larger transcription model was suggested by `transcribe` but I did not use one; I had no clear signal of whether `base.en` was good enough for this clip, and `doctor` did not say what the alternatives cost.

- **Not sure it's right:**
  - The retake choice (keep "Then ... The first thing you need is an old laptop", drop the earlier "is, UM, the first thing." take) was judged from the transcript, not by listening to the audio. I did not run `cutroom listen` on the retake or aside joins.
  - The `words` check and `transcript` on the final file use the same `base.en` recogniser that built the timeline, so a misheard word would pass both. The transcript I read back matches what I expect the speaker said, but that is not independent evidence.
  - I looked at one caption frame (0:09.5) on the preview; I did not visually inspect the final render's captions at other times, or the `look --cuts` image.
  - Inter-sentence pauses after tighten are 0.30 to 0.33 s, which I left as the tool's default; I did not judge whether that is too slow for social pacing.

- **Best part:** the chained flow `transcript --suspects` -> `edit cut --dry-run` -> `timeline --joins` -> `captions --burn` -> `render --preview` -> `check`. The dry-run join preview and the `check` lines (captions, words, sync, loudness) gave clear, verifiable pass/fail signal per requirement.

- **One change:** let `cutroom render` (or a `deliver`-style command) take `--srt <path>` and `--burn` and write the video plus the sidecar subtitle file in one call, so the final deliverables come out with the names I asked for and no separate captions step needs to be remembered.
