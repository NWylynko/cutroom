#!/usr/bin/env bash
# Prepare T2 eval run directories (LOOP.md "Running a T2 task"): evals/setup-runs.sh <round> <task>...
# Creates evals/work/runs/<round>-<task>-<model> for haiku and sonnet, copies the task's fixtures and
# runs its setup there with bin/ on PATH. Refuses to reuse an existing run directory.
set -euo pipefail
cd "$(dirname "$0")/.."
R=$PWD
export PATH="$R/bin:$PATH"
round=$1; shift
for t in "$@"; do
  spec=evals/tasks/$t.json
  [ -f "$spec" ] || { echo "no task $t"; exit 1; }
  for m in haiku sonnet; do
    d="$R/evals/work/runs/$round-$t-$m"
    if [ -e "$d" ]; then echo "exists: $d (delete it first)"; exit 1; fi
    mkdir -p "$d"
    for f in $(jq -r '.fixtures[]' "$spec"); do cp "evals/work/fixtures/$f" "$d/"; done
    setup=$(jq -r '.setup // empty' "$spec")
    if [ -n "$setup" ]; then
      (cd "$d" && bash -c "$setup" >/dev/null 2>&1) && echo "setup ok  $t-$m" || { echo "SETUP FAIL $t-$m"; exit 1; }
    else
      echo "ready     $t-$m"
    fi
  done
done
