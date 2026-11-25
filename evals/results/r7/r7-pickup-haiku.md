# Friction report: r7-pickup-haiku

- **Outcome:** done. `out/final.mp4` (13.800 s) passes `cutroom check` with 0 failures and 0 warnings: the retake, the cat aside and the 3 s pauses are gone, and "Then install Linux on it." comes from pickup.mp4. The only open item is the audio-mixed warning, which I could not settle (see below).

- **Commands run:** about 42 cutroom invocations in 12 shell calls, in this order:
  `guide`, `guide workflow`, `guide edit`, `guide render`, `guide verify`, `guide perceive`, `guide pitfalls`, `doctor`,
  `init talk.mp4 --source pickup=pickup.mp4`, `inspect`, `transcribe`, `transcript`, `transcript --suspects`,
  `transcript pickup`, `transcript pickup --suspects`, `transcript talk --words --from 0:00 --to 0:31`,
  `edit replace --help`, `edit --help`, `edit tighten --dry-run`, `edit tighten`,
  `edit cut --from "The first thing you need" --near 6.31 --to "first thing." --source talk --dry-run` (then without --dry-run),
  `edit cut --from "By the way" --to "keyboard." --dry-run` (then without),
  `edit replace --text "Then install Linux on it." --with pickup --dry-run` (then without),
  `timeline`, `timeline --joins`, `render --preview --dry-run`, `render --preview`, `check out/preview.mp4`,
  `render`, `check out/final.mp4`, `listen --output 5.6-7.2`, `transcript --file out/final.mp4`, `status`,
  `timeline --help`, `render --help`.

- **Errors hit:** none. No command failed.

- **Guessed wrong:**
  - I used `init talk.mp4 --source pickup=pickup.mp4`, copying the separate-recorder form. It worked, but cutroom made pickup a second video take. The "Several takes" guide section uses plain `init take1.mov take2.mov`, and I only noticed the difference after the fact. I should have used the plain form.
  - I expected `--from/--to` with `--near` to need the source-time form `--near 6.31`. The warning's suggested command gave exactly that, so this was fine.

- **Confusing output:**
  - `WARN[audio-mixed]` is a long single line. It says "fine as it is" if the pickup came from the same microphone, but it gives no way to check that. The warning stays open after the replace, so `status` reports "0 error(s), 1 warning(s)" on a result that is otherwise clean.
  - `transcript talk --words` printed about 90 lines and was cut off at about 40 with "… 24 more lines", so I had to re-run it with narrower bounds.
  - `listen --output 5.6-7.2` says "words planned here: old laptop. Then install" but heard only "laptop, then install". The window ended before "Linux on it", so the check was not useful for that join. A window that follows the segment boundaries would help.

- **Missing:**
  - A way to compare the replaced segment's room tone or level with its neighbours. The audio-mixed warning asks the user, not the tool, to judge it.
  - Some way to confirm that `--with pickup` used the re-recording and not the original line. I only confirmed it through `timeline` (`s8 pickup 0:00.367–0:02.133`) and the transcript text, which are both base.en output.

- **Not sure it's right:**
  - The audio-mixed join at 0:06.400 (pickup's sound next to talk's sound) is unverified. I did not listen to it properly. The loudness of both is normalized to -16 LUFS, but the tone and mic match is untested.
  - The pickup's own pauses at its start and end (`0:00.000–0:00.367` and after `0:02.133`) were trimmed by the replace's pads. I did not verify that the pickup's first word is not clipped beyond what `check` measures. `check` reported "all 44 expected words heard in order", but that check uses base.en transcription of the render, so it cannot catch a clipped sound.
  - Tighten also removed the 0.32 s sound at 0:16.970 that suspects flagged. I read that as a cough or breath in a pause and accepted it without listening.

- **Best part:** `edit replace --text "…" --with pickup --dry-run` showed the exact source ranges and the audio warning before anything was saved. Then `timeline --joins` plus `check` gave a clear pass/fail picture with no guessing.

- **One change:** have `edit replace --with <take>` (and the audio-mixed warning) print the replaced segment's room-tone and level difference against its neighbours, so an agent can judge the mic match with a number instead of a warning it can't act on.
