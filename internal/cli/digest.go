package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/costap/vger/internal/adapters/cache"
	"github.com/costap/vger/internal/adapters/gemini"
	"github.com/costap/vger/internal/adapters/youtube"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/costap/vger/internal/domain"
	"github.com/spf13/cobra"
)

var digestPlaylist string
var digestAI bool
var digestOutput string

var digestCmd = &cobra.Command{
	Use:   "digest",
	Short: "Produce an overview digest of all cached talks in a playlist",
	Long: `Produce an overview digest of all previously scanned talks in a playlist.

Layer 1 (always): reads cached analyses — compact talk table and technology radar showing
which technologies appeared across how many talks. Zero API cost.

Layer 2 (--ai flag): sends all summaries to Gemini for cross-playlist synthesis —
overarching theme, recommended learning path, and priority talks to watch first.

Use --output to write a Markdown report file.

Examples:
  vger digest --playlist "https://www.youtube.com/playlist?list=PLj6h78yzYM2P..."
  vger digest --playlist PLj6h78yzYM2P... --ai
  vger digest --playlist PLj6h78yzYM2P... --ai --output kubecon2024.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if digestPlaylist == "" {
			return fmt.Errorf("--playlist is required")
		}

		ytClient := youtube.New(youtubeAPIKey)

		playlistID, err := ytClient.ExtractPlaylistID(digestPlaylist)
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

		ui.Status(fmt.Sprintf("Found %d videos — loading cached analyses...", len(videos)))

		cacheDir, err := cache.DefaultDir()
		if err != nil {
			ui.RedAlert(err)
			return err
		}
		c := cache.New(cacheDir)

		ids := make([]string, len(videos))
		for i, v := range videos {
			ids[i] = v.VideoID
		}

		entries, err := c.LoadByVideoIDs(cmd.Context(), ids)
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		missing := len(videos) - len(entries)
		if len(entries) == 0 {
			ui.Complete(fmt.Sprintf("No cached analyses found. Run: vger scan --playlist %s", playlistID))
			return nil
		}

		if missing > 0 {
			ui.Status(fmt.Sprintf("Warning: %d video(s) not yet scanned — run vger scan --playlist to fill gaps.", missing))
		}

		ui.Status(fmt.Sprintf("Loaded %d cached analyses.", len(entries)))

		// Build technology frequency map
		techFreq := buildTechFrequency(entries)

		// Render the local digest (Layer 1)
		renderDigestTable(entries, techFreq)

		// Layer 2: AI synthesis
		var digest *domain.DigestReport
		if digestAI {
			ui.Status("Transmitting summaries to Gemini synthesis array...")
			gmClient := gemini.New(geminiAPIKey, geminiModel)
			digest, err = gmClient.Synthesise(cmd.Context(), entries, userContext)
			if err != nil {
				ui.RedAlert(err)
				return err
			}
			renderDigestAI(digest)
		}

		// Optional Markdown export
		if digestOutput != "" {
			if err := writeMarkdownReport(digestOutput, playlistID, entries, techFreq, digest); err != nil {
				ui.Status(fmt.Sprintf("Warning: could not write report: %v", err))
			} else {
				ui.Status(fmt.Sprintf("Report written to: %s", digestOutput))
			}
		}

		ui.Complete("Digest complete. Captain.")
		return nil
	},
}

func init() {
	digestCmd.Flags().StringVar(&digestPlaylist, "playlist", "", "Playlist ID or URL (required)")
	digestCmd.Flags().BoolVar(&digestAI, "ai", false, "Use Gemini to synthesise a cross-playlist learning path")
	digestCmd.Flags().StringVar(&digestOutput, "output", "", "Write a Markdown report to this file")
}

// buildTechFrequency counts how many talks mention each technology.
func buildTechFrequency(entries []*domain.CachedAnalysis) []domain.TechCount {
	freq := make(map[string]int)
	for _, e := range entries {
		for _, t := range e.Report.Technologies {
			freq[t.Name]++
		}
	}
	counts := make([]domain.TechCount, 0, len(freq))
	for name, count := range freq {
		counts = append(counts, domain.TechCount{Name: name, Count: count})
	}
	sort.Slice(counts, func(i, j int) bool {
		if counts[i].Count != counts[j].Count {
			return counts[i].Count > counts[j].Count
		}
		return counts[i].Name < counts[j].Name
	})
	return counts
}

// renderDigestTable renders the talk table and technology radar.
func renderDigestTable(entries []*domain.CachedAnalysis, techFreq []domain.TechCount) {
	ui.SectionHeader(fmt.Sprintf("Talk Overview  (%d talks)", len(entries)))
	fmt.Println()

	for i, e := range entries {
		topTechs := topN(e.Report.Technologies, 3)
		dur := ""
		if e.Report.VideoDurationSec > 0 {
			dur = formatDuration(e.Report.VideoDurationSec)
		}
		fmt.Printf("  %s  %s  %s\n",
			ui.LabelStyle().Render(fmt.Sprintf("%3d.", i+1)),
			ui.DimStyle().Render(fmt.Sprintf("%-8s", dur)),
			truncate(e.Report.VideoTitle, 55),
		)
		if len(topTechs) > 0 {
			fmt.Printf("       %s\n", ui.DimStyle().Render(strings.Join(topTechs, " · ")))
		}
		fmt.Printf("       %s\n", ui.DimStyle().Render(e.Report.VideoURL))
		fmt.Println()
	}

	// Technology radar
	if len(techFreq) > 0 {
		ui.SectionHeader("Technology Radar")
		fmt.Println()
		maxCount := techFreq[0].Count
		barWidth := 30
		limit := 20
		if len(techFreq) < limit {
			limit = len(techFreq)
		}
		for _, tc := range techFreq[:limit] {
			filled := int(float64(tc.Count) / float64(maxCount) * float64(barWidth))
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
			fmt.Printf("  %-25s %s %s\n",
				ui.LabelStyle().Render(truncate(tc.Name, 24)),
				ui.DimStyle().Render(bar),
				fmt.Sprintf("%d", tc.Count),
			)
		}
		fmt.Println()
	}
}

// renderDigestAI renders the Gemini synthesis section.
func renderDigestAI(d *domain.DigestReport) {
	ui.SectionHeader("AI Synthesis")

	if d.OverarchingTheme != "" {
		fmt.Printf("\n  %s\n  %s\n\n", ui.LabelStyle().Render("OVERARCHING THEME"), d.OverarchingTheme)
	}

	if len(d.LearningPath) > 0 {
		fmt.Printf("  %s\n", ui.LabelStyle().Render("RECOMMENDED LEARNING PATH"))
		for i, tech := range d.LearningPath {
			fmt.Printf("  %s %s\n", ui.DimStyle().Render(fmt.Sprintf("%2d.", i+1)), tech)
		}
		fmt.Println()
	}

	if len(d.PriorityTalks) > 0 {
		fmt.Printf("  %s\n", ui.LabelStyle().Render("WATCH THESE FIRST"))
		for i, pt := range d.PriorityTalks {
			fmt.Printf("\n  %s %s\n", ui.LabelStyle().Render(fmt.Sprintf("%d.", i+1)), pt.Title)
			fmt.Printf("     %s\n", ui.DimStyle().Render(pt.URL))
			fmt.Printf("     %s\n", pt.Reason)
		}
		fmt.Println()
	}

	if d.KeyInsights != "" {
		fmt.Printf("  %s\n  %s\n\n", ui.LabelStyle().Render("KEY INSIGHTS"), d.KeyInsights)
	}
}

// writeMarkdownReport writes a full Markdown digest to a file.
func writeMarkdownReport(path, playlistID string, entries []*domain.CachedAnalysis, techFreq []domain.TechCount, digest *domain.DigestReport) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Playlist Digest — %s\n\n", playlistID))
	sb.WriteString(fmt.Sprintf("*%d talks analysed by V'Ger*\n\n", len(entries)))

	if digest != nil && digest.OverarchingTheme != "" {
		sb.WriteString("## Overarching Theme\n\n")
		sb.WriteString(digest.OverarchingTheme + "\n\n")
	}

	if digest != nil && len(digest.PriorityTalks) > 0 {
		sb.WriteString("## Watch These First\n\n")
		for i, pt := range digest.PriorityTalks {
			sb.WriteString(fmt.Sprintf("%d. **[%s](%s)**  \n   %s\n\n", i+1, pt.Title, pt.URL, pt.Reason))
		}
	}

	if digest != nil && len(digest.LearningPath) > 0 {
		sb.WriteString("## Recommended Learning Path\n\n")
		for i, tech := range digest.LearningPath {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, tech))
		}
		sb.WriteString("\n")
	}

	if digest != nil && digest.KeyInsights != "" {
		sb.WriteString("## Key Insights\n\n")
		sb.WriteString(digest.KeyInsights + "\n\n")
	}

	if len(techFreq) > 0 {
		sb.WriteString("## Technology Radar\n\n")
		sb.WriteString("| Technology | Talks |\n|---|---|\n")
		limit := 20
		if len(techFreq) < limit {
			limit = len(techFreq)
		}
		for _, tc := range techFreq[:limit] {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", tc.Name, tc.Count))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## All Talks\n\n")
	for i, e := range entries {
		topTechs := topN(e.Report.Technologies, 5)
		dur := ""
		if e.Report.VideoDurationSec > 0 {
			dur = " · " + formatDuration(e.Report.VideoDurationSec)
		}
		sb.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, e.Report.VideoTitle))
		sb.WriteString(fmt.Sprintf("[%s](%s)%s\n\n", e.Report.VideoURL, e.Report.VideoURL, dur))
		sb.WriteString(e.Report.Summary + "\n\n")
		if len(topTechs) > 0 {
			sb.WriteString("**Technologies:** " + strings.Join(topTechs, ", ") + "\n\n")
		}
	}

	return os.WriteFile(path, []byte(sb.String()), 0o644)
}

func topN(techs []domain.Technology, n int) []string {
	names := make([]string, 0, n)
	for i, t := range techs {
		if i >= n {
			break
		}
		names = append(names, t.Name)
	}
	return names
}
