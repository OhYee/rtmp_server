ffmpeg -stream_loop -1 -i "../../tools/big_buck_bunny.mp4" -vcodec libx264 -acodec mp3 -ar 11025 -ab 16K -preset ultrafast -r 60 -f flv rtmp://127.0.0.1:19356/test/pc
