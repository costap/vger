package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/costap/vger/internal/adapters/cache"
	"github.com/costap/vger/internal/adapters/cncf"
	"github.com/costap/vger/internal/adapters/gemini"
	"github.com/costap/vger/internal/adapters/signals"
	"github.com/costap/vger/internal/adapters/youtube"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/costap/vger/internal/domain"
)

var (
	researchDiscover  bool
	researchAnywhere  bool
	researchChannel   string
	researchLens      string
	researchMaxVideos int
	researchMaxDepth  int
	researchOutput    string
)

// signalSearcher is a local interface satisfied by both signals.JSONStore and
// signals.MarkdownStore via their Search() method added in search.go.
type signalSearcher interface {
	Search(ctx context.Context, query string) ([]*domain.Signal, error)
}

var researchLongDesc = `Research a technology topic by searching all available sources:
  • Local cache of previously scanned conference talks
  • CNCF landscape for related projects
  • Your tracked technology signals
  • YouTube talks (optional, with --discover or --anywhere)

The findings are synthesised by Gemini into a structured research brief with
a landscape map, evidence from cached videos, investigation paths, and a verdict.

Use --lens to apply an analytical preset to the synthesis (architect, engineer,
radar, brief). Use --discover to also search YouTube for unscanned relevant talks.

Use --anywhere to search all of YouTube instead of a specific channel.
Note: --anywhere uses the YouTube search.list API (100 quota units/call).

Use --max-depth to enable Phase 2 investigation: Gemini will autonomously query
cached videos and search for additional evidence before synthesising the report.
  --max-depth 0  fast single-pass synthesis (default)
  --max-depth 3  light investigation (3 tool rounds)
  --max-depth 5  standard investigation (recommended)
  --max-depth 8  deep investigation (broader topics, higher cost)

Examples:
  vger research "eBPF"
  vger research "multi-cluster networking" --discover
  vger research "multi-cluster networking" --anywhere
  vger research "WASM in Kubernetes" --lens architect
  vger research "service mesh" --max-depth 5
  vger research "service mesh" --output service-mesh-brief.md`

var researchCmd = &cobra.Command{
	Use:   "research <topic>",
	Short: "Search all sources about a topic and synthesise investigation paths",
	Long:  researchLongDesc,
	Args:  cobra.ExactArgs(1),
	RunE:  runResearch,
}

func runResearch(cmd *cobra.Command, args []string) error {
	topic := args[0]
	ctx := cmd.Context()

	cacheDir, err := cache.DefaultDir()
	if err != nil {
		ui.RedAlert(err)
		return err
	}

	cacheClient := cache.New(cacheDir)
	cncfClient := cncf.New(cacheDir)
	gmClient := gemini.NewWithTools(geminiAPIKey, geminiModel, cncfClient)

	// Resolve signal store (mirrors track command logic).
	sigStore, err := resolveResearchSignalStore()
	if err != nil {
		// Signal search is best-effort — continue without it.
		sigStore = nil
	}

	// --- Phase 1: Search local cache ---
	ui.Status(fmt.Sprintf("Searching knowledge cache for %q…", topic))
	hits, err := cacheClient.Search(ctx, topic, researchMaxVideos)
	if err != nil {
		ui.RedAlert(err)
		return err
	}

	// --- Phase 2: CNCF landscape lookup ---
	ui.Status("Scanning CNCF landscape…")
	projects := cncfClient.LookupByTopic(ctx, topic)

	// --- Phase 3: Tracked signals ---
	var matchedSignals []*domain.Signal
	if sigStore != nil {
		ui.Status("Checking tracked signals…")
		matchedSignals, _ = sigStore.Search(ctx, topic)
	}

	// --- Phase 4: YouTube discovery (--discover / --anywhere) ---
	var discoveredTalks []domain.VideoListing
	if researchDiscover || researchAnywhere {
		ytClient := youtube.New(youtubeAPIKey)
		var all []domain.VideoListing

		if researchAnywhere {
			ui.Status(fmt.Sprintf("Searching all of YouTube for %q… (uses search.list quota)", topic))
			all, err = ytClient.SearchVideos(ctx, topic, 20)
			if err != nil {
				ui.RedAlert(err)
				return err
			}
		} else {
			ui.Status(fmt.Sprintf("Discovering unscanned talks on %q…", topic))
			channelRef := researchChannel
			channelID, channelName, err := ytClient.ResolveChannel(ctx, channelRef)
			if err != nil {
				ui.RedAlert(fmt.Errorf("resolve channel %q: %w", channelRef, err))
				return err
			}
			saveChannelHistory(channelRef, channelID, channelName)
			all, err = ytClient.ListVideos(ctx, channelID, topic, 20)
			if err != nil {
				ui.RedAlert(err)
				return err
			}
		}

		// Exclude videos already in the local cache.
		cachedIDs := make(map[string]bool, len(hits))
		for _, h := range hits {
			cachedIDs[h.VideoID] = true
		}
		for _, v := range all {
			if !cachedIDs[v.VideoID] {
				discoveredTalks = append(discoveredTalks, v)
			}
		}
	}

	// --- Phase 5: Gemini synthesis ---
	if researchMaxDepth > 0 {
		ui.Status(fmt.Sprintf("Running deep investigation (max depth %d)…", researchMaxDepth))
	} else {
		ui.Status("Synthesising research brief…")
	}

	var lc *gemini.LensContext
	if researchLens != "" {
		lens, ok := lookupLens(researchLens)
		if !ok {
			return fmt.Errorf("unknown lens %q — available: %s", researchLens, lensNames())
		}
		lc = &gemini.LensContext{RoleContext: lens.RoleContext}
	}

	report, err := gmClient.ResearchSynthesize(ctx, topic, hits, projects, matchedSignals, discoveredTalks, lc, researchMaxDepth, cacheClient)
	if err != nil {
		ui.RedAlert(err)
		return err
	}

	// Attach signal summaries (not returned by Gemini, sourced locally).
	report.RelatedSignals = signalSummaries(matchedSignals)

	ui.Complete("Research brief ready.")

	// --- Phase 6: Render ---
	ui.RenderResearchReport(report)

	if researchOutput != "" {
		if err := writeResearchMarkdown(report, researchOutput); err != nil {
			ui.RedAlert(fmt.Errorf("write output: %w", err))
			return err
		}
		ui.Complete(fmt.Sprintf("Report saved to %s", researchOutput))
	}

	return nil
}

// resolveResearchSignalStore returns a signalSearcher backed by the appropriate
// store. Mirrors the logic in track/cmd.go. Returns nil (no error) when no
// store is available — signal search is best-effort for the research command.
func resolveResearchSignalStore() (signalSearcher, error) {
	dir := os.Getenv("TECHDR_DIR")
	if dir == "" {
		home, _ := os.UserHomeDir()
		candidate := home + "/code/github.com/costap/tech-signals"
		if _, err := os.Stat(candidate + "/.next-id"); err == nil {
			dir = candidate
		}
	}
	if dir != "" {
		if _, err := os.Stat(dir); err != nil {
			return nil, nil // TECHDR_DIR missing — silently skip
		}
		return signals.NewMarkdownStore(dir), nil
	}

	jsonDir, err := signals.DefaultDir()
	if err != nil {
		return nil, nil
	}
	return signals.New(jsonDir), nil
}

func init() {
	researchCmd.Flags().BoolVar(&researchDiscover, "discover", false, "Search YouTube for unscanned relevant talks (channel-restricted)")
	researchCmd.Flags().BoolVar(&researchAnywhere, "anywhere", false, "Search all of YouTube (uses search.list; implies --discover; ignores --channel)")
	researchCmd.Flags().StringVar(&researchChannel, "channel", "@cncf", "YouTube channel to search when --discover is used (ignored with --anywhere)")
	researchCmd.Flags().StringVar(&researchLens, "lens", "", fmt.Sprintf("Apply a built-in analytical lens (%s)", lensNames()))
	researchCmd.Flags().IntVar(&researchMaxVideos, "max-videos", 10, "Maximum cached videos to include in context")
	researchCmd.Flags().IntVar(&researchMaxDepth, "max-depth", 0, "Investigation depth: 0=fast single-pass, 3=light, 5=standard, 8=deep (Phase 2)")
	researchCmd.Flags().StringVar(&researchOutput, "output", "", "Write report to a Markdown file")

	_ = researchCmd.RegisterFlagCompletionFunc("lens", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		names := make([]string, len(builtinLenses))
		for i, l := range builtinLenses {
			names[i] = l.Name + "\t" + l.ShortDesc
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	})
	_ = researchCmd.RegisterFlagCompletionFunc("channel", channelCompletionFunc)
}

// signalSummaries converts matched signals into compact SignalSummary values.
func signalSummaries(sigs []*domain.Signal) []domain.SignalSummary {
	out := make([]domain.SignalSummary, len(sigs))
	for i, s := range sigs {
		out[i] = domain.SignalSummary{
			ID:       s.ID,
			Title:    s.Title,
			Status:   s.Status,
			Category: s.Category,
			Note:     s.Note,
		}
	}
	return out
}

// writeResearchMarkdown writes the ResearchReport as a Markdown file.
func writeResearchMarkdown(r *domain.ResearchReport, path string) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Research Brief: %s\n\n", r.Topic))
	sb.WriteString(fmt.Sprintf("%s\n\n", r.Brief))

	if len(r.LandscapeMap) > 0 {
		sb.WriteString("## CNCF Landscape\n\n")
		sb.WriteString("| Project | Stage | Category | Relevance |\n")
		sb.WriteString("|---------|-------|----------|-----------|\n")
		for _, p := range r.LandscapeMap {
			stage := p.CNCFStage
			if stage == "" {
				stage = "—"
			}
			sb.WriteString(fmt.Sprintf("| [%s](%s) | %s | %s | %s |\n",
				p.Name, p.Homepage, stage, p.Category, p.Relevance))
		}
		sb.WriteString("\n")
	}

	if len(r.EvidenceVideos) > 0 {
		sb.WriteString("## Evidence from Cache\n\n")
		for i, e := range r.EvidenceVideos {
			sb.WriteString(fmt.Sprintf("%d. **[%s](%s)** — %s\n", i+1, e.VideoTitle, e.VideoURL, e.Relevance))
		}
		sb.WriteString("\n")
	}

	if len(r.RelatedSignals) > 0 {
		sb.WriteString("## Tracked Signals\n\n")
		for _, s := range r.RelatedSignals {
			sb.WriteString(fmt.Sprintf("- **%s** `%s` [%s, %s]", s.ID, s.Title, s.Status, s.Category))
			if s.Note != "" {
				sb.WriteString(fmt.Sprintf(": %s", s.Note))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	if len(r.InvestigationPaths) > 0 {
		sb.WriteString("## Investigation Paths\n\n")
		for i, p := range r.InvestigationPaths {
			sb.WriteString(fmt.Sprintf("### Path %c: %s\n\n", 'A'+rune(i), p.Title))
			sb.WriteString(fmt.Sprintf("%s\n\n", p.Description))
			for _, a := range p.Actions {
				sb.WriteString(fmt.Sprintf("- %s\n", a))
			}
			sb.WriteString("\n")
		}
	}

	if len(r.CompetingApproaches) > 0 {
		sb.WriteString("## Competing Approaches\n\n")
		for _, a := range r.CompetingApproaches {
			sb.WriteString(fmt.Sprintf("- %s\n", a))
		}
		sb.WriteString("\n")
	}

	if r.Verdict != "" {
		sb.WriteString("## Verdict\n\n")
		sb.WriteString(fmt.Sprintf("%s\n\n", r.Verdict))
	}

	if len(r.DiscoveredTalks) > 0 {
		sb.WriteString("## Undiscovered Talks\n\n")
		for i, t := range r.DiscoveredTalks {
			sb.WriteString(fmt.Sprintf("%d. [%s](%s)\n", i+1, t.Title, t.URL))
		}
		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0o640)
}
