# vger — Analysis Data Flow

---

## 1. Current Pipeline (MVP / Stub)

Sequential execution. The agent layer acts as a coordinator, calling each port in order and passing results downstream. No LLM reasoning drives the sequencing.

```mermaid
flowchart TD
    subgraph CLI ["CLI Layer (internal/cli)"]
        A([operator]) -->|"vger scan <url>"| B[scan command\nscan.go]
        B -->|"url string"| C[ScanAgent\nagent/react.go]
    end

    subgraph Agent ["Agent Layer (internal/agent)"]
        C -->|"url string"| D[FetchMetadata\ndomain.MetadataFetcher]
        D -->|"*domain.VideoMetadata"| E[AnalyseVideo\ndomain.VideoAnalyser]
        E -->|"*domain.Report"| F[return to CLI]
    end

    subgraph Adapters ["Adapter Layer (internal/adapters)"]
        D --- G[youtube.Client]
        E --- H[gemini.Client]
    end

    subgraph External ["External APIs"]
        G -->|"videos.list\npart=snippet,contentDetails"| I[(YouTube Data API v3)]
        I -->|"title, description,\nduration, channel"| G
        H -->|"YouTube URL +\nmetadata as context"| J[(Gemini 2.5 Pro API)]
        J -->|"summary + technology list\nas JSON"| H
    end

    subgraph Render ["CLI Render (internal/cli/ui)"]
        F -->|"*domain.Report"| K[lcars.go\nrenderer]
        K -->|"LCARS-styled\nterminal output"| L([terminal])
    end
```

---

## 2. Data Types Crossing Layer Boundaries

| Boundary | Type | Direction |
|----------|------|-----------|
| CLI → Agent | `string` (url) | in |
| Agent → MetadataFetcher | `string` (url) | in |
| MetadataFetcher → Agent | `*domain.VideoMetadata` | out |
| Agent → VideoAnalyser | `string` (url), `*domain.VideoMetadata` | in |
| VideoAnalyser → Agent | `*domain.Report` | out |
| Agent → CLI | `*domain.Report` | out |
| CLI → renderer | `*domain.Report` | in |

---

## 3. Planned ReAct Agent Loop

The sequential pipeline in internal/agent/react.go is replaced by a LangChainGo ReAct agent. The agent receives the video URL as its initial input and decides which tools to call, in what order, based on model reasoning. This enables multi-step enrichment (e.g., resolving a technology name against the CNCF Landscape) before the final report is assembled.

```mermaid
flowchart TD
    subgraph CLI ["CLI Layer (internal/cli)"]
        A([operator]) -->|"vger scan <url>"| B[scan command]
        B -->|"url string"| C[ReAct Agent\nLangChainGo]
    end

    subgraph AgentLoop ["Agent Loop (internal/agent)"]
        C --> D{model\nreasoning}
        D -->|"tool call:\nfetch_video_metadata"| E[MetadataFetcher\ntool wrapper]
        E -->|"*VideoMetadata as JSON\nobservation"| D
        D -->|"tool call:\nanalyse_video"| F[VideoAnalyser\ntool wrapper]
        F -->|"*Report as JSON\nobservation"| D
        D -->|"tool call:\nsearch_cncf_landscape\n(future)"| G[CNCF Landscape\ntool wrapper]
        G -->|"project record\nor empty"| D
        D -->|"final answer:\nno more tool calls"| H[assembled\n*domain.Report]
    end

    subgraph Adapters ["Adapter Layer (internal/adapters)"]
        E --- I[youtube.Client]
        F --- J[gemini.Client]
        G --- K[cncf.Client\n(future)]
    end

    subgraph External ["External APIs"]
        I <-->|"videos.list"| L[(YouTube Data API v3)]
        J <-->|"URL passthrough"| M[(Gemini 2.5 Pro)]
        K <-->|"landscape query"| N[(CNCF Landscape API)]
    end

    subgraph Render ["CLI Render (internal/cli/ui)"]
        H --> O[lcars.go renderer]
        O --> P([terminal])
    end
```

---

## 4. Video Analysis Detail (Gemini Adapter)

The Gemini adapter does not download the video. The YouTube URL is passed directly to the Gemini 2.5 Pro multimodal API. The model reads audio, on-screen text (slides, code), and speaker names natively.

```mermaid
sequenceDiagram
    participant Agent as ScanAgent
    participant Adapter as gemini.Client
    participant API as Gemini 2.5 Pro API

    Agent->>Adapter: AnalyseVideo(ctx, url, metadata)
    Adapter->>API: POST /v1/models/gemini-2.5-pro:generateContent<br/>{parts: [{video_url: url}, {text: system_prompt}]}
    Note over API: model fetches video natively<br/>processes audio + visual frames<br/>performs speaker diarisation
    API-->>Adapter: JSON response {summary, technologies[]}
    Adapter->>Adapter: unmarshal JSON → *domain.Report
    Adapter-->>Agent: *domain.Report, nil
```
