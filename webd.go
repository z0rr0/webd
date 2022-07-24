package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
)

var (
	logError = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	logInfo  = log.New(os.Stdout, "INFO: ", log.LstdFlags)

	// Tag is the git tag of the build.
	Tag string
)

func main() {
	var (
		host, root string
		port       uint
		version    bool
		timeout    time.Duration
	)
	flag.StringVar(&host, "host", "127.0.0.1", "host to listen on")
	flag.UintVar(&port, "port", 8080, "port to listen on")
	flag.StringVar(&root, "root", ".", "root directory to serve")
	flag.BoolVar(&version, "version", false, "show version")
	flag.DurationVar(&timeout, "timeout", time.Second*5, "timeout for requests")

	flag.Parse()
	if version {
		showVersion()
		return
	}

	server := newServer(root, host, port, timeout)
	idleConnClosed := make(chan struct{}) // to wait http server shutdown

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, os.Signal(syscall.SIGTERM), os.Signal(syscall.SIGQUIT))
		<-sigint

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logError.Printf("error shutting down server: %v", err)
		}
		close(idleConnClosed)
	}()

	logInfo.Printf("listening on %s, (timeout=%v, directory=%v)", server.Addr, timeout, root)
	if err := server.ListenAndServe(); (err != nil) && (err != http.ErrServerClosed) {
		logError.Printf("error starting server: %v", err)
	}
	<-idleConnClosed
	logInfo.Println("server successfully stopped")
}

func showVersion() {
	const name = "WebD"
	var keys = map[string]string{
		"vcs":          "",
		"vcs.revision": "",
		"vcs.time":     "",
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, bs := range bi.Settings {
			if _, exists := keys[bs.Key]; exists {
				keys[bs.Key] = bs.Value
			}
		}
	}
	fmt.Printf("%s %s\n%s:%s\n%s\n", name, Tag, keys["vcs"], keys["vcs.revision"], keys["vcs.time"])
}
