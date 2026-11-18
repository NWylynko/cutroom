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

# init / inspect on fixtures
P="$W/proj"; mkdir -p "$P"; cp $FIX/talk.mp4 $FIX/talk-vfr.mp4 $FIX/talk-rot.mov $FIX/recorder.wav "$P/"
pushd "$P" >/dev/null
expect "init adds sources"     'talk-vfr .*\(added\)'               cutroom init talk.mp4 talk-vfr.mp4 talk-rot.mov --source mic=recorder.wav
expect "init is idempotent"    '^project updated'                    cutroom init talk.mp4
expect "no duplicate source"   '^\{|"unchanged": \[\s*$|"talk"'     cutroom init talk.mp4 --json
n=$(python3 -c 'import json;print(len(json.load(open("cutroom.json"))["sources"]))')
[ "$n" = 4 ] && ok "4 sources after re-init" || bad "4 sources after re-init" "got $n"
expect "inspect vfr"           'VARIABLE frame rate \(11 gaps'        cutroom inspect talk-vfr
expect "inspect rotation"      'rotated 180° by metadata'            cutroom inspect talk-rot
expect "inspect audio-only"    'audio  pcm_s16le 48000 Hz mono'      cutroom inspect mic
expect "unknown id suggests"   'did you mean source "talk"'          cutroom inspect tlk
expect "inspect a loose file"  'no problems found'                   cutroom inspect --fast talk.mp4
expect "transcribe"            'talk +[0-9]+ words, 7 phrases'        cutroom transcribe talk talk-vfr
expect "transcribe cached"     'cached'                              cutroom transcribe talk
expect "phrases"               '0:10\.[0-9]+–0:12\.[0-9]+  The first thing you need is an old laptop' cutroom transcript talk
expect "vfr times shifted"     '^0:06\.(39|40)[0-9]–'                cutroom transcript talk-vfr
expect "retake suspect"        'repeat .*first thing you need'       cutroom transcript talk --suspects
expect "sound suspect"         'sound .*no words'                    cutroom transcript talk --suspects
expect "search"                '2 match\(es\)'                       cutroom transcript --search "first thing you need"
expect "untranscribed → fix"   'cutroom transcribe mic'              cutroom transcript mic
popd >/dev/null
expect "no project → fix"      'run `cutroom init'                   cutroom inspect --project "$W/nothing-here"

echo "e2e: $pass passed, $fail failed"
[ "$fail" -eq 0 ]
