package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	fswatch "github.com/andreaskoch/go-fswatch"
)

var (
	app    *exec.Cmd
	dir    string
	file   string
	binary string
	path   string

	GO_DIR     string = os.Getenv("GOPATH")
	GO_SRC_DIR string = GO_DIR + "/src/"
	BUILD_DIR  string = GO_DIR + "/gob/build"
)

func main() {
	if !isValidSrc() {
		return
	}

	gobPrint("initializing program...")

	setup()
	build()
	run()
	watch()
}

func gobPrint(msg string) {
	fmt.Println("[gob] " + msg)
}

// Simple argument validation
// to short circuit compilation errors
func isValidSrc() bool {
	// Make sure the user provided enough args to cmd line
	if len(os.Args) < 2 {
		gobPrint("Please provide a valid source file to run.")
		return false
	}

	path = os.Args[1]
	dir, file = filepath.Split(path)
	ext := filepath.Ext(file)
	binary = BUILD_DIR + "/" + file[0:len(file)-len(ext)]

	// Make sure the file we're trying to build exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) || filepath.Ext(path) != ".go" {
		gobPrint("Please provide a valid source file to run.")
		return false
	}

	return true
}

func setup() {
	_, err := os.Stat(BUILD_DIR)
	if os.IsNotExist(err) {
		gobPrint("creating temporary build directory... " + BUILD_DIR)
		os.MkdirAll(BUILD_DIR, 0777)
	}
}

// Build the application in a separate process
// and pipe the output to the terminal
func build() {
	cmd := exec.Command("go", "build", "-o", binary, path)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	gobPrint("building src... " + path)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

// Run our program binary
func run() {
	cmd := exec.Command(binary)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	gobPrint("starting application...")
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	app = cmd
}

// Watch the filesystem for any changes
// and restart the application if detected
func watch() {
	skipNoFile := func(path string) bool {
		return false
	}

	appWatcher := fswatch.NewFolderWatcher(GO_SRC_DIR, true, skipNoFile).Start()

	for appWatcher.IsRunning() {
		select {
		case <-appWatcher.Change:
			fmt.Println("[gob] restarting application...")
			app.Process.Kill()
			build()
			run()
		}
	}
}
