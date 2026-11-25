# Friction report: r9-social-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s, 1280x720, burned-in captions) and `out/final.srt` (5 cues) both exist; `cutroom check out/final.mp4` passed with 0 failures and 0 warnings, and a re-transcription of the render reads the kept words in order.

- **Commands run:** about 37 `cutroom` invocations, plus `mkdir -p out`, `ls`, and reading PNGs. In order:
  `guide`, `guide workflow`, `guide perceive`, `guide edit`, `guide captions`, `guide render`, `guide verify`, `guide overlays`, `guide pitfalls`, `doctor`, `init talk.mp4`, `inspect`, `transcribe`, `transcript`, `transcript --suspects`, `transcript --words --from 6 --to 10`, `edit tighten`, `timeline`, `edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`, `edit cut --from "By the way" --to "keyboard." --dry-run`, `edit cut` (same two, for real), `timeline`, `timeline --joins`, `captions --output out/final.srt`, `captions --burn --output out/final.srt`, `status`, `render --preview --dry-run`, `render --preview`, `check out/preview.mp4`, `look --file out/preview.mp4 2.0`, `look --file out/preview.mp4 --count 6 6-12`, `look --cuts`, `render`, `check out/final.mp4`, `transcript --file out/final.mp4`, `status`, `edit history`.

- **Errors hit:** none. No command failed, so no `fix:` line was needed.

- **Guessed wrong:** Nothing failed, but `captions --burn` also writes the SRT, which I had to find out from its output line ("also wrote out/final.srt"). I ran `captions` twice as a result. Also `look --file` writes its PNG under `.cutroom/look/` inside the project, not under `out/`, which I had to discover from the "wrote" line.

- **Confusing output:** `transcript --suspects` printed `(1 breath(s) not listed: quiet, normal; --all shows them)`. That was clear, but I could not tell whether the hidden breaths included anything I should check. `status` printed `next:` lines that repeated `timeline --joins` and `render --preview`, which I had already run.

- **Missing:**
  - A way to listen to the sound at 0:16.970 before `tighten` removed it. `tighten` says it removes it, but I never confirmed it was non-speech.
  - A check that `out/final.srt` matches the burned captions. `check` covers the burned picture only. I could only confirm the SRT with `captions --check`, which I did not run separately; the first `captions` run printed the same checks.
  - A way to see the "words" check's model separate from the transcript's model. The `words` check re-transcribes with base.en, the same model that made the transcript, so it confirms the cut but not the recognition. The guide says this, but the check line itself makes it sound stronger than it is.

- **Not sure it's right:**
  - Retake choice: I kept "The first thing you need is an old laptop" (10.17 s source) and cut the earlier "…is, UM, the first thing." I chose it from the transcript and the `repeat` suspect only. I did not `listen` to either take.
  - The cat aside (18.5–21.9 s) was judged off-topic from the transcript text alone.
  - The base.en transcript may have mis-recognised words, so the caption text could contain errors I cannot see. I read the captions against the frames I looked at, not every cue.
  - The joins are at pauses of 0.30–0.33 s. `timeline --joins` read them as clean, but I did not `listen` to any join.

- **Best part:** `timeline` with its `WARN[retake-kept]` line gave the exact `edit cut` command, and `--dry-run` showed the join before I applied it. `check` gave a clear PASS list per check, and `look --cuts` made the joins easy to inspect.

- **One change:** Have `cutroom check` also verify the subtitle file (`out/final.srt`) against the timeline's words and timing, and report it as its own line. That would have confirmed the SRT deliverable with the same tool, instead of relying on the file written by `captions`.
