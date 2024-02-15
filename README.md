# ffmpeg-to-webrtc

ffmpeg-to-webrtc demonstrates how to send video from ffmpeg to your browser using [pion](https://github.com/pion/webrtc).

This is forked from the awesome https://github.com/ashellunts/ffmpeg-to-webrtc with a few changes that I wanted:

1. Streams from local file, already in proper format, rather than running ffmpeg itself
2. Accepts ivf (vp8/9) instead of h264 formats (you can pre-format with ffmpeg outside of this)
3. Minimal signaling server to allow to pass the SDP configs over the wire rather than cutting and pasting

## How to run it

### Open example web page

```bash
# go install github.com/go-serve/goserve@latest
go install github.com/philippgille/serve@latest
serve -d html -p 7777
```

Open [localhost:7777](http://localhost:7777)

### Copy browser's SDP

Open a new shell and:

```bash
cd src
go run . <path/to/file.ivf>
```

### Connect the browser

One you see "Ready" in logs, click on "Start Session".

Now you see the video

## Operating Systems

I only tested this on Ubuntu Linux. It *should* work on MacOS