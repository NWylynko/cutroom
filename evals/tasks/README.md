# Agent eval tasks (tier T2)

Each task is `<name>.md` (the brief the agent sees, verbatim) and `<name>.json` (grading spec, the
agent never sees it). The runner copies the named fixtures into a fresh working directory, gives the
agent the brief plus this preamble, and grades the result with `go run ./evals/grade <task> <dir>`.

Preamble given to every agent:

> You are editing video with the `cutroom` CLI, which is on your PATH. Start with `cutroom guide`.
> Work only inside <dir>. Don't look for cutroom's source code. When you finish, write a friction
> report to <dir>/FRICTION.md using the template that follows.

Grading spec fields: `fixtures` (files to copy), `output` (path the result must be at),
`must_contain` / `must_not_contain` (phrases, matched loosely against an independent whisper pass
of the output), `duration` ([min, max] seconds), `checks` (named checks: `upright`, `no_rotation_meta`,
`av_sync`, `size=WxH`, `loudness=-16±1`, `srt`).
