package cli

import (
	"fmt"
	"strings"

	"github.com/costap/vger/internal/adapters/gemini"
	"github.com/costap/vger/internal/adapters/youtube"
	"github.com/costap/vger/internal/agent"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan <youtube-url>",
	Short: "Analyse a conference video and extract a technology summary",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		ui.Status("Initialising knowledge assimilation sequence...")
		ui.Status(fmt.Sprintf("Target: %s", url))
		ui.Status("Contacting YouTube metadata relay...")

		ytClient := youtube.New(youtubeAPIKey)
		gmClient := gemini.New(geminiAPIKey, geminiModel)
		a := agent.New(ytClient, gmClient)

		ui.Status("Transmitting video to Gemini multimodal array...")

		report, err := a.Run(cmd.Context(), url)
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		report.Stardate = ui.Stardate()

		ui.Complete("Analysis complete. Captain.")

		// Render report
		ui.SectionHeader("Mission Report")
		ui.Field("title", report.VideoTitle)
		ui.Field("url", report.VideoURL)
		ui.Field("stardate", report.Stardate)

		ui.SectionHeader("Summary")
		fmt.Printf("\n  %s\n", report.Summary)

		ui.SectionHeader("Technologies Identified")
		for i, t := range report.Technologies {
			fmt.Printf("\n  %s\n", ui.LabelStyle().Render(fmt.Sprintf("%d. %s", i+1, t.Name)))
			if t.CNCFStage != "" {
				fmt.Printf("  %s\n", ui.DimStyle().Render(fmt.Sprintf("   CNCF: %s", strings.ToUpper(t.CNCFStage))))
			}
			fmt.Printf("  %s\n", t.Description)
			fmt.Printf("  %s %s\n", ui.DimStyle().Render("Why relevant:"), t.WhyRelevant)
			fmt.Printf("  %s %s\n", ui.DimStyle().Render("Learn more:  "), t.LearnMore)
		}

		fmt.Println()
		return nil
	},
}
