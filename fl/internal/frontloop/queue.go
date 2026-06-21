package frontloop

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// StatusClarify is the queue for new tasks that need human review.
	StatusClarify = "clarify"
	// StatusReady is the queue for reviewed tasks ready to work.
	StatusReady = "ready"
	// StatusInProgress is the queue for the task currently being worked on.
	StatusInProgress = "in_progress"
	// StatusDone is the queue for completed tasks.
	StatusDone = "done"

	// DefaultEpicSlug is the built-in epic bucket for unscoped tasks.
	DefaultEpicSlug = "default"
	// ArchiveDirName is the reserved top-level directory for archived epics.
	ArchiveDirName = "_archive"
)

// Statuses is the ordered list of all task status directories.
var Statuses = []string{StatusClarify, StatusReady, StatusInProgress, StatusDone}

// Dirs is kept as the legacy name for callers that still treat statuses as
// top-level queue directories.
var Dirs = Statuses

// Epic represents an active frontloop epic directory.
type Epic struct {
	Slug     string
	Path     string
	Archived bool
}

// prefixedDirs are the directories that use priority prefixes on filenames.
var prefixedDirs = map[string]bool{
	StatusReady:      true,
	StatusInProgress: true,
}

// IsStatus reports whether status is one of the known task statuses.
func IsStatus(status string) bool {
	for _, s := range Statuses {
		if status == s {
			return true
		}
	}
	return false
}

// IsReservedEpicName reports whether name is reserved and cannot be an active
// epic slug. Top-level names beginning with '_' are reserved for frontloop
// internals, and status directory names are reserved for legacy layout support.
func IsReservedEpicName(name string) bool {
	return strings.HasPrefix(name, "_") || IsStatus(name)
}

// ListEpics returns all active epics in root, sorted by slug. Reserved
// top-level directories such as _archive and names beginning with '_' are
// ignored. Legacy flat queues are exposed as a compatibility-only default epic.
func ListEpics(root string) ([]Epic, error) {
	if IsLegacyRoot(root) && !IsV2Root(root) {
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return nil, err
		}
		return []Epic{{Slug: DefaultEpicSlug, Path: absRoot}}, nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var epics []Epic
	for _, entry := range entries {
		if !entry.IsDir() || IsReservedEpicName(entry.Name()) {
			continue
		}

		path := filepath.Join(root, entry.Name())
		if !hasStatusDirs(path) {
			continue
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		epics = append(epics, Epic{Slug: entry.Name(), Path: absPath})
	}

	sort.Slice(epics, func(i, j int) bool {
		return epics[i].Slug < epics[j].Slug
	})
	return epics, nil
}

// ListEpicDir returns all tasks for a specific epic and status, sorted by
// filename. In v2 roots the path is .frontloop/<epic>/<status>; in legacy roots
// only the default epic is supported and maps to .frontloop/<status>.
func ListEpicDir(root, epic, status string) ([]Task, error) {
	if !IsStatus(status) {
		return nil, fmt.Errorf("unknown frontloop status %q", status)
	}
	if epic == "" {
		epic = DefaultEpicSlug
	}

	dirPath, err := epicStatusDir(root, epic, status)
	if err != nil {
		return nil, err
	}
	return listTasksInDir(dirPath, epic, status)
}

// ListEpic returns all tasks for one epic grouped by status.
func ListEpic(root, epic string) (map[string][]Task, error) {
	if epic == "" {
		epic = DefaultEpicSlug
	}

	result := make(map[string][]Task, len(Statuses))
	for _, status := range Statuses {
		tasks, err := ListEpicDir(root, epic, status)
		if err != nil {
			return nil, err
		}
		result[status] = tasks
	}
	return result, nil
}

// ListAllByEpic returns all active tasks grouped by epic slug and status.
func ListAllByEpic(root string) (map[string]map[string][]Task, error) {
	epics, err := ListEpics(root)
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string][]Task, len(epics))
	for _, epic := range epics {
		grouped, err := ListEpic(root, epic.Slug)
		if err != nil {
			return nil, err
		}
		result[epic.Slug] = grouped
	}
	return result, nil
}

// ListDir returns all tasks for a status, sorted alphabetically by filename
// within each epic. This legacy compatibility wrapper aggregates that status
// across active epics for v2 roots and reads the top-level status directory for
// legacy roots.
func ListDir(root, status string) ([]Task, error) {
	if !IsStatus(status) {
		return nil, fmt.Errorf("unknown frontloop status %q", status)
	}

	if IsV2Root(root) {
		epics, err := ListEpics(root)
		if err != nil {
			return nil, err
		}
		var tasks []Task
		for _, epic := range epics {
			epicTasks, err := ListEpicDir(root, epic.Slug, status)
			if err != nil {
				return nil, err
			}
			tasks = append(tasks, epicTasks...)
		}
		return tasks, nil
	}

	dirPath := filepath.Join(root, status)
	return listTasksInDir(dirPath, DefaultEpicSlug, status)
}

// ListAll returns all tasks grouped by status. This legacy compatibility
// wrapper aggregates all active epics in v2 roots.
func ListAll(root string) (map[string][]Task, error) {
	result := make(map[string][]Task, len(Statuses))
	for _, status := range Statuses {
		tasks, err := ListDir(root, status)
		if err != nil {
			return nil, err
		}
		result[status] = tasks
	}
	return result, nil
}

// CreateTask writes a task file to task.Epic's clarify/ directory, defaulting
// to the default epic. Legacy flat roots map the default epic to the top-level
// clarify/ directory.
func CreateTask(root string, task Task) error {
	return CreateTaskInEpic(root, task.Epic, task)
}

// CreateTaskInEpic writes a task file to an epic's clarify/ directory.
// It validates that v2 epics already exist so typos do not silently create new
// task buckets.
func CreateTaskInEpic(root, epic string, task Task) error {
	if epic == "" {
		epic = DefaultEpicSlug
	}

	dirPath, err := clarifyDirForEpic(root, epic)
	if err != nil {
		return err
	}

	content := fmt.Sprintf("---\ntitle: %s\npriority: %s\n---\n\n%s\n", task.Title, task.Priority, task.Body)
	path := filepath.Join(dirPath, task.Filename)
	return os.WriteFile(path, []byte(content), 0644)
}

// MoveTask moves a task to destDir, adding or stripping the priority prefix as required.
// - ready/ and in_progress/: add priority prefix
// - clarify/ and done/: strip priority prefix
func MoveTask(root string, task Task, destDir string) error {
	var destName string
	if prefixedDirs[destDir] {
		prefix, ok := PriorityPrefix[task.Priority]
		if !ok {
			prefix = "4-" // default to low
		}
		destName = prefix + task.BaseName()
	} else {
		destName = task.BaseName()
	}

	destPath := filepath.Join(root, destDir, destName)
	return os.Rename(task.Path, destPath)
}

func listTasksInDir(dirPath, epic, status string) ([]Task, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	tasks := make([]Task, 0, len(names))
	for _, name := range names {
		task, err := ParseFile(filepath.Join(dirPath, name))
		if err != nil {
			return nil, err
		}
		task.Epic = epic
		task.Status = status
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func epicStatusDir(root, epic, status string) (string, error) {
	if IsV2Root(root) {
		if IsReservedEpicName(epic) {
			return "", fmt.Errorf("reserved frontloop epic name %q", epic)
		}
		return filepath.Join(root, epic, status), nil
	}

	if epic != DefaultEpicSlug {
		return "", fmt.Errorf("epic %q is unavailable in legacy frontloop layout", epic)
	}
	return filepath.Join(root, status), nil
}

func clarifyDirForEpic(root, epic string) (string, error) {
	if IsV2Root(root) {
		if err := ValidateEpicSlug(epic); err != nil {
			return "", err
		}

		epicPath := filepath.Join(root, epic)
		if !hasStatusDirs(epicPath) {
			return "", fmt.Errorf("frontloop epic %q does not exist (run `fl epic new %s` to create it)", epic, epic)
		}
		return filepath.Join(epicPath, StatusClarify), nil
	}

	if epic != DefaultEpicSlug {
		return "", fmt.Errorf("epic %q is unavailable in legacy frontloop layout", epic)
	}
	return filepath.Join(root, StatusClarify), nil
}
