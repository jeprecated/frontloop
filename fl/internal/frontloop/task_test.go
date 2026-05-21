package frontloop_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jeprecated/frontloop/fl/internal/frontloop"
)

const sampleTask = `---
title: My Task Title
priority: high
---

This is the body of the task.

## Section

Some content here.
`

func writeTempTask(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseFile_ExtractsTitle(t *testing.T) {
	path := writeTempTask(t, "2-my-task.md", sampleTask)
	task, err := frontloop.ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "My Task Title" {
		t.Errorf("got title %q, want %q", task.Title, "My Task Title")
	}
}

func TestParseFile_ExtractsPriority(t *testing.T) {
	path := writeTempTask(t, "2-my-task.md", sampleTask)
	task, err := frontloop.ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Priority != "high" {
		t.Errorf("got priority %q, want %q", task.Priority, "high")
	}
}

func TestParseFile_StoresBody(t *testing.T) {
	path := writeTempTask(t, "2-my-task.md", sampleTask)
	task, err := frontloop.ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Body == "" {
		t.Error("expected non-empty body")
	}
}

func TestParseFile_SetsFilename(t *testing.T) {
	path := writeTempTask(t, "2-my-task.md", sampleTask)
	task, err := frontloop.ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Filename != "2-my-task.md" {
		t.Errorf("got filename %q, want %q", task.Filename, "2-my-task.md")
	}
}

func TestParseFile_SetsPath(t *testing.T) {
	path := writeTempTask(t, "2-my-task.md", sampleTask)
	task, err := frontloop.ParseFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Path != path {
		t.Errorf("got path %q, want %q", task.Path, path)
	}
}

func TestBaseName_StripsPrefix(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"1-my-task.md", "my-task.md"},
		{"2-my-task.md", "my-task.md"},
		{"3-my-task.md", "my-task.md"},
		{"4-my-task.md", "my-task.md"},
		{"my-task.md", "my-task.md"},
	}
	for _, tt := range tests {
		path := writeTempTask(t, tt.filename, sampleTask)
		task, _ := frontloop.ParseFile(path)
		if got := task.BaseName(); got != tt.want {
			t.Errorf("BaseName(%q) = %q, want %q", tt.filename, got, tt.want)
		}
	}
}

func TestPriorityPrefix_Maps(t *testing.T) {
	tests := []struct {
		priority string
		prefix   string
	}{
		{"critical", "1-"},
		{"high", "2-"},
		{"medium", "3-"},
		{"low", "4-"},
	}
	for _, tt := range tests {
		got, ok := frontloop.PriorityPrefix[tt.priority]
		if !ok {
			t.Errorf("PriorityPrefix[%q] not found", tt.priority)
			continue
		}
		if got != tt.prefix {
			t.Errorf("PriorityPrefix[%q] = %q, want %q", tt.priority, got, tt.prefix)
		}
	}
}

func TestPrefixPriority_Maps(t *testing.T) {
	tests := []struct {
		prefix   string
		priority string
	}{
		{"1-", "critical"},
		{"2-", "high"},
		{"3-", "medium"},
		{"4-", "low"},
	}
	for _, tt := range tests {
		got, ok := frontloop.PrefixPriority[tt.prefix]
		if !ok {
			t.Errorf("PrefixPriority[%q] not found", tt.prefix)
			continue
		}
		if got != tt.priority {
			t.Errorf("PrefixPriority[%q] = %q, want %q", tt.prefix, got, tt.priority)
		}
	}
}
