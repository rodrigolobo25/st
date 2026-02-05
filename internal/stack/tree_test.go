package stack

import (
	"testing"
)

// helper to build a Repo with branches for testing.
func makeRepo(trunk string, branches map[string]string, current string) *Repo {
	repo := &Repo{
		Trunk:    trunk,
		Branches: make(map[string]*Branch),
	}
	for name, parent := range branches {
		repo.Branches[name] = &Branch{
			Name:    name,
			Parent:  parent,
			Current: name == current,
		}
	}
	return repo
}

// --- BuildTree ---

func TestBuildTree_TrunkParented(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat-a": "main",
		"feat-b": "main",
	}, "")
	BuildTree(repo)

	if len(repo.Stacks) != 2 {
		t.Fatalf("expected 2 stacks, got %d", len(repo.Stacks))
	}
	if repo.Stacks[0].Name != "feat-a" || repo.Stacks[1].Name != "feat-b" {
		t.Errorf("stacks sorted wrong: %s, %s", repo.Stacks[0].Name, repo.Stacks[1].Name)
	}
}

func TestBuildTree_UntrackedParent(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"my-branch": "some/untracked",
	}, "")
	BuildTree(repo)

	if len(repo.Stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(repo.Stacks))
	}
	if repo.Stacks[0].Name != "my-branch" {
		t.Errorf("expected root to be my-branch, got %s", repo.Stacks[0].Name)
	}
}

func TestBuildTree_MixedParents(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat-a":       "main",
		"feat-a-child": "feat-a",
		"orphan":       "external/branch",
	}, "")
	BuildTree(repo)

	if len(repo.Stacks) != 2 {
		t.Fatalf("expected 2 stacks, got %d", len(repo.Stacks))
	}
	// Sorted by name: feat-a, orphan
	if repo.Stacks[0].Name != "feat-a" {
		t.Errorf("expected first root feat-a, got %s", repo.Stacks[0].Name)
	}
	if repo.Stacks[1].Name != "orphan" {
		t.Errorf("expected second root orphan, got %s", repo.Stacks[1].Name)
	}
}

func TestBuildTree_ChildrenLinked(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"root":  "main",
		"child": "root",
	}, "")
	BuildTree(repo)

	root := repo.Branches["root"]
	if len(root.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children))
	}
	if root.Children[0].Name != "child" {
		t.Errorf("expected child named 'child', got %s", root.Children[0].Name)
	}
}

func TestBuildTree_ChildrenSorted(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"root":    "main",
		"child-z": "root",
		"child-a": "root",
		"child-m": "root",
	}, "")
	BuildTree(repo)

	root := repo.Branches["root"]
	if len(root.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(root.Children))
	}
	expected := []string{"child-a", "child-m", "child-z"}
	for i, name := range expected {
		if root.Children[i].Name != name {
			t.Errorf("child %d: expected %s, got %s", i, name, root.Children[i].Name)
		}
	}
}

func TestBuildTree_NoBranches(t *testing.T) {
	repo := makeRepo("main", map[string]string{}, "")
	BuildTree(repo)

	if len(repo.Stacks) != 0 {
		t.Errorf("expected 0 stacks, got %d", len(repo.Stacks))
	}
}

func TestBuildTree_DeepChain(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"a": "main",
		"b": "a",
		"c": "b",
		"d": "c",
	}, "")
	BuildTree(repo)

	if len(repo.Stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(repo.Stacks))
	}
	if repo.Stacks[0].Name != "a" {
		t.Errorf("expected root a, got %s", repo.Stacks[0].Name)
	}
	// Walk the chain
	b := repo.Branches["a"]
	for _, expected := range []string{"b", "c", "d"} {
		if len(b.Children) != 1 {
			t.Fatalf("expected 1 child at %s, got %d", b.Name, len(b.Children))
		}
		b = b.Children[0]
		if b.Name != expected {
			t.Errorf("expected %s, got %s", expected, b.Name)
		}
	}
	if len(b.Children) != 0 {
		t.Errorf("leaf should have no children, got %d", len(b.Children))
	}
}

func TestBuildTree_MultipleUntrackedParents(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat-on-main":   "main",
		"on-external-a":  "external/a",
		"on-external-b":  "external/b",
		"child-of-ext-a": "on-external-a",
	}, "")
	BuildTree(repo)

	// Roots: feat-on-main (parent=main), on-external-a (parent=external/a), on-external-b (parent=external/b)
	// child-of-ext-a is not a root since its parent is tracked
	if len(repo.Stacks) != 3 {
		t.Fatalf("expected 3 stacks, got %d", len(repo.Stacks))
	}
}

// --- CurrentBranch ---

func TestCurrentBranch_Found(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat-a": "main",
		"feat-b": "main",
	}, "feat-b")
	BuildTree(repo)

	b := CurrentBranch(repo)
	if b == nil {
		t.Fatal("expected current branch, got nil")
	}
	if b.Name != "feat-b" {
		t.Errorf("expected feat-b, got %s", b.Name)
	}
}

func TestCurrentBranch_NoCurrent(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat-a": "main",
	}, "")
	BuildTree(repo)

	b := CurrentBranch(repo)
	if b != nil {
		t.Errorf("expected nil, got %s", b.Name)
	}
}

// --- CurrentStack ---

func TestCurrentStack_TrunkParent(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"root":  "main",
		"child": "root",
	}, "child")
	BuildTree(repo)

	s := CurrentStack(repo)
	if s == nil {
		t.Fatal("expected stack root, got nil")
	}
	if s.Name != "root" {
		t.Errorf("expected root, got %s", s.Name)
	}
}

func TestCurrentStack_CurrentIsRoot(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"root": "main",
	}, "root")
	BuildTree(repo)

	s := CurrentStack(repo)
	if s == nil {
		t.Fatal("expected stack root, got nil")
	}
	if s.Name != "root" {
		t.Errorf("expected root, got %s", s.Name)
	}
}

func TestCurrentStack_UntrackedParent(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"orphan":       "external/branch",
		"orphan-child": "orphan",
	}, "orphan-child")
	BuildTree(repo)

	s := CurrentStack(repo)
	if s == nil {
		t.Fatal("expected stack root, got nil")
	}
	if s.Name != "orphan" {
		t.Errorf("expected orphan as root, got %s", s.Name)
	}
}

func TestCurrentStack_SingleOrphan(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"orphan": "external/branch",
	}, "orphan")
	BuildTree(repo)

	s := CurrentStack(repo)
	if s == nil {
		t.Fatal("expected stack root, got nil")
	}
	if s.Name != "orphan" {
		t.Errorf("expected orphan, got %s", s.Name)
	}
}

func TestCurrentStack_NoCurrent(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat-a": "main",
	}, "")
	BuildTree(repo)

	s := CurrentStack(repo)
	if s != nil {
		t.Errorf("expected nil, got %s", s.Name)
	}
}

// --- PathToTrunk ---

func TestPathToTrunk_NormalChain(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"a": "main",
		"b": "a",
		"c": "b",
	}, "")
	BuildTree(repo)

	path := PathToTrunk(repo, repo.Branches["c"])
	if len(path) != 3 {
		t.Fatalf("expected path len 3, got %d", len(path))
	}
	expected := []string{"a", "b", "c"}
	for i, name := range expected {
		if path[i].Name != name {
			t.Errorf("path[%d]: expected %s, got %s", i, name, path[i].Name)
		}
	}
}

func TestPathToTrunk_OrphanedBranch(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"orphan":       "external/branch",
		"orphan-child": "orphan",
	}, "")
	BuildTree(repo)

	path := PathToTrunk(repo, repo.Branches["orphan-child"])
	if len(path) != 2 {
		t.Fatalf("expected path len 2, got %d", len(path))
	}
	if path[0].Name != "orphan" || path[1].Name != "orphan-child" {
		t.Errorf("unexpected path: %s, %s", path[0].Name, path[1].Name)
	}
}

func TestPathToTrunk_RootBranch(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"root": "main",
	}, "")
	BuildTree(repo)

	path := PathToTrunk(repo, repo.Branches["root"])
	if len(path) != 1 {
		t.Fatalf("expected path len 1, got %d", len(path))
	}
	if path[0].Name != "root" {
		t.Errorf("expected root, got %s", path[0].Name)
	}
}

// --- AllBranchesInStack ---

func TestAllBranchesInStack_Linear(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"a": "main",
		"b": "a",
		"c": "b",
	}, "")
	BuildTree(repo)

	all := AllBranchesInStack(repo.Branches["a"])
	if len(all) != 3 {
		t.Fatalf("expected 3 branches, got %d", len(all))
	}
	expected := []string{"a", "b", "c"}
	for i, name := range expected {
		if all[i].Name != name {
			t.Errorf("all[%d]: expected %s, got %s", i, name, all[i].Name)
		}
	}
}

func TestAllBranchesInStack_Branching(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"root":    "main",
		"child-a": "root",
		"child-b": "root",
	}, "")
	BuildTree(repo)

	all := AllBranchesInStack(repo.Branches["root"])
	if len(all) != 3 {
		t.Fatalf("expected 3 branches, got %d", len(all))
	}
	if all[0].Name != "root" {
		t.Errorf("first should be root, got %s", all[0].Name)
	}
	// Children sorted alphabetically, so child-a then child-b
	if all[1].Name != "child-a" || all[2].Name != "child-b" {
		t.Errorf("children in wrong order: %s, %s", all[1].Name, all[2].Name)
	}
}

func TestAllBranchesInStack_Single(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"alone": "main",
	}, "")
	BuildTree(repo)

	all := AllBranchesInStack(repo.Branches["alone"])
	if len(all) != 1 {
		t.Fatalf("expected 1 branch, got %d", len(all))
	}
}

// --- Leaves ---

func TestLeaves_Linear(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"a": "main",
		"b": "a",
		"c": "b",
	}, "")
	BuildTree(repo)

	leaves := Leaves(repo.Branches["a"])
	if len(leaves) != 1 {
		t.Fatalf("expected 1 leaf, got %d", len(leaves))
	}
	if leaves[0].Name != "c" {
		t.Errorf("expected c, got %s", leaves[0].Name)
	}
}

func TestLeaves_MultipleBranches(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"root":    "main",
		"child-a": "root",
		"child-b": "root",
		"grand":   "child-a",
	}, "")
	BuildTree(repo)

	leaves := Leaves(repo.Branches["root"])
	if len(leaves) != 2 {
		t.Fatalf("expected 2 leaves, got %d", len(leaves))
	}
	// DFS order: grand (under child-a) then child-b
	if leaves[0].Name != "grand" || leaves[1].Name != "child-b" {
		t.Errorf("unexpected leaves: %s, %s", leaves[0].Name, leaves[1].Name)
	}
}

func TestLeaves_SingleNode(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"alone": "main",
	}, "")
	BuildTree(repo)

	leaves := Leaves(repo.Branches["alone"])
	if len(leaves) != 1 {
		t.Fatalf("expected 1 leaf, got %d", len(leaves))
	}
	if leaves[0].Name != "alone" {
		t.Errorf("expected alone, got %s", leaves[0].Name)
	}
}

// --- NavigateUp ---

func TestNavigateUp_OneStep(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"root":  "main",
		"child": "root",
	}, "root")
	BuildTree(repo)

	name, err := NavigateUp(repo, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "child" {
		t.Errorf("expected child, got %s", name)
	}
}

func TestNavigateUp_MultipleSteps(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"a": "main",
		"b": "a",
		"c": "b",
	}, "a")
	BuildTree(repo)

	name, err := NavigateUp(repo, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "c" {
		t.Errorf("expected c, got %s", name)
	}
}

func TestNavigateUp_AlreadyAtTop(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"leaf": "main",
	}, "leaf")
	BuildTree(repo)

	_, err := NavigateUp(repo, 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "already at the top of the stack" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNavigateUp_NotTracked(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat": "main",
	}, "")
	BuildTree(repo)

	_, err := NavigateUp(repo, 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "current branch is not tracked by st" {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- NavigateDown ---

func TestNavigateDown_OneStep(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"root":  "main",
		"child": "root",
	}, "child")
	BuildTree(repo)

	name, err := NavigateDown(repo, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "root" {
		t.Errorf("expected root, got %s", name)
	}
}

func TestNavigateDown_MultipleSteps(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"a": "main",
		"b": "a",
		"c": "b",
	}, "c")
	BuildTree(repo)

	name, err := NavigateDown(repo, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "a" {
		t.Errorf("expected a, got %s", name)
	}
}

func TestNavigateDown_AlreadyAtBottom_TrunkParent(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"root": "main",
	}, "root")
	BuildTree(repo)

	_, err := NavigateDown(repo, 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "already at the bottom of the stack" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNavigateDown_AlreadyAtBottom_UntrackedParent(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"orphan": "external/branch",
	}, "orphan")
	BuildTree(repo)

	_, err := NavigateDown(repo, 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "already at the bottom of the stack" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNavigateDown_NotTracked(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat": "main",
	}, "")
	BuildTree(repo)

	_, err := NavigateDown(repo, 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "current branch is not tracked by st" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNavigateDown_OrphanChain(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"orphan":       "external/branch",
		"orphan-child": "orphan",
	}, "orphan-child")
	BuildTree(repo)

	name, err := NavigateDown(repo, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "orphan" {
		t.Errorf("expected orphan, got %s", name)
	}

	// Now going down from orphan should fail (parent is untracked)
	repo2 := makeRepo("main", map[string]string{
		"orphan":       "external/branch",
		"orphan-child": "orphan",
	}, "orphan")
	BuildTree(repo2)

	_, err = NavigateDown(repo2, 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "already at the bottom of the stack" {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- NavigateTop ---

func TestNavigateTop_FromRoot(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"a": "main",
		"b": "a",
		"c": "b",
	}, "a")
	BuildTree(repo)

	name, err := NavigateTop(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "c" {
		t.Errorf("expected c, got %s", name)
	}
}

func TestNavigateTop_FromMiddle(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"a": "main",
		"b": "a",
		"c": "b",
	}, "b")
	BuildTree(repo)

	name, err := NavigateTop(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "c" {
		t.Errorf("expected c, got %s", name)
	}
}

func TestNavigateTop_AlreadyAtTop(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"leaf": "main",
	}, "leaf")
	BuildTree(repo)

	_, err := NavigateTop(repo)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "already at the top of the stack" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNavigateTop_NotTracked(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat": "main",
	}, "")
	BuildTree(repo)

	_, err := NavigateTop(repo)
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- NavigateBottom ---

func TestNavigateBottom_FromLeaf(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"a": "main",
		"b": "a",
		"c": "b",
	}, "c")
	BuildTree(repo)

	name, err := NavigateBottom(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "a" {
		t.Errorf("expected a, got %s", name)
	}
}

func TestNavigateBottom_FromMiddle(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"a": "main",
		"b": "a",
		"c": "b",
	}, "b")
	BuildTree(repo)

	name, err := NavigateBottom(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "a" {
		t.Errorf("expected a, got %s", name)
	}
}

func TestNavigateBottom_AlreadyAtBottom(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"root": "main",
	}, "root")
	BuildTree(repo)

	_, err := NavigateBottom(repo)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "already at the bottom of the stack" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNavigateBottom_OrphanStack(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"orphan":       "external/branch",
		"orphan-child": "orphan",
	}, "orphan-child")
	BuildTree(repo)

	name, err := NavigateBottom(repo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "orphan" {
		t.Errorf("expected orphan, got %s", name)
	}
}

func TestNavigateBottom_OrphanAlreadyAtBottom(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"orphan": "external/branch",
	}, "orphan")
	BuildTree(repo)

	_, err := NavigateBottom(repo)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "already at the bottom of the stack" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNavigateBottom_NotTracked(t *testing.T) {
	repo := makeRepo("main", map[string]string{
		"feat": "main",
	}, "")
	BuildTree(repo)

	_, err := NavigateBottom(repo)
	if err == nil {
		t.Fatal("expected error")
	}
}
