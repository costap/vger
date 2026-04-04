package cli

import (
	"os"

	"github.com/costap/vger/internal/cli/track"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/spf13/cobra"
)

var geminiAPIKey string
var youtubeAPIKey string
var geminiModel string

// Root is the top-level cobra command.
var Root = &cobra.Command{
	Use:   "vger",
	Short: "V'Ger — conference video knowledge assimilation system",
	Long: `V'Ger ingests online conference videos (KubeCon, CloudNativeCon, etc.)
and produces structured summaries with technology learning recommendations.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Resolve keys: explicit flag > env var (populated by godotenv in main).
		if geminiAPIKey == "" {
			geminiAPIKey = os.Getenv("GEMINI_API_KEY")
		}
		if youtubeAPIKey == "" {
			youtubeAPIKey = os.Getenv("YOUTUBE_API_KEY")
		}
		ui.Header()
	},
}

func init() {
	Root.PersistentFlags().StringVar(&geminiAPIKey, "gemini-key", "", "Gemini API key (env: GEMINI_API_KEY, .env)")
	Root.PersistentFlags().StringVar(&youtubeAPIKey, "youtube-key", "", "YouTube Data API key (env: YOUTUBE_API_KEY, .env)")
	Root.PersistentFlags().StringVar(&geminiModel, "model", "gemini-2.5-flash", "Gemini model to use")

	Root.AddCommand(scanCmd)
	Root.AddCommand(listCmd)
	Root.AddCommand(askCmd)
	Root.AddCommand(digestCmd)
	Root.AddCommand(track.Cmd())
}

