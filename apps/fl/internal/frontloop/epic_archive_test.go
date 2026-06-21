package frontloop

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const archiveTask = `---
title: Archived Task
priority: high
---

Done body.
`

func makeArchiveTestRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), ".frontloop")
	if err := EnsureV2Root(root); err != nil {
		t.Fatal(err)
	}
	return root
}

func writeArchiveTask(t *testing.T, root, epic, status, filename string) {
	t.Helper()
	path := filepath.Join(root, epic, status, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(archiveTask), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestArchiveEpic_MovesCompletedEpicAndUpdatesMetadata(t *testing.T) {
	root := makeArchiveTestRoot(t)
	if err := EnsureEpic(root, "checkout-redesign"); err != nil {
		t.Fatal(err)
	}
	writeArchiveTask(t, root, "checkout-redesign", StatusDone, "0100-archived-task.md")
	completedAt := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)

	archived, err := archiveEpic(root, "checkout-redesign", completedAt)
	if err != nil {
		t.Fatalf("unexpected archive error: %v", err)
	}

	wantArchivePath := filepath.Join(root, ArchiveDirName, "2026-06-21-checkout-redesign")
	if archived.ArchivePath != wantArchivePath {
		t.Errorf("archive path = %q, want %q", archived.ArchivePath, wantArchivePath)
	}
	if _, err := os.Stat(filepath.Join(root, "checkout-redesign")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("active epic should be moved away, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(wantArchivePath, StatusDone, "0100-archived-task.md")); err != nil {
		t.Fatalf("done task was not archived: %v", err)
	}

	metadata, err := os.ReadFile(filepath.Join(wantArchivePath, "epic.md"))
	if err != nil {
		t.Fatalf("archived epic metadata missing: %v", err)
	}
	content := string(metadata)
	for _, want := range []string{"title: checkout redesign", "slug: checkout-redesign", "status: archived", "completed_at: 2026-06-21"} {
		if !strings.Contains(content, want) {
			t.Errorf("metadata missing %q: %q", want, content)
		}
	}
}

func TestArchiveEpic_RefusesUnfinishedTaskFiles(t *testing.T) {
	for _, status := range []string{StatusClarify, StatusReady, StatusInProgress} {
		t.Run(status, func(t *testing.T) {
			root := makeArchiveTestRoot(t)
			if err := EnsureEpic(root, "checkout-redesign"); err != nil {
				t.Fatal(err)
			}
			writeArchiveTask(t, root, "checkout-redesign", status, "unfinished.md")

			_, err := archiveEpic(root, "checkout-redesign", time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC))
			if err == nil {
				t.Fatal("expected unfinished-task error")
			}
			if !strings.Contains(err.Error(), "unfinished") || !strings.Contains(err.Error(), status+string(filepath.Separator)+"unfinished.md") {
				t.Errorf("expected unfinished task path in error, got: %v", err)
			}
			if _, statErr := os.Stat(filepath.Join(root, "checkout-redesign")); statErr != nil {
				t.Errorf("active epic should remain after refusal: %v", statErr)
			}
		})
	}
}

func TestArchiveEpic_RefusesDefaultEpic(t *testing.T) {
	root := makeArchiveTestRoot(t)

	_, err := archiveEpic(root, DefaultEpicSlug, time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("expected default-epic refusal")
	}
	if !strings.Contains(err.Error(), "cannot archive") || !strings.Contains(err.Error(), DefaultEpicSlug) {
		t.Errorf("expected default refusal, got: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, DefaultEpicSlug)); statErr != nil {
		t.Errorf("default epic should remain after refusal: %v", statErr)
	}
}

func TestArchiveEpic_RefusesDestinationCollision(t *testing.T) {
	root := makeArchiveTestRoot(t)
	if err := EnsureEpic(root, "checkout-redesign"); err != nil {
		t.Fatal(err)
	}
	completedAt := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	collisionPath := filepath.Join(root, ArchiveDirName, "2026-06-21-checkout-redesign")
	if err := os.MkdirAll(collisionPath, 0755); err != nil {
		t.Fatal(err)
	}

	_, err := archiveEpic(root, "checkout-redesign", completedAt)
	if err == nil {
		t.Fatal("expected destination collision error")
	}
	if !strings.Contains(err.Error(), "destination already exists") || !strings.Contains(err.Error(), collisionPath) {
		t.Errorf("expected destination collision path in error, got: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "checkout-redesign")); statErr != nil {
		t.Errorf("active epic should remain after collision: %v", statErr)
	}
}

func TestArchiveEpic_RemovesEpicFromActiveListing(t *testing.T) {
	root := makeArchiveTestRoot(t)
	if err := EnsureEpic(root, "checkout-redesign"); err != nil {
		t.Fatal(err)
	}
	writeArchiveTask(t, root, "checkout-redesign", StatusDone, "0100-archived-task.md")

	if _, err := archiveEpic(root, "checkout-redesign", time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("unexpected archive error: %v", err)
	}

	epics, err := ListEpics(root)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}
	for _, epic := range epics {
		if epic.Slug == "checkout-redesign" || strings.Contains(epic.Path, ArchiveDirName) {
			t.Fatalf("archived epic should not be active: %+v", epics)
		}
	}
	if len(epics) != 1 || epics[0].Slug != DefaultEpicSlug {
		t.Fatalf("active epics = %+v, want only default", epics)
	}
}
