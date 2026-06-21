package frontloop_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jeprecated/frontloop/apps/fl/internal/frontloop"
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

func TestCreateTask_QuotesTitleWithColon(t *testing.T) {
	root := makeQueue(t)
	task := frontloop.Task{
		Title:    "Deferred: add real Twenty CRM sink package",
		Priority: "high",
		Body:     "Body.",
		Filename: "colon-title.md",
	}

	if err := frontloop.CreateTask(root, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(root, "clarify", "colon-title.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !containsString(string(data), `title: "Deferred: add real Twenty CRM sink package"`) {
		t.Errorf("expected quoted title frontmatter, got: %q", string(data))
	}

	parsed, err := frontloop.ParseFile(path)
	if err != nil {
		t.Fatalf("created task should parse: %v", err)
	}
	if parsed.Title != task.Title {
		t.Errorf("parsed title = %q, want %q", parsed.Title, task.Title)
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

func TestMoveTask_V2ClarifyToReadyPreservesEpicAndAddsDefaultOrderPrefix(t *testing.T) {
	dir := t.TempDir()
	root := makeEpicFrontloop(t, dir, frontloop.DefaultEpicSlug, "checkout-redesign")
	writeTask(t, root, filepath.Join("checkout-redesign", frontloop.StatusClarify), "render-review.md", taskB)

	tasks, err := frontloop.ListEpicDir(root, "checkout-redesign", frontloop.StatusClarify)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}

	if err := frontloop.MoveTask(root, tasks[0], frontloop.StatusReady); err != nil {
		t.Fatalf("unexpected move error: %v", err)
	}

	expectedName := frontloop.PriorityPrefix["high"] + "render-review.md"
	assertPathExists(t, filepath.Join(root, "checkout-redesign", frontloop.StatusReady, expectedName))
	assertPathMissing(t, filepath.Join(root, frontloop.DefaultEpicSlug, frontloop.StatusReady, expectedName))
	assertPathMissing(t, filepath.Join(root, "checkout-redesign", frontloop.StatusClarify, "render-review.md"))
}

func TestMoveTask_V2ReadyToInProgressPreservesEpicAndOrderPrefix(t *testing.T) {
	dir := t.TempDir()
	root := makeEpicFrontloop(t, dir, frontloop.DefaultEpicSlug, "checkout-redesign")
	writeTask(t, root, filepath.Join("checkout-redesign", frontloop.StatusReady), "0020-render-review.md", taskB)

	tasks, err := frontloop.ListEpicDir(root, "checkout-redesign", frontloop.StatusReady)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}

	if err := frontloop.MoveTask(root, tasks[0], frontloop.StatusInProgress); err != nil {
		t.Fatalf("unexpected move error: %v", err)
	}

	assertPathExists(t, filepath.Join(root, "checkout-redesign", frontloop.StatusInProgress, "0020-render-review.md"))
	assertPathMissing(t, filepath.Join(root, frontloop.DefaultEpicSlug, frontloop.StatusInProgress, "0020-render-review.md"))
	assertPathMissing(t, filepath.Join(root, "checkout-redesign", frontloop.StatusReady, "0020-render-review.md"))
}

func TestMoveTask_V2ReadyToDonePreservesEpicAndOrderPrefix(t *testing.T) {
	dir := t.TempDir()
	root := makeEpicFrontloop(t, dir, frontloop.DefaultEpicSlug, "checkout-redesign")
	writeTask(t, root, filepath.Join("checkout-redesign", frontloop.StatusReady), "0020-render-review.md", taskB)

	tasks, err := frontloop.ListEpicDir(root, "checkout-redesign", frontloop.StatusReady)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}

	if err := frontloop.MoveTask(root, tasks[0], frontloop.StatusDone); err != nil {
		t.Fatalf("unexpected move error: %v", err)
	}

	assertPathExists(t, filepath.Join(root, "checkout-redesign", frontloop.StatusDone, "0020-render-review.md"))
	assertPathMissing(t, filepath.Join(root, frontloop.DefaultEpicSlug, frontloop.StatusDone, "0020-render-review.md"))
	assertPathMissing(t, filepath.Join(root, "checkout-redesign", frontloop.StatusReady, "0020-render-review.md"))
}

func TestMoveTask_V2ReadyToClarifyPreservesEpicAndStripsOrderPrefix(t *testing.T) {
	dir := t.TempDir()
	root := makeEpicFrontloop(t, dir, frontloop.DefaultEpicSlug, "checkout-redesign")
	writeTask(t, root, filepath.Join("checkout-redesign", frontloop.StatusReady), "0020-render-review.md", taskB)

	tasks, err := frontloop.ListEpicDir(root, "checkout-redesign", frontloop.StatusReady)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}

	if err := frontloop.MoveTask(root, tasks[0], frontloop.StatusClarify); err != nil {
		t.Fatalf("unexpected move error: %v", err)
	}

	assertPathExists(t, filepath.Join(root, "checkout-redesign", frontloop.StatusClarify, "render-review.md"))
	assertPathMissing(t, filepath.Join(root, frontloop.DefaultEpicSlug, frontloop.StatusClarify, "render-review.md"))
	assertPathMissing(t, filepath.Join(root, "checkout-redesign", frontloop.StatusReady, "0020-render-review.md"))
}

func TestMoveTask_V2DuplicateFilenamesMoveByTaskPathAndEpic(t *testing.T) {
	dir := t.TempDir()
	root := makeEpicFrontloop(t, dir, frontloop.DefaultEpicSlug, "checkout-redesign")
	writeTask(t, root, filepath.Join(frontloop.DefaultEpicSlug, frontloop.StatusReady), "0020-render-review.md", taskA)
	writeTask(t, root, filepath.Join("checkout-redesign", frontloop.StatusReady), "0020-render-review.md", taskB)

	tasks, err := frontloop.ListEpicDir(root, frontloop.DefaultEpicSlug, frontloop.StatusReady)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}

	if err := frontloop.MoveTask(root, tasks[0], frontloop.StatusInProgress); err != nil {
		t.Fatalf("unexpected move error: %v", err)
	}

	assertPathExists(t, filepath.Join(root, frontloop.DefaultEpicSlug, frontloop.StatusInProgress, "0020-render-review.md"))
	assertPathExists(t, filepath.Join(root, "checkout-redesign", frontloop.StatusReady, "0020-render-review.md"))
	assertPathMissing(t, filepath.Join(root, frontloop.DefaultEpicSlug, frontloop.StatusReady, "0020-render-review.md"))
}

func TestListEpics_ReturnsActiveEpicsSortedAndIgnoresReserved(t *testing.T) {
	dir := t.TempDir()
	root := makeEpicFrontloop(t, dir, "default", "checkout-redesign", "worker-runtime")
	makeEpicDirs(t, filepath.Join(root, frontloop.ArchiveDirName, "2026-06-21-old-epic"))
	makeEpicDirs(t, filepath.Join(root, "_scratch"))
	makeEpicDirs(t, filepath.Join(root, frontloop.StatusReady))

	epics, err := frontloop.ListEpics(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := make([]string, len(epics))
	for i, epic := range epics {
		got[i] = epic.Slug
		if epic.Archived {
			t.Errorf("active epic %q was marked archived", epic.Slug)
		}
	}
	want := []string{"checkout-redesign", "default", "worker-runtime"}
	if len(got) != len(want) {
		t.Fatalf("epics = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("epics = %v, want %v", got, want)
			break
		}
	}
}

func TestListEpics_LegacyRootReturnsDefaultCompatibilityEpic(t *testing.T) {
	root := makeQueue(t)

	epics, err := frontloop.ListEpics(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(epics) != 1 {
		t.Fatalf("got %d epics, want 1", len(epics))
	}
	if epics[0].Slug != frontloop.DefaultEpicSlug {
		t.Errorf("epic slug = %q, want %q", epics[0].Slug, frontloop.DefaultEpicSlug)
	}
	if epics[0].Path != root {
		t.Errorf("epic path = %q, want %q", epics[0].Path, root)
	}
}

func TestListEpicDir_V2ReturnsSortedTasksWithEpicAndStatus(t *testing.T) {
	dir := t.TempDir()
	root := makeEpicFrontloop(t, dir, "default")
	writeTask(t, root, filepath.Join("default", "ready"), "0200-task-b.md", taskB)
	writeTask(t, root, filepath.Join("default", "ready"), "0100-task-a.md", taskA)

	tasks, err := frontloop.ListEpicDir(root, frontloop.DefaultEpicSlug, frontloop.StatusReady)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}
	if tasks[0].Filename != "0100-task-a.md" || tasks[1].Filename != "0200-task-b.md" {
		t.Errorf("task order = [%q, %q], want sorted filenames", tasks[0].Filename, tasks[1].Filename)
	}
	for _, task := range tasks {
		if task.Epic != frontloop.DefaultEpicSlug {
			t.Errorf("task epic = %q, want %q", task.Epic, frontloop.DefaultEpicSlug)
		}
		if task.Status != frontloop.StatusReady {
			t.Errorf("task status = %q, want %q", task.Status, frontloop.StatusReady)
		}
		if task.Dir != filepath.Join(root, "default", "ready") {
			t.Errorf("task dir = %q, want %q", task.Dir, filepath.Join(root, "default", "ready"))
		}
	}
}

func TestListAllByEpic_GroupsActiveTasksByEpicAndStatus(t *testing.T) {
	dir := t.TempDir()
	root := makeEpicFrontloop(t, dir, "default", "checkout-redesign")
	writeTask(t, root, filepath.Join("default", "clarify"), "task-a.md", taskA)
	writeTask(t, root, filepath.Join("checkout-redesign", "ready"), "0100-task-b.md", taskB)
	makeEpicDirs(t, filepath.Join(root, frontloop.ArchiveDirName, "2026-06-21-old-epic"))
	if err := os.WriteFile(filepath.Join(root, frontloop.ArchiveDirName, "2026-06-21-old-epic", "ready", "0001-archived.md"), []byte(taskC), 0644); err != nil {
		t.Fatal(err)
	}

	all, err := frontloop.ListAllByEpic(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("got %d epics, want 2", len(all))
	}
	if len(all["default"]["clarify"]) != 1 {
		t.Errorf("default clarify: got %d, want 1", len(all["default"]["clarify"]))
	}
	if len(all["checkout-redesign"]["ready"]) != 1 {
		t.Errorf("checkout ready: got %d, want 1", len(all["checkout-redesign"]["ready"]))
	}
	if _, ok := all[frontloop.ArchiveDirName]; ok {
		t.Error("archive directory was returned as an active epic")
	}
	if got := all["checkout-redesign"]["ready"][0].Epic; got != "checkout-redesign" {
		t.Errorf("task epic = %q, want checkout-redesign", got)
	}
}

func TestListDir_V2AggregatesStatusAcrossActiveEpics(t *testing.T) {
	dir := t.TempDir()
	root := makeEpicFrontloop(t, dir, "default", "checkout-redesign")
	writeTask(t, root, filepath.Join("default", "ready"), "0200-task-b.md", taskB)
	writeTask(t, root, filepath.Join("checkout-redesign", "ready"), "0100-task-a.md", taskA)

	tasks, err := frontloop.ListDir(root, frontloop.StatusReady)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(tasks))
	}
	if tasks[0].Epic != "checkout-redesign" || tasks[1].Epic != "default" {
		t.Errorf("task epics = [%q, %q], want active epics in slug order", tasks[0].Epic, tasks[1].Epic)
	}
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path %q to exist: %v", path, err)
	}
}

func assertPathMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected path %q to be missing, stat err = %v", path, err)
	}
}

func makeEpicDirs(t *testing.T, epicPath string) {
	t.Helper()
	for _, status := range frontloop.Statuses {
		if err := os.MkdirAll(filepath.Join(epicPath, status), 0755); err != nil {
			t.Fatal(err)
		}
	}
}
