package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jeprecated/frontloop/fl/internal/frontloop"
)

// State represents which UI state the model is in.
type State int

const (
	StateList       State = iota // browsing the task list
	StateSelectDest              // choosing a destination for a selected task
)

// item pairs a task with its queue directory.
type item struct {
	task frontloop.Task
	dir  string
}

// Model is the bubbletea model for the move TUI.
type Model struct {
	root       string
	all        map[string][]frontloop.Task
	items      []item
	cursor     int
	state      State
	destCursor int
	dests      []string
	statusMsg  string
}

// New creates a new Model from a loaded task map.
func New(root string, all map[string][]frontloop.Task) Model {
	m := Model{
		root: root,
		all:  all,
	}
	m.items = buildItems(all)
	return m
}

func buildItems(all map[string][]frontloop.Task) []item {
	var items []item
	for _, dir := range frontloop.Dirs {
		for _, t := range all[dir] {
			items = append(items, item{task: t, dir: dir})
		}
	}
	return items
}

// Cursor returns the current cursor position in StateList.
func (m Model) Cursor() int { return m.cursor }

// DestCursor returns the current cursor position in StateSelectDest.
func (m Model) DestCursor() int { return m.destCursor }

// State returns the current UI state.
func (m Model) State() State { return m.state }

// Destinations returns the valid move destinations for the selected task.
func (m Model) Destinations() []string { return m.dests }

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update handles incoming messages and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case moveResultMsg:
		return m.handleMoveResult(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case StateList:
		return m.handleListKey(msg)
	case StateSelectDest:
		return m.handleDestKey(msg)
	}
	return m, nil
}

func (m Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyDown:
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
		return m, nil
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case tea.KeyEnter:
		if len(m.items) == 0 {
			return m, nil
		}
		return m.openDestSelect()
	}
	switch {
	case msg.Type == tea.KeyRunes && string(msg.Runes) == "q":
		return m, tea.Quit
	case msg.Type == tea.KeyRunes && string(msg.Runes) == "j":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}
		return m, nil
	case msg.Type == tea.KeyRunes && string(msg.Runes) == "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	}
	return m, nil
}

func (m Model) openDestSelect() (Model, tea.Cmd) {
	selected := m.items[m.cursor]
	dests := make([]string, 0, len(frontloop.Dirs)-1)
	for _, d := range frontloop.Dirs {
		if d != selected.dir {
			dests = append(dests, d)
		}
	}
	m.state = StateSelectDest
	m.dests = dests
	m.destCursor = 0
	return m, nil
}

func (m Model) handleDestKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = StateList
		return m, nil
	case tea.KeyDown:
		if m.destCursor < len(m.dests)-1 {
			m.destCursor++
		}
		return m, nil
	case tea.KeyUp:
		if m.destCursor > 0 {
			m.destCursor--
		}
		return m, nil
	case tea.KeyEnter:
		return m, m.execMove()
	}
	switch {
	case msg.Type == tea.KeyRunes && string(msg.Runes) == "q":
		return m, tea.Quit
	case msg.Type == tea.KeyRunes && string(msg.Runes) == "j":
		if m.destCursor < len(m.dests)-1 {
			m.destCursor++
		}
		return m, nil
	case msg.Type == tea.KeyRunes && string(msg.Runes) == "k":
		if m.destCursor > 0 {
			m.destCursor--
		}
		return m, nil
	}
	return m, nil
}

// moveResultMsg carries the result of an async move operation.
type moveResultMsg struct {
	err    error
	newAll map[string][]frontloop.Task
}

func (m Model) execMove() tea.Cmd {
	selected := m.items[m.cursor]
	dest := m.dests[m.destCursor]
	root := m.root
	return func() tea.Msg {
		err := frontloop.MoveTask(root, selected.task, dest)
		if err != nil {
			return moveResultMsg{err: err}
		}
		newAll, err := frontloop.ListAll(root)
		return moveResultMsg{err: err, newAll: newAll}
	}
}

func (m Model) handleMoveResult(msg moveResultMsg) (Model, tea.Cmd) {
	if msg.err != nil {
		m.statusMsg = fmt.Sprintf("error: %v", msg.err)
		m.state = StateList
		return m, nil
	}
	m.all = msg.newAll
	m.items = buildItems(msg.newAll)
	m.state = StateList
	m.statusMsg = "moved"
	// Clamp cursor after refresh
	if m.cursor >= len(m.items) && len(m.items) > 0 {
		m.cursor = len(m.items) - 1
	}
	return m, nil
}

// View renders the TUI.
func (m Model) View() string {
	switch m.state {
	case StateSelectDest:
		return m.viewSelectDest()
	default:
		return m.viewList()
	}
}

var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Underline(true)
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	dimStyle     = lipgloss.NewStyle().Faint(true)
	dirStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3"))
)

func (m Model) viewList() string {
	if len(m.items) == 0 {
		return headerStyle.Render("fl move") + "\n\nQueue is empty — nothing to move.\n\nPress q to quit.\n"
	}

	var sb strings.Builder
	sb.WriteString(headerStyle.Render("fl move") + "\n\n")

	currentDir := ""
	for i, it := range m.items {
		if it.dir != currentDir {
			currentDir = it.dir
			sb.WriteString(dirStyle.Render(strings.ToUpper(currentDir)) + "\n")
		}
		prefix := "  "
		title := it.task.Title
		if i == m.cursor {
			prefix = "> "
			title = selectedStyle.Render(title)
		} else {
			title = dimStyle.Render(title)
		}
		sb.WriteString(fmt.Sprintf("  %s%s\n", prefix, title))
	}

	sb.WriteString("\n")
	if m.statusMsg != "" {
		sb.WriteString(dimStyle.Render(m.statusMsg) + "\n")
	}
	sb.WriteString(dimStyle.Render("j/k: navigate  Enter: move  q: quit"))

	return sb.String()
}

func (m Model) viewSelectDest() string {
	selected := m.items[m.cursor]
	var sb strings.Builder
	sb.WriteString(headerStyle.Render("Move: "+selected.task.Title) + "\n\n")
	sb.WriteString("Choose destination:\n\n")

	for i, d := range m.dests {
		if i == m.destCursor {
			sb.WriteString("> " + selectedStyle.Render(d) + "\n")
		} else {
			sb.WriteString("  " + dimStyle.Render(d) + "\n")
		}
	}

	sb.WriteString("\n" + dimStyle.Render("Enter: confirm  Esc: cancel  q: quit"))
	return sb.String()
}
