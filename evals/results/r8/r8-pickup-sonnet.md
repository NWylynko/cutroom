# Friction report

- **Outcome:** done. out/final.mp4 (13.8 s) rendered and `cutroom check` PASSED with 0 failures and 0 warnings.
- **Commands run:** 11 (guide; guide workflow/edit/pitfalls as one command; init; transcribe; transcript; transcript --suspects; transcript pickup; edit tighten; edit cut --dry-run; edit cut x2; edit replace; timeline; timeline --joins; render; check). Roughly 9 separate shell calls.
- **Errors hit:** none.
- **Guessed wrong:** nothing material. I first assumed "the pickup line" would need `edit add` plus a cut. The workflow guide showed `edit replace --text ... --with pickup`.
- **Confusing output:** `transcribe` printed pickup first, then talk. `transcript` shows only talk unless you name pickup. The pickup was one line, so I had to match it by text with the talk. Cut commands printed "next: cutroom timeline --joins" even when it was already planned.
- **Missing:** nothing. A "which talk line does the pickup match" hint would be nice.
- **Not sure it's right:** I did not listen to the joins. The audio-mixed INFO said the levels match. Transcription used base.en, so the `words` check confirms the cut, not the recogniser.
- **Best part:** `edit replace --text --with pickup`, and `check` covering sync per segment.
- **One change:** have `transcript` list the other sources' phrases alongside the main one, so pickups can be matched without a second command.
