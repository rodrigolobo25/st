package cmd

import (
	"fmt"

	"github.com/rodrigolobo/st/internal/git"
	"github.com/rodrigolobo/st/internal/stack"
	"github.com/spf13/cobra"
)

var reparentCmd = &cobra.Command{
	Use:   "reparent <new-parent>",
	Short: "Change the parent of the current branch",
	Long:  "Changes the parent of the current branch to a new parent. Run 'st restack' afterward to rebase onto the new parent.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		newParent := args[0]

		current, err := git.CurrentBranch()
		if err != nil {
			return fmt.Errorf("could not determine current branch: %w", err)
		}

		trunk, err := git.GetTrunk()
		if err != nil {
			return err
		}

		if current == trunk {
			return fmt.Errorf("cannot reparent trunk branch")
		}

		// Load repo to check if current branch is tracked
		repo, err := stack.LoadRepo()
		if err != nil {
			return err
		}

		branch, ok := repo.Branches[current]
		if !ok {
			return fmt.Errorf("branch %q is not tracked by st", current)
		}

		oldParent := branch.Parent

		// Verify new parent exists as a git branch
		if !git.BranchExists(newParent) {
			return fmt.Errorf("branch %q does not exist", newParent)
		}

		if newParent == current {
			return fmt.Errorf("cannot reparent a branch to itself")
		}

		if newParent == oldParent {
			fmt.Printf("Branch %q is already parented to %q\n", current, newParent)
			return nil
		}

		if err := stack.ReparentBranch(current, newParent); err != nil {
			return fmt.Errorf("failed to reparent: %w", err)
		}

		fmt.Printf("Reparented %q: %s → %s\n", current, oldParent, newParent)
		fmt.Println("Run 'st restack' to rebase onto the new parent.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reparentCmd)
}
