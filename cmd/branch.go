package cmd

import (
	"fmt"

	"github.com/rodrigolobo/st/internal/git"
	"github.com/rodrigolobo/st/internal/tui"
	"github.com/spf13/cobra"
)

var branchCmd = &cobra.Command{
	Use:     "branch",
	Aliases: []string{"b"},
	Short:   "Show info about the current branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := loadAndBuild()
		if err != nil {
			return err
		}

		current, err := git.CurrentBranch()
		if err != nil {
			return fmt.Errorf("could not determine current branch: %w", err)
		}

		branch, ok := repo.Branches[current]
		if !ok {
			return fmt.Errorf("branch %q is not tracked by st", current)
		}

		fmt.Print(tui.RenderBranchInfo(branch, repo))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(branchCmd)
}
