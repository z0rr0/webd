package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

var (
	// ErrBasicAuthRequired is returned when basic auth is required.
	ErrBasicAuthRequired = fmt.Errorf("basic auth required")
	// ErrBasicAuthFailed is returned when basic auth failed.
	ErrBasicAuthFailed = fmt.Errorf("basic auth failed")
)

// Params is server parameters.
type Params struct {
	Root     string
	Host     string
	Port     uint
	User     string
	Password string
	Timeout  time.Duration
	LogInfo  *log.Logger
	LogError *log.Logger
}

// handler is HTTP handler for the server.
type handler struct {
	fileServer http.Handler
	user       string
	password   string
	logInfo    *log.Logger
	logError   *log.Logger
}

// NoAuth checks applying basic auth to the handler.
// It allows empty password and returns true if authentication is not required.
func (h *handler) NoAuth() bool {
	return h.user == ""
}

func (h *handler) auth(w http.ResponseWriter, r *http.Request) error {
	if h.NoAuth() {
		return nil
	}
	w.Header().Set("WWW-Authenticate", `Basic realm="auth"`)
	u, p, ok := r.BasicAuth()
	if !ok {
		return ErrBasicAuthRequired
	}
	if u != h.user || p != h.password {
		return ErrBasicAuthFailed
	}
	return nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer h.logInfo.Printf("%-5v\t%-12v\t%v", r.Method, time.Since(start), r.URL.String())

	if err := h.auth(w, r); err != nil {
		h.logError.Printf("error auth: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	h.fileServer.ServeHTTP(w, r)
}

// New returns new server.
func New(p Params) *http.Server {
	h := &handler{
		fileServer: http.FileServer(http.Dir(p.Root)),
		user:       p.User,
		password:   p.Password,
		logInfo:    p.LogInfo,
		logError:   p.LogError,
	}
	return &http.Server{
		Addr:        net.JoinHostPort(p.Host, fmt.Sprintf("%d", p.Port)),
		Handler:     h,
		ReadTimeout: p.Timeout,
	}
}
