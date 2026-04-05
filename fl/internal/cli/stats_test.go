package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeStatsTask(t *testing.T, root, dir, filename, title, priority string) {
	t.Helper()
	content := fmt.Sprintf("---\ntitle: %s\npriority: %s\n---\n\nBody.\n", title, priority)
	path := filepath.Join(root, ".frontloop", dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func runStats(t *testing.T, dir string) string {
	t.Helper()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	origNoColor := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer os.Setenv("NO_COLOR", origNoColor)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"stats"})
	rootCmd.Execute() //nolint:errcheck
	return out.String()
}

func TestStatsCmd_ShowsEmptySectionsWhenQueueEmpty(t *testing.T) {
	dir := makeCLIFrontloop(t)
	output := runStats(t, dir)

	for _, section := range []string{"IN PROGRESS", "READY", "NEEDS CLARIFICATION", "DONE"} {
		if !strings.Contains(output, section) {
			t.Errorf("expected section %q in output:\n%s", section, output)
		}
	}
	if count := strings.Count(output, "(empty)"); count < 4 {
		t.Errorf("expected 4 '(empty)' markers, got %d in output:\n%s", count, output)
	}
}

func TestStatsCmd_ShowsCountInSectionHeader(t *testing.T) {
	dir := makeCLIFrontloop(t)
	writeStatsTask(t, dir, "ready", "2-task-one.md", "Task One", "high")
	writeStatsTask(t, dir, "ready", "3-task-two.md", "Task Two", "medium")

	output := runStats(t, dir)
	if !strings.Contains(output, "READY (2)") {
		t.Errorf("expected 'READY (2)' in output:\n%s", output)
	}
}

func TestStatsCmd_ShowsTaskLineFormat(t *testing.T) {
	dir := makeCLIFrontloop(t)
	writeStatsTask(t, dir, "ready", "2-my-task.md", "My Task Title", "high")

	output := runStats(t, dir)
	if !strings.Contains(output, "2-my-task") {
		t.Errorf("expected filename without .md in output:\n%s", output)
	}
	if !strings.Contains(output, "[high]") {
		t.Errorf("expected [high] priority label in output:\n%s", output)
	}
	if !strings.Contains(output, "My Task Title") {
		t.Errorf("expected title in output:\n%s", output)
	}
}

func TestStatsCmd_ReadyTasksSortedAlphabetically(t *testing.T) {
	dir := makeCLIFrontloop(t)
	writeStatsTask(t, dir, "ready", "3-beta-task.md", "Beta Task", "medium")
	writeStatsTask(t, dir, "ready", "1-alpha-task.md", "Alpha Task", "critical")

	output := runStats(t, dir)
	alphaIdx := strings.Index(output, "1-alpha-task")
	betaIdx := strings.Index(output, "3-beta-task")
	if alphaIdx == -1 || betaIdx == -1 {
		t.Fatalf("expected both tasks in output:\n%s", output)
	}
	if alphaIdx > betaIdx {
		t.Errorf("expected alpha-task before beta-task (alphabetical order):\n%s", output)
	}
}

func TestStatsCmd_DoneSectionShowsMax5(t *testing.T) {
	dir := makeCLIFrontloop(t)
	for i := 1; i <= 7; i++ {
		filename := fmt.Sprintf("task-%02d.md", i)
		writeStatsTask(t, dir, "done", filename, fmt.Sprintf("Task %d", i), "low")
		// stagger mod times so sort is deterministic
		path := filepath.Join(dir, ".frontloop", "done", filename)
		t := time.Now().Add(time.Duration(i) * time.Second)
		os.Chtimes(path, t, t) //nolint:errcheck
	}

	output := runStats(t, dir)
	if !strings.Contains(output, "... and 2 more") {
		t.Errorf("expected '... and 2 more' in output:\n%s", output)
	}
}

func TestStatsCmd_DoneShowsAllWhenFiveOrFewer(t *testing.T) {
	dir := makeCLIFrontloop(t)
	for i := 1; i <= 5; i++ {
		writeStatsTask(t, dir, "done", fmt.Sprintf("task-%d.md", i), fmt.Sprintf("Task %d", i), "low")
	}

	output := runStats(t, dir)
	if strings.Contains(output, "... and") {
		t.Errorf("expected no truncation with 5 done tasks, got:\n%s", output)
	}
}

func TestStatsCmd_ErrorWhenNoFrontloopDir(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	os.Chdir(dir)           //nolint:errcheck

	origNoColor := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")
	defer os.Setenv("NO_COLOR", origNoColor)

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"stats"})
	err := rootCmd.Execute()
	if err == nil {
		t.Error("expected error when .frontloop not found, got nil")
	}
}

func TestStatsCmd_InProgressSectionLabel(t *testing.T) {
	dir := makeCLIFrontloop(t)
	writeStatsTask(t, dir, "in_progress", "1-current-task.md", "Current Task", "critical")

	output := runStats(t, dir)
	if !strings.Contains(output, "IN PROGRESS (1)") {
		t.Errorf("expected 'IN PROGRESS (1)' in output:\n%s", output)
	}
	if !strings.Contains(output, "1-current-task") {
		t.Errorf("expected task filename in output:\n%s", output)
	}
}

func TestStatsCmd_ClarifySection(t *testing.T) {
	dir := makeCLIFrontloop(t)
	writeStatsTask(t, dir, "clarify", "needs-review.md", "Needs Review", "medium")

	output := runStats(t, dir)
	if !strings.Contains(output, "NEEDS CLARIFICATION (1)") {
		t.Errorf("expected 'NEEDS CLARIFICATION (1)' in output:\n%s", output)
	}
	if !strings.Contains(output, "needs-review") {
		t.Errorf("expected task filename in output:\n%s", output)
	}
}

func TestStatsCmd_NoColorFlag_Accepted(t *testing.T) {
	dir := makeCLIFrontloop(t)
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"stats", "--no-color"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("--no-color flag not accepted: %v", err)
	}
}

func TestStatsCmd_NoColorFlag_OutputContainsExpectedSections(t *testing.T) {
	dir := makeCLIFrontloop(t)
	writeStatsTask(t, dir, "ready", "0001-my-task.md", "My Task", "high")
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs([]string{"stats", "--no-color"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "READY") {
		t.Errorf("expected READY section in output:\n%s", output)
	}
	if strings.Contains(output, "\x1b[") {
		t.Errorf("expected no ANSI escape codes with --no-color, got:\n%s", output)
	}
}
