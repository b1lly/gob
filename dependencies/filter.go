package dependencies

type Filter struct {
	Limit int // Total number of results to limit to
	Graph *Graph

	deps     []string
	toRemove int
}

// NewFilter will take in a dependency graph and try to first prioritize the
// dependencies by importance, and create a new list of dependencies based on
// the limit (total number of deps to watch)
func NewFilter(f *Filter) *Filter {
	filter := Filter{
		Limit:    f.Limit,
		Graph:    f.Graph,
		toRemove: f.Graph.TotalDeps - f.Limit,
	}

	if filter.toRemove > 0 {
		filter.prioritize()
		filter.clean()
	}

	return &filter
}

// ListDeps the dependencies that resulted from the prioritization and clean
func (f *Filter) ListDeps() []string {
	var deps []string
	for _, dep := range f.deps {
		deps = append(deps, dep)
	}
	return deps
}

func (f *Filter) prioritize() {
	var priority = struct {
		high   []string
		medium []string
		low    []string
	}{}

	for _, dep := range f.Graph.ListDeps() {
		node := f.Graph.Nodes[dep]

		if node.IsCoreDep && node.IsDuplicate {
			priority.high = append(priority.high, dep)
		} else if node.IsCoreDep || node.IsDuplicate {
			priority.medium = append(priority.medium, dep)
		} else {
			priority.low = append(priority.low, dep)
		}
	}

	f.deps = append(f.deps, priority.high...)
	f.deps = append(f.deps, priority.medium...)
	f.deps = append(f.deps, priority.low...)
}

func (f *Filter) clean() {
	for i := len(f.deps); i >= 0; i-- {
		if f.toRemove > 0 {
			f.deps = f.deps[:len(f.deps)-1]
			f.toRemove--
		} else {
			return
		}
	}
}
