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
// not already exist in the clarify/ directory. If the base name exists, it
// appends -2, -3, etc. until a free slot is found.
func IdeaFilenameUnique(root, description string) string {
	base := IdeaFilename(description)
	stem := base[:len(base)-3] // strip .md

	candidate := base
	for n := 2; ; n++ {
		path := filepath.Join(root, "clarify", candidate)
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

// CreateIdeaTask writes a skeleton idea task to clarify/ and returns its path.
func CreateIdeaTask(root, description, priority string) (string, error) {
	filename := IdeaFilenameUnique(root, description)
	task := Task{
		Title:    description,
		Priority: priority,
		Body:     IdeaBody(description),
		Filename: filename,
	}
	if err := CreateTask(root, task); err != nil {
		return "", err
	}
	return filepath.Join(root, "clarify", filename), nil
}
