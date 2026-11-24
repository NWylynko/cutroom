# Friction report

- **Outcome:** done. out/final.mp4 rendered (17.233 s, down from 30.26 s) and `cutroom check` passed with 0 failures and 0 warnings. That includes sync at -0 ms median and worst on all 6 segments.
- **Commands run:** 12. guide, guide workflow, guide pitfalls, doctor, init, inspect, transcribe, transcript --suspects, transcript, guide edit, edit tighten, edit cut, timeline, timeline --joins, render --preview --range 0-8, render, check. (Several were chained in one shell call.)
- **Errors hit:** none.
- **Guessed wrong:** nothing. I expected to have to handle the VFR and audio offset myself, but both were reported as `handled`.
- **Confusing output:** `render --preview --range 0-8` rendered the whole 17 s timeline, because it renders whole chunks. The output explains this, but "range" suggests otherwise. It also produced out/preview-range.mp4, which I didn't check. The transcript marked both the false start and the real take as `[repeat]`, so I had to work out which to cut myself.
- **Missing:** nothing important. I'd have liked an explicit statement that the cat remark was kept on purpose, since the task only asked for pauses and the false start.
- **Not sure it's right:** I kept the "cat walked across the keyboard" aside because it wasn't requested to go. I didn't `listen` to the joins. The `sync` check says lip sync matches the source, and I trust it.
- **Best part:** `edit tighten` and quoted-text `edit cut` were fast. `check` reported the sync measurement in plain terms and covered VFR automatically.
- **One change:** have the tighten output name the false start directly, with a ready `cut` command. The retake-kept hint came with a command, but it printed the wrong span for my case, since the false start contained "UM".
