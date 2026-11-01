# AGENTS.md — AI Agent Guide for V'Ger

Welcome, agent. This document provides the essential architecture, conventions, and operational workflows for working with and modifying the **V'Ger** codebase.

---

## 1. Project Overview

**V'Ger** (`github.com/costap/vger`) is a Go-based CLI AI agent and knowledge assimilation system. It ingests online technical conference videos (primarily KubeCon, CloudNativeCon, and related cloud-native events from YouTube) and synthesises structured intelligence:
- Video analysis summaries, speaker lists, and technology radars with CNCF project stage context.
- Local caching of analyses to enable fast, zero-LLM-cost follow-up Q&A and full-text search.
- Technology signal tracking (`vger track`) to capture, enrich with AI, and synthesise architectural backlogs into actionable digests.
- Multi-source deep research synthesis (`vger research`) across local cached talks, tracked signals, the CNCF landscape, and discovered YouTube talks.
- Star Trek LCARS-styled terminal output (`charmbracelet/lipgloss`).

---

## 2. Architectural Blueprint: Onion / Hexagonal (Ports & Adapters)

V'Ger follows a strict **Hexagonal Architecture (Ports and Adapters)**. Dependencies point inward: outer layers depend on inner layers, **never** the reverse.

```
       ┌──────────────────────────────────────────────────┐
       │                   Composition                    │
       │           cmd/vger/main.go, internal/cli         │
       └──────────────┬───────────────────┬───────────────┘
                      │                   │
                      ▼                   ▼
       ┌────────────────────────┐  ┌──────────────────────┐
       │   internal/adapters    │  │    internal/agent    │
       │   gemini, youtube,     │  │    (react.go)        │
       │   signals, cache,      │  └──────────┬───────────┘
       │   cncf, genkit, quota  │             │
       └──────────────┬─────────┘             │
                      │                       │
                      └───────────┬───────────┘
                                  ▼
                     ┌────────────────────────┐
                     │    internal/domain     │
                     │  (ports.go, model.go)  │
                     │ ZERO EXTERNAL IMPORTS  │
                     └────────────────────────┘
```

### The Architectural Layers

1. **`internal/domain/` (Innermost Core)**
   - Contains pure Go domain models (`model.go`, `research.go`) and port interface definitions (`ports.go`).
   - **Critical Rule:** ZERO external library imports outside the Go standard library (e.g. `context`, `time`, `strings`). No Cobra, no Genkit, no Gemini SDK, no third-party JSON/YAML tags.
   - Defines ports: `MetadataFetcher`, `VideoAnalyser`, `ChannelLister`, `AnalysisCache`, `CacheSearcher`, `SignalStore`, `SignalEnricher`, `VideoQA`.

2. **`internal/agent/` (Orchestration Layer)**
   - Coordinates domain ports (e.g. `ScanAgent` orchestrates `MetadataFetcher` and `VideoAnalyser`).
   - Operates solely on domain interfaces; has no knowledge of concrete adapters or external APIs.
   - Keeps outer pipeline execution deterministic while allowing internal adapter logic to use dynamic LLM function calling.

3. **`internal/adapters/` (Outer Infrastructure)**
   - Concrete implementations of domain ports.
   - Each sub-package corresponds to an external integration or persistence target:
     - `adapters/gemini`: Multimodal video analysis, Q&A, signal enrichment, research synthesis, and function-calling tools using `google.golang.org/genai`.
     - `adapters/youtube`: YouTube Data API v3 client (`google.golang.org/api`) for channel, playlist, and video metadata.
     - `adapters/cache`: JSON file-backed analysis storage under `~/.vger/cache/<video_id>.json` with relevance search.
     - `adapters/signals`: JSON / Markdown signal store under `~/.vger/signals/` with search and pulse metrics.
     - `adapters/cncf`: Embedded CNCF landscape database and project classification (graduated, incubating, sandbox).
     - `adapters/genkit`: Google Genkit flows for structured playlist and signal synthesis.
     - `adapters/quota`: Request and token-per-minute (TPM) quota tracking with daily persistence.
     - `adapters/config`: User settings in `~/.vger/config.yaml`.

4. **`internal/cli/` & `cmd/vger/` (Composition Root)**
   - `cmd/vger/main.go` loads `.env` and executes `cli.Root`.
   - `internal/cli` defines Cobra commands (`scan`, `list`, `ask`, `digest`, `track`, `research`, `quota`, `config`).
   - Instantiates concrete adapters, injects them into domain ports/agents, and runs workflows.
   - Renders terminal output using `internal/cli/ui` (LCARS theme via `charmbracelet/lipgloss`).

---

## 3. Directory Layout

```
vger/
├── .github/
│   ├── renovate.json               # Renovate dependency management configuration
│   └── workflows/
│       ├── ci.yml                  # PR & push CI test and build workflow
│       └── release.yml             # GoReleaser release workflow (triggered on tags)
├── cmd/
│   └── vger/
│       └── main.go                 # Application entrypoint; loads .env, runs cli.Root
├── internal/
│   ├── domain/                     # Pure domain core (no external dependencies)
│   │   ├── model.go                # VideoMetadata, Report, Signal, Digest models
│   │   ├── ports.go                # Core interfaces (VideoAnalyser, SignalStore, etc.)
│   │   └── research.go             # ResearchReport, RelatedProject, InvestPath
│   ├── agent/                      # Pipeline orchestration
│   │   └── react.go                # ScanAgent orchestrating fetch -> analyse -> report
│   ├── adapters/                   # Concrete port implementations
│   │   ├── cache/                  # ~/.vger/cache/ JSON storage and search
│   │   ├── cncf/                   # CNCF project stage categorization
│   │   ├── config/                 # ~/.vger/config.yaml parser and persistence
│   │   ├── gemini/                 # Google GenAI client, video prompt, QA, retry
│   │   ├── genkit/                 # Firebase Genkit Go flows
│   │   ├── quota/                  # API request tracker & rate limiters
│   │   ├── signals/                # Tech signal store (JSON / Markdown)
│   │   └── youtube/                # YouTube Data API v3 metadata & listings
│   └── cli/                        # Cobra commands and UI
│       ├── root.go                 # Root command, global flags, composition root
│       ├── scan.go                 # `vger scan <url>`
│       ├── list.go                 # `vger list videos|playlists|tags|speakers`
│       ├── ask.go                  # `vger ask <video-id> "<question>"`
│       ├── digest.go               # `vger digest <playlist-id>`
│       ├── research.go             # `vger research "<topic>"`
│       ├── quota.go                # `vger quota`
│       ├── config.go               # `vger config set|show`
│       ├── channels.go             # Curated list of conference channels & tab-completion
│       ├── track/                  # `vger track` subcommand tree
│       │   ├── cmd.go              # Track root command
│       │   ├── add.go              # `track add`
│       │   ├── list.go             # `track list`
│       │   ├── show.go             # `track show`
│       │   ├── enrich.go           # `track enrich`
│       │   ├── link.go             # `track link`
│       │   ├── status.go           # `track status`
│       │   └── digest.go           # `track digest`
│       └── ui/                     # Terminal UI presentation (LCARS / lipgloss)
├── docs/                           # Architectural design specs & roadmaps
├── digests/                        # Sample generated digests
├── go.mod                          # Go module configuration (Go 1.25.0)
└── go.sum                          # Checksums
```

---

## 4. Key Technology Stack & Environment

- **Language:** Go 1.25.0
- **CLI Framework:** `github.com/spf13/cobra`
- **Terminal UI / Styling:** `github.com/charmbracelet/lipgloss`
- **AI / Multimodal SDK:** `google.golang.org/genai` (Gemini 2.5 Flash / Pro)
- **AI Framework:** `github.com/firebase/genkit/go`
- **YouTube API:** `google.golang.org/api/youtube/v3`
- **Config & Env:** `github.com/joho/godotenv`, `go.yaml.in/yaml/v3`

### Required Environment Variables
- `GEMINI_API_KEY`: API key for Google Gemini (Google AI Studio).
- `YOUTUBE_API_KEY`: API key for YouTube Data API v3 (Google Cloud Console).

Both can be placed in a `.env` file at repository root or passed via CLI flags (`--gemini-key`, `--youtube-key`).

---

## 5. Development & Testing Commands

### Build
```bash
go build -o vger ./cmd/vger
```

### Test
Run all unit tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test -v ./...
```

Run tests for a specific adapter:
```bash
go test -v ./internal/adapters/quota/...
go test -v ./internal/adapters/gemini/...
go test -v ./internal/adapters/youtube/...
```

### Code Quality & Formatting
Verify formatting and vet:
```bash
go vet ./...
go fmt ./...
```

### Dependency Management & CI
- Dependencies are monitored and updated automatically by **Renovate** via [`.github/renovate.json`](.github/renovate.json).
- Pull requests and commits to `main` trigger [`.github/workflows/ci.yml`](.github/workflows/ci.yml) (`go mod verify`, `go vet ./...`, `go test -v ./...`, `go build`).
- Validate Renovate config locally using:
```bash
npx --yes --package renovate renovate-config-validator .github/renovate.json
```

---

## 6. Development Rules and Invariants for AI Agents

When modifying or adding code to V'Ger, agents must observe the following rules:

1. **Preserve Hexagonal Boundaries:**
   - Do **NOT** import `internal/adapters`, `internal/cli`, or third-party libraries into `internal/domain`.
   - Domain models must remain standard Go types.
   - If a new external integration is needed, declare a port interface in `internal/domain/ports.go`, implement it in `internal/adapters/<new_adapter>`, and wire it in `internal/cli`.

2. **Context Propagation:**
   - Always accept `context.Context` as the first argument in any adapter, port, or long-running function.
   - Respect context cancellation and deadlines.

3. **Error Handling:**
   - Wrap errors with descriptive context using `%w`: `fmt.Errorf("fetching metadata for %s: %w", url, err)`.
   - Errors for missing resources should return `(nil, nil)` where documented (e.g. `SignalStore.Load` or `AnalysisCache.Load`).

4. **Persistence & Filesystem Safety:**
   - User data defaults to `~/.vger/` (`cache/`, `signals/`, `config.yaml`).
   - Use `0o750` for directories and `0o600` (or `0o640`) for files.
   - In unit tests, **never** write to `~/.vger/`. Always use `t.TempDir()`.

5. **Terminal UI & Styling:**
   - Output meant for user consumption in CLI commands should use the LCARS styling helpers in `internal/cli/ui` (`ui.Header()`, `ui.Card()`, `ui.Section()`, etc.).
   - Error messages go to `os.Stderr`.

6. **Quota & Rate Limiting Awareness:**
   - Video ingestion and multimodal analysis are token and quota heavy. Ensure any batching or retry logic respects `internal/adapters/quota` to avoid hitting rate limits on free or tier-1 Gemini quotas.

---

## 7. Adding Features: Quick Reference

### Adding a New Domain Port & Adapter
1. Define the interface in `internal/domain/ports.go`.
2. Add necessary models to `internal/domain/model.go`.
3. Create an adapter package under `internal/adapters/<name>/`.
4. Write unit tests in `internal/adapters/<name>/<name>_test.go` using `t.TempDir()` or test doubles.
5. Wire the adapter in the corresponding command in `internal/cli/`.

### Adding a New CLI Command
1. Create a new command file in `internal/cli/<command>.go`.
2. Define a `var <command>Cmd = &cobra.Command{ ... }`.
3. Register the command in `internal/cli/root.go` inside `init()`.
4. If it requires API keys or config, read from the shared persistent variables in `root.go` (`geminiAPIKey`, `youtubeAPIKey`, `userContext`, `quotaTier`).
