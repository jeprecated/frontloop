package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/jeprecated/frontloop/fl/internal/frontloop"
	"github.com/jeprecated/frontloop/fl/internal/tui"
)

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Interactively move tasks between queue directories",
	RunE:  runMoveCmd,
}

func init() {
	rootCmd.AddCommand(moveCmd)
}

func runMoveCmd(_ *cobra.Command, _ []string) error {
	root, err := findV2FrontloopRoot()
	if err != nil {
		return err
	}
	all, err := frontloop.ListAll(root)
	if err != nil {
		return err
	}

	m := tui.New(root, all)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
