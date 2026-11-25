# Friction report (r8-pickup-haiku)

- **Outcome:** done. `out/final.mp4` (13.8 s) passes `cutroom check` with 0 failures and 0 warnings. The talk now reads "Hi, today I'm going to show you how to set up a home server. The first thing you need is an old laptop, then install Linux on it. Finally, open the terminal and run the update command. And that's it, your server is ready." The pickup's "Then install Linux on it." is in place, the false start and the cat aside are gone, and the long pauses are tightened. Not verified by ear: which take's audio plays on the pickup line (the timeline says pickup's own sound), and the isolated `listen` heard "Lineage" for "Linux", which I read as recogniser error, not a cut problem.

- **Commands run:** 35 cutroom invocations across 13 shell calls, in order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `cutroom guide perceive`, `cutroom guide edit`
  3. `cutroom guide render`, `cutroom guide verify`, `cutroom guide pitfalls`, `cutroom doctor`
  4. `cutroom init talk.mp4 pickup.mp4`, `cutroom inspect`, `cutroom transcribe`, `cutroom transcript`
  5. `cutroom transcript talk --suspects`, `cutroom transcript pickup`, `cutroom transcript pickup --suspects`, `cutroom transcript talk --words --from 6 --to 12`
  6. `cutroom edit replace --help`, `cutroom edit --help`
  7. `cutroom edit tighten`, `cutroom timeline`
  8. `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --source talk --dry-run`, `cutroom edit cut --from "By the way" --to "keyboard." --source talk --dry-run`
  9. `cutroom edit cut` (same two, applied), `cutroom edit replace --text "Then install Linux on it." --with pickup --dry-run`
  10. `cutroom edit replace --text "Then install Linux on it." --with pickup`, `cutroom timeline`, `cutroom timeline --joins`
  11. `cutroom render --preview --dry-run`, `cutroom render --preview`, `cutroom check out/preview.mp4`
  12. `cutroom render`, `cutroom check out/final.mp4`, `cutroom transcript --file out/final.mp4`
  13. `cutroom listen --output 5.4-7.4 --file out/final.mp4`, then I opened the PNG it wrote

- **Errors hit:** none. No command failed, and I never needed a `fix:` line.

- **Guessed wrong:** nothing material. I did not need `--json`. `cutroom guide` said the `--near` value is a source time, and that matched the `retake-kept` suggestion. I first assumed `edit replace` would only accept a text span, but `--help` showed it works as I needed.

- **Confusing output:**
  - The `replace` dry run said `pickup 0:00.380–0:02.130`, but the timeline afterwards showed `pickup 0:00.367–0:02.133`. The two numbers differ with no explanation, and the pickup's own transcript words were 0.48–1.91. I could not tell whether the difference is padding.
  - The `audio-mixed` INFO line repeated in four outputs (the replace, both timeline calls, and the joins output). It is useful once and noisy after that.
  - `cutroom timeline` printed `WARN[retake-kept]` after tighten, which had to be cleared by hand. The suggested fix is good, but the fix is not applied automatically.
  - `cutroom transcript --file out/final.mp4` prints the whole edit as one line. It does not show where the joins are, so you can't check the cuts from it. I used `timeline --joins` and `listen` instead.
  - The `cutroom check` output is clear. Each line states what it measured and the limit.

- **Missing:**
  - No way to see the pickup's and talk's word-level times alongside the output times in one view. I had to combine `transcript`, `timeline` and `timeline --joins` myself.
  - No `listen` or `look` summary of all joins at once. I checked one join by hand, and the others by reading `--joins`.
  - No command to show that a replaced segment actually uses the pickup's audio (the `audio-mixed` note only says "pickup's own sound"). I had to trust the timeline.

- **Not sure it's right:**
  - I checked only the join at 6.4 s by ear and spectrogram. The false-start cut and the aside cut were checked by dry-run and `timeline --joins`, not by `listen`.
  - The pickup clip has a 0.033 s audio offset, which the render says it kept "as recorded". I did not confirm lip sync visually.
  - The small model (`base.en`) misheard "Linux" as "Lineage" in the isolated listen, so the transcript-based word check proves the words are in the right order, not that they are the right words.
  - I looked at a spectrogram only for the one join. I did not look at any picture frames.

- **Best part:** `cutroom edit replace --text "…" --with pickup` did the whole swap in one step, and `timeline` plus `timeline --joins` plus `check` made it easy to confirm the result. The `retake-kept` warning also gave a ready-to-run cut command, so the retake removal took one copy-paste.

- **One change:** have `cutroom timeline` (or `timeline --joins`) print a per-segment table with the source audio file, the source word times, and the padding applied, so a replaced line can be checked at a glance.
