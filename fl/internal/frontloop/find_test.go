package frontloop_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ohare93/frontloop/fl/internal/frontloop"
)

func makeFrontloop(t *testing.T, base string) string {
	t.Helper()
	root := filepath.Join(base, ".frontloop")
	for _, sub := range []string{"clarify", "ready", "in_progress", "done"} {
		if err := os.MkdirAll(filepath.Join(root, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestFindRoot_FindsInCurrentDir(t *testing.T) {
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

func TestFindRoot_FindsInParent(t *testing.T) {
	dir := t.TempDir()
	makeFrontloop(t, dir)
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
