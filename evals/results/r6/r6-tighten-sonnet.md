# Friction report

- **Outcome:** done. final.mp4 is 13.77 s, `cutroom check` PASS with 0 warnings, and the re-transcript of the render reads as the intended script.
- **Commands run:** 11 (not counting `ls`): guide; guide --all; doctor; init talk.mp4; transcribe; transcript; transcript --suspects; edit tighten; edit cut x2 with --dry-run; edit cut x2 for real; timeline --joins (printed by the cuts); render --preview; render; check; transcript --file out/final.mp4.
- **Errors hit:** none.
- **Guessed wrong:** nothing. Quoting the transcript words for --from/--to worked first time.
- **Confusing output:** `edit tighten` prints a long retake-kept warning that is easy to skim past. The "um" is shown as "UM" in the transcript, but the guide says case is ignored, so it was fine.
- **Missing:** the off-topic aside (the cat) is not flagged by any suspect; I found it by reading the transcript. A rough off-topic hint would help. Dry-run cut edges were 0.2 s earlier than the printed phrase times; fine, but I did not check it by ear.
- **Not sure it's right:** I never ran `look` or `listen` on the joins. The checks pass and the words match, but the picture at the cuts is unseen. The "kept" retake is the one I judged the good take from the base.en transcript only.
- **Best part:** the suspects list plus quoted-text cuts, with `--dry-run` showing the join. `check` covering words, sync and stray sounds in one command.
- **One change:** flag off-topic/aside phrases (e.g. "by the way") in --suspects.
