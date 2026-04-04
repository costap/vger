package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	genkitadapter "github.com/costap/vger/internal/adapters/genkit"
	"github.com/costap/vger/internal/adapters/gemini"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/costap/vger/internal/domain"
	"github.com/spf13/cobra"
)

var (
	trackDigestStatus   string
	trackDigestCategory string
	trackDigestEnrich   bool
	trackDigestOutput   string
)

var trackDigestCmd = &cobra.Command{
	Use:   "digest",
	Short: "AI-powered synthesis of your signal backlog",
	Long: `Synthesise your tracked signals into a prioritised backlog review.

Gemini analyses your signals and returns:
  • Weekly focus — 3 signals to investigate this week (with reasons)
  • Clusters     — related signals grouped by technology theme
  • Learning path — suggested investigation order
  • Key insights  — narrative of patterns and trends

Use --enrich to fill AI context for any unenriched signals before synthesis.
Use --output to save the report as a Markdown file.

Examples:
  vger track digest
  vger track digest --status spotted --enrich
  vger track digest --output ~/tech-review.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if geminiAPIKey == "" {
			err := fmt.Errorf("GEMINI_API_KEY is required — set it as an env var or pass --gemini-key")
			ui.RedAlert(err)
			return err
		}

		store, err := resolveSignalStore()
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		// Load signals (filtered if flags set).
		var sigs []*domain.Signal
		switch {
		case trackDigestStatus != "":
			sigs, err = store.LoadByStatus(cmd.Context(), trackDigestStatus)
		case trackDigestCategory != "":
			sigs, err = store.LoadByCategory(cmd.Context(), trackDigestCategory)
		default:
			sigs, err = store.LoadAll(cmd.Context())
		}
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		if len(sigs) == 0 {
			fmt.Println(ui.DimStyle().Render("  no signals found — run: vger track add"))
			return nil
		}

		ui.SectionHeader(fmt.Sprintf("track digest — %d signals", len(sigs)))

		// Optional: enrich unenriched signals.
		if trackDigestEnrich {
			gmClient := gemini.New(geminiAPIKey, geminiModel)
			enriched := 0
			for _, sig := range sigs {
				if sig.Enrichment != nil {
					continue
				}
				fmt.Printf("  %s %s\n", ui.DimStyle().Render("enriching:"), sig.Title)
				enrichment, err := gmClient.EnrichSignal(cmd.Context(), sig)
				if err != nil {
					ui.RedAlert(fmt.Errorf("enrich %s: %w", sig.ID, err))
					return err
				}
				sig.Enrichment = enrichment
				sig.UpdatedAt = time.Now().UTC()
				if err := store.Save(cmd.Context(), sig); err != nil {
					ui.RedAlert(err)
					return err
				}
				enriched++
			}
			if enriched > 0 {
				fmt.Printf("  %s\n\n", ui.DimStyle().Render(fmt.Sprintf("%d signal(s) enriched", enriched)))
			}
		}

		// Build pulse stats.
		pulse := buildPulse(sigs)

		// Dereference slice of pointers to slice of values for Genkit input.
		flatSigs := make([]domain.Signal, len(sigs))
		for i, s := range sigs {
			flatSigs[i] = *s
		}

		fmt.Println(ui.DimStyle().Render("  calling Gemini via Genkit…"))

		report, err := genkitadapter.DigestSignals(cmd.Context(), geminiAPIKey, geminiModel,
			genkitadapter.DigestInput{
				Signals: flatSigs,
				Pulse:   pulse,
			},
		)
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		// Render to terminal.
		renderDigestReport(report, pulse, len(sigs))

		// Optionally write Markdown file.
		if trackDigestOutput != "" {
			if err := writeDigestMarkdown(report, pulse, sigs, trackDigestOutput); err != nil {
				ui.RedAlert(fmt.Errorf("write output: %w", err))
				return err
			}
			ui.Complete(fmt.Sprintf("report saved to %s", trackDigestOutput))
		}

		return nil
	},
}

func init() {
	trackDigestCmd.Flags().StringVar(&trackDigestStatus, "status", "", "Filter signals by status before digesting")
	trackDigestCmd.Flags().StringVar(&trackDigestCategory, "category", "", "Filter signals by category before digesting")
	trackDigestCmd.Flags().BoolVar(&trackDigestEnrich, "enrich", false, "Enrich unenriched signals with AI context before synthesis")
	trackDigestCmd.Flags().StringVar(&trackDigestOutput, "output", "", "Write a Markdown report to this file path")
}

func buildPulse(sigs []*domain.Signal) domain.SignalPulse {
	pulse := domain.SignalPulse{
		ByStatus:   make(map[string]int),
		ByCategory: make(map[string]int),
	}
	for _, s := range sigs {
		pulse.ByStatus[s.Status]++
		pulse.ByCategory[s.Category]++
	}
	return pulse
}

func renderDigestReport(report *domain.SignalDigestReport, pulse domain.SignalPulse, total int) {
	dimSty := ui.DimStyle()
	labelSty := ui.LabelStyle()

	// Pulse summary.
	fmt.Println()
	ui.SectionHeader("pulse")
	fmt.Printf("  %s  ", labelSty.Render("TOTAL"))
	fmt.Printf("%s\n\n", dimSty.Render(fmt.Sprintf("%d signals", total)))

	fmt.Printf("  %s\n", labelSty.Render("BY STATUS"))
	for _, status := range domain.ValidSignalStatuses {
		if n, ok := pulse.ByStatus[status]; ok {
			fmt.Printf("    %s %s\n", dimSty.Render(fmt.Sprintf("%-12s", status)), dimSty.Render(fmt.Sprintf("%d", n)))
		}
	}
	fmt.Println()

	// Weekly focus.
	if len(report.WeeklyFocus) > 0 {
		ui.SectionHeader("weekly focus")
		for i, f := range report.WeeklyFocus {
			fmt.Printf("  %s %s\n", labelSty.Render(fmt.Sprintf("[%d]", i+1)), dimSty.Render(f.Title))
			if f.URL != "" {
				fmt.Printf("      %s\n", dimSty.Render(f.URL))
			}
			fmt.Printf("      %s\n\n", dimSty.Render(f.Reason))
		}
	}

	// Clusters.
	if len(report.Clusters) > 0 {
		ui.SectionHeader("clusters")
		for _, c := range report.Clusters {
			fmt.Printf("  %s\n", labelSty.Render(strings.ToUpper(c.Theme)))
			fmt.Printf("    %s\n", dimSty.Render(c.Summary))
			fmt.Printf("    %s %s\n\n", dimSty.Render("signals:"), dimSty.Render(strings.Join(c.SignalIDs, ", ")))
		}
	}

	// Learning path.
	if len(report.LearningPath) > 0 {
		ui.SectionHeader("learning path")
		for i, item := range report.LearningPath {
			fmt.Printf("  %s %s\n", dimSty.Render(fmt.Sprintf("%2d.", i+1)), dimSty.Render(item))
		}
		fmt.Println()
	}

	// Key insights.
	if report.KeyInsights != "" {
		ui.SectionHeader("key insights")
		fmt.Printf("  %s\n\n", dimSty.Render(report.KeyInsights))
	}
}

func writeDigestMarkdown(report *domain.SignalDigestReport, pulse domain.SignalPulse, sigs []*domain.Signal, path string) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Tech Signal Digest — %s\n\n", time.Now().Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("**Total signals reviewed:** %d\n\n", len(sigs)))

	sb.WriteString("## Pulse\n\n")
	sb.WriteString("| Status | Count |\n|--------|-------|\n")
	for _, status := range domain.ValidSignalStatuses {
		if n, ok := pulse.ByStatus[status]; ok {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", status, n))
		}
	}
	sb.WriteString("\n")

	if len(report.WeeklyFocus) > 0 {
		sb.WriteString("## Weekly Focus\n\n")
		for i, f := range report.WeeklyFocus {
			sb.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, f.Title))
			if f.URL != "" {
				sb.WriteString(fmt.Sprintf("**URL:** %s\n\n", f.URL))
			}
			sb.WriteString(fmt.Sprintf("%s\n\n", f.Reason))
		}
	}

	if len(report.Clusters) > 0 {
		sb.WriteString("## Clusters\n\n")
		for _, c := range report.Clusters {
			sb.WriteString(fmt.Sprintf("### %s\n\n", c.Theme))
			sb.WriteString(fmt.Sprintf("%s\n\n", c.Summary))
			sb.WriteString(fmt.Sprintf("**Signals:** %s\n\n", strings.Join(c.SignalIDs, ", ")))
		}
	}

	if len(report.LearningPath) > 0 {
		sb.WriteString("## Learning Path\n\n")
		for i, item := range report.LearningPath {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
		}
		sb.WriteString("\n")
	}

	if report.KeyInsights != "" {
		sb.WriteString("## Key Insights\n\n")
		sb.WriteString(report.KeyInsights)
		sb.WriteString("\n\n")
	}

	sb.WriteString("---\n*Generated by vger track digest*\n")

	return os.WriteFile(path, []byte(sb.String()), 0644)
}
