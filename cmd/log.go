package cmd

import (
	"fmt"

	"github.com/rodrigolobo/st/internal/stack"
	"github.com/rodrigolobo/st/internal/tui"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:     "log",
	Aliases: []string{"ls"},
	Short:   "Show the stack tree",
	Long:    "Displays the tree of all stacked branches with commit counts and status.",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := stack.LoadRepo()
		if err != nil {
			return err
		}
		stack.BuildTree(repo)

		if len(repo.Stacks) == 0 {
			fmt.Println("No stacked branches found. Use 'st create <name>' to create one.")
			return nil
		}

		fmt.Print(tui.RenderTree(repo))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}
