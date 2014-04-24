package dependencies

import (
	"fmt"
	"go/build"
	"path/filepath"
	"strconv"
	"strings"
)

type Graph struct {
	StdLib bool     // the value 'false' will ignore stdlib imports
	SrcDir string   // the root src directory of all the packages
	Pkgs   []string // list of packages to use when building our depdendency tree

	TotalDeps int // total number of dependencies used across all packages

	RootNode *Node
	Nodes    map[string]*Node // map of all the dependencies across all our projects
}

// NewGraph is used to build out a dependency tree, and provide useful helpers
// for traversing and doing analysis. It uses a list of packages, analysis the imports
// and build the tree.
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

// ListNodes returns a unique list of nodes and their path based on the dependency tree
func (d *Graph) ListNodes() (nodes []string) {
	for n := range d.Nodes {
		nodes = append(nodes, d.Nodes[n].Path)
	}

	return
}

// ListDeps returns a unique list of all the dependencies based on the graph
func (d *Graph) ListDeps() (nodes []string) {
	for n := range d.Nodes {
		if d.Nodes[n].IsDep {
			nodes = append(nodes, d.Nodes[n].Path)
		}
	}

	return
}

// Node contains a data for a given dependency and it's relationship to others.
// It also contains some extra helpful data for traversing
type Node struct {
	Parent        *Node
	Children      []*Node
	TotalChildren int

	Path string

	IsDep       bool // "true" means this node contains a dependency
	IsCoreDep   bool // "true" means that this node stems from the pkg that is importing it
	IsDuplicate bool // "true" means the dependency is used more then once
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

		fmt.Println(d.Pkgs[p])

		// Iterate through the imports and build our tree
		for i := range imports {
			// The full path of our current import
			path := imports[i]

			// When dealing with multiple packages, we can't assume that imports
			// are unique. Thus the nodes may already exist and we shouldn't do any work
			if d.Nodes[path] != nil {
				d.Nodes[path].IsDuplicate = true
				continue
			}

			// Ignore the GO standard library imports
			if _, ok := stdlib[strings.Split(path, "/")[0]]; ok && !d.StdLib {
				continue
			}

			// Keep track when traversing the path
			var currentNode = &Node{
				Path:      path,
				IsDep:     true,
				IsCoreDep: strings.HasPrefix(d.Pkgs[p], path),
			}

			// Keep track of the number of dependencies
			d.TotalDeps++

			// Link our dependency node to it's ancestors
			for path != "" {
				// Constant time lookup to all of our nodes
				// based on their full path string
				d.Nodes[path] = currentNode

				// Keep popping off the tip of the path
				path, _ = filepath.Split(path)

				if len(path) > 0 {
					// Trailing slash in file path causes issues, remove it
					if strings.HasSuffix(path, "/") {
						path = path[:len(path)-1]
					}

					// Create nodes for all directory paths if they don't exist
					if d.Nodes[path] == nil {
						currentNode.addParent(&Node{
							Path: path,
						})

						// Change the current node to the newly created item
						currentNode = currentNode.Parent
					} else {
						// Otherwise, assume the common ancestor already has it's tree built
						currentNode.addParent(d.Nodes[path])
						currentNode = nil
						break
					}
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
