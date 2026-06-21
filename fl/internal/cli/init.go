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
	Short: "Create a v2 .frontloop/ directory tree in the current directory",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		root := filepath.Join(cwd, ".frontloop")
		hasLegacyTasks, err := frontloop.HasLegacyTasks(root)
		if err != nil {
			return err
		}
		if hasLegacyTasks {
			return fmt.Errorf("legacy .frontloop task files detected at %s; run `fl migrate epic-layout` to move them into the v2 default epic", root)
		}

		if err := frontloop.EnsureV2Root(root); err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Initialized v2 .frontloop task queue: %s\n", root)
		fmt.Fprintf(cmd.OutOrStdout(), "  %s/        - default epic for unscoped tasks\n", frontloop.DefaultEpicSlug)
		fmt.Fprintf(cmd.OutOrStdout(), "  %s/       - archived completed epics\n", frontloop.ArchiveDirName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
