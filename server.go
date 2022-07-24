package main

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

type httpHandler struct {
	fileServer http.Handler
	user       string
	password   string
}

func (h *httpHandler) isAuth() bool {
	return h.user != ""
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer logInfo.Printf("%-5v\t%-12v\t%v", r.Method, time.Since(start), r.URL.String())

	if err := h.auth(w, r); err != nil {
		logError.Printf("error auth: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	h.fileServer.ServeHTTP(w, r)

}

func (h *httpHandler) auth(w http.ResponseWriter, r *http.Request) error {
	if !h.isAuth() {
		return nil
	}
	w.Header().Set("WWW-Authenticate", `Basic realm="auth"`)
	u, p, ok := r.BasicAuth()
	if !ok {
		return fmt.Errorf("basic auth required")
	}
	if u != h.user || p != h.password {
		return fmt.Errorf("basic auth failed")
	}
	return nil
}

func newServer(root, host, user, password string, port uint, timeout time.Duration) *http.Server {
	handler := &httpHandler{
		fileServer: http.FileServer(http.Dir(root)),
		user:       user,
		password:   password,
	}
	return &http.Server{
		Addr:        net.JoinHostPort(host, fmt.Sprintf("%d", port)),
		Handler:     handler,
		ReadTimeout: timeout,
	}
}
