package tui

import (
	"fmt"
	"strings"

	"github.com/rodrigolobo/st/internal/git"
	"github.com/rodrigolobo/st/internal/stack"
)

// RenderTree renders the full stack tree as a string.
func RenderTree(repo *stack.Repo) string {
	var sb strings.Builder

	// Group roots by parent
	groups := make(map[string][]*stack.Branch)
	var parentOrder []string
	for _, root := range repo.Stacks {
		if _, seen := groups[root.Parent]; !seen {
			parentOrder = append(parentOrder, root.Parent)
		}
		groups[root.Parent] = append(groups[root.Parent], root)
	}

	for gi, parent := range parentOrder {
		if gi > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(TrunkStyle.Render(parent))
		sb.WriteString("\n")

		roots := groups[parent]
		for i, root := range roots {
			isLast := i == len(roots)-1
			renderBranch(&sb, root, repo, "", isLast)
		}
	}

	return sb.String()
}

func renderBranch(sb *strings.Builder, branch *stack.Branch, repo *stack.Repo, prefix string, isLast bool) {
	// Draw connector
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	// Branch name
	var nameStr string
	if branch.Current {
		nameStr = CurrentBranchStyle.Render(branch.Name)
	} else {
		nameStr = BranchStyle.Render(branch.Name)
	}

	// Commit count
	commitCount := ""
	count, err := git.CommitCount(branch.Parent, branch.Name)
	if err == nil {
		if count == 1 {
			commitCount = DimStyle.Render("  1 commit")
		} else {
			commitCount = DimStyle.Render(fmt.Sprintf("  %d commits", count))
		}
	}

	// Needs restack indicator
	restackIndicator := ""
	if stack.NeedsRestack(branch) {
		restackIndicator = WarningStyle.Render("  ⟳ needs restack")
	}

	// Current marker
	hereMarker := ""
	if branch.Current {
		hereMarker = HereMarker
	}

	sb.WriteString(prefix + connector + nameStr + commitCount + restackIndicator + hereMarker + "\n")

	// Children
	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	for i, child := range branch.Children {
		childIsLast := i == len(branch.Children)-1
		renderBranch(sb, child, repo, childPrefix, childIsLast)
	}
}

// RenderBranchInfo renders detailed info about a branch.
func RenderBranchInfo(branch *stack.Branch, repo *stack.Repo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Branch: %s\n", CurrentBranchStyle.Render(branch.Name)))
	sb.WriteString(fmt.Sprintf("Parent: %s\n", InfoStyle.Render(branch.Parent)))

	// Children
	if len(branch.Children) > 0 {
		children := make([]string, len(branch.Children))
		for i, c := range branch.Children {
			children[i] = c.Name
		}
		sb.WriteString(fmt.Sprintf("Children: %s\n", InfoStyle.Render(strings.Join(children, ", "))))
	} else {
		sb.WriteString(fmt.Sprintf("Children: %s\n", DimStyle.Render("none")))
	}

	// Commit count
	count, err := git.CommitCount(branch.Parent, branch.Name)
	if err == nil {
		sb.WriteString(fmt.Sprintf("Commits: %d\n", count))
	}

	// Restack status
	if stack.NeedsRestack(branch) {
		sb.WriteString(WarningStyle.Render("Status: needs restack") + "\n")
	} else {
		sb.WriteString(SuccessStyle.Render("Status: up to date") + "\n")
	}

	return sb.String()
}
