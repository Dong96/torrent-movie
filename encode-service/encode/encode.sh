#!/bin/sh

input=$1
audioIndex=$2
audioChannel=$3
audioBitrate=$4
dir=$(dirname "$input")
file=$(basename "$input")
#output="${file%.*}"
# echo $input
# echo $dir

channel=$(
    cd "$dir" && \
    ffprobe -v error -show_entries stream=channels -of default=noprint_wrappers=1:nokey=1 "$file"
)

echo "$channel"

command=$(
    cd "$dir" && \
    mkdir -p dash && \
    ffmpeg -progress dash/block.txt -i "$file" \
    -c:a aac -b:a $audioBitrate -ac $audioChannel \
    -c:v libx264 -preset veryfast -keyint_min 180 -g 180 \
    -filter_complex "[0:v]scale=-2:360[V0];[0:v]scale=-2:1080[V1]" \
    -map 0:${audioIndex} \
    -map [V0] -crf:v:0 40 \
    -map [V1] -crf:v:1 20 \
    -use_template 1 -use_timeline 1 -single_file 1 -b_strategy 0 \
    -adaptation_sets "id=0,streams=v id=1,streams=a" \
    -f dash dash/manifest.mpd 2> dash/out.txt
)

# command=$(
#     cd "$dir" && \
#     mkdir -p test && \ 
#     ffmpeg -progress test/block.txt -i "$file" \
#     -c:a aac -b:a 1M -ac $channel \
#     -c:v libx264 -preset veryfast -keyint_min 180 -g 180 \
#     -filter_complex "[0:v]scale=-2:360[V0];[0:v]scale=-2:1080[V1]" \
#     -map 0:a:0 test/audio.m4a\
#     -map [V0] -crf:v:0 40 test/video_360.mp4 \
#     -map [V1] -crf:v:1 20 test/video_1080.mp4 && \

#     # cd "${dir}/test" && \
#     # ffmpeg -f dash -i video_360.mp4 \
#     # -f dash -i video_1080.mp4 \
#     # -f dash -i audio.m4a \
#     # -c copy -map 0 -map 1 -map 2 \
#     # -f dash manifest.mpd > out.txt 2>&1

# )

echo $command