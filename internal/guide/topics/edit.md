<!-- summary: building the timeline — cut and keep by quoted words, tighten pauses, check joins, undo -->
# Edit: the timeline

The timeline (`timeline.json`) is the list of source ranges the video keeps, in order. It is created
by your first edit as the whole first video source. Sources are never touched.

## The usual sequence for a talking video
1. `cutroom edit tighten` — shortens every long pause (keeps ~0.32 s), trims silence at the ends.
2. Read `cutroom transcript` and `--suspects`. For each retake, false start, or off-topic aside:
   `cutroom edit cut --from "first words" --to "last words"`. Quote the transcript's words exactly
   as printed (case, punctuation and um/uh don't matter, but include an "um" if it tells two takes apart).
3. `cutroom timeline` — every segment with its words, plus problems cutroom found
   (`retake-kept`, `kept-sound`, `edge-in-word`, …). Fix each one or decide it is intended.
   A `retake-kept` warning right after tighten is expected: it prints a ready `edit cut` that
   removes the earlier take(s) and keeps the last; run it with `--dry-run` first to see the join.
4. `cutroom timeline --joins` — every cut point with the words either side and the pause across it.
   Read them like a viewer would hear them. `cutroom listen` any join you doubt.

## Cutting
- `--text "exact words"` cuts just those words; `--from "…" --to "…"` cuts a span (inclusive).
  `--from` must be unique: if it repeats you get a numbered list, then add `--occurrence N` or
  `--near <source time>`, or quote more words. `--to` is the first match after `--from`.
- Cuts by words land inside the pauses: 0.22 s is kept after the previous word, 0.10 s before the
  next (word endings trail off; this keeps joins natural). A warning means a cut is inside
  continuous speech — listen to it.
- `cutroom edit cut 16.9-17.4` cuts a source time range (for sounds, not words).
- `cutroom edit keep --from "…" --to "…"` keeps only that span (e.g. to start a short).
- Explicit cuts are pinned: a later `tighten` never moves them.
- `cutroom edit audio <source>` takes the sound from another source (a recorder) once
  `cutroom sync <source>` has aligned it.
- `--dry-run` previews; `cutroom edit undo` reverts the last edit; `cutroom edit history` lists them.

Times in `edit` are **source** times (what `transcript` prints). Times in `timeline` output
columns are **output** times (where it lands in the video), followed by the source range.
