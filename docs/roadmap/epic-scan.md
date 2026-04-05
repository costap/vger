# Epic: `vger scan` — Video Ingestion & Batch Processing

Covers all improvements to how vger ingests, processes, and pre-processes video content.

---

## Stories

### T2-5 · `vger scan --from-file <urls.txt>` — Batch from file
**Priority:** Tier 2 · Effort: S

**As a** researcher with a curated list of talks,  
**I want to** point vger at a plain-text file of YouTube URLs,  
**so that** I can batch-scan them without running the command repeatedly.

**Acceptance criteria:**
- `--from-file <path>` reads one URL per line, blank lines and `#` comments ignored
- Each URL processed in sequence with per-video progress output
- Errors on individual URLs logged and skipped; remaining URLs still processed
- Already-cached videos skipped unless `--force` is passed

**Implementation notes:**
- `internal/cli/scan.go` — read URL list, iterate existing `runSingleScan()` logic
- No new domain types needed

---

### T1-4 (partial) · `vger scan --watch` — Monitor and auto-scan new videos
**Priority:** Tier 1 · Effort: M  
*(Full watch daemon is in [epic-new-commands.md](epic-new-commands.md); this story covers the scan-side integration)*

**As a** user with a polling watcher running,  
**I want** newly discovered videos to be scanned automatically,  
**so that** my cache stays up to date without manual intervention.

**Acceptance criteria:**
- `vger watch` can call scan logic for each new video ID
- Scan result is persisted to cache as normal
- Scan errors are logged but do not abort the watch loop

---

### ✅ Speaker/presenter detection
**Priority:** ~~Unscheduled~~ · **Status: DONE** · Effort: S

**As a** user reading a scan report,  
**I want** speaker names extracted from video metadata,  
**so that** I can filter and search by presenter.

**Acceptance criteria:**
- ✅ Speaker name(s) added to `domain.Report` type
- ✅ Gemini prompt updated to extract speaker from title/description (`"Name (Affiliation)"` format, deduplication)
- ✅ Displayed in `vger scan` report, `vger list` chips, and `vger research` evidence output
- ✅ `--tags` filtering matches speaker names
- ✅ Cache search scores speaker names

---

### · Auto-assign signal category from scan output
**Priority:** Unscheduled · Effort: S

**As a** user running a scan,  
**I want** vger to suggest a `track add` signal for detected technologies,  
**so that** I don't have to manually create signals after scanning.

**Acceptance criteria:**
- Post-scan prompt: "3 new technologies found. Add as signals? [y/N]"
- Creates signals with category inferred from `Report.Technologies`
- Signal status defaults to `Evaluating`

---

### T3-4 · Non-YouTube inputs (local files, Vimeo)
**Priority:** Tier 3 · Effort: L

**As a** user with conference recordings not on YouTube,  
**I want** to scan local MP4/WebM files or Vimeo URLs,  
**so that** vger covers my full video corpus.

**Acceptance criteria:**
- Local files: upload via Gemini Files API before analysis
- Vimeo: new `MetadataFetcher` implementation using Vimeo oEmbed API
- `yt-dlp` used as optional pre-processor for unsupported platforms
- Cache key derived from content hash or URL rather than YouTube ID

**Implementation notes:**
- New `internal/adapters/vimeo/` package for metadata
- New `internal/adapters/localmedia/` package for file upload
- `domain.MetadataFetcher` port already exists — add new implementations

---

### · Confidence scores per technology
**Priority:** Unscheduled · Effort: M

**As a** user reviewing technology mentions,  
**I want** each detected technology to carry a confidence score,  
**so that** I can distinguish certain mentions from speculative ones.

**Acceptance criteria:**
- `domain.Technology` gains `Confidence float64` field (0–1)
- Gemini prompt instructs model to rate confidence per technology
- Displayed in scan output and surfaced in research/ask

---

## Implementation Map

```
internal/cli/scan.go          — --from-file flag, runSingleScan loop
internal/domain/report.go     — Speaker, Confidence fields
internal/adapters/vimeo/      — new MetadataFetcher impl
internal/adapters/localmedia/ — Gemini Files API upload
```
