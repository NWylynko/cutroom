# Friction report

- **Outcome:** done. out/final.mp4 rendered (13.767 s), `cutroom check` PASS with 0 warnings; the frame I looked at shows an upright "2" and the check reports display rotation 0°.
- **Commands run:** 12 total. guide; guide --all | tail; init talk-rot.mov; inspect; transcribe; transcript; transcript --suspects (the last five were chained in one shell call); edit tighten; edit cut --from "By the way" --to "keyboard"; edit cut --from "The first thing you need is, UM" --to "the first thing"; timeline --joins; timeline; render --preview; render; check; look out/final.mp4 2.
- **Errors hit:** none.
- **Guessed wrong:** nothing material. I cut the "UM" false-start take as well as the cat aside, since the transcript flagged it as a retake and the task said "cut the cat and the long pauses". That was my judgement, not an instruction.
- **Confusing output:** `cutroom timeline` segment ids jump (s1, s3, s4, s6, s7). The retake-kept WARN stayed after the cat cut, and I had to work out that it was about the "first thing you need is" retake rather than the cat. The "this also removes the sound with no words at 0:16.970" note from tighten was useful.
- **Missing:** nothing. A one-line "rotation: upright in output" confirmation on render would be nice; `check` did give this in the format line.
- **Not sure it's right:** I judged orientation from one frame (the test-pattern digit looked upright). I didn't test in a rotation-ignoring player, but check reports rotation 0° and no metadata. I used the base.en model, so transcript word accuracy is unverified beyond check's "all 44 words heard".
- **Best part:** the guide's workflow was exact. Rotation is handled automatically, and `check` verifies format, picture, sync and words in one command.
- **One change:** have `render` print the output's rotation and orientation ("upright, rotation metadata 0") so the phone-footage requirement is confirmed without running `check` or `look`.
