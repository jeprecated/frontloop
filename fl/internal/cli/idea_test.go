package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeCLIFrontloop(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"clarify", "ready", "in_progress", "done"} {
		if err := os.MkdirAll(filepath.Join(dir, ".frontloop", sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestIdeaCmd_CreatesFileInClarify(t *testing.T) {
	dir := makeCLIFrontloop(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"idea", "add retry logic to the API client"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(dir, ".frontloop", "clarify", "add-retry-logic-to-the-api-client.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created at %q: %v", path, err)
	}
}

func TestIdeaCmd_DefaultPriorityMedium(t *testing.T) {
	dir := makeCLIFrontloop(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"idea", "add retry logic"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".frontloop", "clarify", "add-retry-logic.md"))
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if !strings.Contains(string(data), "priority: medium") {
		t.Errorf("expected priority medium in file: %q", string(data))
	}
}

func TestIdeaCmd_PriorityFlagOverridesDefault(t *testing.T) {
	dir := makeCLIFrontloop(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"idea", "--priority", "high", "add retry logic"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".frontloop", "clarify", "add-retry-logic.md"))
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if !strings.Contains(string(data), "priority: high") {
		t.Errorf("expected priority high in file: %q", string(data))
	}
}

func TestIdeaCmd_ShortPriorityFlag(t *testing.T) {
	dir := makeCLIFrontloop(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"idea", "-p", "critical", "urgent fix needed"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".frontloop", "clarify", "urgent-fix-needed.md"))
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if !strings.Contains(string(data), "priority: critical") {
		t.Errorf("expected priority critical in file: %q", string(data))
	}
}

func TestIdeaCmd_MultiplePositionalArgsJoined(t *testing.T) {
	dir := makeCLIFrontloop(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"idea", "add", "retry", "logic"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(dir, ".frontloop", "clarify", "add-retry-logic.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created at %q: %v", path, err)
	}
}

func TestIdeaCmd_PrintsConfirmationMessage(t *testing.T) {
	dir := makeCLIFrontloop(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"idea", "add retry logic"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "add-retry-logic.md") {
		t.Errorf("expected confirmation with filename in output: %q", output)
	}
}

func TestIdeaCmd_ErrorWhenNoFrontloopDir(t *testing.T) {
	dir := t.TempDir() // no .frontloop here
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"idea", "add retry logic"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when .frontloop not found, got nil")
	}
}
