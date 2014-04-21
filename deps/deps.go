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

		// For each import, create a node
		base := imports[i]

		currentNode := nodes[base]
		for base != "" {
			// If the full path exits, start at it
			if currentNode != nil {
				// Before creating a new parent node,
				// check to see if one exists
				if nodes[base] == nil {
					currentNode.setParent(&Node{
						Path: base,
					})
					// Change the current node to the newly created item
					currentNode = currentNode.Parent
				} else {
					currentNode.setParent(nodes[base])
					currentNode = nil
					break
				}
			} else {
				currentNode = &Node{
					Path: base,
				}
			}

			// Constant time lookup to all of our nodes
			// based on their full path string
			nodes[base] = currentNode

			// Keep popping off the tip of the path
			base, _ = filepath.Split(base)

			// Trailing slash in file path causes issues, remove it
			if strings.HasSuffix(base, "/") {
				base = base[:len(base)-1]
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
