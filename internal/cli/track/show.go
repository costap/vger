package track

import (
	"fmt"
	"strings"

	"github.com/costap/vger/internal/cli/ui"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:               "show <id>",
	Short:             "Display a signal in detail",
	Long: `Display the full details of a tracked signal, including any AI enrichment.

Example:
  vger track show 0001`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: signalIDCompletionFunc,
	RunE: func(cmd *cobra.Command, args []string) error {
		id := fmt.Sprintf("%04s", args[0])

		store, err := resolveSignalStore()
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		sig, err := store.Load(cmd.Context(), id)
		if err != nil {
			ui.RedAlert(err)
			return err
		}
		if sig == nil {
			err := fmt.Errorf("signal %s not found", id)
			ui.RedAlert(err)
			return err
		}

		ui.SectionHeader(fmt.Sprintf("signal %s", sig.ID))

		ui.Field("Title", sig.Title)
		ui.Field("Date", sig.Date)
		ui.Field("Status", sig.Status)
		ui.Field("Category", sig.Category)
		ui.Field("Source", sig.Source)
		if sig.URL != "" {
			ui.Field("URL", sig.URL)
		}
		if len(sig.Tags) > 0 {
			ui.Field("Tags", strings.Join(sig.Tags, ", "))
		}
		if len(sig.LinkedVideoIDs) > 0 {
			ui.Field("Linked Videos", strings.Join(sig.LinkedVideoIDs, ", "))
		}

		dimSty := ui.DimStyle()
		labelSty := ui.LabelStyle()

		if sig.Note != "" {
			fmt.Printf("\n  %s\n  %s\n\n",
				labelSty.Render("WHY CAPTURED"),
				dimSty.Render(sig.Note),
			)
		}

		if sig.Enrichment != nil {
			e := sig.Enrichment
			ui.SectionHeader("ai enrichment")

			if e.WhatItIs != "" {
				fmt.Printf("  %s\n  %s\n\n", labelSty.Render("WHAT IT IS"), dimSty.Render(e.WhatItIs))
			}
			if e.Maturity != "" {
				fmt.Printf("  %s\n  %s\n\n", labelSty.Render("MATURITY & RISK"), dimSty.Render(e.Maturity))
			}
			if len(e.Alternatives) > 0 {
				fmt.Printf("  %s\n", labelSty.Render("ALTERNATIVES"))
				for _, a := range e.Alternatives {
					fmt.Printf("    %s\n", dimSty.Render("• "+a))
				}
				fmt.Println()
			}
			if e.StackFit != "" {
				fmt.Printf("  %s\n  %s\n\n", labelSty.Render("STACK FIT"), dimSty.Render(e.StackFit))
			}
			if len(e.NextSteps) > 0 {
				fmt.Printf("  %s\n", labelSty.Render("NEXT STEPS"))
				for _, s := range e.NextSteps {
					fmt.Printf("    %s\n", dimSty.Render("☐ "+s))
				}
				fmt.Println()
			}
		}

		return nil
	},
}
