package track

import (
	"context"
	"fmt"
	"time"

	"github.com/costap/vger/internal/adapters/gemini"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/costap/vger/internal/domain"
	"github.com/spf13/cobra"
)

var enrichCmd = &cobra.Command{
	Use:   "enrich <id>",
	Short: "AI-enrich a signal with context, alternatives, and next steps",
	Long: `Call Gemini to analyse a captured signal and fill the AI enrichment section.

Enrichment includes: what the technology is, its maturity and risks, alternatives,
how it fits your stack, and concrete next steps for evaluation.

Any existing enrichment is overwritten (idempotent — safe to re-run).

Example:
  vger track enrich 0001`,
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: signalIDCompletionFunc,
	RunE: func(cmd *cobra.Command, args []string) error {
		key := geminiKey(cmd)
		if key == "" {
			err := fmt.Errorf("GEMINI_API_KEY is required — set it as an env var or pass --gemini-key")
			ui.RedAlert(err)
			return err
		}

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

		gmClient := gemini.New(key, model(cmd))

		if err := enrichSignalAndSave(cmd.Context(), gmClient, store, sig); err != nil {
			ui.RedAlert(err)
			return err
		}

		ui.Complete("enrichment saved")
		return nil
	},
}

// enrichSignalAndSave calls Gemini to enrich sig, saves it, and prints the enrichment section.
// Shared by enrichCmd and addCmd (--enrich flag).
func enrichSignalAndSave(ctx context.Context, gmClient *gemini.Client, store domain.SignalStore, sig *domain.Signal) error {
	ui.Field("Enriching", sig.ID+" — "+sig.Title)
	fmt.Println(ui.DimStyle().Render("  calling Gemini…"))

	enrichment, err := gmClient.EnrichSignal(ctx, sig)
	if err != nil {
		return err
	}

	sig.Enrichment = enrichment
	sig.UpdatedAt = time.Now().UTC()

	if err := store.Save(ctx, sig); err != nil {
		return err
	}

	dimSty := ui.DimStyle()
	labelSty := ui.LabelStyle()

	fmt.Println()
	ui.SectionHeader("ai enrichment")
	fmt.Printf("  %s\n  %s\n\n", labelSty.Render("WHAT IT IS"), dimSty.Render(enrichment.WhatItIs))
	fmt.Printf("  %s\n  %s\n\n", labelSty.Render("MATURITY & RISK"), dimSty.Render(enrichment.Maturity))
	if len(enrichment.Alternatives) > 0 {
		fmt.Printf("  %s\n", labelSty.Render("ALTERNATIVES"))
		for _, a := range enrichment.Alternatives {
			fmt.Printf("    %s\n", dimSty.Render("• "+a))
		}
		fmt.Println()
	}
	fmt.Printf("  %s\n  %s\n\n", labelSty.Render("STACK FIT"), dimSty.Render(enrichment.StackFit))
	if len(enrichment.NextSteps) > 0 {
		fmt.Printf("  %s\n", labelSty.Render("NEXT STEPS"))
		for _, s := range enrichment.NextSteps {
			fmt.Printf("    %s\n", dimSty.Render("☐ "+s))
		}
		fmt.Println()
	}

	return nil
}
