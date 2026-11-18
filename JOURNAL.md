# JOURNAL

One entry per loop iteration, newest first. Format: `### N · YYYY-MM-DD · title`, then Did / Measured / Surprises / Next.

### 3 · 2026-11-18 · M1 init + inspect
Did: `internal/media` (ffprobe facts; VFR from packet timestamps — no decode; rotation from the
display matrix; HDR from transfer; AV offset; EBU R128 loudness + true peak), warnings tagged
`handled` vs `ACTION` with lesson ids; `internal/project` (cutroom.json, ids from file names, staleness
by size+mtime, source identity by `os.SameFile`); `init` (idempotent, adds/re-inspects) and `inspect`
(ids, loose files, did-you-mean on ids); guide topic `workflow` (only lists what exists; future steps
marked not available). Tests: TestL02_DetectVFRAndAudioOffset, TestL25_DetectRotation, clean/audio-only,
project path identity. e2e 14/14.
Measured: VFR scan via packets 37 ms on 885 frames. Fixture loudness: camera -20.4 LUFS, recorder -16.8.
Surprises: (1) re-init added a duplicate source: stored relative path vs absolute compare with
`--project .`; then macOS /var → /private/var symlink broke string compare again. Identity is now
`os.SameFile`. (2) AAC stream copy keeps the priming frame: the stream starts 0.0637 s while speech
content is at 0.085 s. Timestamp-based alignment places content correctly, so inspect reports the
stream timestamp; the renderer must align audio by pts (not by "first sample = 0").
Next: `transcribe` + `transcript` views.

### 2 · 2026-11-18 · independent grader + its self-test
Did: `evals/grade` (shares no code with cutroom): own whisper pass with fuzzy in-order phrase
counting; duration; A/V lengths; frame matching of output against the reference testsrc video (unique
frame counters → exact source frame, upright vs flipped); A/V sync by cross-correlating 0.5 s output
audio windows with the reference at the matched source time (speech windows only, corr ≥ 0.6);
`no_rotation_meta`; `size=`; `audio_from=` (4 kHz+ speech energy compared against both candidate
sources). `evals/grade/selftest.sh`: 9 cases (good cut passes; unedited, late audio, flipped, rotation
metadata, wrong audio, unsynced recorder fail on the right check; corrected recorder passes).
Fixture change: camera audio now band-limited (180–3500 Hz) so recorder vs camera is measurable.
Exact T2 runner protocol written into LOOP.md. `make grade-selftest`.
Measured: self-test 35 s. Good cut: 74/74 frames matched (err 0.1), sync median 0, worst 34–40 ms —
the reference cut uses float-second trims, i.e. lesson L01 visible in the grader itself. cutroom's
frame-exact renders should sit well under that; if a cutroom render shows > 20 ms worst, investigate.
Surprises: zsh treats `$a:end` as `$a` + the `:e` modifier — write eval shell in bash only.
`say` renders at 22.05 kHz, so little energy above 5 kHz; measure the recorder/camera difference
relative to the sources instead of with absolute thresholds.
Next: M1 `init` + `inspect`.

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
