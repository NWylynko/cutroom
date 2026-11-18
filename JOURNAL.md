# JOURNAL

One entry per loop iteration, newest first. Format: `### N · YYYY-MM-DD · title`, then Did / Measured / Surprises / Next.

### 1 · 2026-11-18 · M0 skeleton: CLI core, guide, doctor, fixtures, tasks
Did: Go module; `internal/cli` (registry, long-flag parser, did-you-mean for commands/flags/topics,
text vs `--json` envelope `{ok,result,next}` / `{ok:false,error:{code,message,fix}}`, exit codes);
`internal/timecode` (exact rationals, frame/sample snapping, `12.5 | 1:02.5 | f372`, TestL01_*);
`guide` with embedded topics (start, pitfalls); `doctor` (ffmpeg filters, whisper-cli, model in
~/.cache/cutroom/models or $CUTROOM_MODELS / $CUTROOM_WHISPER_MODEL); fixture generator (talk, talk-vfr,
talk-rot, recorder + ground-truth manifests); `make check`, `make e2e` (4 passing); 4 T2 task briefs.
Measured: fixtures build in 15 s. talk.mp4 30.233 s, 907 frames. VFR variant 885 frames.
Surprises: `-itsoffset 0.085` on an AAC stream shows start_time 0.0637 — the 1024-sample priming
delay is subtracted. inspect must reason about the first *audible* sample, not just start_time.
Next: grader, then M1 (init/inspect).

### 0 · 2026-11-17 · charter
Did: wrote docs/LESSONS.md (56 lessons from four real productions, categorised), docs/SPEC.md,
LOOP.md, BACKLOG.md. Installed Go 1.27.4 and whisper.cpp (brew); downloaded `ggml-base.en.bin`.
Measured: whisper-cli on `say` speech gives word-level JSON (`-ojf -ml 1 -sow`), and it kept a planted
repeat ("the first thing, the first thing"). ffmpeg 9.0.1 here has no drawtext/subtitles/libass/zscale;
has scale_vt, scdet, ebur128, showspectrumpic, tile, xstack, libx264, videotoolbox.
Next: M0 skeleton.
