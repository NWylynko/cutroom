<!-- summary: text cards and images over the edit — anchored to words, looked at, rendered, checked -->
# Overlays

An overlay is a text card or an image shown over the edit. It is **anchored to words** (`--at`, the
words it appears on), not to a time: re-cut the edit and it still appears on those words.

```
$ cutroom overlay add text "Install Linux" --sub "Ubuntu Server 24.04" --at "Then install" --duration 2.5
added o1: text "Install Linux", lower-third, output 0:06.500–0:09.000 (2.500 s), appears on "Then install" (talk 0:14.020)
look at it: `cutroom look --output 7.75`
```

- `--duration 3` (default 4) or `--to "on it."` (until those words end).
- Layouts: `lower-third` (text default: left, above the bottom), `title` and `center` (big, centred),
  `top`, `corner` (top right; image default). `--sub` adds a smaller second line to text.
- `cutroom overlay add image logo.png --at "Finally" --duration 4 [--size 0.2]` — PNG or JPEG (alpha
  kept); `--size` is its width as a fraction of the frame.
- `--fade 0.2` (default) fades in and out; `--fade 0` cuts in and out on exact frames.
- `cutroom overlay list` shows each overlay's output time, or `NOT SHOWN` and why (its anchor words
  were cut — `cutroom timeline` warns `overlay-orphaned`). Two overlays in the same layout at the same
  time warn `overlay-overlap`. `cutroom overlay rm o1` removes one; `cutroom edit undo` brings it back.

Look before rendering: `cutroom look --output <time>` draws the cards exactly as the render will
(labels show `+o1`). Text is drawn by cutroom itself (no ffmpeg text filters needed, L47).

`cutroom check` then verifies each overlay on the render: visible inside its window, gone on the
frame before it starts and the frame after it ends (`ok overlay o1 … visible over frames 195–269`), and
it compares the rest of the picture around the cards.
