// copied from https://github.com/pion/example-webrtc-applications/blob/master/internal/signal/http.go
// with some modifications

// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/pion/webrtc/v3"
)

type ServerConfig struct {
	Port  int
	Offer webrtc.SessionDescription
}

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

// HTTPSDPServer starts a HTTP Server that consumes SDPs
func HTTPSDPServer(config ServerConfig) chan string {
	offerStr := Encode(config.Offer)
	port := config.Port
	if port == 0 {
		port = 9999
	}

	sdpChan := make(chan string)
	http.HandleFunc("/offer", func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		w.Write([]byte(offerStr))
	})
	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, "done")
		sdpChan <- string(body)
	})

	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(port), nil) // nolint:gosec
		if err != nil {
			panic(err) //nolint
		}
	}()

	return sdpChan
}
