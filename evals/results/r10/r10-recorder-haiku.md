# Friction report: r10-recorder-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s) uses the recorder's audio, `cutroom check` passed with 0 warnings (sync median -0 ms, worst -6 ms, limit 45), and the two long pauses plus the retake and the cat aside were removed. I did not listen to the joins myself; I checked them with `look --cuts` (frames) and `check` (words, sync, stray sound).

- **Commands run:** 24 cutroom commands and 1 image read (the cut-frames PNG), in this order:
  1. `cutroom guide`
  2. `cutroom guide workflow`
  3. `cutroom guide edit`
  4. `cutroom guide perceive`
  5. `cutroom doctor`
  6. `cutroom init talk.mp4 --source mic=recorder.wav`
  7. `cutroom sync mic`
  8. `cutroom inspect`
  9. `cutroom edit audio mic`
  10. `cutroom transcribe`
  11. `cutroom transcript`
  12. `cutroom transcript --suspects`
  13. `cutroom edit tighten`
  14. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`
  15. `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing."`
  16. `cutroom edit cut --from "By the way" --to "keyboard" --dry-run`
  17. `cutroom edit cut --from "By the way" --to "keyboard"`
  18. `cutroom timeline`
  19. `cutroom timeline --joins`
  20. `cutroom render --preview --range 0:00-0:06`
  21. `cutroom render`
  22. `cutroom check out/final.mp4`
  23. `cutroom transcript --file out/final.mp4`
  24. `cutroom look --cuts`

- **Errors hit:** None. No command failed, and no `fix:` line was needed.

- **Guessed wrong:** Nothing major. I guessed `init --source mic=recorder.wav` would work from the workflow guide, and it did. I expected the `--to "the first thing."` match to land inside the first take, and the dry-run confirmed it.

- **Confusing output:**
  - `cutroom edit audio mic` printed `timeline: 1 segment, 30.233 s (was 30.233 s)`, but the talk video is 30.261 s. The 28 ms gap is never explained, and "was" and "now" were identical, so the line told me nothing.
  - `cutroom transcribe` transcribed the `mic` source too (`mic  64 words ... 4.2 s`). That took time, and I did not need it for the camera-timed edit. The output is also long for two sources when I only need the camera one.
  - `cutroom transcript --file out/final.mp4` splits the first phrase as "Hi!" then "Today I'm going..." with a 0.18 s pause. The cut is not at fault, since the source segment is continuous; base.en just splits it. A reader could take it as a cut.

- **Missing:**
  - A way to see the recorder's coverage against the video. `sync` reports the offset and drift, but not whether the recorder starts or ends before or after the video. Here it did not matter, but I had to check it by arithmetic.
  - An explanation of why the timeline length (30.233 s) differs from the source length (30.261 s) after `edit audio`.

- **Not sure it's right:**
  - I did not listen to any join (`cutroom listen`) or compare the recorder against the camera's own sound by ear. The check's sync and stray-sound checks passed, but a subjective check of the cut joins is still open.
  - The `check` word test uses base.en, the same model that made the transcript, so it confirms the cut, not whether the recogniser heard the words correctly. The guide says so.
  - The tighten pause changes (2.13→0.32 s, etc.) and the removed 0.32 s sound at 16.97 s look right on paper, but I did not hear them.

- **Best part:** The "Separate audio recorder" section of `cutroom guide workflow` gave me the three-step recipe (`init --source`, `sync`, `edit audio`). Each step's `next:` line then pointed to the next one, so I never had to guess. `cutroom check` reporting sync per segment with a limit and a median and worst value made the result easy to trust.

- **One change:** `cutroom transcribe` should transcribe only the camera source by default, or take a flag for the others. Transcribing the recorder doubled the wait and added output I did not need, since the camera's own speech is what the edit is timed against.
