# Friction report

- **Outcome:** done. out/final.mp4 is 13.8 s, two retake/aside removals, both cards placed, `cutroom check` PASS with 0 warnings; I viewed both card frames.
- **Commands run:** ~12 cutroom commands: guide; guide workflow/edit/overlays; init; transcribe; transcript; transcript --suspects; edit tighten; edit cut (retake); edit cut (cat aside); timeline; overlay add x2; overlay list; look x2; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant. I guessed `--layout title` exists on overlay add (it did).
- **Confusing output:** tighten output prints "retake-kept" warning with a ready-made cut using `--near`; I wrote my own cut instead, which worked, but the guide's `--from "..." --to "..."` with "UM" needed me to copy the transcript's odd capitalisation. Title card was 3.6 s (to end of first sentence) - fine, but its start (0.100) is before the first word, unexplained.
- **Missing:** nothing needed. Mildly: a way to view the whole card over time (e.g. first/last frame) in one look call.
- **Not sure it's right:** transcript is from base.en; I could not hear that the kept "first thing you need is an old laptop" take is the better-delivered one (only that it is the last). Lower-third lasts only 1.43 s (exact words "Then install Linux on it"), short but correct to the request.
- **Best part:** quote-anchored cuts and overlays, plus `check` verifying overlay windows and every word.
- **One change:** have `tighten` optionally apply the suggested retake cut (or make the warning's command a one-flag `--apply`).
