package main

import (
	"flag"
	"github.com/b1lly/gob"
	"github.com/b1lly/gob/agent"
)

var (
	// Normal gob flags
	noRunMode            = flag.Bool("norun", false, "hot compile code and build it, but don't run it")
	watchTemplates       = flag.Bool("agent", false, "watch templates and notify gob agent of changes")
	port                 = flag.String("port", "9034", "port for gob to listen for subscribers on")
	watchDeps            = flag.Bool("deps", false, "watch dependencies of your package for changes")
	depInterval          = flag.Int("intvl", 1, "time between dependency checks")
	recursivelyWatchDeps = flag.Bool("recWatch", true, "recursively watch dependencies")

	// flags pertaining to gob config file usage
	useConfig  = flag.Bool("useConfig", true, "use gob config file if it exists")
	saveConfig = flag.Bool("saveConfig", false, "save the current flags as th default config")
)

func main() {
	flag.Parse()

	// Creates a new instance of Gob including the
	// settings passed in from the user flags
	gb := gob.NewGob(&gob.GobFlags{
		NoRunMode:                    *noRunMode,
		WatchTemplates:               *watchTemplates,
		GobServerPort:                *port,
		WatchPkgDependencies:         *watchDeps,
		DependencyCheckInterval:      *depInterval,
		RecursivelyWatchDependencies: *recursivelyWatchDeps,
	})

	if !gb.IsValidSrc() {
		return
	}

	if *saveConfig {
		gb.WriteConfigToPackage()
	}

	if *useConfig {
		gb.LoadConfig()
	}

	// Decides whether or not to start up the GobServer
	// for the GobAgent client to connect to
	if *watchTemplates {
		// TODO (billy) Make a part of Gob struct
		gb.GobServer = agent.NewGobServer(*port)
		go gb.GobServer.Start()
	}

	// Notify the user that gob has started up
	gb.Print("initializing program...")

	// For performance purposes, we ignore dependencies by default
	if *watchDeps {
		gb.GetPkgDeps()
	}

	// Create the temporary build directory for our programs
	// and setup other basic things
	gb.Setup()

	// Build the application and attempt to start it up
	// The gob.Run() will short circuit if -norun is specified
	build := gb.Build()
	if build {
		gb.Run()
	}

	// Start watching the filesystem for updates
	gb.Watch()
}
