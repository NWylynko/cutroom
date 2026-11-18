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
expect "look sheet"            'contact sheet of talk: 8 frames'     cutroom look 6-12 --count 8
expect "look frame rotated"    'frame 375 of talk-rot'               cutroom look 12.5 --source talk-rot
expect "look crop"             '_crop100-50-200-100.png'             cutroom look 3 --crop 100,50,200,100
expect "look past end → fix"   'is 30\.[0-9]+ s long'                cutroom look 400
expect "listen finds burst"    'sound with no transcript words 0:16\.9' cutroom listen 15.3-18.9
expect "listen heard"          'heard \(isolated clip\): .*by the way' cutroom listen 15.3-18.9
expect "listen too long → fix" 'at most 30 s'                        cutroom listen 0-29.9 --pad 1
# rotated source must come out upright: compare a frame with the clean source's frame
cutroom look 12.5 --source talk --json >/dev/null; cutroom look 12.5 --source talk-rot --json >/dev/null
d=$(ffmpeg -v error -i .cutroom/look/talk_12_500.png -i .cutroom/look/talk-rot_12_500.png -lavfi "[0][1]psnr" -f null - 2>&1; ffmpeg -i .cutroom/look/talk_12_500.png -i .cutroom/look/talk-rot_12_500.png -lavfi "[0:v]crop=iw:ih-40:0:0[a];[1:v]crop=iw:ih-40:0:0[b];[a][b]psnr" -f null - 2>&1 | grep -o 'average:[0-9.inf]*')
case "$d" in *average:inf*|*average:[3-9][0-9]*) ok "look shows rotated source upright";; *) bad "look shows rotated source upright" "$d";; esac
popd >/dev/null

# edit flow on a fresh project: reach the ground-truth edit of talk.mp4 by words
E="$W/edit"; mkdir -p "$E"; cp $FIX/talk.mp4 "$E/"; pushd "$E" >/dev/null
cutroom init talk.mp4 >/dev/null; cutroom transcribe --passes 1 >/dev/null
expect "timeline before edit → fix" 'make the first edit'             cutroom timeline
expect "tighten creates timeline"   'created timeline from talk'      cutroom edit tighten
expect "tighten flags retake"       'retake-kept'                     cutroom timeline
expect "ambiguous --from lists"     'occurs 2 times'                  cutroom edit cut --from "The first thing you need is" --to "thing."
expect "cut retake by words"        'removed\): "The first thing you need is, UM, the first thing\."' cutroom edit cut --from "The first thing you need is, um" --to "the first thing."
expect "join pause is real"         'join: .*\(pause 0\.[23][0-9] s\)'  cutroom edit cut --from "By the way" --to "keyboard." --dry-run
expect "dry run not saved"          'my cat'                          cutroom timeline
expect "cut aside"                  'removed\): "By the way'           cutroom edit cut --from "By the way" --to "keyboard."
expect "clean timeline"             'no problems found'               cutroom timeline
out=$(cutroom timeline --json)
dur=$(python3 -c 'import json,sys;print(json.load(sys.stdin)["result"]["duration"])' <<<"$out")
python3 -c "import sys;d=float('$dur');sys.exit(0 if 13.0<=d<=18.5 else 1)" && ok "edited duration $dur in ground-truth bounds" || bad "edited duration" "$dur"
if grep -qiE '"text": "[^"]*(UM|cat)' <<<"$out"; then bad "no retake/aside words left" "$out"; else ok "no retake/aside words left"; fi
expect "undo"                       'undid: edit cut'                 cutroom edit undo
expect "history"                    'edit tighten'                    cutroom edit history
expect "unknown op suggests"        'did you mean `cutroom edit cut`' cutroom edit cutt
popd >/dev/null
expect "no project → fix"      'run `cutroom init'                   cutroom inspect --project "$W/nothing-here"

echo "e2e: $pass passed, $fail failed"
[ "$fail" -eq 0 ]
