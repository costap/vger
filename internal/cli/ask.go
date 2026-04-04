package cli

import (
	"fmt"
	"strings"

	"github.com/costap/vger/internal/adapters/cache"
	"github.com/costap/vger/internal/adapters/gemini"
	"github.com/costap/vger/internal/adapters/youtube"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/spf13/cobra"
)

var askDeep bool
var askLens string

var askCmd = &cobra.Command{
	Use:   "ask <youtube-url> [question]",
	Short: "Ask a follow-up question about a previously scanned video",
	Long: `Ask a follow-up question about a video.

By default, the answer is generated from the cached analysis (fast, cheap).
The cache includes a detailed notes section that covers everything mentioned
in the video, so most questions can be answered without re-uploading.

Use --deep to re-submit the video URL to Gemini for questions that require
direct access to the full video content (exact quotes, timestamps, details
not captured in the notes). This is slower and uses more tokens.

Use --lens <name> to apply a built-in analytical preset so you don't have to
type the same verbose prompt each time. The question argument becomes optional
when a lens is set — the lens provides a default question. If you supply both
a lens and a question, the lens role context prefixes your question.

Available lenses: ` + lensNames() + `

The video must have been scanned with 'vger scan' at least once.`,
	Args: func(cmd *cobra.Command, args []string) error {
		lens, _ := cmd.Flags().GetString("lens")
		if lens != "" {
			// Lens set: URL required, question optional
			if len(args) < 1 || len(args) > 2 {
				return fmt.Errorf("accepts 1 or 2 args (url [question]) when --lens is set, received %d", len(args))
			}
			return nil
		}
		// No lens: both URL and question required
		if len(args) != 2 {
			return fmt.Errorf("accepts 2 args (url question), received %d — or use --lens to skip the question", len(args))
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		userQuestion := ""
		if len(args) > 1 {
			userQuestion = args[1]
		}

		// Resolve question: lens overrides or prefixes user question.
		question := userQuestion
		lensLabel := ""
		if askLens != "" {
			lens, ok := lookupLens(askLens)
			if !ok {
				return fmt.Errorf("unknown lens %q — available: %s", askLens, lensNames())
			}
			question = buildLensPrompt(lens, userQuestion)
			lensLabel = fmt.Sprintf(" [lens: %s]", lens.Name)
		}
		if strings.TrimSpace(question) == "" {
			return fmt.Errorf("a question is required (provide as argument or use --lens)")
		}

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
			ui.Status(fmt.Sprintf("Deep mode%s: re-transmitting video to Gemini for direct analysis...", lensLabel))
			answer, err = gmClient.AskDeep(cmd.Context(), question, url, cached)
		} else {
			ui.Status(fmt.Sprintf("Transmitting question to Gemini knowledge core%s...", lensLabel))
			answer, err = gmClient.Ask(cmd.Context(), question, cached)
		}

		if err != nil {
			ui.RedAlert(err)
			return err
		}

		ui.Complete("Response received.")

		ui.SectionHeader("Question")
		displayQ := userQuestion
		if displayQ == "" && askLens != "" {
			displayQ = fmt.Sprintf("[%s lens default question]", askLens)
		}
		fmt.Printf("\n  %s\n", ui.LabelStyle().Render(displayQ))
		if askLens != "" {
			fmt.Printf("  %s\n", ui.DimStyle().Render(fmt.Sprintf("lens: %s", askLens)))
		}

		ui.SectionHeader("Answer")
		fmt.Printf("\n  %s\n\n", answer)

		return nil
	},
}

func init() {
	askCmd.Flags().BoolVar(&askDeep, "deep", false, "Re-submit the video to Gemini for questions beyond the cached notes")
	askCmd.Flags().StringVar(&askLens, "lens", "", fmt.Sprintf("Apply a built-in analytical preset (%s)", lensNames()))
}


