# Friction report

- **Outcome:** done. out/final.mp4 is 1280x720 SDR bt709, upright, 13.767 s, `cutroom check` PASS with 0 warnings.
- **Commands run:** 9 cutroom calls. guide, guide --all, init, inspect, transcribe, transcript, transcript --suspects, edit tighten, edit cut (x2), timeline --joins, render, check, look out/final.mp4 3.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant. I cut the first "The first thing you need is, UM, the first thing." by quoting it myself rather than using the printed retake-kept command; it worked.
- **Confusing output:** `guide --all` is long (I tailed it). The retake-kept WARN printed after tighten was stale by the time I'd cut, harmless.
- **Missing:** nothing needed. HDR/rotation needed no flags, which was great, but I had no way to look at the source's raw colour for comparison beyond the `tone` check line.
- **Not sure it's right:** the look frame was a test pattern; I could see colours were not washed out and the "3" digit read upright, but I did not compare against the source frame side by side. The `tone` check (saturation 87 vs 87) covers this. Transcript used the small base.en model.
- **Best part:** init/inspect flagged rotation and HDR as "handled", and render+check explicitly verified tone mapping, rotation baked in, and sync. Text anchors for cuts were easy.
- **One change:** have `edit tighten` not emit a retake-kept WARN that suggests a command, or make `guide workflow` say explicitly that HDR/rotation need no action, so an agent doesn't wonder.
