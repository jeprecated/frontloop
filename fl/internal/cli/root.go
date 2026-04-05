package cli

import (
	"errors"

	"github.com/spf13/cobra"
)

// ErrInterrupted is returned when the run is stopped by a signal.
var ErrInterrupted = errors.New("interrupted by signal")

var rootCmd = &cobra.Command{
	Use:   "fl",
	Short: "Frontloop task queue CLI",
}

// SetVersion sets the version string (injected at build time).
func SetVersion(v string) {
	rootCmd.Version = v
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
