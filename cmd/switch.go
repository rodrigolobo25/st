package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rodrigolobo/st/internal/git"
	"github.com/rodrigolobo/st/internal/tui"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:     "switch",
	Aliases: []string{"sw"},
	Short:   "Interactively switch between stacked branches",
	Long:    "Opens an interactive TUI to browse and switch between stacked branches.",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := loadAndBuild()
		if err != nil {
			return err
		}

		if len(repo.Stacks) == 0 {
			return fmt.Errorf("no stacked branches found. Use 'st create <name>' to create one")
		}

		model := tui.NewSwitcherModel(repo)
		p := tea.NewProgram(model, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}

		m := finalModel.(tui.SwitcherModel)
		chosen := m.Chosen()
		if chosen == "" {
			return nil // user quit without selecting
		}

		if err := git.Checkout(chosen); err != nil {
			return fmt.Errorf("failed to checkout %s: %w", chosen, err)
		}
		fmt.Printf("Switched to %s\n", chosen)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
