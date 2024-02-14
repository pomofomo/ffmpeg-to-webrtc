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

Open [localhost:7777/test.html](http://localhost:7777/test.html)

### Copy browser's SDP

Open a new shell and:

```bash
cd src
go run . <path/to/file.ivf>
```

You can test this in another shell via `curl localhost:9999/offer`

### Connect the browser

Copy the string on terminal to the clipboard, and paste it in the website

Now you see the video

## Operating Systems

I only tested this on Ubuntu Linux. I *should* work on MacOS