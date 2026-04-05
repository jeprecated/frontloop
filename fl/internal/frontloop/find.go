package frontloop

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrNotFound is returned when no valid .frontloop directory can be found.
var ErrNotFound = errors.New("no .frontloop directory found")

var requiredSubdirs = []string{"clarify", "ready", "in_progress", "done"}

// FindRoot walks up from startDir looking for a .frontloop/ directory that
// contains all four required subdirectories. Returns the absolute path to the
// .frontloop/ directory, or ErrNotFound if none is found.
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

func isValidRoot(path string) bool {
	for _, sub := range requiredSubdirs {
		info, err := os.Stat(filepath.Join(path, sub))
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}
