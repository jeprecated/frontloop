package cli

import (
	"bytes"
	"strings"
	"testing"
)

func runCompletion(t *testing.T, shell string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"completion", shell})
	err := rootCmd.Execute()
	return out.String(), err
}

func TestCompletionCmd_BashProducesOutput(t *testing.T) {
	output, err := runCompletion(t, "bash")
	if err != nil {
		t.Fatalf("completion bash failed: %v", err)
	}
	if len(output) == 0 {
		t.Error("expected bash completion output, got empty string")
	}
}

func TestCompletionCmd_ZshProducesOutput(t *testing.T) {
	output, err := runCompletion(t, "zsh")
	if err != nil {
		t.Fatalf("completion zsh failed: %v", err)
	}
	if len(output) == 0 {
		t.Error("expected zsh completion output, got empty string")
	}
}

func TestCompletionCmd_FishProducesOutput(t *testing.T) {
	output, err := runCompletion(t, "fish")
	if err != nil {
		t.Fatalf("completion fish failed: %v", err)
	}
	if len(output) == 0 {
		t.Error("expected fish completion output, got empty string")
	}
}

func TestCompletionCmd_BashContainsFlCommands(t *testing.T) {
	output, err := runCompletion(t, "bash")
	if err != nil {
		t.Fatalf("completion bash failed: %v", err)
	}
	for _, cmd := range []string{"idea", "stats", "move"} {
		if !strings.Contains(output, cmd) {
			t.Errorf("expected bash completion to mention %q command, output:\n%s", cmd, output)
		}
	}
}
