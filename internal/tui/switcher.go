package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rodrigolobo/st/internal/stack"
)

// Panel represents which panel is active.
type Panel int

const (
	StackPanel Panel = iota
	BranchPanel
)

// SwitcherModel is the bubbletea model for the interactive switcher.
type SwitcherModel struct {
	repo          *stack.Repo
	stacks        []*stack.Branch // stack roots
	filteredIdx   []int           // indices into stacks after filtering
	selectedStack int             // index into filteredIdx
	selectedBranch int            // index into flat branch list for selected stack
	activePanel   Panel
	searchInput   textinput.Model
	searching     bool
	chosen        string // the branch to checkout (set on Enter)
	quitting      bool
	width         int
	height        int
}

// NewSwitcherModel creates a new switcher model.
func NewSwitcherModel(repo *stack.Repo) SwitcherModel {
	ti := textinput.New()
	ti.Placeholder = "search..."
	ti.CharLimit = 50

	m := SwitcherModel{
		repo:      repo,
		stacks:    repo.Stacks,
		searchInput: ti,
		width:     80,
		height:    24,
	}
	m.resetFilter()
	return m
}

func (m *SwitcherModel) resetFilter() {
	m.filteredIdx = make([]int, len(m.stacks))
	for i := range m.stacks {
		m.filteredIdx[i] = i
	}
}

func (m *SwitcherModel) applyFilter() {
	query := strings.ToLower(m.searchInput.Value())
	if query == "" {
		m.resetFilter()
		return
	}
	m.filteredIdx = nil
	for i, s := range m.stacks {
		if fuzzyMatch(s.Name, query) {
			m.filteredIdx = append(m.filteredIdx, i)
		}
	}
	if m.selectedStack >= len(m.filteredIdx) {
		m.selectedStack = max(0, len(m.filteredIdx)-1)
	}
}

func fuzzyMatch(s, query string) bool {
	s = strings.ToLower(s)
	qi := 0
	for i := 0; i < len(s) && qi < len(query); i++ {
		if s[i] == query[qi] {
			qi++
		}
	}
	return qi == len(query)
}

func (m SwitcherModel) currentStackBranches() []*stack.Branch {
	if len(m.filteredIdx) == 0 {
		return nil
	}
	root := m.stacks[m.filteredIdx[m.selectedStack]]
	return stack.AllBranchesInStack(root)
}

// Chosen returns the branch name the user selected, or empty string.
func (m SwitcherModel) Chosen() string {
	return m.chosen
}

func (m SwitcherModel) Init() tea.Cmd {
	return nil
}

func (m SwitcherModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.searching {
			switch msg.String() {
			case "esc":
				m.searching = false
				m.searchInput.Blur()
				m.resetFilter()
				return m, nil
			case "enter":
				m.searching = false
				m.searchInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.applyFilter()
				return m, cmd
			}
		}

		switch msg.String() {
		case "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "/":
			m.searching = true
			m.searchInput.Focus()
			return m, textinput.Blink

		case "tab":
			if m.activePanel == StackPanel {
				m.activePanel = BranchPanel
			} else {
				m.activePanel = StackPanel
			}
			return m, nil

		case "up", "k":
			if m.activePanel == StackPanel {
				if m.selectedStack > 0 {
					m.selectedStack--
					m.selectedBranch = 0
				}
			} else {
				if m.selectedBranch > 0 {
					m.selectedBranch--
				}
			}
			return m, nil

		case "down", "j":
			if m.activePanel == StackPanel {
				if m.selectedStack < len(m.filteredIdx)-1 {
					m.selectedStack++
					m.selectedBranch = 0
				}
			} else {
				branches := m.currentStackBranches()
				if m.selectedBranch < len(branches)-1 {
					m.selectedBranch++
				}
			}
			return m, nil

		case "enter":
			if m.activePanel == BranchPanel {
				branches := m.currentStackBranches()
				if m.selectedBranch < len(branches) {
					m.chosen = branches[m.selectedBranch].Name
					return m, tea.Quit
				}
			} else {
				// Switch to branch panel on enter from stack panel
				m.activePanel = BranchPanel
				m.selectedBranch = 0
				return m, nil
			}
			return m, nil
		}
	}

	return m, nil
}

func (m SwitcherModel) View() string {
	if m.quitting {
		return ""
	}

	panelHeight := m.height - 4 // leave room for search + borders
	if panelHeight < 5 {
		panelHeight = 5
	}
	leftWidth := m.width/3 - 2
	if leftWidth < 20 {
		leftWidth = 20
	}
	rightWidth := m.width - leftWidth - 6
	if rightWidth < 20 {
		rightWidth = 20
	}

	// Left panel: Stacks
	leftTitle := " Stacks "
	var leftContent strings.Builder
	for i, idx := range m.filteredIdx {
		s := m.stacks[idx]
		branchCount := len(stack.AllBranchesInStack(s))
		line := fmt.Sprintf("%s (%d)", s.Name, branchCount)

		if i == m.selectedStack {
			if m.activePanel == StackPanel {
				leftContent.WriteString(SelectedItemStyle.Render("> " + line))
			} else {
				leftContent.WriteString(NormalItemStyle.Bold(true).Render("> " + line))
			}
		} else {
			leftContent.WriteString(NormalItemStyle.Render("  " + line))
		}
		leftContent.WriteString("\n")
	}

	// Right panel: Branches in selected stack
	rightTitle := " Branches "
	var rightContent strings.Builder
	branches := m.currentStackBranches()
	if len(branches) > 0 {
		// Build a mini tree view
		root := m.stacks[m.filteredIdx[m.selectedStack]]
		renderSwitcherTree(&rightContent, root, m.repo, "", true, m.selectedBranch, m.activePanel == BranchPanel, &branchCounter{})
	}

	// Style panels
	var leftStyle, rightStyle lipgloss.Style
	if m.activePanel == StackPanel {
		leftStyle = ActiveBorderStyle.Width(leftWidth).Height(panelHeight)
		rightStyle = InactiveBorderStyle.Width(rightWidth).Height(panelHeight)
	} else {
		leftStyle = InactiveBorderStyle.Width(leftWidth).Height(panelHeight)
		rightStyle = ActiveBorderStyle.Width(rightWidth).Height(panelHeight)
	}

	leftPanel := leftStyle.Render(HeaderStyle.Render(leftTitle) + "\n" + leftContent.String())
	rightPanel := rightStyle.Render(HeaderStyle.Render(rightTitle) + "\n" + rightContent.String())

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Search bar
	searchBar := ""
	if m.searching {
		searchBar = "\n / " + m.searchInput.View()
	}

	help := DimStyle.Render("  ↑↓/jk: navigate • tab: switch panel • enter: select • /: search • q/esc: quit")

	return panels + searchBar + "\n" + help
}

type branchCounter struct {
	count int
}

func renderSwitcherTree(sb *strings.Builder, branch *stack.Branch, repo *stack.Repo, prefix string, isLast bool, selectedIdx int, isActive bool, counter *branchCounter) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	// For root, show trunk first
	if prefix == "" {
		nameStr := TrunkStyle.Render(repo.Trunk)
		sb.WriteString("  " + nameStr + "\n")
		prefix = "  "
	}

	idx := counter.count
	counter.count++

	var nameStr string
	isSelected := idx == selectedIdx

	if isSelected && isActive {
		nameStr = SelectedItemStyle.Render(branch.Name)
	} else if branch.Current {
		nameStr = CurrentBranchStyle.Render(branch.Name)
	} else {
		nameStr = NormalItemStyle.Render(branch.Name)
	}

	marker := ""
	if branch.Current {
		marker = DimStyle.Render(" ← here")
	}

	cursor := "  "
	if isSelected && isActive {
		cursor = "> "
	}

	sb.WriteString(cursor + prefix + connector + nameStr + marker + "\n")

	childPrefix := prefix
	if isLast {
		childPrefix += "    "
	} else {
		childPrefix += "│   "
	}

	for i, child := range branch.Children {
		childIsLast := i == len(branch.Children)-1
		renderSwitcherTree(sb, child, repo, childPrefix, childIsLast, selectedIdx, isActive, counter)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
