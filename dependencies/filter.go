package dependencies

import "strings"

type Filter struct {
	Limit int

	Graph *Graph
	Deps  []string

	toRemove int
}

func NewFilter(f *Filter) *Filter {
	filter := Filter{
		Limit:    f.Limit,
		Graph:    f.Graph,
		Deps:     f.Graph.ListNodes(),
		toRemove: f.Graph.TotalDeps - f.Limit,
	}

	f.Graph.RootNode.print("\t")

	if filter.toRemove > 0 {
		filter.moveToCommon()
	}

	return &filter
}

// TODO(billy) make smarter
func (f *Filter) moveToCommon() {
	// Find all the nodes with children and also the node with the most children
	for d := range f.Deps {
		// The node representation of a particular path
		node := f.Graph.Nodes[f.Deps[d]]

		if len(strings.Split(node.Path, "/")) == 2 {
			delete(f.Graph.Nodes, node.Path)
		}
	}
}
