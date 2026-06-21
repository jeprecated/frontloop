package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jeprecated/frontloop/fl/internal/frontloop"
)

func TestMigrateEpicLayoutCmd_IsRegistered(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use != "migrate" {
			continue
		}
		for _, subcmd := range cmd.Commands() {
			if subcmd.Use == "epic-layout" {
				return
			}
		}
	}
	t.Error("migrate epic-layout command not registered with root command")
}

func TestMigrateEpicLayoutCmd_MovesLegacyTasksPreservingContents(t *testing.T) {
	dir := makeCLIFrontloop(t)
	readyContent := []byte("---\ntitle: Ready\npriority: high\n---\n\nReady body.\n")
	clarifyContent := []byte("---\ntitle: Clarify\npriority: medium\n---\n\nClarify body.\n")
	writeFile(t, filepath.Join(dir, ".frontloop", frontloop.StatusReady, "0001-ready.md"), readyContent)
	writeFile(t, filepath.Join(dir, ".frontloop", frontloop.StatusClarify, "clarify.md"), clarifyContent)

	output, err := runCLIInDir(t, dir, "migrate", "epic-layout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Migrated 2 legacy task") {
		t.Errorf("expected migrated count in output, got: %q", output)
	}

	assertFileContent(t, filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, frontloop.StatusReady, "0001-ready.md"), readyContent)
	assertFileContent(t, filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, frontloop.StatusClarify, "clarify.md"), clarifyContent)
	if _, err := os.Stat(filepath.Join(dir, ".frontloop", frontloop.StatusReady, "0001-ready.md")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("legacy ready task should be moved away, stat err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".frontloop", frontloop.ArchiveDirName)); err != nil {
		t.Errorf("archive directory not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, "epic.md")); err != nil {
		t.Errorf("default epic metadata not created: %v", err)
	}
}

func TestMigrateEpicLayoutCmd_RefusesConflictsBeforeMoving(t *testing.T) {
	dir := makeCLIFrontloop(t)
	legacyReady := []byte("legacy ready")
	destReady := []byte("v2 ready")
	legacyClarify := []byte("legacy clarify")
	writeFile(t, filepath.Join(dir, ".frontloop", frontloop.StatusReady, "same.md"), legacyReady)
	writeFile(t, filepath.Join(dir, ".frontloop", frontloop.StatusClarify, "only-legacy.md"), legacyClarify)
	writeFile(t, filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, frontloop.StatusReady, "same.md"), destReady)

	_, err := runCLIInDir(t, dir, "migrate", "epic-layout")
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if !strings.Contains(err.Error(), "conflict") || !strings.Contains(err.Error(), "same.md") {
		t.Errorf("expected helpful conflict error, got: %v", err)
	}

	assertFileContent(t, filepath.Join(dir, ".frontloop", frontloop.StatusReady, "same.md"), legacyReady)
	assertFileContent(t, filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, frontloop.StatusReady, "same.md"), destReady)
	assertFileContent(t, filepath.Join(dir, ".frontloop", frontloop.StatusClarify, "only-legacy.md"), legacyClarify)
	if _, err := os.Stat(filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, frontloop.StatusClarify, "only-legacy.md")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("non-conflicting task should not move when a conflict exists, stat err = %v", err)
	}
}

func TestMigrateEpicLayoutCmd_EmptyLegacyQueueCreatesV2Layout(t *testing.T) {
	dir := makeCLIFrontloop(t)

	output, err := runCLIInDir(t, dir, "migrate", "epic-layout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "No legacy tasks to migrate") {
		t.Errorf("expected no-op migration output, got: %q", output)
	}
	if !frontloop.IsV2Root(filepath.Join(dir, ".frontloop")) {
		t.Fatal("expected empty legacy queue to be initialized as v2")
	}
	if _, err := os.Stat(filepath.Join(dir, ".frontloop", frontloop.ArchiveDirName)); err != nil {
		t.Errorf("archive directory not created: %v", err)
	}
}

func TestMigrateEpicLayoutCmd_NoLegacyTasksInV2RootIsSafe(t *testing.T) {
	dir := t.TempDir()
	if err := frontloop.EnsureV2Root(filepath.Join(dir, ".frontloop")); err != nil {
		t.Fatal(err)
	}

	output, err := runCLIInDir(t, dir, "migrate", "epic-layout")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "No legacy tasks to migrate") {
		t.Errorf("expected no-op migration output, got: %q", output)
	}
}

func runCLIInDir(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return out.String(), err
}

func writeFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
}

func assertFileContent(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	if string(got) != string(want) {
		t.Fatalf("%s content = %q, want %q", path, string(got), string(want))
	}
}
