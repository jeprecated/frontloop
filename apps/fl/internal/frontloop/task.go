package frontloop

import (
	"encoding/json"
	"fmt"
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

// ValidatePriority reports whether priority is one of the supported task
// priorities.
func ValidatePriority(priority string) error {
	if _, ok := PriorityPrefix[priority]; ok {
		return nil
	}
	return fmt.Errorf("invalid frontloop priority %q (use critical, high, medium, or low)", priority)
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

func formatTaskMarkdown(task Task) string {
	return fmt.Sprintf("---\ntitle: %s\npriority: %s\n---\n\n%s\n", yamlScalar(task.Title), yamlScalar(task.Priority), task.Body)
}

func yamlScalar(value string) string {
	if isPlainYAMLScalar(value) {
		return value
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%q", value)
	}
	return string(encoded)
}

func isPlainYAMLScalar(value string) bool {
	if value == "" || strings.TrimSpace(value) != value {
		return false
	}
	if strings.ContainsAny(value, "\r\n\t") {
		return false
	}
	if strings.Contains(value, ": ") || strings.Contains(value, " #") {
		return false
	}

	switch value {
	case "null", "Null", "NULL", "~", "true", "True", "TRUE", "false", "False", "FALSE":
		return false
	}

	if strings.ContainsAny(value[:1], "-?:!&*#{}[],|>%@`\"'") {
		return false
	}
	return true
}

// ParseFile reads a task markdown file and returns a Task.
func ParseFile(path string) (Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Task{}, err
	}

	var fm taskFrontmatter
	rest, err := parseTaskFrontmatter(string(data), &fm)
	if err != nil {
		return Task{}, fmt.Errorf("parse task frontmatter in %s: %w", path, err)
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

func parseTaskFrontmatter(content string, fm *taskFrontmatter) ([]byte, error) {
	rest, err := frontmatter.Parse(strings.NewReader(content), fm)
	if err == nil {
		return rest, nil
	}

	legacyFM, legacyRest, ok := parseLegacyUnquotedColonFrontmatter(content)
	if ok {
		*fm = legacyFM
		return legacyRest, nil
	}

	return nil, err
}

func parseLegacyUnquotedColonFrontmatter(content string) (taskFrontmatter, []byte, bool) {
	block, rest, ok := splitFrontmatterBlock(content)
	if !ok {
		return taskFrontmatter{}, nil, false
	}

	var fm taskFrontmatter
	recoveredUnquotedColon := false
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		key, value, found := strings.Cut(trimmed, ":")
		if !found {
			continue
		}

		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "title":
			fm.Title = value
			if hasUnquotedMappingColon(value) {
				recoveredUnquotedColon = true
			}
		case "priority":
			fm.Priority = value
		}
	}

	return fm, []byte(rest), recoveredUnquotedColon && fm.Title != "" && fm.Priority != ""
}

func splitFrontmatterBlock(content string) (string, string, bool) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.SplitAfter(normalized, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", "", false
	}

	var block strings.Builder
	restStart := len(lines[0])
	for _, line := range lines[1:] {
		restStart += len(line)
		if strings.TrimSpace(line) == "---" {
			return block.String(), normalized[restStart:], true
		}
		block.WriteString(line)
	}
	return "", "", false
}

func hasUnquotedMappingColon(value string) bool {
	if value == "" || strings.HasPrefix(value, "\"") || strings.HasPrefix(value, "'") {
		return false
	}
	return strings.Contains(value, ": ")
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
