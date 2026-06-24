package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jeprecated/frontloop/apps/fl/internal/frontloop"
)

func writeStatsTask(t *testing.T, root, status, filename, title, priority string) {
	t.Helper()
	writeStatsTaskInEpic(t, root, frontloop.DefaultEpicSlug, status, filename, title, priority)
}

func writeStatsTaskInEpic(t *testing.T, root, epic, status, filename, title, priority string) {
	t.Helper()
	content := fmt.Sprintf("---\ntitle: %s\npriority: %s\n---\n\nBody.\n", title, priority)
	path := filepath.Join(root, ".frontloop", epic, status, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func resetStatsFlags(t *testing.T) {
	t.Helper()
	for name, value := range map[string]string{
		"no-color": "false",
		"epic":     "",
	} {
		flag := statsCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("stats flag %q is not registered", name)
		}
		if err := flag.Value.Set(value); err != nil {
			t.Fatalf("failed to reset stats flag %q: %v", name, err)
		}
		flag.Changed = false
	}
}

func runStats(t *testing.T, dir string) string {
	t.Helper()
	origNoColor := os.Getenv("NO_COLOR")
	os.Setenv("NO_COLOR", "1")               //nolint:errcheck
	defer os.Setenv("NO_COLOR", origNoColor) //nolint:errcheck

	output, err := runStatsCommand(t, dir, "stats")
	if err != nil {
		t.Fatalf("unexpected stats error: %v", err)
	}
	return output
}

func runStatsCommand(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	resetStatsFlags(t)

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir) //nolint:errcheck
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return out.String(), err
}

func TestStatsCmd_HidesEmptyEpicsWhenQueueEmpty(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
	output := runStats(t, dir)

	if !strings.Contains(output, "No active frontloop tasks.") {
		t.Errorf("expected no-active-tasks message in output:\n%s", output)
	}
	for _, unexpected := range []string{"EPIC: default (default)", "IN PROGRESS", "READY", "NEEDS CLARIFICATION", "DONE", "(empty)"} {
		if strings.Contains(output, unexpected) {
			t.Errorf("did not expect %q in empty stats output:\n%s", unexpected, output)
		}
	}
}

func TestStatsCmd_ShowsCountInSectionHeader(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
	writeStatsTask(t, dir, "ready", "2-task-one.md", "Task One", "high")
	writeStatsTask(t, dir, "ready", "3-task-two.md", "Task Two", "medium")

	output := runStats(t, dir)
	if !strings.Contains(output, "READY (2)") {
		t.Errorf("expected 'READY (2)' in output:\n%s", output)
	}
}

func TestStatsCmd_ShowsTaskLineFormat(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
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

func TestStatsCmd_ReadyTasksSortedAlphabeticallyWithinEpic(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
	writeStatsTask(t, dir, "ready", "3-beta-task.md", "Beta Task", "medium")
	writeStatsTask(t, dir, "ready", "1-alpha-task.md", "Alpha Task", "critical")

	output := runStats(t, dir)
	alphaIdx := strings.Index(output, "1-alpha-task")
	betaIdx := strings.Index(output, "3-beta-task")
	if alphaIdx == -1 || betaIdx == -1 {
		t.Fatalf("expected both tasks in output:\n%s", output)
	}
	if alphaIdx > betaIdx {
		t.Errorf("expected alpha-task before beta-task within the epic:\n%s", output)
	}
}

func TestStatsCmd_DoneSectionShowsMax5PerEpic(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
	root := filepath.Join(dir, ".frontloop")
	if err := frontloop.EnsureEpic(root, "checkout-redesign"); err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= 7; i++ {
		filename := fmt.Sprintf("task-%02d.md", i)
		writeStatsTask(t, dir, "done", filename, fmt.Sprintf("Default Task %d", i), "low")
		// stagger mod times so sort is deterministic
		path := filepath.Join(dir, ".frontloop", frontloop.DefaultEpicSlug, "done", filename)
		modTime := time.Now().Add(time.Duration(i) * time.Second)
		os.Chtimes(path, modTime, modTime) //nolint:errcheck
	}
	for i := 1; i <= 6; i++ {
		filename := fmt.Sprintf("checkout-task-%02d.md", i)
		writeStatsTaskInEpic(t, dir, "checkout-redesign", "done", filename, fmt.Sprintf("Checkout Task %d", i), "low")
		path := filepath.Join(dir, ".frontloop", "checkout-redesign", "done", filename)
		modTime := time.Now().Add(time.Duration(i) * time.Second)
		os.Chtimes(path, modTime, modTime) //nolint:errcheck
	}

	output := runStats(t, dir)
	for _, want := range []string{"... and 2 more", "... and 1 more"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in output:\n%s", want, output)
		}
	}
}

func TestStatsCmd_DoneShowsAllWhenFiveOrFewer(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
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

	_, err := runStatsCommand(t, dir, "stats")
	if err == nil {
		t.Error("expected error when .frontloop not found, got nil")
	}
}

func TestStatsCmd_ErrorWhenLegacyLayoutNeedsMigration(t *testing.T) {
	dir := makeCLIFrontloop(t)

	_, err := runStatsCommand(t, dir, "stats")
	if err == nil {
		t.Fatal("expected legacy-layout error")
	}
	if !strings.Contains(err.Error(), "legacy .frontloop layout") || !strings.Contains(err.Error(), "fl migrate epic-layout") {
		t.Errorf("expected migration hint, got: %v", err)
	}
}

func TestStatsCmd_InProgressSectionLabel(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
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
	dir := makeCLIV2Frontloop(t)
	writeStatsTask(t, dir, "clarify", "needs-review.md", "Needs Review", "medium")

	output := runStats(t, dir)
	if !strings.Contains(output, "NEEDS CLARIFICATION (1)") {
		t.Errorf("expected 'NEEDS CLARIFICATION (1)' in output:\n%s", output)
	}
	if !strings.Contains(output, "needs-review") {
		t.Errorf("expected task filename in output:\n%s", output)
	}
}

func TestStatsCmd_GroupsOutputByActiveEpicInSlugOrder(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
	root := filepath.Join(dir, ".frontloop")
	for _, slug := range []string{"checkout-redesign", "worker-runtime"} {
		if err := frontloop.EnsureEpic(root, slug); err != nil {
			t.Fatal(err)
		}
	}
	writeStatsTaskInEpic(t, dir, "checkout-redesign", "ready", "0100-render-review.md", "Render Review", "high")
	writeStatsTask(t, dir, "ready", "0200-default-task.md", "Default Task", "medium")
	writeStatsTaskInEpic(t, dir, "worker-runtime", "ready", "0100-heartbeat.md", "Heartbeat", "critical")

	output := runStats(t, dir)
	assertInOrder(t, output,
		"EPIC: checkout-redesign",
		"0100-render-review",
		statsEpicSeparator,
		"EPIC: default (default)",
		"0200-default-task",
		statsEpicSeparator,
		"EPIC: worker-runtime",
		"0100-heartbeat",
	)
	if count := strings.Count(output, "READY (1)"); count != 3 {
		t.Errorf("expected each epic to have its own READY (1) count, got %d in:\n%s", count, output)
	}
	if count := strings.Count(output, statsEpicSeparator); count != 2 {
		t.Errorf("expected separator between each epic, got %d in:\n%s", count, output)
	}
}

func TestStatsCmd_HidesEmptyEpicsWhenShowingAllEpics(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
	root := filepath.Join(dir, ".frontloop")
	for _, slug := range []string{"checkout-redesign", "empty-epic"} {
		if err := frontloop.EnsureEpic(root, slug); err != nil {
			t.Fatal(err)
		}
	}
	writeStatsTaskInEpic(t, dir, "checkout-redesign", "clarify", "0100-translate-copy.md", "Translate Copy", "high")

	output := runStats(t, dir)
	if !strings.Contains(output, "EPIC: checkout-redesign") || !strings.Contains(output, "0100-translate-copy") {
		t.Errorf("expected non-empty epic in output:\n%s", output)
	}
	for _, unexpected := range []string{"EPIC: default (default)", "EPIC: empty-epic"} {
		if strings.Contains(output, unexpected) {
			t.Errorf("did not expect empty epic %q in output:\n%s", unexpected, output)
		}
	}
}

func TestStatsCmd_EpicFilterShowsOnlySelectedEpic(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
	root := filepath.Join(dir, ".frontloop")
	if err := frontloop.EnsureEpic(root, "checkout-redesign"); err != nil {
		t.Fatal(err)
	}
	writeStatsTask(t, dir, "ready", "0200-default-task.md", "Default Task", "medium")
	writeStatsTaskInEpic(t, dir, "checkout-redesign", "ready", "0100-render-review.md", "Render Review", "high")

	output, err := runStatsCommand(t, dir, "stats", "--epic", "checkout-redesign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "EPIC: checkout-redesign") || !strings.Contains(output, "0100-render-review") {
		t.Errorf("expected selected epic in output:\n%s", output)
	}
	if !strings.Contains(output, "READY (1)") {
		t.Errorf("expected selected epic READY count in output:\n%s", output)
	}
	for _, unexpected := range []string{"EPIC: default (default)", "0200-default-task", "Default Task"} {
		if strings.Contains(output, unexpected) {
			t.Errorf("did not expect %q in filtered output:\n%s", unexpected, output)
		}
	}
}

func TestStatsCmd_EpicFilterShowsEmptySelectedEpic(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
	root := filepath.Join(dir, ".frontloop")
	if err := frontloop.EnsureEpic(root, "checkout-redesign"); err != nil {
		t.Fatal(err)
	}

	output, err := runStatsCommand(t, dir, "stats", "--epic", "checkout-redesign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(output, "EPIC: checkout-redesign") {
		t.Errorf("expected selected empty epic header in output:\n%s", output)
	}
	if count := strings.Count(output, "(empty)"); count != 4 {
		t.Errorf("expected selected empty epic to show all empty sections, got %d markers in:\n%s", count, output)
	}
}

func TestStatsCmd_EpicFilterRejectsMissingEpic(t *testing.T) {
	dir := makeCLIV2Frontloop(t)

	_, err := runStatsCommand(t, dir, "stats", "--epic", "checkout-redesign")
	if err == nil {
		t.Fatal("expected missing epic error")
	}
	if !strings.Contains(err.Error(), "frontloop epic \"checkout-redesign\" does not exist") || !strings.Contains(err.Error(), "fl epic list") {
		t.Errorf("expected helpful missing-epic error, got: %v", err)
	}
}

func TestStatsCmd_IgnoresArchivedEpics(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
	archiveReady := filepath.Join(dir, ".frontloop", frontloop.ArchiveDirName, "2026-06-21-old-epic", "ready")
	if err := os.MkdirAll(archiveReady, 0755); err != nil {
		t.Fatal(err)
	}
	archivedTask := filepath.Join(archiveReady, "0001-archived-task.md")
	content := []byte("---\ntitle: Archived Task\npriority: low\n---\n\nArchived body.\n")
	if err := os.WriteFile(archivedTask, content, 0644); err != nil {
		t.Fatal(err)
	}
	writeStatsTask(t, dir, "ready", "0001-active-task.md", "Active Task", "high")

	output := runStats(t, dir)
	if !strings.Contains(output, "Active Task") {
		t.Errorf("expected active task in output:\n%s", output)
	}
	for _, unexpected := range []string{"Archived Task", "0001-archived-task", "old-epic", frontloop.ArchiveDirName} {
		if strings.Contains(output, unexpected) {
			t.Errorf("archived epic data should not appear in output, found %q in:\n%s", unexpected, output)
		}
	}
}

func TestStatsCmd_NoColorFlag_Accepted(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
	if _, err := runStatsCommand(t, dir, "stats", "--no-color"); err != nil {
		t.Fatalf("--no-color flag not accepted: %v", err)
	}
}

func TestStatsCmd_NoColorFlag_OutputContainsExpectedSections(t *testing.T) {
	dir := makeCLIV2Frontloop(t)
	writeStatsTask(t, dir, "ready", "0001-my-task.md", "My Task", "high")

	output, err := runStatsCommand(t, dir, "stats", "--no-color")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "READY") {
		t.Errorf("expected READY section in output:\n%s", output)
	}
	if strings.Contains(output, "\x1b[") {
		t.Errorf("expected no ANSI escape codes with --no-color, got:\n%s", output)
	}
}

func assertInOrder(t *testing.T, output string, terms ...string) {
	t.Helper()
	lastIdx := -1
	for _, term := range terms {
		start := lastIdx + 1
		idx := strings.Index(output[start:], term)
		if idx == -1 {
			t.Fatalf("expected %q after previous term in output:\n%s", term, output)
		}
		lastIdx = start + idx
	}
}
