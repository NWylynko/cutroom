<!-- summary: what cutroom is, how to read its output, the 60-second version -->
# cutroom — start here

cutroom edits video for you, an AI agent. You can't watch video or hear audio, so cutroom turns
media into text and images you can reason over, turns your text decisions into a frame-exact video,
and then checks the render the same way so you can prove it matches what you intended.

The loop is: **perceive → decide → act → verify.**

## Reading cutroom's output
- Default output is short text. Add `--json` to any command for the complete result as one object:
  `{"ok": true, "result": {...}, "next": [...]}` or `{"ok": false, "error": {"code", "message", "fix"}}`.
- Lines starting `next:` are the suggested next commands. Following them is usually right.
- Errors print `error[code]: message` and `fix: what to do`. Do the fix; don't guess around it.
- Exit codes: 0 ok · 1 usage · 2 invalid project/timeline · 3 missing dependency · 4 check failed ·
  5 external tool (ffmpeg, whisper) failed.
- Lists are truncated with a note saying how to see more. Ask for more only when you need it.
- Paths you type are relative to your current folder, like any command; `--project <dir>` only says
  which project to work on (its default outputs, like out/final.mp4, go inside it).

## Times
Accepted everywhere a time is: `12.5` (seconds), `1:02.5`, `1:00:02.5`, `f372` (frame number).
Prefer quoting transcript text over typing times when a command accepts text anchors: you quote
text reliably; you miscount seconds.

## Rules cutroom enforces for you
Source files are never modified. Cuts snap to whole frames and audio samples. Video is encoded once
from the sources. You don't need to manage any of that.

## Rules you must follow
1. Never call an edit done without running `cutroom check` on the render and reading every line.
2. Look at images cutroom gives you (frames, contact sheets, spectrograms) before judging visuals.
3. Test small first: `render --preview --range` before a full render.

Topics: run `cutroom guide <topic>`. For most edits read `workflow`, then `edit` and `verify`
(~1,500 words); the others when a task needs them (`cutroom guide --all` prints all ~4,000 words).
