# Friction report: r7-fix-edit-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s) is rendered from the corrected timeline and passes `cutroom check` with 0 failures and 0 warnings, and its transcript reads as one clean explanation with no suspects.

- **Commands run:** about 30 cutroom invocations across 10 shell calls, in this order: `cutroom guide`; `guide workflow`, `guide edit`; `guide verify`, `guide render`, `guide perceive`; `status`; `timeline`; `transcript`; `transcript --suspects`; `edit history`; `guide pitfalls`; `edit reset`; `edit tighten --min-cut 0.4 --dry-run`; `edit tighten --min-cut 0.4`; `edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`; `edit cut --from "By the way, my cat" --to "across the keyboard." --dry-run`; `edit cut` (same two, applied); `timeline`; `timeline --joins`; `render --preview --dry-run`; `render --preview`; `check out/preview.mp4`; `transcript --file out/preview.mp4`; `look --cuts --file out/preview.mp4`; `render`; `check out/final.mp4`; `transcript --file out/final.mp4 --suspects`; `status`. Did not run `cutroom doctor` or `cutroom init` (the project already existed).

- **Errors hit:** none. No command failed. The inherited timeline raised warnings instead:
  - `WARN[edge-in-word] at output 0:01.433: segment s1 ends inside the word "I'm"`
  - `WARN[broken-sentence] at output 0:01.433: the cut at 1.433 removes "going to show you how" from inside a sentence`
  - `WARN[retake-kept] at output 0:03.093: "the first thing you need is" is still said 2× in the timeline`
  The `fix:` lines were useful. `edit reset` plus re-tightening was simpler than the suggested undo, and the `retake-kept` fix line was usable as written.

- **Guessed wrong:** nothing failed, but I assumed `edit tighten --min-cut 0.4` was the right setting. I took it from `edit history` rather than the guide, which says only "shortens every long pause (keeps ~0.32 s)". It produced the intended result.

- **Confusing output:**
  - `edit tighten` said "2 segment edges trimmed" without naming them. I had to infer which edges moved from `timeline` afterwards.
  - The `retake-kept` fix line gives a `--near 6.31` source time, but the `--to "first thing."` span it describes (6.310–9.220) does not match the segment's words (6.200–9.433). I used the plain `--from`/`--to` form instead. Both worked, but the mismatch made me re-check.
  - The render transcript lowercases the sentence start after the comma ("…is an old laptop, then install Linux on it."). That looks like a broken sentence to a reader, though the render is correct.
  - `timeline` prints the full warning text with a fix for each. That is more useful than the summary line, but it is long.

- **Missing:**
  - A single view of an inherited edit against the transcript: for each sentence, whether it is kept whole, trimmed, retaken, or cut. I had to combine `transcript`, `timeline` and `timeline --joins` to get it.
  - I did not check the original `out/final.mp4` before resetting. `status` showed it as "current, not checked", and I went straight to the timeline. A quick `check` on the original would have shown what the viewer actually saw.

- **Not sure it's right:**
  - The transcript is from `base.en` and cannot show whether the sound at the joins is clean. I did not run `cutroom listen` on any join. The `look --cuts` PNG shows the frame cuts (1–4 of 4) and the words either side, and the 0.30–0.33 s pauses at each join.
  - "By the way, my cat…" was removed as an off-topic aside. That is a judgement from the words alone, not a viewing of the picture.
  - The sentence "Hi, today I'm going to show you how to set up a home server." is kept whole in s1 (source 0.467–4.400). The original had cut mid-sentence, so I assume the reset fixed the edge-in-word and broken-sentence warnings, but I did not listen to confirm it.

- **Best part:** `cutroom edit cut --from … --to … --dry-run` printed the join ("…up a home server. | The first thing you…", pause 0.30 s) and the resulting length before applying. `timeline` plus the warnings with fix lines made the inherited problems easy to find, and `check` gave a clear PASS.

- **One change:** an `inspect`-style summary of an inherited timeline that compares it with the transcript sentence by sentence (kept whole / trimmed / retake / aside / missing), so the agent can see every error in one screen.
