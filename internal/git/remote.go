package git

import (
	"fmt"
	"strings"
)

// Fetch fetches from a remote.
func Fetch(remote string) error {
	return RunSilent("fetch", remote)
}

// Push pushes a branch to a remote.
func Push(remote, branch string) error {
	return RunSilent("push", remote, branch)
}

// PushForce force-pushes a branch to a remote.
func PushForce(remote, branch string) error {
	return RunSilent("push", "--force-with-lease", remote, branch)
}

// FastForward fast-forwards a local branch to its remote tracking branch.
func FastForward(branch, remote string) error {
	remoteBranch := fmt.Sprintf("%s/%s", remote, branch)
	// Check if remote branch exists
	if !RemoteBranchExists(remote, branch) {
		return nil
	}
	// If currently on the branch, use pull --ff-only
	current, _ := CurrentBranch()
	if current == branch {
		return RunSilent("merge", "--ff-only", remoteBranch)
	}
	// Otherwise update the ref directly
	return RunSilent("fetch", remote, fmt.Sprintf("%s:%s", branch, branch))
}

// RemoteBranchExists checks if a branch exists on the remote.
func RemoteBranchExists(remote, branch string) bool {
	err := RunSilent("show-ref", "--verify", "--quiet", fmt.Sprintf("refs/remotes/%s/%s", remote, branch))
	return err == nil
}

// RemoteTrackingBranch returns the remote tracking ref for a local branch.
func RemoteTrackingBranch(branch string) (string, error) {
	return Run("rev-parse", "--abbrev-ref", fmt.Sprintf("%s@{upstream}", branch))
}

// HasRemote checks if any remote is configured.
func HasRemote() bool {
	out, err := Run("remote")
	return err == nil && strings.TrimSpace(out) != ""
}
