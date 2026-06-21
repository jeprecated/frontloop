package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jeprecated/frontloop/apps/fl/internal/frontloop"
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

func resetIdeaFlags(t *testing.T) {
	t.Helper()
	for name, value := range map[string]string{
		"priority": "medium",
		"epic":     frontloop.DefaultEpicSlug,
	} {
		flag := ideaCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("idea flag %q is not registered", name)
		}
		if err := flag.Value.Set(value); err != nil {
			t.Fatalf("failed to reset idea flag %q: %v", name, err)
		}
		flag.Changed = false
	}
}

func TestIdeaCmd_CreatesFileInClarify(t *testing.T) {
	resetIdeaFlags(t)
	dir := makeCLIV2Frontloop(t)
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

	path := filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, frontloop.StatusClarify, "add-retry-logic-to-the-api-client.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created at %q: %v", path, err)
	}
}

func TestIdeaCmd_DefaultPriorityMedium(t *testing.T) {
	resetIdeaFlags(t)
	dir := makeCLIV2Frontloop(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"idea", "add retry logic"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, frontloop.StatusClarify, "add-retry-logic.md"))
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if !strings.Contains(string(data), "priority: medium") {
		t.Errorf("expected priority medium in file: %q", string(data))
	}
}

func TestIdeaCmd_PriorityFlagOverridesDefault(t *testing.T) {
	resetIdeaFlags(t)
	dir := makeCLIV2Frontloop(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"idea", "--priority", "high", "add retry logic"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, frontloop.StatusClarify, "add-retry-logic.md"))
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if !strings.Contains(string(data), "priority: high") {
		t.Errorf("expected priority high in file: %q", string(data))
	}
}

func TestIdeaCmd_ShortPriorityFlag(t *testing.T) {
	resetIdeaFlags(t)
	dir := makeCLIV2Frontloop(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"idea", "-p", "critical", "urgent fix needed"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, frontloop.StatusClarify, "urgent-fix-needed.md"))
	if err != nil {
		t.Fatalf("file not found: %v", err)
	}
	if !strings.Contains(string(data), "priority: critical") {
		t.Errorf("expected priority critical in file: %q", string(data))
	}
}

func TestIdeaCmd_InvalidPriorityRejected(t *testing.T) {
	resetIdeaFlags(t)
	dir := makeCLIV2Frontloop(t)

	_, err := runCLIInDir(t, dir, "idea", "--priority", "urgent", "urgent fix needed")
	if err == nil {
		t.Fatal("expected invalid priority error")
	}
	if !strings.Contains(err.Error(), "critical, high, medium, or low") {
		t.Errorf("expected priority guidance, got: %v", err)
	}

	path := filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, frontloop.StatusClarify, "urgent-fix-needed.md")
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Errorf("invalid priority should not create task, stat err = %v", statErr)
	}
}

func TestIdeaCmd_MultiplePositionalArgsJoined(t *testing.T) {
	resetIdeaFlags(t)
	dir := makeCLIV2Frontloop(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"idea", "add", "retry", "logic"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, frontloop.StatusClarify, "add-retry-logic.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created at %q: %v", path, err)
	}
}

func TestIdeaCmd_PrintsConfirmationMessage(t *testing.T) {
	resetIdeaFlags(t)
	dir := makeCLIV2Frontloop(t)
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

func TestIdeaCmd_EpicFlagWritesSelectedEpicClarify(t *testing.T) {
	resetIdeaFlags(t)
	dir := makeCLIV2Frontloop(t)
	root := filepath.Join(dir, ".frontloop")
	if err := frontloop.EnsureEpic(root, "checkout-redesign"); err != nil {
		t.Fatal(err)
	}

	output, err := runCLIInDir(t, dir, "idea", "--epic", "checkout-redesign", "render review page")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(root, "checkout-redesign", frontloop.StatusClarify, "render-review-page.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created at %q: %v", path, err)
	}
	if !strings.Contains(output, path) {
		t.Errorf("expected confirmation to include selected epic path %q, got: %q", path, output)
	}
}

func TestIdeaCmd_EpicFlagRejectsMissingEpic(t *testing.T) {
	resetIdeaFlags(t)
	dir := makeCLIV2Frontloop(t)
	root := filepath.Join(dir, ".frontloop")

	_, err := runCLIInDir(t, dir, "idea", "--epic", "checkout-redesign", "render review page")
	if err == nil {
		t.Fatal("expected missing epic error")
	}
	if !strings.Contains(err.Error(), "frontloop epic \"checkout-redesign\" does not exist") || !strings.Contains(err.Error(), "fl epic new checkout-redesign") {
		t.Errorf("expected helpful missing-epic error, got: %v", err)
	}

	unexpected := filepath.Join(root, "checkout-redesign")
	if _, err := os.Stat(unexpected); !os.IsNotExist(err) {
		t.Errorf("missing epic should not be auto-created, stat err = %v", err)
	}
}

func TestIdeaCmd_DeduplicatesFilenamesPerEpic(t *testing.T) {
	resetIdeaFlags(t)
	dir := makeCLIV2Frontloop(t)
	root := filepath.Join(dir, ".frontloop")
	if err := frontloop.EnsureEpic(root, "checkout-redesign"); err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{
		{"idea", "add retry logic"},
		{"idea", "--epic", "checkout-redesign", "add retry logic"},
		{"idea", "--epic", "checkout-redesign", "add retry logic"},
	} {
		if _, err := runCLIInDir(t, dir, args...); err != nil {
			t.Fatalf("%v failed: %v", args, err)
		}
	}

	if _, err := os.Stat(filepath.Join(root, frontloop.DefaultEpicSlug, frontloop.StatusClarify, "add-retry-logic.md")); err != nil {
		t.Errorf("default epic base filename missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "checkout-redesign", frontloop.StatusClarify, "add-retry-logic.md")); err != nil {
		t.Errorf("selected epic base filename missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "checkout-redesign", frontloop.StatusClarify, "add-retry-logic-2.md")); err != nil {
		t.Errorf("selected epic duplicate filename missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, frontloop.DefaultEpicSlug, frontloop.StatusClarify, "add-retry-logic-2.md")); !os.IsNotExist(err) {
		t.Errorf("default epic should not see duplicate from selected epic, stat err = %v", err)
	}
}

func TestIdeaCmd_DoesNotWriteEpicFrontmatter(t *testing.T) {
	resetIdeaFlags(t)
	dir := makeCLIV2Frontloop(t)
	root := filepath.Join(dir, ".frontloop")
	if err := frontloop.EnsureEpic(root, "checkout-redesign"); err != nil {
		t.Fatal(err)
	}

	if _, err := runCLIInDir(t, dir, "idea", "--epic", "checkout-redesign", "render review page"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "checkout-redesign", frontloop.StatusClarify, "render-review-page.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "epic:") {
		t.Errorf("task frontmatter should not include epic membership: %q", string(data))
	}
}

func TestIdeaCmd_ErrorWhenLegacyLayoutNeedsMigration(t *testing.T) {
	resetIdeaFlags(t)
	dir := makeCLIFrontloop(t)

	_, err := runCLIInDir(t, dir, "idea", "add retry logic")
	if err == nil {
		t.Fatal("expected legacy-layout error")
	}
	if !strings.Contains(err.Error(), "legacy .frontloop layout") || !strings.Contains(err.Error(), "fl migrate epic-layout") {
		t.Errorf("expected migration hint, got: %v", err)
	}
}

func TestIdeaCmd_ErrorWhenNoFrontloopDir(t *testing.T) {
	resetIdeaFlags(t)
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
