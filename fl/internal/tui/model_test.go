package tui_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jeprecated/frontloop/fl/internal/frontloop"
	"github.com/jeprecated/frontloop/fl/internal/tui"
)

// makeTestQueue creates a temporary .frontloop root with tasks in given dirs.
func makeTestQueue(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range frontloop.Dirs {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func writeTestTask(t *testing.T, root, dir, filename, content string) {
	t.Helper()
	path := filepath.Join(root, dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

const taskAlpha = `---
title: Alpha Task
priority: high
---

Body of alpha.
`

const taskBeta = `---
title: Beta Task
priority: medium
---

Body of beta.
`

func TestModel_InitialCursorIsZero(t *testing.T) {
	root := makeTestQueue(t)
	writeTestTask(t, root, "ready", "2-alpha.md", taskAlpha)

	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	if m.Cursor() != 0 {
		t.Errorf("initial cursor = %d, want 0", m.Cursor())
	}
}

func TestModel_JKeyMovesCursorDown(t *testing.T) {
	root := makeTestQueue(t)
	writeTestTask(t, root, "ready", "2-alpha.md", taskAlpha)
	writeTestTask(t, root, "ready", "3-beta.md", taskBeta)

	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = next.(tui.Model)

	if m.Cursor() != 1 {
		t.Errorf("cursor after j = %d, want 1", m.Cursor())
	}
}

func TestModel_KKeyMovesCursorUp(t *testing.T) {
	root := makeTestQueue(t)
	writeTestTask(t, root, "ready", "2-alpha.md", taskAlpha)
	writeTestTask(t, root, "ready", "3-beta.md", taskBeta)

	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	// Move down first
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = next.(tui.Model)
	// Then move up
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = next.(tui.Model)

	if m.Cursor() != 0 {
		t.Errorf("cursor after j then k = %d, want 0", m.Cursor())
	}
}

func TestModel_CursorClampsAtZero(t *testing.T) {
	root := makeTestQueue(t)
	writeTestTask(t, root, "ready", "2-alpha.md", taskAlpha)

	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = next.(tui.Model)

	if m.Cursor() != 0 {
		t.Errorf("cursor at top with k = %d, want 0", m.Cursor())
	}
}

func TestModel_CursorClampsAtBottom(t *testing.T) {
	root := makeTestQueue(t)
	writeTestTask(t, root, "ready", "2-alpha.md", taskAlpha)

	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = next.(tui.Model)

	if m.Cursor() != 0 {
		t.Errorf("cursor at bottom with j = %d, want 0", m.Cursor())
	}
}

func TestModel_DownArrowMovesCursorDown(t *testing.T) {
	root := makeTestQueue(t)
	writeTestTask(t, root, "ready", "2-alpha.md", taskAlpha)
	writeTestTask(t, root, "ready", "3-beta.md", taskBeta)

	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(tui.Model)

	if m.Cursor() != 1 {
		t.Errorf("cursor after down arrow = %d, want 1", m.Cursor())
	}
}

func TestModel_UpArrowMovesCursorUp(t *testing.T) {
	root := makeTestQueue(t)
	writeTestTask(t, root, "ready", "2-alpha.md", taskAlpha)
	writeTestTask(t, root, "ready", "3-beta.md", taskBeta)

	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(tui.Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = next.(tui.Model)

	if m.Cursor() != 0 {
		t.Errorf("cursor after down then up = %d, want 0", m.Cursor())
	}
}

func TestModel_EnterOpensDestSelectState(t *testing.T) {
	root := makeTestQueue(t)
	writeTestTask(t, root, "ready", "2-alpha.md", taskAlpha)

	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(tui.Model)

	if m.State() != tui.StateSelectDest {
		t.Errorf("state after enter = %v, want StateSelectDest", m.State())
	}
}

func TestModel_EscReturnsToListState(t *testing.T) {
	root := makeTestQueue(t)
	writeTestTask(t, root, "ready", "2-alpha.md", taskAlpha)

	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(tui.Model)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(tui.Model)

	if m.State() != tui.StateList {
		t.Errorf("state after esc = %v, want StateList", m.State())
	}
}

func TestModel_DestinationsExcludeCurrentDir(t *testing.T) {
	root := makeTestQueue(t)
	writeTestTask(t, root, "ready", "2-alpha.md", taskAlpha)

	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(tui.Model)

	for _, d := range m.Destinations() {
		if d == "ready" {
			t.Error("destinations should not include current dir 'ready'")
		}
	}
}

func TestModel_QQuits(t *testing.T) {
	root := makeTestQueue(t)
	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	if cmd == nil {
		t.Error("expected quit command from 'q', got nil")
	}
}

func TestModel_CtrlCQuits(t *testing.T) {
	root := makeTestQueue(t)
	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	if cmd == nil {
		t.Error("expected quit command from ctrl+c, got nil")
	}
}

func TestModel_ViewShowsEmptyMessage(t *testing.T) {
	root := makeTestQueue(t)
	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	view := m.View()

	if !strings.Contains(view, "empty") {
		t.Errorf("view with no tasks should mention empty, got: %q", view)
	}
}

func TestModel_ViewShowsTaskTitle(t *testing.T) {
	root := makeTestQueue(t)
	writeTestTask(t, root, "ready", "2-alpha.md", taskAlpha)

	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	view := m.View()

	if !strings.Contains(view, "Alpha Task") {
		t.Errorf("view should contain task title 'Alpha Task', got: %q", view)
	}
}

func TestModel_ViewGroupsByDirectory(t *testing.T) {
	root := makeTestQueue(t)
	writeTestTask(t, root, "ready", "2-alpha.md", taskAlpha)
	writeTestTask(t, root, "clarify", "beta.md", taskBeta)

	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	view := strings.ToLower(m.View())

	if !strings.Contains(view, "ready") {
		t.Errorf("view should show 'ready' group header, got: %q", view)
	}
	if !strings.Contains(view, "clarify") {
		t.Errorf("view should show 'clarify' group header, got: %q", view)
	}
}

func TestModel_ViewShowsDestsInSelectState(t *testing.T) {
	root := makeTestQueue(t)
	writeTestTask(t, root, "ready", "2-alpha.md", taskAlpha)

	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(tui.Model)

	view := m.View()

	// Should show some destination options (clarify, in_progress, done)
	hasDestination := strings.Contains(view, "clarify") ||
		strings.Contains(view, "in_progress") ||
		strings.Contains(view, "done")
	if !hasDestination {
		t.Errorf("select dest view should show destinations, got: %q", view)
	}
}

func TestModel_DestCursorJMovesDown(t *testing.T) {
	root := makeTestQueue(t)
	writeTestTask(t, root, "ready", "2-alpha.md", taskAlpha)

	all, _ := frontloop.ListAll(root)
	m := tui.New(root, all)

	// Enter select mode
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(tui.Model)

	// Move dest cursor down
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = next.(tui.Model)

	if m.DestCursor() != 1 {
		t.Errorf("dest cursor after j = %d, want 1", m.DestCursor())
	}
}
