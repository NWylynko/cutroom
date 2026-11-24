<!-- summary: proving a render is right with `cutroom check`, and what to do with each finding -->
# Verify

Run `cutroom check` after every render (it checks out/final.mp4, else out/preview.mp4; or name a
file). It measures the file itself:

| check | what it means if it fails |
|---|---|
| frames, av-length | the render is short or audio and video lengths differ: re-render |
| format | wrong size, rotation metadata left in, or untagged colour |
| picture | frames don't match their source frames (shifted/frozen video); it says where |
| sync | audio vs picture per segment, limit 45 ms. Separate recorder? Align it first |
| sync-camera | separate recorder only: the render's sound vs the camera's own sound, so a wrong `sync` offset can't pass |
| words | the render re-transcribed vs the timeline's words, in order |
| clipped-word | a cut goes through a word; cut by words instead |
| stray-sound | sound with no words kept (breath, bump, piece of a cut take) |
| loudness | integrated loudness / peak vs the render settings |
| stale | the timeline changed after this render: render again |

`PASS` with warnings is not automatically done: read each `WARN` and either fix it or decide it is
fine. Word differences are often the recogniser mishearing (especially with small models); settle
each with `cutroom listen` around the source time before changing the edit.

Exit code 4 means a check failed (so `cutroom render && cutroom check` stops at check, not at render). Fix, re-render (unchanged chunks are cached), check again.
