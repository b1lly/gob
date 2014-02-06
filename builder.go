package gob

import (
	"flag"
	"fmt"
	"github.com/b1lly/gob/agent"
	"github.com/howeyc/fsnotify"
	"go/build"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type Gob struct {
	GobServer  *agent.GobServer
	Cmd        *exec.Cmd
	CmdArgs    []string
	Config     *Config   // See "config.go"
	FlagConfig *GobFlags // See "config.go"

	InputPath string // The user input path of the file to build

	Dir      string // The absolute path of the input path
	Filename string // The file name of the user input

	PackagePath string   // The "package" path of the program we're building
	Binary      string   // The path to the binary file (e.g. output of build)
	PkgDeps     []string // The 3rd-party dependencies of the package we're building
}

// NewGob returns a new instance of Gob
// with the configuration settings applied
func NewGob(gobFlags *GobFlags) *Gob {
	return &Gob{
		Config:     DefaultConfig(),
		FlagConfig: gobFlags,
	}
}

// Simple print method to prefix all output with "[gob]"
func (g *Gob) Print(msg string) {
	fmt.Println("[gob] " + msg)
}

// IsValidSrc is a simple "input" validation helper
// that will try to determine the package that a particular path belongs too.
func (g *Gob) IsValidSrc() bool {
	var err error

	nonGobArgs := flag.Args()

	// Make sure the user provided enough args to cmd line
	if len(nonGobArgs) < 1 {
		g.Print("Please provide a valid source file to run.")
		return false
	}

	// Store the users `path/to/src` input
	g.InputPath = nonGobArgs[0]

	// Store the users `runtime` flags
	g.CmdArgs = nonGobArgs[1:]

	// Stores the absolute path of our file or package
	// Used to check to see if the package/file exists from root
	var absPath string

	// Handle filename and "package" as inputs
	g.Dir, g.Filename = filepath.Split(g.InputPath)

	pkgPath := filepath.Join(g.Config.SrcDir, g.InputPath)
	if _, err = os.Stat(pkgPath); !os.IsNotExist(err) {
		// we were passed the package name to build and run
		g.PackagePath = g.InputPath
		absPath = pkgPath
	} else if strings.Contains(g.Filename, ".") {
		g.Dir, err = filepath.Abs(g.Dir)
		if err != nil {
			fmt.Println(err)
			return false
		}

		g.PackagePath, err = filepath.Rel(g.Config.SrcDir, g.Dir)
		if err != nil {
			fmt.Println(err)
			return false
		}

		absPath = filepath.Join(g.Config.SrcDir, g.PackagePath, g.Filename)
	} else {
		g.Dir, err = filepath.Abs(g.InputPath)
		if err != nil {
			fmt.Println(err)
			return false
		}

		g.PackagePath, err = filepath.Rel(g.Config.SrcDir, g.Dir)
		if err != nil {
			fmt.Println(err)
			return false
		}

		absPath = filepath.Join(g.Config.SrcDir, g.PackagePath)
	}

	// Save our "tmp" binary as the last element of the input
	g.Binary = g.Config.BuildDir + "/" + filepath.Base(g.Dir)

	// Make sure the file/package we're trying to build exists
	_, err = os.Stat(absPath)
	if os.IsNotExist(err) {
		g.Print("Please provide a valid source file/package to run.")
		return false
	}

	return true
}

// Create a temp directory based on the configuration.
// This is where all the binaries are stored
func (g *Gob) Setup() {
	_, err := os.Stat(g.Config.BuildDir)
	if os.IsNotExist(err) {
		g.Print("creating temporary build directory... " + g.Config.BuildDir)
		os.MkdirAll(g.Config.BuildDir, 0777)
	}
}

// Build the source and save the binary
func (g *Gob) Build() bool {
	cmd := exec.Command("go", "build", "-o", g.Binary, g.PackagePath)
	cmd.Stdout = g.Config.Stdout
	cmd.Stderr = g.Config.Stderr

	g.Print("building src... " + g.InputPath)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		cmd.Process.Kill()
		return false
	}

	return true
}

// Run will attempt to run the binary that was previously compiled by Gob.
func (g *Gob) Run() {
	if g.FlagConfig.NoRunMode {
		g.Print("waiting for changes to recompile...")
		return
	}

	cmd := exec.Command(g.Binary, g.CmdArgs...)

	cmd.Stdout = g.Config.Stdout
	cmd.Stderr = g.Config.Stderr

	g.Print("starting application...")
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		cmd.Process.Kill()
		g.Cmd = nil
	} else {
		g.Cmd = cmd
	}
}

func (g *Gob) GetPkgDeps() {
	config := build.Default
	pkgsToCheck := []string{g.PackagePath}
	deps := make(map[string]bool)
	for i := 0; i < len(pkgsToCheck); i++ {
		pkg := pkgsToCheck[i]
		p, err := config.Import(pkg, path.Join(config.GOPATH, "src"), build.AllowBinary)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, dep := range p.Imports {
			// TODO(ttacon) Smarter filtering of 3rd party packages
			if ok, _ := deps[dep]; !ok && strings.Contains(dep, ".") {
				pkgsToCheck = append(pkgsToCheck, dep)
				deps[dep] = true
				g.PkgDeps = append(g.PkgDeps, dep)
			}
		}
	}
}

// Watch the filesystem for any changes
// and restart the application if detected
func (g *Gob) Watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	// Process events
	go func() {
		// Used to keep track of previous updates
		var lastUpdate time.Time

		// Used as a buffer to track multiple unique file changes
		fileChanges := map[string]bool{}

		for {
			select {
			case ev := <-watcher.Event:
				_, file := filepath.Split(ev.Name)

				// Short circuit if a file is renamed
				// or if the file is hidden
				if ev.IsRename() || !strings.HasPrefix(file, ".") {
					return
				}

				// Buffer up a bunch of files from our events
				// until the next update
				fileChanges[file] = true

				// Avoid excess rebuilds
				if time.Since(lastUpdate).Seconds() > 1 {
					lastUpdate = time.Now()

					app, views := g.getChangeType(fileChanges)

					// If they are application files, rebuild
					if app {
						g.Print("restarting application...")
						if g.Cmd != nil {
							g.Cmd.Process.Kill()
						}
						build := g.Build()
						if build {
							g.Run()
						}
					}

					// Talk to the Gob Agent when a view has been updated
					// and notify the subscribers
					if len(views) > 0 && g.GobServer != nil {
						g.GobServer.NotifySubscribers(views)
					}

					// Empty the files that were updated
					for k := range fileChanges {
						delete(fileChanges, k)
					}
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	// Our watchers require absolute paths for our dependencies
	toWatch := make([]string, len(g.PkgDeps)+1)
	for i, dep := range g.PkgDeps {
		toWatch[i] = path.Join(g.Config.SrcDir, dep)
	}

	// The last package to watch is the application package
	toWatch[len(toWatch)-1] = path.Join(g.Config.SrcDir, g.PackagePath)

	// Watch the packages
	for i, path := range toWatch {
		if i < len(toWatch)-1 {
			err = watcher.Watch(path)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			// If it's our application package, recursively
			// watch all sub directories
			f := func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// Ignore hidden directories
				if strings.HasPrefix(filepath.Base(path), ".") {
					return filepath.SkipDir
				}

				return watcher.Watch(path)
			}

			filepath.Walk(path, f)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	<-done
}

// Looks at the files that were modified and checks their extension
func (g *Gob) getChangeType(fileChanges map[string]bool) (app bool, views []string) {
	for filename := range fileChanges {
		fileExt := filepath.Ext(filename)

		// For now, just hardcode since we only have one of each type
		// TODO(billy) handle a slice better
		switch fileExt {
		case g.Config.BuildTypes[0]:
			app = true
		case g.Config.TemplateTypes[0]:
			views = append(views, filename)
		}
	}

	return
}
