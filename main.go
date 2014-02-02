package main

import (
	"flag"
	"github.com/b1lly/gob/builder"

	"github.com/b1lly/gob/server"
)

var (
	// Normal gob flags
	watchTemplates       = flag.Bool("agent", false, "watch templates and notify gob agent of changes")
	port                 = flag.String("port", "9034", "port for gob to listen for subscribers on")
	watchDeps            = flag.Bool("deps", false, "watch dependencies of your package for changes")
	depInterval          = flag.Int("intvl", 1, "time between dependency checks")
	recursivelyWatchDeps = flag.Bool("recWatch", true, "recursively watch dependencies")

	// flags pertaining to gob config file usage
	useConfig  = flag.Bool("useConfig", true, "use gob config file if it exists")
	saveConfig = flag.Bool("saveConfig", false, "save the current flags as th default config")

	gobFlagsConfig gobFlags
)

func main() {
	flag.Parse()

	gobFlagsConfig = gobFlags{
		WatchTemplates:               *watchTemplates,
		GobServerPort:                *port,
		WatchPkgDependencies:         *watchDeps,
		DependencyCheckInterval:      *depInterval,
		RecursivelyWatchDependencies: *recursivelyWatchDeps,
	}

	gob := builder.NewGob()
	if !gob.IsValidSrc() {
		return
	}

	if *saveConfig {
		WriteConfigToPackage(gob)
	}

	if *useConfig {
		LoadConfig(gob)
	}

	if *watchTemplates {
		builder.GobServer = server.NewGobServer(*port)
		go builder.GobServer.Start()
	}

	gob.Print("initializing program...")

	if *watchDeps {
		gob.GetPkgDeps()
	}

	gob.Setup()
	build := gob.Build()
	if build {
		gob.Run()
	}

	gob.Watch()
}
