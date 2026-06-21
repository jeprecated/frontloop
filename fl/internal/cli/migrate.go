package cli

import (
	"fmt"
	"os"

	"github.com/jeprecated/frontloop/fl/internal/frontloop"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate frontloop queue layouts",
}

var migrateEpicLayoutCmd = &cobra.Command{
	Use:   "epic-layout",
	Short: "Move legacy flat queues into the v2 default epic",
	Args:  cobra.NoArgs,
	RunE:  runMigrateEpicLayoutCmd,
}

func init() {
	migrateCmd.AddCommand(migrateEpicLayoutCmd)
	rootCmd.AddCommand(migrateCmd)
}

func runMigrateEpicLayoutCmd(cmd *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	root, err := frontloop.FindRoot(cwd)
	if err != nil {
		return fmt.Errorf("no .frontloop directory found (run fl init to create one)")
	}

	result, err := frontloop.MigrateLegacyToEpicLayout(root)
	if err != nil {
		return fmt.Errorf("could not migrate legacy .frontloop layout at %s: %w", root, err)
	}

	if result.TasksMoved == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No legacy tasks to migrate; v2 layout is ready at %s\n", root)
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Migrated %d legacy task(s) into %s/%s/\n", result.TasksMoved, root, frontloop.DefaultEpicSlug)
	return nil
}
