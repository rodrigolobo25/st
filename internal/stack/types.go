package stack

// Branch represents a tracked branch in a stack.
type Branch struct {
	Name     string
	Parent   string
	Children []*Branch
	Current  bool
}

// Repo holds the full state of all tracked stacks.
type Repo struct {
	Trunk    string
	Branches map[string]*Branch // all tracked branches
	Stacks   []*Branch          // root branches (parent == trunk)
}
