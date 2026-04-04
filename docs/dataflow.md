# vger — Analysis Data Flow

---

## 1. Current Pipeline (MVP / Stub)

Sequential execution. The agent layer acts as a coordinator, calling each port in order and passing results downstream. No LLM reasoning drives the sequencing.

```mermaid
flowchart TD
    subgraph CLI ["CLI Layer (internal/cli)"]
        A([operator]) -->|"vger scan url"| B["scan command / scan.go"]
        B -->|"url string"| C["ScanAgent / agent/react.go"]
    end

    subgraph Agent ["Agent Layer (internal/agent)"]
        C -->|"url string"| D["FetchMetadata / domain.MetadataFetcher"]
        D -->|"domain.VideoMetadata"| E["AnalyseVideo / domain.VideoAnalyser"]
        E -->|"domain.Report"| F[return to CLI]
    end

    subgraph Adapters ["Adapter Layer (internal/adapters)"]
        D --- G[youtube.Client]
        E --- H[gemini.Client]
    end

    subgraph External ["External APIs"]
        G -->|"videos.list part=snippet,contentDetails"| I[(YouTube Data API v3)]
        I -->|"title, description, duration, channel"| G
        H -->|"YouTube URL + metadata as context"| J[(Gemini 2.5 Pro API)]
        J -->|"summary + technology list as JSON"| H
    end

    subgraph Render ["CLI Render (internal/cli/ui)"]
        F -->|"domain.Report"| K["lcars.go renderer"]
        K -->|"LCARS-styled terminal output"| L([terminal])
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

## 3. Implemented: Hybrid ReAct — Go Outer Pipeline + Gemini Tool-Calling Inner Loop

The sequential pipeline is preserved for the deterministic outer steps (metadata fetch, cache
check) that must always run in the same order. Inside `AnalyseVideo`, the Gemini adapter runs
a native function-calling loop where the model controls when to call enrichment tools.

**Why the outer pipeline stays in Go (not LLM-controlled):**
- `FetchMetadata` and `AnalyseVideo` are always required, always sequential — LLM autonomy here adds tokens with zero benefit
- The cache check happens *before* any LLM cost; this would be impossible if the LLM controlled the flow
- `AnalyseVideo` is itself the primary Gemini call — it cannot be a tool call inside another Gemini session

```mermaid
flowchart TD
    subgraph CLI ["CLI Layer (internal/cli)"]
        A([operator]) -->|"vger scan url"| B[scan command]
        B -->|cache hit?| CACHE[(~/.vger/cache)]
        CACHE -->|hit| RENDER([terminal])
        CACHE -->|miss| C["ScanAgent / agent/react.go"]
    end

    subgraph Agent ["Agent Layer — Fixed Pipeline"]
        C -->|"url"| D["FetchMetadata (YouTube)"]
        D -->|"VideoMetadata"| E["AnalyseVideo (Gemini)"]
        E -->|"Report"| F[return to CLI]
    end

    subgraph GeminiLoop ["Inside AnalyseVideo — Gemini Function-Calling Loop"]
        E --> G{"model reasoning\nover video"}
        G -->|"lookup_cncf_project(name)"| H["cncf.LookupProject()"]
        H -->|"stage, found"| G
        G -->|"validate_url(url)"| I["cncf.ValidateURL() → HEAD"]
        I -->|"reachable: bool"| G
        G -->|"no more tool calls"| J["final JSON → domain.Report"]
    end

    subgraph External ["External APIs"]
        D <-->|"videos.list"| K[(YouTube Data API v3)]
        G <-->|"YouTube URL passthrough"| L[(Gemini 2.5 Flash API)]
        H <-->|"landscape.yml cache"| M[(CNCF GitHub / local cache)]
    end
```

---

## 4. Forward Look: Fully LLM-Controlled ReAct (Future `vger research`)

For a future `vger research <topic>` command, the model would genuinely need to decide
sequencing: which videos to pick, whether to search the web for context, when it has
enough evidence to synthesise. This is the right use case for full LLM autonomy over
the pipeline — unlike single-video scan where the steps are always the same.

```mermaid
flowchart TD
    subgraph CLI ["CLI Layer (internal/cli)"]
        A([operator]) -->|"vger research topic"| B[research command]
        B -->|"topic string"| C["ReAct Agent (future)"]
    end

    subgraph AgentLoop ["Agent Loop — fully LLM-controlled"]
        C --> D{"model reasoning"}
        D -->|"tool: search_playlist(query)"| E["PlaylistSearch tool"]
        E -->|"video list"| D
        D -->|"tool: get_cached_analysis(id)"| F["Cache tool"]
        F -->|"CachedAnalysis or miss"| D
        D -->|"tool: search_web(query)"| G["Web Search tool (future)"]
        G -->|"snippets"| D
        D -->|"tool: lookup_cncf_project(name)"| H["CNCF tool"]
        H -->|"stage"| D
        D -->|"final answer"| I["synthesised research report"]
    end
```

**Key tools for this future pattern:**
- `search_playlist(query)` — find relevant videos without pre-selecting them
- `get_cached_analysis(video_id)` — read already-scanned results without re-uploading
- `search_web(query)` — fill gaps for technologies not in training data
- `lookup_cncf_project(name)` — already implemented ✅

---

## 5. Video Analysis Detail (Gemini Adapter)

The Gemini adapter does not download the video. The YouTube URL is passed directly to the Gemini 2.5 Flash multimodal API. The model reads audio, on-screen text (slides, code), and speaker names natively.

```mermaid
sequenceDiagram
    participant Agent as ScanAgent
    participant Adapter as gemini.Client
    participant API as Gemini 2.5 Flash API

    Agent->>Adapter: AnalyseVideo(ctx, url, metadata)
    Adapter->>API: POST /v1/models/gemini-2.5-flash:generateContent<br/>{parts: [{video_url: url}, {text: system_prompt}, {text: metadata}], tools: [...]}
    Note over API: model fetches video natively<br/>processes audio + visual frames<br/>performs speaker diarisation
    API-->>Adapter: FunctionCall{lookup_cncf_project, "Cilium"}
    Adapter->>Adapter: cncf.LookupProject("Cilium") → "graduated"
    Adapter->>API: FunctionResponse{lookup_cncf_project, {stage: "graduated"}}
    API-->>Adapter: FunctionCall{validate_url, "https://cilium.io"}
    Adapter->>Adapter: HEAD https://cilium.io → 200 OK
    Adapter->>API: FunctionResponse{validate_url, {reachable: true}}
    API-->>Adapter: JSON response {summary, technologies[]}
    Adapter->>Adapter: unmarshal JSON → *domain.Report
    Adapter-->>Agent: *domain.Report, nil
```

