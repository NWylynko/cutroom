# Friction report (fill this in as the last thing you do)

- **Outcome:** partly done. `out/final.mp4` has the recorder's audio, in sync to about 10–30 ms across the cuts, with pauses tightened and the first retake removed. But cutroom has no way to use a separate audio source, so I built that audio with raw ffmpeg and remuxed it onto cutroom's video. It is not cutroom's voice chain, and I could not run `cutroom check`.

- **Commands run:** 19 shell commands, plus one rejected `update_plan` call (not a cutroom call).
  1. `cutroom guide`
  2. `cutroom guide workflow|edit|render|pitfalls`, `ls -la`
  3. `cutroom doctor`, `cutroom init talk.mp4 --source mic=recorder.wav`, `cutroom inspect`
  4. `cutroom --help`
  5. `cutroom init|edit|render|inspect --help`
  6. `cutroom guide --all | grep ...`, `cutroom timeline --help`, `cutroom sync --help`, `cutroom audio --help`
  7. `cutroom transcribe`, `cutroom transcript`
  8. `cutroom transcript mic`, `transcript talk --suspects`, `transcript mic --suspects`
  9. `cutroom edit tighten`, `cutroom timeline`, `cutroom render --dry-run`
  10. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`, `cutroom transcript --search "the first thing you need is"`
  11. `cutroom edit cut ... --near 6.3`, `cutroom timeline --joins`
  12. `cutroom inspect mic --json`, `cat cutroom.json`, `ls`
  13. `cutroom render --preview`, `cutroom check`
  14. `python3 -c "import numpy"`, `ffmpeg` extract to 8 kHz WAV in /tmp
  15. `cutroom transcript talk|mic --words ...` (four calls for anchor words)
  16. `cutroom render`, `cat timeline.json`, `ffprobe out/final.mp4`
  17. `ffmpeg ... -filter_complex_script ...` (failed), then `ffmpeg ... -/filter_complex ...` remux and frame extract
  18. `cutroom init --source out=out/final.mp4`, `cutroom transcribe out --passes 1`, `cutroom transcript out --words ...`

- **Errors hit:**
  - `error[unknown-command]: unknown command "check"` with `fix: run cutroom to list commands`. This did not get me unstuck. The fix line points to a list that does not contain `check`. The render output and `guide workflow` both mention `cutroom check` as something to run, so an agent is told to run a command that does not exist.
  - `error[unknown-command]: unknown command "sync"` and `"audio"` (my guesses, see below). The fix line was the same generic one, which helped only by pointing to `cutroom` with no args.
  - `Unrecognized option 'filter_complex_script'` from ffmpeg. This was my fault for using the older flag. It is not a cutroom error, but cutroom's guide doesn't say anything about raw ffmpeg.
  - `ls -la out` in a command where `out/` did not exist printed to stderr and made the call report `Exit code 1`, even though the cutroom `inspect --json` part succeeded. Not a cutroom bug, but it looked like one.

- **Guessed wrong:**
  - I assumed `init --source mic=recorder.wav` would make the render use the recorder's audio. It does not. `timeline.json` has `"audio": "talk"`, and nothing in `edit`, `render` or `timeline` lets you change it.
  - I assumed there was a sync or offset command. The guide mentions only `init --source` for a separate recorder, with no sync step.
  - `inspect mic` reports `av_offset: 0` and says nothing about the 1.70 s start offset relative to the camera. I had to work it out myself from matching words.
  - The guide's "Separate audio recorder? Add it as a source too" implies the recorder gets used. In practice it is only transcribed.

- **Confusing output:**
  - `render --preview` ends with `next: check the render before calling it done (look at frames; cutroom check is coming)`. This is a pointer to a command that doesn't exist yet, so it reads like a missing piece rather than a plan.
  - `timeline` after `tighten` printed `WARN[retake-kept]` for the first retake, and I could only confirm it went away by re-running `timeline`. A "resolved" line would help.
  - Timeline output times and source times are in different columns, and `edit` takes source times. Easy to mix up, though the guide says so.
  - `transcript --words` prints `talk — 64 words, 0:30.261 (base.en, 2 pass(es))` before every word table, which is noise when you call it repeatedly.

- **Missing:**
  - A way to use a separate audio source as the timeline's audio, with a measured offset and drift, e.g. `cutroom sync talk mic` that prints `offset 1.700 s, drift X ppm`, and something like `edit audio mic` or `render --audio mic` that applies it.
  - An audio cross-correlation or offset measurement. I had no numpy for it, so I used word-timestamp anchors (`Hi,` 0.58 vs 2.28, `server.` 4.18 vs 5.88, `ready.` 29.29 vs 30.99, cat sound 16.97 vs 18.67). All give 1.70 s, but that is by eye.
  - `cutroom check`, which the guide says is the required verification step.
  - Removing a source (`cutroom init` added `out` for my verification and I could not take it back out).
  - A way to render with a non-default audio chain applied to another source.

- **Not sure it's right:**
  - The 1.70 s offset is from four word anchors, each good to about ±10 ms, and it looked constant over 30 s. I did not measure drift precisely.
  - The final audio is my own ffmpeg chain (48 kHz resample, 20 ms fades at cuts, `highpass=f=80`, `loudnorm` to -16 LUFS / -1.5 dBTP). It is NOT cutroom's speech chain (rumble cut, compression, shelf), so loudness and tone differ from what `cutroom render` would give. I verified sync with a whisper pass on the output: `ready.` at 16.76–16.99 s, which matches the expected 16.99 s.
  - I did not check any frame visually. The video is cutroom's own render, so I trust it, but `cutroom check` did not exist to confirm it.
  - Left in the project: `out/preview.mp4`, and the `out` source that I added to `cutroom.json` for verification. The mic source and the "By the way, my cat" aside are both still in the cut (the aside was not part of the brief).

- **Best part:** The transcribe → transcript → `--suspects` → `edit cut --from/--to --near` → `timeline --joins` loop was quick and easy to follow. `--dry-run` on the cut showed the exact join and removed duration before I committed. `render` reported the duration matching the timeline, and `--near 6.3` resolved the repeated phrase in one go.

- **One change:** Add a real way to use a separate recorder as the audio source. For example, `cutroom sync talk mic` that measures the offset and drift and writes them to `cutroom.json`, plus `render --audio mic` (or an `edit` option) that applies that offset and the normal voice chain. That would have made the whole task one command set instead of a hand-built ffmpeg remux that skips the chain.
