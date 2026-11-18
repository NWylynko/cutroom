# Friction report: r3-tighten-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s, from 30.261 s) passes `cutroom check` with 0 failures and 0 warnings. I did not listen to the cut joins myself, so the audio at the joins rests on the check and the cut frames.

- **Commands run:** 27 shell calls (all cutroom except `ls`/`test`), plus 2 image reads.
  1. `cutroom guide`
  2. `cutroom guide workflow`, `perceive`, `edit`
  3. `cutroom guide render`, `verify`, `pitfalls`
  4. `cutroom doctor`; `cutroom init talk.mp4`; `cutroom inspect`
  5. `cutroom transcribe`; `cutroom transcript`; `cutroom transcript --suspects`
  6. `cutroom transcript talk --words --from 6 --to 12`; `cutroom listen 16.5-17.8 --json`
  7. `cutroom edit tighten`; `cutroom edit cut ... --dry-run` (x2)
  8. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing."`; `cutroom edit cut --from "By the way, my cat" --to "across the keyboard."`; `cutroom timeline`; `cutroom timeline --joins`
  9. `cutroom render --preview --range 0:00-0:14 --dry-run`; `cutroom render --preview`; `cutroom check out/preview.mp4`
  10. `cutroom look --cuts --file out/preview.mp4`; `cutroom render`; `cutroom check`

- **Errors hit:** none. No command failed.

- **Guessed wrong:** `--range` on `render --preview --dry-run` still reported `1 chunk(s) ... 13.767 s of video`, the whole timeline, not the 0:00-0:14 range I asked for. Either `--range` is ignored under `--dry-run`, or I misread it. The guide only says `--range` "renders only the chunks covering that output range". I did not work out which. I also guessed that `--to` matches the first occurrence after `--from`. The dry-run confirmed it, so I used `--dry-run` before each cut.

- **Confusing output:** The `timeline` output after the aside cut shows `s4 talk 0:13.933–0:15.667` followed directly by `s6 talk 0:22.833`. The cut was reported as `0:18.500–0:21.933`, and nothing says what happened to the source stretch 15.667–18.5 or the sound at 16.97 that `cutroom transcript --suspects` had flagged. The check's `stray-sound` passing is what told me the sound is gone. I would have liked the removed stretch and its reason listed in the cut output. The `listen --json` output also prints `"heard_isolated": ""` with no hint of whether empty means "silence" or "recogniser failed".

- **Missing:** a way to explain a cut's effect on neighbouring pauses. Tighten and cut both change pauses, and `timeline` only shows the resulting segments, not the removed spans. A `--json` option on `timeline` and `transcript --suspects` would have helped, but I did not check whether these exist.

- **Not sure it's right:**
  - The transcript came from `base.en`, which the tool itself warns about. The check's `words` pass matches the render's re-transcription against the timeline's words, so it cannot catch a word the recogniser misheard in both passes.
  - I kept the second take of "The first thing you need is ..." (10.25 s) and dropped the first (6.31 s, with "UM"). I judged that from the transcript and did not listen to either take.
  - The 0.32 s burst at 16.97 s is broadband and has no words, so it is probably a bump or click. I didn't confirm that. It is gone from the output, so I cut it by removing the stretch around it.
  - I checked the cut frames only on the preview render, which has identical cuts. I did not look at the final file's frames directly.

- **Best part:** `cutroom transcript --suspects` surfaced the retake, the stray sound and the repeat in one listing. `--dry-run` on each cut let me check the exact span before changing the timeline. `cutroom check` gave a per-check PASS/fail list that made the final verification fast.

- **One change:** have `cutroom edit cut` (and `edit tighten`) print what was removed, as source ranges with their words or sound, plus what happened to the pauses on either side of each cut. That would have replaced the guessing and the extra `timeline` and `listen` calls.
