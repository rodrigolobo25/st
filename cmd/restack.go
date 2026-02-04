package cmd

import (
	"fmt"

	"github.com/rodrigolobo/st/internal/git"
	"github.com/rodrigolobo/st/internal/stack"
	"github.com/spf13/cobra"
)

var restackCmd = &cobra.Command{
	Use:   "restack",
	Short: "Rebase all branches in the current stack",
	Long:  "Walks the stack from bottom to top, rebasing each branch onto its parent.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if git.IsRestackInProgress() {
			return fmt.Errorf("a restack is already in progress. Run 'st continue' to resume or resolve conflicts first")
		}

		repo, err := loadAndBuild()
		if err != nil {
			return err
		}

		// Save current branch to return to after restack
		currentBranch, _ := git.CurrentBranch()

		all, _ := cmd.Flags().GetBool("all")
		var result *stack.RestackResult
		if all {
			result, err = stack.RestackAll(repo)
		} else {
			result, err = stack.RestackCurrent(repo)
		}
		if err != nil {
			return err
		}

		// Print results
		for _, b := range result.Rebased {
			fmt.Printf("  ✓ Rebased %s\n", b)
		}
		for _, b := range result.Skipped {
			fmt.Printf("  · %s (already up to date)\n", b)
		}

		if result.Conflict != "" {
			fmt.Printf("\n  ✗ Conflict on %s\n", result.Conflict)
			fmt.Println("  Resolve conflicts, then run 'st continue'")
			return nil
		}

		// Return to the original branch
		if currentBranch != "" {
			_ = git.Checkout(currentBranch)
		}

		fmt.Println("Restack complete")
		return nil
	},
}

func init() {
	restackCmd.Flags().Bool("all", false, "restack all stacks, not just the current one")
	rootCmd.AddCommand(restackCmd)
}
