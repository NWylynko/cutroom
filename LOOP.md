# LOOP.md — the development loop

cutroom is built by an autonomous loop: design → build → test → have fresh agents use it → fix what
confused them → repeat, until it's golden. The human is not in the loop. This file is the loop's
charter; read it at the start of every iteration.

Files: `docs/SPEC.md` (what we're building), `docs/LESSONS.md` (requirements from real productions,
IDs L01…), `BACKLOG.md` (prioritised work), `JOURNAL.md` (one entry per iteration, newest first),
`evals/` (fixtures, tasks, graders, results).

## One iteration (≈ one backlog item, ≤ ~45 min of work)

1. **Orient.** Read this file, the top of `BACKLOG.md`, the last 3 `JOURNAL.md` entries, `git log -5`.
2. **Pick** the top unblocked item. If an eval round is due (every 5th iteration, or after finishing a
   milestone), do that instead.
3. **Design briefly** in the journal entry: what, why, which lessons/eval findings it serves, how it
   will be tested. For a new command, write its `guide` text and example output FIRST — the agent
   interface is the product; the implementation serves it.
4. **Build** with tests. Every INVARIANT lesson touched gets a test named `TestLxx_…`.
5. **Verify** — all must pass before commit:
   - `make check` (gofmt, go vet, `go test ./...`)
   - `make e2e` (scripted runs on synthetic fixtures; asserts on outputs, not just exit codes)
   - Look at any new visual output yourself (Read the PNG). Compute, don't estimate.
6. **Commit** locally (`git commit`, message = journal title). Never push, never publish.
7. **Journal** the entry: done, measured numbers, what surprised you, follow-ups → `BACKLOG.md`.
8. **Schedule** the next iteration (or stop, see below).

## Eval rounds (the part that makes it agent-native)

Run every 5th iteration and at each milestone. Three tiers:

- **T1 scripted** — `make e2e`: fixed command sequences on synthetic fixtures.
- **T2 agent evals** — for each task in `evals/tasks/*.md`, spawn a **fresh** subagent (no access to
  this conversation or cutroom's source) with only: the task brief, a working directory with the
  fixture, and `cutroom` on PATH. It must say "Start with `cutroom guide`" and nothing else about
  usage. Run each task with sonnet (the gate: real users are expected to drive cutroom with
  Sonnet/Opus-class models) and haiku (cheap early-warning signal for friction). The subagent ends with
  a **friction report** (`evals/FRICTION_TEMPLATE.md`). Then grade with `evals/grade.sh <task> <dir>`,
  which checks the output independently of cutroom's own `check` (ffprobe, an independent whisper
  pass, expected/forbidden phrases, duration bounds, sync).
- **T3 real footage** (each milestone, local only) — 60–120 s excerpts of real recordings listed in
  `evals/local/` (gitignored; never commit real footage or anything from it). A reviewer subagent
  (strongest model) gets the output transcript and a contact sheet and judges the edit as an editor would.

### Running a T2 task (exact protocol)
1. `make build fixtures`, then `evals/setup-runs.sh <round> <task>…`: it creates
   `evals/work/runs/<round>-<task>-<model>` for haiku and sonnet, copies the task's `fixtures` and runs
   its `setup` there with `bin/` on PATH.
2. Spawn the subagent (Agent tool, `general-purpose`, `model` = haiku or sonnet). Prompt = the preamble
   in `evals/tasks/README.md` with `<dir>` = the absolute `$d`, then the task brief verbatim, then the
   contents of `evals/FRICTION_TEMPLATE.md`, then: "Put `<repo>/bin` first on PATH in every shell command
   (`export PATH=<repo>/bin:$PATH`)." Nothing else — no hints about commands.
3. When it returns: `go run ./evals/grade -fixtures evals/work/fixtures <task> $d` → pass/fail.
   Read `$d/FRICTION.md` and the subagent's transcript summary; count commands and errors from it.
4. Delete `$d` renders after recording results (disk guardrail).

The grader is itself tested: `make grade-selftest` (known-good outputs pass, deliberately broken ones fail
on the right check). Run it whenever the grader or fixtures change.

Record every round in `evals/results/<date>-<n>.md`: per task × model: pass/fail, commands run, errors
hit, guide lookups, wasted steps, friction items. Turn every friction item into a backlog item tagged
`[friction]` with the quote that shows it. Friction ranks above new features.

### What counts as friction
The agent: guessed a flag/field that didn't exist; had to read `--help` twice for one thing; parsed
output with a workaround; called something done that wasn't (grader caught it); couldn't tell what to
do next; hit an error that didn't say how to fix it; produced output > 100 lines unasked; needed a
command that didn't exist; used seconds where text anchors would have been safer and got it wrong.

## Golden (stop criteria)

All of these, verified in the same round:
1. Every SPEC v1 command implemented, in `guide`, with e2e coverage.
2. Every INVARIANT in LESSONS.md has a passing guard test; every DETECT has a fixture that triggers it.
3. T2: 100% task pass rate with sonnet, two consecutive rounds, with no friction item rated medium or
   higher. Haiku runs are informational (their friction still goes to the backlog), not a gate.
   (Decided by the user 2026-11-24.)
4. T3: the reviewer accepts the edits on all local real-footage excerpts; `check` passes on them.
5. `go install` + `cutroom doctor` works on a clean PATH; README quick start verified by a fresh agent.

When golden: write `RELEASE_NOTES.md`, stop the loop, and notify the user. Publishing is their call.

## Guardrails

- Never write inside any source media folder; real footage is read-only and stays out of git.
- Disk: keep `evals/work` and caches under 5 GB; delete renders after grading. Check `df` each
  iteration; stop and notify if free space < 20 GB.
- No pushes, no remote repos, no publishing, no accounts. Downloads only for declared dependencies
  (whisper models) from their official sources.
- If blocked on a human decision (naming, licensing, something only a human can judge), write it in
  `BACKLOG.md` under **Needs human**, skip to other work, and mention it in the next notification.
- If 3 iterations in a row fail to make progress on the same item, stop and notify.
