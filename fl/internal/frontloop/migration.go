package frontloop

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MigrationResult summarizes a legacy-to-v2 layout migration.
type MigrationResult struct {
	TasksMoved int
}

// MigrationConflict identifies a legacy task that cannot be moved because the
// v2 default-epic destination already exists.
type MigrationConflict struct {
	Status     string
	Filename   string
	SourcePath string
	DestPath   string
}

// MigrationConflictError reports one or more migration destination conflicts.
type MigrationConflictError struct {
	Conflicts []MigrationConflict
}

func (e MigrationConflictError) Error() string {
	if len(e.Conflicts) == 0 {
		return "legacy frontloop migration conflict"
	}
	if len(e.Conflicts) == 1 {
		conflict := e.Conflicts[0]
		return fmt.Sprintf("legacy frontloop migration conflict: %s already exists for %s/%s", conflict.DestPath, conflict.Status, conflict.Filename)
	}
	first := e.Conflicts[0]
	return fmt.Sprintf("legacy frontloop migration conflicts: %d destination files already exist (first: %s for %s/%s)", len(e.Conflicts), first.DestPath, first.Status, first.Filename)
}

// EnsureV2Root creates the v2 frontloop root structure under root. It is
// idempotent and never overwrites an existing default epic.md.
func EnsureV2Root(root string) error {
	if err := EnsureEpic(root, DefaultEpicSlug); err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(root, ArchiveDirName), 0755)
}

// EnsureEpic creates an active epic directory with epic.md and all status
// directories. Existing epic.md files are preserved.
func EnsureEpic(root, slug string) error {
	if slug == "" {
		slug = DefaultEpicSlug
	}
	if IsReservedEpicName(slug) {
		return fmt.Errorf("reserved frontloop epic name %q", slug)
	}

	epicPath := filepath.Join(root, slug)
	for _, status := range Statuses {
		if err := os.MkdirAll(filepath.Join(epicPath, status), 0755); err != nil {
			return err
		}
	}

	epicFile := filepath.Join(epicPath, "epic.md")
	if _, err := os.Stat(epicFile); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	content := DefaultEpicMetadata(slug, time.Now())
	return os.WriteFile(epicFile, []byte(content), 0644)
}

// DefaultEpicMetadata returns the initial epic.md content for slug.
func DefaultEpicMetadata(slug string, createdAt time.Time) string {
	title := strings.TrimSpace(strings.ReplaceAll(slug, "-", " "))
	if slug == DefaultEpicSlug {
		title = "Default"
	}
	if title == "" {
		title = "Default"
	}

	return fmt.Sprintf("---\ntitle: %s\nstatus: active\ncreated_at: %s\ncompleted_at:\n---\n\n## Goal\n\n", title, createdAt.Format("2006-01-02"))
}

// HasLegacyTasks reports whether root contains task markdown files in the
// legacy top-level status directories.
func HasLegacyTasks(root string) (bool, error) {
	tasks, err := legacyTaskFiles(root)
	if err != nil {
		return false, err
	}
	return len(tasks) > 0, nil
}

// MigrateLegacyToEpicLayout moves legacy top-level status-directory task files
// into the default epic. It preserves filenames and file contents, creates the
// v2 root structure as needed, and refuses to move anything if a destination
// conflict is found.
func MigrateLegacyToEpicLayout(root string) (MigrationResult, error) {
	legacyTasks, err := legacyTaskFiles(root)
	if err != nil {
		return MigrationResult{}, err
	}

	conflicts := make([]MigrationConflict, 0)
	for _, task := range legacyTasks {
		destPath := filepath.Join(root, DefaultEpicSlug, task.status, task.filename)
		if _, err := os.Stat(destPath); err == nil {
			conflicts = append(conflicts, MigrationConflict{
				Status:     task.status,
				Filename:   task.filename,
				SourcePath: task.sourcePath,
				DestPath:   destPath,
			})
		} else if !errors.Is(err, os.ErrNotExist) {
			return MigrationResult{}, err
		}
	}
	if len(conflicts) > 0 {
		return MigrationResult{}, MigrationConflictError{Conflicts: conflicts}
	}

	if err := EnsureV2Root(root); err != nil {
		return MigrationResult{}, err
	}

	result := MigrationResult{}
	for _, task := range legacyTasks {
		destPath := filepath.Join(root, DefaultEpicSlug, task.status, task.filename)
		if err := os.Rename(task.sourcePath, destPath); err != nil {
			return result, err
		}
		result.TasksMoved++
	}
	return result, nil
}

type legacyTaskFile struct {
	status     string
	filename   string
	sourcePath string
}

func legacyTaskFiles(root string) ([]legacyTaskFile, error) {
	var tasks []legacyTaskFile
	for _, status := range Statuses {
		dirPath := filepath.Join(root, status)
		entries, err := os.ReadDir(dirPath)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			tasks = append(tasks, legacyTaskFile{
				status:     status,
				filename:   entry.Name(),
				sourcePath: filepath.Join(dirPath, entry.Name()),
			})
		}
	}
	return tasks, nil
}
