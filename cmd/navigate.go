package cmd

import (
	"fmt"
	"strconv"

	"github.com/rodrigolobo/st/internal/git"
	"github.com/rodrigolobo/st/internal/stack"
	"github.com/spf13/cobra"
)

func loadAndBuild() (*stack.Repo, error) {
	repo, err := stack.LoadRepo()
	if err != nil {
		return nil, err
	}
	stack.BuildTree(repo)
	return repo, nil
}

var upCmd = &cobra.Command{
	Use:   "up [n]",
	Short: "Move up the stack (away from trunk)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n := 1
		if len(args) > 0 {
			var err error
			n, err = strconv.Atoi(args[0])
			if err != nil || n < 1 {
				return fmt.Errorf("invalid count: %s", args[0])
			}
		}

		repo, err := loadAndBuild()
		if err != nil {
			return err
		}

		target, err := stack.NavigateUp(repo, n)
		if err != nil {
			return err
		}

		if err := git.Checkout(target); err != nil {
			return fmt.Errorf("failed to checkout %s: %w", target, err)
		}
		fmt.Printf("Switched to %s\n", target)
		return nil
	},
}

var downCmd = &cobra.Command{
	Use:   "down [n]",
	Short: "Move down the stack (toward trunk)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n := 1
		if len(args) > 0 {
			var err error
			n, err = strconv.Atoi(args[0])
			if err != nil || n < 1 {
				return fmt.Errorf("invalid count: %s", args[0])
			}
		}

		repo, err := loadAndBuild()
		if err != nil {
			return err
		}

		target, err := stack.NavigateDown(repo, n)
		if err != nil {
			return err
		}

		if err := git.Checkout(target); err != nil {
			return fmt.Errorf("failed to checkout %s: %w", target, err)
		}
		fmt.Printf("Switched to %s\n", target)
		return nil
	},
}

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Jump to the top of the stack (leaf branch)",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := loadAndBuild()
		if err != nil {
			return err
		}

		target, err := stack.NavigateTop(repo)
		if err != nil {
			return err
		}

		if err := git.Checkout(target); err != nil {
			return fmt.Errorf("failed to checkout %s: %w", target, err)
		}
		fmt.Printf("Switched to %s\n", target)
		return nil
	},
}

var bottomCmd = &cobra.Command{
	Use:   "bottom",
	Short: "Jump to the bottom of the stack (closest to trunk)",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := loadAndBuild()
		if err != nil {
			return err
		}

		target, err := stack.NavigateBottom(repo)
		if err != nil {
			return err
		}

		if err := git.Checkout(target); err != nil {
			return fmt.Errorf("failed to checkout %s: %w", target, err)
		}
		fmt.Printf("Switched to %s\n", target)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(topCmd)
	rootCmd.AddCommand(bottomCmd)
}
