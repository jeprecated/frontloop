package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/ohare93/fl/internal/frontloop"
	"github.com/spf13/cobra"
)

const maxDoneShown = 5

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show frontloop queue status",
	RunE:  runStatsCmd,
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

func runStatsCmd(cmd *cobra.Command, _ []string) error {
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
	renderStats(cmd.OutOrStdout(), all)
	return nil
}

func renderStats(w io.Writer, all map[string][]frontloop.Task) {
	renderer := lipgloss.NewRenderer(w)

	bold := renderer.NewStyle().Bold(true)
	green := renderer.NewStyle().Foreground(lipgloss.Color("2"))
	yellow := renderer.NewStyle().Foreground(lipgloss.Color("3"))
	dim := renderer.NewStyle().Faint(true)

	priorityStyle := map[string]lipgloss.Style{
		"critical": renderer.NewStyle().Foreground(lipgloss.Color("1")),
		"high":     renderer.NewStyle().Foreground(lipgloss.Color("3")),
		"medium":   renderer.NewStyle().Foreground(lipgloss.Color("6")),
		"low":      renderer.NewStyle().Foreground(lipgloss.Color("8")),
	}

	printSection := func(label string, hdrStyle lipgloss.Style, tasks []frontloop.Task, truncate bool) {
		fmt.Fprintf(w, "%s (%d)\n", hdrStyle.Render(label), len(tasks))
		if len(tasks) == 0 {
			fmt.Fprintln(w, "  (empty)")
		} else {
			shown := tasks
			if truncate && len(tasks) > maxDoneShown {
				shown = tasks[:maxDoneShown]
			}
			for _, t := range shown {
				name := strings.TrimSuffix(t.Filename, ".md")
				ps, ok := priorityStyle[t.Priority]
				if !ok {
					ps = renderer.NewStyle()
				}
				fmt.Fprintf(w, "  %s %s %s\n", name, ps.Render("["+t.Priority+"]"), t.Title)
			}
			if truncate && len(tasks) > maxDoneShown {
				fmt.Fprintf(w, "  %s\n", dim.Render(fmt.Sprintf("... and %d more", len(tasks)-maxDoneShown)))
			}
		}
		fmt.Fprintln(w)
	}

	printSection("IN PROGRESS", bold, all["in_progress"], false)
	printSection("READY", green, all["ready"], false)
	printSection("NEEDS CLARIFICATION", yellow, all["clarify"], false)
	printSection("DONE", dim, sortByModTime(all["done"]), true)
}

func sortByModTime(tasks []frontloop.Task) []frontloop.Task {
	type entry struct {
		task    frontloop.Task
		modTime time.Time
	}
	entries := make([]entry, 0, len(tasks))
	for _, t := range tasks {
		var mt time.Time
		if info, err := os.Stat(t.Path); err == nil {
			mt = info.ModTime()
		}
		entries = append(entries, entry{t, mt})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].modTime.After(entries[j].modTime)
	})
	result := make([]frontloop.Task, len(entries))
	for i, e := range entries {
		result[i] = e.task
	}
	return result
}
