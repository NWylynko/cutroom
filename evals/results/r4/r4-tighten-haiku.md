# Friction report: r4-tighten-haiku

- **Outcome:** done. The final render passes `cutroom check` with 0 failures and 0 warnings, and the timeline keeps the 44 expected words (64 minus the 10-word false start and the 10-word cat aside).

- **Commands run:** 32 cutroom invocations, in order: `guide`, `guide workflow`, `guide perceive`, `guide edit`, `guide render`, `guide verify`, `guide pitfalls`, `doctor`, `init talk.mp4`, `inspect`, `transcribe`, `transcript talk`, `transcript talk --suspects`, `transcript talk --words --from 0 --to 31`, `transcript talk --words --from 18 --to 31`, `listen 16.5-17.8`, `edit tighten`, `edit cut --from "The first thing you need is, UM" --to "first thing." --dry-run`, `edit cut` (same, applied), `edit cut --from "By the way" --to "keyboard." --dry-run`, `edit cut` (same, applied), `timeline`, `timeline --joins`, `render --preview --dry-run`, `render --preview`, `check`, `render`, `check out/final.mp4`, `look --cuts`, `transcript out/final.mp4` (failed), `listen --file out/final.mp4 3.5-4.6`, `listen --file out/final.mp4 6.0-6.8`.

- **Errors hit:**
  - `error[unknown-source]: no source "out/final.mp4"` / `fix: add it first: cutroom init out/final.mp4`. I did not follow the fix. Running `init` on the output would add it as a source, so I used `listen --file` instead. I wanted to re-read the final transcript and could not do it without the `init` workaround.

- **Guessed wrong:**
  - I assumed `cutroom transcript` would accept any video file, as the guide says "any time printed by cutroom can be passed back". It only accepts sources, and the error is the only signal.
  - `listen --file` on the final: the isolated-clip "heard" text did not match the expected words ("The first thing you do is to make sure you're not ready to do it." for a span that should say "server. The first"). The guide does warn that isolated re-transcription is unreliable, but it was jarring at verify time.

- **Confusing output:**
  - The `retake-kept` WARN after `tighten` suggested `--from "the first thing you need is" --near 6.31 --to "first thing."`. I used `--from "The first thing you need is, UM"` instead. It worked, and the dry-run confirmed the join, but I had to decide which phrasing was safe.
  - `timeline` prints output times and source ranges in one column layout. It is readable, but I had to keep two time bases straight.
  - `transcript` on the source lists `[repeat]` on the opening take even though the earlier take had "UM" in it. The suspect list was clear enough.

- **Missing:**
  - A way to transcribe the rendered file directly (for example `transcript out/final.mp4` or `transcript --file`), so the final can be re-read as text without `init`ing it as a source.
  - A summary line in `check` that counts expected words against the total. I had to count 64 - 20 = 44 myself to confirm the number.

- **Not sure it's right:**
  - Verification relies on the base.en model. Source transcript words such as "Hi, today I'm going to show you how to set up a home server" were read by base.en, which may mishear. The final `check` words test is a self-consistency check, not an independent read of the render.
  - I checked the cut joins by `look --cuts` (frame pairs, pause lengths about 0.30 s each) and by `listen --file`, but the isolated listen re-transcription was not reliable enough to confirm the audio at each join.
  - The `talk.mp4` source was only 30 s, so I did not check for pacing or energy beyond the removed pauses and the two cut segments.

- **Best part:** `cutroom edit cut --dry-run` showed the exact span removed and the new join, and `cutroom check` plus `look --cuts` gave a clear pass that I could trust.

- **One change:** allow `cutroom transcript` (and `check`'s word list) to take a rendered file such as `out/final.mp4` directly, so the final can be re-read and compared word-by-word without the `init` workaround.
