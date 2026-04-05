package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/ohare93/frontloop/fl/internal/frontloop"
	"github.com/ohare93/frontloop/fl/internal/tui"
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
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err := frontloop.FindRoot(cwd)
	if err != nil {
		return fmt.Errorf("no .frontloop directory found (run fl init to create one)")
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
