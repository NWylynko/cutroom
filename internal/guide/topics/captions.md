<!-- summary: caption files (SRT/VTT) of the edit: how cues are made, what the caption check catches -->
# Captions

`cutroom captions` writes the edit's words as a caption file in **output** time: `out/captions.srt`
(`--vtt` for WebVTT; `--output out/final.srt` to name it). Upload it next to the video (YouTube, Vimeo) — it is not burned into the picture.

```
$ cutroom captions
wrote out/captions.srt: 8 cues, 45 words over 0:13.767 (≤ 42 chars × 2 lines, 1.1–3.4 s each)
   1  0:00.000–0:03.300  Hi, today I'm going to show you
                         how to set up a home server.
   2  0:03.400–0:05.633  The first thing you need is
                         an old laptop.
   …
ok   reading  longest line 38 chars, fastest cue 15 chars/s (limit 20)
ok   seams    3 cut(s): no cue opens on a sentence tail, every sentence before a cut ends
next: cutroom captions --check out/captions.srt   (after editing the file by hand)
```

How cues are made (so they read like a person made them):
- Words come from the transcript of what the timeline keeps; um/uh are left out.
- A cue holds at most 2 lines of 42 characters and 7 s; a new sentence starts a new cue when the
  current one has room to end there; a pause ≥ 0.7 s always starts a new cue.
- A cue or line never ends on a weak word ("the", "a", "to", "and", …) when it can be avoided, and
  never leaves a 1–2 word sentence tail to open the next cue (L48, L49).
- Each cue shows from its first word until its last word + 0.3 s (at least 1 s when there is room),
  never overlapping the next.

Checks (printed every time; `--check <file>` runs them on an existing or hand-edited file):
- **reading**: line length, lines per cue, reading speed (characters per second ≤ 20).
- **seams** (L49): a cue that opens on the tail of a sentence whose start was cut; a sentence that
  stops at a cut without ending ("you need is | Then install") — fix the edit, not the captions:
  `cutroom timeline --joins` shows the join.
- **words**: the file's words are the timeline's words, in order (catches a stale file after re-editing).
- **timing**: cues in order, no overlaps, all inside the video.

Captions follow the timeline: after any edit, run `cutroom captions` again (`cutroom status` says
when the file is stale).

## Burned-in captions (in the picture)

`cutroom captions --burn` draws captions into the video itself (and writes the subtitle file in the
same call; `--output` names it), the short-form way: 1–3 words at a
time (≤ 20 characters; a 4th word only when it avoids ending on "is", "the" and the like), never ending on a weak word, never leaving a lone short word (L48), white on a
dark box low in the frame (on vertical video, above the platform's buttons). They are one overlay
over the whole edit: the groups are made from the cut at render time, so they follow every re-cut.
`cutroom look --output <time>` shows them (`+cc`); `cutroom check` samples groups on screen and the
gaps between them; lower thirds move up above them. Remove with `cutroom overlay rm <id>`.
Burned-in captions show every misheard word to everyone, so `captions --burn` lists the words the
recogniser wasn't sure of (low confidence, or its two passes disagreed). Listen (`cutroom listen <source
time>`), then correct one with `cutroom transcript --fix "other work providers" --as "other OAuth
providers"` — captions, subtitle files and word checks use the fix; the recording is unchanged
(`--near <time>` for one place only, `--fixes` lists them, `--unfix` removes one). In upper-case
captions, `style captions --case upper-keep-names` leaves iCloud, OAuth, GitHub as written.
That look is only the default: font, size, colours, box, outline, place, words at a time and a
karaoke highlight are all yours to set — `cutroom guide style`.
