package gob

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/b1lly/gob/agent"
	"github.com/howeyc/fsnotify"
	"go/build"
	"io/ioutil"
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
	World       []string // All packages described in GobMultiPackage build file
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
	fmt.Println("[gob]", msg)
}

func (g *Gob) PrintErr(err error) {
	fmt.Println("[gob]", err)
}

func (g *Gob) checkIsSource(srcDir, buildDir, path string) ([]string, bool) {
	var absPath string
	toReturn := make([]string, 4) // [0]=dir, [1]=filename, [2]=pkgPath [3]=binary
	// Handle filename and "package" as inputs
	toReturn[0], toReturn[1] = filepath.Split(path)

	pkgPath := filepath.Join(srcDir, path)
	if _, err := os.Stat(pkgPath); !os.IsNotExist(err) {
		// we were passed the package name to build and run
		toReturn[2] = path
		absPath = pkgPath
	} else if strings.Contains(toReturn[1], ".") {
		toReturn[0], err = filepath.Abs(toReturn[0])
		if err != nil {
			g.PrintErr(err)
			return nil, false
		}

		toReturn[2], err = filepath.Rel(srcDir, toReturn[0])
		if err != nil {
			g.PrintErr(err)
			return nil, false
		}

		absPath = filepath.Join(srcDir, toReturn[2], toReturn[1])
	} else {
		toReturn[0], err = filepath.Abs(path)
		if err != nil {
			g.PrintErr(err)
			return nil, false
		}

		toReturn[2], err = filepath.Rel(srcDir, toReturn[0])
		if err != nil {
			g.PrintErr(err)
			return nil, false
		}

		absPath = filepath.Join(srcDir, toReturn[2])
	}

	// Save our "tmp" binary as the last element of the input
	toReturn[3] = buildDir + "/" + filepath.Base(toReturn[2])

	// Make sure the file/package we're trying to build exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		g.Print("Please provide a valid source file/package to run.")
		return nil, false
	}

	return toReturn, true
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

	// Check if multi-package build file
	// currently must just be package def
	data, err := ioutil.ReadFile(filepath.Join(g.Config.SrcDir, g.InputPath))
	if err == nil {
		var packagesToBuild []string
		err = json.Unmarshal(data, &packagesToBuild)
		if err == nil {
			g.World = packagesToBuild
			// Make sure these are packages
			badPackages := false
			for _, pkgName := range g.World {
				if _, isValidSrc := g.checkIsSource(g.Config.SrcDir, g.Config.BuildDir, pkgName); !isValidSrc {
					fmt.Printf("[gob] '%s' is not a valid source file to build\n", pkgName)
					badPackages = true
				}
			}
			return !badPackages
		}
	}

	// Store the users `runtime` flags
	g.CmdArgs = nonGobArgs[1:]

	// Stores the absolute path of our file or package
	// Used to check to see if the package/file exists from root
	pkgValues, isValidSrc := g.checkIsSource(g.Config.SrcDir, g.Config.BuildDir, g.InputPath)
	g.Dir = pkgValues[0]
	g.Filename = pkgValues[1]
	g.PackagePath = pkgValues[2]
	g.Binary = pkgValues[3]
	return isValidSrc
}

// Create a temp directory based on the configuration.
// This is where all the binaries are stored
func (g *Gob) Setup() {
	if _, err := os.Stat(g.Config.BuildDir); os.IsNotExist(err) {
		g.Print("creating temporary build directory... " + g.Config.BuildDir)
		os.MkdirAll(g.Config.BuildDir, 0777)
	}
}

// Build the source and save the binary
func (g *Gob) Build() bool {
	var pkgs []string
	if len(g.World) == 0 {
		pkgs = []string{g.PackagePath}
	} else {
		pkgs = g.World
	}
	// keep track of which package build failed
	buildSucceeded := true
	for _, pkg := range pkgs {
		binaryName := filepath.Base(pkg)
		cmd := exec.Command("go", "build", "-o", filepath.Join(g.Config.BuildDir, binaryName), pkg)
		cmd.Stdout = g.Config.Stdout
		cmd.Stderr = g.Config.Stderr

		g.Print("building src... " + pkg)
		if err := cmd.Run(); err != nil {
			g.PrintErr(err)
			cmd.Process.Kill()
			buildSucceeded = false
		}
	}
	return buildSucceeded
}

// Run will attempt to run the binary that was previously compiled by Gob.
func (g *Gob) Run() {
	if g.FlagConfig.NoRunMode {
		g.Print("waiting for changes to recompile...")
		return
	}

	if len(g.World) == 0 {
		cmd := exec.Command(g.Binary, g.CmdArgs...)
		cmd.Stdout = g.Config.Stdout
		cmd.Stderr = g.Config.Stderr

		g.Print("starting application...")
		if err := cmd.Start(); err != nil {
			g.PrintErr(err)
			cmd.Process.Kill()
			g.Cmd = nil
		} else {
			g.Cmd = cmd
		}
	} else {
		for _, pkgName := range g.World {
			binaryName := filepath.Base(pkgName)
			cmd := exec.Command(filepath.Join(g.Config.BuildDir, binaryName))
			cmd.Stdout = g.Config.Stdout
			cmd.Stderr = g.Config.Stderr

			// pair pkgnames with cmd
			g.Print("starting " + pkgName + "[" + binaryName + "]...")
			if err := cmd.Start(); err != nil {
				g.PrintErr(err)
				cmd.Process.Kill()
				g.Cmd = nil
			} else {
				g.Cmd = cmd
			}
		}
	}
}

// GetValidPkgRoots returns a list of all the root package
// paths that are not part of the GO standard library
func (g *Gob) GetValidPkgRoots() map[string]bool {
	pkgRoots := make(map[string]bool)

	srcPaths, err := ioutil.ReadDir(g.Config.SrcDir)
	if err != nil {
		g.PrintErr(err)
	}

	for i := 0; i < len(srcPaths); i++ {
		if srcPaths[i].IsDir() {
			pkgRoots[srcPaths[i].Name()] = true
		}
	}

	return pkgRoots
}

// GetPkgDeps will return all of the dependencies for the root packages
// that we're building and running
func (g *Gob) GetPkgDeps() {
	config := build.Default
	var pkgsToCheck []string

	// Used to compare packages against standard GO library
	validPkgsRoots := g.GetValidPkgRoots()

	// If check if we're building multiple packages or just
	// one package
	if len(g.World) == 0 {
		pkgsToCheck = []string{g.PackagePath}
	} else {
		pkgsToCheck = g.World
	}

	// Contains a list of all the dependencies
	deps := make(map[string]bool)

	// Iterate through our list of packages and look for dependencies
	for i := 0; i < len(pkgsToCheck); i++ {
		pkg := pkgsToCheck[i]

		p, err := config.Import(pkg, g.Config.SrcDir, build.AllowBinary)
		if err != nil {
			g.PrintErr(err)
			return
		}

		// Iterate through all the dependencies
		// and filter out GO standard packages
		for _, dep := range p.Imports {
			// Avoid extra work
			if ok, _ := deps[dep]; !ok {
				// Grab the base path of the dependency
				// and match it against our valid package root paths
				if ok, _ := validPkgsRoots[strings.Split(dep, "/")[0]]; ok {
					pkgsToCheck = append(pkgsToCheck, dep)
					deps[dep] = true
					g.PkgDeps = append(g.PkgDeps, dep)
				}
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
				// Short circuit if a file is renamed
				if ev.IsRename() {
					break
				}

				_, file := filepath.Split(ev.Name)

				// Buffer up a bunch of files from our events
				// until the next update
				fileChanges[file] = true

				// Avoid excess rebuilds
				if time.Since(lastUpdate).Seconds() > 1 {
					lastUpdate = time.Now()

					app, views := g.getChangeType(fileChanges)

					// Ignore hidden files
					// TODO(billy) Figure out why this prevents duplicate events
					if !strings.HasPrefix(file, ".") {
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

	// Create watchers in each of the packages
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
