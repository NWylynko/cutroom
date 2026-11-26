# Friction report

- **Outcome:** done. out/final.mp4 rendered (17.233 s), `cutroom check` PASS with 0 failures/warnings, sync worst 0 ms.
- **Commands run:** 9 cutroom calls: guide; guide workflow/edit/pitfalls; init; inspect; transcribe; transcript; edit tighten; edit cut --from "The first thing you need is, UM" --to "the first thing."; timeline; timeline --joins; timeline --accept aside-kept; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** the retake-kept warning suggested cutting the take that included "UM"; I cut that one and kept the clean take. Source times in warnings (e.g. 18.60) differ slightly from timeline columns, mildly confusing.
- **Missing:** nothing needed. A prompt saying whether to cut "By the way, my cat..." would help; I kept it (task asked only for pauses and false start) and accepted the warning with a reason.
- **Not sure it's right:** whether the user wanted the cat aside removed; whether the base.en word check misses subtle mishears. Picture sync is verified by check, not by eye.
- **Best part:** VFR and the 0.064 s audio offset were handled automatically and reported plainly, and check measured sync per segment.
- **One change:** have `tighten` output the single next command (cut the false start) directly instead of requiring reading the timeline warnings.
