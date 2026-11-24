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
5. **Render.** `cutroom render --preview`, check it, then `cutroom render`. See `cutroom guide render`.
6. **Verify.** `cutroom check` measures the render: frames, sync, every word, stray sounds,
   loudness. Fix what fails, re-render, check again. See `cutroom guide verify`.

## Separate audio recorder (a mic recorded apart from the camera)
1. `cutroom init talk.mp4 --source mic=recorder.wav` — add both.
2. `cutroom sync mic` — measures the offset and clock drift against the camera's own sound.
3. `cutroom edit audio mic` — the timeline now takes its sound from the recorder, aligned.
Transcribe and cut using the camera source as usual; renders and `check` use the alignment
automatically. Never point the timeline at another source's audio without `sync`: cutroom refuses.
