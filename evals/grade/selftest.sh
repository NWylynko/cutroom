#!/usr/bin/env bash
# Prove the grader: known-good outputs must PASS, deliberately broken ones must FAIL on the right check.
set -uo pipefail
cd "$(dirname "$0")/../.."
FIX=evals/work/fixtures
W=$(mktemp -d evals/work/selftest.XXXX)
trap 'rm -rf "$W"' EXIT
go build -o "$W/grade" ./evals/grade || exit 1

# Ground-truth cut from the manifest: keep/retake-good lines, pad 0.10 before / 0.22 after.
segs=$(python3 - <<'EOF'
import json
m=json.load(open("evals/work/fixtures/talk.json"))
print(" ".join(f"{l['start']-0.10:.3f}:{l['end']+0.22:.3f}" for l in m["lines"] if l.get("tag") in ("keep","retake-good")))
EOF
)
cut() { # in out [extra video filter] [audio file] [audio delay ms]
  local in=$1 out=$2 vx=${3:-null} ain=${4:-$1} adelay=${5:-0} fc="" i=0 cat=""
  for s in $segs; do
    a=${s%:*}; b=${s#*:}
    fc+="[0:v]trim=start=$a:end=$b,setpts=PTS-STARTPTS[v$i];[1:a]atrim=start=$a:end=$b,asetpts=PTS-STARTPTS[a$i];"
    cat+="[v$i][a$i]"; i=$((i+1))
  done
  fc+="${cat}concat=n=$i:v=1:a=1[v][a];[v]$vx[vo];[a]adelay=$adelay:all=1[ao]"
  mkdir -p "$(dirname "$out")"
  ffmpeg -v error -y -i "$in" -i "$ain" -filter_complex "$fc" -map "[vo]" -map "[ao]" -c:v libx264 -preset veryfast -c:a aac -ar 48000 -shortest "$out"
}

pass=0; fail=0
want() { # expect(PASS|FAIL) task dir [check-that-must-fail]
  local exp=$1 task=$2 dir=$3 chk=${4:-}
  local out; out=$("$W/grade" -fixtures $FIX "$task" "$dir" 2>&1); local code=$?
  local got=PASS; [ $code -ne 0 ] && got=FAIL
  if [ "$got" != "$exp" ]; then fail=$((fail+1)); echo "FAIL selftest: $task $dir expected $exp got $got"; echo "$out" | grep -v '^JSON' | sed 's/^/     /'; return; fi
  if [ -n "$chk" ] && ! grep -qE "^FAIL $chk" <<<"$out"; then fail=$((fail+1)); echo "FAIL selftest: $task $dir should fail on '$chk'"; echo "$out" | grep -v '^JSON' | sed 's/^/     /'; return; fi
  pass=$((pass+1)); echo "ok   selftest: $task $(basename $dir) → $exp ${chk:+($chk)}"
}

cut $FIX/talk.mp4 "$W/good/out/final.mp4"
want PASS tighten "$W/good"

mkdir -p "$W/raw/out" && cp $FIX/talk.mp4 "$W/raw/out/final.mp4"
want FAIL tighten "$W/raw" "absent: my cat"

cut $FIX/talk.mp4 "$W/late/out/final.mp4" null $FIX/talk.mp4 200
want FAIL tighten "$W/late" "av_sync"

cut $FIX/talk.mp4 "$W/flip/out/final.mp4" "hflip,vflip"
want FAIL phone-footage "$W/flip" "frames_from_source/upright"

mkdir -p "$W/rotmeta/out" && ffmpeg -v error -y -display_rotation 180 -i "$W/flip/out/final.mp4" -c copy "$W/rotmeta/out/final.mp4"
want FAIL phone-footage "$W/rotmeta" "no_rotation_meta"

want PASS phone-footage "$W/good"
want FAIL recorder "$W/good" "audio_from=recorder.wav"

cut $FIX/talk.mp4 "$W/rec/out/final.mp4" null $FIX/recorder.wav
# recorder starts 1.7 s early, so the naive cut above is out of sync: grader must catch it
want FAIL recorder "$W/rec" "av_sync"

# recorder corrected for its known offset and drift → must PASS
ffmpeg -v error -y -i $FIX/recorder.wav -af "atrim=start=1.7,asetpts=PTS-STARTPTS,asetrate=48000/1.00005,aresample=48000" "$W/rec-fixed.wav"
cut $FIX/talk.mp4 "$W/recok/out/final.mp4" null "$W/rec-fixed.wav"
want PASS recorder "$W/recok"

# HDR (HLG, rotated): a correct tone map of the right cut passes; a plain matrix conversion fails on tone
ffmpeg -v error -y -hwaccel videotoolbox -hwaccel_output_format videotoolbox_vld -display_rotation 0 -i $FIX/talk-hdr.mov \
  -vf "scale_vt=color_matrix=bt709:color_primaries=bt709:color_transfer=bt709,hwdownload,format=p010le,hflip,vflip,format=yuv420p" \
  -c:v libx264 -preset veryfast -crf 18 -color_trc bt709 -color_primaries bt709 -colorspace bt709 -c:a copy "$W/hdr-sdr.mp4"
cut "$W/hdr-sdr.mp4" "$W/hdrgood/out/final.mp4"
want PASS hdr-phone "$W/hdrgood"
ffmpeg -v error -y -i $FIX/talk-hdr.mov -vf "colorspace=all=bt709:iall=bt2020:format=yuv420p" -c:v libx264 -preset veryfast \
  -color_trc bt709 -color_primaries bt709 -colorspace bt709 -c:a copy "$W/hdr-naive.mp4"
cut "$W/hdr-naive.mp4" "$W/hdrnaive/out/final.mp4"
want FAIL hdr-phone "$W/hdrnaive" "tone_mapped"
cut "$W/hdr-sdr.mp4" "$W/hdrlate/out/final.mp4" null "$W/hdr-sdr.mp4" 100
want FAIL hdr-phone "$W/hdrlate" "av_sync"

# pickup: the good cut with the pickup's line in place of the original passes; ignoring the pickup fails
fc=$(python3 - <<'PY'
import json
m=json.load(open("evals/work/fixtures/talk.json")); pk=json.load(open("evals/work/fixtures/pickup.json"))
parts=[];cat="";i=0
for l in m["lines"]:
    if l.get("tag") not in ("keep","retake-good"): continue
    if l["say"].startswith("Then install Linux"):
        a=l["start"]-0.10-pk["facts"]["talk_start"]; b=l["end"]+0.22-pk["facts"]["talk_start"]; src=1
    else:
        a=l["start"]-0.10; b=l["end"]+0.22; src=0
    parts.append(f"[{src}:v]trim=start={a:.3f}:end={b:.3f},setpts=PTS-STARTPTS[v{i}];[{src}:a]atrim=start={a:.3f}:end={b:.3f},asetpts=PTS-STARTPTS[a{i}];")
    cat+=f"[v{i}][a{i}]"; i+=1
print("".join(parts)+cat+f"concat=n={i}:v=1:a=1[v][a]")
PY
)
mkdir -p "$W/pick/out" && ffmpeg -v error -y -i $FIX/talk.mp4 -i $FIX/pickup.mp4 -filter_complex "$fc" -map "[v]" -map "[a]" -c:v libx264 -preset veryfast -c:a aac -ar 48000 "$W/pick/out/final.mp4"
want PASS pickup "$W/pick"
want FAIL pickup "$W/good" "alt_used"

# cards and captions: drawn with drawbox (no text filters needed), timed from the ground-truth lines
eval "$(python3 - <<'PY'
import json
m=json.load(open("evals/work/fixtures/talk.json"))
keep=[l for l in m["lines"] if l.get("tag") in ("keep","retake-good")]
t=0.0; win=[]
for l in keep:
    a=t+0.10; b=a+(l["end"]-l["start"]); win.append((a,b,l["say"])); t+=(l["end"]+0.22)-(l["start"]-0.10)
def en(ws): return "+".join(f"between(t,{a:.3f},{b:.3f})" for a,b,_ in ws)
cards=[w for w in win if "home server" in w[2] or "install Linux" in w[2]]
print(f"CARDS='{en(cards)}'")
print(f"SPEECH='{en(win)}'")
srt=[]; k=1
for a,b,say in win:
    ws=say.replace(",","").split(); n=len(ws); step=(b-a)/((n+3)//4)
    for i in range(0,n,4):
        ca=a+step*(i//4); cb=min(b,ca+step)
        f=lambda x: f"{int(x//3600):02d}:{int(x%3600//60):02d}:{int(x%60):02d},{int(round(x%1*1000)):03d}"
        srt.append(f"{k}\n{f(ca)} --> {f(cb)}\n{' '.join(ws[i:i+4])}\n"); k+=1
open("/tmp/cutroom-selftest.srt","w").write("\n".join(srt))
PY
)"
cut $FIX/talk.mp4 "$W/titles/out/final.mp4" "drawbox=x=300:y=280:w=680:h=110:color=black@0.8:t=fill:enable='$CARDS'"
want PASS titles "$W/titles"
want FAIL titles "$W/good" "card during: show you how"
cut $FIX/talk.mp4 "$W/allcards/out/final.mp4" "drawbox=x=300:y=280:w=680:h=110:color=black@0.8:t=fill"
want FAIL titles "$W/allcards" "no card during"
cut $FIX/talk.mp4 "$W/social/out/final.mp4" "drawbox=x=360:y=600:w=560:h=70:color=black@0.7:t=fill:enable='$SPEECH'"
cp /tmp/cutroom-selftest.srt "$W/social/out/final.srt"
want PASS social "$W/social"
mkdir -p "$W/nocc/out"; cp "$W/good/out/final.mp4" /tmp/cutroom-selftest.srt "$W/nocc/out/"; mv "$W/nocc/out/cutroom-selftest.srt" "$W/nocc/out/final.srt"
want FAIL social "$W/nocc" "burned_captions"
mkdir -p "$W/late-srt/out"; cp "$W/social/out/final.mp4" "$W/late-srt/out/"
python3 -c 'import re,sys
def sh(m):
    h,mi,s,ms=map(int,m.groups()); t=h*3600+mi*60+s+ms/1000+1.5
    return f"{int(t//3600):02d}:{int(t%3600//60):02d}:{int(t%60):02d},{int(round(t%1*1000)):03d}"
print(re.sub(r"(\d+):(\d\d):(\d\d),(\d\d\d)",sh,open("/tmp/cutroom-selftest.srt").read()))' > "$W/late-srt/out/final.srt"
want FAIL social "$W/late-srt" "srt: in sync"
rm -f /tmp/cutroom-selftest.srt

# short: a 9:16 crop of the steps with captions; broll: a full-screen cutaway during a line
eval "$(python3 - <<'PY'
import json
m=json.load(open("evals/work/fixtures/talk.json"))
keep=[l for l in m["lines"] if l.get("tag") in ("keep","retake-good")]
steps=[l for l in keep if l["say"].startswith(("The first thing you need is an","Then install","Finally"))]
print("SHORTSEGS='"+" ".join(f"{l['start']-0.06:.3f}:{l['end']+0.16:.3f}" for l in steps)+"'")
t=0.0; sw=[]
for l in steps:
    a=t+0.06; sw.append((a,a+l["end"]-l["start"])); t+=(l["end"]+0.16)-(l["start"]-0.06)
print("SHORTSPEECH='"+"+".join(f"between(t,{a:.3f},{b:.3f})" for a,b in sw)+"'")
t=0.0
for l in keep:
    a=t+0.10; b=a+(l["end"]-l["start"])
    if l["say"].startswith("Finally"): print(f"CUTA={a:.3f}; CUTB={b:.3f}")
    t+=(l["end"]+0.22)-(l["start"]-0.10)
PY
)"
full_segs=$segs; segs=$SHORTSEGS
cut $FIX/talk.mp4 "$W/short/out/short.mp4" "crop=ih*9/16:ih:'clip(0.6*iw-ow/2,0,iw-ow)':0,scale=1080:1920,drawbox=x=140:y=1300:w=800:h=140:color=black@0.75:t=fill:enable='$SHORTSPEECH'"
want PASS short "$W/short"
cut $FIX/talk.mp4 "$W/shortlow/out/short.mp4" "crop=ih*9/16:ih:'clip(0.6*iw-ow/2,0,iw-ow)':0,scale=1080:1920,drawbox=x=140:y=1700:w=800:h=140:color=black@0.75:t=fill:enable='$SHORTSPEECH'"
want FAIL short "$W/shortlow" "safe_bottom"
cut $FIX/talk.mp4 "$W/shortwide/out/short.mp4" "drawbox=x=300:y=520:w=680:h=60:color=black@0.75:t=fill:enable='$SHORTSPEECH'"
want FAIL short "$W/shortwide" "vertical_9x16"
segs=$full_segs
cut $FIX/talk.mp4 "$W/brollbase/out/final.mp4"
mkdir -p "$W/broll/out" "$W/brolltone/out"
ffmpeg -v error -y -i "$W/brollbase/out/final.mp4" -i $FIX/terminal.mp4 -filter_complex "[1:v]setpts=PTS-STARTPTS+$CUTA/TB[t];[0:v][t]overlay=0:0:eof_action=pass:enable='between(t,$CUTA,$CUTB)'[v]" -map "[v]" -map 0:a -c:v libx264 -preset veryfast -c:a copy "$W/broll/out/final.mp4"
want PASS broll "$W/broll"
ffmpeg -v error -y -i "$W/broll/out/final.mp4" -i $FIX/terminal.mp4 -filter_complex "[1:a]adelay=${CUTA}s:all=1[t];[0:a][t]amix=inputs=2:duration=first:normalize=0[a]" -map 0:v -map "[a]" -c:v copy -c:a aac "$W/brolltone/out/final.mp4"
want FAIL broll "$W/brolltone" "no_tone=6000"
want FAIL broll "$W/brollbase" "cutaway during"

echo "selftest: $pass passed, $fail failed"
[ $fail -eq 0 ]
