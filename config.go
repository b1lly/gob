package gob

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

// Config are all the basic configuration options
type Config struct {
	GoPath   string
	BuildDir string
	SrcDir   string

	BuildTypes    []string // File extensions that cause the app to rebuild
	TemplateTypes []string // File extensions that cause the templating engine to re-render
	IgnoreTypes   []string // File extensions to let the filewatcher ignore

	Stdout io.Writer
	Stderr io.Writer
}

// DefaultConfig will return all the default
// configuration options used for setting up our application
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

// GobFlags represents the options gob uses when building and watching
// the target package. These are specified in the CLI
type GobFlags struct {
	NoRunMode                    bool   `json:"noRunMode"`                   // Listen and hot compile code, but don't run the program
	WatchTemplates               bool   `json:"watchTemplates"`              // whether or not to watch templates and notify subscribed gob agents
	GobServerPort                string `json:"gobServerPort"`               // what port to run the GobServer on (where GobClients can register)
	WatchPkgDependencies         bool   `json:"watchPackageDependencies"`    // whether or not to watch dependencies of the target package
	DependencyCheckInterval      int    `json:"dependencyCheckInterval"`     // the interval to sue when monitoring dependencies
	RecursivelyWatchDependencies bool   `json:recursivelyWatchDependencies"` // whether or not to watch dependencies recursively
}

// WriteConfigToPackage writes a gob config file to the directory of the target package
func (gb *Gob) WriteConfigToPackage() {
	data, err := json.Marshal(&gb.FlagConfig)
	if err != nil {
		fmt.Printf("[gob] failed to save config file due to error: %v\n", err)
		return
	}

	buffer := &bytes.Buffer{}
	err = json.Indent(buffer, data, "", "  ")
	if err != nil {
		fmt.Printf("[gob] failed to save config file due to error: %v\n", err)
		return
	}

	err = ioutil.WriteFile(path.Join(gb.Config.SrcDir, gb.PackagePath, ".gob.json"),
		buffer.Bytes(),
		0644)
	if err != nil {
		fmt.Printf("[gob] failed to save config file due to error: %v\n", err)
	}
}

// LoadConfig loads the data from the gob config file into the passed GobFlags struct
func (gb *Gob) LoadConfig() {
	data, err := ioutil.ReadFile(path.Join(gb.Config.SrcDir, gb.PackagePath, ".gob.json"))
	if err != nil {
		fmt.Printf("[gob] failed to load config: %v\n", err)
		return
	}
	err = json.Unmarshal(data, &gb.FlagConfig)
	if err != nil {
		fmt.Printf("[gob] failed to load config: %v\n", err)
	}
}
