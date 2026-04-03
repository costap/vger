package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/costap/vger/internal/adapters/cache"
	"github.com/costap/vger/internal/adapters/gemini"
	"github.com/costap/vger/internal/adapters/youtube"
	"github.com/costap/vger/internal/agent"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/costap/vger/internal/domain"
	"github.com/spf13/cobra"
)

var scanRefresh bool

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

		// Resolve video ID for cache key
		videoID, err := ytClient.ExtractVideoID(url)
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		// Load from cache unless --refresh requested
		cacheDir, err := cache.DefaultDir()
		if err != nil {
			ui.RedAlert(err)
			return err
		}
		c := cache.New(cacheDir)

		if !scanRefresh {
			cached, err := c.Load(cmd.Context(), videoID)
			if err == nil && cached != nil {
				ui.Status(fmt.Sprintf("Loaded from cache: %s/%s.json", cacheDir, videoID))
				renderReport(&cached.Report)
				return nil
			}
		}

		ui.Status("Transmitting video to Gemini multimodal array...")

		report, meta, err := a.Run(cmd.Context(), url)
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		report.Stardate = ui.Stardate()

		// Persist to cache
		entry := &domain.CachedAnalysis{
			VideoID:  videoID,
			CachedAt: time.Now(),
			Report:   *report,
		}
		if meta != nil {
			entry.Metadata = *meta
		}
		if saveErr := c.Save(cmd.Context(), entry); saveErr != nil {
			ui.Status(fmt.Sprintf("Warning: could not save cache: %v", saveErr))
		} else {
			ui.Status(fmt.Sprintf("Analysis cached: %s/%s.json", cacheDir, videoID))
		}

		ui.Complete("Analysis complete. Captain.")
		renderReport(report)
		return nil
	},
}

func init() {
	scanCmd.Flags().BoolVar(&scanRefresh, "refresh", false, "Re-analyse even if a cached result exists")
}

func renderReport(report *domain.Report) {
	ui.SectionHeader("Mission Report")
	ui.Field("title", report.VideoTitle)
	ui.Field("url", report.VideoURL)
	if report.VideoDurationSec > 0 {
		ui.Field("duration", formatDuration(report.VideoDurationSec))
	}
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
}

// formatDuration converts a duration in seconds to a human-readable string (e.g. "1h 15m 30s").
func formatDuration(secs int) string {
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	switch {
	case h > 0:
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	case m > 0:
		return fmt.Sprintf("%dm %ds", m, s)
	default:
		return fmt.Sprintf("%ds", s)
	}
}
