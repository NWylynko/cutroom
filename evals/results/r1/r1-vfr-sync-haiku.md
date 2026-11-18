# Friction report (fill this in as the last thing you do)

- **Outcome:** partly done. The false start and long pauses are cut and `out/final.mp4` renders at 17.233 s, matching the timeline, but I could not verify lip sync, which was the user's main concern, because `cutroom check` is absent and nothing re-transcribes or inspects a render.

- **Commands run:** 10 shell calls, about 21 cutroom invocations. In order:
  1. `cutroom guide`
  2. `cutroom guide workflow`, `edit`, `render`, `pitfalls`, `perceive`
  3. `ls -la`, `cutroom doctor`, `cutroom init talk-vfr.mp4`, `cutroom inspect`
  4. `cutroom transcribe`, `cutroom transcript`, `cutroom transcript --suspects`
  5. `cutroom edit tighten`, `cutroom edit cut --from "The first thing you need is, UM" --to "the first thing." --dry-run`
  6. `cutroom edit cut` (same args, applied), `cutroom timeline`, `cutroom timeline --joins`, `cutroom listen 16.5-17.8 --json`
  7. `cutroom render --dry-run`, `cutroom render --preview`
  8. `cutroom --help`, `cutroom render --help`
  9. `ffprobe` on `out/preview.mp4` and `talk-vfr.mp4`, `cutroom look 3.9 --json`
  10. `cutroom render`, `ffprobe` on `out/final.mp4`, `cutroom transcript --search "need is"`

- **Errors hit:** none. No `error[...]` lines and no `fix:` lines appeared.

- **Guessed wrong:** nothing failed, but I was unsure whether `--from` would match the whole phrase. The guide says it must be unique, and it is not here, so I expected a numbered list. It was not one because I quoted the full false-start phrase. I also expected `cutroom check` to exist from the guide's workflow step 6, but it is marked "not available in this version yet".

- **Confusing output:**
  - `listen --json` returned `"heard_isolated": "BLANK_AUDIO]"`, with the opening bracket missing. It looks like a truncation artefact.
  - The `tighten` output said `WARN[retake-kept] ... still said 2× in the timeline`. It came before the cut, so I had to work out that the cut was what fixed it.
  - The `init` output says "project created" but lists no `cutroom.json`, and I could not see whether init wrote one.

- **Missing:**
  - **A post-render check (`cutroom check`).** This is the big gap for a lip-sync task. The guide says to check before calling anything done, and it does not exist yet.
  - **Any way to transcribe or inspect a rendered file.** `transcribe` only takes sources, so I could not re-transcribe `out/final.mp4` to compare word times with the video.
  - **A sync report.** `inspect` says the audio starts +0.064 s and that this is "handled", but I could not see how the renderer applies the offset or verify it on the output. The only evidence I have is ffprobe: both output streams start at 0.000 s and have the same duration. That does not prove they are in sync.
  - A way to view `look` on the output rather than the source, so I could check a cut in the rendered frames.
  - I could not tell from the guide whether the cut-free "cat" aside (0:18.6 to 0:21.7) should go. The task only asked for pauses and the false start, so I left it in and am flagging it.

- **Not sure it's right:**
  - **Lip sync is unverified.** The source is VFR with an audio start offset of +0.064 s. The tool says it is handled, and the output streams start at 0 with equal length, but I could not confirm that the speech lines up with the on-screen action.
  - The false-start cut at 6.230–9.430 s removed "The first thing you need is, UM, the first thing." and kept "The first thing you need is an old laptop." I checked this from the transcript only, not by listening to the join.
  - The 0.33 s unworded sound at 16.97 s sits inside a pause that is now removed (`listen` confirmed it is the only sound in that range), so the output should not have it.
  - Base.en whisper model: the word times I cut on may be slightly off, and I did not run a larger model.

- **Best part:** `cutroom transcript --suspects` flagged the retake and the stray sound directly, and `cutroom edit cut --from/--to --dry-run` showed exactly which range would go before I applied it. The `timeline` and `timeline --joins` output was clear and short.

- **One change:** add `cutroom check <render>` that re-transcribes the rendered file, compares word times with the timeline's output times, and reports the audio-to-video offset in ms per segment. That would have let me verify the one thing the user asked about.
