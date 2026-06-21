package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// ErrInterrupted is returned when the run is stopped by a signal.
var ErrInterrupted = errors.New("interrupted by signal")

var rootCmd = &cobra.Command{
	Use:   "fl",
	Short: "Frontloop task queue CLI",
	// Disable Cobra's auto-generated completion command so we can provide our
	// own that correctly routes output through cmd.OutOrStdout() at call time.
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

var completionCmd = &cobra.Command{
	Use:       "completion [bash|zsh|fish]",
	Short:     "Generate shell completion scripts",
	ValidArgs: []string{"bash", "zsh", "fish"},
	Args:      cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		w := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(w)
		case "zsh":
			return cmd.Root().GenZshCompletion(w)
		case "fish":
			return cmd.Root().GenFishCompletion(w, true)
		default:
			return fmt.Errorf("unsupported shell: %q (use bash, zsh, or fish)", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

// SetVersion sets the version string (injected at build time).
func SetVersion(v string) {
	rootCmd.Version = v
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
