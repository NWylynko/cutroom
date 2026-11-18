# JOURNAL

One entry per loop iteration, newest first. Format: `### N · YYYY-MM-DD · title`, then Did / Measured / Surprises / Next.

### 6 · 2026-11-18 · M2 timeline + edit
Did: `internal/timeline`: segments of kept source ranges, edges snapped to the frame grid anchored at
the audio start (L01/L02), output duration from whole frames, source↔output mapping, Cut (pins
explicit edges), Keep, CutWords (cuts inside the pauses with pads 0.22/0.10, warns on continuous
speech), Tighten (pause − pads ≥ min-cut, trims edges, never moves pinned edges — L38/L39),
Validate (out-of-source, short segments, repeated material, edge-in-word, retake-kept, kept-sound),
Joins, KeptBetween, checkpoint/undo history. CLI `edit cut|keep|tighten|undo|history|reset` with
`--text`, `--from/--to`, `--near`, `--occurrence`, `--dry-run`, auto-validation after every edit;
`timeline` (segments with words, problems) and `timeline --joins`. Guide topic `edit`. Unit tests
TestL01_EdgesOnFrameGrid, TestL39_*, TestL38_*; e2e reaches the ground-truth edit of talk.mp4 by
words alone (13.767 s, no retake/aside words, "no problems found"). e2e 44/44.
Surprises: (1) search dropped "um" from the *query*, so quoting the filler to pick the abandoned take
was "ambiguous" — fillers are now skipped only in the transcript when the query lacks them.
(2) `--to` matched at the same position as `--from` and silently cut 3 words; caught at once by the
continuous-speech warning. --to now must start after --from starts and end at or after it ends.
(3) The join pause printed after a cut ignored what tighten had already removed (said 2.84 s, real
0.30 s); now measured from what the timeline keeps.
Next: `render` (chunked, cached, frame/sample exact), then the first T2 eval round.

### 5 · 2026-11-18 · M1 look + listen (eval round deferred)
Eval round due (5th iteration) deferred to right after `render` exists: every T2 task needs an output
video, so a round now would only measure "can't do it". Logged in BACKLOG.
Did: `internal/raster` (Go fonts via golang.org/x/image, text/labels/fills — L47, also the base for
captions later); `look` (one frame as displayed, rotation applied, labelled; `--crop` at full
resolution; ranges → contact sheet with time, frame number and the word being spoken per frame);
`listen` (≤ 30 s: ffmpeg spectrogram + our loudness trace with room-tone/threshold lines, sound
regions green / no-words red, word blocks with start times, time axis; text: transcript words, what an
isolated re-transcription hears (L30), peak/room tone, sound regions, unworded sounds). Guide
`perceive` updated ("open the PNG"). e2e 30/30 incl. a PSNR check that the rotated source looks
identical to the clean one in `look`.
Measured: listen 0.9 s for 3.6 s; 8-frame sheet ~1 s.
Surprises: `listen` hung forever: label truncation cut 2 bytes and appended "…" (3 bytes), so the
string grew. Now rune-based with a termination test. Viewing the first listen PNG found 3 readability
bugs (no-words region drawn teal instead of red, overlapping time labels, labels on the trace) — the
"look at your output" rule earned its place again.
Next: M2 timeline schema + edit commands.

### 4 · 2026-11-18 · M1 transcribe + transcript
Did: `internal/speech`: 16 kHz extraction (source time = wav time + audio stream start); 10 ms RMS
envelope after an 80 Hz high-pass (L46); sound regions (threshold = room tone + max(10 dB, 35% of
range), 80 ms gap closing, 30 ms hangover); whisper.cpp passes with **DTW word onsets** (`-nfa -dtw
<model>`: flash attention silently disables DTW, t_dtw = -1); tokens merged into words, punctuation
never extends timing; refinement: first word of a region starts at the region start, last word ends
at the region end; suspects: repeat (3–8-gram within 30 s → retake, "keep the last"), sound (region
with no words), long-word, low-confidence, pass-diff (LCS per 60 s window). Commands `transcribe`
(cached by size/mtime/model/passes/algorithm) and `transcript` (phrases with pauses + suspect markers,
`--words`, `--search` with context, `--suspects`, `--from/--to`, truncation at 40 lines). Guide topic
`perceive`. Tests incl. TestL28_FixturePhrasesMatchGroundTruth. e2e 22/22.
Measured: whisper base.en 0.8 s/pass on 30 s (1.5 s with DTW). Raw whisper word mode stretched
"server." 3.69→6.38 over a 2 s pause and placed "Then install Linux on it" ~1.5 s late (overlapping
the noise burst). After refinement every phrase start is within 50 ms of truth; burst found at
16.97–17.30 (truth 17.008–17.258); VFR fixture times = clean + 0.084 s (truth 0.085).
Surprises: the disfluency prompt changed nothing on TTS speech (no real stumbles), so pass-diff is
only unit-tested. Synthetic audio is cleaner than real rooms: threshold tuning needs T3 footage.
Next: `look` + `listen`.

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
