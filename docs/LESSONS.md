# Lessons → requirements

Source: a log of producing several published YouTube videos headless (ffmpeg + whisperx + Python,
driven by an AI agent), Oct–Nov 2026. Each lesson cost a debugging session. Every row says what it
becomes in cutroom. IDs (L01…) are referenced from code comments and tests: a test that guards a
lesson is named after it.

Categories: **INVARIANT** (built in, impossible to get wrong) · **DETECT** (`inspect` warns + fix) ·
**CHECK** (`check` flags it) · **DEFAULT** · **PROVIDER** · **GUIDE** (taught in `cutroom guide`) ·
**NON-GOAL** (proven dead end; don't build, tell agents not to try).

## Time and sync

| ID | Lesson | Becomes |
|---|---|---|
| L01 | Float-second `trim` dropped a frame on some segments: A/V drift grew to +113 ms over 14 cuts. Cutting with `trim=start_frame/end_frame` + `atrim=start_sample/end_sample` kept every segment within ±17 ms. | INVARIANT: timeline stores rational time; every cut resolves to an integer frame and an integer sample before any ffmpeg call. |
| L02 | VFR source (dropped-frame gaps, audio starting 0.085 s in): counting real frames fell 8+ frames behind by 12 min. Fix: `fps=<rate>:start_time=<first audio sample>` before trimming so frame k = audio time k/fps. | DETECT + INVARIANT: inspect measures VFR and the audio start; the renderer always resamples to a CFR grid anchored on the audio. |
| L03 | Extracting frames with `-ss … -f rawvideo` on VFR without `-fps_mode passthrough` duplicates frames into gaps (phantom +141 ms). | INVARIANT inside `look`/`check` frame extraction. |
| L04 | Concat of camera takes by stream copy shifted video 43 ms (HEVC reorder delay, edit list dropped). | DETECT: inspect reports first-frame pts vs audio; multi-source timelines never rely on a joined file's pts. |
| L05 | Separate audio recorder: offset ~11 s and clock drift −27 ppm (30 ms over 20 min). Robust fit of windowed cross-correlation; plain least squares was pulled off by silent windows. | Command `sync` (offset + drift, robust fit) writing a corrected audio track. |
| L06 | A render one frame short is harmless when the last segment runs past the source's last frame. | CHECK tolerates exactly that case and says why. |
| L07 | `-itsoffset` after an accurate `-ss` stamps from the keyframe, not the seek point (inset ran 0.65 s early, then froze before every cut). | INVARIANT: never `-itsoffset` on a seeked long-GOP read; shift with `setpts`. Lint test on generated commands. |
| L08 | `concat` output pts are truncated microseconds; overlays took the previous frame every third frame. Restamp `settb=1/fps,setpts=N`. | INVARIANT in the graph builder. |
| L09 | Seeking exactly on a frame boundary is ambiguous; seek mid-frame when comparing. Seek a spare half-frame early and read 3 spare frames past a piece. | INVARIANT in renderer and checker. |

## Rendering and graphs

| ID | Lesson | Becomes |
|---|---|---|
| L10 | Every overlay in one graph costs time on every frame, on screen or not (131 overlays: 60 s of output took 165 s instead of 15 s). | INVARIANT: chunked rendering; each chunk only includes the overlays visible in it. |
| L11 | Chunking at cuts (~45 s), parallel jobs, content-addressed cache: 21 min → 5m41s; unchanged re-render ~1 s, bit-identical. | INVARIANT: chunk cache keyed by content (segments, overlay file hashes, encoder settings, tool version). |
| L12 | Never chain lossy encodes; join chunks by stream copy, encode once from source. | INVARIANT. |
| L13 | Chunk boundaries inside a fade cause visible seams. | INVARIANT: no chunk boundary inside any overlay fade. |
| L14 | Each chunk checks its frame count; a short chunk must fail loudly. | INVARIANT. |
| L15 | Feeding an overlay from a branch (`split`) of the main stream stalls the graph (buffers until the overlay starts; render hung 19 min). | INVARIANT: every overlay/inset has its own input. |
| L16 | `overlay eof_action=pass` drops the secondary's last frame (one-frame flash at cuts). Use `repeat` + an `enable` window when the last frame matters. | INVARIANT. |
| L17 | Audio rendered once for the whole video (loudness measured globally), muxed onto the joined chunks; audio-only fixes remux with `-c copy`. | INVARIANT; audio cache keyed only on audio inputs (a picture change must not re-run audio). |
| L18 | Audio pass on a long 48 kHz source was slow (every `atrim` reads up to its end). | Design: cut audio per segment with seeks, then measure. |
| L19 | Card/asset caches keyed on the HTML only missed changed referenced files (logos, CSS). | INVARIANT: cache keys hash every referenced file. |
| L20 | Stills at cuts missed four one-frame bugs; a frame-by-frame PSNR comparison against an overlay-free render of the same timeline found them. | CHECK: `check --frames` compares against an overlay-free reference. |
| L21 | Renders at different scales can't be compared per frame (±1 frame scatter); motion cross-correlation over windows settles it. | CHECK implementation detail. |
| L22 | 20 ms audio fades at segment edges cover ≤1 frame of a clipped onset. | DEFAULT. |

## Source formats

| ID | Lesson | Becomes |
|---|---|---|
| L23 | iPhone HLG/HDR (bt2020, arib-std-b67): software decode with a matrix change is +20/255 washed out. VideoToolbox `scale_vt` to bt709 matched Apple's own SDR export within 1–2/255. | DETECT + DEFAULT: HDR sources get a tone-map path (VideoToolbox on macOS; report clearly when unavailable). |
| L24 | Downloaded hw frames keep the bt2020nc tag → mis-tagged output. `setparams` retag needed. | INVARIANT: output always tagged bt709. |
| L25 | Display-matrix rotation −180: hardware frames aren't autorotated, render came out upside down. PNG stills hid it (ffmpeg rotates PNGs). | DETECT + CHECK: orientation is checked on an encoded sample, never a PNG. |
| L26 | Two takes of one shoot came back rotated 0.45° and colour-shifted; region-mean colour matching made skin look sick. | GUIDE: show face crops of options to a human; NON-GOAL: automatic grading in v1. |
| L27 | `loudnorm` outputs 192 kHz. | INVARIANT: always `-ar 48000` (or the timeline rate) on output. |

## Transcripts (the agent's eyes and ears)

| ID | Lesson | Becomes |
|---|---|---|
| L28 | The transcriber smooths over stumbles and stretches word times over hidden restarts and dropped takes. | DETECT: `transcribe` flags suspects: words longer than ~0.75 s, loud gaps, repeated n-grams. |
| L29 | A second pass with a disfluency prompt surfaced restarts the default pass hid. Diffing the two passes finds them. | PROVIDER feature: `transcribe --passes 2` diffs and reports disagreements. |
| L30 | The long-form pass skips whole phrases; isolated 10–15 s clips hear them. Don't use the disfluency prompt on short clips (it hallucinates). | `listen` transcribes an isolated clip without the prompt. GUIDE: settle "missing" words with `listen`. |
| L31 | Sub-second restarts: re-transcription can't tell; a 10 ms RMS envelope plus a spectrogram sets the cut. | `listen` outputs spectrogram + RMS envelope. |
| L32 | Word ends land early: sibilants run 0.1–0.2 s past the aligned end. | DEFAULT: pad 0.22 s after a word, 0.10 s before. |
| L33 | A loud tail pad (last word end → out) after rendering means a fragment of a dropped take. | CHECK: tail/head pad RMS vs room tone. |
| L34 | Spectrograms beat waveforms for "listening" to gaps. | `listen` default output is a spectrogram. |
| L35 | When a phrase repeats near a cut, transcribe each side as an isolated 2–4 s clip before choosing the edge. | GUIDE + `listen` around a cut. |
| L36 | Ad-libbed recordings: script matching fails; a gap-split phrase list makes retakes obvious (repeated openings). | `transcript --phrases` view; `retakes` detector (repeated openings). |
| L37 | Cue times taken from the source transcript land wrong where its word times are invented; cues from a transcript of the render are right. | Overlays anchor to text; `check` re-transcribes the render and validates anchors. |

## Edits and pacing

| ID | Lesson | Becomes |
|---|---|---|
| L38 | Pause tightening: `min-cut 0.4 s` gave half the cuts of 0.15 s for ~10 s less trimmed — the casual default. | DEFAULT. |
| L39 | Exact, hand-chosen edges must win over automatic tightening, which cut a real word next to a restart. | INVARIANT: edits have provenance; `tighten` never moves a pinned edge. |
| L40 | Punch-ins: 1.2× max on 1080p; anchor on the face. | DEFAULT + validation warning above the limit. |
| L41 | Vertical shorts: open on the claim, drop every aside, pauses ≤0.2 s, edge pads 0.06/0.16, safe zones (title ≥ y 260, captions ≤ y 1490, nothing in right ~120 px). | Preset `short`; CHECK safe zones. |

## Audio

| ID | Lesson | Becomes |
|---|---|---|
| L42 | DeepFilterNet (attenuation cap 18 dB) removed a siren −29 dB with voice −0.5 dB; full strength ate soft word endings. | PROVIDER (optional denoise) with the cap as default. |
| L43 | Single-pass `loudnorm` missed target; two-pass `linear` silently fell back to dynamic. Measured gain + `alimiter` (level=0, latency=1) then re-measure hit −14.2 LUFS exactly with sync unchanged. | INVARIANT: loudness is measure → gain → limit → re-measure; report achieved LUFS/TP. |
| L44 | Harsh on phone speakers: compressor + gain raised hiss ~3 dB; a static high shelf fixed it. Phone speakers are the test. | DEFAULT chain; GUIDE. |
| L45 | `deesser`, `adynamicequalizer` were no-ops; `afftdn` and gates barely touch pause noise; Demucs worse than DeepFilterNet; nothing removes mic-handling thumps. | NON-GOAL; GUIDE tells agents not to try. |
| L46 | Silence detection found nothing until an 80 Hz highpass removed rumble. | INVARIANT inside pause/gap analysis. |

## Text, captions, graphics

| ID | Lesson | Becomes |
|---|---|---|
| L47 | This ffmpeg has no `drawtext` and no `subtitles`/libass. Text must arrive as images. | INVARIANT: cutroom rasterises text itself (Go, embedded font); no dependency on ffmpeg text filters. |
| L48 | Captions: groups of 1–3 words, ≤20 chars, never end on a weak word, never strand a short word; one alpha clip instead of ~150 overlays. | Built-in caption renderer as one overlay clip. |
| L49 | SRT cues opening on a stranded sentence tail; drops leave unpunctuated sentence seams. | CHECK on SRT output. |
| L50 | A card's animation not started yet was ghosted at ~10% opacity in Chrome. | GUIDE for HTML provider. NON-GOAL v1: HTML rendering in core. |
| L51 | Check a contact sheet of every overlay's full state over a real frame before wiring cues. | `look --overlays`. |

## Process (how agents went wrong)

| ID | Lesson | Becomes |
|---|---|---|
| L52 | Called a grade done from numbers and a 640 px contact sheet; the human saw it at full size. | GUIDE: judge on full-size crops; `look` returns full-resolution crops on request. |
| L53 | Plan → approve → render; test small first (~30 s). | `render --range`, `--preview`, `--dry-run` with cost estimate. |
| L54 | Durations and totals must be computed, never estimated. | Every command reports computed durations. |
| L55 | Sources are read-only. | INVARIANT: cutroom never writes next to or over a source. |
| L56 | Measured costs: 1080p x264 CRF 18 medium ≈ 0.35× realtime on M1 Pro; 4K HEVC chunk ~6 min per 49 s; whisper ~3 min per 8 min of audio (CPU). | Cost model seed for `--dry-run`; refined from the project's own timing history. |
