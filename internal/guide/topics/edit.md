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
   removes the false start — usually the earlier take(s), but the later one when it breaks off ("…
   worrying–") after a complete sentence; run it with `--dry-run` first to see the join.
4. `cutroom timeline --joins` — every cut point with the words either side and the pause across it.
   Read them like a viewer would hear them. `cutroom listen` any join you doubt.

## Cutting
- `--text "exact words"` cuts just those words; `--from "…" --to "…"` cuts a span (inclusive).
  `--from` must be unique: if it repeats you get a numbered list, then add `--occurrence N` or
  `--near <source time>`, or quote more words. `--to` is the first match after `--from`.
- Cuts by words land inside the pauses: 0.22 s is kept after the previous word, 0.10 s before the
  next (word endings trail off; this keeps joins natural). Where words run together with no pause,
  the cut goes to the quietest frame of the measured sound near the word boundary. A warning means
  a cut is inside continuous speech — listen to it.
- `cutroom edit cut 16.9-17.4` cuts a source time range (for sounds, not words). Prefer words: the
  transcript's word starts run 20–80 ms late, so a range ending "just before" a word keeps a piece
  of it. `edge-in-sound` flags any segment edge that sits in speech, with the cut that moves it to
  the quiet point.
- `cutroom edit keep --from "…" --to "…"` keeps only that span (e.g. to start a short).
- Explicit cuts are pinned: a later `tighten` never moves them. `cutroom edit pin s3 [--in|--out]`
  pins edges you set some other way (`--unpin` releases them).
- `cutroom edit split --text "Then install"` (or a source time: `edit split 12.4`) splits a segment
  in two without removing anything — in the pause before those words. Then move, zoom or cut each
  part on its own (e.g. zoom only the second half of a long take).
- `cutroom edit audio <source>` takes the sound from another source (a recorder) once
  `cutroom sync <source>` has aligned it.
- `cutroom edit zoom 1.15 --from "…" --to "…"` punches in on a span (zoom 1 undoes it).
  `--anchor x,y` (0–1, default 0.5,0.4 ≈ a face) is the point to zoom on; the crop stays inside the
  frame. Keep the upscale ≤ 1.2× (a 1080p source at 1.3× looks soft; 4K sources can go 2× for a
  1080p output) — edit warns above it. Check the framing with `cutroom look --output <time>`;
  `timeline` shows `[zoom …]` on zoomed segments and `--joins` marks zoom changes (not cuts).
- Every WARN is decided: fix it, or keep it on purpose. `cutroom timeline --apply retake-kept@4.09`
  runs the warning's own suggested edit; `cutroom timeline --accept audio-mixed --reason "same mic"`
  keeps it — still listed, no longer open (status stops pointing at it). Both are undo-able.
- `--dry-run` previews; `cutroom edit undo` reverts the last edit; `cutroom edit history` lists them;
  `cutroom edit reset` starts over from the whole video (itself undo-able) — the quickest fix when an
  inherited edit is wrong in several places. `cutroom edit <sub> --help` shows one subcommand.

Times in `edit` are **source** times (what `transcript` prints). Times in `timeline` output
columns are **output** times (where it lands in the video), followed by the source range.
