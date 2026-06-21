package frontloop

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var validEpicSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ArchivedEpic summarizes a successful epic archive operation.
type ArchivedEpic struct {
	Slug        string
	ActivePath  string
	ArchivePath string
	CompletedAt time.Time
}

// ValidateEpicSlug reports whether slug is valid for an active epic directory.
// Slugs must use lower-case letters, digits, and hyphens, and must not use
// reserved frontloop names.
func ValidateEpicSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("frontloop epic slug is required")
	}
	if IsReservedEpicName(slug) {
		return fmt.Errorf("reserved frontloop epic name %q", slug)
	}
	if !validEpicSlug.MatchString(slug) {
		return fmt.Errorf("invalid frontloop epic slug %q (use lower-case letters, digits, and hyphens; start and end with a letter or digit)", slug)
	}
	return nil
}

// CreateEpic creates a new active epic directory. Unlike EnsureEpic, it refuses
// to reuse an existing path so command callers do not accidentally overwrite or
// complete a preexisting epic directory.
func CreateEpic(root, slug string) error {
	if err := ValidateEpicSlug(slug); err != nil {
		return err
	}

	epicPath := filepath.Join(root, slug)
	if _, err := os.Stat(epicPath); err == nil {
		return fmt.Errorf("frontloop epic %q already exists", slug)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return EnsureEpic(root, slug)
}

// ArchiveEpic moves a completed active epic to .frontloop/_archive/ and marks
// its metadata archived. An epic is archivable only when clarify/, ready/, and
// in_progress/ contain no task markdown files. The default epic cannot be
// archived.
func ArchiveEpic(root, slug string) (ArchivedEpic, error) {
	return archiveEpic(root, slug, time.Now())
}

func archiveEpic(root, slug string, completedAt time.Time) (ArchivedEpic, error) {
	if !IsV2Root(root) {
		return ArchivedEpic{}, fmt.Errorf(".frontloop at %s is not a v2 epic layout", root)
	}
	if slug == DefaultEpicSlug {
		return ArchivedEpic{}, fmt.Errorf("cannot archive the %q epic", DefaultEpicSlug)
	}
	if err := ValidateEpicSlug(slug); err != nil {
		return ArchivedEpic{}, err
	}

	activePath := filepath.Join(root, slug)
	if !hasStatusDirs(activePath) {
		return ArchivedEpic{}, fmt.Errorf("frontloop epic %q does not exist", slug)
	}

	unfinished, err := unfinishedEpicTaskFiles(activePath)
	if err != nil {
		return ArchivedEpic{}, err
	}
	if len(unfinished) > 0 {
		return ArchivedEpic{}, fmt.Errorf("frontloop epic %q still has unfinished task(s): %s", slug, strings.Join(unfinished, ", "))
	}

	archiveRoot := filepath.Join(root, ArchiveDirName)
	if err := os.MkdirAll(archiveRoot, 0755); err != nil {
		return ArchivedEpic{}, err
	}

	archivePath := filepath.Join(archiveRoot, completedAt.Format("2006-01-02")+"-"+slug)
	if _, err := os.Stat(archivePath); err == nil {
		return ArchivedEpic{}, fmt.Errorf("archived epic destination already exists: %s", archivePath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return ArchivedEpic{}, err
	}

	if err := os.Rename(activePath, archivePath); err != nil {
		return ArchivedEpic{}, err
	}
	if err := markEpicMetadataArchived(filepath.Join(archivePath, "epic.md"), slug, completedAt); err != nil {
		return ArchivedEpic{}, err
	}

	return ArchivedEpic{
		Slug:        slug,
		ActivePath:  activePath,
		ArchivePath: archivePath,
		CompletedAt: completedAt,
	}, nil
}

func unfinishedEpicTaskFiles(epicPath string) ([]string, error) {
	var unfinished []string
	for _, status := range []string{StatusClarify, StatusReady, StatusInProgress} {
		entries, err := os.ReadDir(filepath.Join(epicPath, status))
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			unfinished = append(unfinished, filepath.Join(status, entry.Name()))
		}
	}
	return unfinished, nil
}

func markEpicMetadataArchived(path, slug string, completedAt time.Time) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		data = []byte(DefaultEpicMetadata(slug, completedAt))
	} else if err != nil {
		return err
	}

	updated := archiveEpicMetadata(string(data), completedAt)
	return os.WriteFile(path, []byte(updated), 0644)
}

func archiveEpicMetadata(content string, completedAt time.Time) string {
	date := completedAt.Format("2006-01-02")
	if !strings.HasPrefix(content, "---\n") {
		return fmt.Sprintf("---\nstatus: archived\ncompleted_at: %s\n---\n\n%s", date, content)
	}

	lines := strings.Split(content, "\n")
	closing := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closing = i
			break
		}
	}
	if closing == -1 {
		return fmt.Sprintf("---\nstatus: archived\ncompleted_at: %s\n---\n\n%s", date, content)
	}

	frontmatterLines := append([]string(nil), lines[1:closing]...)
	frontmatterLines = setFrontmatterValue(frontmatterLines, "status", "archived")
	frontmatterLines = setFrontmatterValue(frontmatterLines, "completed_at", date)

	updatedLines := make([]string, 0, len(lines)+2)
	updatedLines = append(updatedLines, "---")
	updatedLines = append(updatedLines, frontmatterLines...)
	updatedLines = append(updatedLines, lines[closing:]...)
	return strings.Join(updatedLines, "\n")
}

func setFrontmatterValue(lines []string, key, value string) []string {
	prefix := key + ":"
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			lines[i] = fmt.Sprintf("%s: %s", key, value)
			return lines
		}
	}
	return append(lines, fmt.Sprintf("%s: %s", key, value))
}
