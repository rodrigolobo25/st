package git

import (
	"fmt"
	"strconv"
	"strings"
)

// CurrentBranch returns the name of the currently checked-out branch.
func CurrentBranch() (string, error) {
	return Run("symbolic-ref", "--short", "HEAD")
}

// BranchExists checks if a branch exists locally.
func BranchExists(name string) bool {
	err := RunSilent("show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", name))
	return err == nil
}

// CreateBranch creates a new branch at the current HEAD.
func CreateBranch(name string) error {
	return RunSilent("checkout", "-b", name)
}

// Checkout switches to an existing branch.
func Checkout(name string) error {
	return RunSilent("checkout", name)
}

// DeleteBranch deletes a local branch.
func DeleteBranch(name string) error {
	return RunSilent("branch", "-D", name)
}

// MergeBase returns the merge-base of two refs.
func MergeBase(a, b string) (string, error) {
	return Run("merge-base", a, b)
}

// RevParse returns the SHA for a ref.
func RevParse(ref string) (string, error) {
	return Run("rev-parse", ref)
}

// CommitCount returns the number of commits between ancestor and descendant.
func CommitCount(ancestor, descendant string) (int, error) {
	out, err := Run("rev-list", "--count", fmt.Sprintf("%s..%s", ancestor, descendant))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(out))
}

// BranchTip returns the commit SHA at the tip of a branch.
func BranchTip(branch string) (string, error) {
	return RevParse(branch)
}

// RebaseOnto performs git rebase --onto.
func RebaseOnto(newBase, oldBase, branch string) error {
	_, err := Run("rebase", "--onto", newBase, oldBase, branch)
	return err
}

// RebaseContinue runs git rebase --continue.
func RebaseContinue() error {
	_, err := Run("rebase", "--continue")
	return err
}

// IsRebaseInProgress checks if a git rebase is in progress.
func IsRebaseInProgress() bool {
	out, _ := Run("rev-parse", "--git-dir")
	if out == "" {
		return false
	}
	// Check for rebase-merge or rebase-apply directory
	err1 := RunSilent("rev-parse", "--verify", "--quiet", "REBASE_HEAD")
	return err1 == nil
}

// CommitAmend amends the current HEAD commit.
func CommitAmend(message string) error {
	if message != "" {
		return RunSilent("commit", "--amend", "-m", message)
	}
	return RunSilent("commit", "--amend", "--no-edit")
}

// Commit creates a new commit with the given message.
func Commit(message string) error {
	return RunSilent("commit", "-m", message)
}

// StageAll stages all changes.
func StageAll() error {
	return RunSilent("add", "-A")
}

// HasStagedChanges checks if there are staged changes.
func HasStagedChanges() bool {
	err := RunSilent("diff", "--cached", "--quiet")
	return err != nil
}

// ListLocalBranches returns all local branch names.
func ListLocalBranches() ([]string, error) {
	out, err := Run("branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	var branches []string
	for _, b := range strings.Split(out, "\n") {
		b = strings.TrimSpace(b)
		if b != "" {
			branches = append(branches, b)
		}
	}
	return branches, nil
}

// IsBranchMergedInto checks if branch is merged into target.
func IsBranchMergedInto(branch, target string) bool {
	mb, err := MergeBase(branch, target)
	if err != nil {
		return false
	}
	tip, err := BranchTip(branch)
	if err != nil {
		return false
	}
	return mb == tip
}

// ShortLog returns a one-line log for a branch relative to its parent.
func ShortLog(parent, branch string) (string, error) {
	return Run("log", "--oneline", fmt.Sprintf("%s..%s", parent, branch))
}
