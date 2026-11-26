# Friction report

- **Outcome:** done. out/short.mp4 is 720x1280, 7.27 s, steps only, aside cut, pink outlined 2-word caps captions, white rounded pop-in title card; `check` PASS with 0 warnings.
- **Commands run:** ~16: guide; guide shorts/style/workflow; init; transcribe; transcript --suspects; transcript; short; edit cut; timeline; edit frame --size 720x1280; style captions; guide overlays; style --help; overlay add; style cards; style o1; timeline; look x2; captions --burn; render; check; look x2.
- **Errors hit:** none.
- **Guessed wrong:** nothing major. Did not know captions need `captions --burn` after `style captions` (style alone doesn't add them); found it in the shorts guide.
- **Confusing output:** `short` keeps the "By the way" aside and only warns; had to cut it manually (expected, documented). `style cards` prints a long field table every time, and `style o1` too. The card default has align left on cards but o1 shows center, minor.
- **Missing:** a one-shot flag for a title card position per safe zone; no way to see an animation (pop) in `look`, only a still. Could not verify the pop actually animates.
- **Not sure it's right:** pop animation and the pink exact colour in the final mp4 (looked only at the `look` still; check confirms captions/card present). Source is 720p so 1080x1920 would be soft; I chose 720x1280.
- **Best part:** `check` safe-zones/overlay/captions lines and text-anchored cuts; `look` draws overlays and captions as rendered.
- **One change:** have `short` auto-flag/offer to cut the detected aside and mention `captions --burn` in the `style captions` next: lines.
