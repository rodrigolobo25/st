package cmd

import (
	"fmt"

	"github.com/rodrigolobo/st/internal/git"
	"github.com/rodrigolobo/st/internal/stack"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create <branch-name>",
	Short: "Create a new stacked branch",
	Long:  "Creates a new branch off the current branch and tracks it in the stack.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		branchName := args[0]

		trunk, err := git.GetTrunk()
		if err != nil {
			return err
		}

		current, err := git.CurrentBranch()
		if err != nil {
			return fmt.Errorf("could not determine current branch: %w", err)
		}

		if git.BranchExists(branchName) {
			return fmt.Errorf("branch %q already exists", branchName)
		}

		// If -a flag, stage all changes
		stageAll, _ := cmd.Flags().GetBool("all")
		if stageAll {
			if err := git.StageAll(); err != nil {
				return fmt.Errorf("failed to stage changes: %w", err)
			}
		}

		// If -m flag and there are staged changes, commit first
		message, _ := cmd.Flags().GetString("message")
		if message != "" && git.HasStagedChanges() {
			if err := git.Commit(message); err != nil {
				return fmt.Errorf("failed to commit: %w", err)
			}
			fmt.Println("Committed changes on current branch")
		}

		// Create the new branch
		if err := git.CreateBranch(branchName); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}

		// Track it: parent is either current branch or trunk
		parent := current
		if current == trunk {
			parent = trunk
		}

		if err := stack.TrackBranch(branchName, parent); err != nil {
			return fmt.Errorf("failed to track branch: %w", err)
		}

		fmt.Printf("Created and checked out branch %q (parent: %s)\n", branchName, parent)
		return nil
	},
}

func init() {
	createCmd.Flags().StringP("message", "m", "", "commit staged changes with message before creating branch")
	createCmd.Flags().BoolP("all", "a", false, "stage all changes before committing")
	rootCmd.AddCommand(createCmd)
}
