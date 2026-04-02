# V'Ger — Architecture Options

> *"It has gathered so much knowledge... it wants to share it."*  
> — Star Trek: The Motion Picture

V'Ger is a Go-based AI agent that ingests online conference videos (KubeCon, CloudNativeCon, etc.) and produces structured summaries with technology recommendations.

---

## Core Pipeline

Every architecture must solve these stages:

```
Video Source → Transcription/Understanding → Analysis → Structured Output
```

---

## Stage 1: Video Ingestion

| Option | Pros | Cons |
|--------|------|------|
| **YouTube Data API v3** | Official, rich metadata (chapters, description, auto-captions) | Quota limits, no raw video download |
| **yt-dlp** (subprocess from Go) | Downloads any public video, audio-only option to save bandwidth | External binary dependency, ToS grey area |
| **Direct Gemini YouTube URL** | Zero download — pass URL directly to Gemini API | Gemini-only, no control over chunking |
| **Manual file upload** | Full control over source material | Poor UX for automation |

**Recommendation:** Use YouTube Data API for metadata and playlist scanning, and pass the YouTube URL directly to Gemini for analysis. Fall back to yt-dlp for non-YouTube sources.

---

## Stage 2: Video Understanding

This is the most critical architectural choice.

### Option A — Gemini Native Video ⭐ Recommended

Gemini 2.5 Pro accepts YouTube URLs or video files directly. It processes up to **6 hours of video natively** and simultaneously understands audio, slides, on-screen code, and speaker names.

```
YouTube URL ──► Gemini 2.5 Pro ──► Structured analysis
```

**Pros:**
- Single API call, no intermediate steps
- Reads on-screen text (demo code, architecture diagrams, slide titles)
- Speaker diarisation built-in
- Visual context preserved (logos, project names shown on screen)

**Cons:**
- Google/Gemini lock-in
- Cost scales with video length at high volume

---

### Option B — Transcribe-then-Analyze

Download audio → transcribe with Whisper → feed transcript to any LLM.

```
yt-dlp (audio) ──► OpenAI Whisper / whisper.cpp ──► GPT-4o / Claude ──► Analysis
```

**Pros:**
- Provider-agnostic, portable
- Transcript is reusable — store once, query many times
- Cheaper at scale once transcripts are cached

**Cons:**
- Loses all visual context (slides, code shown on screen)
- Two separate API calls with error surface between them
- Whisper accuracy degrades with technical jargon

---

### Option C — Hybrid (Transcribe + Key Frames)

Extract audio transcript with Whisper AND sample video frames with ffmpeg, then feed both to a vision-capable LLM.

```
yt-dlp ──► Whisper transcript
        └─► ffmpeg key frames ──► Claude / GPT-4V ──► Analysis
```

**Pros:**
- Provider-agnostic
- Captures visual context from slides and demos

**Cons:**
- Most complex orchestration
- High token cost (images are expensive)
- Frame sampling strategy is non-trivial to get right

---

## Stage 3: Agent Architecture Pattern

### Pattern 1 — Simple Pipeline

No agent loop — just a sequential chain of Go functions. Best for an MVP or CLI tool.

```go
metadata := FetchYouTubeMetadata(url)
analysis := GeminiAnalyzeVideo(url)
report   := FormatReport(analysis)
```

**Best for:** Getting something working fast, single-purpose tool.

---

### Pattern 2 — Tool-use ReAct Agent ⭐ Recommended

The LLM drives the flow by deciding which tools to call. This allows multi-step reasoning and enrichment.

```
Tools available to the agent:
  - fetch_video_metadata     → YouTube title, description, chapters
  - analyze_video            → Gemini video understanding
  - search_cncf_landscape    → Validate/enrich tech mentions against CNCF projects
  - generate_summary         → Structured summary from transcript
  - extract_technologies     → Named tech list with novelty scoring
```

**Best for:** Richer output, extensibility over time. The agent can look up whether a mentioned project is CNCF-incubating vs graduated, cross-reference speaker bios, etc.

---

### Pattern 3 — Multi-Agent

Separate specialised agents coordinated by an orchestrator. Enables parallelism for batch processing.

```
Orchestrator
├── Ingestion Agent   → fetches metadata and video
├── Analysis Agent    → transcription + summarisation
└── Radar Agent       → technology extraction + novelty scoring
```

**Best for:** Processing entire KubeCon playlists in parallel, production scale.

---

## Stage 4: Go Framework Options

| Framework | Best fit | Notes |
|-----------|----------|-------|
| **LangChainGo** | Pattern 2 (tool-use agent) | Most mature Go LLM library; 10+ LLM providers; solid tool/agent abstraction |
| **Google Genkit** | Pattern 1 or 2 with Gemini | First-class Gemini support; simpler if staying Google-only |
| **Eino** (ByteDance) | Pattern 3 at scale | Production-oriented; built for ReAct-style agents |
| **Raw Go + official SDKs** | Any pattern | Maximum control; use `google.golang.org/genai` directly |

For V'Ger, **LangChainGo + Gemini** or **Genkit** are the pragmatic starting points.

---

## Recommended Starting Architecture

```
┌──────────────────────────────────────────────────────────┐
│                   vger CLI                               │
│                vger scan <youtube-url>                   │
├──────────────────────────────────────────────────────────┤
│  ReAct Agent (LangChainGo)                               │
│  ┌─────────────────┐ ┌─────────────┐ ┌───────────────┐  │
│  │ YouTube Metadata│ │Gemini Video │ │ CNCF Landscape│  │
│  │ Tool            │ │ Tool        │ │ Tool          │  │
│  └─────────────────┘ └─────────────┘ └───────────────┘  │
├──────────────────────────────────────────────────────────┤
│  Structured Output                                       │
│  {                                                       │
│    "title": "...",                                       │
│    "summary": "...",                                     │
│    "technologies": [                                     │
│      { "name": "Cilium", "why_relevant": "...",          │
│        "cncf_stage": "graduated", "learn_more": "..." }  │
│    ]                                                     │
│  }                                                       │
└──────────────────────────────────────────────────────────┘
```

**Initial tech stack:**
- **Language:** Go 1.22+
- **Agent framework:** LangChainGo or Google Genkit
- **LLM / video understanding:** Gemini 2.5 Pro (YouTube URL passthrough)
- **Video metadata:** YouTube Data API v3
- **Tech enrichment:** CNCF Landscape API or local snapshot
- **Output:** JSON → rendered as Markdown report

---

## Key Decisions to Revisit Before Building

| Decision | Options |
|----------|---------|
| Gemini-only vs provider-agnostic? | Gemini wins on video quality; multi-provider adds complexity early |
| Single video vs playlist batch? | KubeCon has 300+ talks — design for batch from the start |
| Persist transcripts? | Strongly recommended — cache to avoid re-processing cost |
| Output format? | Start with CLI Markdown; add JSON flag for piping |
| CNCF enrichment? | Nice-to-have V2 feature; skip for MVP |
