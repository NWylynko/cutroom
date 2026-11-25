# Friction report

- **Outcome:** done. out/final.mp4 is 13.8 s, SDR bt709, rotation baked in upright, and `cutroom check` passed with 0 failures and 0 warnings.
- **Commands run:** 9 cutroom commands, plus 1 file read.
  1. `cutroom guide`
  2. `cutroom guide workflow; cutroom guide pitfalls`
  3. `cutroom init clip.mov`
  4. `cutroom inspect`
  5. `cutroom transcribe`
  6. `cutroom transcript`
  7. `cutroom guide edit`
  8. `cutroom edit tighten`
  9. `cutroom edit cut` (false start)
  10. `cutroom edit cut` (cat remark)
  11. `cutroom timeline`
  12. `cutroom render`
  13. `cutroom check`
  14. `cutroom look --output 3`
- **Errors hit:** none.
- **Guessed wrong:** nothing.
- **Confusing output:** none. The `retake-kept` warning from `tighten` appeared before I cut, and I ignored it because I cut the retake myself. That was fine.
- **Missing:** nothing.
- **Not sure it's right:** the tone-mapped look was checked numerically by `check` (saturation 87 vs 87). My only visual check was one frame via `look`. The frame showed in the viewer, and I did not study it closely.
- **Best part:** HDR and rotation were handled automatically, and `check` reported them. The `inspect` warnings were marked `handled`, so no action was needed.
- **One change:** have `render` run `check` itself, or print the `look` image path automatically.
