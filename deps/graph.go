package main

import (
	"fmt"
	"go/build"
	"path/filepath"
	"strconv"
	"strings"
)

// TODO(billy) Remove when done with the interface. This is for testing purposes
func main() {
	Graph := NewGraph(&Graph{
		StdLib: false,
		SrcDir: "/src/",
		Pkgs:   []string{"yext/pages/storepages/storm", "yext/pages/storepages/admin"},
	})

	// Unique list of dependencies
	list := Graph.ListDeps()
	for i := range list {
		fmt.Println(list[i])
	}

	// Print our graph
	Graph.RootNode.print("\t")
}

// TODO(billy) Remove once we make this a real package and not a main
var stdlib = map[string]struct{}{
	"archive":   {},
	"bufio":     {},
	"bytes":     {},
	"compress":  {},
	"container": {},
	"crypto":    {},
	"database":  {},
	"debug":     {},
	"encoding":  {},
	"errors":    {},
	"expvar":    {},
	"flag":      {},
	"fmt":       {},
	"go":        {},
	"hash":      {},
	"html":      {},
	"image":     {},
	"index":     {},
	"io":        {},
	"log":       {},
	"math":      {},
	"mime":      {},
	"net":       {},
	"os":        {},
	"path":      {},
	"reflect":   {},
	"regexp":    {},
	"runtime":   {},
	"sort":      {},
	"strconv":   {},
	"strings":   {},
	"sync":      {},
	"syscall":   {},
	"testing":   {},
	"text":      {},
	"time":      {},
	"unicode":   {},
}

type Graph struct {
	StdLib bool     // the value 'false' will ignore stdlib imports
	SrcDir string   // the root src directory of all the packages
	Pkgs   []string // list of packages to use when building our depdendency tree

	TotalDeps int // total number of dependencies used across all packages

	RootNode *Node
	Nodes    map[string]*Node // map of all the dependencies across all our projects
}

// NewGraph is used to build out a dependency tree and provides uselful helpers
// for traversing and figuring based on a list of packages
func NewGraph(d *Graph) *Graph {
	Graph := Graph{
		StdLib: d.StdLib,
		SrcDir: d.SrcDir,
		Pkgs:   d.Pkgs,

		RootNode: &Node{
			Path: d.SrcDir,
		},
		Nodes: make(map[string]*Node),
	}

	Graph.buildTree()
	return &Graph
}

// ListGraph returns a unique list of dependencies based on the dependency tree
func (d *Graph) ListDeps() (deps []string) {
	for n := range d.Nodes {
		deps = append(deps, d.Nodes[n].Path)
	}

	return
}

// Node contains a data for a given dependency and it's relationship to others.
// It also contains some extra helpful data for traversing
type Node struct {
	Parent   *Node
	Children []*Node

	Path          string
	TotalChildren int
}

func (n *Node) addParent(node *Node) {
	n.Parent = node
	node.addChild(n)
}

func (n *Node) addChild(node *Node) {
	n.Children = append(n.Children, node)
	n.TotalChildren++
}

func (n *Node) print(prefix string) {
	fmt.Println(prefix + n.Path + " (" + strconv.Itoa(n.TotalChildren) + " children)")

	for _, child := range n.Children {
		child.print(prefix + "\t")
	}
}

// buildTree iterates through a list of packages to figure out all the unique
// imports and builds a dependency graph based on what it finds
func (d *Graph) buildTree() {
	config := build.Default

	// For each package, look for the dependencies and build out a tree
	for p := range d.Pkgs {
		pkg, _ := config.Import(d.Pkgs[p], d.SrcDir, build.AllowBinary)
		imports := pkg.Imports

		// Iterate through the imports and build our tree
		for i := range imports {
			// The full path of our current import
			path := imports[i]

			// When dealing with multiple packages, we can't assume that imports
			// are unique. Thus the nodes may already exist and we shouldn't do any work
			if d.Nodes[path] != nil {
				continue
			}

			// Ignore the GO standard library imports
			if _, ok := stdlib[strings.Split(path, "/")[0]]; ok && !d.StdLib {
				continue
			}

			// Keep track when traversing the path
			var currentNode *Node

			// For each directory in the path, create a node and link it
			// to it's parent and children
			for path != "" {
				// The first node (full path of import) should always get created
				if currentNode == nil {
					currentNode = &Node{
						Path: path,
					}
				} else {
					// Before creating a new parent node, check to see if there
					// is a common ancestor and use it if it exists
					if d.Nodes[path] == nil {
						currentNode.addParent(&Node{
							Path: path,
						})

						// Change the current node to the newly created item
						currentNode = currentNode.Parent
					} else {
						// Assume the common ancestor already has it's node tree setup
						currentNode.addParent(d.Nodes[path])
						currentNode = nil
						break
					}
				}

				// Keep track of the number of dependencies
				d.TotalDeps++

				// Constant time lookup to all of our nodes
				// based on their full path string
				d.Nodes[path] = currentNode

				// Keep popping off the tip of the path
				path, _ = filepath.Split(path)

				// Trailing slash in file path causes issues, remove it
				if strings.HasSuffix(path, "/") {
					path = path[:len(path)-1]
				}
			}

			// currentNode will be nil if there was already a common ancestor --
			// which means the root node already exists for that import path
			if currentNode != nil {
				d.RootNode.addChild(currentNode)
			}
		}
	}
}
