package stack

import (
	"fmt"
	"strings"

	"github.com/rodrigolobo/st/internal/git"
)

// RestackResult holds the result of a restack operation.
type RestackResult struct {
	Rebased  []string // branches that were rebased
	Skipped  []string // branches that were already up to date
	Conflict string   // branch where a conflict occurred (empty if none)
}

// RestackAll restacks all stacks in the repo.
func RestackAll(repo *Repo) (*RestackResult, error) {
	result := &RestackResult{}
	for _, root := range repo.Stacks {
		if err := restackBranch(root, repo.Trunk, result); err != nil {
			return result, err
		}
		if result.Conflict != "" {
			return result, nil
		}
	}
	return result, nil
}

// RestackCurrent restacks only the current stack.
func RestackCurrent(repo *Repo) (*RestackResult, error) {
	root := CurrentStack(repo)
	if root == nil {
		return nil, fmt.Errorf("current branch is not in a tracked stack")
	}

	result := &RestackResult{}
	if err := restackBranch(root, repo.Trunk, result); err != nil {
		return result, err
	}
	return result, nil
}

// RestackRemaining continues restacking from a saved state.
func RestackRemaining() (*RestackResult, error) {
	remaining, err := git.GetRestackState()
	if err != nil {
		return nil, fmt.Errorf("no restack in progress")
	}

	branches := strings.Split(remaining, ",")
	result := &RestackResult{}

	for i, branchName := range branches {
		branchName = strings.TrimSpace(branchName)
		if branchName == "" {
			continue
		}

		parent, err := git.GetStackParent(branchName)
		if err != nil {
			continue
		}

		rebased, err := doRebase(branchName, parent)
		if err != nil {
			// Save remaining branches
			remainingBranches := branches[i:]
			if err := git.SetRestackState(strings.Join(remainingBranches, ",")); err != nil {
				return result, fmt.Errorf("failed to save restack state: %w", err)
			}
			result.Conflict = branchName
			return result, nil
		}

		if rebased {
			result.Rebased = append(result.Rebased, branchName)
		} else {
			result.Skipped = append(result.Skipped, branchName)
		}
	}

	// All done, clear state
	git.ClearRestackState()
	return result, nil
}

func restackBranch(branch *Branch, expectedParent string, result *RestackResult) error {
	rebased, err := doRebase(branch.Name, expectedParent)
	if err != nil {
		// Save remaining branches for continue
		remaining := collectRemaining(branch)
		if saveErr := git.SetRestackState(strings.Join(remaining, ",")); saveErr != nil {
			return fmt.Errorf("rebase conflict on %s, and failed to save state: %w", branch.Name, saveErr)
		}
		result.Conflict = branch.Name
		return nil
	}

	if rebased {
		result.Rebased = append(result.Rebased, branch.Name)
	} else {
		result.Skipped = append(result.Skipped, branch.Name)
	}

	for _, child := range branch.Children {
		if err := restackBranch(child, branch.Name, result); err != nil {
			return err
		}
		if result.Conflict != "" {
			return nil
		}
	}
	return nil
}

// doRebase rebases branch onto expectedParent if needed.
// Returns true if a rebase was performed.
func doRebase(branchName, expectedParent string) (bool, error) {
	mb, err := git.MergeBase(branchName, expectedParent)
	if err != nil {
		return false, fmt.Errorf("could not find merge-base for %s and %s: %w", branchName, expectedParent, err)
	}

	parentTip, err := git.BranchTip(expectedParent)
	if err != nil {
		return false, fmt.Errorf("could not get tip of %s: %w", expectedParent, err)
	}

	if mb == parentTip {
		// Already up to date
		return false, nil
	}

	err = git.RebaseOnto(expectedParent, mb, branchName)
	if err != nil {
		return false, err
	}
	return true, nil
}

// collectRemaining collects remaining branch names from the current branch onward (DFS).
func collectRemaining(branch *Branch) []string {
	var result []string
	result = append(result, branch.Name)
	for _, child := range branch.Children {
		result = append(result, collectRemaining(child)...)
	}
	return result
}
