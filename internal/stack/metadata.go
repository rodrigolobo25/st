package stack

import (
	"fmt"
	"strings"

	"github.com/rodrigolobo/st/internal/git"
)

// LoadRepo loads the full repo state from git config.
func LoadRepo() (*Repo, error) {
	trunk, err := git.GetTrunk()
	if err != nil {
		return nil, err
	}

	current, err := git.CurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("could not determine current branch: %w", err)
	}

	entries, err := git.ConfigGetRegexp(`^stack\.`)
	if err != nil {
		return nil, fmt.Errorf("could not read stack config: %w", err)
	}

	repo := &Repo{
		Trunk:    trunk,
		Branches: make(map[string]*Branch),
	}

	// Parse all stack.<name>.parent entries
	for _, entry := range entries {
		key := entry[0]
		value := entry[1]

		// key is like "stack.feat-auth.parent"
		parts := strings.SplitN(key, ".", 3)
		if len(parts) != 3 || parts[2] != "parent" {
			continue
		}
		branchName := parts[1]
		parentName := value

		branch := &Branch{
			Name:    branchName,
			Parent:  parentName,
			Current: branchName == current,
		}
		repo.Branches[branchName] = branch
	}

	return repo, nil
}

// TrackBranch adds a new branch to the stack metadata.
func TrackBranch(name, parent string) error {
	return git.SetStackParent(name, parent)
}

// UntrackBranch removes a branch from the stack metadata.
func UntrackBranch(name string) error {
	return git.RemoveStackSection(name)
}

// ReparentBranch changes the parent of a branch.
func ReparentBranch(name, newParent string) error {
	return git.SetStackParent(name, newParent)
}
