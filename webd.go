package main

import (
	"flag"
	"fmt"
	"runtime/debug"
)

const (
	name = "WebD"
)

func main() {
	var (
		host    string
		port    int
		root    string
		version bool
	)
	flag.StringVar(&host, "host", "127.0.0.1", "host to listen on")
	flag.IntVar(&port, "port", 8080, "port to listen on")
	flag.StringVar(&root, "root", ".", "root directory to serve")
	flag.BoolVar(&version, "version", false, "show version")

	flag.Parse()
	if version {
		fmt.Println("webd version 0.0.1")
		return
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		fmt.Printf("%v\n", bi)
	}
}
