package cli

import (
	"os"
	"strings"
	"testing"
)

// TestMoveCmd_IsRegistered checks that the move command is registered with the root command.
func TestMoveCmd_IsRegistered(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "move" {
			return
		}
	}
	t.Error("move command not registered with root command")
}

func TestMoveCmd_ErrorWhenNoFrontloopDir(t *testing.T) {
	dir := t.TempDir() // no .frontloop here
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	rootCmd.SetArgs([]string{"move"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when .frontloop not found, got nil")
	}
}

func TestMoveCmd_ErrorWhenLegacyLayoutNeedsMigration(t *testing.T) {
	dir := makeCLIFrontloop(t)

	_, err := runCLIInDir(t, dir, "move")
	if err == nil {
		t.Fatal("expected legacy-layout error")
	}
	if !strings.Contains(err.Error(), "legacy .frontloop layout") || !strings.Contains(err.Error(), "fl migrate epic-layout") {
		t.Errorf("expected migration hint, got: %v", err)
	}
}
