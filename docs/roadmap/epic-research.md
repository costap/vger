# Epic: `vger research` — Transversal Topic Research

Covers improvements to the research command, which synthesises insights across all local sources and optionally the web.

> **Current state (v0.7):** Phase 1 implemented — structured pipeline with one Gemini synthesis call.  
> Cache search → CNCF lookup → signal search → (optional) YouTube discover → Gemini synthesis.

---

## Stories

### T1-5 · Phase 2 — LLM-directed deep-dives with tools
**Priority:** Tier 1 · Effort: M

**As a** user researching a topic,  
**I want** Gemini to autonomously decide which videos to read in full and which CNCF projects to look up,  
**so that** the research report is richer and less dependent on pre-filtered inputs.

**Acceptance criteria:**
- Gemini can call `ask_video(video_id, question)` to query a specific cached video
- Gemini can call `lookup_cncf(project_name)` to fetch detailed project info
- Gemini can call `search_cache(query)` to surface additional relevant videos
- Research loop runs up to a configurable max iterations (default: 5)
- Final synthesis uses same `ResearchReport` domain type — fully backward compatible
- `--max-depth <n>` flag controls iteration budget

**Design reference:** See Phase 2 section in `docs/roadmap/` plan notes and `plan.md`.

**Implementation notes:**
- `internal/adapters/gemini/research.go` — add tool declarations (`ask_video`, `lookup_cncf`, `search_cache`) and multi-turn loop before final synthesis call
- All domain types, CLI flags, UI rendering unchanged
- `gemini.NewWithTools()` already used; tools just need registering for the research call

---

### T2-2 · Web search integration (Tavily / Brave Search)
**Priority:** Tier 2 · Effort: M

**As a** user researching a topic with few cached talks,  
**I want** vger to search the web for additional context,  
**so that** the research report covers current blog posts, docs, and news beyond my local cache.

**Acceptance criteria:**
- New `domain.WebSearcher` port with `Search(ctx, query, maxResults) ([]WebResult, error)`
- Default adapter: Tavily API (`TAVILY_API_KEY` env var); fallback: Brave Search
- Web results included as a fourth evidence section in the Gemini synthesis prompt
- `--no-web` flag to disable web search even if API key is present
- Web results cited in `ResearchReport.InvestPaths` with source URLs

**Implementation notes:**
- New `internal/adapters/websearch/` package (tavily.go, brave.go)
- `internal/domain/ports.go` — `WebSearcher` interface
- `internal/adapters/gemini/research.go` — add web results section to `buildResearchPrompt()`

---

### T2-3 · `--create-signal` — Auto-capture signals from research
**Priority:** Tier 2 · Effort: S

**As a** user who ran a research pass and found interesting technologies,  
**I want to** create track signals for the top findings without switching commands,  
**so that** the research-to-tracking loop is seamless.

**Acceptance criteria:**
- `--create-signal` flag triggers interactive confirmation after synthesis
- Shows top N `InvestPaths` with suggested signal category
- User selects which to capture (multi-select prompt)
- Created signals linked to research report (source URL or title)
- Signals written to the active signal store

**Implementation notes:**
- `internal/cli/research.go` — post-synthesis prompt block
- Calls existing `track add` domain logic directly (no subprocess)

---

### · `--since <date>` — Scope research to recent talks
**Priority:** Unscheduled · Effort: S

**As a** user doing a periodic review,  
**I want to** research a topic using only talks scanned in the last 90 days,  
**so that** I get current rather than historical signals.

**Acceptance criteria:**
- `--since 90d` or `--since 2024-06-01` date filter on cache search results
- Surfaced videos older than threshold excluded from evidence
- CNCF and signal search unaffected (not date-scoped)

---

### · Research history — save and diff runs over time
**Priority:** Unscheduled · Effort: M

**As a** user running research on the same topic periodically,  
**I want** previous research runs saved and diff'd against the current one,  
**so that** I can see what changed in my knowledge base since last time.

**Acceptance criteria:**
- Research reports saved to `~/.vger/research/<topic>/<timestamp>.json`
- `--diff-last` flag compares current run against most recent saved run
- Diff highlights: new evidence, new CNCF projects, changed verdicts, new invest paths
- `vger research --history <topic>` lists past runs

---

### · Multi-topic comparison (`vger research "eBPF" vs "WASM"`)
**Priority:** Tier 2 · Effort: M  
*(See also [epic-new-commands.md](epic-new-commands.md) for `vger compare`)*

**As a** architect evaluating two technologies,  
**I want** a head-to-head research comparison,  
**so that** I can make an informed adopt/trial/hold decision.

**Acceptance criteria:**
- `vger research --compare "eBPF" "WASM"` runs two research passes
- Gemini synthesises a comparative brief: use-cases, maturity, ecosystem overlap
- Output section: "Recommendation" with explicit trade-off analysis

---

### · Research watch — notify on new signals/videos
**Priority:** Tier 3 · Effort: L

**As a** user tracking a topic long-term,  
**I want** vger to notify me when new cached videos or signals appear for my research topics,  
**so that** I stay current without manually re-running research.

**Acceptance criteria:**
- `vger research --watch "eBPF"` persists topic and polls (integrates with `vger watch`)
- Notification via Slack webhook or terminal bell on new evidence
- Weekly digest mode: `--watch --notify weekly`

---

## Implementation Map

```
internal/adapters/gemini/research.go — Phase 2 tool loop, web results section
internal/adapters/websearch/         — new package: tavily.go, brave.go
internal/domain/ports.go             — WebSearcher interface
internal/domain/research.go          — ResearchHistory, DiffReport types (future)
internal/cli/research.go             — --create-signal, --since, --compare, --max-depth flags
```
