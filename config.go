package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/b1lly/gob/builder"
)

type gobFlags struct {
	WatchTemplates               bool   `json:"watchTemplates"`
	GobServerPort                string `json:"gobServerPort"`
	WatchPkgDependencies         bool   `json:"watchPackageDependencies"`
	DependencyCheckInterval      int    `json:"dependencyCheckInterval"`
	RecursivelyWatchDependencies bool   `json:recursivelyWatchDependencies"`
}

func WriteConfigToPackage(gob *builder.Gob) {
	data, err := json.Marshal(&gobFlagsConfig)
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

func LoadConfig(gob *builder.Gob) {
	data, err := ioutil.ReadFile(path.Join(gob.Config.SrcDir, gob.PackagePath, ".gob.json"))
	if err != nil {
		fmt.Printf("[gob] failed to load config: %v\n", err)
		return
	}
	err = json.Unmarshal(data, &gobFlagsConfig)
	if err != nil {
		fmt.Printf("[gob] failed to load config: %v\n", err)
	}
}
