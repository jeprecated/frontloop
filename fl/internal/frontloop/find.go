package frontloop

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrNotFound is returned when no valid .frontloop directory can be found.
var ErrNotFound = errors.New("no .frontloop directory found")

// Layout identifies the filesystem layout used by a .frontloop root.
type Layout string

const (
	// LayoutUnknown means the path is not a recognized frontloop root.
	LayoutUnknown Layout = "unknown"
	// LayoutLegacy is the original flat .frontloop/<status>/ layout.
	LayoutLegacy Layout = "legacy"
	// LayoutEpic is the v2 .frontloop/<epic>/<status>/ layout.
	LayoutEpic Layout = "epic"
)

// FindRoot walks up from startDir looking for a .frontloop/ directory that
// contains either the v2 default epic queue or the legacy flat queue. Returns
// the absolute path to the .frontloop/ directory, or ErrNotFound if none is
// found.
func FindRoot(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(dir, ".frontloop")
		if isValidRoot(candidate) {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return "", ErrNotFound
		}
		dir = parent
	}
}

// DetectLayout reports which frontloop layout root uses.
func DetectLayout(root string) Layout {
	if IsV2Root(root) {
		return LayoutEpic
	}
	if IsLegacyRoot(root) {
		return LayoutLegacy
	}
	return LayoutUnknown
}

// IsV2Root reports whether path is a v2 epic-first frontloop root. The minimum
// v2 root has .frontloop/default/{clarify,ready,in_progress,done}.
func IsV2Root(path string) bool {
	return hasStatusDirs(filepath.Join(path, DefaultEpicSlug))
}

// IsLegacyRoot reports whether path is a legacy flat frontloop root with
// .frontloop/{clarify,ready,in_progress,done}.
func IsLegacyRoot(path string) bool {
	return hasStatusDirs(path)
}

func isValidRoot(path string) bool {
	return DetectLayout(path) != LayoutUnknown
}

func hasStatusDirs(path string) bool {
	for _, sub := range Statuses {
		info, err := os.Stat(filepath.Join(path, sub))
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}
