package cli

import (
	"fmt"

	"github.com/costap/vger/internal/adapters/cache"
	"github.com/costap/vger/internal/adapters/gemini"
	"github.com/costap/vger/internal/adapters/youtube"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/spf13/cobra"
)

var askDeep bool

var askCmd = &cobra.Command{
	Use:   "ask <youtube-url> <question>",
	Short: "Ask a follow-up question about a previously scanned video",
	Long: `Ask a follow-up question about a video.

By default, the answer is generated from the cached analysis (fast, cheap).
The cache includes a detailed notes section that covers everything mentioned
in the video, so most questions can be answered without re-uploading.

Use --deep to re-submit the video URL to Gemini for questions that require
direct access to the full video content (exact quotes, timestamps, details
not captured in the notes). This is slower and uses more tokens.

The video must have been scanned with 'vger scan' at least once.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		question := args[1]

		ytClient := youtube.New(youtubeAPIKey)

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

		ui.Status(fmt.Sprintf("Loading cached analysis for video: %s", videoID))

		cached, err := c.Load(cmd.Context(), videoID)
		if err != nil {
			ui.RedAlert(err)
			return err
		}
		if cached == nil {
			err := fmt.Errorf("no cached analysis found for %s — run 'vger scan %s' first", videoID, url)
			ui.RedAlert(err)
			return err
		}

		ui.Status(fmt.Sprintf("Cache hit: %s (cached %s)", cached.Report.VideoTitle, cached.CachedAt.Format("2006-01-02")))

		gmClient := gemini.New(geminiAPIKey, geminiModel)
		var answer string

		if askDeep {
			ui.Status("Deep mode: re-transmitting video to Gemini for direct analysis...")
			answer, err = gmClient.AskDeep(cmd.Context(), question, url, cached)
		} else {
			ui.Status("Transmitting question to Gemini knowledge core...")
			answer, err = gmClient.Ask(cmd.Context(), question, cached)
		}

		if err != nil {
			ui.RedAlert(err)
			return err
		}

		ui.Complete("Response received.")

		ui.SectionHeader("Question")
		fmt.Printf("\n  %s\n", ui.LabelStyle().Render(question))

		ui.SectionHeader("Answer")
		fmt.Printf("\n  %s\n\n", answer)

		return nil
	},
}

func init() {
	askCmd.Flags().BoolVar(&askDeep, "deep", false, "Re-submit the video to Gemini for questions beyond the cached notes")
}

