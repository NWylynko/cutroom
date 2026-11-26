# T3 r65 — editor review of three styled real-footage shorts (2026-11-26)

Shorts made with `short --look channel` (Avenir Black upper-case captions, yellow highlight, pop;
light rounded card) from three productions: sso (securing, 4K HLG, rotated), mail (my-personal,
4K 59.94 fps camera + synced recorder), github2 (getting-started, 1080p). Reviewer: fresh opus
agent as a short-form editor, with full-resolution frames.

Verdicts: REJECT ×3 (sso and github2 "one caption fix from ACCEPT").

Tool problems found → fixed in iteration 65:
- caption groups over the asked count (3 words when 2 were asked) → `--words` is a hard limit
- groups across sentence/clause seams and cuts ("INBOX I", "DO THIS SO") → groups never span a
  cut; a recorder's punctuation-free words borrow the camera transcript's punctuation; crossing a
  comma costs more than breaking at it
- 2-word groups wrapping to 2 lines → captions use the safe width before wrapping
- pop read as a fade → 70%→108%→100%, opaque by the second frame
- endings cut on the last word → shorts hold up to 0.45 s after it
- "signed in with that" stumble: the camera transcript missed "Apple.", tightening cut it as a pause →
  `short` uses a synced recorder's sound and words from the start
- found while fixing: caption cards snapped to 40 ms (±1 frame) by the concat demuxer → L58

Not fixed (→ BACKLOG): misheard technical words burned in ("work"/OAuth, "greater"/great at);
card that restates the payoff; sso's crop is tight because the 4K frame is full-height (a fill crop
can't zoom out — `--fit --bars blur` is the alternative); brand casing in upper case (iCloud).
