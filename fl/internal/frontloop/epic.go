package frontloop

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var validEpicSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ValidateEpicSlug reports whether slug is valid for an active epic directory.
// Slugs must use lower-case letters, digits, and hyphens, and must not use
// reserved frontloop names.
func ValidateEpicSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("frontloop epic slug is required")
	}
	if IsReservedEpicName(slug) {
		return fmt.Errorf("reserved frontloop epic name %q", slug)
	}
	if !validEpicSlug.MatchString(slug) {
		return fmt.Errorf("invalid frontloop epic slug %q (use lower-case letters, digits, and hyphens; start and end with a letter or digit)", slug)
	}
	return nil
}

// CreateEpic creates a new active epic directory. Unlike EnsureEpic, it refuses
// to reuse an existing path so command callers do not accidentally overwrite or
// complete a preexisting epic directory.
func CreateEpic(root, slug string) error {
	if err := ValidateEpicSlug(slug); err != nil {
		return err
	}

	epicPath := filepath.Join(root, slug)
	if _, err := os.Stat(epicPath); err == nil {
		return fmt.Errorf("frontloop epic %q already exists", slug)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return EnsureEpic(root, slug)
}
