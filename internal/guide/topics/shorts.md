<!-- summary: vertical shorts from a span of the talk — their own project, short-form defaults, safe zones -->
# Shorts

`cutroom short --from "first words" --to "last words." --name laptop` makes a vertical short of that
span in `shorts/laptop/`: a project of its own with the same sources and transcripts. Every command
works on it with `--project shorts/laptop` (render writes `shorts/laptop/out/final.mp4`; paths you
give are relative to your current folder, so `cutroom render out/short.mp4 --project shorts/laptop`
writes `./out/short.mp4`).

```
$ cutroom short --from "The first thing you need is an" --to "update command." --name steps
made shorts/steps: talk 0:10.250–0:25.990 "The first thing you need is" … "and run the update command."
  1080x1920, picture filled and cropped around 0.50,0.50; pads 0.06/0.16 s; 3 pause(s) tightened (5.4 s out)
  timeline: 4 segments, 10.567 s
```

Short-form defaults (L41):
- 1080×1920, or 720×1280 for a 720p-class source (it would look soft enlarged); `--size WxH` to
  choose, later `cutroom edit frame --size 720x1280 --project …`. The landscape picture is scaled
  to fill the frame and cropped; `short` finds the speaker (skin-coloured
  movement — a face talking, hands) and centres the crop on them, saying where (`framing: on the
  speaker, found at x 0.74`) or that it found no one (then it is centred; a screen recording, say).
  `cutroom edit frame --anchor auto` finds them again after re-cutting; `--anchor 0.4,0.5` centres the crop
  on that point of the source, e.g. the speaker's face (0–1 across, down) — look with `cutroom look --output <time> --project …`.
  `edit frame --fit` keeps the whole picture with bars instead — `--bars blur` (the picture blurred
  behind itself, as many shorts do), `--bars "#101718"` (a colour) or black.
- Pads at cuts 0.06 s before / 0.16 s after; every pause tightened to ~0.2 s.
- Open on the claim and drop every aside: `short` keeps everything between --from and --to, but
  names any aside it finds there; `--cut-asides` cuts them in the same call ("By the way, my cat…").
  Still read `cutroom timeline --project …` sentence by sentence and cut side points (`edit cut`).
  It warns above 60 s.
- Safe zones: on a vertical frame the app's title bar, buttons and caption area cover the top
  (above y 260 of 1920), the bottom (below y 1490) and the right 120 px. Title, lower-third and
  corner layouts and burned captions (`cutroom captions --burn --project …`) are placed inside them;
  `timeline` warns `overlay-unsafe` for anything that isn't (a big image, say) and `check` reports
  `safe-zones` on the render.
- The captions' and cards' look is yours: `cutroom style captions --font "…" --highlight "#ffd400"
  --words 2 --case upper --project …` (karaoke highlight, outline, no box, any font) — `cutroom guide style`.
  Keep it as a look (`cutroom style save channel`) and start the next short with `--look channel`.
