# Friction report

- **Outcome:** done. out/final.mp4 is 13.77 s with the cat remark, the pauses and the earlier "first thing" takes removed. The rotation is baked into the pixels, and `cutroom check` passed with 0 failures and 0 warnings.
- **Commands run:** 11.
  1. guide
  2. guide workflow / edit / pitfalls
  3. doctor
  4. init talk-rot.mov
  5. inspect
  6. transcribe
  7. transcript
  8. edit tighten
  9. edit cut --from "By the way" --to "keyboard"
  10. timeline
  11. edit cut (retake, the command the warning suggested), then render, then check
- **Errors hit:** none.
- **Guessed wrong:** nothing. I did not need to handle rotation myself, because render does it.
- **Confusing output:** `timeline` is long, and the retake warning repeats after every edit. The first `tighten` printed the retake warning before I had cut anything, which was a bit noisy. Cut times print as source times, which is easy to mix up with the output times in the same listing.
- **Missing:** a visual way to check that the picture is upright. `check` reports that sampled frames match, but I did not view a frame with `look`.
- **Not sure it's right:** I did not look at frames myself, so I am trusting the `picture` check that the video is upright. I also decided on my own to drop the earlier takes of "The first thing you need". The task did not ask for that, but they were clear retakes.
- **Best part:** the `retake-kept` warning came with a ready-to-run command. The render output said outright that rotation is baked in and the metadata is gone, which was exactly the requirement.
- **One change:** have `tighten` or `timeline` offer a single combined "suggested cuts" list (retakes plus asides). It could also print in the `render` output that the final is upright, which it already does.
