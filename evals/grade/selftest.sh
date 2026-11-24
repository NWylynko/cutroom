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

echo "selftest: $pass passed, $fail failed"
[ $fail -eq 0 ]
