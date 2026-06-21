package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jeprecated/frontloop/apps/fl/internal/frontloop"
	"github.com/spf13/cobra"
)

var epicCmd = &cobra.Command{
	Use:   "epic",
	Short: "Manage active frontloop epics",
	Long:  "Create and list active frontloop epics. Epic slugs must use lower-case letters, digits, and hyphens.",
}

var epicNewCmd = &cobra.Command{
	Use:   "new <slug>",
	Short: "Create a new active epic",
	Long:  "Create a new active epic. Slugs must use lower-case letters, digits, and hyphens, such as checkout-redesign.",
	Args:  cobra.ExactArgs(1),
	RunE:  runEpicNewCmd,
}

var epicListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active epics",
	Args:  cobra.NoArgs,
	RunE:  runEpicListCmd,
}

var epicArchiveCmd = &cobra.Command{
	Use:   "archive <slug>",
	Short: "Archive a completed epic",
	Long:  "Archive a completed epic by moving it to .frontloop/_archive/. The epic must have no tasks in clarify, ready, or in_progress.",
	Args:  cobra.ExactArgs(1),
	RunE:  runEpicArchiveCmd,
}

func init() {
	epicCmd.AddCommand(epicNewCmd, epicListCmd, epicArchiveCmd)
	rootCmd.AddCommand(epicCmd)
}

func runEpicNewCmd(cmd *cobra.Command, args []string) error {
	root, err := findV2FrontloopRoot()
	if err != nil {
		return err
	}

	slug := args[0]
	if err := frontloop.CreateEpic(root, slug); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created epic: %s\n", slug)
	fmt.Fprintf(cmd.OutOrStdout(), "Path: %s\n", filepath.Join(root, slug))
	return nil
}

func runEpicListCmd(cmd *cobra.Command, _ []string) error {
	root, err := findV2FrontloopRoot()
	if err != nil {
		return err
	}

	epics, err := frontloop.ListEpics(root)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Active epics:")
	for _, epic := range epics {
		if epic.Slug == frontloop.DefaultEpicSlug {
			fmt.Fprintf(cmd.OutOrStdout(), "  %s (default)\n", epic.Slug)
			continue
		}
		fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", epic.Slug)
	}
	return nil
}

func runEpicArchiveCmd(cmd *cobra.Command, args []string) error {
	root, err := findV2FrontloopRoot()
	if err != nil {
		return err
	}

	archived, err := frontloop.ArchiveEpic(root, args[0])
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Archived epic: %s\n", archived.Slug)
	fmt.Fprintf(cmd.OutOrStdout(), "Path: %s\n", archived.ArchivePath)
	fmt.Fprintf(cmd.OutOrStdout(), "Manual restore: move %s back to %s and update epic.md to `status: active` with an empty `completed_at`.\n", archived.ArchivePath, archived.ActivePath)
	return nil
}

func findV2FrontloopRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	root, err := frontloop.FindRoot(cwd)
	if err != nil {
		return "", fmt.Errorf("no .frontloop directory found (run fl init to create one)")
	}

	switch frontloop.DetectLayout(root) {
	case frontloop.LayoutEpic:
		return root, nil
	case frontloop.LayoutLegacy:
		return "", fmt.Errorf("legacy .frontloop layout detected at %s; run `fl migrate epic-layout` to move tasks into the v2 epic layout", root)
	default:
		return "", fmt.Errorf(".frontloop at %s is not a v2 epic layout (run fl init or fl migrate epic-layout)", root)
	}
}
