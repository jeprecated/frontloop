package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jeprecated/frontloop/apps/fl/internal/frontloop"
	"github.com/spf13/cobra"
)

const maxDoneShown = 5

type epicStats struct {
	Slug  string
	Tasks map[string][]frontloop.Task
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show frontloop queue status",
	RunE:  runStatsCmd,
}

func init() {
	statsCmd.Flags().Bool("no-color", false, "Disable color output (for piping/scripting)")
	statsCmd.Flags().String("epic", "", "Only show tasks for the given active epic slug")
	rootCmd.AddCommand(statsCmd)
}

func runStatsCmd(cmd *cobra.Command, _ []string) error {
	root, err := findV2FrontloopRoot()
	if err != nil {
		return err
	}

	epicFilter, _ := cmd.Flags().GetString("epic")
	epics, err := loadStatsEpics(root, epicFilter)
	if err != nil {
		return err
	}

	noColor, _ := cmd.Flags().GetBool("no-color")
	renderStatsByEpic(cmd.OutOrStdout(), epics, noColor)
	return nil
}

func loadStatsEpics(root, epicFilter string) ([]epicStats, error) {
	activeEpics, err := frontloop.ListEpics(root)
	if err != nil {
		return nil, err
	}

	if epicFilter != "" {
		if err := frontloop.ValidateEpicSlug(epicFilter); err != nil {
			return nil, err
		}
		if !hasActiveEpic(activeEpics, epicFilter) {
			return nil, fmt.Errorf("frontloop epic %q does not exist (run `fl epic list` to see active epics)", epicFilter)
		}

		tasks, err := frontloop.ListEpic(root, epicFilter)
		if err != nil {
			return nil, err
		}
		return []epicStats{{Slug: epicFilter, Tasks: tasks}}, nil
	}

	allByEpic, err := frontloop.ListAllByEpic(root)
	if err != nil {
		return nil, err
	}

	groups := make([]epicStats, 0, len(activeEpics))
	for _, epic := range activeEpics {
		groups = append(groups, epicStats{Slug: epic.Slug, Tasks: allByEpic[epic.Slug]})
	}
	return groups, nil
}

func hasActiveEpic(epics []frontloop.Epic, slug string) bool {
	for _, epic := range epics {
		if epic.Slug == slug {
			return true
		}
	}
	return false
}

func renderStatsByEpic(w io.Writer, epics []epicStats, noColor bool) {
	if noColor {
		prev := os.Getenv("NO_COLOR")
		os.Setenv("NO_COLOR", "1")        //nolint:errcheck
		defer os.Setenv("NO_COLOR", prev) //nolint:errcheck
	}
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

	for i, epic := range epics {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "%s\n\n", bold.Render(epicHeader(epic.Slug)))
		printSection("IN PROGRESS", bold, epic.Tasks[frontloop.StatusInProgress], false)
		printSection("READY", green, epic.Tasks[frontloop.StatusReady], false)
		printSection("NEEDS CLARIFICATION", yellow, epic.Tasks[frontloop.StatusClarify], false)
		printSection("DONE", dim, sortByModTime(epic.Tasks[frontloop.StatusDone]), true)
	}
}

func epicHeader(slug string) string {
	if slug == frontloop.DefaultEpicSlug {
		return fmt.Sprintf("EPIC: %s (default)", slug)
	}
	return "EPIC: " + slug
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
