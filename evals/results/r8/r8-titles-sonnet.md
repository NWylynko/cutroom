# Friction report

- **Outcome:** done. out/final.mp4 is 13.77 s (from 30.26 s); `cutroom check` PASS, 0 failures, 0 warnings, and both cards were checked visually with `cutroom look`.
- **Commands run:** 14 (including guide x3): guide; guide workflow; guide edit; guide overlays; init; transcribe; transcript; transcript --suspects; edit tighten; edit cut (retake); edit cut (cat aside); edit cut 16.9-17.4; timeline; overlay add title; overlay add lower-third; overlay list; render; check; overlay set o2 --duration 2.1; render; check; look x2.
- **Errors hit:** `error[nothing-cut]: nothing to cut in talk 16.9-17.4` - my own redundant cut; tighten had already removed the cough. Its fix line was accurate.
- **Guessed wrong:** nothing significant. I wasn't sure whether `--to "home server."` matched "server." with punctuation; it worked.
- **Confusing output:** the tighten output says "this also removes the sound with no words" which is helpful. o2 default window (1.43 s) was shorter than reading time, so I extended it to 2.1 s, which now runs ~0.47 s into the next sentence ("Finally...") - the tool gave no way to see this except by arithmetic.
- **Missing:** a way to make a card extend only within its spoken words yet satisfy readability (maybe warn when an extension crosses into other words). `look` shows only a single frame per call; fine.
- **Not sure it's right:** the base.en transcript is coarse; "Linux" step number is taken from the user's text. Title card runs 0.1-3.7 s, covering the whole first sentence.
- **Best part:** quoted-word anchors for cuts and overlays, `timeline` warning with a ready-made cut command, and `check` verifying overlays per frame.
- **One change:** make `overlay add` auto-extend (or warn up front with a suggested --to) when the spoken anchor is shorter than the readable time, so no second render is needed.
