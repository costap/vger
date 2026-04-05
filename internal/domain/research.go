package domain

// ResearchReport is the output of a vger research run.
// It aggregates evidence from all local sources and a Gemini synthesis.
type ResearchReport struct {
	Topic               string
	Brief               string           // 2–3 sentence what-and-why
	LandscapeMap        []RelatedProject // related CNCF and detected projects
	EvidenceVideos      []EvidenceEntry  // cached videos that mention the topic
	RelatedSignals      []SignalSummary  // tracked signals matching the topic
	DiscoveredTalks     []VideoListing   // unscanned talks found via YouTube (--discover)
	InvestigationPaths  []InvestPath     // ordered options to explore further
	CompetingApproaches []string         // alternative technologies or approaches
	Verdict             string           // bottom-line recommendation
}

// RelatedProject is a technology project relevant to the research topic,
// drawn from the CNCF landscape.
type RelatedProject struct {
	Name      string
	CNCFStage string // graduated | incubating | sandbox | ""
	Category  string
	Homepage  string
	Relevance string // why relevant to this topic (LLM-generated)
}

// EvidenceEntry is a reference to a cached video that mentions the topic.
type EvidenceEntry struct {
	VideoTitle string
	VideoURL   string
	Speakers   []string
	Relevance  string // what the video said about the topic (LLM-generated)
}

// SignalSummary is a compact view of a tracked Signal, used inside ResearchReport.
type SignalSummary struct {
	ID       string
	Title    string
	Status   string
	Category string
	Note     string
}

// InvestPath is a suggested investigation route returned as part of a ResearchReport.
type InvestPath struct {
	Title       string
	Description string
	Actions     []string
}
