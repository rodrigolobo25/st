package stack

import (
	"fmt"
	"sort"

	"github.com/rodrigolobo/st/internal/git"
)

// BuildTree links branches into a tree and identifies stack roots.
func BuildTree(repo *Repo) {
	// Link children to parents
	for _, branch := range repo.Branches {
		if parent, ok := repo.Branches[branch.Parent]; ok {
			parent.Children = append(parent.Children, branch)
		}
	}

	// Sort children by name for deterministic output
	for _, branch := range repo.Branches {
		sort.Slice(branch.Children, func(i, j int) bool {
			return branch.Children[i].Name < branch.Children[j].Name
		})
	}

	// Find stack roots (branches whose parent is not a tracked branch)
	for _, branch := range repo.Branches {
		_, parentTracked := repo.Branches[branch.Parent]
		if !parentTracked {
			repo.Stacks = append(repo.Stacks, branch)
		}
	}
	sort.Slice(repo.Stacks, func(i, j int) bool {
		return repo.Stacks[i].Name < repo.Stacks[j].Name
	})
}

// CurrentStack returns the stack (root branch) containing the current branch.
// Returns nil if the current branch is trunk or untracked.
func CurrentStack(repo *Repo) *Branch {
	// Find the current branch
	var currentBranch *Branch
	for _, b := range repo.Branches {
		if b.Current {
			currentBranch = b
			break
		}
	}
	if currentBranch == nil {
		return nil
	}

	// Walk up to find the root
	b := currentBranch
	for b.Parent != repo.Trunk {
		parent, ok := repo.Branches[b.Parent]
		if !ok {
			return b
		}
		b = parent
	}
	return b
}

// CurrentBranch returns the branch marked as current in the repo.
func CurrentBranch(repo *Repo) *Branch {
	for _, b := range repo.Branches {
		if b.Current {
			return b
		}
	}
	return nil
}

// PathToTrunk returns the list of branches from the given branch down to trunk (exclusive).
func PathToTrunk(repo *Repo, branch *Branch) []*Branch {
	var path []*Branch
	b := branch
	for b != nil && b.Name != repo.Trunk {
		path = append(path, b)
		parent, ok := repo.Branches[b.Parent]
		if !ok {
			break
		}
		b = parent
	}
	// Reverse so trunk-adjacent is first
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// AllBranchesInStack returns all branches in a stack rooted at the given branch (DFS order).
func AllBranchesInStack(root *Branch) []*Branch {
	var result []*Branch
	var walk func(b *Branch)
	walk = func(b *Branch) {
		result = append(result, b)
		for _, child := range b.Children {
			walk(child)
		}
	}
	walk(root)
	return result
}

// Leaves returns all leaf branches (no children) in a stack.
func Leaves(root *Branch) []*Branch {
	var result []*Branch
	var walk func(b *Branch)
	walk = func(b *Branch) {
		if len(b.Children) == 0 {
			result = append(result, b)
		}
		for _, child := range b.Children {
			walk(child)
		}
	}
	walk(root)
	return result
}

// NeedsRestack checks if a branch needs to be rebased onto its parent.
func NeedsRestack(branch *Branch) bool {
	if branch.Parent == "" {
		return false
	}
	mb, err := git.MergeBase(branch.Name, branch.Parent)
	if err != nil {
		return false
	}
	parentTip, err := git.BranchTip(branch.Parent)
	if err != nil {
		return false
	}
	return mb != parentTip
}

// NavigateUp moves n branches away from trunk (toward leaves).
func NavigateUp(repo *Repo, n int) (string, error) {
	current := CurrentBranch(repo)
	if current == nil {
		return "", fmt.Errorf("current branch is not tracked by st")
	}

	target := current
	for i := 0; i < n; i++ {
		if len(target.Children) == 0 {
			return "", fmt.Errorf("already at the top of the stack")
		}
		target = target.Children[0]
	}
	return target.Name, nil
}

// NavigateDown moves n branches toward trunk.
func NavigateDown(repo *Repo, n int) (string, error) {
	current := CurrentBranch(repo)
	if current == nil {
		return "", fmt.Errorf("current branch is not tracked by st")
	}

	target := current
	for i := 0; i < n; i++ {
		if target.Parent == repo.Trunk {
			return "", fmt.Errorf("already at the bottom of the stack")
		}
		parent, ok := repo.Branches[target.Parent]
		if !ok {
			return "", fmt.Errorf("already at the bottom of the stack")
		}
		target = parent
	}
	return target.Name, nil
}

// NavigateTop moves to the top (leaf) of the current stack.
func NavigateTop(repo *Repo) (string, error) {
	current := CurrentBranch(repo)
	if current == nil {
		return "", fmt.Errorf("current branch is not tracked by st")
	}

	target := current
	for len(target.Children) > 0 {
		target = target.Children[0]
	}
	if target.Name == current.Name {
		return "", fmt.Errorf("already at the top of the stack")
	}
	return target.Name, nil
}

// NavigateBottom moves to the bottom (closest to trunk) of the current stack.
func NavigateBottom(repo *Repo) (string, error) {
	current := CurrentBranch(repo)
	if current == nil {
		return "", fmt.Errorf("current branch is not tracked by st")
	}

	root := CurrentStack(repo)
	if root == nil {
		return "", fmt.Errorf("current branch is not in a stack")
	}
	if root.Name == current.Name {
		return "", fmt.Errorf("already at the bottom of the stack")
	}
	return root.Name, nil
}
