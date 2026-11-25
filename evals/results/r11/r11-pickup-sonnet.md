# Friction report

- **Outcome:** done. out/final.mp4 rendered (13.8 s) and `cutroom check` PASS with 0 warnings.
- **Commands run:** 8 — guide; guide workflow + guide edit; init talk.mp4 pickup.mp4; transcribe + transcript + `transcript --source pickup --suspects` (errored); transcript pickup + edit replace --help; edit tighten + 2x edit cut + edit replace + timeline --joins; render + check; this report.
- **Errors hit:** `error[unknown-flag]: transcript has no flag --source` (the fix line showed positional usage; I got unstuck at once).
- **Guessed wrong:** assumed `transcript --source pickup`; it takes a positional id.
- **Confusing output:** a transcript for only one source is printed by default; the "also transcribed" hint is easy to miss. Otherwise clear.
- **Missing:** nothing important.
- **Not sure it's right:** I did not listen to the pickup join; check reported sync and words OK, and audio-mixed said levels match. The 2.13 s opening pause is tightened but not removed, which seems fine.
- **Best part:** `edit replace --text ... --with pickup` in one step, and the workflow guide's Several takes section described it directly.
- **One change:** let `transcript --source X` work or alias the positional id.
