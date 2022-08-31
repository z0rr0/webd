package server

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"path"
	"strings"
	"time"
)

const maxNameLength = 32

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

type fileItem struct {
	name string
	url  string
}

// toHTTPError writes an error to the response.
func (h *handler) toHTTPError(w http.ResponseWriter, err error) {
	var (
		code = http.StatusInternalServerError
		msg  = "500 Internal Server Error"
	)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		code, msg = http.StatusNotFound, "404 page not found"
	case errors.Is(err, fs.ErrPermission):
		code, msg = http.StatusForbidden, "403 Forbidden"
	case errors.Is(err, ErrBasicAuthRequired):
		code, msg = http.StatusUnauthorized, "401 Unauthorized (auth required)"
	case errors.Is(err, ErrBasicAuthFailed):
		code, msg = http.StatusUnauthorized, "401 Unauthorized (auth failed)"
	}
	h.logError.Printf("error response [%d]: %v / %v", code, err, msg)
	http.Error(w, msg, code)
}

// localRedirect gives a Moved Permanently response.
// It is a copy of http.localRedirect
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}

// handler is HTTP handler for the server.
type handler struct {
	root       http.FileSystem
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

// ServeHTTP is HTTP handler for the server.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer h.logInfo.Printf("%-5v\t%-12v\t%v", r.Method, time.Since(start), r.URL.String())

	if err := h.auth(w, r); err != nil {
		h.toHTTPError(w, err)
		return
	}

	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	// TODO: replace after dev is ready
	// h.serveFile(w, r)  // custom implementation
	h.fileServer.ServeHTTP(w, r) // standard library implementation
}

// serveFile is URL HTTP handler.
func (h *handler) serveFile(w http.ResponseWriter, r *http.Request) {
	var name = path.Clean(r.URL.Path)

	f, err := h.root.Open(name)
	if err != nil {
		h.toHTTPError(w, err)
		return
	}
	defer func() {
		if errClose := f.Close(); errClose != nil {
			h.logError.Printf("error closing file '%s': %v", name, errClose)
		}
	}()
	d, err := f.Stat()
	if err != nil {
		h.toHTTPError(w, err)
		return
	}
	if d.IsDir() {
		url := r.URL.Path
		// redirect if the directory name doesn't end in a slash
		if url == "" || url[len(url)-1] != '/' {
			localRedirect(w, r, path.Base(url)+"/")
			return
		}
	}
}

func (h *handler) dirList(w http.ResponseWriter, r *http.Request, f http.File) {

}

// New returns new server.
func New(p Params) *http.Server {
	root := http.Dir(p.Root)
	h := &handler{
		root:       root,
		fileServer: http.FileServer(root),
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
