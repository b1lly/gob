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
	// Only update the parent if it doesn't exist
	n.Parent = node
	node.addChild(n)
}

func (n *Node) addChild(node *Node) {
	n.Children = append(n.Children, node)
}

func main() {
	config := build.Default

	// Testing path the package details of the current context
	pkg, _ := config.Import("yext/pages/storepages/storm", "/src/", build.AllowBinary)

	// A map from node string last node of a branch
	nodes := make(map[string]*Node)
	imports := pkg.Imports

	rootNode := Node{
		Path: "/src/",
	}

	// Iterate through the packages imports
	for i := range imports {
		// Store the full path of our import
		path := imports[i]

		// For each directory in the path, create a node and link it
		// to it's parent and children
		for path != "" {
			// Create our starter node
			if currentNode == nil {
				currentNode = &Node{
					Path: path,
				}
			} else {
				// Before creating a new parent node,
				// check to see if one exists
				if nodes[path] == nil {
					currentNode.setParent(&Node{
						Path: path,
					})
					// Change the current node to the newly created item
					currentNode = currentNode.Parent
				} else {
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
