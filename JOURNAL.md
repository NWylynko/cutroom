# JOURNAL

One entry per loop iteration, newest first. Format: `### N · YYYY-MM-DD · title`, then Did / Measured / Surprises / Next.

### 60 · 2026-11-26 · Shorts frame the speaker by default (`edit frame --anchor auto`)
Did: `short` cropped the centre unless told otherwise; on the real footage the speaker sits at
x≈0.74 and the short showed an empty desk (every check passed). New perceive.FindSpeaker: 16 frames
at 160×90 over the short's span; pixels that look like skin (YCbCr) and move weigh most, upper rows
more than lower (face over hands/mic); the crop centre is the weighted centroid inside the best
9:16-wide window. Below 2% skin area it reports "no speaker" and stays centred. `short` uses it when
no --anchor is given (and warns when a given anchor leaves the speaker out); `edit frame --anchor
auto` re-finds them. Prototyped first on all five real takes.
Measured: estimate vs the face by eye — getting-started 0.76 (0.74), deploying 0.52 (0.52),
my-personal 0.54 (0.53), securing take-1 0.70 (0.74), take-2 0.75 (0.74); colour-bar fixture: 0.5%
skin → not found. The re-made GitHub short is framed on him (looked at 3 frames). ~1 s added to
`short`. make check green; e2e 279/279.
Next: `check` could say when a render's crop holds no one (today only `short`/`edit frame` do).

### 59 · 2026-11-26 · Named looks: `style save/use/looks`, `short --look`
Did: r12 haiku's "one change" — a channel look took ~10 style calls per short. `style save <name>`
keeps the project's caption + card style (font files made absolute) in the user config folder
($CUTROOM_LOOKS overrides; a path like looks/brand.json keeps it with the project);
`style use <name>` applies it (undoable; per-card styles stay); `style looks` lists them; `short
--look <name>` starts a short with it. Unknown names suggest the closest saved one.
Measured: saved the real short's look, made a new short from another moment of the real footage with
`--look channel`: captions (Avenir Black, yellow highlight, pop) and card (light rounded, slide-up)
as saved; render + `check` PASS. make check green; e2e 278/278.
Surprises: that new short framed the empty desk — `short` crops the centre by default and the
speaker sits at x≈0.74, so he is mostly off the right edge; every check passed. Framing is the next
real-footage problem.
Next: find the speaker and frame them by default (and let `check` notice a frame with nobody in it).

### 58 · 2026-11-26 · Eval r12 styled-short: both pass; friction fixed (burn hint, compact style, 720p shorts)
Did: new T2 task `styled-short` (a channel look the short must match) and grader check
`caption_color` (pixels of the asked colour in sampled frames; selftest: a pink-boxed short passes,
the plain short fails — yellow was useless because the fixture's colour bars contain it, so the
brief uses #FF5FA2). Fresh sonnet and haiku both PASS with 0 cutroom errors (evals/results/
2026-11-26-12.md). Fixed their friction: `style captions` without burned captions says
`captions --burn`; style changes print a one-line "now set" instead of the full table; --animate
says how to see the motion; `short` picks 720×1280 for a 720p-class source (and the soft-source tip
no longer suggests the size it already uses); `captions --burn` mentions lower thirds only when
there are some.
Measured: grader selftest 28/28; make check green; e2e 272/272.
Next: named looks (save a style set once, apply it to every short) — haiku's "one change".

### 57 · 2026-11-26 · Accent bars follow a rounded card's corners
Did: on `style cards --radius …` the lower-third accent bar was its own little pill sitting on the
panel's curve. It is now the left strip of the panel's own rounded shape (FillRoundPart), a little
wider on rounded panels so it shows inside the corner; square panels unchanged.
Measured: real short card at radius 0.5, accent #307351: the bar curves with the corner (looked at
full resolution); test asserts the corner pixel is clear and mid-height is the accent. make check
green; e2e 271/271.
Next: T2 eval task whose brief asks for a specific look (fresh sonnet + haiku).

### 56 · 2026-11-26 · Animated captions and cards: pop, slide, fade (`style --animate`)
Did: cards could only alpha-fade, captions only cut in. New style fields `animate` (none | fade | pop |
slide-up/down/left/right) and `animate-time`. Captions: each group's first frames are drawn as
scaled/moved/faded copies of its card (pre-rendered PNGs, one frame each, in the same concat list —
no new filters; karaoke animates once per group); export folds them back into the group's card.
Cards: cropped to their box (+ shadow room) and moved/scaled per frame with eased ffmpeg expressions
(overlay x/y and scale with eval=frame, frame number n), in and out; chunk splits avoid them like fades.
Measured: real short, captions pop (0.15 s), cards slide-up/pop (0.3 s): per-frame card brightness
44→52 ramps then holds (settled on frame 52 = 0.3 s); `check` PASS (overlays visible over their
windows, gone before/after; 301 boxes in the safe zones). make check green; e2e 271/271.
Surprises: two bugs found only by looking — the chunk copied its overlay before the anim flags were
set (cards rendered static), and the caption union box started from the first *animation* frame's
zero box, sending a lower third to y 0 (check's safe-zones caught it).
Next: inset the accent bar on rounded cards; a T2 task that asks for a specific look.

### 55 · 2026-11-26 · Lower thirds keep clear of where the captions really are
Did: with burned-in captions a lower third was raised above the *default* caption band; after
`style captions --y/--size` it could sit on top of the captions, or jump up for no reason. Placements
now lays out the captions first and gives text/image lower thirds their real box (Overlay.CapBox,
not saved): a lower third stays in its usual place when that doesn't touch the captions, else goes
just above them (below, when above would leave the vertical safe area). Caption card boxes are
remembered by key, so the extra layout pass doesn't redraw cached cards.
Measured: real short, card "Nick / builds personal software": captions at y 0.78 → card above them;
captions at y 0.45 → card stays at its lower-third place. make check green; e2e 269/269.
Next: rounded cards with an accent bar look slightly off (bar on the curve); entrance animations.

### 54 · 2026-11-26 · `cutroom style`: captions and cards take any look (no house style forced)
Did: every caption and card was one hard-coded look (Go Bold, white on a black box, amber bar, fixed
place, 3 words). New `timeline.Style` (all fields optional, layered defaults ← `style cards` ←
`style <id>`; captions ← `style captions`) and `cutroom style`: font (built-in, a font file, or an
installed font by face name — `style fonts` lists them; .ttc faces by name or #n), size, colour,
sub-line font/size/colour, box (any alpha, or none), rounded corners, padding, outline, letter
shadow, soft box shadow, accent bar, case, alignment, x/y centre, wrap width, words/chars at once,
and a karaoke highlight (one card per word, timed on word starts). Bad colours and unknown fonts
are refused when set (with close font names); unsafe placement on vertical frames warns, with a
style fix. Guide topic `style`; captions/overlays/shorts guides point to it.
Measured: on the real-footage short: Avenir Black upper-case, outlined, no box, yellow highlight,
2 words, y 0.62 + a light rounded card (Avenir Next Heavy/Medium, box shadow) — rendered, `check`
PASS (captions 124 groups sampled on screen, safe zones ok, sync 0 ms). make check green; e2e
269/269 (+12 style cases).
Surprises: a letter shadow on a light card looks like smudged print — card panels want a box shadow,
so `shadow` (letters) and `box-shadow` (panel) are separate fields.
Next: lower thirds raised above captions ignore the captions' own y/size; entrance animation
(pop/slide) for captions and cards; T2 task where the brief asks for a specific look.

### 53 · 2026-11-26 · Fill crop centres on its anchor (vertical shorts framed the speaker off-edge)
Did: a real-footage short (speaker at x≈0.74, `edit frame --anchor 0.70,0.40`) came out with the face
against the right edge and the mic hand cut off. FrameFit cropped at `(iw-ow)*fx`: the anchor was a slide
position, while `short`/`edit frame` help says "point of the source kept in frame" and ZoomCrop centres
on its anchor. FrameFit now centres too: `clip(fx*iw-ow/2,0,iw-ow)` (and y). The grader's verticalRef
and the selftest's known-good short use the same rule; shorts guide says the crop centres on the point.
Measured: at 0.70 on a 1920×1080 source the old window was centred on 0.64; the face sat at ~83% of
the width. Re-rendered: face centred, `check` passes. 0.5 is unchanged. make check green; e2e 257/257;
grader selftest 26/26.
Surprises: `check` can't see framing — every check passed on the badly framed short.
Next: styling — captions and cards should take the look the user wants, not one house style.

### 52 · 2026-11-25 · Golden: eval round 11 24/24, all five criteria in the same round
Did: eval round 11 — 12 tasks × sonnet + haiku, fresh agents (brief + "Start with `cutroom guide`").
Sonnet 12/12, haiku 12/12 on the independent grader; every friction item rated low (round summary
evals/results/2026-11-25-11.md, reports in evals/results/r11/). Same round: T3 r51 4/4 accepted,
make check green, e2e 257/257, grader selftest 26/26, `go install` + `doctor` on a clean PATH (exit 3
with each fix) and with tools (exit 0). Added `transcript --source` (an r11 sonnet guess). Wrote
RELEASE_NOTES.md; r11's low friction to BACKLOG.
Measured: cutroom errors in 24 runs: 4, each recovered by its fix line at once.
Surprises: the last real gaps were all in the sound, not the text — transcripts put word onsets late,
and only measuring the audio at each edge closed them.
Next: golden — loop stopped; publishing is the maintainer's call.

### 51 · 2026-11-25 · Golden 4 met: T3 r51 accepts all four real-footage edits
Did: T3 r51, fresh sonnet editors on all four excerpts, opus reviewer: deploy, started, mps, sec all
ACCEPT; `check` PASS on all four (sync 0 ms vs each recorder, ≤ 15 ms vs camera sound; −16 to
−16.8 LUFS). Minor notes only: deploy's one join could breathe a little longer; started keeps a faint
0.1 s tail of a weak "up"; mps dropped 7.5 s of on-topic content to avoid opening on "And they…" (a
content choice); sec's take change shows a framing/colour-temperature difference.
Fixes from r51 friction: `edit cut` now reports every word a cut moved to the quiet point really
removes (and its warnings name the true neighbours) — "to create" also took a 40 ms "up";
ambiguous text in several sources suggests --source; --from/--to/--text given a time say to cut a
range; the workflow guide names the output size (4K sources render 4K; `--size 1080p`).
Measured: make check green; e2e 257/257. Across r48–r51 the reviewer's rejects went 1 → 1 → 1 → 0,
each traced to a cause in cutroom rather than the editor: dangling ending (ends-mid-speech),
fragments at the camera's end and in continuous speech (mapped-edge words, quiet-frame cuts),
fragments from time-range cuts (edge-in-sound).
Next: eval round 11 (T2, sonnet gate + haiku), all criteria in the same round.

### 50 · 2026-11-25 · T3 r50: 3/4 again; edge-in-sound catches every fragment the reviewers heard
Did: T3 r50, fresh editors on all four. deploy, started, mps ACCEPT (mps: the editor moved the
earlier take that completes the camera's cut-off last sentence to the end — a real editor's move);
sec REJECT again, new fragments: 80 ms of "and" after "with a PIN", and "meaning | you" cut mid-word.
Both edges came from time-range cuts the editor computed by hand from recorder times, 20 ms "before"
words whose real onsets were 50–80 ms earlier than the transcript — inside edge-in-word's 40 ms
tolerance. Transcripts can't see this; the sound can. New check edge-in-sound (Validate, via the
Quiet hook): a segment edge (not a recording's own edge, not a split) whose level is ≥ 18 dB above
the pauses within ±0.4 s (5th percentile, digital silence ignored) is in speech; the fix trims to the
quiet frame within 150 ms nearest the edge, or says there is no pause and to cut by words. quietest()
now prefers, among frames within 1 dB of the minimum, the one nearest the original edge.
Also: the timeline header names each take's sound ("mic1 (for sec-take1), mic2 (for sec-take2)");
a recorder's transcript says `transcript <camera>` shows its words on the camera's clock; a
transcript range prints whole. Guide: prefer word cuts; edge-in-sound explained.
Measured: on the eight T3 projects of r49+r50 it flags exactly the four fragments the reviewers
rejected (r49 sec ×2, r50 sec ×2; r49's join fix = 13.367, the gap found by hand) and nothing on
the six accepted edits. `timeline` with levels: 35 ms on a 1-min excerpt. e2e 256/256.
Surprises: a first relative rule (edge vs quietest point within 150 ms) missed "meaning | you" —
there is no pause there at all; the room level is the right reference. A 20th percentile was too
high around dense speech; the 5th works.
Next: T3 r51 on all four, then eval round 11.

### 49 · 2026-11-25 · T3 r49: 3/4 accepted; sec's word fragments found and fixed at the source
Did: T3 round r49 (fresh sonnet editors, opus reviewer; packs built by evals/local/t3/pack.sh).
deploy ACCEPT (≈ the channel's own edit: 8.70–14.83 + 41.57–56.50 vs 8.61–14.86 + 41.52–56.65),
started ACCEPT (the r48 reject: the editor followed the new ends-mid-speech warning), mps ACCEPT,
sec REJECT: two word fragments — 20 ms of the next "and" at the very end, 40 ms of a cut "you" after
"access,". Causes and fixes: (1) MapTo dropped recorder words running past the camera's end, so no
check knew "and" was there → words overlapping the edge are kept, clipped; edge-in-word now gives a
runnable time-range cut (so `timeline --apply` fixes it). (2) In continuous speech the cut sat on
whisper's boundary, 50–80 ms late → Timeline.Boundary puts it at the quietest frame boundary from
120 ms before to 30 ms after, measured in the sound the edit uses (recorder via offset+drift), for
cut/keep/add/zoom edges; edge-in-word tolerates that window. (3) Both recorder editors hit recorder
words whisper couldn't place (uncertain-timing, "and" 0.9 s early) → ReconcileTiming takes such a
word's time from the camera mic's transcript when a neighbour agrees. Friction fixes: --to may equal
--from; `listen 0 3` is a range; file-not-found lists the media files that are there.
Measured: sec's replayed add now ends at 13.367 (mic 18.775, −68 dB gap) instead of 13.433 (inside
"you"); its last segment raises edge-in-word with `edit cut 59.825-59.900 --source sec-take2`.
make check green; e2e 256/256.
Surprises: a symmetric ±60 ms search missed the gap — recogniser onsets are biased late, so the
search must lean earlier.
Next: T3 r50 on all four (the cut placement changed for every edit), then eval round 11.

### 48 · 2026-11-25 · T3 round r48: 3/4 accepted; fixes for the reject and the friction
Did: T3 on the four real-footage excerpts (deploy, started, mps, sec): a sonnet editor per excerpt
(cutroom guide only) wrote EDIT.md + FRICTION.md, then a reviewer (strongest model) judged each edit
from transcripts, joins, contact sheets and `listen`. Verdicts: deploy ACCEPT, mps ACCEPT, sec ACCEPT,
started REJECT — it ended on "Those are much better.", the first words of a sentence the recording
cuts off, contradicting the line before it; its last segment also ran past the source's end
(out-of-source error, yet check passed). Fixes (commit 67e1c46): segments are clamped to where the
source's picture and sound really end (INFO trimmed-to-source-end); new WARN ends-mid-speech /
starts-mid-speech when the edit opens or closes within 0.3 s of the recording's edge with speech
there, with a runnable fix (end on the sentence before); the L14 short-chunk error now says re-running
won't help and how to find the real end; out-of-source fix no longer says `undo`; output size is
shown at render time and settable with `render --size 1080p` (two of four editors rendered 4K by
accident, 157–267 s); recorder transcripts mapped onto the camera clock for word anchors.
Measured: make check green; e2e 253/253. Re-opening r48/started: `timeline` now raises
ends-mid-speech on exactly the reviewer's complaint with the fix the reviewer proposed; the accepted
retake-kept stays accepted and is not listed as "still open" after later edits.
Surprises: a camera file's container duration can exceed its last video frame by ~0.1 s (started:
60.10 vs 60.009) — durations must come from the streams, not the container.
Next: T3 round r49 on all four excerpts with fresh editors, then final eval round 11.

### 47 · 2026-11-24 · Golden 5: README, go install + doctor on a clean PATH, a fresh agent follows it
Did: README.md (what cutroom is, install, doctor, quick start, deciding warnings, agents → `cutroom
guide`, development, license). `GOBIN=… go install ./cmd/cutroom` with PATH=/usr/bin:/bin and an empty
HOME: doctor lists each missing tool with its install command and the model download (exit 3); with
the tools on PATH and CUTROOM_MODELS set: all ok (exit 0). Ran the quick start verbatim: every step
exit 0, preview and final check PASS. A fresh sonnet agent (README only, fresh HOME/GOBIN): README
WORKS — install, doctor, the whole quick start, check PASS twice, captions, status "done". Its gaps,
fixed: how to decide each WARN (one at a time, code@time, apply vs accept), go install needs the
network once, examples for look/listen.
Measured: binary 8.3 MB; quick start on the 30 s fixture renders in ~1 s per pass.
Surprises: re-ran the quick start in zsh with `for c in "init talk.mp4"…; cutroom $c` — zsh doesn't
word-split, every command failed; bash only for scripted CLI runs (again).
Next: T3 — reviewer pass on the four real-footage excerpts; then the final eval round.

### 46 · 2026-11-24 · Golden 2: every INVARIANT guarded, every DETECT on a fixture (two real fixes)
Did: audited docs/LESSONS.md — 26 INVARIANT/DETECT rows, 15 had no TestLxx. Wrote guards:
render L10 (a chunk holds only overlays visible in it), L11 (key = content), L12 (encoded from
sources, joined by stream copy — joinArgs extracted), L13, L14 (a short chunk errors and isn't
cached), L15 (one input per overlay, no split), L17, L27+L43 (−30 dB tone → −16 LUFS ±1, TP ≤ −1,
48 kHz out), L47 (no drawtext/subtitles; cards drawn in Go); verify L03 (frame reads across the VFR
fixture's dropped frames keep the gap); speech L46 (brown-noise rumble louder than the voice: the
pause is still found); cli L55 (a whole session with the source folder read-only leaves it
byte-for-byte unchanged; render over a source → would-overwrite-source; export --name with a path
refused); media L04 on a new fixture talk-late.mp4 (picture 43 ms after sound) + e2e that it
renders and checks PASS. L09 was already in the L07/L08 test (renamed).
Fixed (the guards found them): L13 — Plan could end a chunk inside an overlay's fade, so the next
chunk drew it without the fade (a seam); now a boundary inside a fade waits for the next segment.
L17 — the audio cache key included chunk keys (which hash overlays/zoom/framing), so any picture
change re-ran the audio; it now hashes the range's audio pieces only. writeTarget guards render and
captions outputs against overwriting a source (L55).
Measured: each new guard fails on a mutation of the code it protects (L13, L17, L03, L46 checked by
hand). make check green; e2e 253/253; grader selftest 26/26.
Surprises: a 10 Hz sine is a bad rumble for an envelope test — a 10 ms frame catches it near zero
and the floor reads low with or without the high-pass; brown noise behaves like the real thing.
Next: Golden 5 (README, go install, doctor on a clean PATH), then an eval round on the new commands.

### 45 · 2026-11-24 · the last SPEC rows: split, pin, retakes, output transcript, look --overlays, validate
Did: `edit split <time> | --text "words"` (Timeline.Split: the segment holding that source time
becomes two at a frame, nothing removed, both new edges pinned; with --text it splits in the pause
before the words); `edit pin <seg> [--in|--out] [--unpin]`; `move` and `pin` take `--text` for the
segment that keeps those words (SPEC: "all accept text anchors"). `transcript --retakes` (each repeat,
its takes kept/cut, resolved/open), `--output` (the kept words in output time by sentence, `|` at
cuts — no render needed), `--range a-b`, `--phrases`. `look --overlays` (one frame per overlay at
mid-window, drawn as rendered; the output cell loop is now shared with --output). `timeline validate`
(problems only, exit 3 on errors). `look --crop` already existed. Guide perceive/edit/overlays;
TestSplitKeepsFrames; 13 new e2e.
Measured: make check green; e2e 251/251. Split at --text "an old laptop" lands at 11.500 (the gap
midpoint) and the edit's duration is unchanged.
Surprises: a Python str.replace on a help line hit both the main usage and the move sub-usage —
check sub-help after editing shared text.
Next: Golden 2 audit (LESSONS invariants ↔ TestLxx, DETECT ↔ fixtures).

### 44 · 2026-11-24 · M6 `export --otio --edl`
Did: `internal/export`: Events (each segment in frames at the timeline rate: source in, length,
record in; its sound source, a recorder's start mapped through its sync offset and drift);
OTIO (Timeline.1 JSON written to the published schema — Stack/Track.1/Clip.2 with media_references,
Gap.1, RationalTime.1: V1 picture, V2 overlays and V3 burned captions as their image files — text
and caption cards copied as PNGs to out/<name>_cards/ — A1 the sound); EDL (CMX 3600, NON-DROP,
record from 01:00:00:00, a B event per segment or V + A when the sound comes from a recorder, reel
names ≤ 8 A–Z0–9, overlays as comments). `cutroom export --otio --edl [--name]`; guide render +
workflow. No OTIO reader is installed (not a declared dependency), so correctness rests on
round-trip tests that parse both files back and compare frame by frame.
Tests: TestOTIORoundTrip, TestEDLRoundTrip; e2e 238/238 (+5).
Surprises: overlays added out of time order were dropped from the OTIO overlay track (the first-in
-list card took the slot and the earlier one looked like an overlap); cards are now sorted by start.
Next: the remaining SPEC rows (edit split/pin, transcript --retakes/--output, look --overlays/--crop),
then the lesson-guard audit.

### 43 · 2026-11-24 · r10: undo --dry-run previews, edit redo, moved warnings, listen's default
Did: (1) bug: `edit undo --dry-run` ran the undo (the undo branch returned before the dry-run
handling). Now it previews: "would undo: <op>; timeline would be: N segments, D" and changes
nothing. (2) `edit redo`: undo keeps the undone edit (op + the timeline it produced) on a redo
stack in .cutroom/history/redo; redo checkpoints and restores it; any new edit clears the stack;
`edit history` lists what redo would put back. (3) `--accept/--apply code@time` with a time an edit
has since moved: the only warning of that code is taken; with several, the error lists their
current times. (4) `listen` with no source uses the edit's main video, else the longest video —
it picked the 2.6 s pickup (first alphabetically) and failed "range too short". (5) guide: burned
caption groups are 1–3 words, a 4th only to avoid a weak ending (matches the output); the topic
index points at a ~1,500-word reading path (workflow, edit, verify) instead of `--all` (~4,000).
Tests: TestUndoPreviewRedo; e2e 233/233 (+5).
Next: golden criteria 1 (export OTIO/EDL; SPEC rows), 2 (lesson guard audit), 4 (T3 reviewer),
5 (README/install/doctor).

### 42 · 2026-11-24 · eval round 10: sonnet 12/12, haiku 12/12 — golden criterion 3 met
Did: all 12 tasks × 2 models; every run PASS. Second consecutive 100% sonnet round (9, 10): golden
criterion 3 holds. Asides: agents with pause-only briefs accepted the new warning with a reason,
the rest cut it. Results in `evals/results/2026-11-24-10.md`.
Found: `edit undo --dry-run` performs the undo (a real bug; no redo to recover).
Next: the r10 bug and friction, then the remaining golden criteria: 1 (SPEC commands — export
OTIO/EDL first), 2 (lesson guards audit), 4 (T3 reviewer), 5 (README, install, clean-PATH doctor).

### 41 · 2026-11-24 · r9 friction: renders find their project, asides flagged, subtitles in check
Did: (1) the render manifest records its project; `look/listen/transcript --file` and `check <file>`
without --project use it (else the nearest folder above the file with a project) — a short rendered
to ./out/short.mp4 is checked against shorts/<name>. (2) bug: `look --file x a-b` with b at/after the
file's end asked for a frame past it (frame-failed); ranges clamp to the last frame. (3) asides:
`speech.Asides` flags sentences opening with a digression marker ("by the way", "anyway", "side
note", "oops", …) as suspects (computed on load, so old transcripts get them); the timeline warns
`aside-kept` with a ready cut (`--apply`-able) or `--accept`. Asked for since round 6; no false
positives on the four T3 projects. (4) status: when the only full render is stale, the next step
says "cutroom render (out/final.mp4 is stale …)". (5) `transcript` with no source and no edit yet
shows the longest *video* (the recorder's clock isn't the edit's). (6) `check` verifies a subtitle
file delivered with the render (same name .srt/.vtt, or out/captions.*) against the edit it was
rendered from: words, timing, seams. (7) check's format line states the rotation tag; edits with
fresh WARNs end with how to decide them.
Tests: TestAsidesAtSentenceStarts; e2e 228/228 (+10; two older expectations updated for the new
aside warning).
Next: eval round 10 — the second clean sonnet round needed for golden criterion 3.

### 40 · 2026-11-24 · eval round 9: sonnet 12/12, haiku 12/12
Did: all 12 tasks × 2 models, fresh agents; every run PASS on the independent grader, broll included
(the round-8 colour fix holds: frames worst 0.2, cutaway 11/11). 2 cutroom errors in 24 runs, both
recovered by their fix lines. Results in `evals/results/2026-11-24-9.md`; friction → backlog.
Next: r9 friction (project from --file paths, the look range bug, aside suspects, stale-render next
step, transcript default, SRT in check), then round 10 — the second clean sonnet round.

### 39 · 2026-11-24 · r8 friction: paths relative to the shell, retake range, render transcripts by sentence
Did: (1) every path typed on the command line resolves against the shell's folder, --project or
not (render output, captions --output/--check, overlay add image/clip; check/init/look already did);
guide start says so. `render out/short.mp4 --project shorts/x` writes ./out/short.mp4 (short's next
step says so). (2) `no-timeline` names the folder it looked in and lists `--project shorts/…`.
(3) retake-kept quotes the range its suggested cut really removes (it runs CutWords on a copy:
6.200–9.433, as the dry-run shows; was the words' 6.31–9.22). (4) the av-offset note says the sound
track starts N s after/before the picture track — when the device began recording each — and is
kept, "not a lip-sync error". (5) `look <source-id> <time>` (a source id first means --source);
`look --file x` with no time → a 12-frame sheet of the whole file. (6) `transcript --file` (renders)
splits at sentence ends and at pauses ≥ 0.15 s (`speech.Sentences`): a tight edit's ~0.2 s pauses
made the whole render one line, and whisper may punctuate a render with commas only. (7) `edit frame
--size WxH` resizes an existing short; short's next steps say to read every sentence for asides.
Tests: TestSentencesSplitTightEdits; e2e 218/218 (+9).
Surprises: the first sentence split did nothing on the short's render — whisper wrote it as one
comma-joined sentence; the pause at a cut was exactly 0.20 s, so the split threshold is 0.15 s.
Next: eval round 9 (all 12 tasks; rounds 9 and 10 must both be 100% sonnet).

### 38 · 2026-11-24 · eval round 8 (sonnet 11/12): a clip colour bug found and fixed
Did: B-roll fixture terminal.mp4 (mandelbrot, unique frames, a 6 kHz tone as its sound); grader:
vertical-crop search (21 anchors), `alt_during` (timed whisper → frames from the clip while the
phrase is said), `speaker elsewhere`, `safe_bottom`, `no_tone=`, `vertical_9x16`; self-test 26/26.
Tasks short + broll; 24 fresh agents; results in `evals/results/2026-11-24-8.md`.
Round: sonnet 11/12, haiku 11/12 — both broll runs failed on the same cutroom bug: clip frames were
colour-shifted (luma mean 144.2 → 138.5). ffmpeg's overlay now negotiates colour spaces: the base is
labelled bt709, so the untagged clip was converted from bt601. Fix: the clip chain relabels bt709
(setparams, as sources are) and stays YUV (yuva420p); CardVersion 3 so cached chunks re-render.
cutroom's clip check had passed it (difference 14) because it compared through `format=gray`
(an RGB-matrix path) with a 150 limit; it now compares stored luma planes, limit 60: the agents'
render fails at 117, the fixed one scores 0. Re-rendered broll-sonnet's own project: grader PASS
(frames 69/69 worst 0.2, cutaway 11/11, 6 kHz tone −62 dB vs speech −29 dB).
Tests: TestClipKeepsColours; e2e 209/209.
Surprises: the agents looked at the frames and saw the clip; a 5-level luma shift is invisible on a
fractal but would tint real B-roll. The independent grader's strict frame match (30) is what caught
it — the reason it shares no code with cutroom.
Next: r8 friction (paths with --project, retake range, offset wording, look/transcript conveniences,
short output), then round 9.

### 37 · 2026-11-24 · M5 shorts: vertical sub-projects, fill framing, safe zones (L41)
Did: `cutroom short --from … --to … --name x` makes shorts/x/: a project of its own (same sources —
paths rewritten for the folder — and the parent's transcripts copied), so every command works on it
with `--project shorts/x`. Defaults (L41): 1080×1920 (--size), pads 0.06/0.16, pauses tightened with
min-cut 0.1 (~0.2 s left), the span kept from the first to the last quoted word; warns over 60 s and
when the 9:16 crop is enlarged > 1.5× (720p → 2.7×). New timeline `fill` (+ fill_x/y): the picture is
scaled to cover the frame and cropped around the anchor (`render.FrameFit`); `PictureVF` = punch-in +
fill crop is used by render, `look` (segVF) and every `check` comparison (picture, tone, overlay
windows via Manifest.SegVF), so fill timelines verify like any other. `edit frame --anchor x,y |
--fill | --fit`. Safe zones (`render.SafeZone`, 260/1490/right 120 of 1080×1920, scaled): vertical
layouts placed inside them (title centred in the safe area, top below the bar, corner left of the
buttons, lower thirds and captions above the bottom zone and centred in the safe width); text sizes
follow the frame's short side (landscape unchanged). `timeline` warns `overlay-unsafe`; `check`
reports `safe-zones` from the manifest boxes (caption groups included, full-frame clips excepted).
`status` lists a project's shorts. Contact sheets cap vertical thumbnails at 480 px tall. The argument
parser splits a `--flag value` that arrives as one word (an unsplit shell variable).
Tests: TestL41_VerticalLayoutsInSafeZones; e2e 209/209 (+11).
Measured: a 10.6 s short renders (preview) in 0.6 s; check PASS with picture worst 0.4 through the
fill crop, 11 overlay boxes inside the zones; an image at --size 0.9 is flagged ("reaches right to x
1036, safe to 960").
Surprises: (1) text sized from frame height made a vertical title 163 px tall, wrapping past the
bottom zone; sizing from the short side gives the same look in both orientations. (2) testing with an
unsplit `$P` in zsh showed the parser saying "did you mean --project?" for "--project shorts/steps" —
an agent could hit the same; now it just works.
Next: eval round 8 (short + B-roll tasks), then M6 (export, README, doctor).

### 36 · 2026-11-24 · M4 clip overlays (B-roll): cutaways and insets, checked against the clip's frames
Did: `overlay add clip <video> --at … [--to|--duration] [--clip-in s] [--layout full|corner|… --size]`
(clip probed on add: duration, displayed size; full by default, no fade, muted). Render: per chunk
the clip is read from ClipIn + the part's offset through the same SDR/rotation path as sources,
scaled to fill its box (increase + crop), `tpad=stop=-1:stop_mode=clone` holds its last frame when the
window outlasts it (`clip-short` warns with the duration that fits), `tpad=start=From` places it on
the chunk's frame grid, overlay `enable=between(n,…)` at the box position. Chunk keys include the
clip's size+mtime. `look --output` draws the clip frame for that moment (ClipCard, clamped to its end).
`check`: the window check matches on the whole frame when the box covers most of it (a full cutaway
leaves nothing around it), and a new clip-frame comparison crops the box at the render's own pixels
and compares it at 64 px wide with the clip at ClipIn + offset (5-frame burst).
`overlay set` gains --clip-in; --size works for clips.
Tests: TestClipBoxes; e2e 198/198 (+9) incl. an HDR, 180°-rotated phone clip as an inset, the
clip-short warning, and a manifest with the wrong clip-in → FAIL "doesn't show the clip's frame".
Measured: clip frames vs render: worst difference 15 (full) and 18 (inset); wrong --clip-in: 925.
Surprises: (1) the first clip-frame check compared at the 160×90 grey size: an inset is ~48×28 px
there and rounding its edges gave differences of 300–1400 on a correct render; measuring the render
frame by frame against the clip showed the render was right to the frame. (2) a full-frame box left
no picture to locate the source frame by, so the "gone after" probe compared against a neighbour
frame and reported the clip still on screen.
Next: M4 is complete (text, image, clip overlays, captions SRT + burned). M5: `short` preset and
safe zones (L41).

### 35 · 2026-11-24 · deciding warnings: `timeline --apply` and `--accept`
Did: every WARN can now be decided in one step. `cutroom timeline --apply <code>[@<output time>]` runs
the warning's own suggested edit (the backticked `cutroom edit|overlay …` in its fix; refused with
"needs your judgement" when the fix has a <placeholder> or none), `--dry-run` passes through.
`--accept <code>[@time] --reason "…"` stores {code, key, reason} in timeline.json; the key is the code
plus the message with its numbers taken out, so it survives edits that move output times but not a
different warning of the same code. Accepted warnings get level `accepted`: not printed by edits,
not counted by status, listed apart by `timeline` ("accepted (decided, not open): …"). Errors can't be
accepted. Unknown codes get did-you-mean; a repeated code without @time lists the choices. `timeline`
ends with how to decide; `status`'s next step names the first warning's `--apply` and the edit it runs.
Both go through edit history (undo-able). Answers r5 (recorder-haiku, phone-haiku), r6 (recorder ×2)
and r7 (titles-sonnet, hdr-sonnet, phone-sonnet).
Tests: TestAcceptedSurvivesMovingTimes; e2e 189/189 (+9).
Next: clip overlays (M4), then M5 (short preset + safe zones).

### 34 · 2026-11-24 · r7 friction: burn + SRT in one call, `overlay set`, audio-mixed numbers, VFR stated
Did: (1) `captions --burn` also writes the subtitle file (`--output` names it); guide shows
`--output out/final.srt`. (2) `overlay set <id>` changes --duration/--to/--at/--layout/--text/--sub/
--fade/--size in place, id kept (sub-help; refuses captions with the rm fix). New `overlay-short`
warning: a text card needs ~0.9 s + 0.3 s a word (≥ 1.5 s) — the titles-haiku card ended with its
words at 1.43 s. (3) `audio-mixed` quotes both sources' speech level and room tone (from the
transcripts): matching (< 3 dB / < 4 dB) → an INFO note with a `listen --output` across the join, not
an open warning; speech and room moved together → "same setup, N dB louder/quieter"; their gap
changed → "different room or microphone". Info notes aren't carried as "still open". (4) render says
"variable frame rate (11 gaps): placed by timestamp on a constant 30 fps grid … can't drift" and
"sound starts +0.064 s from its picture; kept as recorded"; check's sync line names the offset.
(5) init points at `cutroom sync <id>` for an audio-only source. (6) check's words line names the
model and says when it is the transcript's own ("confirms the cut, not the recogniser").
Tests: TestAudioMixedClassifies; e2e 180/180 (+11).
Surprises: the e2e take made 12 dB quieter was first called "different room or microphone":
gain moves speech and room tone together, so the speech-to-room gap, not the room level, is what
tells a different setup apart. The test caught it before any agent saw it.
Next: apply/accept a suggested fix (`timeline --apply`/`--accept`), then clip overlays (M4).

### 33 · 2026-11-24 · eval round 7: sonnet 10/10, haiku 10/10 (titles + social tasks)
Did: grader learned cards, burned captions and SRT (tile-robust frame matching only for tasks with
overlays, so earlier tasks keep their tuned thresholds; ground-truth line times from the fixture
manifest; timed whisper words for SRT sync); self-test 20/20 incl. 6 new cases. Two tasks (titles,
social); 20 fresh agents; graded; results in `evals/results/2026-11-24-7.md`, friction → backlog.
Measured: all 20 PASS, 1 cutroom error in 20 runs (a harmless `nothing-cut`), sonnet median 6 tool
calls, haiku 13. Fifth clean sonnet round (3–7).
Surprises: every agent found `overlay add` and `captions --burn` from the guide first try; the one
gap was `captions --output` (one sonnet copied the file instead).
Next: r7 friction (burn + SRT in one call, `overlay set`, audio-mixed numbers, VFR statement).

### 32 · 2026-11-24 · M4 burned-in captions: 1–3 word groups as one alpha clip per chunk (L48)
Did: `captions.Groups`: per sentence (and pauses ≥ 1 s), a small DP picks group boundaries: ≤ 20
chars, ≤ 3 words (a 4th only to avoid a weak ending), costs for weak endings (determiners/"to"
worst), lone words (short ones worse), unbalanced lengths; cheaper breaks after commas and at
pauses ≥ 0.3 s; trailing ,/. dropped; groups shown until the next starts (no flicker), else last
word + 0.4 s. `cutroom captions --burn` adds one `captions` overlay over the whole edit; groups are
computed from the cut at render time, so they follow every re-cut. Cards: bold white on a dark box,
0.068 × the short side, bottom at 0.92 h (landscape) or 0.77 h (vertical, L41 safe zone); lower
thirds move up above the caption band when captions are on. Render: each chunk's groups become one
ffconcat of card PNGs (blank PNG in gaps) with frame-count durations → `fps,format=rgba,trim` →
one overlay (L48: not ~150 overlays). `look --output` shows them (+cc); `check` samples ≤ 6 groups
on screen and the gaps between groups clean. Tests: TestL48_CaptionGroups, TestGroupsPauseAndHold,
TestL48_CaptionsOneClipPerChunk; e2e 169/169 (+6) incl. a render without the captions checked
against a manifest that has them → FAIL "not visible".
Measured: fixture 17 groups for 44 words; preview 1.4 s (1.0 s without); card vs picture ≥ 3184.
Real speech (deploy 153 words → 57 groups; sec 123 → 49): greedy grouping first gave "care about
the | ongoing cost", "If | you don't know" — pauses forced splits and the weak-word rule only looked
at one word. The DP plus soft pause breaks gives "care about | the ongoing cost", "If you don't know |
where to start", "to start with | a Docker file".
Surprises: treating a short pause as a hard split was the main cause of bad groups: speakers pause
*after* "the" and "if" while thinking.
Next: eval round 7 (overlay + captions tasks), then clip overlays.

### 31 · 2026-11-24 · M4 overlays: text cards and images anchored to words, checked frame-exactly
Did: guide topic `overlays` first. Timeline gains `overlays` (id, kind text|image, anchor source +
source time from `--at` words, `--to` words or `--duration`, layout, `--sub`, `--size`, `--fade`);
`Window` derives the output span on the frame grid (anchor in a cut pad → first kept moment of the
phrase); validate warns `overlay-orphaned` (anchor cut) and `overlay-overlap` (same layout, same
time). `cutroom overlay add|list|rm` (with sub-help; undo-able through edit history). Cards are drawn
by cutroom (Go fonts, L47): lower-third with accent bar, title/center, top, corner; images scaled
with alpha. Card PNGs are cached by a key over the spec, output size and the referenced image's bytes
(L19). Render: per chunk, each card is a looped still overlaid with `enable='between(n,a,b)'` (frame
numbers, exact on the grid), `eof_action=repeat` (L16), alpha fades counted in frames, only at the
part's own ends when it spans chunks; then `trim=end_frame` so the chunk keeps its frame count (L14);
chunk keys change only for chunks an overlay touches. Range renders shift overlays into file time.
Manifest records placements (file time + box). `look --output` composites cards as rendered (+oN).
`check`: picture compares around the cards (masked MSE); new `overlay` check per card: box differs
from the source picture inside its window (3 frames), matches it on the frame before and after.
TimelineHash includes overlays (renders go stale); captions use a CutHash without them.
Tests: TestL16_OverlayAcrossChunksIsFrameExact, TestL19_CardKeyFollowsImageBytes,
TestCardsStayInFrame; e2e 163/163 (+16) incl. two planted faults (manifest window 0.5 s late →
"still on screen"; render without the card → "not visible") and the orphaned-anchor warning.
Measured: fixture with 2 overlays: preview render 1.0 s; card vs picture difference 2014 (lower
third) and 5528 (title) inside, ≤ 60 outside; picture worst 0.3 around the cards.
Surprises: the first overlay render had 416 frames for 413: the looped still (−t chunk + 3 spare
frames) made `overlay` run past the base. A frame-count invariant (L14) caught it at once.
Next: burned-in captions (one alpha clip, L48), then clip overlays; eval round 7 after that (adds
overlay and captions tasks).

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
footage for T3 indexed in evals/local, gitignored). Ran 5 tasks × sonnet/haiku at deafd74: **10/10
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
