package cmd

import (
	"fmt"

	"github.com/rodrigolobo/st/internal/git"
	"github.com/rodrigolobo/st/internal/stack"
	"github.com/spf13/cobra"
)

var continueCmd = &cobra.Command{
	Use:   "continue",
	Short: "Continue a restack after resolving conflicts",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !git.IsRestackInProgress() {
			return fmt.Errorf("no restack in progress")
		}

		// First, continue the git rebase
		if git.IsRebaseInProgress() {
			if err := git.RebaseContinue(); err != nil {
				return fmt.Errorf("rebase --continue failed: %w\nResolve remaining conflicts and run 'st continue' again", err)
			}
		}

		// Then continue restacking remaining branches
		result, err := stack.RestackRemaining()
		if err != nil {
			return err
		}

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

		fmt.Println("Restack complete")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(continueCmd)
}
