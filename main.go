package main

import (
	"flag"

	"github.com/b1lly/gob/builder"
	"github.com/b1lly/gob/server"
)

var (
	watchTemplates = flag.Bool("agent", false, "watch templates and notify gob agent of changes")
	port           = flag.String("port", "9034", "port for gob to listen for subscribers on")
)

func main() {
	flag.Parse()

	gob := builder.NewGob()
	if !gob.IsValidSrc() {
		return
	}

	if *watchTemplates {
		builder.GobServer = server.NewGobServer(*port)
		go builder.GobServer.Start()
	}

	gob.Print("initializing program...")

	gob.Setup()
	build := gob.Build()
	if build {
		gob.Run()
	}

	gob.Watch()
}
