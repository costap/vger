# Epic: `vger list` — Channel & Playlist Browsing

Covers improvements to how users discover, filter, and act on video listings.

---

## Stories

### T1-1 · Cache hit indicator (★) in `vger list`
**Priority:** Tier 1 · Effort: S

**As a** user browsing a playlist,  
**I want to** see at a glance which videos are already scanned,  
**so that** I know what's cached and what still needs to be processed.

**Acceptance criteria:**
- Each video row shows `★` if a cached analysis exists, `·` if not
- Indicator loads from the local cache index without making network calls
- Works for both channel listings and playlist listings
- Performance: index loaded once per `list` invocation, O(1) lookup per video

**Implementation notes:**
- `internal/adapters/cache/json.go` — add `LoadIndex() map[string]bool` returning set of cached video IDs
- `internal/cli/list.go` — load index before rendering, pass to UI renderer
- `internal/cli/ui/list.go` (or equivalent) — render `★` / `·` in the ID column

---

### · Interactive picker — select videos to scan from list
**Priority:** Unscheduled · Effort: M

**As a** user browsing a playlist,  
**I want to** select multiple videos with arrow keys and scan them in one step,  
**so that** I don't have to copy-paste URLs manually.

**Acceptance criteria:**
- `vger list --pick` enters an interactive multi-select mode (Bubble Tea)
- Space to toggle selection, Enter to confirm, Esc to abort
- Selected videos immediately piped to scan logic
- Already-cached videos shown as pre-selected (greyed out, skippable)

---

### · Filter by duration (`--min-min` / `--max-min`)
**Priority:** Unscheduled · Effort: S

**As a** user with limited time,  
**I want to** filter the list to videos within a duration range,  
**so that** I only see talks that fit my schedule.

**Acceptance criteria:**
- `--min-min 20 --max-min 45` shows videos between 20 and 45 minutes long
- Duration sourced from YouTube metadata (already fetched)
- Out-of-range videos omitted from output

---

### · Multi-channel search
**Priority:** Unscheduled · Effort: M

**As a** user tracking multiple conference channels,  
**I want to** list videos from several channels at once,  
**so that** I can compare and prioritise across sources.

**Acceptance criteria:**
- `--channel` flag is repeatable: `vger list --channel @cncf --channel @linuxfoundation`
- Results merged and deduplicated by video ID
- Source channel shown in output row
- Sort order: by publish date descending across channels

---

### · Export listing (`--format json` / `--format csv`)
**Priority:** Unscheduled · Effort: S

**As a** user scripting vger into a pipeline,  
**I want** structured output from `vger list`,  
**so that** I can pipe it to other tools.

**Acceptance criteria:**
- `--format json` outputs a JSON array of video objects
- `--format csv` outputs headers + rows
- Default format remains the current LCARS table

---

## Implementation Map

```
internal/adapters/cache/json.go  — LoadIndex() map[string]bool
internal/cli/list.go             — cache index load, duration filter, multi-channel
internal/cli/ui/list.go          — ★/· indicator rendering
```
