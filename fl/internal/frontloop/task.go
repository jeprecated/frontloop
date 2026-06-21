package frontloop

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
)

// PriorityPrefix maps a priority name to its filename prefix.
var PriorityPrefix = map[string]string{
	"critical": "1-",
	"high":     "2-",
	"medium":   "3-",
	"low":      "4-",
}

// PrefixPriority maps a filename prefix to its priority name.
var PrefixPriority = map[string]string{
	"1-": "critical",
	"2-": "high",
	"3-": "medium",
	"4-": "low",
}

// Task represents a frontloop task file.
type Task struct {
	Title    string
	Priority string
	Body     string
	Filename string
	Epic     string
	Status   string
	Dir      string
	Path     string
}

type taskFrontmatter struct {
	Title    string `yaml:"title"`
	Priority string `yaml:"priority"`
}

// ParseFile reads a task markdown file and returns a Task.
func ParseFile(path string) (Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Task{}, err
	}

	var fm taskFrontmatter
	rest, err := frontmatter.Parse(strings.NewReader(string(data)), &fm)
	if err != nil {
		return Task{}, err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return Task{}, err
	}

	return Task{
		Title:    fm.Title,
		Priority: fm.Priority,
		Body:     strings.TrimSpace(string(rest)),
		Filename: filepath.Base(absPath),
		Dir:      filepath.Dir(absPath),
		Path:     absPath,
	}, nil
}

// BaseName returns the filename with any priority prefix (1-, 2-, 3-, 4-) stripped.
func (t Task) BaseName() string {
	name := t.Filename
	for prefix := range PrefixPriority {
		if strings.HasPrefix(name, prefix) {
			return name[len(prefix):]
		}
	}
	return name
}
