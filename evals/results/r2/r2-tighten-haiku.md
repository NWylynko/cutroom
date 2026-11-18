# Friction report (fill this in as the last thing you do)
Be blunt and specific. The tool's authors will read this to make the tool better for agents like you. Quote exact commands and output where it helps.
- **Outcome:** done — `out/final.mp4` (13.767 s, 5 segments) rendered and `cutroom check out/final.mp4` passes every line; the retake and the cat aside are gone, all other sentences kept. Not checked: the words in the final file were not re-transcribed separately, only checked by `check`'s word test.
- **Commands run:** 8 shell calls, 27 `cutroom` invocations, plus 2 image reads. In order:
  1. `cutroom guide`
  2. `cutroom guide workflow && guide edit && guide perceive && guide render && guide verify && guide pitfalls`
  3. `cutroom doctor; cutroom init talk.mp4; cutroom inspect; cutroom transcribe`
  4. `cutroom transcript talk; transcript talk --suspects; transcript talk --words --from 0 --to 31`
  5. `cutroom edit tighten talk`; `cutroom edit cut --from "The first thing you need is, UM" --to "first thing." --dry-run`; `cutroom edit cut --from "By the way" --to "keyboard." --dry-run`
  6. `cutroom edit cut` (both, for real); `cutroom timeline`; `cutroom timeline --joins`
  7. `cutroom render --preview && cutroom check out/preview.mp4`
  8. `cutroom render && cutroom check out/final.mp4; cutroom look --cuts; cutroom listen --output 3-5`
  Plus Read on `.cutroom/look/cuts_p1.png` and the listen PNG.
- **Errors hit:** none. No `error[...]` lines and no non-zero exits. The only `WARN` was `WARN[retake-kept] at 0:04.093: "the first thing you need is" is still said 2× in the timeline`, which came from the tighten step before I cut the retake. It went away once the cut was applied (`timeline: ... no problems found`). The `fix:` pointer was not needed.
- **Guessed wrong:** nothing material. `--from`/`--to` with quoted text worked on the first try, and I quoted "first thing." with its period to make `--to` unique. I did not need `--occurrence` or `--near`. I did not try `transcript --search`, because `--suspects` already showed the retake.
- **Confusing output:**
  - The dry-run of the cat-aside cut also printed the `WARN[retake-kept]` line from the earlier state. It looks like a problem with the cut, when it is only the pre-retake state. A dry-run should either say that it is showing warnings from the current timeline, or not repeat them.
  - `transcript` says `0:06.310–0:09.220 The first thing you need is, UM, the first thing. [repeat]`, but the cut that actually ran was `0:06.200–0:09.433`. The padding is only visible in the cut's own output. That was fine, but I had to work it out.
  - `timeline` shows output times on the left and source ranges on the right. That is documented, but easy to mix up when I was reading the join list next to the transcript.
  - Times in `edit`/`transcript` are source times, and `look --output`/`listen --output` take output times. I used `listen --output 3-5` and it reported `cuts in range: 0:03.933`, which is correct, but I had to check the unit.
- **Missing:**
  - No way to re-transcribe the final render directly from the command line. `transcript` seems to work only on sources. `check` compares the render's words with the timeline's words, but I could not print the final's transcript to read it myself. I expected something like `cutroom transcript --file out/final.mp4`, and `look`/`listen` support `--file`, but I did not try it on transcript.
  - No way to see the whole timeline's words as continuous prose in one place. `timeline` lists words per segment, which was enough for 5 segments but would be tedious for a long video.
  - Nothing said which repeat to keep. The guide says "the last take is usually the keeper" and `--suspects` shows both, but I had to read both takes' text and timing myself. The first take had "UM" (a false start), and the second was clean, so the rule held here.
- **Not sure it's right:**
  - Whisper ran on `base.en` (the default in `doctor`), and the guide itself warns about misheard words. The word test passed against the timeline's words, but both sides come from the same recogniser, so a shared mishearing would pass. The intro line "Hi, today I'm going to show you how to set up a home server." was not isolated and listened to.
  - I judged the retake by reading two takes at 6.31 and 10.25. I did not listen to either take in isolation (`cutroom listen 6.2-9.5` and `10.1-12.6`). The spectrogram I looked at was only at the cut 3–5 s in output.
  - The video is a colour test pattern with a counter, so the picture check tells me the frames match the source, but I cannot judge anything about the speaker from the image.
  - The cut at 0:18.5–0:21.9 removed a sound at 0:16.97 that `tighten`/`transcript` flagged as a suspect. It is inside the removed aside, so `stray-sound` passing is consistent, but I did not listen to it.
- **Best part:** `cutroom transcript --suspects` pointed straight at the retake and the aside, and `cutroom edit cut --from ... --to ... --dry-run` showed the exact range and the resulting join before I committed to it. The `timeline --joins` view and `look --cuts` image made the result easy to check.
- **One change:** let `cutroom transcript` (and `cutroom transcribe`) take `--file out/final.mp4`, so I can read what the render actually says instead of trusting the timeline's words. That is the one place I wanted to check the output and could not.
