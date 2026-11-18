#!/usr/bin/env bash
# Scripted end-to-end runs on synthetic fixtures (LOOP.md tier T1). Asserts on outputs.
set -euo pipefail
cd "$(dirname "$0")/.."
FIX=evals/work/fixtures
W=$(mktemp -d evals/work/e2e.XXXX)
trap 'rm -rf "$W"' EXIT
pass=0; fail=0
ok()   { pass=$((pass+1)); echo "ok   $1"; }
bad()  { fail=$((fail+1)); echo "FAIL $1"; [ -n "${2:-}" ] && echo "$2" | head -20 | sed 's/^/     /'; }
expect() { # name, command..., checks via grep on output
  local name=$1 pattern=$2; shift 2
  local out; out=$("$@" 2>&1) || true
  if grep -qE -- "$pattern" <<<"$out"; then ok "$name"; else bad "$name" "$out"; fi
}

expect "doctor ok"            'all dependencies ok'                 cutroom doctor
expect "guide lists topics"   'Topics:'                             cutroom guide
expect "typo suggests"        'did you mean `cutroom guide`'        cutroom giude
expect "json envelope"        '"ok": true'                          cutroom guide --json

echo "e2e: $pass passed, $fail failed"
[ "$fail" -eq 0 ]
