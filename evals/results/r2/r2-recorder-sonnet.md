# Friction report
- **Outcome:** done. out/final.mp4 uses recorder audio (offset +1.700 s, drift -42.7 ppm), pauses tightened, one retake cut; `cutroom check` PASS, 0 warnings.
- **Commands run:** 9 (guide, guide workflow + guide edit, init, sync, edit audio, edit tighten [failed], transcribe, edit tighten + render + check, transcript talk, edit cut + render + check).
- **Errors hit:** `error[no-transcript]` on tighten; fix line said run transcribe first; worked.
- **Guessed wrong:** assumed tighten could run before transcribe; `edit audio` created the timeline and succeeded anyway.
- **Confusing output:** none significant. Quoted "um" in --from matched the transcript's "UM" fine.
- **Missing:** nothing big. A single "sync-and-use" command, and a hint in `sync` output that transcription is needed before tighten.
- **Not sure it's right:** the transcript is from a small model (base.en); I did not listen to joins or view frames. I did not run `timeline --joins`.
- **Best part:** the workflow guide's recorder section gave the exact command sequence; check verifies sync per segment.
- **One change:** have `tighten` auto-run transcribe (or `init` suggest it before tighten).
