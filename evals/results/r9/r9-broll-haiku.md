# Friction report (r9-broll-haiku)

- **Outcome:** done — out/final.mp4 (13.767 s) rendered and `cutroom check` passed with 0 failures and 0 warnings; I did not run `cutroom listen` on the joins, so the audible result is unverified beyond the checks.

- **Commands run:** about 31 cutroom invocations, in order: `guide`; `guide workflow`, `guide edit`, `guide overlays`, `guide render`, `guide verify`, `guide pitfalls`; `init talk.mp4 terminal.mp4`; `inspect`; `perceive` (guessed, stderr hidden, output discarded); `transcribe`; `transcript`; `transcript talk --suspects`; `transcript terminal`; `edit tighten`; `timeline`; `timeline --joins`; `edit cut --from "The first thing you need" --near 6.31 --to "first thing." --source talk --dry-run`; the same without `--dry-run`; `edit cut --from "By the way" --to "keyboard." --dry-run`; the same without `--dry-run`; `timeline --joins`; `overlay add clip terminal.mp4 --at "open the terminal" --to "update command."`; `overlay list`; `timeline`; `look --output 10.15`; `render --preview --range 0:08-0:13`; `render --dry-run`; `render`; `check`; `transcript --file out/final.mp4`; `status`.

- **Errors hit:** None surfaced. I hid stderr on `cutroom perceive`, which I had guessed exists, so I do not know what it printed. No `fix:` lines were needed.

- **Guessed wrong:** `cutroom perceive` does not exist as far as I can tell (I never saw its output). I also assumed `edit cut --from ... --to ...` would keep the join pause sensible; the join output confirmed it.

- **Confusing output:** `look` labels read `out 0:10.150 = talk 24.85 +o1`, which mixes output and source times on one line; it took a second read to see which was which. The `retake-kept` warning's suggested command is long; I copied it nearly verbatim, which worked. The `transcript --suspects` output says "1 breath(s) not listed" without saying which, so I could not tell whether an aside was hidden there.

- **Missing:** `transcript --suspects` does not flag an off-topic aside. The "By the way, my cat just walked across the keyboard." sentence was only found by reading the transcript myself, and that was a judgement call. I also had no way to check what terminal.mp4 actually shows beyond a single frame from `look`, and I did not confirm the clip shows the update command. `overlay add` has no `--dry-run`, so I added the overlay without a preview.

- **Not sure it's right:** The false-start cut kept "The first thing you need is an old laptop." and removed the earlier "…is, UM, the first thing." take; the join reads correctly in the transcript. The cat cut joins "install Linux on it." straight to "Finally, open the terminal…" with a 0.33 s pause; I judged it natural but did not listen. The B-roll window is 9.000–11.300 s, starting from the terminal clip's first frame (clip-in 0); `check` confirms the frames match the clip, not that the clip is relevant to the words. Word checks confirm the spoken words are in order, not that the recogniser (base.en) heard them correctly.

- **Best part:** `edit cut --dry-run` showed the join and the removed duration before applying, and `check` gave a per-item line for every overlay and sync point, so the result was easy to verify.

- **One change:** Have `transcript --suspects` flag sentences that do not relate to the rest of the talk (asides, tangents), so off-topic cuts are found without reading the whole transcript by hand.
