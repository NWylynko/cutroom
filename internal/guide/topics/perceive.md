<!-- summary: reading transcripts, suspects, and how far to trust word times -->
# Perceive: what was said, and where

`cutroom transcribe` makes a word-level transcript of every source with audio (cached). Then:

- `cutroom transcript` — one line per phrase with its time range and the pause after it. Read this
  first; it is the fastest way to understand a recording. Retakes stand out as repeated openings.
- `cutroom transcript --search "exact words"` — every occurrence with times and context. Use it to
  find edit points; quote words as the recogniser wrote them (case, punctuation and um/uh ignored).
- `cutroom transcript --suspects` — things to check before cutting near them:
  - `repeat`: the same 3–8 words said more than once within 30 s. Almost always a retake; the last
    take is usually the good one, but read both.
  - `sound`: sound with no words in a pause: a cough, a bump, a breath, or words the recogniser
    skipped. Don't keep it in a pause you leave in; if it might be words, check it.
  - `long-word`: a word lasting > 0.75 s; it may hide a restart.
  - `pass-diff`: the second, stumble-sensitive pass heard something different. Look there.
  - `low-confidence`: the recogniser was unsure of the word.
- `cutroom transcript --words --from 6 --to 12` — word-by-word timing for a stretch.

## See and hear
- `cutroom look 12.5` — one frame as a viewer sees it (rotation applied), labelled. `--crop x,y,w,h`
  gives a full-resolution crop: judge faces, text and colour on crops, not thumbnails.
- `cutroom look 6-12 --count 12` — a contact sheet; each frame shows its time, frame number and the
  word being spoken. Good for choosing where a cut lands visually.
- `cutroom listen 16.5-17.8` — a spectrogram with loudness and word timing drawn under it, plus what an
  isolated re-transcription of just that stretch hears. Use it to settle every doubt about a suspect:
  whether a sound is words, where a word really ends, whether a pause is clean. Red = sound with no words.
- **The edit and the render:**
  - `cutroom look --cuts` — every cut as a pair (last frame before | first frame after) with the
    words either side and the pause. Check this once after editing: jump cuts that look wrong show here.
  - `cutroom look --output 0:03-0:06` — frames of the edit in OUTPUT time (cuts marked in red).
  - `cutroom listen --output 3-5` — listen to the current render around output 3–5 s, cuts marked.
  - `--file out/final.mp4` on either command — any video file, in that file's time.
Both write a PNG and print its path. **Open the PNG** with your image/file viewing tool; the text alone
is a summary, not the picture.

## How far to trust times
Word starts come from the recogniser; the start of each phrase and the end of every word before a
pause are measured from the audio's loudness, so **pause boundaries are accurate to ~30–50 ms**.
Inside continuous speech, word boundaries are approximate (±100 ms). Cut in pauses when you can.

Times are source times in seconds (`12.5`) or `m:ss.mmm`. Any time printed by cutroom can be
passed back to any command.
