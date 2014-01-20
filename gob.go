package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	fswatch "github.com/andreaskoch/go-fswatch"
)

type Gob struct {
	Cmd    *exec.Cmd
	Config *Config

	// The full source path of the file to build
	InputPath string

	// The absolute path of the user input
	Dir string

	// The file name of the user input
	Filename string

	// The path to the binary file (e.g. output of build)
	Binary string
}

// Builder configuration options
type Config struct {
	GoPath   string
	BuildDir string
	SrcDir   string

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
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
	}
}

// Simple print method to prefix all output with "[gob]"
func (g *Gob) Print(msg string) {
	fmt.Println("[gob] " + msg)
}

// Simple argument validation
// to short circuit compilation errors
func (g *Gob) IsValidSrc() bool {
	// Make sure the user provided enough args to cmd line
	if len(os.Args) < 2 {
		g.Print("Please provide a valid source file to run.")
		return false
	}

	g.InputPath = os.Args[1]
	g.Dir, g.Filename = filepath.Split(g.InputPath)
	ext := filepath.Ext(g.Filename)
	g.Binary = g.Config.BuildDir + "/" + g.Filename[0:len(g.Filename)-len(ext)]

	// Make sure the file we're trying to build exists
	_, err := os.Stat(g.InputPath)
	if os.IsNotExist(err) || filepath.Ext(g.InputPath) != ".go" {
		g.Print("Please provide a valid source file to run.")
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
	cmd := exec.Command("go", "build", "-o", g.Binary, g.InputPath)
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
	cmd := exec.Command(g.Binary)

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

	appWatcher := fswatch.NewFolderWatcher(g.Config.SrcDir, true, skipNoFile).Start()

	for appWatcher.IsRunning() {
		select {
		case <-appWatcher.Change:
			g.Print("restarting application...")
			g.Cmd.Process.Kill()
			build := g.Build()
			if build {
				g.Run()
			}
		}
	}
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
