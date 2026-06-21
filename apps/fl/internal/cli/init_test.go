package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jeprecated/frontloop/apps/fl/internal/frontloop"
)

func TestInitCmd_IsRegistered(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "init" {
			return
		}
	}
	t.Error("init command not registered with root command")
}

func TestInitCmd_CreatesV2DefaultEpicAndArchive(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"init"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, sub := range frontloop.Statuses {
		path := filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, sub)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("directory not created: %s: %v", sub, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected directory, got file: %s", sub)
		}
	}

	archivePath := filepath.Join(dir, ".frontloop", frontloop.ArchiveDirName)
	info, err := os.Stat(archivePath)
	if err != nil {
		t.Fatalf("archive directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected archive directory, got file")
	}
}

func TestInitCmd_CreatesDefaultEpicMetadata(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, "epic.md"))
	if err != nil {
		t.Fatalf("default epic metadata not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "title: Default") {
		t.Errorf("expected default epic title, got: %q", content)
	}
	if !strings.Contains(content, "status: active") {
		t.Errorf("expected active status, got: %q", content)
	}
}

func TestInitCmd_PrintsConfirmation(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"init"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), ".frontloop") {
		t.Errorf("expected confirmation mentioning .frontloop, got: %q", out.String())
	}
}

func TestInitCmd_IdempotentWhenAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, ".frontloop")
	if err := frontloop.EnsureV2Root(root); err != nil {
		t.Fatal(err)
	}

	metadataPath := filepath.Join(root, frontloop.DefaultEpicSlug, "epic.md")
	customMetadata := []byte("---\ntitle: Custom Default\nstatus: active\n---\n\nExisting notes.\n")
	if err := os.WriteFile(metadataPath, customMetadata, 0644); err != nil {
		t.Fatal(err)
	}
	taskPath := filepath.Join(root, frontloop.DefaultEpicSlug, frontloop.StatusClarify, "existing-task.md")
	if err := os.WriteFile(taskPath, []byte("existing task"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error on re-init, got: %v", err)
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(customMetadata) {
		t.Errorf("init overwrote epic.md: got %q, want %q", string(data), string(customMetadata))
	}
	if _, err := os.Stat(taskPath); err != nil {
		t.Errorf("init removed existing task: %v", err)
	}
}

func TestInitCmd_LegacyTasksReportMigration(t *testing.T) {
	dir := makeCLIFrontloop(t)
	legacyTaskPath := filepath.Join(dir, ".frontloop", frontloop.StatusReady, "0001-existing.md")
	if err := os.WriteFile(legacyTaskPath, []byte("legacy task"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"init"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected init to report legacy task files")
	}
	if !strings.Contains(err.Error(), "fl migrate epic-layout") {
		t.Errorf("expected migration hint, got: %v", err)
	}
	if _, err := os.Stat(legacyTaskPath); err != nil {
		t.Errorf("legacy task should remain in place: %v", err)
	}
}
