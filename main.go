package main

import (
	"github.com/b1lly/gob/agent"
	"github.com/b1lly/gob/builder"
)

func main() {
	gob := builder.NewGob()
	if !gob.IsValidSrc() {
		return
	}

	gob.Print("initializing program...")

	gob.Setup()
	build := gob.Build()
	if build {
		gob.Run()
	}
	ga := agent.NewGobAgent()
	go ga.NewServer()

	gob.Watch()
}
