# vger — Technical Specification

module: github.com/costap/vger  
language: Go 1.22+  
version: 0.1.0-dev  

---

## Purpose

vger is a command-line tool that accepts a conference video URL (primarily YouTube), submits it to a multimodal language model for analysis, and returns a structured report containing a talk summary and a ranked list of technologies the operator should investigate further. The primary use case is KubeCon/CloudNativeCon session analysis, though the architecture is not domain-specific.

---

## Architectural Pattern: Onion / Ports and Adapters

The codebase is organised as concentric layers. Dependencies point inward; outer layers depend on inner layers, never the reverse.

Layer 1 — Domain (innermost)  
Location: internal/domain/  
Contains Go structs and interface definitions (ports). Has zero external imports outside the Go standard library. This is the stable core that all other layers depend on.

Layer 2 — Agent  
Location: internal/agent/  
Contains orchestration logic that coordinates the domain ports. Currently a sequential pipeline (fetch metadata, then analyse video). Wires domain.MetadataFetcher and domain.VideoAnalyser together. Has no knowledge of which concrete adapter is used; it operates solely on interfaces.

Layer 3 — Adapters (outermost application code)  
Location: internal/adapters/  
Contains concrete implementations of the domain ports. Each adapter subdirectory corresponds to one external service or technology. Adapters import domain types but nothing from agent or cli.

Layer 4 — CLI  
Location: internal/cli/, cmd/vger/  
Contains Cobra command definitions and terminal UI rendering. This layer instantiates concrete adapters, injects them into the agent, and invokes agent.Run(). It is the composition root.

---

## Directory Structure

```
vger/
├── cmd/vger/main.go                    entry point; calls cli.Root.Execute()
├── internal/
│   ├── domain/
│   │   ├── model.go                    VideoMetadata, Technology, Report structs
│   │   └── ports.go                    MetadataFetcher, VideoAnalyser interfaces
│   ├── agent/
│   │   └── react.go                    ScanAgent: orchestrates fetcher + analyser
│   ├── adapters/
│   │   ├── gemini/analyser.go          domain.VideoAnalyser via Gemini 2.5 Pro API
│   │   └── youtube/metadata.go         domain.MetadataFetcher via YouTube Data API v3
│   └── cli/
│       ├── root.go                     cobra root command; global flags; composition root
│       ├── scan.go                     `vger scan <url>` command handler
│       └── ui/lcars.go                 LCARS terminal renderer (lipgloss)
└── docs/
    ├── architecture.md                 architecture options document
    └── spec.md                         this file
```

---

## Domain Model

VideoMetadata  
Fields: URL string, Title string, Description string, ChannelName string, PublishedAt string, DurationSec int  
Source: populated by MetadataFetcher before the analysis call. Passed to VideoAnalyser as context.

Technology  
Fields: Name string, Description string, WhyRelevant string, LearnMore string, CNCFStage string  
CNCFStage values: "graduated", "incubating", "sandbox", or empty string if not a CNCF project.

Report  
Fields: VideoTitle string, VideoURL string, Stardate string, Summary string, Technologies []Technology  
This is the final output of a single scan run. Rendered by the CLI layer.

---

## Port Interfaces

MetadataFetcher  
Method: FetchMetadata(ctx context.Context, url string) (*VideoMetadata, error)  
Responsibility: retrieve title, description, duration, and channel information for the given video URL. Must not perform any analysis.

VideoAnalyser  
Method: AnalyseVideo(ctx context.Context, url string, metadata *VideoMetadata) (*Report, error)  
Responsibility: submit the video URL (and optional metadata as context) to a multimodal model and return a populated Report. The implementation may pass the URL directly to the model without downloading the video.

---

## Adapter Implementations

gemini.Client (implements VideoAnalyser)  
Location: internal/adapters/gemini/analyser.go  
Strategy: passes the YouTube URL directly to Gemini 2.5 Pro via the google.golang.org/genai SDK. The model receives the full video natively — no download or transcription step is performed by vger. The prompt instructs the model to return a JSON object matching the Report schema.  
Config: API key via GEMINI_API_KEY environment variable or --gemini-key flag.  
Current state: stub returning hardcoded data. Real API call to be wired in a subsequent iteration.

youtube.Client (implements MetadataFetcher)  
Location: internal/adapters/youtube/metadata.go  
Strategy: calls the YouTube Data API v3 videos.list endpoint with part=snippet,contentDetails.  
Config: API key via YOUTUBE_API_KEY environment variable or --youtube-key flag.  
Current state: stub returning hardcoded data.

---

## Agent Layer

ScanAgent (internal/agent/react.go)  
Current implementation: sequential pipeline — calls FetchMetadata then AnalyseVideo.

Planned evolution: replace the sequential pipeline with a LangChainGo ReAct agent. The agent will receive a set of tools wrapping the same port interfaces:

  tool: fetch_video_metadata  
  input: url string  
  output: VideoMetadata as JSON  

  tool: analyse_video  
  input: url string, metadata JSON  
  output: Report as JSON  

  tool: search_cncf_landscape (future)  
  input: technology name string  
  output: CNCF project record or empty  

The ReAct loop allows the model to interleave reasoning with tool calls, enabling enrichment steps (e.g., resolving a mentioned project name against the CNCF Landscape) before producing the final report.

---

## CLI Layer

Root command: vger  
Persistent flags: --gemini-key, --youtube-key  
PersistentPreRun: renders the LCARS header banner on every command invocation.

Subcommand: scan  
Usage: vger scan <youtube-url>  
Behaviour:  
1. Prints LCARS status lines at each pipeline stage.  
2. Instantiates youtube.Client and gemini.Client with keys from flags/env.  
3. Constructs agent.ScanAgent and calls Run(ctx, url).  
4. On success: renders the Report using LCARS-style section headers and field labels.  
5. On error: renders a RED ALERT banner and returns a non-zero exit code.

---

## Terminal UI

Location: internal/cli/ui/lcars.go  
Library: github.com/charmbracelet/lipgloss  

Colour palette:  
  amber   #FF9900  — labels, section headers, banner  
  blue    #99CCFF  — completion messages, stardate value  
  red     #CC4444  — error banners (RED ALERT)  
  white   #FFFFFF  — body text  
  dimGrey #666666  — decorators, timestamps, secondary labels  

Stardate calculation: TNG-style numeric stardate derived from year and day-of-year. Formula: (year - 1900) * 10 + (day_of_year / days_in_year) * 10. Printed as a single decimal, e.g. 1262.5.

Exported functions:  
  Header()                 — prints ASCII V'Ger banner and current stardate  
  Status(msg string)       — prints a dim-timestamped status line  
  Complete(msg string)     — prints a blue completion line in uppercase  
  RedAlert(err error)      — prints a full-width red error block  
  SectionHeader(title)     — prints an amber section divider  
  Field(key, value string) — prints a labelled key/value row  
  LabelStyle() Style       — returns the amber lipgloss.Style  
  DimStyle() Style         — returns the dim grey lipgloss.Style  

---

## Configuration

All secrets are passed via environment variables or CLI flags. No config file is read in the current implementation. The --config flag path is reserved for a future YAML config file.

  GEMINI_API_KEY   required for video analysis in production  
  YOUTUBE_API_KEY  required for metadata fetching in production  

In stub mode (current), both keys may be empty — the stub adapters do not validate them.

---

## Extension Points

Adding a new video source  
Implement domain.MetadataFetcher for the new source. Register the adapter in cli/root.go or cli/scan.go via a --source flag. No changes to domain or agent layers required.

Adding a new LLM provider  
Implement domain.VideoAnalyser using the provider's SDK. Inject via --provider flag in the CLI layer.

Adding CNCF enrichment  
Implement a new tool wrapping the CNCF Landscape API (landscape.cncf.io/api). Register it with the LangChainGo ReAct agent in internal/agent/react.go.

Adding batch/playlist processing  
Add a `playlist` subcommand in internal/cli/. Iterate over video URLs from the YouTube Data API playlistItems.list endpoint. Each URL is processed by a separate agent.ScanAgent call. Use goroutines and a semaphore for bounded concurrency.

---

## Dependencies

  github.com/spf13/cobra            CLI framework  
  github.com/charmbracelet/lipgloss terminal styling  
  github.com/tmc/langchaingo        LLM agent framework (planned for ReAct wiring)  
  google.golang.org/genai           Gemini SDK (planned for production adapter)  
  google.golang.org/api             YouTube Data API client (planned for production adapter)  

---

## Build and Run

  go build ./cmd/vger             build the binary  
  go run ./cmd/vger scan <url>    run without building  
  GEMINI_API_KEY=... YOUTUBE_API_KEY=... ./vger scan <url>   production invocation  
