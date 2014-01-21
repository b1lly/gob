package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	fswatch "github.com/b1lly/go-fswatch"
)

type Gob struct {
	Cmd     *exec.Cmd
	CmdArgs []string
	Config  *Config

	InputPath string // The user input path of the file to build

	Dir      string // The absolute path of the input path
	Filename string // The file name of the user input

	PackagePath string // The "package" path of the program we're building
	Binary      string // The path to the binary file (e.g. output of build)
}

// Gob configuration options
type Config struct {
	GoPath   string
	BuildDir string
	SrcDir   string

	TemplateEngine interface{}

	BuildTypes    []string // File extensions that cause the app to rebuild
	TemplateTypes []string // File extensions that cause the templating engine to re-render
	IgnoreTypes   []string // File extensions to let the filewatcher ignore

	Stdout io.Writer
	Stderr io.Writer
}

func NewGob() *Gob {
	return &Gob{
		Config: DefaultConfig(),
	}
}

func DefaultConfig() *Config {
	goPath := os.Getenv("GOPATH")

	return &Config{
		GoPath:   goPath,
		BuildDir: goPath + "/gob/build",
		SrcDir:   goPath + "/src/",

		Stdout: os.Stdout,
		Stderr: os.Stderr,

		BuildTypes:    []string{".go"},
		TemplateTypes: []string{".soy"},
		IgnoreTypes:   []string{".js", ".css", ".scss", ".png", ".jpg", ".gif"},
	}
}

// Simple print method to prefix all output with "[gob]"
func (g *Gob) Print(msg string) {
	fmt.Println("[gob] " + msg)
}

// Simple argument validation
// to short circuit compilation errors
func (g *Gob) IsValidSrc() bool {
	var err error

	// Make sure the user provided enough args to cmd line
	if len(os.Args) < 2 {
		g.Print("Please provide a valid source file to run.")
		return false
	}

	// Store the users `path/to/src` input
	g.InputPath = os.Args[1]

	// Store the users `runtime` flags
	g.CmdArgs = os.Args[2:]

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

// Run our programs binary
func (g *Gob) Run() {
	cmd := exec.Command(g.Binary, g.CmdArgs...)

	cmd.Stdout = g.Config.Stdout
	cmd.Stderr = g.Config.Stderr

	g.Print("starting application...")
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		cmd.Process.Kill()
	}

	g.Cmd = cmd
}

// Watch the filesystem for any changes
// and restart the application if detected
func (g *Gob) Watch() {
	skipNoFile := func(path string) bool {
		return false
	}

	folder := fswatch.NewFolderWatcher(filepath.Join(g.Config.SrcDir, g.PackagePath), true, skipNoFile).Start()

	for folder.IsRunning() {
		select {
		case changes := <-folder.Change:
			go func() {
				app, views := g.getChangeType(changes)

				if app {
					g.Print("restarting application...")
					g.Cmd.Process.Kill()
					build := g.Build()
					if build {
						g.Run()
					}
				}

				// Special case for Soy template renderings
				// TODO(billy) Implement Template Hook
				for _, filename := range views {
					g.Print("re-rendering soy template... " + filename)
				}
			}()
		}
	}
}

// Looks at the files that were modified and checks their extension
func (g *Gob) getChangeType(changes *fswatch.FolderChange) (app bool, views []string) {
	var fileChanges []string
	fileChanges = append(fileChanges, changes.New()...)
	fileChanges = append(fileChanges, changes.Modified()...)
	fileChanges = append(fileChanges, changes.Moved()...)

	for _, filename := range fileChanges {
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

func main() {
	gob := NewGob()
	if !gob.IsValidSrc() {
		return
	}

	gob.Print("initializing program...")

	gob.Setup()
	build := gob.Build()
	if build {
		gob.Run()
	}
	gob.Watch()
}
