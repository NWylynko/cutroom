# cutroom — spec

A video editing CLI whose user is an AI agent. Built on ffmpeg. Open source, generic.

**Thesis.** Agents can read text and look at images, but cannot watch video or hear audio. Every
editing decision depends on perception. cutroom's job is to (1) turn media into text and images an
agent can reason over, (2) turn the agent's text decisions into a frame-exact video, and (3) run the
same perception on the output to prove it matches intent. Perceive → decide → act → verify.

**v1 scope.** Footage with speech: talking heads, tutorials, podcasts, interviews, shorts. Cutting,
tightening, retake removal, punch-ins/reframing, image/clip overlays, built-in text and captions,
audio cleanup and loudness, multi-source (several takes, a separate audio recorder), SRT, OTIO export.
**Not v1:** motion graphics engine, colour grading beyond a LUT, multicam switching, effects.

## Design rules (agent ergonomics)

1. **Text in, text out.** Edits can be addressed by quoted transcript text ("from 'so the first'
   to 'anyway'"), segment IDs or times. Text is preferred: agents quote accurately and miscount seconds.
2. **Token-frugal by default.** Default output is compact text, ≤ ~40 lines, with counts and
   truncation notes ("… 212 more, use --all"). `--json` gives a complete, stable, schema'd object.
3. **Every response says what to do next.** A `next:` line with 1–3 concrete commands.
4. **Errors teach.** `error[CODE]: what happened` + `fix: exact command or field`. Unknown flags,
   fields and commands get "did you mean". Exit codes: 0 ok · 1 usage · 2 invalid project/timeline ·
   3 missing dependency · 4 check failed · 5 external tool failed.
5. **Images are for eyes.** Anything visual (`look`, `listen`) writes PNGs sized for vision models
   (≤ 1568 px long edge by default) and prints their paths plus a one-line text summary of what is in them.
6. **State lives on disk.** An agent with no memory runs `cutroom status` and knows everything: sources,
   what is stale, the last render and check, suggested next step.
7. **Safe by construction.** Sources are never written. Cuts snap to frames and samples. One lossy encode.
   Every LESSONS.md INVARIANT is enforced in code and guarded by a test named after its ID.
8. **Cost before work.** Long operations support `--dry-run` with a time estimate; long jobs print
   progress lines and end with a timing summary that feeds the estimator.
9. **Idempotent and cached.** Re-running a command with no changes is ~instant and says "cached".
10. **The manual ships in the binary.** `cutroom guide` is the agent's manual for this exact version.

## Project layout

```
<project>/
  cutroom.json        # project: sources (id → path, probed facts), settings, presets
  timeline.json       # the edit (see below); every edit command rewrites it and logs history
  .cutroom/           # caches, transcripts, renders' chunk cache, history, timings (disposable except history)
  out/                # renders, captions, exports
```

Sources are referenced by path, never copied or modified.

## Timeline format (`timeline.json`)

Flat and agent-editable; times are seconds as decimal strings or numbers on input, stored as
rational frame/sample counts alongside a readable seconds value.

```json
{
  "version": 1,
  "fps": "30/1", "width": 1920, "height": 1080, "sample_rate": 48000,
  "segments": [
    {"id": "s1", "source": "cam", "in": 12.4, "out": 31.85,
     "text": "So the first thing you want to do…", "reason": "skip false start",
     "pinned": false, "zoom": 1.0, "anchor": [0.5, 0.4]}
  ],
  "overlays": [
    {"id": "o1", "kind": "text|image|clip|captions", "at": {"text": "open the terminal"},
     "duration": 3.0, "layout": "lower-third", "content": "…"}
  ],
  "audio": {"source": "cam", "chain": "default-voice", "loudness": -16},
  "output": {"preset": "youtube-1080p"}
}
```

Overlays are anchored to transcript text or to a segment; output times are derived, so re-cutting
never breaks them.

## Commands (v1)

| Command | Purpose |
|---|---|
| `guide [topic]` | Agent manual: workflow, timeline, editing, verification, pitfalls, recipes |
| `doctor` | Dependencies present and versions; exact install commands for what's missing |
| `init [--source id=path]…` | Create a project, probe sources |
| `status` | Project state, staleness, next steps |
| `inspect <source|path>` | Facts (codec, size, fps, VFR, HDR, rotation, audio offset, loudness) + warnings with fixes |
| `transcribe [source]` | Word-level transcript, suspect flags; `--passes 2` diff |
| `transcript` | Query: `--phrases`, `--search TEXT`, `--range`, `--suspects`, `--retakes`, `--output` (through the timeline) |
| `look <time|range>` | Frames, contact sheets, `--cuts`, `--overlays`, `--crop` |
| `listen <range>` | Spectrogram + RMS envelope PNG, isolated-clip transcript, loudness numbers |
| `edit <op>` | `keep`, `cut`, `tighten`, `move`, `split`, `zoom`, `pin`, `undo`, `history`; all accept text anchors |
| `timeline` | Compact view (id, output time, duration, text) and `validate` |
| `overlay <op>` | `add`, `list`, `rm`; text/image/clip/captions |
| `sync` | Align a separate recorder: offset + drift, corrected track |
| `render` | Chunked, cached render; `--preview`, `--range`, `--dry-run` |
| `check` | Verify a render: duration, transcript vs timeline text, A/V drift, frame match, tail pads, loudness, safe zones |
| `captions` | SRT/VTT of the output |
| `export` | `--otio`, `--edl` (CMX 3600) |
| `short` | Derive a vertical timeline between two phrases with short-form defaults |

## Providers

Transcription: whisper.cpp (default), others behind an interface. Denoise: optional DeepFilterNet.
Hardware decode/tone-map: VideoToolbox on macOS. All shell out; cutroom is one static Go binary.
