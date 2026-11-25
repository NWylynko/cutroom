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
expect "init: recorder needs sync" 'mic is audio only .*cutroom sync mic' cutroom init talk.mp4
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
expect "breath not a suspect"  'breath\(s\) not listed'              cutroom transcript talk --suspects
expect "breath shown with --all" 'breath +breath or mouth noise'      cutroom transcript talk --suspects --all
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
expect "range dry run is the range" 'range: output 0:02\.000–0:04\.000' cutroom render --dry-run --range 0:02-0:04
expect "range renders the range"   '2\.000 s, 60 frames'             cutroom render --preview --range 0:02-0:04
expect "check passes a range"      '^PASS out/preview-range\.mp4'      cutroom check out/preview-range.mp4
expect "check passes a good render" '^PASS out/final.mp4'              cutroom check
expect "captions written"          'wrote out/captions\.srt: [0-9]+ cues'  cutroom captions
expect "captions words match"      'ok   words +all [0-9]+ words of the edit' cutroom captions --dry-run
expect "captions --check passes"   'ok   seams'                      cutroom captions --check out/captions.srt
expect "captions vtt"              '^WEBVTT'                         sh -c 'cutroom captions --vtt >/dev/null && head -1 out/captions.vtt'
expect "status lists captions"     'captions: out/captions\.srt \([0-9]+ cues\) current' cutroom status
expect "captions not a render"     '^renders \([0-9]+\):'           sh -c 'cutroom status | grep -v "captions" | grep -c "out/captions" | grep -q "^0$" && cutroom status | grep "^renders"'
printf '1\n00:00:00,000 --> 00:00:02,000\nBy the way, my cat just walked across\n\n2\n00:00:02,100 --> 00:00:30,000\nthe keyboard.\n' > out/stale.srt
expect "captions --check fails stale" 'FAIL words .*in the file but not the edit' cutroom captions --check out/stale.srt
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
expect "render refuses hand-set unsynced audio" 'audio-not-synced|have not been aligned|is not aligned to' cutroom render
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
expect "edit: open warning in one line" 'still open \(unchanged by this edit\): retake-kept at output' cutroom edit cut --text "Then install Linux on it." --dry-run
expect "status: open warnings first"  'next: cutroom timeline +\([0-9]+ open warning' cutroom status
expect "status: points to history"    'cutroom edit history. lists them all' cutroom status
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

# punch-ins (L40): zoom a span, warn above 1.2× upscale, render, check through the crop
Z="$W/zoom"; mkdir -p "$Z"; cp $FIX/talk.mp4 "$Z/"; pushd "$Z" >/dev/null
cutroom init talk.mp4 >/dev/null; cutroom transcribe --passes 1 >/dev/null; cutroom edit tighten >/dev/null
expect "zoom a span"               'zoom 1\.15× on talk .*centred at 0\.50,0\.40' cutroom edit zoom 1.15 --from "Then install" --to "on it."
expect "zoom upscale warning"      'upscales the source 1\.50×.*L40'   cutroom edit zoom 1.5 --text "Your server is ready." --anchor 0.5,0.5
expect "timeline shows zoom"       '\[zoom 1\.50× at 0\.50,0\.50\]' cutroom timeline
expect "joins: zoom is not a cut"  'zoom 1\.00× → 1\.50× \(no cut'    cutroom timeline --joins
expect "look --output labels zoom" 'wrote .*output_'                  cutroom look --output 19.5
cutroom render >/dev/null 2>&1
expect "check through the zoom"    'ok   picture +[0-9]+ sampled frames match' cutroom check
python3 -c "import json;m=json.load(open('out/final.mp4.cutroom.json'));[g.pop('zoom',None) for g in m['segments']];json.dump(m,open('out/final.mp4.cutroom.json','w'))"
expect "check: framing mismatch"   'match no nearby source frame.*framing or zoom' cutroom check
expect "bad zoom factor"           'between 1 \(none\) and 4'          cutroom edit zoom 0.5 --text "Your server is ready."
popd >/dev/null

# multi-take: replace a line with another take's version, reorder, render, check per-segment sound
T="$W/takes"; mkdir -p "$T"; cp $FIX/talk.mp4 "$T/"; pushd "$T" >/dev/null
ffmpeg -v error -y -ss 13.5 -t 2.5 -i talk.mp4 -vf negate -c:v libx264 -preset veryfast -c:a aac take2.mp4
cutroom init talk.mp4 take2.mp4 >/dev/null; cutroom transcribe --passes 1 >/dev/null
expect "first edit uses longest take" 'created timeline from talk .*longest video' cutroom edit tighten
cutroom edit cut --from "Then install" --to "on it." >/dev/null
expect "add another take's line"   'added take2 .* after s3: "Then install Linux on it\."' cutroom edit add --source take2 --text "Then install Linux on it." --after s3
expect "same-mic take is a note"    'INFO\[audio-mixed\].*take2 use take2.s own sound next to talk; they match' cutroom timeline
ffmpeg -v error -y -i take2.mp4 -c:v copy -af volume=-12dB -c:a aac take3.mp4; cutroom init take3.mp4 >/dev/null; cutroom transcribe take3 --passes 1 >/dev/null
expect "quieter take warned"       'WARN\[audio-mixed\].*take3 .*differ \(speech .* dB.*louder/quieter' cutroom edit add --source take3 --text "Then install Linux on it." --dry-run
expect "recorder for wrong take refused" 'not been aligned to take2' cutroom edit audio talk --for take2 --dry-run
expect "move a segment"            'moved s1 to the end'               cutroom edit move s1 --before end
cutroom edit undo >/dev/null
cutroom render >/dev/null 2>&1
expect "check: each take's sound"  'ok   sync .*each segment.s own sound source \(take2, talk\)' cutroom check
expect "check: take2 picture"      'ok   picture'                     cutroom check
ffmpeg -v error -y -ss 10.5 -i out/final.mp4 -frames:v 1 o.png; ffmpeg -v error -y -ss 1.3 -i take2.mp4 -frames:v 1 t2.png
d=$(ffmpeg -i o.png -i t2.png -lavfi "[0:v]format=gray[a];[1:v]format=gray[b];[a][b]psnr" -f null - 2>&1 | grep -o 'average:[0-9.]*' | cut -d: -f2)
awk -v d="$d" 'BEGIN{exit !(d > 20)}' && ok "render shows take2 where added (psnr $d)" || bad "render shows take2 where added" "psnr $d"
popd >/dev/null

# pickup: replace a line with the re-recorded version in one command, in the original's place
K="$W/pickup"; mkdir -p "$K"; cp $FIX/talk.mp4 $FIX/pickup.mp4 "$K/"; pushd "$K" >/dev/null
cutroom init talk.mp4 pickup.mp4 >/dev/null; cutroom transcribe --passes 1 >/dev/null; cutroom edit tighten >/dev/null
expect "transcript defaults to the main take" 'also transcribed: pickup' cutroom transcript
expect "edit sub --help is focused"  'cutroom edit undo'              sh -c 'cutroom edit undo --help | head -1'
expect "replace with pickup"       'in its place: pickup .*"Then install Linux on it\."' cutroom edit replace --text "Then install Linux on it." --with pickup
expect "pickup sits where the line was" 'laptop\.\n?.*' sh -c 'cutroom timeline | grep -A1 "old laptop" | tail -1 | grep -q "pickup .*Then install Linux" && echo laptop.'
expect "look --output takes no value" '--output is a switch' cutroom look 2 --output 0:02
cutroom render >/dev/null 2>&1
expect "check pickup edit"         '^PASS'                            cutroom check
popd >/dev/null

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
expect "check HDR render: tone"      'ok   tone +HDR \(HLG\) source tone-mapped' cutroom check
expect "render says HDR tone-mapped"  'is HDR \(HLG\): tone-mapped to SDR bt709' cutroom render --preview
popd >/dev/null

# overlays: text and image cards anchored to words; rendered, looked at and checked frame-exactly
O="$W/overlay"; mkdir -p "$O"; cp $FIX/talk.mp4 "$O/"; pushd "$O" >/dev/null
cutroom init talk.mp4 >/dev/null; cutroom transcribe --passes 1 >/dev/null; cutroom edit tighten >/dev/null
cutroom edit cut --from "By the way" --to "keyboard." >/dev/null
ffmpeg -v error -y -f lavfi -i color=c=orange:s=240x120 -frames:v 1 logo.png
expect "overlay add text"          'added o1: text "Install Linux", lower-third, output' cutroom overlay add text "Install Linux" --sub "Ubuntu Server" --at "Then install" --duration 2
expect "overlay add image"         'added o2: image "logo.png", corner'  cutroom overlay add image logo.png --at "Finally" --to "update command." --fade 0
expect "overlay add needs --at"    'overlay add needs --at'              cutroom overlay add text "x"
expect "overlay bad layout"        'layouts: lower-third'                cutroom overlay add text "x" --at "Finally" --layout side
expect "overlay list"              'o2 +image "logo.png" +corner'        cutroom overlay list
expect "overlay set duration"      'set o1: --duration 1; now output .*\(1\.000 s\)' cutroom overlay set o1 --duration 1
expect "short card warned"         'overlay-short.*o1.*overlay set o1 --duration' cutroom timeline
expect "overlay set keeps the id"  'set o1: --duration 2\.5'               cutroom overlay set o1 --duration 2.5
expect "overlay set nothing"       'nothing to change'                   cutroom overlay set o1
expect "overlay sub-help"          'cutroom overlay add'                 sh -c 'cutroom overlay add --help | head -1'
t1=$(cutroom overlay list --json | python3 -c 'import json,sys;o=json.load(sys.stdin)["result"]["overlays"][0];print(round((o["output_start"]+o["output_end"])/2,2))')
expect "look --output shows overlay" 'with overlay\(s\) o1 drawn as rendered' cutroom look --output $t1
expect "render with overlays"      '^wrote out/final\.mp4'               cutroom render
expect "check: overlay text"       'ok   overlay +o1 "Install Linux" visible over frames .*gone on the frame before and after' cutroom check
expect "check: overlay image"      'ok   overlay +o2 "logo.png" visible'  cutroom check
expect "check: picture around cards" 'ok   picture'                       cutroom check
# planted faults: the manifest says the card shows 0.5 s later than the render does / a render without it
python3 -c 'import json;m=json.load(open("out/final.mp4.cutroom.json"));m["overlays"][0]["start"]+=0.5;m["overlays"][0]["end"]+=0.5;json.dump(m,open("out/late.mp4.cutroom.json","w"))'; cp out/final.mp4 out/late.mp4
expect "check: overlay window wrong" 'FAIL overlay +o1 .*still on screen'  cutroom check out/late.mp4
cutroom overlay rm o1 >/dev/null; cutroom overlay rm o2 >/dev/null; cutroom render out/bare.mp4 >/dev/null 2>&1
python3 -c 'import json;m=json.load(open("out/final.mp4.cutroom.json"));m["output"]="out/bare.mp4";json.dump(m,open("out/bare.mp4.cutroom.json","w"))'
expect "check: overlay missing"    'FAIL overlay +o1 .*not visible'      cutroom check out/bare.mp4
cutroom edit undo >/dev/null; cutroom edit undo >/dev/null
expect "overlay back after undo"   'o1 +text'                            cutroom overlay list
expect "captions --burn"           'burning in captions: [0-9]+ groups of 1–3 words' cutroom captions --burn --output out/final.srt
expect "burn wrote the srt too"    '^1$'                                 sh -c 'head -1 out/final.srt' 
expect "captions --burn twice"     'already on \(o3\)'                 cutroom captions --burn
expect "look shows captions"       'o3 \(burned-in captions, \+cc\)'    cutroom look --output 1
expect "render with captions"      '^wrote out/final\.mp4'               cutroom render
expect "check: captions on screen" 'ok   captions +burned-in captions: [0-9]+ groups; [0-9]+ sampled on screen' cutroom check
cutroom overlay rm o3 >/dev/null; cutroom render out/nocc.mp4 >/dev/null 2>&1
python3 -c 'import json;m=json.load(open("out/final.mp4.cutroom.json"));m["output"]="out/nocc.mp4";json.dump(m,open("out/nocc.mp4.cutroom.json","w"))'
expect "check: captions missing"   'FAIL captions .*not visible'          cutroom check out/nocc.mp4
cutroom edit undo >/dev/null
cutroom edit cut --text "Then install Linux on it." >/dev/null
expect "orphaned overlay warned"   'overlay-orphaned.*o1'                cutroom timeline
expect "list says not shown"       'NOT SHOWN: its anchor "Then install"' cutroom overlay list
popd >/dev/null

# VFR source with an audio offset: render and check say how both were handled
V="$W/vfr"; mkdir -p "$V"; cp $FIX/talk-vfr.mp4 "$V/"; pushd "$V" >/dev/null
cutroom init talk-vfr.mp4 >/dev/null; cutroom transcribe --passes 1 >/dev/null; cutroom edit tighten >/dev/null
expect "render states VFR handled"  'variable frame rate \(11 gaps\).*constant 30 fps grid' cutroom render --preview
expect "render states audio offset" "talk-vfr's sound starts \+0\.064 s from its picture" cutroom render --preview
expect "check states the offset"    'audio starts at \+0\.064 s, and that offset is in the render' cutroom check out/preview.mp4
expect "check names the model"      're-transcribed with base\.en, the model the transcript used' cutroom check out/preview.mp4
popd >/dev/null

# deciding warnings: apply a suggested fix, or accept one with a reason
D="$W/decide"; mkdir -p "$D"; cp $FIX/talk.mp4 "$D/"; pushd "$D" >/dev/null
cutroom init talk.mp4 >/dev/null; cutroom transcribe --passes 1 >/dev/null; cutroom edit tighten >/dev/null
expect "timeline says how to decide" 'decide each WARN: .*--apply.*--accept' cutroom timeline
expect "status names the apply"     'next: cutroom timeline --apply retake-kept@[0-9.]+ +\(runs: cutroom edit cut' cutroom status
expect "accept needs a reason"      '--accept needs --reason'            cutroom timeline --accept retake-kept
expect "unknown code suggests"      'did you mean retake-kept'           cutroom timeline --apply retake
expect "accept a warning"           'accepted retake-kept .*0 still open' cutroom timeline --accept retake-kept --reason "both takes wanted"
expect "accepted is listed apart"   'accepted \(decided, not open\): retake-kept' cutroom timeline
expect "accepted survives an edit"  'accepted \(decided, not open\): retake-kept' sh -c 'cutroom edit cut --from "By the way" --to "keyboard." >/dev/null; cutroom timeline'
cutroom edit undo >/dev/null; cutroom edit undo >/dev/null
expect "apply runs the fix"         'applying the retake-kept fix: cutroom edit cut .*cut talk 0:06\.200' sh -c 'cutroom timeline --apply retake-kept | tr "\n" " "'
expect "after apply: no retake"     'no problems found'                  cutroom timeline
popd >/dev/null

echo "e2e: $pass passed, $fail failed"
[ "$fail" -eq 0 ]
