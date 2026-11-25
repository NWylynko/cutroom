# Friction report

- **Outcome:** done. out/short.mp4 is 720x1280, 7.27 s, steps only, cat aside cut, captions burned in, `check` PASS with 0 warnings.
- **Commands run:** ~14: guide; guide shorts; guide workflow; init talk.mp4; transcribe; transcript --suspects; transcript; short (1080x1920); edit cut; (rm -rf, redo) short --size 720x1280; edit cut; look --output; captions --burn; timeline; render; check; look out/short.mp4; plus a cp to out/short.mp4.
- **Errors hit:** `error[unknown-flag]: look has no flag --render` (I guessed; the fix line pointed to usage and got me unstuck).
- **Guessed wrong:** `cutroom short` leaves the aside in (cat sentence), which I had to cut separately. Assumed short had a --force/overwrite; I used rm -rf shorts/steps to redo with a different size. Guessed `look --render`.
- **Confusing output:** the WARN about enlargement says "--size 720x1280 halves the enlargement" even when I had already passed it. Final render lives in shorts/steps/out/final.mp4, and I had to copy it to out/short.mp4 myself.
- **Missing:** a way to set size on an existing short without rebuilding; a `short --output` path; a hint in the short output that the transcript has an aside between the start and end.
- **Not sure it's right:** the picture is a synthetic test pattern, so the crop around 0.5,0.5 could not be judged for a speaker; softness at 1.8x enlargement unjudged. The faint stray sound at 16.97 s was inside a tightened pause, and check reports none kept.
- **Best part:** `short` plus `captions --burn` plus `check` (safe-zones, words, sync) made verification concrete.
- **One change:** have `short` propose or auto-flag aside/off-topic phrases inside its span, and accept an output path.
