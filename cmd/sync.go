package cmd

import (
	"fmt"

	"github.com/rodrigolobo/st/internal/git"
	"github.com/rodrigolobo/st/internal/stack"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync with remote and restack",
	Long:  "Fetches from remote, fast-forwards trunk, cleans merged branches, and restacks all stacks.",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := loadAndBuild()
		if err != nil {
			return err
		}

		// Save current branch
		currentBranch, _ := git.CurrentBranch()

		// Fetch from origin
		if git.HasRemote() {
			fmt.Println("Fetching from origin...")
			if err := git.Fetch("origin"); err != nil {
				return fmt.Errorf("failed to fetch: %w", err)
			}

			// Fast-forward trunk
			fmt.Printf("Updating %s...\n", repo.Trunk)
			if err := git.FastForward(repo.Trunk, "origin"); err != nil {
				fmt.Printf("  Warning: could not fast-forward %s: %v\n", repo.Trunk, err)
			}
		}

		// Detect and clean merged branches
		var merged []string
		for name, branch := range repo.Branches {
			if git.IsBranchMergedInto(name, repo.Trunk) {
				merged = append(merged, name)
				// Reparent children first
				for _, child := range branch.Children {
					if err := stack.ReparentBranch(child.Name, branch.Parent); err != nil {
						fmt.Printf("  Warning: failed to reparent %s: %v\n", child.Name, err)
					}
				}
			}
		}

		for _, name := range merged {
			if name == currentBranch {
				// Switch to trunk before deleting current branch
				if err := git.Checkout(repo.Trunk); err != nil {
					fmt.Printf("  Warning: could not switch to trunk: %v\n", err)
					continue
				}
				currentBranch = repo.Trunk
			}
			if err := stack.UntrackBranch(name); err != nil {
				fmt.Printf("  Warning: failed to untrack %s: %v\n", name, err)
			}
			if err := git.DeleteBranch(name); err != nil {
				fmt.Printf("  Warning: failed to delete %s: %v\n", name, err)
			}
			fmt.Printf("  Cleaned merged branch %s\n", name)
		}

		// Reload repo after cleaning
		repo, err = loadAndBuild()
		if err != nil {
			return err
		}

		if len(repo.Stacks) == 0 {
			fmt.Println("No stacks to restack")
			return nil
		}

		// Restack all
		fmt.Println("Restacking...")
		result, err := stack.RestackAll(repo)
		if err != nil {
			return err
		}

		for _, b := range result.Rebased {
			fmt.Printf("  ✓ Rebased %s\n", b)
		}
		for _, b := range result.Skipped {
			fmt.Printf("  · %s (up to date)\n", b)
		}

		if result.Conflict != "" {
			fmt.Printf("\n  ✗ Conflict on %s\n", result.Conflict)
			fmt.Println("  Resolve conflicts, then run 'st continue'")
			return nil
		}

		// Return to original branch if it still exists
		if currentBranch != "" && currentBranch != repo.Trunk && git.BranchExists(currentBranch) {
			_ = git.Checkout(currentBranch)
		}

		fmt.Println("Sync complete")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
