# cutroom v1 — release notes

cutroom is video editing for AI agents: it turns footage into text and images an agent can reason
over (transcripts, contact sheets, spectrograms), turns the agent's text decisions into a
frame-exact render, and then checks the render the same way. Perceive → decide → act → verify.
Apache-2.0.

## What's in v1

- **Start:** `doctor` (ffmpeg, whisper.cpp, a speech model; prints each fix), `guide` (the agent
  manual, by topic), `init`, `status`.
- **Perceive:** `inspect` (frame rate, VFR, rotation, HDR, audio offset, loudness and what each
  means for editing), `transcribe` (word level, two passes; flags retakes, asides, unworded sounds,
  doubtful timing), `transcript` (phrases, words, search, suspects, retakes, the edit's own words),
  `look` (frames, contact sheets, every cut, overlays, renders), `listen` (spectrogram, loudness,
  isolated re-transcription).
- **Edit:** `edit` cut/keep/add/replace/move/split/pin/zoom/frame/tighten/audio by quoted words or
  times, with undo/redo/history; `timeline` (segments, joins, warnings with runnable fixes,
  `--apply` / `--accept`); `overlay` (text cards, lower thirds, images, B-roll clips, anchored to
  words so they follow re-cuts); `sync` (a separate recorder: offset and clock drift, per take);
  `short` (vertical shorts with safe zones).
- **Render:** `render` (cached chunks, fast previews, one encode, HDR tone-mapping, rotation baked
  in, −16 LUFS), `captions` (SRT/VTT and burned-in, checked for reading speed and seams), `export`
  (OpenTimelineIO, CMX 3600 EDL).
- **Verify:** `check` measures the render independently of the timeline: frame count, picture vs
  source (upright, right moment), A/V sync per segment (against the recorder and the camera's own
  sound), every word in order, stray sounds, overlays, captions, tone mapping, loudness.

Sources are read-only, cuts land on whole frames and audio samples, and every command has `--help`,
`--json` and a `next:` hint; every error has a `fix:` line.

## How it was verified (all in the same round, 2026-11-25)

1. Every SPEC v1 command is implemented, documented in `guide`, and covered end to end:
   `make e2e` 257/257 on synthetic fixtures; `make check` green.
2. Every INVARIANT lesson in `docs/LESSONS.md` has a `TestLxx_…` guard; every DETECT lesson has a
   fixture that triggers it.
3. Agent evals (round 11): 12 tasks, fresh agents given only the brief and `cutroom guide`, graded
   by an independent grader (its own whisper pass, frame matching, sync cross-correlation):
   **sonnet 12/12, haiku 12/12**; round 10 was also 12/12 sonnet. Every friction item rated low.
   Grader self-test 26/26. (`evals/results/2026-11-25-11.md`)
4. Real footage (four excerpts of the maintainer's own recordings: single camera, 4K HDR with a
   separate recorder, two takes with a recorder each): fresh editors made each edit with cutroom;
   a reviewer accepted all four, and `check` passes on each (T3 round r51).
5. `go install` + `cutroom doctor` on a clean PATH lists each missing tool with its install command;
   with the tools present it reports all dependencies ok. A fresh agent followed the README quick
   start to a checked render.

## Highlights from the last iterations (real footage)

- Cuts where words run together go to the quietest frame of the measured sound, not the
  recogniser's word boundary (which runs 50–80 ms late).
- `edge-in-sound`: any segment edge that sits in speech is flagged with the trim to the nearest
  quiet frame — this caught every word fragment the real-footage reviewers heard, and nothing in
  the edits they accepted.
- `ends-mid-speech` / `starts-mid-speech`: an edit that opens or closes where the recording cuts a
  speaker off is warned, with the cut that ends on the sentence before.
- Recorder words cut off by the camera's last frame are kept; recorder words whisper couldn't place
  take their time from the camera microphone when a neighbour agrees.
- Output size is shown at render time and settable (`render --size 1080p`).

## Known limits and next steps

- Word checks re-transcribe with the installed model (base.en by default): they confirm the cut,
  not the recogniser; `transcribe --model` with a larger model checks both.
- The `.otio` export is verified by round-trip tests, not yet by the OpenTimelineIO reader.
- Fixture generation on Linux (espeak-ng path) is untested; development was on macOS.
- Low-rated friction from the last round is in `BACKLOG.md` (one-step retake/aside cuts after
  tighten, a 720p-appropriate default size for shorts, labelled times in suggested commands, …).
