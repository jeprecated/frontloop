package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jeprecated/frontloop/fl/internal/frontloop"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a .frontloop/ directory tree in the current directory",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		root := filepath.Join(cwd, ".frontloop")
		for _, sub := range frontloop.Dirs {
			if err := os.MkdirAll(filepath.Join(root, sub), 0755); err != nil {
				return err
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", root)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
