package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/jeprecated/frontloop/fl/internal/frontloop"
	"github.com/spf13/cobra"
)

var ideaCmd = &cobra.Command{
	Use:   "idea [description]",
	Short: "Quickly capture a task idea into .frontloop/clarify/",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		description := strings.Join(args, " ")
		priority, _ := cmd.Flags().GetString("priority")

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		root, err := frontloop.FindRoot(cwd)
		if err != nil {
			return fmt.Errorf("no .frontloop directory found (run fl init to create one)")
		}

		path, err := frontloop.CreateIdeaTask(root, description, priority)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", path)
		return nil
	},
}

func init() {
	ideaCmd.Flags().StringP("priority", "p", "medium", "Priority (critical, high, medium, low)")
	rootCmd.AddCommand(ideaCmd)
}
