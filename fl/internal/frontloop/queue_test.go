package frontloop_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ohare93/fl/internal/frontloop"
)

// makeQueue creates a full .frontloop directory tree and returns the root path.
func makeQueue(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	makeFrontloop(t, dir)
	return filepath.Join(dir, ".frontloop")
}

func writeTask(t *testing.T, root, dir, filename, content string) {
	t.Helper()
	path := filepath.Join(root, dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

const taskA = `---
title: Task A
priority: critical
---

Body of task A.
`

const taskB = `---
title: Task B
priority: high
---

Body of task B.
`

const taskC = `---
title: Task C
priority: medium
---

Body of task C.
`

func TestListDir_ReturnsTasksSorted(t *testing.T) {
	root := makeQueue(t)
	writeTask(t, root, "ready", "2-task-b.md", taskB)
	writeTask(t, root, "ready", "1-task-a.md", taskA)

	tasks, err := frontloop.ListDir(root, "ready")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}
	if tasks[0].Filename != "1-task-a.md" {
		t.Errorf("tasks[0] = %q, want %q", tasks[0].Filename, "1-task-a.md")
	}
	if tasks[1].Filename != "2-task-b.md" {
		t.Errorf("tasks[1] = %q, want %q", tasks[1].Filename, "2-task-b.md")
	}
}

func TestListDir_ReturnsEmptyForEmptyDir(t *testing.T) {
	root := makeQueue(t)

	tasks, err := frontloop.ListDir(root, "ready")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("got %d tasks, want 0", len(tasks))
	}
}

func TestListDir_SetsDirField(t *testing.T) {
	root := makeQueue(t)
	writeTask(t, root, "ready", "1-task-a.md", taskA)

	tasks, err := frontloop.ListDir(root, "ready")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tasks[0].Dir != filepath.Join(root, "ready") {
		t.Errorf("Dir = %q, want %q", tasks[0].Dir, filepath.Join(root, "ready"))
	}
}

func TestListAll_ReturnsTasksByDir(t *testing.T) {
	root := makeQueue(t)
	writeTask(t, root, "ready", "1-task-a.md", taskA)
	writeTask(t, root, "clarify", "task-b.md", taskB)
	writeTask(t, root, "in_progress", "2-task-c.md", taskC)

	all, err := frontloop.ListAll(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all["ready"]) != 1 {
		t.Errorf("ready: got %d, want 1", len(all["ready"]))
	}
	if len(all["clarify"]) != 1 {
		t.Errorf("clarify: got %d, want 1", len(all["clarify"]))
	}
	if len(all["in_progress"]) != 1 {
		t.Errorf("in_progress: got %d, want 1", len(all["in_progress"]))
	}
	if len(all["done"]) != 0 {
		t.Errorf("done: got %d, want 0", len(all["done"]))
	}
}

func TestCreateTask_WritesToClarify(t *testing.T) {
	root := makeQueue(t)
	task := frontloop.Task{
		Title:    "New Task",
		Priority: "high",
		Body:     "## Goal\n\nDo the thing.",
		Filename: "new-task.md",
	}

	if err := frontloop.CreateTask(root, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(root, "clarify", "new-task.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created at %q: %v", path, err)
	}
}

func TestCreateTask_FileContainsFrontmatter(t *testing.T) {
	root := makeQueue(t)
	task := frontloop.Task{
		Title:    "New Task",
		Priority: "high",
		Body:     "## Goal\n\nDo the thing.",
		Filename: "new-task.md",
	}

	if err := frontloop.CreateTask(root, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "clarify", "new-task.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if content == "" {
		t.Error("expected non-empty file content")
	}
	// Should contain title in frontmatter
	if !containsString(content, "New Task") {
		t.Errorf("file does not contain title: %q", content)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStringHelper(s, substr))
}

func containsStringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestMoveTask_ToReady_AddsPrefix(t *testing.T) {
	root := makeQueue(t)
	writeTask(t, root, "clarify", "task-a.md", taskA)

	tasks, _ := frontloop.ListDir(root, "clarify")
	task := tasks[0]

	if err := frontloop.MoveTask(root, task, "ready"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// clarify file should be gone
	if _, err := os.Stat(filepath.Join(root, "clarify", "task-a.md")); err == nil {
		t.Error("original file still exists in clarify")
	}

	// ready file should exist with prefix
	expectedName := frontloop.PriorityPrefix["critical"] + "task-a.md"
	if _, err := os.Stat(filepath.Join(root, "ready", expectedName)); err != nil {
		t.Errorf("file not found in ready as %q: %v", expectedName, err)
	}
}

func TestMoveTask_ToInProgress_AddsPrefix(t *testing.T) {
	root := makeQueue(t)
	writeTask(t, root, "clarify", "task-a.md", taskA)

	tasks, _ := frontloop.ListDir(root, "clarify")
	task := tasks[0]

	if err := frontloop.MoveTask(root, task, "in_progress"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedName := frontloop.PriorityPrefix["critical"] + "task-a.md"
	if _, err := os.Stat(filepath.Join(root, "in_progress", expectedName)); err != nil {
		t.Errorf("file not found in in_progress as %q: %v", expectedName, err)
	}
}

func TestMoveTask_ToDone_StripsPrefix(t *testing.T) {
	root := makeQueue(t)
	writeTask(t, root, "in_progress", "1-task-a.md", taskA)

	tasks, _ := frontloop.ListDir(root, "in_progress")
	task := tasks[0]

	if err := frontloop.MoveTask(root, task, "done"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "done", "task-a.md")); err != nil {
		t.Errorf("file not found in done as task-a.md: %v", err)
	}
}

func TestMoveTask_ToClarify_StripsPrefix(t *testing.T) {
	root := makeQueue(t)
	writeTask(t, root, "ready", "2-task-b.md", taskB)

	tasks, _ := frontloop.ListDir(root, "ready")
	task := tasks[0]

	if err := frontloop.MoveTask(root, task, "clarify"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "clarify", "task-b.md")); err != nil {
		t.Errorf("file not found in clarify as task-b.md: %v", err)
	}
}
