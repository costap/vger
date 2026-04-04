# Epic: `vger track` — Technology Signal Management

Covers improvements to the technology signal backlog: search, cross-referencing, export, and team collaboration.

---

## Stories

### T1-6 · `vger track search <query>` — First-class search command
**Priority:** Tier 1 · Effort: S

**As a** user with a growing signal backlog,  
**I want to** search signals by keyword from the CLI,  
**so that** I can quickly find relevant signals without browsing the full list.

**Acceptance criteria:**
- `vger track search <query>` searches Title, Note, Category, and enrichment fields
- Results displayed as a compact LCARS table (title, category, status, match excerpt)
- `--status <status>` filter: narrow to Evaluating / Adopted / etc.
- `--category <cat>` filter: narrow by technology category
- Case-insensitive; partial matches supported

**Implementation notes:**
- `Search()` already implemented on both `JSONStore` and `MarkdownStore` (from research Phase 1)
- `internal/cli/track/` — new `search.go` subcommand, wire existing `Search()` to CLI
- `internal/cli/ui/track.go` — new `RenderSearchResults()` renderer

---

### · Notify when new scans match tracked signals
**Priority:** Tier 2 · Effort: M

**As a** user tracking specific technologies,  
**I want** vger to alert me when a new scan mentions a tracked signal's technology,  
**so that** my signal backlog stays fresh with new evidence.

**Acceptance criteria:**
- After `vger scan` completes, cross-reference detected technologies with active signals
- If match found: "📡 1 tracked signal matched: [Cilium — Evaluating]"
- `--no-notify` flag to suppress match output
- Matched signal IDs appended to the scan result cache entry

---

### · `track export --format obsidian`
**Priority:** Tier 2 · Effort: M

**As a** user with an Obsidian vault,  
**I want to** export my signal backlog as Obsidian notes,  
**so that** I can link and search them alongside my other knowledge.

**Acceptance criteria:**
- One `.md` file per signal in `<output-dir>/signals/`
- Frontmatter: `id`, `status`, `category`, `tags`, `created`
- Body: `## Summary`, `## Notes`, `## Enrichment`, `## Related Links`
- Technology mentions as `[[WikiLinks]]` to enable Obsidian graph
- Index note `signals-index.md` with table of all signals
- `--format logseq` outputs Logseq-compatible block format

**Implementation notes:**
- New `internal/adapters/export/` package — `obsidian.go`, `logseq.go`
- `internal/cli/track/export.go` — new subcommand

---

### · `track import` — Import from CSV or bookmark export
**Priority:** Unscheduled · Effort: M

**As a** user migrating from a spreadsheet or bookmarks,  
**I want to** bulk-import technology signals,  
**so that** I don't have to recreate them manually.

**Acceptance criteria:**
- `vger track import --format csv <file.csv>` reads headers: `title,category,status,note,url`
- `--format bookmarks <file.html>` parses Netscape bookmark format
- Duplicate detection by title (case-insensitive): skip or update with `--upsert`
- Dry-run mode: `--dry-run` shows what would be created without writing

---

### T3-5 · Decision records — Auto-generate ADR from signal
**Priority:** Tier 3 · Effort: M

**As a** team lead,  
**I want** vger to generate an Architecture Decision Record when I adopt or reject a signal,  
**so that** the decision is documented with full context.

**Acceptance criteria:**
- `vger track status <id> adopted` (or `rejected`) triggers ADR generation prompt
- ADR template filled from: signal title, notes, enrichment, linked talks
- Saved as `<output-dir>/adr/<id>-<title-slug>.md` in standard ADR format
- Optional: push to GitHub via `gh` CLI

---

### T3-6 · Team signal store — S3/GCS remote backend
**Priority:** Tier 3 · Effort: L

**As a** team,  
**I want** a shared signal store backed by cloud storage,  
**so that** all team members see the same signal backlog without manual sync.

**Acceptance criteria:**
- New `SignalStore` implementation: `S3Store` and `GCSStore`
- Configured via `~/.vger/config.yaml`: `signal_store: s3://bucket/path`
- Optimistic locking: conflict detection on concurrent writes
- `vger track sync` to pull latest from remote
- Auth: AWS default credentials chain / GCP ADC

**Implementation notes:**
- New `internal/adapters/signals/s3store.go` and `gcsstore.go`
- `domain.SignalStore` port unchanged — new implementations only

---

### · `track compare <id1> <id2>` — Side-by-side signal comparison
**Priority:** Unscheduled · Effort: S

**As a** user evaluating two competing technologies,  
**I want to** view two signals side-by-side,  
**so that** I can compare their notes, enrichments, and linked talks at a glance.

**Acceptance criteria:**
- `vger track compare <id1> <id2>` renders a two-column LCARS view
- Sections: Status, Category, Summary, Notes, Enrichment, Links
- Differences highlighted

---

### · Roadmap / Gantt view
**Priority:** Unscheduled · Effort: L

**As a** team planning adoption timelines,  
**I want** signals displayed on a timeline by status change date,  
**so that** I can see the adoption journey at a glance.

**Acceptance criteria:**
- `vger track timeline` renders signals on a horizontal Gantt in the terminal (Bubble Tea)
- X-axis: months; Y-axis: signals grouped by category
- Status transitions shown as colour bands

---

## Implementation Map

```
internal/cli/track/search.go    — new track search subcommand
internal/cli/track/export.go    — new track export subcommand
internal/cli/track/import.go    — new track import subcommand
internal/adapters/export/       — obsidian.go, logseq.go
internal/adapters/signals/s3store.go — team remote store
internal/adapters/signals/gcsstore.go
internal/cli/ui/track.go        — RenderSearchResults(), RenderCompare()
```
