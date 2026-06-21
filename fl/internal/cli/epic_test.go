package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jeprecated/frontloop/fl/internal/frontloop"
)

func TestEpicCmd_IsRegistered(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use != "epic" {
			continue
		}
		for _, subcmd := range cmd.Commands() {
			if subcmd.Use == "new <slug>" {
				return
			}
		}
	}
	t.Error("epic new command not registered with root command")
}

func TestEpicNewCmd_CreatesActiveEpic(t *testing.T) {
	dir := makeCLIV2Frontloop(t)

	output, err := runCLIInDir(t, dir, "epic", "new", "checkout-redesign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "Created epic: checkout-redesign") {
		t.Errorf("expected creation confirmation, got: %q", output)
	}

	epicPath := filepath.Join(dir, ".frontloop", "checkout-redesign")
	for _, status := range frontloop.Statuses {
		info, err := os.Stat(filepath.Join(epicPath, status))
		if err != nil {
			t.Errorf("status directory %q not created: %v", status, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("status path %q is not a directory", status)
		}
	}

	metadata, err := os.ReadFile(filepath.Join(epicPath, "epic.md"))
	if err != nil {
		t.Fatalf("epic metadata not created: %v", err)
	}
	content := string(metadata)
	for _, want := range []string{"title: checkout redesign", "slug: checkout-redesign", "status: active"} {
		if !strings.Contains(content, want) {
			t.Errorf("epic metadata missing %q: %q", want, content)
		}
	}
}

func TestEpicNewCmd_RefusesDuplicateEpic(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
	if err := frontloop.EnsureEpic(filepath.Join(dir, ".frontloop"), "checkout-redesign"); err != nil {
		t.Fatal(err)
	}

	_, err := runCLIInDir(t, dir, "epic", "new", "checkout-redesign")
	if err == nil {
		t.Fatal("expected duplicate epic error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected duplicate error, got: %v", err)
	}
}

func TestEpicNewCmd_RefusesInvalidSlugs(t *testing.T) {
	dir := makeCLIV2Frontloop(t)

	for _, slug := range []string{"Checkout", "checkout_redesign", "-checkout", "checkout-", "checkout--redesign"} {
		t.Run(slug, func(t *testing.T) {
			args := []string{"epic", "new", slug}
			if strings.HasPrefix(slug, "-") {
				args = []string{"epic", "new", "--", slug}
			}
			_, err := runCLIInDir(t, dir, args...)
			if err == nil {
				t.Fatalf("expected invalid slug error for %q", slug)
			}
			if !strings.Contains(err.Error(), "lower-case letters, digits, and hyphens") {
				t.Errorf("expected slug rule in error, got: %v", err)
			}
		})
	}
}

func TestEpicNewCmd_RefusesReservedSlugs(t *testing.T) {
	dir := makeCLIV2Frontloop(t)

	for _, slug := range []string{frontloop.ArchiveDirName, "_scratch", frontloop.StatusReady, frontloop.StatusClarify, frontloop.StatusInProgress, frontloop.StatusDone} {
		t.Run(slug, func(t *testing.T) {
			_, err := runCLIInDir(t, dir, "epic", "new", slug)
			if err == nil {
				t.Fatalf("expected reserved slug error for %q", slug)
			}
			if !strings.Contains(err.Error(), "reserved frontloop epic name") {
				t.Errorf("expected reserved-name error, got: %v", err)
			}
		})
	}
}

func TestEpicNewCmd_ErrorWhenNoFrontloopDir(t *testing.T) {
	dir := t.TempDir()

	_, err := runCLIInDir(t, dir, "epic", "new", "checkout-redesign")
	if err == nil {
		t.Fatal("expected error when .frontloop is missing")
	}
	if !strings.Contains(err.Error(), "no .frontloop directory found") || !strings.Contains(err.Error(), "fl init") {
		t.Errorf("expected helpful no-root error, got: %v", err)
	}
}

func TestEpicNewCmd_ErrorWhenLegacyLayoutNeedsMigration(t *testing.T) {
	dir := makeCLIFrontloop(t)

	_, err := runCLIInDir(t, dir, "epic", "new", "checkout-redesign")
	if err == nil {
		t.Fatal("expected legacy-layout error")
	}
	if !strings.Contains(err.Error(), "legacy .frontloop layout") || !strings.Contains(err.Error(), "fl migrate epic-layout") {
		t.Errorf("expected migration hint, got: %v", err)
	}
}

func TestEpicListCmd_ListsActiveEpicsAndDefaultMarker(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
	root := filepath.Join(dir, ".frontloop")
	for _, slug := range []string{"checkout-redesign", "worker-runtime"} {
		if err := frontloop.EnsureEpic(root, slug); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, frontloop.ArchiveDirName, "2026-06-21-old-epic", frontloop.StatusReady), 0755); err != nil {
		t.Fatal(err)
	}

	output, err := runCLIInDir(t, dir, "epic", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"Active epics:", "checkout-redesign", "default (default)", "worker-runtime"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected output to contain %q, got: %q", want, output)
		}
	}
	if strings.Contains(output, frontloop.ArchiveDirName) || strings.Contains(output, "old-epic") {
		t.Errorf("archived epics should not appear in active list: %q", output)
	}
}

func TestEpicListCmd_ErrorWhenLegacyLayoutNeedsMigration(t *testing.T) {
	dir := makeCLIFrontloop(t)

	_, err := runCLIInDir(t, dir, "epic", "list")
	if err == nil {
		t.Fatal("expected legacy-layout error")
	}
	if !strings.Contains(err.Error(), "fl migrate epic-layout") {
		t.Errorf("expected migration hint, got: %v", err)
	}
}

func makeCLIV2Frontloop(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := frontloop.EnsureV2Root(filepath.Join(dir, ".frontloop")); err != nil {
		t.Fatal(err)
	}
	return dir
}
