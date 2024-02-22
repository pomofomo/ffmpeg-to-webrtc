# ffmpeg-to-webrtc

ffmpeg-to-webrtc demonstrates how to send video from ffmpeg to your browser using [pion](https://github.com/pion/webrtc).

This is forked from the awesome https://github.com/ashellunts/ffmpeg-to-webrtc with a few changes that I wanted:

1. Streams from local file, already in proper format, rather than running ffmpeg itself
2. Accepts ivf (vp8/9) instead of h264 formats (you can pre-format with ffmpeg outside of this)
3. Minimal signaling server to allow to pass the SDP configs over the wire rather than cutting and pasting
4. Allow multiple clients to subscribe to the stream at the same time

## How to run it

### Start the server 

Open a new shell and pass whatever video you want to show. (This may be a named pipe if you use output from elsewhere)

```bash
cd src
go run . <path/to/file.ivf>
```

### Connect the browser

Open [localhost:9999](http://localhost:9999) to view it

One you see "Ready" in logs, click on "Start Session".

Now you see the video

## Operating Systems

I only tested this on Ubuntu Linux. It *should* work on MacOS

## TODO

Make the signalling server more functional as only entry point

* (DONE) Serve the html files in the same `serve.go` file, so we have the same domain/port (no CORS problem) and remove running one more process
* (IMPORTANT) Check about needing https for other machines, proxy? 
    * [focused on quest](https://medium.com/@lazerwalker/how-to-easily-test-your-webvr-and-webxr-projects-locally-on-your-oculus-quest-eec26a03b7ee)
      * [adb reverse explained](https://blog.grio.com/2015/07/android-tip-adb-reverse.html)
      * [turn on dev mode](https://medium.com/sidequestvr/how-to-turn-on-developer-mode-for-the-quest-3-509244ccd386)
      * [adb wirelessly](https://tlmpartners.com/2022/11/14/oculus-quest-adb-android-debug-bridge/)
      * [Side Quest](https://sidequestvr.com/) - crazy kung fu?
    * [local certs](https://web.dev/articles/how-to-use-local-https)
    * [more local certs](https://stackoverflow.com/questions/63150089/is-there-a-way-to-have-https-for-local-network-webapps-without-buying-any-domain)
* Investigate proxy through (to other local http port) as simple net/http handler (or just use nginx? one more tech?)
