package main

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

type httpHandler struct {
	fileServer http.Handler
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	h.fileServer.ServeHTTP(w, r)
	logInfo.Printf("%-5v\t%-12v\t%v", r.Method, time.Since(start), r.URL.String())
}

func newServer(root, host string, port uint, timeout time.Duration) *http.Server {
	return &http.Server{
		Addr:        net.JoinHostPort(host, fmt.Sprintf("%d", port)),
		Handler:     &httpHandler{fileServer: http.FileServer(http.Dir(root))},
		ReadTimeout: timeout,
	}
}
