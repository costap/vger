package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/costap/vger/internal/adapters/cache"
	"github.com/costap/vger/internal/adapters/signals"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/spf13/cobra"
)

var trackLinkVideo string

var trackLinkCmd = &cobra.Command{
	Use:   "link <id> --video <youtube-url>",
	Short: "Link a signal to a vger video scan",
	Long: `Associate a tracked signal with a previously scanned conference talk.

The video must already be in the vger cache (run vger scan first).
Multiple links can be added to the same signal.

Examples:
  vger track link 0001 --video https://www.youtube.com/watch?v=abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := fmt.Sprintf("%04s", args[0])

		if trackLinkVideo == "" {
			err := fmt.Errorf("--video is required")
			ui.RedAlert(err)
			return err
		}

		// Resolve video ID from URL or raw ID.
		videoID := extractVideoID(trackLinkVideo)
		if videoID == "" {
			err := fmt.Errorf("could not extract video ID from: %s", trackLinkVideo)
			ui.RedAlert(err)
			return err
		}

		// Verify the video is in the cache.
		cacheDir, err := cache.DefaultDir()
		if err != nil {
			ui.RedAlert(err)
			return err
		}
		c := cache.New(cacheDir)
		entry, err := c.Load(cmd.Context(), videoID)
		if err != nil {
			ui.RedAlert(err)
			return err
		}
		if entry == nil {
			err := fmt.Errorf("video %s is not in the vger cache — run: vger scan %s", videoID, trackLinkVideo)
			ui.RedAlert(err)
			return err
		}

		// Load the signal.
		sigDir, err := signals.DefaultDir()
		if err != nil {
			ui.RedAlert(err)
			return err
		}
		store := signals.New(sigDir)

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

		// Check for duplicate.
		for _, existing := range sig.LinkedVideoIDs {
			if existing == videoID {
				ui.Complete(fmt.Sprintf("video %s already linked to signal %s", videoID, id))
				return nil
			}
		}

		sig.LinkedVideoIDs = append(sig.LinkedVideoIDs, videoID)
		sig.UpdatedAt = time.Now().UTC()

		if err := store.Save(cmd.Context(), sig); err != nil {
			ui.RedAlert(err)
			return err
		}

		ui.Field("Signal", sig.ID+" — "+sig.Title)
		ui.Field("Linked Video", videoID+" — "+entry.Report.VideoTitle)
		ui.Field("All Links", strings.Join(sig.LinkedVideoIDs, ", "))
		ui.Complete("link saved")
		return nil
	},
}

func init() {
	trackLinkCmd.Flags().StringVar(&trackLinkVideo, "video", "", "YouTube URL or video ID to link")
}

// extractVideoID pulls the video ID from a YouTube URL or returns the raw string
// if it looks like a bare video ID (11 characters).
func extractVideoID(input string) string {
	// Try URL parsing first (handles v= parameter).
	if strings.Contains(input, "youtube.com") || strings.Contains(input, "youtu.be") {
		for _, part := range strings.Split(input, "?") {
			for _, kv := range strings.Split(part, "&") {
				if strings.HasPrefix(kv, "v=") {
					return strings.TrimPrefix(kv, "v=")
				}
			}
		}
		// youtu.be/<id>
		if strings.Contains(input, "youtu.be/") {
			parts := strings.Split(input, "youtu.be/")
			if len(parts) == 2 {
				return strings.Split(parts[1], "?")[0]
			}
		}
	}
	// Bare video ID (YouTube IDs are 11 chars).
	if len(input) == 11 && !strings.Contains(input, "/") {
		return input
	}
	return ""
}
