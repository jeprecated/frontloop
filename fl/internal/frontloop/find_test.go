package frontloop_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jeprecated/frontloop/fl/internal/frontloop"
)

func makeFrontloop(t *testing.T, base string) string {
	t.Helper()
	root := filepath.Join(base, ".frontloop")
	for _, sub := range frontloop.Statuses {
		if err := os.MkdirAll(filepath.Join(root, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func makeEpicFrontloop(t *testing.T, base string, epics ...string) string {
	t.Helper()
	root := filepath.Join(base, ".frontloop")
	if len(epics) == 0 {
		epics = []string{frontloop.DefaultEpicSlug}
	}
	for _, epic := range epics {
		for _, status := range frontloop.Statuses {
			if err := os.MkdirAll(filepath.Join(root, epic, status), 0755); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := os.MkdirAll(filepath.Join(root, frontloop.ArchiveDirName), 0755); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestFindRoot_FindsLegacyInCurrentDir(t *testing.T) {
	dir := t.TempDir()
	makeFrontloop(t, dir)

	got, err := frontloop.FindRoot(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(dir, ".frontloop")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFindRoot_FindsV2DefaultLayoutInCurrentDir(t *testing.T) {
	dir := t.TempDir()
	makeEpicFrontloop(t, dir)

	got, err := frontloop.FindRoot(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(dir, ".frontloop")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFindRoot_FindsInParent(t *testing.T) {
	dir := t.TempDir()
	makeEpicFrontloop(t, dir)
	child := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(child, 0755); err != nil {
		t.Fatal(err)
	}

	got, err := frontloop.FindRoot(child)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(dir, ".frontloop")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDetectLayout_DistinguishesV2AndLegacy(t *testing.T) {
	legacyDir := t.TempDir()
	legacyRoot := makeFrontloop(t, legacyDir)
	if got := frontloop.DetectLayout(legacyRoot); got != frontloop.LayoutLegacy {
		t.Errorf("legacy layout = %q, want %q", got, frontloop.LayoutLegacy)
	}

	v2Dir := t.TempDir()
	v2Root := makeEpicFrontloop(t, v2Dir)
	if got := frontloop.DetectLayout(v2Root); got != frontloop.LayoutEpic {
		t.Errorf("v2 layout = %q, want %q", got, frontloop.LayoutEpic)
	}
}

func TestFindRoot_ErrNotFound(t *testing.T) {
	dir := t.TempDir()
	// No .frontloop directory

	_, err := frontloop.FindRoot(dir)
	if !errors.Is(err, frontloop.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFindRoot_IgnoresIncompleteDir(t *testing.T) {
	dir := t.TempDir()
	// .frontloop exists but is missing subdirs
	if err := os.MkdirAll(filepath.Join(dir, ".frontloop", "clarify"), 0755); err != nil {
		t.Fatal(err)
	}

	_, err := frontloop.FindRoot(dir)
	if !errors.Is(err, frontloop.ErrNotFound) {
		t.Errorf("expected ErrNotFound for incomplete .frontloop, got %v", err)
	}
}
