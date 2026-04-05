package ui

import (
	"fmt"
	"strings"

	"github.com/costap/vger/internal/domain"
)

// RenderResearchReport prints a LCARS-styled research brief to the terminal.
func RenderResearchReport(r *domain.ResearchReport) {
	fmt.Println()
	SectionHeader(fmt.Sprintf("RESEARCH BRIEF: %q", strings.ToUpper(r.Topic)))

	// Brief.
	fmt.Println()
	fmt.Printf("  %s\n\n", bodyStyle.Render(r.Brief))

	// CNCF Landscape.
	if len(r.LandscapeMap) > 0 {
		SectionHeader(fmt.Sprintf("CNCF LANDSCAPE  (%d projects)", len(r.LandscapeMap)))
		fmt.Println()
		for _, p := range r.LandscapeMap {
			stage := p.CNCFStage
			if stage == "" {
				stage = "—"
			}
			stageStyle := dimStyle
			switch stage {
			case "graduated":
				stageStyle = GreenStyle()
			case "incubating":
				stageStyle = blueStyle
			case "sandbox":
				stageStyle = labelStyle
			}
			fmt.Printf("  %s  %s  %s\n",
				stageStyle.Render(fmt.Sprintf("%-12s", strings.ToUpper(stage))),
				labelStyle.Render(fmt.Sprintf("%-30s", p.Name)),
				dimStyle.Render(p.Category),
			)
			if p.Relevance != "" {
				fmt.Printf("    %s\n", dimStyle.Render(p.Relevance))
			}
		}
		fmt.Println()
	}

	// Evidence from cache.
	if len(r.EvidenceVideos) > 0 {
		SectionHeader(fmt.Sprintf("EVIDENCE FROM CACHE  (%d talks)", len(r.EvidenceVideos)))
		fmt.Println()
		for i, e := range r.EvidenceVideos {
			fmt.Printf("  %s  %s\n",
				labelStyle.Render(fmt.Sprintf("[%d]", i+1)),
				bodyStyle.Render(e.VideoTitle),
			)
			if len(e.Speakers) > 0 {
				fmt.Printf("    %s\n", dimStyle.Render("Speakers: "+strings.Join(e.Speakers, ", ")))
			}
			if e.Relevance != "" {
				fmt.Printf("    %s\n", dimStyle.Render(e.Relevance))
			}
			if e.VideoURL != "" {
				fmt.Printf("    %s\n", blueStyle.Render(e.VideoURL))
			}
		}
		fmt.Println()
	}

	// Related signals.
	if len(r.RelatedSignals) > 0 {
		SectionHeader(fmt.Sprintf("TRACKED SIGNALS  (%d)", len(r.RelatedSignals)))
		fmt.Println()
		for _, s := range r.RelatedSignals {
			fmt.Printf("  %s  %s  %s  %s\n",
				labelStyle.Render(s.ID),
				bodyStyle.Render(s.Title),
				dimStyle.Render(fmt.Sprintf("[%s]", s.Status)),
				dimStyle.Render(s.Category),
			)
			if s.Note != "" {
				note := s.Note
				if len(note) > 120 {
					note = note[:120] + "…"
				}
				fmt.Printf("    %s\n", dimStyle.Render(note))
			}
		}
		fmt.Println()
	}

	// Investigation paths.
	if len(r.InvestigationPaths) > 0 {
		SectionHeader("INVESTIGATION PATHS")
		fmt.Println()
		for i, p := range r.InvestigationPaths {
			fmt.Printf("  %s  %s\n",
				labelStyle.Render(fmt.Sprintf("PATH %c", 'A'+rune(i))),
				bodyStyle.Render(p.Title),
			)
			if p.Description != "" {
				fmt.Printf("    %s\n", dimStyle.Render(p.Description))
			}
			for _, a := range p.Actions {
				fmt.Printf("    %s  %s\n", dimStyle.Render("→"), bodyStyle.Render(a))
			}
			fmt.Println()
		}
	}

	// Competing approaches.
	if len(r.CompetingApproaches) > 0 {
		SectionHeader("COMPETING APPROACHES")
		fmt.Println()
		for _, a := range r.CompetingApproaches {
			fmt.Printf("  %s  %s\n", dimStyle.Render("●"), bodyStyle.Render(a))
		}
		fmt.Println()
	}

	// Verdict.
	if r.Verdict != "" {
		SectionHeader("VERDICT")
		fmt.Println()
		fmt.Printf("  %s\n\n", bodyStyle.Render(r.Verdict))
	}

	// Undiscovered talks (--discover).
	if len(r.DiscoveredTalks) > 0 {
		SectionHeader(fmt.Sprintf("UNDISCOVERED TALKS  (%d found via YouTube)", len(r.DiscoveredTalks)))
		fmt.Println()
		for i, t := range r.DiscoveredTalks {
			fmt.Printf("  %s  %s\n",
				labelStyle.Render(fmt.Sprintf("[%d]", i+1)),
				bodyStyle.Render(t.Title),
			)
			fmt.Printf("    %s\n", blueStyle.Render(t.URL))
		}
		fmt.Println()
	}
}
