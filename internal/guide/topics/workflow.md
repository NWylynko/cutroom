<!-- summary: the steps from raw footage to a verified video, and which commands do each -->
# Workflow

1. **Set up.** `cutroom doctor` once. Then, in your working directory:
   `cutroom init <video> [<more files>]` creates `cutroom.json` and inspects every source.
   Each source gets an id (file name by default); other commands take ids or paths.
   Running `init` again is safe: it adds new files and re-inspects changed ones.
2. **Inspect.** `cutroom inspect [id]` shows frame rate, variable frame rate, rotation, HDR, the audio
   offset and loudness. Warnings marked `handled` need nothing from you. `ACTION` ones say what to do.
3. **Perceive the content.** `cutroom transcribe`, then `cutroom transcript` to read it phrase by
   phrase, `--suspects` for retakes and stray sounds, `--search` to find exact words. See
   `cutroom guide perceive`. *(Looking at frames and listening closely: not available yet.)*
4. **Edit the timeline.** *(not available in this version yet)*
5. **Render** a preview, then the final. *(not available in this version yet)*
6. **Verify** with `cutroom check` before calling anything done. *(not available in this version yet)*

Separate audio recorder? Add it as a source too (`cutroom init --source mic=recorder.wav`).
