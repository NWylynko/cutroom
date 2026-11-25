# JOURNAL

One entry per loop iteration, newest first. Format: `### N · YYYY-MM-DD · title`, then Did / Measured / Surprises / Next.

### 30 · 2026-11-24 · M4 `captions`: SRT/VTT of the edit, with reading and seam checks (L48, L49)
Did: guide topic `captions` written first (example output, rules, checks), then `internal/captions`:
OutputWords (the words the timeline keeps, in output time, clamped to segments; um/uh dropped with
their punctuation moved to the word before; first word after a real cut marked — zoom splits aren't
cuts); Build (≤ 2 × 42 chars, ≤ 7 s; a sentence end or a ≥ 0.7 s pause starts a new cue; a forced
break moves back to a comma in the cue's second half, never strands a 1–2 word sentence tail, never
ends on a function word or a subject pronoun after a conjunction; lines balanced, preferring
punctuation; cue = first word … last word + 0.3 s, ≥ 1 s when there's room, 40 ms gaps); SRT/VTT
Format/Parse; Check: reading (chars/line, lines, chars/s), timing (order, overlap, inside the video),
words (LCS vs the edit; ≥ 10% missing or ≥ 3 extra = FAIL: a stale file), seams (L49: a sentence that
stops at a cut, a sentence whose start was cut, a cue opening on a stranded tail). `cutroom captions`
[--vtt] [--check file] [--dry-run]; a sidecar records the timeline hash so `status` says current/STALE.
Fixed on the way: `status`/`check` treated any `*.cutroom.json` in out/ as a render manifest (the
captions sidecar showed up as a render); now only .mp4 ones.
Tests: TestL48_CuesReadLikeAPersonMadeThem, TestL49_NoStrandedSentenceTail, TestL49_SeamsAtCuts,
SRT/VTT round trip + stale file, OutputWords fillers/cuts; e2e 147/147 (+7).
Measured: fixture full edit: 5 cues, all 40 words, 0 warnings; with "an old laptop." and "Finally,"
cut, both seams are flagged at the right output times. Real speech (T3 deploy 44 s, sec 40 s): all
153/123 words, 0 seam warnings, 1–2 reading-speed warnings on genuinely fast speech (20–22 chars/s).
Surprises: the first real-speech pass broke "personal | software", ended a line on "if you" and a
cue on "how". Preferring a comma in the cue's second half for forced breaks fixed most; subject
pronouns are weak only after a conjunction or punctuation (object "you" at a line end reads fine).
Next: text rasteriser + `overlay add` (M4), then burned-in captions.

### 29 · 2026-11-24 · r6 friction: exact `render --range`, multi-source `transcript`, sub-help
Did: (1) `render --range a-b` renders exactly that output range to whole frames (Timeline.Trim: clip
segments, output restarts at 0), into out/<name>-range.mp4; loudness is still measured on the whole
edit (L17). It used to render the whole ~45 s chunk(s) covering the range — on a short edit, the whole
thing, which the agent read as the flag being ignored. The manifest marks segment edges that are range
edges (`range_in/range_out`); `check` skips clipped-word warnings there and doesn't count word
differences within 0.5 s of them. (2) `transcript` with no source shows the edit's main video (else
the longest), with "also transcribed: pickup — `cutroom transcript pickup`"; before, the alphabetically
first source (a 2.6 s pickup) looked like the whole answer. (3) `cutroom edit <sub> --help` prints that
subcommand's usage and flags only (Command.Subs); guide edit mentions `edit reset` and sub-help.
Every usage line in the sub-help was run through the CLI before keeping it.
Tests: TestL01_TrimIsFrameExact; e2e 140/140 (range dry run, range render 2.000 s / 60 frames, check
PASS on a range, transcript default, sub-help).
Measured: ranges 0-3, 2-4, 3.5-8.2, 5-11, 17-25 (past end) on the tightened fixture: all frame-exact,
all `check` PASS with 0 warnings; 2–4 renders in 0.4 s.
Surprises: the first range check flagged "show" as clipped and "need" as misheard — the recogniser
at a file's very edge is unreliable, so word differences there can't be blamed on the edit.
A words-check `at` before the segment start printed "output -0.09 s is in segment " (empty id); the
expected-word time is now clamped to the segment start.
Next: M4 captions.

### 28 · 2026-11-24 · eval round 6 recorded: sonnet 8/8, haiku 8/8
Did: graded the 16 kept round-6 runs with the independent grader (they had friction reports but no
grades), wrote `evals/results/2026-11-24-6.md`, moved the friction into the backlog.
Measured: all PASS, 0 cutroom errors, no medium friction; fourth clean sonnet round (3–6). pickup:
≈1.8 s of pickup.mp4 frames in both outputs, the line said once.
Surprises: the loop had stopped between writing the friction reports and grading — iteration 27's
commit included the reports but the round was never recorded. Grading after the fact works because
run dirs are kept until recorded (LOOP step 4).
Next: r6 friction (`render --range`, multi-source `transcript`, subcommand help), then M4 captions.

### 27 · 2026-11-24 · retakes: one warning per retake, cut at sentence boundaries
Did: repeat suspects whose attempts interleave (each first take before the other's last take) are one
retake: one `retake-kept` warning ("… and N more phrase(s) are said again later"), one suggested cut.
The cut runs from the start of the sentence holding the first attempt's earliest repeated phrase to the
start of the sentence where the next attempt begins — capped at the end of the first attempt's last
repeated sentence, so words only the next attempt has are never cut. Falls back to the phrase cut when
no clean boundary exists (and the fixture's retake, already sentence-aligned, is unchanged). Test:
TestL36_RetakeCutAtSentences (the sec pattern). e2e 132/132.
Measured (real footage, each suggestion dry-run through the CLI): sec take 1: 7 → 1 warning; cuts
2.17–14.70 "Instead of making sure … layer of security." → join "…write secure code. | My solution is
to…" (the editor's cut). deploy: 5 → 2 warnings; 0.73–7.47 (the published edit cut 0.12–8.61, the same
attempt) and 21.57–31.47.
Surprises: (1) the first grouping cut into "My solution is to bare-hug the software," — the next
attempt's sentence started with new words; capping at the first attempt's sentence end fixed it.
(2) an off-by-one (`j <= firstLast`) silently sent the perfect case to the fallback.
Next: eval round 6 is due (iteration 26 was 5 after round 5); then M4 captions.

### 26 · 2026-11-24 · T3 multi-take on real footage: two HEVC HDR takes, two recorders
Did: excerpts (local) of securing-personal-software take-1/take-2 (4K HLG, 180° rotation) and their
recorders. `sync` failed on both ("only 2 confident windows"): the camera across the room correlates
0.35–0.55 with the lapel recorder even at the right lag, under the 0.5 window threshold. Lowered to
0.35, 60 windows instead of 40, and a consensus guard (≥ 60% of correlating windows must sit on the
fitted line, else no-match). TestL05_ReverberantCamera: the fixture camera with echoes, band limits
and room noise must fail at 0.5 and land within 2.5 ms at 0.35. Then the full flow: tighten take 1,
`edit audio mic1`, `edit audio mic2 --for sec-take2`, `edit add` a line from take 2, render --preview,
check.
Measured: mic1 +9.203 s (truth +9.205), mic2 +5.409 s (truth +5.410); drift on a 60 s span is loose
(mic2 −68 ppm vs −28: ≤ 2.4 ms over the excerpt). Preview of 39.9 s from two 4K HEVC HDR takes in
11.6 s. check PASS: picture worst 3.2, tone 16 vs 17, sync 0 ms against mic1/mic2 per segment,
sync-camera worst 15 ms (independent of both offsets), 116/123 words with 5 recogniser variants.
Contact sheet across the take join: upright, natural colour, take 2's slightly different framing.
Surprises: the retake suggestion on real speech cuts from mid-sentence ("does the right
authentication…") to the same words in the next take — a valid splice, but an editor would cut the
whole first attempt at sentence boundaries. Backlog.
Next: retake cuts at sentence boundaries; then M3/M4 (captions, overlays).

### 25 · 2026-11-24 · multi-take: `edit add`, `edit move`, each take's own sound
Did: `edit add` puts a span of any source into the edit (by words with --text/--from/--to across all
transcripts, or a range with --source; --after/--before id, --after start; default end; edges pinned;
warns when it duplicates kept material). `edit move <id> --after/--before`. Timeline `audio_for` maps a
video source to its recorder (`edit audio mic2 --for take2`). `render.AudioFor` resolves each take's
sound — mapped recorder; else the timeline audio when it is that take or synced to it; else, for
another take, its own sound — and still refuses a recorder synced to nothing (L05). Render assembles
audio from several files (mono when mixed); the manifest records each segment's audio source; check
measures sync, camera cross-check, edges and clipped words per segment against the right sound.
`timeline` warns `audio-mixed` where takes use different microphones. The first edit now starts from
the longest video (it was the alphabetically first: a 2.5 s pickup became the whole timeline) and says
so. Guide: "Several takes" recipe. Tests: TestAudioForMultiTake; e2e 132/132 (+9) incl. a negated
second take placed by PSNR (25.8 dB vs take2, 3.7 dB vs talk).
Surprises: (1) the default-video bug only shows with two videos — every fixture so far had one.
(2) a hue-rotated second take was invisible to every luma-based check (check's picture and the
grader's), so the e2e take is negated.
Next: real multi-take footage (securing-personal-software: two HEVC takes + two recorders) as T3.

### 24 · 2026-11-24 · `edit zoom`: punch-ins rendered, looked at and checked through the crop
Did: `cutroom edit zoom <factor> <span> [--anchor x,y]` (span = range, --text or --from/--to; 1 =
zoom out). Timeline.Zoom splits at the span edges and pins them (continuous speech: tighten must not
open a gap there). Render crops around the anchor, clamped inside the frame, before scaling; on the
HDR path `scale_vt` keeps z× the pixels for the crop. Edit warns when the upscale (zoom × out ÷ source
height) exceeds 1.2× (L40) and suggests the largest clean factor. `timeline` tags `[zoom 1.50× at
0.50,0.50]`; joins carry `kind` ("cut"/"zoom"), `--joins` prints zoom changes as "no cut"; `look
--output`/`--cuts` show the cropped picture (no cut bar at a zoom change). `check` compares zoomed
segments through the same crop (picture and tone; it skipped them before) and a frame that matches
nothing nearby now reads "wrong picture, framing or zoom, not timing" instead of "N frames off".
Manifest stores the anchor. Tests: TestL40_ZoomSplitsPinsAndStaysContinuous (incl. tighten after
zoom), TestL40_ZoomCropInChunk; e2e 124/124 (+10).
Measured: fixture zoom 1.5×: picture worst difference 26.1 (upscale softness; limit 60); manifest
stripped of the zoom → picture FAIL difference 4258.9 "framing or zoom". HDR fixture 1.2×: picture
26.9, tone 92 vs 92 once the tone sample also goes through the crop (92 vs 88 without).
Surprises: the old picture failure text would have sent an agent hunting a timing bug for a framing
problem.
Next: `edit move` / multi-take segments (M2), then the render side of multi-source.

### 23 · 2026-11-24 · breaths aren't suspects
Did: `speech.IsBreath` — a wordless sound shorter than 0.8 s and more than 12 dB below speech level is
a breath or mouth noise. Transcripts label it `breath` (Algorithm 3), so retake/kept-sound warnings,
tighten's "also removes the sound" notes and `transcript --suspects` skip it ("N breath(s) not listed;
--all shows them"); `check` stray-sound counts kept breaths as normal instead of warning. Fixture talk
gains a soft inhale before "Then install Linux on it." (−35 dB, 19 dB under speech; timings
unchanged). Tests: TestL28_BreathIsNotASoundSuspect (real + fixture levels); e2e 116/116 (+2);
selftest 12/12 on the regenerated fixtures.
Measured: deploy excerpt (speech −24 dB): 7 "sound" suspects → 1 sound (−31 dB, right after a word)
+ 6 breaths (−39…−43 dB). Fixture: breath −35 dB → breath; planted bump −21 dB → sound.
Surprises: the threshold falls in a wide gap on both real and synthetic audio (breaths 15–19 dB under
speech, bumps/coughs 5–7 dB under), so 12 dB isn't a fragile choice.
Next: M2 — `edit zoom` (punch-ins).

### 22 · 2026-11-24 · round-5 friction: HDR measured in check, one-line open warnings, clocks labelled
Did: `check` has a **tone** item for HDR sources: the render's saturation and brightness range vs the
source read through the tone map at the same moments (±15%); `render` says "is HDR (HLG): tone-mapped
to SDR bt709 (videotoolbox)". Edit commands print only warnings the edit introduced in full; ones
already open collapse to "still open (unchanged by this edit): retake-kept at output 0:04.093 — …".
Warnings say "at output …"; the retake fix says "--near is a source time". `status`: open warnings
put `cutroom timeline (N open warning(s)…)` first; timeline line names `cutroom edit history`.
`check` always suggests `cutroom transcript --file <render>` to read what it says. `-o` hint shows the
usage line. "1 segment". e2e 114/114 (+5).
Measured: HDR fixture render: tone ok (saturation 87 vs 87, range 217 vs 217); a plain bt2020→bt709
conversion with the same manifest: tone FAIL (saturation 103 vs 87).
Surprises: phone-haiku reported the retake warning as stale after it was fixed; reproducing showed it
was gone ("no problems found") — the agent saw it repeated after its other cuts. The repetition was the
real problem.
Next: breaths (real footage), then M2.

### 21 · 2026-11-24 · eval round 5 (+ hdr-phone, resume tasks)
Did: two new tasks; grader learned HDR: `sdr_bt709`, `tone_mapped` (signalstats saturation + luma
range vs a VideoToolbox-tone-mapped reference, ±10%), reference decoded from the HDR fixture itself
(`ref`, `ref_sdr`), per-task `frame_err` and `sync_limit_ms`, and frame matching by short runs instead
of single frames. Selftest 12/12 (+3: HDR good PASS, naive FAIL on tone_mapped, 100 ms late FAIL on
av_sync). `evals/setup-runs.sh` prepares run dirs. Ran 7 tasks × sonnet/haiku: **14/14 PASS,
tool-only, no medium friction**; third clean sonnet round (golden criterion 3 holds).
Measured: VideoToolbox decode vs software decode of the HDR fixture: identical frames at identical
indices (8/8). The grader's −100 ms on the first HDR attempt was its own matching (tone-shifted
test-pattern neighbours differ by 0.4 MSE vs ~149 of encoding noise), not the render.
Surprises: an independent grader for HDR is hard on synthetic content: the picture can only be placed
to ±1 frame, so hdr-phone's sync limit is 70 ms. Backlog: a per-frame marker in the fixture.

### 20 · 2026-11-24 · HDR → SDR: iPhone HLG/PQ tone-mapped in render, look and check
Did: `media.ReadSDR` — the one way render, look and check's picture burst read a source. HDR (HLG/PQ)
goes through VideoToolbox (`-hwaccel videotoolbox`, `scale_vt` to bt709 — scaled on the GPU to the
output size — `hwdownload`, `setparams` retag), the recipe the real productions used; zscale+tonemap
(hable) where VideoToolbox is missing; SDR sources untouched. Rotation on the hardware path is applied
by hand after download. `inspect`'s hdr warning is now "handled" when a tone mapper exists, else says
what ffmpeg is needed. New fixture talk-hdr.mov (HEVC 10-bit, HLG/bt2020 tags, 180° display
rotation, like iPhone files). Render Version 2 (cache keys). Tests: TestL23_HardwareToneMapOverrides
Rotation, TestL24_OutputAlwaysTaggedBT709 (L24 had no guard). e2e 109/109 (+5), selftest 9/9.
Measured: fixture: render 29.7 s in 2.9 s; check PASS, picture worst 0.3, rotation 0°, bt709; look
upright (PSNR 24.2 vs 1.2 against the flipped frame). Real iPhone 4K HLG excerpt: `look` frame shows
natural skin, saturated neon, proper contrast; a naive ffmpeg conversion of the same frame is visibly
washed out and flat (both viewed).
Surprises: (1) `-noautorotate` with hardware frames copied the input's −180° display matrix into the
output — players would have flipped it back upside down. `-display_rotation 0` on the input overrides
it. Only caught because frame counting choked on the extra side-data CSV field. (2) facts stored at
inspect time keep an old "ACTION hdr" line until the file changes — warning text that depends on the
machine shouldn't be frozen into cutroom.json (backlog).
Next: breaths (real footage), then M2 (`edit zoom`, multi-take).

### 19 · 2026-11-24 · T3 prep on real footage: one retake = one suspect; changed sources can't be used stale
Did: 60 s stream-copy excerpts (local only) from three published productions: ProRes 1080p single take,
h264 1080p with two video streams, iPhone 4K HLG + separate 24-bit recorder. `init`/`transcribe`/
`sync` on each. Fixes from what they exposed: (1) a retake longer than 8 words was reported once per
overlapping n-gram window (4 suspects for one retake) — windows whose takes overlap pairwise are now
merged (TestL36_LongRetakeReportedOnce; transcript Algorithm 2 invalidates old caches). (2) Bare
`cutroom init` didn't re-inspect a file that changed on disk although the guide and every
stale-source fix line promise it; now it does, says "nothing changed" otherwise, and warns that a
changed recorder's alignment was reset. (3) `sync` and `transcribe` used caches keyed on the facts
recorded at inspect time, so a replaced file was silently aligned/transcribed as the old one; both
now refuse a changed source with "run `cutroom init`", and sync's wav cache compares with the file on
disk. e2e 104/104 (+6).
Measured (real footage): deploy excerpt suspects 7 → 5 repeats; both retake spans match what the
published edit removed (excerpt 0.12–8.61 s and 14.86–41.52 s). iPhone+recorder: `sync` +15.167 s,
−23.8 ppm, 21 windows, residual 0.04 ms, 0.4 s — the production's own tool measured +15.17 s (from
−104.8265 s at cam 0) and −24.36 ppm. Transcribe ~4 s per minute of speech (base.en, 2 passes).
Surprises: `status` caught the stale recorder before I did. The no-match error on my first (wrong)
audio window was correct and its fix line pointed the right way.
Next: HDR→SDR tone mapping moves up — 2 of 4 real productions are iPhone HLG and render "washed out"
today [L23]. Then breath/room-tone thresholds (deploy: 6 "sound" suspects at −39…−43 dB peaks over
a −55 dB floor are breaths).

### 18 · 2026-11-24 · `status`
Did: `cutroom status` (SPEC rule 6): sources (kind, size, length, transcribed, changed on disk, recorder
sync), timeline (segments, length vs sources, sound source, edits and the last one, error/warning
codes), every render in out/ with a manifest (full/preview/partial, current vs STALE by timeline hash,
last check verdict while the file is unchanged), and one next step: transcribe → first edit →
render --preview → check → render → check → "done: out/final.mp4 is current and passed check".
`check` now records its verdict in .cutroom/checks.json (pass, counts, first problem, file size+mtime).
Guide workflow points to it. e2e 98/98 (+6: each next-step state, recorder sync shown).
Surprises: with two current full renders the newest won; out/final.mp4 is the deliverable, so it wins.
Next: T3 prep on real footage (excerpts, inspect/transcribe vs the published cut).

### 17 · 2026-11-24 · round-4 friction: render paths, honest fix lines, independent recorder sync check
Did: `transcript out/final.mp4` reads the file (as `--file`); any command given a render path where a
source is expected says "is a render, not a source; read it as a file" instead of suggesting `init`.
`check <missing>` lists the renders that exist. `render --output PATH` works (agents guessed it three
rounds running; `-o` now says "did you mean --output"); dry-run names the output path and `next:`
keeps --preview/--range/path. Render says when a source's display rotation was baked into the pixels.
`retake-kept` names what its command removes ("removes the earlier take, source 0:06.314–0:09.214
\"The first thing you need is, UM, …\""). New check **sync-camera**: with a separate recorder, the
render's sound vs the camera's own sound (both band-passed), independent of `sync`'s offset. Guide:
render output path, sync-camera row. e2e 92/92 (+6), grader selftest 9/9.
Measured: recorder fixture, correct offset: sync-camera 7/7 segments, worst 0 ms. Offset planted
+100 ms in cutroom.json then rendered: `sync` still PASS (worst 0 ms — it trusts the same offset),
sync-camera FAIL worst −100 ms with "re-run `cutroom sync mic`".
Surprises: the recorder-haiku agent found a real verification gap by reasoning about what `check`
compares, not by hitting an error. vfr-sonnet reported the retake hint as wrong; dry-running it shows
it was right — the hint just didn't say what it would remove, which is the actual fix.
Next: golden criterion 3 is met (sonnet clean in rounds 3 and 4). Move to spec coverage: `status`,
then M2 (`edit zoom`, `edit move`, multi-take render), and T3 prep on real footage.

### 16 · 2026-11-24 · eval round 4
Did: decisions recorded (name cutroom, Apache-2.0 LICENSE, sonnet gates T2, full v1 target; real
footage for T3 indexed in evals/local, gitignored). Ran 5 tasks × sonnet/haiku at 3e8fe0e: **10/10
grader passes, all tool-only, no medium friction**. Sonnet median 7 tool calls (round 3: 9), haiku 10
(13). Five cutroom errors in total, each unstuck. Results: evals/results/2026-11-24-4.md.
Surprises: haiku agents now re-read the render as text by habit and hit the one wrong fix line left.
Next: fix the seven low items.

### 15 · 2026-11-18 · round-3 friction: runnable retake fix, tighten/cut say what they removed
Did: `retake-kept` now prints a ready `cutroom edit cut --from "<take>" --near <t> --to "<last words>"`
that removes the earlier take(s) and keeps the last; the --to phrase is grown until the CLI's own
resolution (first kept match after --from) lands on the last word before the next take. Tighten prints
its rule (pads + --min-cut) and each pause as `0:04.180 2.13→0.32 s`; JSON has `shortened`. Any edit
that drops a "sound" suspect says so ("this also removes the sound with no words at …"). Unknown flags
on commands with positional args show the usage line (`render --output` → `cutroom render
[output.mp4]`). `look`/`listen` take a file path first (`look out/final.mp4 3.9`). `render --range`
says which output span its chunks cover, and "= the whole timeline" when it fits in one chunk.
`check` sync line/help say what it proves: the render's sound vs the source's sound at the moment the
picture shows, so with picture ok lip sync is as the source plays. Guide: retake warning after tighten
is expected and carries the fix. e2e 86/86 (+8), incl. running the warning's command and asserting it
removes exactly the false start.
Surprises: the first draft of the --to phrase ("first thing.") would have been the wrong match if the
false start had contained the phrase twice; checking against the same Search the CLI uses keeps the
suggestion honest instead of plausible.
Next: eval round 4 (second clean round needed for golden criterion 3).

### 14 · 2026-11-18 · eval round 3
Did: 5 tasks × haiku/sonnet at f7f96a1. **10/10 grader passes, all tool-only**, no medium friction:
the round-2 VFR false FAIL is gone (both vfr runs: check PASS, 0 warnings). Median 12.5 tool calls
(round 2: 16); 4 cutroom errors in total, all guessed syntax (`render --output` ×2, `look <file>`, cut
of an already-removed sound), each unstuck by its fix line. Results: evals/results/2026-11-18-3.md.
Six low friction items, top one from 5 runs: the retake warning should carry a runnable cut.
Surprises: agents now lean entirely on `check`; none listened to a join. The vfr-haiku agent doubted
that a passing sync check meant lip sync — wording, not measurement.
Next: fix the six items, then round 4.

### 13 · 2026-11-18 · round-2 friction: VFR picture false FAIL, transcript --file, 8 small fixes
Did: `check` picture now judged by real timestamps: source bursts decoded with `-copyts` + `showinfo`;
the expected frame is the one whose display interval contains the target time; pass if it matches as
well as the best match (held VFR frames, static shots), else report the best match's offset in frames
and ms. Guard: a picture check that measured nothing now FAILS instead of passing. `transcript --file`
(any file, cached; all views). Short flags: `-o` → "positional arguments need no flag"; `-p` → "did you
mean --preview". Dry run says "would create timeline". New validator `broken-sentence` (words removed
from inside a sentence, retakes exempt). Resolved repeats marked in `--suspects`. doctor's next: line
matches the guide; `sync` says "mic started recording 1.700 s before talk"; `check` names the audio
reference; `look` notes nominal frame numbers on VFR; guide notes check's exit 4 in chains. e2e 78/78.
Surprises: (1) the first rewrite PASSED the VFR render by measuring nothing — showinfo logs one frame
more than `-frames:v` emits, so the length check discarded every burst and "0 sampled frames match"
read as success. Never let a check pass on zero measurements. (2) Picking "the nearest of the
equally-good frames" turned a planted 3-frame shift into "+2": on test patterns neighbouring frames
differ by < 1.0 MSE. Judge the *expected* frame against the best one instead.
Next: eval round 3.

### 12 · 2026-11-18 · eval round 2
Did: new task `fix-edit` (scripted `setup` field: a pre-broken project; grader fails its start state 3
ways). Ran 5 tasks × haiku/sonnet. **10/10 tool-only passes** (round 1: 6/8). Results:
evals/results/2026-11-18-2.md, reports in evals/results/r2/. Both models solved fix-edit with `edit
history`/`undo` and the validator's warnings. One medium friction: `check` reported a false picture
FAIL on the VFR source in both vfr runs (offset counted in decoded frames across a dropped-frame gap);
the haiku agent spent ~34 calls investigating. Plus 9 small items. Not golden: needs that fixed, then
two clean rounds.
Next: fix the round-2 list (VFR picture check first, `transcript --file`), then round 3.

### 11 · 2026-11-18 · perceive the edit and the render (friction #3 from round 1)
Did: contact sheets generalised to labelled cells (any file/time, cut markers). `look --cuts`
(pairs: last frame before | first after each cut, words either side, measured pause; 12 cuts/page),
`look --output <t|range>` (output time mapped through the timeline to source frames; red bar where a
cut falls), `look --file <video>`; `listen --output <range>` (needs a render of the *current* timeline,
else a clear stale/no-render error) and `listen --file`, both with words mapped through the render
manifest and cuts drawn on the spectrogram. Guide `perceive` updated. e2e 70/70.
Looked at all three new images: cut pairs show the expected source jumps; the output sheet marks the
cut between output 3.857 and 4.143; the render spectrogram shows the cut centred in a clean 0.30 s pause.
Next: eval round 2.

### 10 · 2026-11-18 · sync + audio source (round-1 correctness bug) + 5 small friction fixes
Did: `internal/align`: 10 ms log-envelope FFT cross-correlation (coarse, whole files), then 40 × 2 s
waveform windows band-passed to 300–3000 Hz, xcorr ±60 ms, robust line fit (3×MAD outlier rejection,
silent windows skipped — L05) → offset + drift + residual + overlap. `sync <src> [--to cam]` stores it
in cutroom.json. `edit audio <src>` (refuses unsynced), `timeline` shows the sound source; validation
ERROR `audio-not-synced` (render refuses with the exact fix); renderer maps recorder audio through the
fit in ≤ 10 s pieces (≤ ½ sample drift error per piece, no resampling); manifest + `check` measure
sync through the same mapping. Friction: dry runs say "would …" + DRY RUN line; out-of-range
`--from/--to` explains they are seconds; suspects already cut are marked; whisper `[BLANK_AUDIO]` /
`(music)` annotations stripped (the "[" was being glued to the previous word as punctuation);
`init` names cutroom.json; `init --remove` (refuses sources in use). Guide: recorder recipe.
Tests: TestL05_RecorderFixture, FFT round trip, annotation stripping. e2e 66/66 incl. the round-1
failure (hand-set `"audio": "mic"`) now refused, and the grader passing the recorder task done
with cutroom alone.
Measured: recorder fixture: offset 1.700055 s (truth 1.700), drift −42.7 ppm (truth −50; 30 s is too
short to pin drift — mapping error < 1 ms either way), residual 0.02 ms, 37/38 windows. Before
band-passing: 21 of 38 windows rejected and residual 0.21 ms (camera mic band-limited vs full-band
recorder — real pairs differ the same way). Render + check: sync worst 6 ms; grader: audio from the
recorder, sync worst 6 ms.
Next: perceive the render (`look`/`listen --output`), then eval round 2.

### 9 · 2026-11-18 · `check` (friction #1 from round 1)
Did: render writes `<out>.cutroom.json` (segments with file-relative output times, sources, settings,
timeline hash). `internal/verify` measures the file: frames vs manifest, audio vs video length, size /
rotation / colour tags, picture (render frames vs a 7-frame source burst, offset in frames, static-shot
ambiguity tolerated), per-segment sync (audio xcorr vs the source at the mapped time), words (re-
transcribe, LCS vs the timeline's kept words; ≥3 missing in a row = FAIL), clipped words at edges,
stray sound without words (L33), loudness, stale timeline. `check` CLI (exit 4 on fail, each finding
with output time, segment and a command to look). Guide `verify`; guide rule 1 restored; render's
next: is now `cutroom check <file>`. Also: CLI no longer panics on an undeclared flag, and any panic
becomes `error[internal]` with a fix line. e2e 55/55 incl. five planted faults.
Measured: good render: PASS in 1.7 s (413/413 frames, 15 picture samples err 0.0, sync 0 ms on 5/5
segments, 44/44 words, -16.0 LUFS). Planted: audio +100 ms → FAIL sync +100; picture +3 frames → FAIL
picture offset +3; cut through a sentence → WARN clipped-word "I'm"; burst kept → WARN stray-sound
at 16.97; edited after render → WARN stale.
Surprises: (1) the first sync measurement of the late file read −21 ms — extracting audio to WAV
drops the stream's start offset (third time this trap has bitten; now padded back from start_time).
(2) The sign of the lag was inverted; a unit test now pins "late = +". (3) Planting the burst by
undoing a cut didn't keep the burst (tighten had removed it) — the first fault test was wrong, not
the checker; verify the fault is really planted before trusting a "missed" result.
Next: audio from another source (correctness) + `sync`.

### 8 · 2026-11-18 · eval round 1
Did: first T2 round: 4 tasks × haiku/sonnet, fresh subagents with only the brief + `cutroom guide`.
Results in evals/results/2026-11-18-1.md (friction reports in evals/results/r1/). Grader passes 8/8 after
loosening phone-footage's min duration (both agents also cut a retake, prompted by `retake-kept`), but
**tool-only 6/8**: both recorder runs needed workarounds outside cutroom (own xcorr, ffmpeg-built audio,
hand-edited timeline.json). Every cutroom render measured frame-exact; sync ≤ 30 ms.
Top friction: (1) `check` mandated by the guide but missing — all 8 runs; guide contradiction fixed now
(rule 1 rewritten, render's next: line no longer points at a missing command). (2) **Correctness bug**:
timeline audio set to another source renders 1.7 s late silently. (3) Can't look/listen/transcribe the
render. Plus 5 small items. All in BACKLOG "NOW", above features.
Measured: agents took 70–150 s and 8–22 tool calls; haiku ~40% more commands than sonnet, same outcomes.
Surprises: the validator's warnings steer agents beyond the brief (retake removal). The recorder agents
found the 1.7 s offset immediately from two transcripts' phrase times — `sync` can use the same idea as a
cross-check of the xcorr result.
Next: `check`, then the audio-source correctness bug + `sync`.

### 7 · 2026-11-18 · M2 render
Did: `internal/render`: audio decoded once to a cached 48 kHz float WAV and assembled in Go by
sample index (pieces = frames × SR/fps, so audio length = video length exactly), 20 ms raised-cosine
fades (L22), voice chain, loudness measured on the whole timeline then exact gain + alimiter, AAC
stereo 48 kHz (L17/L27/L43); video planned into ~45 s chunks at cuts, each segment its own seek
0.25 s early + `fps=…:start_time=<lead>` on the audio-anchored grid + `trim=end_frame` + 3 spare
frames (L02/L09), concat + restamp (L08), bt709 tags, x264, frame count checked (L14), cached by
content key, parallel, joined by stream copy (L12) and muxed. `render` CLI: --preview (540p),
--range (output time), --dry-run with estimates from per-project timing history (L56 defaults),
--loudness, --true-peak, --audio-chain. Guide `render`. Tests: TestL07_L08_ChunkCommandInvariants,
TestL01_AssembleSampleExactWithFades, chunk planning/caching. e2e 49/49, including the independent
grader passing the e2e render.
Measured: talk edit (13.767 s): 1.2 s final render, 0.08 s cached; -16.0 LUFS exactly. Grader on
renders of all three video fixtures: frames 69–70/70 exact; sync worst 0 ms (CFR, rotated) and 30 ms
(VFR: nearest-frame rounding around dropped frames). Rotated source: upright pixels, no rotation
metadata. The grader's own float-trim reference cut drifts 34–40 ms — L01 visible side by side.
Surprises: (1) the grader failed every frame (err 216) though luma PSNR was 47.7 dB: `format=gray`
treats untagged source and bt709-tagged output differently. Grader now compares raw Y planes; the
real issue (ffmpeg assumes bt601 for untagged HD, players assume bt709) → backlog for `look`.
(2) The VFR fixture itself was out of sync: `-itsoffset` shifted audio timestamps without dropping
content, so the file played 85 ms late. Real cameras that start audio late keep timestamps honest.
Fixture fixed (trim then offset); the renderer, which honours timestamps, then passed.
Next: **first T2 eval round** (agents with only `cutroom guide`), then `check`.

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
