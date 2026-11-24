# Friction report

- **Outcome:** done. out/final.mp4 is 13.77 s (from 30.26 s) and `cutroom check` passed with 0 failures and 0 warnings.
- **Commands run:** 9. `guide`, `guide --all`, `init talk.mp4`, `transcribe`, `transcript`, `transcript --suspects`, `edit tighten`, `edit cut --from "By the way" --to "keyboard"`, `timeline`, `edit cut --from "the first thing you need is" --near 6.31 --to "first thing."`, `timeline --joins`, `render`, `check`. (That is 13 calls in 9 shell invocations; several were chained with &&.)
- **Errors hit:** none.
- **Guessed wrong:** nothing. The retake-kept warning printed a ready-made cut command that I used as given.
- **Confusing output:** `cutroom edit tighten` printed the retake-kept WARN, and the following `edit cut` printed it again, then `timeline` printed it a third time. Repeating it was noisy. The `[repeat]` suspect covered the "um" false start and the retake, so I had to work out for myself that the first one was the one to drop.
- **Missing:** nothing needed. A built-in "is this off-topic?" hint would help; I judged the cat aside by reading the text.
- **Not sure it's right:** the transcript comes from the small base.en model. `check` re-transcribed the render and found all 44 words in order, but I did not look at or listen to the joins myself (no `listen`/`look --cuts`). Join pauses were 0.30-0.33 s.
- **Best part:** the copy-paste `edit cut` line in the retake-kept warning, and `check` listing each check with a plain verdict.
- **One change:** have `tighten` offer (or `--auto-retakes` apply) the retake cut itself, so the common case is one command.
