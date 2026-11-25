# Friction report: r10-tighten-haiku

- **Outcome:** done. out/final.mp4 (13.767 s, down from 30.261 s) passes `cutroom check` with 0 failures and 0 warnings, the cut frames were viewed in `look --cuts`, and the transcript reads as the intended script.

- **Commands run:** 28 cutroom invocations, in order: `guide`, `guide workflow`, `guide perceive`, `guide edit`, `guide render`, `guide verify`, `guide pitfalls`, `init talk.mp4`, `inspect`, `transcribe`, `transcript talk`, `transcript talk --suspects`, `edit tighten`, `transcript talk --words --from 6 --to 12`, `listen 16.5-17.8`, `edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`, `edit cut ... (applied)`, `edit cut --from "By the way," --to "the keyboard." --dry-run`, `edit cut --from "By the way," --to "the keyboard."`, `timeline`, `timeline --joins`, `render --preview --dry-run`, `render --preview`, `check out/preview.mp4`, `render`, `check out/final.mp4`, `transcript --file out/final.mp4`, `look --cuts`. Plus one image read of the cuts PNG outside cutroom.

- **Errors hit:** none. No command failed.

- **Guessed wrong:** nothing material. I expected `--from "The first thing you need"` to be ambiguous (it is said twice), and the tool's own suggested command with `--near 6.31` resolved it. I did not have to guess a flag.

- **Confusing output:**
  - `transcript --suspects` and the plain `transcript` both tag BOTH "The first thing you need is, UM…" and "The first thing you need is an old laptop." as `[repeat]`, with no sign of which take to keep. I only knew the keeper from the later `timeline` warning (`retake-kept` with a ready `edit cut`). The suspects output should say which take is the likely keeper.
  - `tighten` reports "9.800 s removed" but `timeline` then says 20.433 s; the two numbers come from different measurements (removed vs. remaining from 30.233 s). Took a moment to reconcile.
  - The `sound` suspect at 0:16.970 is inside a pause that `tighten` removes. The tighten output mentions this, but only after the fact, so I ran `listen` on it before realising it was already gone.

- **Missing:**
  - A way to tell `transcript --suspects` (or the transcript itself) which take is the keeper, or to list only the take that is retained.
  - Nothing else; the cut warnings with ready commands covered everything.

- **Not sure it's right:**
  - The transcript is from `base.en`, so words could be misheard. The render's `check words` step uses the same model, so it confirms the cut, not the transcription. I did not re-transcribe with a larger model (`cutroom transcribe --model`), so I have not verified the spoken words beyond the small model's output.
  - Removing "By the way, my cat just walked across the keyboard." is my editorial call that it is off-topic. The tool only flagged it as a probable digression.
  - The "Hi," opening was kept at 0:00.110. I did not listen to the first half-second.

- **Best part:** The `guide` topics plus the `WARN` lines with ready-to-run `edit cut` commands and `--dry-run` previews made the whole edit a few deterministic steps. `check` gave a clear PASS on both preview and final.

- **One change:** Have `transcript --suspects` mark the take the tool expects to keep (for example "keep" on the last take, "retake" on the earlier one), so the agent can pick the right span without first running `timeline` to get the `retake-kept` suggestion.
