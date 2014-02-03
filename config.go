package gob

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/b1lly/gob/builder"
)

// GobFlags represents the options gob uses when building and watching
// the target package
type GobFlags struct {
	WatchTemplates               bool   `json:"watchTemplates"`              // whether or not to watch templates and notify subscribed gob agents
	GobServerPort                string `json:"gobServerPort"`               // what port to run the GobServer on (where GobClients can register)
	WatchPkgDependencies         bool   `json:"watchPackageDependencies"`    // whether or not to watch dependencies of the target package
	DependencyCheckInterval      int    `json:"dependencyCheckInterval"`     // the interval to sue when monitoring dependencies
	RecursivelyWatchDependencies bool   `json:recursivelyWatchDependencies"` // whether or not to watch dependencies recursively
}

// WriteConfigToPackage writes a gob config file to the directory of the target package
func WriteConfigToPackage(gob *builder.Gob, GobFlagsConfig GobFlags) {
	data, err := json.Marshal(&GobFlagsConfig)
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

	err = ioutil.WriteFile(path.Join(gob.Config.SrcDir, gob.PackagePath, ".gob.json"),
		buffer.Bytes(),
		0644)
	if err != nil {
		fmt.Printf("[gob] failed to save config file due to error: %v\n", err)
	}
}

// LoadConfig loads the data from the gob config file into the passed GobFlags struct
func LoadConfig(gob *builder.Gob, GobFlagsConfig GobFlags) {
	data, err := ioutil.ReadFile(path.Join(gob.Config.SrcDir, gob.PackagePath, ".gob.json"))
	if err != nil {
		fmt.Printf("[gob] failed to load config: %v\n", err)
		return
	}
	err = json.Unmarshal(data, GobFlagsConfig)
	if err != nil {
		fmt.Printf("[gob] failed to load config: %v\n", err)
	}
}
