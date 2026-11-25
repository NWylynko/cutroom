# Friction report: r10-resume-haiku

- **Outcome:** done. `out/final.mp4` (13.767 s) passes `cutroom check` with 0 failures and 0 warnings, and the timeline has no open problems. The one open item I had, the retake, was fixed; the previous editor had already cut the cat aside.

- **Commands run:** 22 shell calls in order: `ls -la` and `cutroom guide`; `cat cutroom.json timeline.json`, `find`, `cutroom guide workflow`; `cutroom guide edit`, `cutroom guide perceive`; `cutroom status`, `cutroom edit history`, `cutroom timeline`, `cutroom transcript`; `cutroom transcript --suspects`, `cutroom edit cut ... --dry-run`, `cutroom timeline --joins`; `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing."`, `cutroom timeline`, `cutroom timeline --joins`, `cutroom render --preview --range 0:00-0:14`; `cutroom render`, `cutroom check out/final.mp4`; `cutroom transcript --file out/final.mp4`, `cutroom look --cuts`, `cutroom status`. Plus one image read of the cuts contact sheet.

- **Errors hit:** none. No command failed.

- **Guessed wrong:** nothing that broke. I expected `--range` to take output time and it did (the output line confirmed it: `range: output 0:00.000–0:13.767`), though `0:00-0:14` covered the whole edit, so the preview was not really partial.

- **Confusing output:**
  - The `next:` block at the end of `cutroom status` and `timeline` lists several overlapping suggestions (`--apply retake-kept@4.09`, `cutroom check out/preview.mp4`). The preview it points at is stale, so I had to work out which suggestion applied.
  - `status` still lists `out/preview.mp4` as STALE, not checked, after the final render. That is correct but noisy.
  - The retake warning's suggested command is long and has quotes inside a single line. It worked, but it is hard to read.

- **Missing:**
  - `cutroom check` verifies that the render matches the timeline and that the remaining words are heard in order. It does not tell you whether the cut content was the right content. The previous editor's `tighten --min-cut 0.4` and the earlier `cut` of the cat aside are only visible through `edit history`, and I had to trust them.
  - There is no single command that prints the final transcript next to the timeline's source ranges for a quick review. I used `transcript --file out/final.mp4`, which works but has no source-time mapping.

- **Not sure it's right:**
  - I did not watch or listen to the video. The contact sheet looked clean, and the transcript of `out/final.mp4` reads as complete sentences: "The first thing you need is an old laptop, then install Linux on it. Finally, open the terminal and run the update command. And that's it, your server is ready."
  - The opening "Hi, today I'm going to show you how to set up a home server." was kept, which I think is the intent.
  - I did not re-check the cut the previous editor made at 18.61–21.72 (the cat aside) or the sound at 16.97. Both are outside the timeline now, and `--suspects` reports them as already cut.
  - The `timeline --joins` pauses are about 0.30–0.33 s, which sounds natural in the transcript, but I did not listen to any join.

- **Best part:** the retake warning gave the exact `edit cut` command, `--dry-run` showed the join before I applied it, and `check` gave a clear PASS with per-check lines.

- **One change:** when `status` or `timeline` shows a `retake-kept` warning, add the join's pause and the words either side to the same line, so the decision is visible without a second `--joins` call. Also drop the stale preview from the `next:` suggestions once a final render passes.
