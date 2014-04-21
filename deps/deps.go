package main

import (
	"fmt"
	"go/build"
	"path/filepath"
	"strings"
)

type Deps struct {
	Nodes    map[string]string
	RootNode *Node
}

type Node struct {
	Parent   *Node
	Children []*Node
	Path     string
}

func (n *Node) setParent(node *Node) {
	n.Parent = node
	node.addChild(n)
}

func (n *Node) addChild(node *Node) {
	n.Children = append(n.Children, node)
}

func main() {
	config := build.Default

	// Testing path where the package details/imports exist
	pkg, _ := config.Import("yext/pages/storepages/storm", "/src/", build.AllowBinary)

	// A map from "full path" string to the node
	nodes := make(map[string]*Node)
	imports := pkg.Imports

	rootNode := Node{
		Path: "/src/",
	}

	// Iterate through the packages imports
	for i := range imports {
		// The full path of our current dependency
		path := imports[i]

		// Used to keep track when traversing the path
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
				if nodes[path] == nil {
					currentNode.setParent(&Node{
						Path: path,
					})

					// Change the current node to the newly created item
					currentNode = currentNode.Parent
				} else {
					// Assume the common ancestor already has it's node tree setup
					currentNode.setParent(nodes[path])
					currentNode = nil
					break
				}
			}

			// Constant time lookup to all of our nodes
			// based on their full path string
			nodes[path] = currentNode

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
			rootNode.addChild(currentNode)
		}
	}

	// Print out our results of our tree
	rootNode.Print("\t")
}

func (n *Node) Print(prefix string) {
	fmt.Println(prefix + n.Path)

	for _, child := range n.Children {
		child.Print(prefix + "\t")
	}
}
