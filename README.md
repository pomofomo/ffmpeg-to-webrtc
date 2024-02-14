# ffmpeg-to-webrtc

ffmpeg-to-webrtc demonstrates how to send video from ffmpeg to your browser using [pion](https://github.com/pion/webrtc).

This is forked from the awesome https://github.com/ashellunts/ffmpeg-to-webrtc with a few changes that I wanted:

1. Streams from local file, already in proper format, rather than running ffmpeg itself
2. Accepts ivf (vp8/9) instead of h264 formats (you can pre-format with ffmpeg outside of this)
3. Minimal signaling server to allow to pass the SDP configs over the wire rather than cutting and pasting

## How to run it

TODO: this is all obsolete

### Open example web page

```bash
# go install github.com/go-serve/goserve@latest
go install github.com/philippgille/serve@latest
serve -d html -p 7777
```

Open [localhost:7777/test.html](http://localhost:7777/test.html)

### Copy browser's SDP

In the website the top textarea is your browser's SDP, copy that to clipboard.

Open a new shell and:

```bash
cd src
# Paste the SDP inside vi
vi SDP.txt

go run . <path/to/file.ivf> < SDP.txt
```



### Start the server

the directory with the video file in it (we call it `stream.ivf` in this example)


#### Windows
1. `cd src`
1. Paste the SDP into a file `src/SDP.txt`.
2. Make sure ffmpeg in your PATH and golang is installed.
3. Run `go run . <ffmpeg command line options> - < SDP.txt`
4. Note dash after ffmpeg options. It makes ffmpeg to write output to stdout. The app will read h264 stream from ffmpeg stdout.
5. ffmpeg output format should be h264. Browsers don't support all h264 profiles so it may not always work. Here is an example of format that works: `-pix_fmt yuv420p -c:v libx264 -bsf:v h264_mp4toannexb -b:v 2M -max_delay 0 -bf 0 -f h264`.

### Put SDP from ffmpeg-to-webrtc into your browser
When you see SDP in base64 format printed it means that SDP is already in clipboard. So you can go to jsfiddle page and paste that into Application SDP text area.

### Hit 'Start Session' in jsfiddle
A video should start playing in your browser below the input boxes.

## Examples (windows)
### Share camera stream
```go run . -rtbufsize 100M -f dshow -i video="PUT_DEVICE_NAME" -pix_fmt yuv420p -c:v libx264 -bsf:v h264_mp4toannexb -b:v 2M -max_delay 0 -bf 0 -f h264 - < SDP```. 
There is a delay of several seconds. Should be possible to fix it with better ffmpeg configuration.

To check list of devices: `ffmpeg -list_devices true -f dshow -i dummy`.  
It is possible also to set a resolution and a format, for example `-pixel_format yuyv422 -s 640x480`.
Possible formats: `ffmpeg -list_options true -f dshow -i video=PUT_DEVICE_NAME`.
### Share screen or window
See `.bat` files in src folder

## Linux, macOS

Should work on other operating systems, though I haven't tried.
