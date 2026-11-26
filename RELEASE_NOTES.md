# cutroom v1.1.0

Looks, framing, and tighter checks on top of v1.0.0. Same idea: perceive → decide → act → verify.
Apache-2.0. Thirty-three commits since v1.0.0.

## Style

`cutroom style` sets the look of burned-in captions and text cards. Nothing about the default is
fixed: font, size, case, colour, box, outline, shadow, accent, placement, width, and tracking.
Styles layer (defaults ← `style cards` ← one card).

- **Motion:** `--animate` fade, pop, or slide; captions animate each group, cards come in and out.
- **Karaoke:** `--highlight` recolors the word being said; `--highlight-box` puts a rounded box
  behind it.
- **Cards:** `--eyebrow` and `--layout left|right`, an SVG `--icon` in an accent tile, `*emphasis*`
  in titles, letter-spacing. A text card stays up for its reading time; `--exact` ends it on the words.
- **Images and inset clips:** rounded corners, border, and shadow.
- **Named looks:** `style save` / `style use` / `style looks`, and `short --look`, so one channel
  look applies to every short. Lower thirds stay clear of where the captions actually sit. Accent
  bars follow a rounded card's corners.

## Framing

- Shorts frame the speaker by default (`edit frame --anchor auto`), including per segment
  (`--per-segment`). A plain `--fill` frames on the speaker too.
- Fitted frames: `edit frame --fit --bars blur` or a colour, instead of cropping.
- Punch-ins: `edit zoom --alternate` switches wide / zoom at every jump cut, on the speaker's face.
- A fill crop centres on its anchor. Vertical renders from HDR sources keep the source shape
  (they were squashed sideways).

## Speech and shorts

- `transcribe --vocab` for project names the recogniser should spell.
- Misheard words are flagged at `captions --burn` and corrected with `transcript --fix`.
- `short --cut-asides` cuts off-topic asides inside the span and names each one.
- `sync` says the alignment in plain words.
- A sound kept inside a sentence is something to listen to and keep, not an automatic cut.
- Caption groups respect cuts and clauses, and land on the exact frame.

## Verify

- `check` sums up the edit: the words kept and the phrases removed.
- Warns when a crop leaves the speaker out, and when a card, image, or captions cover the speaker's face.

## How it was verified

Agent evals, fresh agents given only the brief and `cutroom guide`, graded independently:

- Round 12 (styled short): sonnet and haiku both pass.
- Round 13: sonnet 13/13, haiku 4/4.
- Round 14 (house style: eyebrow, SVG icon, panel colour, left layout): sonnet and haiku both pass.

## Known limits

- Word checks re-transcribe with the installed model (base.en by default): they confirm the cut,
  not the recogniser; `transcribe --model` with a larger model checks both.
- The `.otio` export is verified by round-trip tests, not yet by the OpenTimelineIO reader.
- Fixture generation on Linux (espeak-ng path) is untested; development was on macOS.
- Face boxes are skin heuristics (a warm wall can read as a face). Variable fonts use the default
  instance. Open low-rated friction is in `BACKLOG.md`.
