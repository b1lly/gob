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
	goCmd     *exec.Cmd
	goPath    string
	goSrcPath string
	path      string
	packages  []string
)

func main() {
	if !IsValidSrc() {
		return
	}

	// Run go build in separate process and pipe output to terminal
	StartApp()

	// For now, just listen for any changes from root
	goPath = os.Getenv("GOPATH")
	goSrcPath = goPath + "/src/"
	WatchForUpdate()
}

// Simple argument validation
// to short circuit compilation errors
func IsValidSrc() bool {
	// Make sure the user provided enough args to cmd line
	if len(os.Args) < 2 {
		fmt.Println("[gob] Please provide a valid source file to run.")
		return false
	}

	path = os.Args[1]

	// Make sure the file we're trying to build exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) || filepath.Ext(path) != ".go" {
		fmt.Println("[gob] Please provide a valid source file to run.")
		return false
	}

	return true
}

// Build the application in a separate process
// and pipe the output to the terminal
func StartApp() {
	goCmd = exec.Command("go", "run", path)

	goCmd.Stdout = os.Stdout
	goCmd.Stderr = os.Stderr

	fmt.Println("[gob] building src... " + path)
	if err := goCmd.Start(); err != nil {
		log.Fatal(err)
	}
}

// Watch the filesystem for any changes
// and restart the application if detected
func WatchForUpdate() {
	skipNoFile := func(path string) bool {
		return false
	}

	appWatcher := fswatch.NewFolderWatcher(goSrcPath, true, skipNoFile).Start()

	for appWatcher.IsRunning() {
		select {
		case <-appWatcher.Change:
			fmt.Println("[gob] restarting application...")
			goCmd.Process.Kill()
			StartApp()
		}
	}
}
