# Friction report

- **Outcome:** done. Reset the inherited edit, redid it, rendered, `check` PASS with 0 warnings.
- **Commands run:** 8: guide; guide workflow/edit/verify + cat timeline.json; status + transcribe + transcript --suspects + transcript (one shell call, 4 commands); edit reset; edit tighten; edit cut (retake); edit cut (aside); timeline; timeline --joins; render; check.
- **Errors hit:** none.
- **Guessed wrong:** nothing significant. I did not use --occurrence; the unique-match worked.
- **Confusing output:** `transcribe` re-ran although status said "transcribed" (harmless). Inherited timeline.json was read by hand to spot the bad cut (1.9-2.9 removed "going to show you how"); `status` only showed warnings (broken-sentence), not which edit caused it.
- **Missing:** a one-line "what's wrong with this inherited edit" summary linking each warning to the edit that caused it.
- **Not sure it's right:** removing the "by the way" cat aside was my judgment (off-topic); check can't tell me that. Small model base.en used for word check.
- **Best part:** quoted-word cuts, tighten, and `edit reset` made a full redo 3 commands; check output is clear.
- **One change:** have `status` name the offending edit for each inherited warning.
