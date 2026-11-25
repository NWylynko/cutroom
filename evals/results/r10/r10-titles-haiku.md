# Friction report: r10-titles-haiku

- **Outcome:** done. out/final.mp4 (13.767 s) passes `cutroom check` with 0 failures and 0 warnings. I did not watch or listen to it myself; the verification rests on check output and two overlay frames I looked at.

- **Commands run:** about 40 cutroom invocations, plus 2 image reads of the look PNGs. In order: `cutroom guide`; `guide workflow`, `guide perceive`, `guide edit`, `guide overlays`, `guide render`, `guide verify`, `guide pitfalls`; `doctor`; `init talk.mp4`; `inspect`; `transcribe`; `transcript`; `transcript --suspects`; `transcript --words --from 5.5 --to 13`; `edit tighten`; `timeline`; `edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`; same with `--from "By the way," --to "the keyboard."`; the two real `edit cut` runs; `timeline`; `timeline --joins`; `transcript --help`; `overlay add text` x2; `overlay list`; `timeline`; `overlay set o2 --duration 2.5`; `overlay list`; `timeline`; `look --output 1.9`; `look --output 7.2`; `render --preview --dry-run`; `render --preview`; `check out/preview.mp4`; `render`; `check out/final.mp4`; `transcript --file out/final.mp4`.

- **Errors hit:** none. No command failed. The only WARNs were `retake-kept` and `aside-kept` after tighten (both resolved by my cuts) and `overlay-short` on o2 (resolved by lengthening it).

- **Guessed wrong:** nothing failed, but I guessed `--layout title` and `--to "home server."` on `overlay add` and both worked first time. I had assumed `--near` takes a source time; the guide says so, and the `retake-kept` suggestion used it correctly.

- **Confusing output:**
  - `timeline` prints output times in its segment list, but the `retake-kept` and `aside-kept` WARNs print the suggested fix in source times (`--near 6.31`, `source 0:06.200–0:09.433`). Two time bases in one line made me re-read it.
  - The `overlay-short` WARN says "needs ~2.1 s" but the card at 2.1 s is the minimum, not a comfortable duration. I chose 2.5 s. Fine, but the hint reads as if 2.1 is the target.
  - `tighten` reports "this also removes the sound with no words at source 0:16.970". Useful, but it was the only hint that the cough/sound was gone, and it is easy to miss.

- **Missing:**
  - A way to dry-run `overlay add` or see the overlay's relation to the surrounding joins before adding it.
  - A single "check the whole plan" view that shows timeline problems and overlay problems together. I had to run `timeline` and `overlay list` separately to be sure both were clean.
  - No speaker-framing check on a full frame of the real picture: `look` on the title and lower-third frames showed the test-pattern background, so I could not judge framing of the speaker with the overlays on.

- **Not sure it's right:**
  - The transcript is from base.en, the small model. `check` words passed, but it re-transcribed with the same model, so it confirms the cut rather than the words. The final transcript reads correctly to me; I did not verify each word against audio beyond that.
  - Title card timing: "Hi, today" to "home server." is 0.10–3.70 s, i.e. over the intro sentence. Lower-third: 6.50–9.00 s, over "then install Linux on it." Both verified by check, not by watching.
  - I did not inspect the audio edit near the removed aside (joins are 0.30–0.33 s pauses). `timeline --joins` read naturally, and check's stray-sound check passed, but I did not `listen` to them.

- **Best part:** The workflow: `transcript --suspects` surfaced the retake, the sound, and the aside, and `edit cut --dry-run` showed each join before I applied it. `overlay add` anchored to quoted words meant I never had to time the cards by hand.

- **One change:** Print every WARN's suggested command with the same time base the rest of that output uses (output time, or a note that says `--near` is a source time), and give `overlay-short` a recommended duration with the same readable-time margin the card needs. This would have removed the one re-read I had to do.
