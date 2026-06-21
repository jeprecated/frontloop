package frontloop

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
)

// PriorityPrefix maps a priority name to its default four-digit ordering
// prefix. The ranges match the documented frontloop ordering convention.
var PriorityPrefix = map[string]string{
	"critical": "0001-",
	"high":     "2500-",
	"medium":   "5000-",
	"low":      "7500-",
}

// PrefixPriority maps known default ordering prefixes to their priority names.
// Short legacy prefixes are kept for backwards-compatible parsing/stripping.
var PrefixPriority = map[string]string{
	"0001-": "critical",
	"2500-": "high",
	"5000-": "medium",
	"7500-": "low",
	"1-":    "critical",
	"2-":    "high",
	"3-":    "medium",
	"4-":    "low",
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

// BaseName returns the filename with any numeric ordering prefix stripped.
func (t Task) BaseName() string {
	base, _ := stripOrderPrefix(t.Filename)
	return base
}

func hasOrderPrefix(name string) bool {
	_, ok := stripOrderPrefix(name)
	return ok
}

func stripOrderPrefix(name string) (string, bool) {
	dash := strings.Index(name, "-")
	if dash <= 0 {
		return name, false
	}

	for _, r := range name[:dash] {
		if r < '0' || r > '9' {
			return name, false
		}
	}
	return name[dash+1:], true
}
