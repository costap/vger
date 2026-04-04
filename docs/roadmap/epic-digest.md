# Epic: `vger digest` — Cross-Talk Playlist Synthesis

Covers improvements to how vger synthesises insights across multiple talks.

---

## Stories

### T1-3 · Playlist diff (`vger digest --diff <playlist-a> <playlist-b>`)
**Priority:** Tier 1 · Effort: M

**As a** conference follower,  
**I want to** compare two KubeCon playlists (e.g. 2023 vs 2024),  
**so that** I can see which technologies rose, fell, or emerged between editions.

**Acceptance criteria:**
- `--diff <playlist-url-a> <playlist-url-b>` accepts two playlist URLs
- Output sections: Emerging (new in B), Growing (up in B), Declining (down in B), Stable
- Each technology entry lists supporting talk titles
- Gemini performs the comparative synthesis; raw tech lists from cache
- `--output` saves diff report as Markdown

**Implementation notes:**
- `internal/domain/` — new `DiffReport` type with `Emerging`, `Growing`, `Declining`, `Stable []TechDelta`
- `internal/adapters/gemini/synthesise.go` — new `DiffSynthesize(a, b []domain.CachedAnalysis)` method
- `internal/cli/digest.go` — new `--diff` flag path, fetch both playlists, call diff synthesis
- `internal/cli/ui/digest.go` — new `RenderDiff()` LCARS renderer

---

### · Cross-playlist digest (combine N playlists)
**Priority:** Unscheduled · Effort: S

**As a** user tracking multiple related playlists,  
**I want to** run a single digest across all of them,  
**so that** I get a unified view without running digest separately per playlist.

**Acceptance criteria:**
- `--playlist` flag is repeatable
- All videos across all playlists deduplicated and treated as one corpus
- Output credits source playlist per technology mention

---

### T2-4 · Technology trend analysis (`vger digest --trend`)
**Priority:** Tier 2 · Effort: M

**As a** technology strategist,  
**I want to** see how technology adoption has shifted over the scanned time period,  
**so that** I can identify what's rising vs declining in the ecosystem.

**Acceptance criteria:**
- Groups cached talks by quarter/year of publication
- Per-period: top 10 technologies by mention frequency
- Delta column: +N or -N vs previous period
- Gemini narrative summary of trend story
- Requires at least 2 distinct time periods in cache

**Implementation notes:**
- `internal/domain/` — new `TechTrend`, `TrendPeriod` types
- `internal/cli/digest.go` — `--trend` flag, group by `CachedAt` / publish date
- No new Gemini method needed — pass trend data as structured input to existing synthesise path

---

### · `vger digest --since <duration>` — Recent talks digest
**Priority:** Unscheduled · Effort: S

**As a** user checking in weekly,  
**I want to** digest only talks scanned in the last N days,  
**so that** I get a quick summary of recent activity.

**Acceptance criteria:**
- `--since 30d` or `--since 2024-01-01` date filter on `CachedAt`
- Falls back to full cache if fewer than 3 videos match
- Works with or without `--playlist`

---

### · Email/Slack export for digest output
**Priority:** Unscheduled · Effort: M

**As a** team lead,  
**I want to** send a digest to Slack or email,  
**so that** I can share insights without requiring colleagues to run vger.

**Acceptance criteria:**
- `--notify slack` posts to `VGER_SLACK_WEBHOOK` env var
- `--notify email` sends via SMTP configured in `~/.vger/config.yaml`
- Fallback: `--output <file>` for manual sharing (already planned)

---

## Implementation Map

```
internal/domain/digest.go         — DiffReport, TechTrend, TrendPeriod types
internal/adapters/gemini/synthesise.go — DiffSynthesize(), TrendSynthesize()
internal/cli/digest.go            — --diff, --trend, --since, --cross-playlist flags
internal/cli/ui/digest.go         — RenderDiff(), RenderTrend() renderers
```
