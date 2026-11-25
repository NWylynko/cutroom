# Friction report: r8-titles-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s) passes `cutroom check` with 0 failures and 0 warnings. The retake and the cat aside are cut, both overlays are placed and verified, and the audio is at -16.0 LUFS.

- **Commands run:** about 27 cutroom invocations in 10 shell calls, plus one `ls`, and two Read calls on overlay PNGs. In order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `guide edit`, `guide overlays`
  3. `cutroom guide perceive`, `guide render`, `guide verify`, `guide pitfalls`
  4. `cutroom doctor`, `cutroom init talk.mp4`, `cutroom inspect`, `cutroom transcribe`, `cutroom transcript`
  5. `cutroom transcript talk --suspects`, `cutroom transcript talk --words --from 0 --to 30`, `cutroom edit tighten`, `cutroom timeline`
  6. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`, and the same for `--from "By the way" --to "keyboard."`
  7. `cutroom edit cut` (both, applied), `cutroom timeline`, `cutroom timeline --joins`
  8. `cutroom overlay add text ... --layout title --at "Hi," --to "server."`, `cutroom overlay add text ... --layout lower-third --at "Then install" --duration 2.5`, `cutroom overlay list`, `cutroom timeline`
  9. `cutroom render --preview --range 0:00-0:10`, `cutroom look --output 1.90`, `cutroom look --output 7.75`
  10. `cutroom render --dry-run`, `cutroom render`, `cutroom check`

- **Errors hit:** none. Every command succeeded on the first try.

- **Guessed wrong:** nothing material. I expected `cutroom timeline --apply retake-kept@4.09` to be the natural fix. I used a manual `edit cut --near` instead, because the warning's suggested cut would have removed the later take, and the guide says to keep the last one. I did not check which take the `--apply` version would keep, so I did not test that path.

- **Confusing output:**
  - `transcript --words --from 0 --to 30` printed 40 lines and truncated the rest with "24 more lines". I had to guess a narrower window to see the cut points. A "last word" summary would have helped.
  - The `retake-kept` warning gave a ready `edit cut` line that has to be copied exactly, including `--near 6.31`. That worked, but the `--near` value is a source time in seconds, while the transcript prints `0:06.310`. Accepting both formats in `--near` would be friendlier.
  - `overlay add` printed "look at it" and "next: render --preview" hints, which were slightly repetitive. Harmless.

- **Missing:**
  - A way to see the transcript and suspect output on the timeline (output times) without converting source times myself.
  - A `--fit-to` or "span this sentence" option for the title card, so the card's end lines up with the sentence's end. I used `--to "server."`, which ends at output 3.70 s. The sentence ends at about 3.93 s, so the card is about 0.2 s shorter than the sentence.

- **Not sure it's right:**
  - The transcript uses `base.en`, and `check`'s words test uses the same model. So "all 44 expected words heard in order" confirms the cut, not the recogniser. A misheard word would not be caught. I did not run a larger model.
  - I did not `cutroom listen` the two joins at output 3.93 and 6.40 s. The pauses are 0.30 s and the audio stray-sound check passed, but I judged the joins by the transcript only.
  - I checked the overlay frames at 1.90 s and 7.75 s by eye and the picture check passed 15 sampled frames. I did not inspect every frame of the final for stray text.
  - The title card runs output 0.100 to 3.700 s, which covers "Hi, ... server." but ends about 0.2 s before the audio sentence ends.

- **Best part:** the flow of `transcript --suspects`, then `edit cut --dry-run`, then `timeline --joins`, then `overlay add`, then `look`, then `check`. Each step printed the next command. `check` listed every test with its result, which made the pass easy to trust.

- **One change:** let `transcript` and `timeline` show output times for each phrase, or let `overlay add --at` take a source sentence and place the card to match its exact start and end. That would remove most of the time-conversion and guesswork I did.
