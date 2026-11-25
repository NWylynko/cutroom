# Friction report

- **Outcome:** done. `out/final.mp4` (13.767 s, 1280x720) renders, passes `cutroom check` with 0 failures and 0 warnings, and contains the retake, aside and both text cards as specified. I checked the card frames on the preview render (`.cutroom/look/overlays.png`, `output_1_900`, `output_7_750`) but did not view final-render frames directly.

- **Commands run:** about 35 cutroom invocations in 11 shell calls, in order: `cutroom guide`; `cutroom guide workflow`, `edit`, `overlays`, `verify`; `cutroom guide perceive`, `render`, `pitfalls`; `cutroom doctor`; `cutroom init talk.mp4`; `cutroom inspect`; `cutroom transcribe`; `cutroom transcript`; `cutroom transcript talk --suspects`; `cutroom transcript talk --words --range 5-13`; `cutroom edit tighten --dry-run`; `cutroom edit tighten`; `cutroom timeline`; `cutroom edit cut --from "The first thing you need" --near 6.31 --to "first thing." --dry-run`; `cutroom edit cut --from "By the way," --to "the keyboard." --dry-run`; `cutroom edit cut` (both, for real); `cutroom timeline`; `cutroom timeline --joins`; `cutroom transcript --output`; `cutroom overlay add text "Home Server in 3 Steps" --layout title --at "Hi, today" --to "home server."`; `cutroom overlay add text "Step 2: Install Linux" --layout lower-third --at "Then install" --duration 2.5`; `cutroom overlay list`; `cutroom timeline`; `cutroom render --preview --dry-run`; `cutroom render --preview`; `cutroom check out/preview.mp4`; `cutroom look --overlays`; `cutroom look --output 1.9`; `cutroom look --output 7.75`; `cutroom render`; `cutroom check out/final.mp4`; `cutroom transcript --file out/final.mp4`.

- **Errors hit:** none. No command failed and no `fix:` line was needed.

- **Guessed wrong:** nothing material. I expected `--from` on `edit cut` to need a unique phrase. It did, but the `retake-kept` WARN printed the exact command with `--near`, so I copied it.

- **Confusing output:**
  - Every `cutroom timeline` call reprints the full WARN block (`retake-kept`, `aside-kept`) and the `decide each WARN` footer, including after the warnings were resolved in the same session. I had to check by hand whether a WARN was new or stale. The `edit cut` output also says "still open (unchanged by this edit)", which is useful, but it is one more line to parse.
  - `timeline --joins` writes segment IDs as `s1|s3`, and the `…text… | text…` join lines are compact but not obvious on first read.
  - `cutroom edit tighten --dry-run` and the real run print nearly the same block; the dry run was useful, but I had to read both to confirm that the real run matched.

- **Missing:**
  - A way to see overlays on the timeline output. `timeline` lists segments and overlays only as a footer line (`overlays: o1 at 0:00.100, o2 at 0:06.500`), so the overlay's end time and its anchor words are only in `overlay list`.
  - A single "is this done" summary across timeline warnings, overlays and the last render. `cutroom status` exists, but I did not run it, so I can't say whether it covers this.

- **Not sure it's right:**
  - The title card runs 0:00.100–0:03.700 and the lower-third 0:06.500–0:09.000. Both are placed by anchor words, and the card frames I looked at (preview render) were correct. The final render's overlay visibility is confirmed only by `check` (`overlay o1 visible over frames 3–110`, `overlay o2 visible over frames 195–269`), not by my own look at final frames.
  - The transcript that `check` used to confirm words is base.en. It said "all 44 expected words heard in order", which confirms the cut but not the recogniser. Some word boundaries, such as "The first thing you need is an old laptop," could have been misheard in the source, and I did not verify them with a larger model or `cutroom listen`.
  - The aside removal ("By the way, my cat just walked across the keyboard.") is a judgement call (off-topic, so cut). The task asked for that, but the cut's join ("…install Linux on it. | Finally, open the terminal…", pause 0.33 s) was read only from `--joins` output, not listened to.

- **Best part:** the WARN lines carry ready-to-run commands, and `--dry-run` on `edit cut` shows the join and the output length before anything changes. Together that made the whole retake and aside pass quick to do safely. `check` then gave a clear per-item pass list.

- **One change:** have `cutroom timeline` mark each WARN as new or already printed in this session (or print only changes since the last `timeline` call), so an agent can tell at a glance which warnings it has not yet handled.
