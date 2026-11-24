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
expect "vfr times on source clock" '^0:06\.3[0-2][0-9]–'             cutroom transcript talk-vfr
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
expect "tighten lists pauses"       'pauses \(source time\): 0:04\.[0-9]+ 2\.[0-9]+→0\.32 s' cutroom edit tighten --dry-run
expect "tighten names lost sound"   'would also remove the sound with no words at source 0:16\.9' cutroom edit tighten --dry-run
expect "tighten creates timeline"   'created timeline from talk'      cutroom edit tighten
expect "tighten flags retake"       'retake-kept'                     cutroom timeline
expect "ambiguous --from lists"     'occurs 2 times'                  cutroom edit cut --from "The first thing you need is" --to "thing."
fixcmd=$(cutroom timeline | grep -o 'keep the last take: `[^`]*`' | sed 's/^keep the last take: `//; s/`$//')
expect "retake warning has a cut command" '^cutroom edit cut --from' echo "$fixcmd"
expect "its cut removes the false start"  'removed\): "The first thing you need is, UM, the first thing\."' sh -c "$fixcmd"
cutroom edit undo >/dev/null
expect "cut retake by words"        'removed\): "The first thing you need is, UM, the first thing\."' cutroom edit cut --from "The first thing you need is, um" --to "the first thing."
expect "join pause is real"         'join: .*\(pause 0\.[23][0-9] s\)'  cutroom edit cut --from "By the way" --to "keyboard." --dry-run
expect "dry run not saved"          'my cat'                          cutroom timeline
expect "cut aside"                  'removed\): "By the way'           cutroom edit cut --from "By the way" --to "keyboard."
expect "clean timeline"             'no problems found'               cutroom timeline
out=$(cutroom timeline --json)
dur=$(python3 -c 'import json,sys;print(json.load(sys.stdin)["result"]["duration"])' <<<"$out")
python3 -c "import sys;d=float('$dur');sys.exit(0 if 13.0<=d<=18.5 else 1)" && ok "edited duration $dur in ground-truth bounds" || bad "edited duration" "$dur"
if grep -qiE '"text": "[^"]*(UM|cat)' <<<"$out"; then bad "no retake/aside words left" "$out"; else ok "no retake/aside words left"; fi
expect "render dry run"             'estimate: ~'                     cutroom render --dry-run
expect "render"                     'wrote out/final.mp4'             cutroom render
expect "render cached"              '1 of 1 chunk\(s\) cached'         cutroom render
expect "render preview 540p"        '\(preview, 960x540\)'             cutroom render --preview
popd >/dev/null
g=$(go run ./evals/grade -fixtures $FIX tighten "$E" 2>&1)
grep -q 'GRADE tighten: PASS' <<<"$g" && ok "independent grader passes the render" || bad "independent grader passes the render" "$(grep -E 'FAIL' <<<"$g")"
pushd "$E" >/dev/null
expect "look --cuts"                'cuts 1–4 of 4'                   cutroom look --cuts
expect "look --output marks cut"    'red bar = a cut'                 cutroom look --output 3-5 --count 8
expect "listen --output (render)"   'cuts in range: 0:03\.933'         cutroom listen --output 3-5
expect "look --file"                'of out/final.mp4 at 0:03\.900 \(file time\)' cutroom look 3.9 --file out/final.mp4
expect "look <file> <time>"         'of out/final.mp4 at 0:03\.900 \(file time\)' cutroom look out/final.mp4 3.9
expect "listen <file> needs range"  'cutroom listen out/final.mp4 3-5' cutroom listen out/final.mp4
expect "render --output works"     "out/x\.mp4"                       cutroom render --output out/x.mp4 --dry-run
expect "range says whole chunk"     'whole timeline \(it fits in one chunk\)' cutroom render --dry-run --range 0:02-0:04
expect "check passes a good render" '^PASS out/final.mp4'              cutroom check
# planted faults: each must be caught by the right check
M=out/final.mp4.cutroom.json
ffmpeg -v error -y -i out/final.mp4 -itsoffset 0.1 -i out/final.mp4 -map 0:v -map 1:a -c copy -t 13.767 out/late.mp4; cp $M out/late.mp4.cutroom.json
expect "check: audio 100 ms late"   'FAIL sync .*worst \+(9[0-9]|10[0-9]) ms'  cutroom check out/late.mp4
ffmpeg -v error -y -i out/final.mp4 -vf "trim=start_frame=3,setpts=PTS-STARTPTS,tpad=stop=3:stop_mode=clone" -c:v libx264 -crf 18 -c:a copy out/shift.mp4; cp $M out/shift.mp4.cutroom.json
expect "check: picture 3 frames off" 'FAIL picture .*about \+3 frames' cutroom check out/shift.mp4
cutroom edit cut 1.9-2.9 >/dev/null; cutroom render out/clip.mp4 >/dev/null 2>&1
expect "check: clipped word"        'WARN clipped-word .*inside the word' cutroom check out/clip.mp4
expect "timeline: broken sentence"  'broken-sentence'                 cutroom timeline
expect "transcript --file render"   'out/clip\.mp4 — [0-9]+ words'     cutroom transcript --file out/clip.mp4
expect "transcript <render path>"   'out/clip\.mp4 — [0-9]+ words'     cutroom transcript out/clip.mp4
expect "render as source → --file"  'is a render, not a source.*transcript --file out/clip\.mp4' cutroom sync out/clip.mp4
expect "check missing → lists renders" 'renders here: .*out/clip\.mp4' cutroom check out/nothere.mp4
cutroom edit undo >/dev/null
expect "check: stale timeline"      'WARN stale'                      cutroom check out/clip.mp4
expect "check exit code 4 on fail"  '^4$'                             sh -c 'cutroom check out/late.mp4 >/dev/null; echo $?'
expect "undo"                       'undid: edit cut'                 cutroom edit undo
expect "history"                    'edit tighten'                    cutroom edit history
expect "unknown op suggests"        'did you mean `cutroom edit cut`' cutroom edit cutt
popd >/dev/null
# separate recorder: the round-1 bug (hand-set audio rendered 1.7 s late) must now be refused
S="$W/sync"; mkdir -p "$S"; cp $FIX/talk.mp4 $FIX/recorder.wav "$S/"; pushd "$S" >/dev/null
cutroom init talk.mp4 --source mic=recorder.wav >/dev/null; cutroom transcribe talk --passes 1 >/dev/null; cutroom edit tighten >/dev/null
expect "edit audio refuses unsynced"  'cutroom sync mic --to talk'      cutroom edit audio mic
python3 -c "import json;t=json.load(open('timeline.json'));t['audio']='mic';json.dump(t,open('timeline.json','w'))"
expect "render refuses hand-set unsynced audio" 'audio-not-synced|have not been aligned' cutroom render
cutroom edit undo >/dev/null; cutroom edit tighten >/dev/null
expect "sync measures offset"         'mic started recording 1\.70[01] s before talk' cutroom sync mic
expect "edit audio after sync"        'takes its sound from mic'        cutroom edit audio mic
expect "timeline shows audio source"  'sound from mic'                  cutroom timeline
expect "render with recorder audio"   'wrote out/final.mp4'             cutroom render
expect "check passes synced render"   '^PASS'                           cutroom check
expect "status: done after check"     'done: out/final\.mp4 is current and passed check' cutroom status
expect "status: shows recorder sync"  'synced to talk \(\+1\.70[01] s'  cutroom status
expect "check: camera cross-check"    'ok   sync-camera .*worst [+-]?[0-9] ms' cutroom check
cp cutroom.json cutroom.json.bak
python3 -c "import json;p=json.load(open('cutroom.json'));p['sources']['mic']['sync']['offset']+=0.1;json.dump(p,open('cutroom.json','w'))"
cutroom render out/wrongsync.mp4 >/dev/null 2>&1
expect "check: wrong recorder offset caught" 'FAIL sync-camera .*worst -(9[0-9]|10[0-9]) ms' cutroom check out/wrongsync.mp4
mv cutroom.json.bak cutroom.json; rm -f out/wrongsync.mp4*
popd >/dev/null
g=$(go run ./evals/grade -fixtures $FIX recorder "$S" 2>&1)
grep -q 'GRADE recorder: PASS' <<<"$g" && ok "grader passes recorder task via sync" || bad "grader passes recorder task via sync" "$(grep FAIL <<<"$g")"
pushd "$S" >/dev/null
expect "dry run says would"           '^would cut .*DRY RUN|would cut'  cutroom edit cut --text "Then install Linux on it." --dry-run
expect "transcript range past end"    'not word numbers'                cutroom transcript talk --words --from 38 --to 63
expect "init --remove in use refused" 'used by the timeline'            cutroom init --remove mic
# the recorder file is replaced (0.5 s later start): found on real footage (T3), sync used the old file
expect "bare init: nothing changed"    'nothing changed on disk'         cutroom init
ffmpeg -v error -y -i recorder.wav -af adelay=500 -c:a pcm_s16le rec2.wav && mv rec2.wav recorder.wav
expect "status: changed on disk"       'mic .*CHANGED on disk'           cutroom status
expect "sync refuses changed source"   'run .cutroom init. to re-inspect'  cutroom sync mic
expect "transcribe refuses changed"    'stale-source'                    cutroom transcribe mic
expect "bare init resets sync"         'mic changed, so its alignment was reset' cutroom init
expect "sync re-measures new file"     'mic started recording 2\.(19[0-9]|20[0-9]) s before talk' cutroom sync mic
popd >/dev/null
# round-2 regressions
V="$W/vfrcheck"; mkdir -p "$V"; cp $FIX/talk-vfr.mp4 "$V/"; pushd "$V" >/dev/null
cutroom init talk-vfr.mp4 >/dev/null; cutroom transcribe --passes 1 >/dev/null
expect "dry run would create"        'would create timeline'           cutroom edit tighten --dry-run
cutroom edit tighten >/dev/null; cutroom edit cut --from "The first thing you need is, um" --to "the first thing." >/dev/null; cutroom render >/dev/null 2>&1
expect "check passes VFR render (no false FAIL)" '^PASS'               cutroom check
expect "picture measured on VFR"     'ok   picture     1[0-9] sampled'   cutroom check
expect "short flag maps to --output"  'did you mean --output'           cutroom render -o x.mp4
expect "short flag maps to long"     'did you mean --preview'          cutroom render -p
expect "resolved repeat marked"      'resolved: only one take'         cutroom transcript --suspects
popd >/dev/null
expect "no project → fix"      'run `cutroom init'                   cutroom inspect --project "$W/nothing-here"

# rotated phone footage: render says the rotation is baked in
mkdir -p "$W/rot" && cp $FIX/talk-rot.mov "$W/rot/" && pushd "$W/rot" >/dev/null
cutroom init talk-rot.mov >/dev/null
expect "status: untranscribed → transcribe" 'next: cutroom transcribe talk-rot' cutroom status
expect "status: no timeline yet"   'timeline: none yet'              cutroom status
cutroom edit cut 0-0.5 >/dev/null
expect "status: no render → preview" 'renders: none'                 cutroom status
expect "render reports baked rotation" 'rotation is baked into the pixels' cutroom render --preview
expect "status: preview not checked" 'out/preview\.mp4 +preview +current, not checked' cutroom status
popd >/dev/null

# iPhone-style HDR (HLG 10-bit, 180° rotation): tone-mapped to SDR everywhere, upright, no metadata [L23 L25]
mkdir -p "$W/hdr" && cp $FIX/talk-hdr.mov $FIX/talk.mp4 "$W/hdr/" && pushd "$W/hdr" >/dev/null
expect "inspect: hdr handled"      'handled hdr: HDR source \(HLG'     cutroom init talk-hdr.mov talk.mp4
cutroom look 12.5 --source talk-hdr >/dev/null; cutroom look 12.5 --source talk >/dev/null
ffmpeg -v error -y -i .cutroom/look/talk_12_500.png -vf hflip,vflip flipped.png
up=$(ffmpeg -i .cutroom/look/talk-hdr_12_500.png -i .cutroom/look/talk_12_500.png -lavfi "[0:v]crop=iw:ih-40:0:0[a];[1:v]crop=iw:ih-40:0:0[b];[a][b]psnr" -f null - 2>&1 | grep -o 'average:[0-9.]*' | cut -d: -f2)
dn=$(ffmpeg -i .cutroom/look/talk-hdr_12_500.png -i flipped.png -lavfi "[0:v]crop=iw:ih-40:0:0[a];[1:v]crop=iw:ih-40:0:0[b];[a][b]psnr" -f null - 2>&1 | grep -o 'average:[0-9.]*' | cut -d: -f2)
awk -v u="$up" -v d="$dn" 'BEGIN{exit !(u > d + 3)}' && ok "look: HDR source upright (psnr $up vs flipped $dn)" || bad "look: HDR source upright" "psnr upright $up, flipped $dn"
cutroom edit cut 0-0.5 --source talk-hdr >/dev/null 2>&1 || true
expect "render HDR: rotation baked"  'rotation is baked into the pixels' cutroom render
expect "check HDR render: SDR, no rotation" 'format +1280x720, display rotation 0°, colour bt709' cutroom check
expect "check HDR render: picture"   'ok   picture'                      cutroom check
popd >/dev/null

echo "e2e: $pass passed, $fail failed"
[ "$fail" -eq 0 ]
