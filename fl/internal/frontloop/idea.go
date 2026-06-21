package frontloop

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// IdeaFilename converts a description to a kebab-case .md filename.
// Non-alphanumeric characters are stripped and words are truncated to 6 max.
func IdeaFilename(description string) string {
	words := strings.Fields(strings.ToLower(description))

	var cleaned []string
	for _, w := range words {
		var sb strings.Builder
		for _, r := range w {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				sb.WriteRune(r)
			}
		}
		if sb.Len() > 0 {
			cleaned = append(cleaned, sb.String())
		}
	}

	if len(cleaned) > 7 {
		cleaned = cleaned[:7]
	}

	return strings.Join(cleaned, "-") + ".md"
}

// IdeaFilenameUnique returns a filename for the given description that does
// not already exist in the default epic's clarify/ directory. Legacy flat roots
// map the default epic to the top-level clarify/ directory.
func IdeaFilenameUnique(root, description string) string {
	filename, err := IdeaFilenameUniqueInEpic(root, DefaultEpicSlug, description)
	if err != nil {
		// Preserve the historical no-error API for callers that only need a
		// best-effort filename against a legacy clarify/ directory.
		return ideaFilenameUniqueInDir(filepath.Join(root, StatusClarify), description)
	}
	return filename
}

// IdeaFilenameUniqueInEpic returns a filename for the given description that
// does not already exist in the selected epic's clarify/ directory. If the base
// name exists, it appends -2, -3, etc. until a free slot is found.
func IdeaFilenameUniqueInEpic(root, epic, description string) (string, error) {
	if epic == "" {
		epic = DefaultEpicSlug
	}

	dirPath, err := clarifyDirForEpic(root, epic)
	if err != nil {
		return "", err
	}
	return ideaFilenameUniqueInDir(dirPath, description), nil
}

func ideaFilenameUniqueInDir(dirPath, description string) string {
	base := IdeaFilename(description)
	stem := base[:len(base)-3] // strip .md

	candidate := base
	for n := 2; ; n++ {
		path := filepath.Join(dirPath, candidate)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d.md", stem, n)
	}
}

// IdeaBody returns the markdown body for a skeleton idea task.
func IdeaBody(description string) string {
	return fmt.Sprintf("## Goal\n\n%s\n\n## Acceptance Criteria\n\n- [ ] TODO\n", description)
}

// CreateIdeaTask writes a skeleton idea task to the default epic's clarify/
// directory and returns its path. Legacy flat roots map the default epic to the
// top-level clarify/ directory.
func CreateIdeaTask(root, description, priority string) (string, error) {
	return CreateIdeaTaskInEpic(root, DefaultEpicSlug, description, priority)
}

// CreateIdeaTaskInEpic writes a skeleton idea task to an epic's clarify/
// directory and returns its path.
func CreateIdeaTaskInEpic(root, epic, description, priority string) (string, error) {
	if epic == "" {
		epic = DefaultEpicSlug
	}

	dirPath, err := clarifyDirForEpic(root, epic)
	if err != nil {
		return "", err
	}

	filename := ideaFilenameUniqueInDir(dirPath, description)
	task := Task{
		Title:    description,
		Priority: priority,
		Body:     IdeaBody(description),
		Filename: filename,
		Epic:     epic,
	}
	if err := CreateTaskInEpic(root, epic, task); err != nil {
		return "", err
	}
	return filepath.Join(dirPath, filename), nil
}
