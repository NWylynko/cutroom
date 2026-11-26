<!-- summary: the look of captions and text cards — fonts, colours, boxes, outlines, placement, karaoke highlight -->
# Style

cutroom has a default look (bold white captions on a dark box; dark cards with an amber bar), but
nothing about it is fixed. `cutroom style` sets the look of **burned-in captions**, of **every text
card**, or of **one card**; anything you don't set stays at the default. Styles layer: defaults ←
`style cards` ← `style o3`.

```
$ cutroom style captions --font "Avenir Black" --case upper --size 0.08 --box none \
    --outline black --outline-width 0.12 --highlight "#ffd400" --words 2 --y 0.62
captions: font=Avenir Black, case=upper, …
  font           Avenir Black → Avenir Black (/System/Library/Fonts/Avenir.ttc)  set
  size           0.08                                     set
  …
next: cutroom look --output 1.00   (see it)
```

Fields (`cutroom style --help` lists them all; `--x default` unsets one, `--reset` all):
- **font** / **sub-font**: `bold`, `regular`, `mono` (built in), a `.ttf`/`.otf`/`.ttc` file (relative
  to the project, `name#2` = the 3rd face of a collection), or an installed font's name.
  `cutroom style fonts [query]` lists installed fonts. A font that can't be found is refused when set.
- **size** / **sub-size**: text height as a fraction of the frame's short side (captions 0.068).
- **color**, **sub-color**, **box**, **outline**, **shadow** (under the letters), **box-shadow** (soft,
  under the box), **accent** (the bar on lower thirds), **highlight**: `#rrggbb`, `#rrggbbaa` (aa =
  opacity: `#00000099`) or a name; `none` turns that part off.
- **highlight** (captions): karaoke — each word turns this colour while it is said.
- **words** / **chars** (captions): how much shows at once (default 3 words, 20 characters).
- **case** (upper | lower | as-is), **align** (left | center | right), **radius** and **pad**
  (fractions of the font size), **outline-width** (fraction of the font size, default 0.08).
- **x** / **y**: the centre of the box as a point of the frame (0–1 across, down), like
  `edit frame --anchor`; unset, captions sit in the caption band and cards follow `--layout`.
- **width**: the widest line as a fraction of the frame width; longer text wraps.
- **animate**: `none`, `fade`, `pop`, `slide-up`, `slide-down`, `slide-left`, `slide-right`.
  Captions animate each group as it appears (karaoke: once per group, not per word); cards come in
  and go out with it, on top of their `--fade` (`overlay set o1 --fade 0` for motion only).
  **animate-time**: seconds it takes (captions 0.15, cards 0.3).

## Looks: one channel look for every short

Set captions and cards up once, then keep them as a named look and apply it anywhere:
```
$ cutroom style save channel                 # this project's caption + card style
$ cutroom short --from "…" --to "…" --name tip --look channel   # a new short with that look
$ cutroom style use channel --project shorts/old   # or give an existing one the look (undoable)
$ cutroom style looks                        # what's saved
```
Looks live in your user config folder (`$CUTROOM_LOOKS` to choose another); give a path instead of a
name (`style save looks/brand.json`) to keep one next to your project and share it. A look replaces
the project's caption and card style; per-card styles (`style o3 …`) stay.

Look before rendering (`cutroom look --output <time>` draws them as the render will). On vertical
video `timeline` and `check` warn `overlay-unsafe` when text reaches where the apps draw their own
buttons (L41) — that is a warning, not a rule: move it (`--y`), shrink it, or keep it if you mean it.
Style changes are undoable (`cutroom edit undo`). Anything a style can't express — a logo, a
designed graphic — goes in as an image or clip overlay (`cutroom overlay add image card.png …`).
