#!/bin/sh

path=$1
dir=$(dirname "$path")
file=$(basename "$path")

# command=$(
#     ffprobe -hide_banner "$path" 2>&1 | grep -o "Duration: [0-9:.]*"
# )

command1=$(
    cd "$dir" && \
    ffprobe -v error -sexagesimal -show_entries \
    "format=filename,duration,size,bit_rate : \
    stream=index,codec_name,codec_type,profile,width,height,channels,bit_rate : \
    stream_tags=BPS,DURATION" \
    -of json "$file"
)

echo "$command1"