# V'Ger — Prioritised Roadmap

> Version reference: v0.7  
> Structure: each epic has its own file with stories. This file is the master priority index.

---

## Epic Index

| Epic file | Scope |
|-----------|-------|
| [epic-scan.md](epic-scan.md) | `vger scan` — video ingestion and batch processing |
| [epic-ask.md](epic-ask.md) | `vger ask` — follow-up Q&A on cached videos |
| [epic-list.md](epic-list.md) | `vger list` — channel/playlist browsing |
| [epic-digest.md](epic-digest.md) | `vger digest` — cross-talk playlist synthesis |
| [epic-research.md](epic-research.md) | `vger research` — transversal topic research |
| [epic-track.md](epic-track.md) | `vger track` — technology signal management |
| [epic-new-commands.md](epic-new-commands.md) | New commands: watch, compare, export, serve, config, report |

---

## Capability Map (v0.7 baseline)

```
INPUT           ANALYSE          STORE            QUERY            OUTPUT
────────────    ────────────     ────────────     ────────────     ────────────
YouTube URL  →  scan           → cache          → ask           → terminal
Playlist     →  scan batch     → signals        → research      → markdown
Channel      →  list           → CNCF index     → digest        → (file)
Prompt       →  track add      →                → track digest  →
```

---

## Tier 1 — High Value · Low/Medium Effort

> Target: next quarter

| ID | Story | Epic | Effort |
|----|-------|------|--------|
| T1-1 | Cache hit indicator (★) in `vger list` | [list](epic-list.md) | S |
| T1-2 | `vger ask --all` — question across full cache | [ask](epic-ask.md) | M |
| T1-3 | `vger digest --diff <playlist-a> <playlist-b>` | [digest](epic-digest.md) | M |
| T1-4 | `vger watch --playlist <id>` — auto-scan daemon | [new commands](epic-new-commands.md) | M |
| T1-5 | `vger research` Phase 2 — LLM-directed deep-dives ✅ | [research](epic-research.md) | M |
| T1-6 | `vger track search <query>` as first-class command | [track](epic-track.md) | S |

---

## Tier 2 — High Value · Medium Effort

> Target: next half-year

| ID | Story | Epic | Effort |
|----|-------|------|--------|
| T2-1 | Custom lenses via `~/.vger/lenses.yaml` | [ask](epic-ask.md) | M |
| T2-2 | Web search integration (Tavily/Brave) for `vger research` | [research](epic-research.md) | M |
| T2-3 | `vger research --create-signal` — auto-capture signals | [research](epic-research.md) | S |
| T2-4 | Technology trend analysis (`vger digest --trend`) | [digest](epic-digest.md) | M |
| T2-5 | `vger scan --from-file <urls.txt>` — batch from file | [scan](epic-scan.md) | S |
| T2-6 | `vger export --format obsidian` — knowledge base export | [new commands](epic-new-commands.md) | M |
| T2-7 | `vger compare <topic-a> <topic-b>` — side-by-side analysis | [new commands](epic-new-commands.md) | M |

---

## Tier 3 — Medium Value / High Effort

> Future consideration

| ID | Story | Epic | Effort |
|----|-------|------|--------|
| T3-1 | Multi-turn interactive ask session (`--interactive`) | [ask](epic-ask.md) | L |
| T3-2 | `vger serve` — local HTTP API | [new commands](epic-new-commands.md) | L |
| T3-3 | `vger watch` with Slack/email/webhook notifications | [new commands](epic-new-commands.md) | L |
| T3-4 | Non-YouTube inputs — local files, Vimeo (yt-dlp) | [scan](epic-scan.md) | L |
| T3-5 | Decision records auto-generated from adopted/rejected signals | [track](epic-track.md) | M |
| T3-6 | Team signal store — S3/GCS remote backend | [track](epic-track.md) | L |
| T3-7 | `vger report` — styled HTML/PDF knowledge briefing | [new commands](epic-new-commands.md) | L |

---

## Themes

| Theme | Tier 1 | Tier 2 |
|-------|--------|--------|
| Better discovery | T1-1 cache hit in list, T1-4 watch | T2-5 from-file, T2-4 trend |
| Richer Q&A | T1-2 ask --all | T2-1 custom lenses, T2-7 compare |
| Research depth | T1-5 research phase 2 | T2-2 web search, T2-3 create-signal |
| Knowledge output | T1-6 track search | T2-6 export |
| Automation | T1-4 watch | — |
