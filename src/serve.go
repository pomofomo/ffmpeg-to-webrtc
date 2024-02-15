// copied from https://github.com/pion/example-webrtc-applications/blob/master/internal/signal/http.go
// with some modifications

// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

package main

import (
	"io"
	"net/http"
	"strconv"
)

type ServerConfig struct {
	Port int
	// This is where we get answers (gets one value for every offer)
	Answers chan string
	// This is where the server sends offers
	Offers chan string
}

func enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

// HTTPSDPServer starts a HTTP Server that consumes SDPs
func HTTPSDPServer(config ServerConfig) {
	port := config.Port
	if port == 0 {
		port = 9999
	}

	http.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		body, _ := io.ReadAll(r.Body)
		config.Offers <- string(body)
		answer := <-config.Answers
		w.Write([]byte(answer))
	})

	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(port), nil) // nolint:gosec
		if err != nil {
			panic(err) //nolint
		}
	}()
}
