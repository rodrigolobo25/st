package tui

import (
	"strings"
	"testing"

	"github.com/rodrigolobo/st/internal/stack"
)

// makeRepo builds an in-memory Repo and runs BuildTree on it.
func makeRepo(trunk string, branches map[string]string, current string) *stack.Repo {
	repo := &stack.Repo{
		Trunk:    trunk,
		Branches: make(map[string]*stack.Branch),
	}
	for name, parent := range branches {
		repo.Branches[name] = &stack.Branch{
			Name:    name,
			Parent:  parent,
			Current: name == current,
		}
	}
	stack.BuildTree(repo)
	return repo
}

func TestRenderTree_SingleTrunkRoot(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat-a": "main",
	}, "")

	out := RenderTree(repo)
	if !strings.Contains(out, "main") {
		t.Error("output should contain trunk name 'main'")
	}
	if !strings.Contains(out, "feat-a") {
		t.Error("output should contain branch 'feat-a'")
	}
	if !strings.Contains(out, "└── ") {
		t.Error("output should contain tree connector for single branch")
	}
}

func TestRenderTree_MultipleTrunkRoots(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat-a": "main",
		"feat-b": "main",
	}, "")

	out := RenderTree(repo)
	if !strings.Contains(out, "├── ") {
		t.Error("output should contain non-last connector")
	}
	if !strings.Contains(out, "└── ") {
		t.Error("output should contain last connector")
	}
}

func TestRenderTree_UntrackedParentGroup(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat-on-main": "main",
		"orphan":       "external/branch",
	}, "")

	out := RenderTree(repo)
	if !strings.Contains(out, "main") {
		t.Error("output should contain 'main' header")
	}
	if !strings.Contains(out, "external/branch") {
		t.Error("output should contain 'external/branch' header")
	}
	if !strings.Contains(out, "feat-on-main") {
		t.Error("output should contain 'feat-on-main' branch")
	}
	if !strings.Contains(out, "orphan") {
		t.Error("output should contain 'orphan' branch")
	}
}

func TestRenderTree_MultipleUntrackedParents(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat-a":     "main",
		"on-ext-a":   "external/a",
		"on-ext-b":   "external/b",
	}, "")

	out := RenderTree(repo)
	// All three parent headers should appear
	if !strings.Contains(out, "main") {
		t.Error("output should contain 'main'")
	}
	if !strings.Contains(out, "external/a") {
		t.Error("output should contain 'external/a'")
	}
	if !strings.Contains(out, "external/b") {
		t.Error("output should contain 'external/b'")
	}
}

func TestRenderTree_CurrentBranchMarker(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat-a": "main",
	}, "feat-a")

	out := RenderTree(repo)
	if !strings.Contains(out, "you are here") {
		t.Error("output should contain 'you are here' marker for current branch")
	}
}

func TestRenderTree_NoCurrent_NoMarker(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat-a": "main",
	}, "")

	out := RenderTree(repo)
	if strings.Contains(out, "you are here") {
		t.Error("output should not contain 'you are here' when no current branch")
	}
}

func TestRenderTree_NestedChildren(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"root":  "main",
		"child": "root",
		"grand": "child",
	}, "")

	out := RenderTree(repo)
	if !strings.Contains(out, "root") {
		t.Error("output should contain root")
	}
	if !strings.Contains(out, "child") {
		t.Error("output should contain child")
	}
	if !strings.Contains(out, "grand") {
		t.Error("output should contain grand")
	}
}

func TestRenderTree_Empty(t *testing.T) {
	repo := makeRepo("main", map[string]string{}, "")

	out := RenderTree(repo)
	// No stacks means no output (no parent groups)
	if strings.Contains(out, "main") {
		t.Error("output should not contain trunk when there are no branches")
	}
	if out != "" {
		t.Errorf("expected empty output, got %q", out)
	}
}

func TestRenderTree_OrphanWithChildren(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"orphan":       "external/branch",
		"orphan-child": "orphan",
	}, "orphan-child")

	out := RenderTree(repo)
	if !strings.Contains(out, "external/branch") {
		t.Error("output should show untracked parent as header")
	}
	if !strings.Contains(out, "orphan") {
		t.Error("output should show orphan branch")
	}
	if !strings.Contains(out, "orphan-child") {
		t.Error("output should show orphan-child branch")
	}
	if !strings.Contains(out, "you are here") {
		t.Error("output should show 'you are here' on current branch")
	}
}

func TestRenderTree_GroupSeparation(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat-a": "main",
		"orphan": "external/branch",
	}, "")

	out := RenderTree(repo)
	// There should be a blank line separating the two groups
	// The groups are separated by "\n" between them
	lines := strings.Split(out, "\n")
	foundBlank := false
	for i := 1; i < len(lines)-1; i++ {
		if lines[i] == "" {
			foundBlank = true
			break
		}
	}
	if !foundBlank {
		t.Error("expected a blank line between different parent groups")
	}
}

func TestRenderTree_TreeStructure(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"root":    "main",
		"child-a": "root",
		"child-b": "root",
	}, "")

	out := RenderTree(repo)
	// root should use └── (last under main)
	// child-a should use ├── (not last under root)
	// child-b should use └── (last under root)
	if !strings.Contains(out, "├── ") {
		t.Error("expected ├── connector for non-last child")
	}
	if !strings.Contains(out, "└── ") {
		t.Error("expected └── connector for last child")
	}
}

func TestRenderBranchInfo_Basic(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat-a": "main",
	}, "feat-a")

	branch := repo.Branches["feat-a"]
	out := RenderBranchInfo(branch, repo)

	if !strings.Contains(out, "feat-a") {
		t.Error("output should contain branch name")
	}
	if !strings.Contains(out, "main") {
		t.Error("output should contain parent name")
	}
	if !strings.Contains(out, "none") {
		t.Error("output should show 'none' for children")
	}
}

func TestRenderBranchInfo_WithChildren(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"root":    "main",
		"child-a": "root",
		"child-b": "root",
	}, "root")

	branch := repo.Branches["root"]
	out := RenderBranchInfo(branch, repo)

	if !strings.Contains(out, "child-a") {
		t.Error("output should list child-a")
	}
	if !strings.Contains(out, "child-b") {
		t.Error("output should list child-b")
	}
}

func TestRenderBranchInfo_OrphanParent(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"orphan": "external/branch",
	}, "orphan")

	branch := repo.Branches["orphan"]
	out := RenderBranchInfo(branch, repo)

	if !strings.Contains(out, "external/branch") {
		t.Error("output should show untracked parent name")
	}
}
