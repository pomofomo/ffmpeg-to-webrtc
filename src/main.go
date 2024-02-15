//go:build !js
// +build !js

package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/ivfreader"
)

const (
	// Make this variable from args to match original
	frameDuration = time.Millisecond * 100
)

func main() { //nolint
	// Create a video track
	videoTrack, videoTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video", "pion")
	if videoTrackErr != nil {
		panic(videoTrackErr)
	}

	// serve video to the track
	if len(os.Args) != 2 {
		fmt.Printf("Usage: ffmpeg-to-webrtc [file.ivf]")
	}
	go serveVideo(os.Args[1], videoTrack)

	// Server waits for client to make offer and responds with an answer

	// At this point we are "ready". Now we publish this offer on our local server
	// And wait for a response
	config := ServerConfig{
		Port:    9999,
		Offers:  make(chan string),
		Answers: make(chan string),
	}
	HTTPSDPServer(config)
	fmt.Println("Serving on :9999 waiting for connection...")

	// perpare to handle all connections
	conns := 0
	peers := []*webrtc.PeerConnection{}
	defer func() {
		for i, p := range peers {
			if cErr := p.Close(); cErr != nil {
				fmt.Printf("cannot close peerConnection %d: %v\n", i, cErr)
			}
		}
	}()

	// loop forever adding new connections
	for {
		conns += 1
		peer, err := addConnection(&config, videoTrack, conns)
		if err != nil {
			panic(err)
		}
		peers = append(peers, peer)
	}
}

func addConnection(config *ServerConfig, videoTrack *webrtc.TrackLocalStaticSample, myConn int) (*webrtc.PeerConnection, error) {
	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	rtpSender, err := peerConnection.AddTrack(videoTrack)
	if err != nil {
		return nil, err
	}

	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection %d State has changed %s \n", myConn, connectionState.String())
	})

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection %d State has changed: %s\n", myConn, s.String())
	})

	// get offer from first client
	offer := webrtc.SessionDescription{}
	Decode(<-config.Offers, &offer)
	fmt.Println(offer.Type)

	// Set the remote SessionDescription
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		return nil, err
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// send answer to server
	// fmt.Println("\nReturning answer")
	fmt.Println(answer.Type)
	// fmt.Println(answer.SDP)
	sdp := Encode(answer)
	config.Answers <- sdp

	return peerConnection, nil
}

func serveVideo(filename string, videoTrack *webrtc.TrackLocalStaticSample) {
	dataPipe, err := os.Open(filename)
	if err != nil {
		panic(fmt.Errorf("opening %s: %s", filename, err))
	}

	ivf, ivfHeader, ivfErr := ivfreader.NewWith(dataPipe)
	_ = ivfHeader // maybe use later?
	if ivfErr != nil {
		panic(ivfErr)
	}

	// Send our video file frame at a time. Pace our sending so we send it at the same speed it should be played back as.
	// This isn't required since the video is timestamped, but we will such much higher loss if we send all at once.
	//
	// It is important to use a time.Ticker instead of time.Sleep because
	// * avoids accumulating skew, just calling time.Sleep didn't compensate for the time spent parsing the data
	// * works around latency issues with Sleep (see https://github.com/golang/go/issues/44343)
	// spsAndPpsCache := []byte{}
	ticker := time.NewTicker(frameDuration)
	frames := 0
	for ; true; <-ticker.C {
		frame, frameHeader, ivfErr := ivf.ParseNextFrame()
		_ = frameHeader
		if ivfErr == io.EOF || frame == nil {
			fmt.Printf("All video frames parsed and sent")
			os.Exit(0)
		}
		if ivfErr != nil {
			panic(ivfErr)
		}

		frames += 1
		if frames%10 == 0 {
			fmt.Printf("Frames: %d\n", frames)
		}

		if ivfErr = videoTrack.WriteSample(media.Sample{Data: frame, Duration: frameDuration}); ivfErr != nil {
			panic(ivfErr)
		}
	}
}
