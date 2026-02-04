package cmd

import (
	"fmt"

	"github.com/rodrigolobo/st/internal/git"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize st in the current git repository",
	Long:  "Sets up st by detecting or specifying the trunk branch.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !git.IsInsideWorkTree() {
			return fmt.Errorf("not inside a git repository")
		}

		trunk, _ := cmd.Flags().GetString("trunk")

		if trunk == "" {
			// Auto-detect trunk
			if git.BranchExists("main") {
				trunk = "main"
			} else if git.BranchExists("master") {
				trunk = "master"
			} else {
				return fmt.Errorf("could not auto-detect trunk branch. Use --trunk to specify")
			}
		} else {
			if !git.BranchExists(trunk) {
				return fmt.Errorf("branch %q does not exist", trunk)
			}
		}

		if err := git.SetTrunk(trunk); err != nil {
			return fmt.Errorf("failed to set trunk: %w", err)
		}

		fmt.Printf("Initialized st with trunk branch: %s\n", trunk)
		return nil
	},
}

func init() {
	initCmd.Flags().StringP("trunk", "t", "", "trunk branch name (default: auto-detect main/master)")
	rootCmd.AddCommand(initCmd)
}
