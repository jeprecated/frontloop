package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ohare93/frontloop/fl/internal/frontloop"
)

func TestInitCmd_IsRegistered(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "init" {
			return
		}
	}
	t.Error("init command not registered with root command")
}

func TestInitCmd_CreatesAllDirs(t *testing.T) {
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

	for _, sub := range frontloop.Dirs {
		path := filepath.Join(dir, ".frontloop", sub)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("directory not created: %s: %v", sub, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected directory, got file: %s", sub)
		}
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
	dir := makeCLIFrontloop(t) // creates .frontloop with all subdirs
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected no error on re-init, got: %v", err)
	}
}
