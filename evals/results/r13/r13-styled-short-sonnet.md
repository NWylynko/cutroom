# Friction report

- **Outcome:** done. `out/short.mp4` is 720x1280, 7.5 s, retakes and the cat aside cut, styled captions and title card burned in, `cutroom check` PASS with 0 warnings.
- **Commands run:** about 17: guide; guide shorts/style/workflow; init; transcribe; transcript; guide edit/overlays; short; timeline; edit cut; style captions; style cards; overlay add text; timeline; look --overlays; look --output; captions --burn; render; check; look --file.
- **Errors hit:** none.
- **Guessed wrong:** nothing big. I did not know whether `style cards --box` meant panel colour (it does) and whether card text colour is `--color` (it is). Guessed `--layout top` for the title; it worked.
- **Confusing output:** `short` kept the aside ("By the way, my cat...") between the steps; only a WARN flagged it, so I had to notice and cut. Output is 720x1280, not 1080x1920, because the source is 720p (explained, fine). The 1080p guide example differs from my result.
- **Missing:** `short` doesn't offer `--cut` for known asides. No way to preview the pop animation except rendering.
- **Not sure it's right:** the pop animation motion itself (only checked settled frames); "a little below the middle" (y=0.6) is my interpretation. Whether the framing is ok: no speaker was found (test pattern), crop centred.
- **Best part:** quoted-word anchors for cut/overlay, and `check` covering overlay, captions, safe-zones and words in one pass.
- **One change:** have `short` auto-drop (or ask about) a flagged aside inside the span, and print the final style/safe-zone summary in `captions --burn`.
