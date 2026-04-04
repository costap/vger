# Epic: New Commands

New top-level commands that expand vger beyond single-video analysis into monitoring, comparison, export, and integration.

---

## Stories

### T1-4 · `vger watch --playlist <id>` — Auto-scan daemon
**Priority:** Tier 1 · Effort: M

**As a** user following a conference playlist,  
**I want** vger to automatically scan new videos as they appear,  
**so that** my cache stays current without manual intervention.

**Acceptance criteria:**
- `vger watch --playlist <url>` polls the playlist every N minutes (default: 60)
- New videos not in cache are automatically scanned
- Terminal output: "▶ New video found: <title> — scanning…"
- `--interval <minutes>` configures poll frequency
- `--notify slack` posts scan summary to Slack webhook
- State persisted to `~/.vger/watches.json` so watches survive restarts
- `vger watch list` shows active watches; `vger watch remove <id>` cancels one
- Graceful shutdown on SIGINT/SIGTERM

**Implementation notes:**
- New `internal/cli/watch.go` command
- `internal/adapters/watch/store.go` — watches.json persistence
- Polling loop calls existing `list` + `scan` adapter chain
- Runs as a long-lived process (not a true daemon); users expected to use `nohup` or a service manager

---

### T2-6 · `vger export --format obsidian` — Knowledge base export
**Priority:** Tier 2 · Effort: M

**As a** user with an Obsidian or Notion vault,  
**I want to** export my full vger knowledge base (cache + signals),  
**so that** I can browse and link it alongside my other notes.

**Acceptance criteria:**
- `vger export --format obsidian --output <dir>` creates a vault structure:
  ```
  <dir>/
    talks/           — one note per cached video
    signals/         — one note per tracked signal
    index.md         — master index linking all notes
    technologies.md  — all detected technologies with talk backlinks
  ```
- Technology names as `[[WikiLinks]]` for Obsidian graph
- `--format logseq` outputs Logseq block format
- `--format json` outputs a single structured JSON dump
- `--since <date>` limits export to recently scanned content

**Implementation notes:**
- New `internal/cli/export.go` command
- New `internal/adapters/export/` package — `obsidian.go`, `logseq.go`, `json.go`
- Reads from both cache and signal store

---

### T2-7 · `vger compare <topic-a> <topic-b>` — Side-by-side topic analysis
**Priority:** Tier 2 · Effort: M

**As a** architect deciding between two technologies,  
**I want** vger to compare them head-to-head using all available evidence,  
**so that** I get a structured trade-off analysis without manual research.

**Acceptance criteria:**
- `vger compare "eBPF" "WASM"` runs two research passes then asks Gemini to compare
- Output sections: Use-cases, Maturity, Ecosystem, Community, Trade-offs, Recommendation
- Adopt/Trial/Hold verdict per technology
- `--lens <name>` applies a lens to focus the comparison (e.g. `architect`, `radar`)
- `--output <file>` saves to Markdown

**Implementation notes:**
- New `internal/cli/compare.go` command
- Reuses `research` pipeline for both topics
- New `CompareSynthesize(a, b ResearchReport)` in `gemini/research.go`

---

### T3-2 · `vger serve` — Local HTTP API
**Priority:** Tier 3 · Effort: L

**As a** developer,  
**I want** vger exposed as a local HTTP API,  
**so that** I can integrate it with editors, scripts, and other tools without the CLI.

**Acceptance criteria:**
- `vger serve --port 8080` starts a local HTTP server
- Endpoints:
  - `POST /scan` — scan a video URL, returns `Report`
  - `GET  /cache` — list cached analyses
  - `GET  /cache/:id` — fetch one cached analysis
  - `POST /ask` — ask a question against a cached video
  - `POST /research` — run research pipeline
  - `GET  /signals` — list tracked signals
- JSON request/response
- API key auth via `X-Vger-Token` header (configured in `~/.vger/config.yaml`)

**Implementation notes:**
- New `internal/cli/serve.go` command
- New `internal/adapters/api/` package — handlers wrapping existing domain logic
- Uses `net/http` stdlib; no framework dependency

---

### T3-3 · `vger watch` with notifications
**Priority:** Tier 3 · Effort: L  
*(Extension of T1-4)*

**As a** user who runs vger watch in the background,  
**I want** notifications when new videos are scanned or new technologies detected,  
**so that** I don't have to poll the terminal output.

**Acceptance criteria:**
- `--notify slack` posts to `VGER_SLACK_WEBHOOK` env var
- `--notify email` sends via SMTP (configured in `~/.vger/config.yaml`)
- `--notify desktop` uses macOS/Linux notification APIs
- Notification content: video title, top 3 technologies, link

---

### · `vger config` — Manage defaults
**Priority:** Unscheduled · Effort: S

**As a** user,  
**I want** a config command to set defaults without editing YAML files,  
**so that** I can configure vger quickly from the CLI.

**Acceptance criteria:**
- `vger config set default-channel @cncf`
- `vger config set default-model gemini-1.5-pro`
- `vger config set user-context "We run on AWS EKS with Istio and Cilium"`
- `vger config get <key>` — read a value
- `vger config list` — show all configured values
- Writes to `~/.vger/config.yaml`

---

### T3-7 · `vger report <playlist>` — Styled HTML/PDF knowledge briefing
**Priority:** Tier 3 · Effort: L

**As a** team lead,  
**I want** a polished HTML or PDF report from a playlist digest,  
**so that** I can share insights with colleagues who don't use the CLI.

**Acceptance criteria:**
- `vger report --playlist <url> --format html --output report.html`
- Sections: Executive summary, Technology table, Talk summaries, Recommended reading
- Styled with CSS (dark theme matching LCARS aesthetic)
- `--format pdf` via headless Chrome (`chromium --headless --print-to-pdf`)
- Embeds video thumbnails via YouTube oEmbed

---

## Implementation Map

```
internal/cli/watch.go           — polling daemon, watches.json persistence
internal/cli/export.go          — knowledge base export command
internal/cli/compare.go         — side-by-side topic comparison
internal/cli/serve.go           — HTTP API server
internal/cli/config.go          — config management command
internal/adapters/export/       — obsidian.go, logseq.go, json.go
internal/adapters/api/          — HTTP handlers
internal/adapters/watch/        — watch state persistence
internal/adapters/gemini/research.go — CompareSynthesize()
```
