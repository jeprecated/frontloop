package cli

import (
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
