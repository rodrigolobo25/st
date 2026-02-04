package cmd

import (
	"fmt"

	"github.com/rodrigolobo/st/internal/git"
	"github.com/rodrigolobo/st/internal/stack"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [branch-name]",
	Short: "Delete a branch from the stack",
	Long:  "Removes a branch, reparents its children to the deleted branch's parent, and deletes the git branch.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := loadAndBuild()
		if err != nil {
			return err
		}

		var branchName string
		if len(args) > 0 {
			branchName = args[0]
		} else {
			current, err := git.CurrentBranch()
			if err != nil {
				return fmt.Errorf("could not determine current branch: %w", err)
			}
			branchName = current
		}

		branch, ok := repo.Branches[branchName]
		if !ok {
			return fmt.Errorf("branch %q is not tracked by st", branchName)
		}

		// Reparent children
		for _, child := range branch.Children {
			if err := stack.ReparentBranch(child.Name, branch.Parent); err != nil {
				return fmt.Errorf("failed to reparent %s: %w", child.Name, err)
			}
			fmt.Printf("  Reparented %s → %s\n", child.Name, branch.Parent)
		}

		// If we're on this branch, checkout parent first
		current, _ := git.CurrentBranch()
		if current == branchName {
			target := branch.Parent
			if err := git.Checkout(target); err != nil {
				return fmt.Errorf("failed to checkout %s: %w", target, err)
			}
			fmt.Printf("  Switched to %s\n", target)
		}

		// Remove metadata
		if err := stack.UntrackBranch(branchName); err != nil {
			return fmt.Errorf("failed to untrack branch: %w", err)
		}

		// Delete git branch
		if err := git.DeleteBranch(branchName); err != nil {
			return fmt.Errorf("failed to delete branch: %w", err)
		}

		fmt.Printf("Deleted %s\n", branchName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
