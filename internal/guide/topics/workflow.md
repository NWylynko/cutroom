<!-- summary: the steps from raw footage to a verified video, and which commands do each -->
# Workflow

Coming back to a project, or not sure where it stands? `cutroom status` shows sources, the timeline,
every render (current or stale, checked or not) and the next step.

1. **Set up.** `cutroom doctor` once. Then, in your working directory:
   `cutroom init <video> [<more files>]` creates `cutroom.json` and inspects every source.
   Each source gets an id (file name by default); other commands take ids or paths.
   Running `init` again is safe: it adds new files and re-inspects changed ones.
2. **Inspect.** `cutroom inspect [id]` shows frame rate, variable frame rate, rotation, HDR, the audio
   offset and loudness. Warnings marked `handled` need nothing from you. `ACTION` ones say what to do.
3. **Perceive the content.** `cutroom transcribe`, then `cutroom transcript` to read it phrase by
   phrase, `--suspects` for retakes and stray sounds, `--search` to find exact words. See
   `cutroom guide perceive`. `cutroom look` shows frames; `cutroom listen` settles doubts about audio.
4. **Edit the timeline.** `cutroom edit tighten`, then `cutroom edit cut --from "…" --to "…"` for
   retakes and asides; check `cutroom timeline` and `cutroom timeline --joins`. See `cutroom guide edit`.
   Text cards and images: `cutroom overlay add …` (see `cutroom guide overlays`).
5. **Render.** `cutroom render --preview`, check it, then `cutroom render`. See `cutroom guide render`.
6. **Verify.** `cutroom check` measures the render: frames, sync, every word, stray sounds,
   loudness. Fix what fails, re-render, check again. See `cutroom guide verify`.
7. **Captions** (optional): `cutroom captions` writes out/captions.srt. See `cutroom guide captions`.
8. **Export** (optional): `cutroom export --otio --edl` hands the edit to another editor.
9. **Shorts** (optional): `cutroom short --from "…" --to "…" --name x` makes a vertical short of a
   span in shorts/x/. See `cutroom guide shorts`.

## Separate audio recorder (a mic recorded apart from the camera)
1. `cutroom init talk.mp4 --source mic=recorder.wav` — add both.
2. `cutroom sync mic` — measures the offset and clock drift against the camera's own sound.
3. `cutroom edit audio mic` — the timeline now takes its sound from the recorder, aligned.
Transcribe and cut using the camera source as usual; renders and `check` use the alignment
automatically. Never point the timeline at another source's audio without `sync`: cutroom refuses.

## Several takes (re-recorded lines, a second camera roll)
1. `cutroom init take1.mov take2.mov` — the first edit starts from the longest video.
2. Cut the weak version from the main take; put the better one in from the other take:
   `cutroom edit add --source take2 --from "So the" --to "works." --after s3` (`--before`, `--after start`).
   A re-recorded line (pickup) swaps in one step: `cutroom edit replace --text "…" --with pickup`
   cuts the edit's version and puts the pickup's in its place.
3. `cutroom edit move s5 --after s2` reorders. `cutroom timeline` shows which take each segment is from.
4. Sound: each take uses its own sound unless a recorder is mapped to it —
   `cutroom sync mic2 --to take2` then `cutroom edit audio mic2 --for take2`. `timeline` compares the
   takes' speech level and room tone (`audio-mixed`): a note when they match, a warning with the
   numbers when they differ; `check` measures sync per segment.
