package frontloop

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Dirs is the ordered list of all queue directories.
var Dirs = []string{"clarify", "ready", "in_progress", "done"}

// prefixedDirs are the directories that use priority prefixes on filenames.
var prefixedDirs = map[string]bool{
	"ready":       true,
	"in_progress": true,
}

// ListDir returns all tasks in a directory, sorted alphabetically by filename.
func ListDir(root, dir string) ([]Task, error) {
	dirPath := filepath.Join(root, dir)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		task, err := ParseFile(filepath.Join(dirPath, name))
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// ListAll returns all tasks grouped by directory.
func ListAll(root string) (map[string][]Task, error) {
	result := make(map[string][]Task)
	for _, dir := range Dirs {
		tasks, err := ListDir(root, dir)
		if err != nil {
			return nil, err
		}
		result[dir] = tasks
	}
	return result, nil
}

// CreateTask writes a task file to the clarify/ directory.
func CreateTask(root string, task Task) error {
	content := fmt.Sprintf("---\ntitle: %s\npriority: %s\n---\n\n%s\n", task.Title, task.Priority, task.Body)
	path := filepath.Join(root, "clarify", task.Filename)
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
