# Friction report: r4-recorder-haiku

- **Outcome:** done. `out/final.mp4` (1280x720, 20.433 s) was rendered from the recorder audio, synced to the camera at a 1.700 s offset with -42.7 ppm drift, with the long pauses tightened, and `cutroom check out/final.mp4` passed with 0 failures and 0 warnings. Not verified by ear.

- **Commands run:** 24 cutroom invocations, plus shell `ls`, `export PATH`, and `cd`.
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom guide edit`
  4. `cutroom guide perceive`
  5. `cutroom guide render`
  6. `cutroom guide verify`
  7. `cutroom guide pitfalls`
  8. `cutroom doctor`
  9. `cutroom init talk.mp4 --source mic=recorder.wav`
  10. `cutroom inspect`
  11. `cutroom sync mic`
  12. `cutroom edit audio mic`
  13. `cutroom transcribe`
  14. `cutroom transcript`
  15. `cutroom transcript --suspects`
  16. `cutroom edit tighten`
  17. `cutroom timeline`
  18. `cutroom timeline --joins`
  19. `cutroom render --preview --dry-run`
  20. `cutroom render --preview`
  21. `cutroom check out/preview.mp4`
  22. `cutroom render`
  23. `cutroom check out/final.mp4`
  24. `cutroom check`

- **Errors hit:** none. No command failed, so no `fix:` line was needed.

- **Guessed wrong:** nothing significant. The "Separate audio recorder" section of `cutroom guide workflow` matched the real sequence exactly (`init --source`, `sync`, `edit audio`). I did not need the `--json` flag.

- **Confusing output:**
  - `cutroom check` sync line reads awkwardly: "the render's sound vs mic (aligned to talk by sync)'s sound at the source moment…". It is hard to tell what was compared. It would help to say "recorder audio, aligned by sync" plainly.
  - `cutroom sync mic` reports "mapping: mic time = +1.700 s + talk time × (1 -42.7 ppm)", which is correct but dense. The sign convention took a re-read.
  - `cutroom edit tighten` lists "pauses longer than pads 0.22+0.10 s + --min-cut 0.4 s" in its summary. This is useful, but the reason it chose 0.32 s is buried in the text.
  - The `retake-kept` warning after tighten prints a ready-made `edit cut` command. That is helpful, but it sits in the same output as the success line, so it is easy to read as a failure. I did not run it because the task only asked for pauses to be removed.
  - `cutroom transcribe` printed 64 words for talk and 63 for mic in the same output. It was not obvious which transcript the timeline uses. The guide says to use the camera source's transcript, and I did.

- **Missing:**
  - A way to tell the tool that only pauses should be tightened, with no retake or aside handling. Each `WARN` then asks for a decision, and I had to judge each one on its own.
  - An independent audio-sync check. `check`'s sync measures the render against the aligned mic audio, which tests that the offset was applied. It does not test that the 1.700 s offset itself is right. That offset comes from `sync`, so a sync error would pass `check`.
  - A listen or look on the final output's join points. I used `timeline --joins` only; I did not run `cutroom listen` on any join.

- **Not sure it's right:**
  - The sync offset (1.700 s, -42.7 ppm) was measured by cutroom. I did not verify it against the audio independently. `check` shows 0 ms median and -6 ms worst error, but that compares the render with the same alignment, so it cannot catch a wrong offset.
  - The loudness gain was +9.9 dB because the recorder is quieter (-16.8 LUFS). The render landed at -16.0 LUFS and peak -4.4 dBTP, which is within the settings. I did not judge the sound quality.
  - Retakes and the cat aside were left in on purpose (the task did not ask for them removed). The final video still has the duplicated "The first thing you need is" opening at about 0:03.9–0:09.6 in output time.
  - The 0.32 s pauses are kept as set by tighten. I did not check whether this reads as natural speech.

- **Best part:** the guide's "Separate audio recorder" recipe. Three commands (`init --source`, `sync`, `edit audio`) took the recorder from raw file to aligned timeline, and `check` covered the rest in one pass.

- **One change:** have `cutroom check` report the sync offset and drift it applied, and state that it verified the render against that offset, not against the camera's own sound. That would make clear what `sync` and `check` each prove.
