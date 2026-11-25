# Friction report: 11-social-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s, 1280x720, burned-in captions) and `out/final.srt` (5 cues) both passed `cutroom check` with 0 failures and 0 warnings, and the captions check passed. Caveat: the only transcription model installed is base.en, so the words are verified against the same model that made the transcript.

- **Commands run:** about 39 `cutroom` invocations over 15 shell calls, plus `ls`, `mkdir -p out`, `cat out/final.srt`. In order:
  `cutroom guide`; `guide workflow`; `guide edit`; `guide verify`; `guide captions`; `guide overlays`; `guide render`; `guide perceive`; `guide pitfalls`; `doctor`; `init talk.mp4`; `inspect`; `transcribe`; `transcript`; `transcript --suspects`; `transcript --retakes`; `transcript --words --range 0-31`; `edit tighten`; `edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`; `edit cut --from "By the way" --to "keyboard." --dry-run`; `edit cut` (same two, applied); `timeline`; `timeline --joins`; `transcript --output`; `captions --help`; `captions --burn --output out/final.srt --dry-run`; `captions --burn --output out/final.srt`; `render --preview --dry-run`; `render --preview`; `check out/preview.mp4`; `look --overlays`; `look --output 0:03-0:06`; `render`; `check`; `captions --check out/final.srt`; `transcribe --help`; `look --output 0:08-0:13.7 --count 8`; `listen --output 4-6`; `status`.

- **Errors hit:** none. No command failed or exited non-zero.

- **Guessed wrong:**
  - `transcribe` printed "`cutroom doctor` shows how to get a larger one." But `doctor` only printed four `ok` lines (ffmpeg, ffprobe, whisper-cli, whisper-model base.en). It did not say how to get a larger model, so the hint was a dead end.
  - I expected `captions --burn` to write the SRT to `out/captions.srt` and needed `--output out/final.srt` to name it. The help text does say `--output` names the file, so this was my oversight, not a gap.

- **Confusing output:**
  - Burned captions are grouped in a different way from the SRT cues. The burn grouping shows "Hi, today I'm going" / "to show you"; the SRT has "Hi, today I'm going to show you" / "how to set up a home server." Both are correct, but nothing says the two files differ on purpose until you read the guide.
  - `timeline` warnings suggest `--near 6.31` on a source time, and a `retake-kept` warning quotes a cut whose range (6.200–9.433) is wider than the words it names. Dry-running it first helped.
  - `cutroom transcript --words --range 0-31` printed 64 lines. Useful, but long; I would have liked a `--range` that prints only the words in a span.

- **Missing:**
  - A larger whisper model, or a way to know which models are available and how to fetch one. The `transcribe` note says to use a bigger model, and I could not find one. `cutroom transcribe --model` only accepts a file path, and only base.en exists locally.
  - A way to confirm the speech is transcribed correctly. `check` re-transcribes with base.en, so `words: all 44 expected words heard` only confirms agreement with base.en, not with the speaker. The check's own note says this.

- **Not sure it's right:**
  - Word accuracy. The transcript is base.en output (64 words for 30 s of speech, with the note "expect misheard words"). The retake choice (keep "The first thing you need is an old laptop", cut "…is, UM, the first thing.") rests on the repeat suspect plus the word timings, and I did not hear the audio myself. `listen` only re-transcribed with base.en.
  - Caption wording. Spot-checked frames in `look` show the burned groups match the SRT text, but I could not confirm every word matches what was said.
  - The stray sound at source 16.97 (0.33 s, peak -22 dB) was removed by `tighten`, which said so. Not verified by ear.
  - Output stays 1280x720 landscape. The request did not specify a format, so I did not make a vertical version.

- **Best part:** `timeline --joins` plus `check`. The join report shows each cut with the words either side and its pause, and `check` then verifies frames, sync, words, stray sound, loudness and the SRT against the timeline. That made "is it correct" answerable with specific numbers.

- **One change:** have `doctor` list every whisper model it can find and tell you how to get a larger one, since the `transcribe` note points there and the doctor output does not deliver it. Failing that, an explicit `cutroom transcribe --model` example that shows the expected model path would have saved a dead end.
