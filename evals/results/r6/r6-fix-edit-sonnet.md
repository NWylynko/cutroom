# Friction report

- **Outcome:** done. out/final.mp4 re-rendered; `check` PASS with 0 warnings; the render's transcript reads as one clean explanation.
- **Commands run:** 12 — `guide`, `guide --all`, `transcribe`, `transcript`, `transcript --suspects`, `timeline`, `edit history`, `edit undo` x2, `edit tighten`, `edit cut` (retake), `edit cut` (cat aside), `timeline`, `timeline --joins`, `render`, `check`, `transcript --file out/final.mp4`. (Several were chained with `&&`, so the call count is lower than the command count.)
- **Errors hit:** none.
- **Guessed wrong:** nothing material. I had to decide that "By the way, my cat..." is an off-topic aside to cut. The task wording did not say so.
- **Confusing output:** `edit undo` undid edits one at a time and the 2nd undo reported "no timeline now". That was fine but surprising. I could have undone in one step by rebuilding from scratch.
- **Missing:** nothing needed. A way to see the problems in the existing timeline (bad cut mid-sentence "I'm | to", retake kept, tight pauses) was `timeline` warnings, which worked well.
- **Not sure it's right:** the cat aside removal is a judgement call. The base.en model is small, so the transcript check is only as good as it. I did not listen to joins.
- **Best part:** `timeline` WARN lines (edge-in-word, broken-sentence, retake-kept) pointed straight at the faults and printed a ready-to-run cut command. `check` plus `transcript --file` closed the loop.
- **One change:** have `timeline` also flag a kept off-topic aside (e.g. "By the way") as a suspect, since that one needed my own judgement.
