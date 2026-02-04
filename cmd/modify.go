package cmd

import (
	"fmt"

	"github.com/rodrigolobo/st/internal/git"
	"github.com/spf13/cobra"
)

var modifyCmd = &cobra.Command{
	Use:     "modify",
	Aliases: []string{"m"},
	Short:   "Amend or commit changes on the current branch",
	Long:    "By default, amends staged changes to HEAD. Use -c to create a new commit instead.",
	RunE: func(cmd *cobra.Command, args []string) error {
		stageAll, _ := cmd.Flags().GetBool("all")
		newCommit, _ := cmd.Flags().GetBool("commit")
		message, _ := cmd.Flags().GetString("message")

		// Stage all if requested
		if stageAll {
			if err := git.StageAll(); err != nil {
				return fmt.Errorf("failed to stage changes: %w", err)
			}
		}

		if !git.HasStagedChanges() && !newCommit {
			// If no staged changes and no -c, allow amend with message only
			if message == "" {
				return fmt.Errorf("no staged changes to commit. Use -a to stage all changes")
			}
		}

		if newCommit {
			// Create a new commit
			if message == "" {
				return fmt.Errorf("commit message required with -c flag")
			}
			if err := git.Commit(message); err != nil {
				return fmt.Errorf("failed to commit: %w", err)
			}
			fmt.Println("Created new commit")
		} else {
			// Amend HEAD
			if err := git.CommitAmend(message); err != nil {
				return fmt.Errorf("failed to amend: %w", err)
			}
			fmt.Println("Amended HEAD commit")
		}

		return nil
	},
}

func init() {
	modifyCmd.Flags().BoolP("all", "a", false, "stage all changes before committing")
	modifyCmd.Flags().BoolP("commit", "c", false, "create a new commit instead of amending")
	modifyCmd.Flags().StringP("message", "m", "", "commit message")
	rootCmd.AddCommand(modifyCmd)
}
