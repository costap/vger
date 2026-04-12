package cli

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/costap/vger/internal/adapters/cache"
	"github.com/costap/vger/internal/adapters/cncf"
	"github.com/costap/vger/internal/adapters/gemini"
	"github.com/costap/vger/internal/adapters/quota"
	"github.com/costap/vger/internal/adapters/youtube"
	"github.com/costap/vger/internal/agent"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/costap/vger/internal/domain"
	"github.com/spf13/cobra"
)

var scanRefresh bool
var scanPlaylist string
var scanConcurrency int
var scanMaxDuration int
var scanFallbackModel string

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
	scanCmd.Flags().IntVar(&scanConcurrency, "concurrency", 1, "Number of videos to analyse in parallel (playlist mode); lower values reduce Gemini API demand")
	scanCmd.Flags().IntVar(&scanMaxDuration, "max-duration", 0, "Skip playlist videos longer than this many minutes (0 = no limit)")
	scanCmd.Flags().StringVar(&scanFallbackModel, "fallback-model", "", "Retry with this model if primary model hits quota (e.g. gemini-2.5-pro)")
}

// newQuotaTracker creates a quota tracker, logging a warning if unavailable.
// Returns nil on error so callers can proceed without tracking.
func newQuotaTracker() *quota.Tracker {
	path, err := quota.DefaultPath()
	if err != nil {
		return nil
	}
	t, err := quota.New(path)
	if err != nil {
		return nil
	}
	return t
}

// estimateVideoTokens estimates the Gemini token cost for a video based on its
// ISO 8601 duration string (e.g. "PT1H23M45S"). Uses ~200 tokens/sec as a
// heuristic derived from ~1M tokens per 90-minute video. Returns 800_000 when
// the duration is empty or cannot be parsed. Result is clamped to [100K, 1M].
func estimateVideoTokens(iso8601 string) int64 {
	secs := youtube.ParseISO8601Duration(iso8601)
	if secs == 0 {
		return 800_000
	}
	tokens := int64(secs) * 200
	if tokens < 100_000 {
		return 100_000
	}
	if tokens > 1_000_000 {
		return 1_000_000
	}
	return tokens
}

// runSingleScan analyses a single video URL.
func runSingleScan(cmd *cobra.Command, url string) error {
	ui.Status("Initialising knowledge assimilation sequence...")
	ui.Status(fmt.Sprintf("Target: %s", url))
	ui.Status("Contacting YouTube metadata relay...")

	cacheDir, err := cache.DefaultDir()
	if err != nil {
		ui.RedAlert(err)
		return err
	}

	ytClient := youtube.New(youtubeAPIKey)
	gmClient := gemini.NewWithTools(geminiAPIKey, geminiModel, cncf.New(cacheDir))
	if qt := newQuotaTracker(); qt != nil {
		gmClient.QuotaFn = func() { qt.Increment(1) }
	}
	a := agent.New(ytClient, gmClient)

	videoID, err := ytClient.ExtractVideoID(url)
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
		if errors.Is(err, gemini.ErrQuotaExhausted) && scanFallbackModel != "" {
			ui.Status(fmt.Sprintf("⚠ Quota exhausted on %s — retrying with fallback model %s...", geminiModel, scanFallbackModel))
			fbGmClient := gemini.NewWithTools(geminiAPIKey, scanFallbackModel, cncf.New(cacheDir))
			if qt := newQuotaTracker(); qt != nil {
				fbGmClient.QuotaFn = func() { qt.Increment(1) }
			}
			fbAgent := agent.New(ytClient, fbGmClient)
			report, meta, err = fbAgent.Run(cmd.Context(), url)
		}
		if err != nil {
			ui.RedAlert(err)
			return err
		}
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

	cacheDir, err := cache.DefaultDir()
	if err != nil {
		ui.RedAlert(err)
		return err
	}

	gmClient := gemini.NewWithTools(geminiAPIKey, geminiModel, cncf.New(cacheDir))
	a := agent.New(ytClient, gmClient)

	// Set up quota tracker and wire it to every Gemini API dispatch.
	qt := newQuotaTracker()

	// Set up TPM throttle window if limits are known for the configured tier.
	var tpmWindow *quota.TPMWindow
	tpmLimit := int64(lookupModelTPM(geminiModel, quotaTier))
	if tpmLimit > 0 {
		tpmWindow = &quota.TPMWindow{}
	}

	// Set up fallback agent (used when primary model hits quota exhaustion).
	var fallbackAgent *agent.ScanAgent
	var fallbackGmClient *gemini.Client
	if scanFallbackModel != "" {
		fallbackGmClient = gemini.NewWithTools(geminiAPIKey, scanFallbackModel, cncf.New(cacheDir))
		fallbackAgent = agent.New(ytClient, fallbackGmClient)
	}

	playlistID, err := ytClient.ExtractPlaylistID(scanPlaylist)
	if err != nil {
		ui.RedAlert(err)
		return err
	}

	ui.Status(fmt.Sprintf("Fetching playlist: %s", playlistID))

	// Resolve the human-readable playlist title for event tagging.
	playlistTitle, err := ytClient.GetPlaylistTitle(cmd.Context(), playlistID)
	if err != nil {
		ui.Status(fmt.Sprintf("(could not resolve playlist title: %v)", err))
		playlistTitle = playlistID
	}

	videos, _, err := ytClient.ListPlaylistVideos(cmd.Context(), playlistID, "", 1000)
	if err != nil {
		ui.RedAlert(err)
		return err
	}

	if len(videos) == 0 {
		ui.Complete("No videos found in playlist.")
		return nil
	}

	// Enrich all videos with duration and view count (needed for max-duration
	// filtering and improves TPM estimation accuracy). Best-effort: if this
	// fails we proceed without durations (filtering/throttling degrade gracefully).
	if err := ytClient.EnrichWithDetails(cmd.Context(), videos); err != nil {
		ui.Status(fmt.Sprintf("Warning: could not enrich video details (duration unavailable): %v", err))
	}

	c := cache.New(cacheDir)

	total := len(videos)
	ui.Status(fmt.Sprintf("Playlist: %s  |  %d videos  |  concurrency: %d", playlistTitle, total, scanConcurrency))

	// Pre-scan quota estimate: ~3 Gemini API rounds per video on average.
	if qt != nil {
		const avgRoundsPerVideo = 3
		todayUsed := qt.Today()
		estimated := int64(total) * avgRoundsPerVideo
		rpd := int64(modelRPD(geminiModel, quotaTier))
		switch {
		case rpd == -1:
			// unlimited RPD — just show the estimate without a warning
			ui.Status(fmt.Sprintf("Quota estimate: ~%d API requests needed (%d videos × %d rounds avg)  |  today: %d used  |  RPD: Unlimited",
				estimated, total, avgRoundsPerVideo, todayUsed))
		case rpd > 0:
			remaining := rpd - todayUsed
			if remaining < 0 {
				remaining = 0
			}
			msg := fmt.Sprintf("Quota estimate: ~%d API requests needed (%d videos × %d rounds avg)  |  today: %d used, ~%d remaining (tier %s: %d/day)",
				estimated, total, avgRoundsPerVideo, todayUsed, remaining, quotaTier, rpd)
			if estimated > remaining {
				ui.Status("  ⚠  " + msg + " — quota may be exceeded mid-scan")
			} else {
				ui.Status(msg)
			}
		default:
			// Tier not configured — show count only
			ui.Status(fmt.Sprintf("Quota estimate: ~%d API requests needed (%d videos × %d rounds avg)  |  today: %d used  |  run 'vger config set quota_tier 1' for limit estimates",
				estimated, total, avgRoundsPerVideo, todayUsed))
		}
	}

	membership := domain.PlaylistMembership{
		PlaylistID:    playlistID,
		PlaylistTitle: playlistTitle,
		AddedAt:       time.Now(),
	}

	var (
		scanned         atomic.Int64
		skipped         atomic.Int64
		failed          atomic.Int64
		quotaSkipped    atomic.Int64
		durationSkipped atomic.Int64
		fallbackUsed    atomic.Int64
		quotaBreaker    atomic.Bool // set true when daily quota is exhausted
		mu              sync.Mutex
		wg              sync.WaitGroup
		sem             = make(chan struct{}, scanConcurrency)
	)

	// Wire fallback notifications so the user sees when a video is retried with
	// a larger-context model. The callback is goroutine-safe via mu.
	logFn := func(msg string) {
		mu.Lock()
		ui.Status("  ⚠  " + msg)
		mu.Unlock()
	}
	gmClient.LogFn = logFn
	if fallbackGmClient != nil {
		fallbackGmClient.LogFn = logFn
	}
	// Wire quota tracker: each GenerateContent call increments the counter.
	if qt != nil {
		gmClient.QuotaFn = func() { qt.Increment(1) }
		if fallbackGmClient != nil {
			fallbackGmClient.QuotaFn = func() { qt.Increment(1) }
		}
	}
	// Wire TPM tracker: record actual token usage after each successful response.
	if tpmWindow != nil {
		gmClient.TokenFn = func(tokens int64) { tpmWindow.Add(tokens) }
		if fallbackGmClient != nil {
			fallbackGmClient.TokenFn = func(tokens int64) { tpmWindow.Add(tokens) }
		}
	}

	for i, v := range videos {
		idx := i + 1
		videoURL := v.URL
		videoID := v.VideoID
		title := v.Title

		// Check cache before spinning up a goroutine.
		if !scanRefresh {
			cached, err := c.Load(cmd.Context(), videoID)
			if err == nil && cached != nil {
				// Record this playlist on the existing entry if not already present.
				if !cached.HasPlaylist(playlistID) {
					cached.Playlists = append(cached.Playlists, membership)
					_ = c.Save(cmd.Context(), cached) // best-effort
				}
				mu.Lock()
				ui.Status(fmt.Sprintf("[%*d/%d] Cached:   %s", len(fmt.Sprintf("%d", total)), idx, total, truncate(title, 60)))
				mu.Unlock()
				skipped.Add(1)
				continue
			}
		}

		// Skip videos that exceed the user-configured max duration.
		// Only applies when --max-duration is set (> 0) and duration was enriched.
		if scanMaxDuration > 0 && v.Duration != "" {
			durationSec := youtube.ParseISO8601Duration(v.Duration)
			if durationSec > 0 && durationSec > scanMaxDuration*60 {
				durMin := durationSec / 60
				mu.Lock()
				ui.Status(fmt.Sprintf("[%*d/%d] Skipped:  %s (duration %dm exceeds --max-duration %dm)",
					len(fmt.Sprintf("%d", total)), idx, total, truncate(title, 42), durMin, scanMaxDuration))
				mu.Unlock()
				durationSkipped.Add(1)
				continue
			}
		}

		// If the daily Gemini quota was exhausted by a previous video, stop
		// dispatching — the quota won't reset until tomorrow.
		if quotaBreaker.Load() {
			mu.Lock()
			ui.Status(fmt.Sprintf("[%*d/%d] Skipped:  %s (quota exhausted)", len(fmt.Sprintf("%d", total)), idx, total, truncate(title, 52)))
			mu.Unlock()
			quotaSkipped.Add(1)
			continue
		}

		// TPM throttle: wait for token budget headroom before dispatching.
		// This runs in the main loop (before spawning a goroutine) so we don't
		// hold the semaphore while sleeping.
		if tpmWindow != nil {
			estimated := estimateVideoTokens(v.Duration)
			current := tpmWindow.TokensLastMinute()
			if current+estimated > tpmLimit {
				mu.Lock()
				ui.Status(fmt.Sprintf("[%*d/%d] ⏳ TPM throttle: ~%dk tokens used/min (limit %dk), waiting for headroom before %s",
					len(fmt.Sprintf("%d", total)), idx, total,
					current/1000, tpmLimit/1000, truncate(title, 30)))
				mu.Unlock()
			}
			if err := tpmWindow.WaitForHeadroom(cmd.Context(), estimated, tpmLimit); err != nil {
				// Context cancelled — abort remaining dispatch.
				break
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
			if err != nil && errors.Is(err, gemini.ErrQuotaExhausted) && fallbackAgent != nil {
				mu.Lock()
				ui.Status(fmt.Sprintf("[%*d/%d] ⚠ %s quota exhausted — retrying with fallback model %s",
					len(fmt.Sprintf("%d", total)), idx, total, geminiModel, scanFallbackModel))
				mu.Unlock()
				report, meta, err = fallbackAgent.Run(cmd.Context(), videoURL)
				if err == nil {
					fallbackUsed.Add(1)
				}
			}
			if err != nil {
				if errors.Is(err, gemini.ErrQuotaExhausted) {
					quotaBreaker.Store(true)
				}
				mu.Lock()
				ui.Status(fmt.Sprintf("[%*d/%d] FAILED:   %s — %v", len(fmt.Sprintf("%d", total)), idx, total, truncate(title, 50), err))
				mu.Unlock()
				failed.Add(1)
				return
			}

			report.Stardate = ui.Stardate()
			entry := &domain.CachedAnalysis{
				VideoID:   videoID,
				CachedAt:  time.Now(),
				Report:    *report,
				Playlists: []domain.PlaylistMembership{membership},
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
	if quotaSkipped.Load() > 0 {
		ui.Field("quota-skipped", fmt.Sprintf("%d", quotaSkipped.Load()))
	}
	if fallbackUsed.Load() > 0 {
		ui.Field("fallback-used", fmt.Sprintf("%d (via %s)", fallbackUsed.Load(), scanFallbackModel))
	}
	if durationSkipped.Load() > 0 {
		ui.Field("duration-skipped", fmt.Sprintf("%d (--max-duration %dm)", durationSkipped.Load(), scanMaxDuration))
	}
	ui.Field("total  ", fmt.Sprintf("%d", total))
	fmt.Println()

	if quotaBreaker.Load() {
		return fmt.Errorf("Gemini daily quota exhausted after %d video(s) — re-run tomorrow to continue (%d skipped, cached videos are skipped automatically)", failed.Load(), quotaSkipped.Load())
	}
	if failed.Load() > 0 {
		return fmt.Errorf("%d video(s) failed to analyse — re-run the same command to retry (cached videos are skipped automatically)", failed.Load())
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
	if len(report.Speakers) > 0 {
		ui.Field("speakers", strings.Join(report.Speakers, ", "))
	}

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
