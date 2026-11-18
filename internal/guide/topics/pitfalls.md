<!-- summary: proven traps in automatic video editing, and dead ends not to try -->
# Pitfalls

These come from real productions. cutroom already protects you from the mechanical ones; the rest
need your judgement.

## Transcripts lie
- Speech-to-text smooths over stumbles and stretches one word's timing over a hidden restart or a
  dropped take. A word lasting longer than ~0.75 s at a cut is a suspect.
- Long transcription passes skip whole phrases. If a phrase seems missing, listen to a 10–15 s
  isolated clip of that spot before believing it.
- When a phrase repeats near a cut (a retake), check each side as a short isolated clip before
  choosing which one to keep.
- For sub-second restarts, a spectrogram and loudness envelope beat re-transcription.

## Cuts
- Word ends are reported early: trailing "s" sounds carry on 0.1–0.2 s. Leave ~0.22 s after a word
  and ~0.10 s before the next one.
- A loud stretch between the last kept word and the cut is a fragment of something you meant to drop.
- Hand-picked edges beat automatic pause tightening near a restart.

## Picture
- Punch-ins above 1.2× on 1080p sources look soft.
- Judge colour and faces on full-size crops, not on averages or tiny contact sheets.
- Phone footage may be rotated by metadata or be HDR; `inspect` tells you, and the renderer handles it.

## Audio
- Judge harshness as if on phone speakers: hiss gets louder with compression and loudness gain.
- Dead ends (don't spend time on them): ffmpeg `deesser`/`adynamicequalizer` (no effect on speech
  here), `afftdn`/gates for pause noise (barely help), stem separation for denoising (worse than a
  speech denoiser). Nothing removes mic-handling thumps; flag them to a human instead.
