package cli

import (
	"fmt"
	"strings"

	"github.com/jeprecated/frontloop/fl/internal/frontloop"
	"github.com/spf13/cobra"
)

var ideaCmd = &cobra.Command{
	Use:   "idea [description]",
	Short: "Quickly capture a task idea into an epic's clarify queue",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		description := strings.Join(args, " ")
		priority, _ := cmd.Flags().GetString("priority")
		epic, _ := cmd.Flags().GetString("epic")
		if epic == "" {
			epic = frontloop.DefaultEpicSlug
		}

		root, err := findV2FrontloopRoot()
		if err != nil {
			return err
		}

		path, err := frontloop.CreateIdeaTaskInEpic(root, epic, description, priority)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", path)
		return nil
	},
}

func init() {
	ideaCmd.Flags().StringP("priority", "p", "medium", "Priority (critical, high, medium, low)")
	ideaCmd.Flags().String("epic", frontloop.DefaultEpicSlug, "Epic slug to receive the new task")
	rootCmd.AddCommand(ideaCmd)
}
