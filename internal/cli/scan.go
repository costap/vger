package cli

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
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
var scanPlaylist string
var scanConcurrency int

var scanCmd = &cobra.Command{
	Use:   "scan [youtube-url]",
	Short: "Analyse a conference video and extract a technology summary",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if scanPlaylist != "" {
			return runPlaylistScan(cmd)
		}
		if len(args) == 0 {
			return fmt.Errorf("provide a YouTube URL or use --playlist <id-or-url>")
		}
		return runSingleScan(cmd, args[0])
	},
}

func init() {
	scanCmd.Flags().BoolVar(&scanRefresh, "refresh", false, "Re-analyse even if a cached result exists")
	scanCmd.Flags().StringVar(&scanPlaylist, "playlist", "", "Playlist ID or URL — scan all videos in the playlist")
	scanCmd.Flags().IntVar(&scanConcurrency, "concurrency", 3, "Number of videos to analyse in parallel (playlist mode)")
}

// runSingleScan analyses a single video URL.
func runSingleScan(cmd *cobra.Command, url string) error {
	ui.Status("Initialising knowledge assimilation sequence...")
	ui.Status(fmt.Sprintf("Target: %s", url))
	ui.Status("Contacting YouTube metadata relay...")

	ytClient := youtube.New(youtubeAPIKey)
	gmClient := gemini.New(geminiAPIKey, geminiModel)
	a := agent.New(ytClient, gmClient)

	videoID, err := ytClient.ExtractVideoID(url)
	if err != nil {
		ui.RedAlert(err)
		return err
	}

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
}

// runPlaylistScan fetches all videos in a playlist and analyses each one,
// running up to scanConcurrency analyses in parallel.
func runPlaylistScan(cmd *cobra.Command) error {
	ytClient := youtube.New(youtubeAPIKey)
	gmClient := gemini.New(geminiAPIKey, geminiModel)
	a := agent.New(ytClient, gmClient)

	playlistID, err := ytClient.ExtractPlaylistID(scanPlaylist)
	if err != nil {
		ui.RedAlert(err)
		return err
	}

	ui.Status(fmt.Sprintf("Fetching playlist: %s", playlistID))

	videos, _, err := ytClient.ListPlaylistVideos(cmd.Context(), playlistID, "", 1000)
	if err != nil {
		ui.RedAlert(err)
		return err
	}

	if len(videos) == 0 {
		ui.Complete("No videos found in playlist.")
		return nil
	}

	cacheDir, err := cache.DefaultDir()
	if err != nil {
		ui.RedAlert(err)
		return err
	}
	c := cache.New(cacheDir)

	total := len(videos)
	ui.Status(fmt.Sprintf("Playlist: %s  |  %d videos  |  concurrency: %d", playlistID, total, scanConcurrency))

	var (
		scanned  atomic.Int64
		skipped  atomic.Int64
		failed   atomic.Int64
		mu       sync.Mutex
		wg       sync.WaitGroup
		sem      = make(chan struct{}, scanConcurrency)
	)

	for i, v := range videos {
		idx := i + 1
		videoURL := v.URL
		videoID := v.VideoID
		title := v.Title

		// Check cache before spinning up a goroutine
		if !scanRefresh {
			cached, err := c.Load(cmd.Context(), videoID)
			if err == nil && cached != nil {
				mu.Lock()
				ui.Status(fmt.Sprintf("[%*d/%d] Cached:   %s", len(fmt.Sprintf("%d", total)), idx, total, truncate(title, 60)))
				mu.Unlock()
				skipped.Add(1)
				continue
			}
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			mu.Lock()
			ui.Status(fmt.Sprintf("[%*d/%d] Scanning: %s", len(fmt.Sprintf("%d", total)), idx, total, truncate(title, 60)))
			mu.Unlock()

			report, meta, err := a.Run(cmd.Context(), videoURL)
			if err != nil {
				mu.Lock()
				ui.Status(fmt.Sprintf("[%*d/%d] FAILED:   %s — %v", len(fmt.Sprintf("%d", total)), idx, total, truncate(title, 50), err))
				mu.Unlock()
				failed.Add(1)
				return
			}

			report.Stardate = ui.Stardate()
			entry := &domain.CachedAnalysis{
				VideoID:  videoID,
				CachedAt: time.Now(),
				Report:   *report,
			}
			if meta != nil {
				entry.Metadata = *meta
			}
			if saveErr := c.Save(cmd.Context(), entry); saveErr != nil {
				mu.Lock()
				ui.Status(fmt.Sprintf("Warning: could not save cache for %s: %v", videoID, saveErr))
				mu.Unlock()
			}
			scanned.Add(1)
		}()
	}

	wg.Wait()

	fmt.Println()
	ui.SectionHeader("Playlist Scan Complete")
	ui.Field("scanned", fmt.Sprintf("%d", scanned.Load()))
	ui.Field("cached ", fmt.Sprintf("%d", skipped.Load()))
	ui.Field("failed ", fmt.Sprintf("%d", failed.Load()))
	ui.Field("total  ", fmt.Sprintf("%d", total))
	fmt.Println()

	if failed.Load() > 0 {
		return fmt.Errorf("%d video(s) failed to analyse", failed.Load())
	}
	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
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
