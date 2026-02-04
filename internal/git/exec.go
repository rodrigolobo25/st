package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Run executes a git command and returns its trimmed stdout.
func Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	result := strings.TrimSpace(string(out))
	if err != nil {
		return result, fmt.Errorf("git %s: %s", strings.Join(args, " "), result)
	}
	return result, nil
}

// RunSilent executes a git command, returning only the error if any.
func RunSilent(args ...string) error {
	_, err := Run(args...)
	return err
}

// RunPassthrough executes a git command with stdin/stdout/stderr attached to the terminal.
func RunPassthrough(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// IsInsideWorkTree returns true if the current directory is inside a git working tree.
func IsInsideWorkTree() bool {
	out, err := Run("rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// TopLevel returns the root directory of the current git repository.
func TopLevel() (string, error) {
	return Run("rev-parse", "--show-toplevel")
}
