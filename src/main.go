//go:build !js
// +build !js

package main

import (
	"context"
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
	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			fmt.Printf("cannot close peerConnection: %v\n", cErr)
		}
	}()

	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())

	// Create a video track
	videoTrack, videoTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video", "pion")
	if videoTrackErr != nil {
		panic(videoTrackErr)
	}

	rtpSender, videoTrackErr := peerConnection.AddTrack(videoTrack)
	if videoTrackErr != nil {
		panic(videoTrackErr)
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

	if len(os.Args) != 2 {
		fmt.Printf("Usage: ffmpeg-to-webrtc [file.ivf]")
	}
	dataPipe, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("Error opening %s: %s", os.Args[1], err)
	}

	go func() {
		ivf, ivfHeader, ivfErr := ivfreader.NewWith(dataPipe)
		_ = ivfHeader // maybe use later?
		if ivfErr != nil {
			panic(ivfErr)
		}

		// Wait for connection established
		<-iceConnectedCtx.Done()

		// Send our video file frame at a time. Pace our sending so we send it at the same speed it should be played back as.
		// This isn't required since the video is timestamped, but we will such much higher loss if we send all at once.
		//
		// It is important to use a time.Ticker instead of time.Sleep because
		// * avoids accumulating skew, just calling time.Sleep didn't compensate for the time spent parsing the data
		// * works around latency issues with Sleep (see https://github.com/golang/go/issues/44343)
		// spsAndPpsCache := []byte{}
		ticker := time.NewTicker(frameDuration)
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

			// nal.Data = append([]byte{0x00, 0x00, 0x00, 0x01}, nal.Data...)

			// if nal.UnitType == h264reader.NalUnitTypeSPS || nal.UnitType == h264reader.NalUnitTypePPS {
			// 	spsAndPpsCache = append(spsAndPpsCache, nal.Data...)
			// 	continue
			// } else if nal.UnitType == h264reader.NalUnitTypeCodedSliceIdr {
			// 	nal.Data = append(spsAndPpsCache, nal.Data...)
			// 	spsAndPpsCache = []byte{}
			// }

			if ivfErr = videoTrack.WriteSample(media.Sample{Data: frame, Duration: frameDuration}); ivfErr != nil {
				panic(ivfErr)
			}
		}
	}()

	// Handshake inspired by https://github.com/pion/example-webrtc-applications/blob/master/gstreamer-send-offer/main.go
	// Where the server makes the offer and the client gets it and provides and answer

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure. It may be reconnected using an ICE Restart.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			fmt.Println("Peer Connection has gone to failed exiting")
			os.Exit(0)
		}
	})

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	if err = peerConnection.SetLocalDescription(offer); err != nil {
		panic(err)
	}

	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// show the offer
	fmt.Println(offer.Type)
	fmt.Println(offer.SDP)

	// At this point we are "ready". Now we publish this offer on our local server
	// And wait for a response
	sdpChan := HTTPSDPServer(ServerConfig{
		Offer: offer,
		Port:  9999,
	})
	fmt.Println("Serving on :9999 waiting for connection...")
	answer := webrtc.SessionDescription{}
	Decode(<-sdpChan, &answer)

	// show the offer
	fmt.Println("")
	fmt.Println(answer.Type)
	fmt.Println(answer.SDP)

	// Set the remote SessionDescription
	if err = peerConnection.SetRemoteDescription(answer); err != nil {
		panic(err)
	}

	// Block forever
	select {}
}
