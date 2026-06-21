package frontloop_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jeprecated/frontloop/apps/fl/internal/frontloop"
)

func TestIdeaFilename_BasicDescription(t *testing.T) {
	got := frontloop.IdeaFilename("add retry logic to the API client")
	want := "add-retry-logic-to-the-api-client.md"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIdeaFilename_TruncatesTo7Words(t *testing.T) {
	got := frontloop.IdeaFilename("one two three four five six seven eight nine")
	want := "one-two-three-four-five-six-seven.md"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIdeaFilename_ExactlySixWords(t *testing.T) {
	got := frontloop.IdeaFilename("one two three four five six")
	want := "one-two-three-four-five-six.md"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIdeaFilename_StripsSpecialChars(t *testing.T) {
	got := frontloop.IdeaFilename("add retry logic!")
	want := "add-retry-logic.md"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIdeaFilename_LowercasesInput(t *testing.T) {
	got := frontloop.IdeaFilename("Add Retry Logic")
	want := "add-retry-logic.md"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIdeaFilename_SingleWord(t *testing.T) {
	got := frontloop.IdeaFilename("authentication")
	want := "authentication.md"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIdeaFilenameUnique_ReturnsBaseWhenNotExists(t *testing.T) {
	root := makeQueue(t)
	got := frontloop.IdeaFilenameUnique(root, "add retry logic")
	want := "add-retry-logic.md"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIdeaFilenameUnique_DeduplicatesWithSuffix2(t *testing.T) {
	root := makeQueue(t)
	writeTask(t, root, "clarify", "add-retry-logic.md", taskA)

	got := frontloop.IdeaFilenameUnique(root, "add retry logic")
	want := "add-retry-logic-2.md"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIdeaFilenameUnique_DeduplicatesWithSuffix3(t *testing.T) {
	root := makeQueue(t)
	writeTask(t, root, "clarify", "add-retry-logic.md", taskA)
	writeTask(t, root, "clarify", "add-retry-logic-2.md", taskA)

	got := frontloop.IdeaFilenameUnique(root, "add retry logic")
	want := "add-retry-logic-3.md"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIdeaBody_ContainsGoalSection(t *testing.T) {
	body := frontloop.IdeaBody("add retry logic to the API client")
	if !containsString(body, "## Goal") {
		t.Errorf("body missing ## Goal section: %q", body)
	}
	if !containsString(body, "add retry logic to the API client") {
		t.Errorf("body missing description: %q", body)
	}
}

func TestIdeaBody_ContainsAcceptanceCriteria(t *testing.T) {
	body := frontloop.IdeaBody("add retry logic to the API client")
	if !containsString(body, "## Acceptance Criteria") {
		t.Errorf("body missing ## Acceptance Criteria section: %q", body)
	}
	if !containsString(body, "- [ ] TODO") {
		t.Errorf("body missing TODO placeholder: %q", body)
	}
}

func TestCreateIdeaTask_CreatesFileInClarify(t *testing.T) {
	root := makeQueue(t)
	path, err := frontloop.CreateIdeaTask(root, "add retry logic to the API client", "medium")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created at %q: %v", path, err)
	}
}

func TestCreateIdeaTask_FileInClarifyDir(t *testing.T) {
	root := makeQueue(t)
	path, err := frontloop.CreateIdeaTask(root, "add retry logic to the API client", "medium")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dir := filepath.Dir(path)
	expected := filepath.Join(root, "clarify")
	if dir != expected {
		t.Errorf("file dir = %q, want %q", dir, expected)
	}
}

func TestCreateIdeaTask_FilenameIsKebabCase(t *testing.T) {
	root := makeQueue(t)
	path, err := frontloop.CreateIdeaTask(root, "add retry logic to the API client", "medium")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	name := filepath.Base(path)
	want := "add-retry-logic-to-the-api-client.md"
	if name != want {
		t.Errorf("filename = %q, want %q", name, want)
	}
}

func TestCreateIdeaTask_FileHasCorrectFrontmatter(t *testing.T) {
	root := makeQueue(t)
	path, err := frontloop.CreateIdeaTask(root, "add retry logic", "high")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !containsString(content, "priority: high") {
		t.Errorf("file missing priority: %q", content)
	}
	if !containsString(content, "title: add retry logic") {
		t.Errorf("file missing title: %q", content)
	}
}

func TestCreateIdeaTask_DeduplicatesFilename(t *testing.T) {
	root := makeQueue(t)
	// Create first file
	_, err := frontloop.CreateIdeaTask(root, "add retry logic", "medium")
	if err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	// Create second with same description
	path2, err := frontloop.CreateIdeaTask(root, "add retry logic", "medium")
	if err != nil {
		t.Fatalf("second create failed: %v", err)
	}

	name := filepath.Base(path2)
	want := "add-retry-logic-2.md"
	if name != want {
		t.Errorf("second filename = %q, want %q", name, want)
	}
}

func TestCreateIdeaTask_V2DefaultsToDefaultEpicClarify(t *testing.T) {
	dir := t.TempDir()
	root := makeEpicFrontloop(t, dir, frontloop.DefaultEpicSlug)

	path, err := frontloop.CreateIdeaTask(root, "add retry logic", "medium")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(root, frontloop.DefaultEpicSlug, frontloop.StatusClarify, "add-retry-logic.md")
	if path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	if _, err := os.Stat(want); err != nil {
		t.Errorf("file not created at %q: %v", want, err)
	}
}

func TestCreateIdeaTaskInEpic_WritesSelectedEpicClarify(t *testing.T) {
	dir := t.TempDir()
	root := makeEpicFrontloop(t, dir, frontloop.DefaultEpicSlug, "checkout-redesign")

	path, err := frontloop.CreateIdeaTaskInEpic(root, "checkout-redesign", "render review page", "high")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := filepath.Join(root, "checkout-redesign", frontloop.StatusClarify, "render-review-page.md")
	if path != want {
		t.Errorf("path = %q, want %q", path, want)
	}
	if _, err := os.Stat(want); err != nil {
		t.Errorf("file not created at %q: %v", want, err)
	}
}

func TestCreateIdeaTaskInEpic_RejectsMissingEpic(t *testing.T) {
	dir := t.TempDir()
	root := makeEpicFrontloop(t, dir, frontloop.DefaultEpicSlug)

	_, err := frontloop.CreateIdeaTaskInEpic(root, "checkout-redesign", "render review page", "high")
	if err == nil {
		t.Fatal("expected missing epic error")
	}
	if !containsString(err.Error(), "frontloop epic \"checkout-redesign\" does not exist") {
		t.Errorf("expected helpful missing epic error, got: %v", err)
	}

	unexpected := filepath.Join(root, "checkout-redesign", frontloop.StatusClarify, "render-review-page.md")
	if _, err := os.Stat(unexpected); !os.IsNotExist(err) {
		t.Errorf("missing epic should not be auto-created, stat err = %v", err)
	}
}

func TestIdeaFilenameUniqueInEpic_DeduplicatesWithinSelectedEpicOnly(t *testing.T) {
	dir := t.TempDir()
	root := makeEpicFrontloop(t, dir, frontloop.DefaultEpicSlug, "checkout-redesign")
	writeTask(t, root, filepath.Join(frontloop.DefaultEpicSlug, frontloop.StatusClarify), "add-retry-logic.md", taskA)
	writeTask(t, root, filepath.Join("checkout-redesign", frontloop.StatusClarify), "add-retry-logic.md", taskA)
	writeTask(t, root, filepath.Join("checkout-redesign", frontloop.StatusClarify), "add-retry-logic-2.md", taskA)

	defaultName, err := frontloop.IdeaFilenameUniqueInEpic(root, frontloop.DefaultEpicSlug, "add retry logic")
	if err != nil {
		t.Fatalf("unexpected default epic error: %v", err)
	}
	if defaultName != "add-retry-logic-2.md" {
		t.Errorf("default epic filename = %q, want add-retry-logic-2.md", defaultName)
	}

	checkoutName, err := frontloop.IdeaFilenameUniqueInEpic(root, "checkout-redesign", "add retry logic")
	if err != nil {
		t.Fatalf("unexpected checkout epic error: %v", err)
	}
	if checkoutName != "add-retry-logic-3.md" {
		t.Errorf("checkout epic filename = %q, want add-retry-logic-3.md", checkoutName)
	}
}

func TestCreateIdeaTaskInEpic_DoesNotWriteEpicFrontmatter(t *testing.T) {
	dir := t.TempDir()
	root := makeEpicFrontloop(t, dir, frontloop.DefaultEpicSlug, "checkout-redesign")

	path, err := frontloop.CreateIdeaTaskInEpic(root, "checkout-redesign", "render review page", "high")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if containsString(content, "epic:") {
		t.Errorf("task frontmatter should not include epic membership: %q", content)
	}
}
