package track

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/costap/vger/internal/domain"
	"github.com/spf13/cobra"
)

var listStatus   string
var listCategory string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tracked signals",
	Long: `List all captured technology signals.

Use --status and --category to filter the list.

Examples:
  vger track list
  vger track list --status spotted
  vger track list --status evaluating --category security`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := resolveSignalStore()
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		var sigs []*domain.Signal

		switch {
		case listStatus != "" && listCategory != "":
			all, err := store.LoadByStatus(cmd.Context(), listStatus)
			if err != nil {
				ui.RedAlert(err)
				return err
			}
			for _, s := range all {
				if s.Category == listCategory {
					sigs = append(sigs, s)
				}
			}
		case listStatus != "":
			sigs, err = store.LoadByStatus(cmd.Context(), listStatus)
			if err != nil {
				ui.RedAlert(err)
				return err
			}
		case listCategory != "":
			sigs, err = store.LoadByCategory(cmd.Context(), listCategory)
			if err != nil {
				ui.RedAlert(err)
				return err
			}
		default:
			sigs, err = store.LoadAll(cmd.Context())
			if err != nil {
				ui.RedAlert(err)
				return err
			}
		}

		if len(sigs) == 0 {
			ui.Complete("No signals found.")
			return nil
		}

		title := "signals"
		if listStatus != "" {
			title += " — " + listStatus
		}
		if listCategory != "" {
			title += " — " + listCategory
		}
		ui.SectionHeader(fmt.Sprintf("%s (%d)", title, len(sigs)))

		labelSty := ui.LabelStyle()
		dimSty := ui.DimStyle()

		fmt.Printf("  %s  %s  %s  %s  %s\n",
			labelSty.Render(fmt.Sprintf("%-6s", "ID")),
			dimSty.Render(fmt.Sprintf("%-12s", "DATE")),
			labelSty.Render(fmt.Sprintf("%-11s", "STATUS")),
			dimSty.Render(fmt.Sprintf("%-22s", "CATEGORY")),
			labelSty.Render("TITLE"),
		)
		fmt.Printf("  %s\n", dimSty.Render(strings.Repeat("─", 80)))

		for _, s := range sigs {
			enriched := ""
			if s.Enrichment != nil {
				enriched = "✓"
			}
			fmt.Printf("  %s  %s  %s  %s  %s %s\n",
				labelSty.Render(fmt.Sprintf("%-6s", s.ID)),
				dimSty.Render(fmt.Sprintf("%-12s", s.Date)),
				signalStatusStyle(s.Status).Render(fmt.Sprintf("%-11s", s.Status)),
				dimSty.Render(fmt.Sprintf("%-22s", truncate(s.Category, 22))),
				labelSty.Render(truncate(s.Title, 50)),
				dimSty.Render(enriched),
			)
		}
		fmt.Println()
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (spotted|evaluating|adopted|rejected|parked)")
	listCmd.Flags().StringVar(&listCategory, "category", "", "Filter by category")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func signalStatusStyle(status string) lipgloss.Style {
	switch status {
	case domain.SignalStatusSpotted:
		return ui.LabelStyle()
	case domain.SignalStatusEvaluating:
		return ui.BlueStyle()
	case domain.SignalStatusAdopted:
		return ui.GreenStyle()
	case domain.SignalStatusRejected:
		return ui.RedStyle()
	default:
		return ui.DimStyle()
	}
}
