# V·G·E·R

```
    ██╗   ██╗ ██████╗ ███████╗██████╗
    ██║   ██║██╔════╝ ██╔════╝██╔══██╗
    ██║   ██║██║  ███╗█████╗  ██████╔╝
    ╚██╗ ██╔╝██║   ██║██╔══╝  ██╔══██╗
     ╚████╔╝ ╚██████╔╝███████╗██║  ██║
      ╚═══╝   ╚═════╝ ╚══════╝╚═╝  ╚═╝
    ════════════════════════════════════
    KNOWLEDGE  ASSIMILATION  UNIT  001
```

[![Release](https://img.shields.io/github/v/release/costap/vger?style=flat-square&label=latest%20release&color=FF9900)](https://github.com/costap/vger/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/costap/vger?style=flat-square)](https://goreportcard.com/report/github.com/costap/vger)
[![License](https://img.shields.io/github/license/costap/vger?style=flat-square)](LICENSE)

> *"It has gathered so much knowledge... it wants to share it."*
> — Star Trek: The Motion Picture

**V'Ger** is a command-line AI agent that ingests online conference videos —
KubeCon, CloudNativeCon, and beyond — and produces structured intelligence
reports: summaries, technology radars, and learning recommendations.

Named after the legendary probe, V'Ger travels the information stream, assimilates
cloud-native knowledge, and returns it to you in a form you can act on.

---

## MISSION PARAMETERS

- Fetch video metadata from the YouTube Data API
- Browse channel video catalogues and playlists, search across complete upload history
- Submit individual videos or entire playlists to **Gemini 2.5 Flash** for multimodal analysis
- Extract a structured summary and a technology radar with CNCF project context
- Cache all analysis locally so follow-up questions are instant and free
- Answer follow-up questions from cached context — or re-submit the video for deep analysis
- Generate playlist digests: technology radar across all talks, AI-synthesised learning paths
- **Track technology signals** as you discover them — capture, AI-enrich, and synthesise your backlog into actionable digests; optionally backed by a git repo for version-controlled signal history

---

## INSTALLATION

### Homebrew (macOS and Linux) — Recommended

```bash
brew install costap/tap/vger
```

### GitHub Releases — Pre-built binaries

Download the latest binary for your platform from the
[Releases page](https://github.com/costap/vger/releases/latest):

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `vger_darwin_arm64.tar.gz` |
| macOS (Intel) | `vger_darwin_amd64.tar.gz` |
| Linux (x86-64) | `vger_linux_amd64.tar.gz` |
| Linux (ARM64) | `vger_linux_arm64.tar.gz` |
| Windows (x86-64) | `vger_windows_amd64.zip` |

```bash
# Example: Linux amd64
curl -L https://github.com/costap/vger/releases/latest/download/vger_linux_amd64.tar.gz \
  | tar xz
sudo mv vger /usr/local/bin/
```

### Go (requires Go toolchain)

```bash
go install github.com/costap/vger/cmd/vger@latest
```

### Build from source

```bash
git clone https://github.com/costap/vger.git
cd vger
go build -o vger ./cmd/vger
```

---

## TRANSMISSION CODES (API KEYS)

V'Ger requires two API keys to operate. Set them as environment variables,
pass them as flags, or place them in a `.env` file in your working directory.

### YouTube Data API v3

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create or select a project
3. Enable **YouTube Data API v3** under APIs & Services
4. Create an **API Key** credential

### Gemini API

1. Go to [Google AI Studio](https://aistudio.google.com/)
2. Click **Get API key** → **Create API key**

### Configure

```bash
# Option A — environment variables
export YOUTUBE_API_KEY=your_youtube_key
export GEMINI_API_KEY=your_gemini_key

# Option B — .env file (auto-loaded from working directory)
cat > .env <<EOF
YOUTUBE_API_KEY=your_youtube_key
GEMINI_API_KEY=your_gemini_key
EOF

# Option C — flags (per command)
vger --youtube-key YOUR_KEY --gemini-key YOUR_KEY scan <url>
```

---

## COMMAND REFERENCE

### `vger scan` — Analyse a video or playlist

Submits a conference video to Gemini for full multimodal analysis.
Results are cached to `~/.vger/cache/` for instant follow-up queries.

```bash
# Single video
vger scan <youtube-url>

# Entire playlist (parallel, resumable)
vger scan --playlist <playlist-id-or-url>
```

**Examples:**

```bash
# Scan a single talk
vger scan https://www.youtube.com/watch?v=H06qrNmGqyE

# Scan all videos in a playlist (3 parallel by default)
vger scan --playlist "https://www.youtube.com/playlist?list=PLj6h78yzYM2P..."

# Faster on a paid Gemini tier
vger scan --playlist PLj6h78yzYM2P... --concurrency 5

# Re-scan everything, ignoring the cache
vger scan --playlist PLj6h78yzYM2P... --refresh
```

**Output includes:**
- Video title, channel, duration, and stardate
- 3–5 sentence technical summary
- Technology radar: each identified project with description, CNCF stage, why it matters, and where to learn more

Playlist scans are **resumable** — already-cached videos are skipped automatically.
Re-run after an interruption and V'Ger picks up where it left off.

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--playlist` | — | Playlist ID or URL — scan all videos in the playlist |
| `--concurrency` | `3` | Number of parallel analyses in playlist mode |
| `--refresh` | `false` | Force re-analysis even if a cached result exists |
| `--model` | `gemini-2.5-flash` | Override the Gemini model |

---

### `vger list` — Browse a channel's videos or playlists

Lists videos or playlists from a YouTube channel ordered by publish date, newest first.
Results are retrieved by walking the channel's **complete upload history** — not limited
by YouTube's search index, so all videos are found regardless of age.

Cached videos are marked with ★ and show the technology tags extracted by Gemini.
Use `--cached` to search everything you have ever scanned without any YouTube API call.

```bash
# List videos from a channel
vger list --channel <channel-id-or-handle>

# List a channel's playlists
vger list --channel <channel-id-or-handle> --playlists

# List videos inside a specific playlist
vger list --playlist <playlist-id-or-url>

# Browse all locally cached videos (no API call)
vger list --cached
```

**Examples:**

```bash
# Browse the CNCF channel
vger list --channel @cncf

# Find all ArgoCD talks (scans full upload history)
vger list --channel @cncf --search argocon

# List all playlists on the CNCF channel
vger list --channel @cncf --playlists

# Filter playlists by name
vger list --channel @cncf --playlists --search kubecon

# List all videos in a specific playlist
vger list --playlist "https://www.youtube.com/playlist?list=PLj6h78yzYM2P..."

# Filter playlist videos by keyword
vger list --playlist PLj6h78yzYM2P... --search "service mesh"

# Filter by Gemini-extracted technology tag (only cached videos match)
vger list --channel @cncf --tags ebpf

# Combine tag and keyword filters
vger list --channel @cncf --tags kubernetes --search 2024

# Search all cached videos for a technology — no channel needed
vger list --cached --tags ebpf

# Search all cached KubeCon talks (matches playlist title)
vger list --cached --tags kubecon

# eBPF talks from KubeCon specifically
vger list --cached --tags kubecon --search ebpf
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--channel` | — | Channel ID (`UCxx...`) or handle (`@cncf`) |
| `--playlist` | — | Playlist ID or URL — list videos from a specific playlist |
| `--playlists` | `false` | List playlists instead of videos |
| `--search` | — | Filter by title/description keyword |
| `--tags` | — | Filter by Gemini technology tag or playlist name (substring match) |
| `--cached` | `false` | Browse all locally cached videos without any YouTube API call |
| `--max` | `50` | Maximum number of results |

> **Tag filtering** matches against both technology names extracted by Gemini (e.g. `--tags cilium`)
> and playlist titles set when scanning (e.g. `--tags kubecon`, `--tags "eu 2025"`).
> Only cached videos can match tag filters — uncached videos show a dim `·` indicator.

---

### `vger ask` — Follow-up questions

Ask a question about a previously scanned video without re-uploading it.

```bash
vger ask <youtube-url> "<question>"
```

V'Ger stores detailed notes during every scan — covering everything mentioned
in the video, not just the top technologies. Most questions can be answered
from this cache instantly and at no extra cost.

For questions that require direct access to the video (exact quotes, timestamps,
demo details not captured in the notes), use `--deep`:

```bash
vger ask --deep <youtube-url> "<question>"
```

#### Analytical lenses

Use `--lens` to apply a built-in analytical preset instead of typing the same
verbose prompt every time. The question argument becomes **optional** when a lens
is set — the lens provides a default question. You can still add a custom question
to focus the lens on something specific.

```bash
# Use a lens (no question needed)
vger ask --lens architect https://www.youtube.com/watch?v=H06qrNmGqyE

# Lens + custom focus
vger ask --lens architect https://www.youtube.com/watch?v=H06qrNmGqyE \
  "focus specifically on the database connection pooling approach"

# Lens + deep video re-read
vger ask --deep --lens radar https://www.youtube.com/watch?v=H06qrNmGqyE
```

**Available lenses:**

| Lens | What it produces |
|------|-----------------|
| `architect` | Solutions architect analysis: approach, decisions, trade-offs, novelty |
| `engineer` | Hands-on deep-dive: implementation patterns, config details, getting started, gotchas |
| `radar` | Tech radar recommendations: Adopt / Trial / Assess / Hold per technology |
| `brief` | 3–5 bullet team briefing: problem, approach, takeaways, action |

**Examples:**

```bash
# Fast — answered from cached notes
vger ask https://www.youtube.com/watch?v=H06qrNmGqyE \
  "Who were the speakers and what companies do they work for?"

vger ask https://www.youtube.com/watch?v=H06qrNmGqyE \
  "Which of the technologies mentioned are production-ready today?"

# Analytical lenses — no question needed
vger ask --lens architect https://www.youtube.com/watch?v=H06qrNmGqyE
vger ask --lens radar     https://www.youtube.com/watch?v=H06qrNmGqyE
vger ask --lens brief     https://www.youtube.com/watch?v=H06qrNmGqyE

# Deep — Gemini re-reads the full video
vger ask --deep https://www.youtube.com/watch?v=H06qrNmGqyE \
  "What exact kubectl command did they run in the live demo?"
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--deep` | Re-submit the video to Gemini for video-grounded answers |
| `--lens <name>` | Apply a built-in analytical preset (`architect`, `engineer`, `radar`, `brief`) |

> Note: `vger scan` must be run at least once before `vger ask` can be used.
> Use `vger scan --refresh` to refresh the cache and pick up the deep notes format
> if you scanned before the notes feature was added.

---

### `vger research` — Topic research brief

Search all available sources about a technology topic and receive a structured
synthesis with a CNCF landscape map, evidence from cached videos, tracked signals,
investigation paths, and a bottom-line verdict.

```bash
vger research <topic> [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--discover` | `false` | Search YouTube for unscanned relevant talks |
| `--channel` | `@cncf` | YouTube channel to search when `--discover` is used |
| `--lens` | — | Apply analytical lens (`architect`, `engineer`, `radar`, `brief`) |
| `--max-videos` | `10` | Max cached videos to include in context |
| `--output` | — | Write full report to a Markdown file |

**Sources searched (always):**
- Local analysis cache — scored full-text search across summaries, notes, technology names
- CNCF landscape — related projects by name and category
- Track signals — matching signals from your backlog

**Source searched with `--discover`:**
- YouTube channel — unscanned talks on the topic (deduplicated against cache)

**Output sections:**
- **Brief** — 2–3 sentence what-and-why
- **CNCF Landscape** — related projects with stage and relevance
- **Evidence from Cache** — cached talks that mention the topic
- **Tracked Signals** — signals from your backlog matching the topic
- **Investigation Paths** — 2–4 distinct routes to explore further
- **Competing Approaches** — alternative technologies
- **Verdict** — bottom-line recommendation
- **Undiscovered Talks** — unscanned talks found via `--discover`

**Examples:**

```bash
# Basic research from local knowledge base
vger research "eBPF"

# Include YouTube discovery for unscanned talks
vger research "multi-cluster networking" --discover

# Apply an architectural lens to the synthesis
vger research "WASM in Kubernetes" --lens architect

# Save full report as Markdown
vger research "service mesh" --output service-mesh-brief.md

# Discover from a specific channel
vger research "eBPF" --discover --channel @isovalent
```

> **Tip:** Run `vger scan --playlist <id>` first on relevant playlists to build up
> your local knowledge base. The more you've scanned, the richer the evidence section.

---

### `vger digest` — Playlist overview and learning path

After scanning a playlist, produce a cross-talk overview without re-running any analysis.

**Layer 1 (always, zero API cost):** reads all cached analyses and renders:
- Compact talk table with duration and top technologies per talk
- Technology radar — bar chart of which technologies appeared across how many talks

**Layer 2 (`--ai`, one Gemini text call):** AI synthesis across all talks:
- Overarching theme of the playlist
- Recommended learning path (technologies in study order)
- 3–5 priority talks to watch first with reasons
- Key insights narrative

```bash
# Local overview — instant, no API call
vger digest --playlist <playlist-id-or-url>

# + AI synthesis
vger digest --playlist <playlist-id-or-url> --ai

# Export a shareable Markdown report
vger digest --playlist <playlist-id-or-url> --ai --output kubecon2024.md
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--playlist` | Playlist ID or URL (required) |
| `--ai` | Use Gemini to synthesise a cross-playlist learning path |
| `--output` | Write a Markdown report to a file |

> `vger scan --playlist` must be run first to populate the cache.

---

### `vger track` — Technology signal tracking

Capture, manage, and review technologies and ideas you encounter day-to-day.
Signals accumulate in a personal backlog; Gemini enriches them on demand and
synthesises your backlog into an actionable digest.

**Storage adapts automatically:**

| Condition | Store |
|-----------|-------|
| `TECHDR_DIR` env set, or `~/code/github.com/costap/tech-signals` exists | Markdown files + git auto-commit (compatible with [tech-signals](https://github.com/costap/tech-signals)) |
| Default | JSON files at `~/.vger/signals/` |

#### `track add` — Capture a new signal

```bash
# Interactive — prompts for title, URL, source, category, note
vger track add

# AI-assisted — describe it in natural language; Gemini extracts the fields
vger track add --ai "saw a KubeCon talk on HolmesGPT for AI-driven k8s remediation https://..."

# AI-assisted + open $EDITOR for review (default when TECHDR_DIR is set)
vger track add --ai "..." --edit
```

| Flag | Description |
|------|-------------|
| `--ai <text>` | Free-text description; Gemini extracts all signal fields |
| `--edit` | Open `$VISUAL`/`$EDITOR` after capture for manual review |

#### `track list` — Browse your backlog

```bash
vger track list
vger track list --status spotted
vger track list --status evaluating --category security
```

| Flag | Description |
|------|-------------|
| `--status` | Filter by status: `spotted` \| `evaluating` \| `adopted` \| `rejected` \| `parked` |
| `--category` | Filter by category: `platform` \| `networking` \| `security` \| `observability` \| `developer-experience` \| `ai-ml` \| `data` \| `other` |

#### `track show` — View a signal in detail

```bash
vger track show 0001
```

Displays all fields including any AI enrichment (what it is, maturity, alternatives, stack fit, next steps).

#### `track enrich` — AI-enrich a signal

```bash
vger track enrich 0001
```

Calls Gemini to fill in: **What it is**, **Maturity & Risk**, **Alternatives**, **How it could fit your stack**, **Suggested next steps**.
Enrichment is stored alongside the signal and shown by `track show`.

#### `track status` — Update investigation status

```bash
vger track status 0001 evaluating
vger track status 0001 adopted
```

Valid progression: `spotted` → `evaluating` → `adopted` | `rejected` | `parked`

When using the Markdown store, produces a `status: 0001 old → new` git commit.

#### `track link` — Link a signal to a conference talk scan

```bash
vger track link 0001 --video https://www.youtube.com/watch?v=abc123
```

Records the video ID on the signal so you can cross-reference `vger scan` results.

| Flag | Description |
|------|-------------|
| `--video <url>` | YouTube URL or video ID to associate with this signal |

#### `track digest` — AI synthesis of your backlog

Synthesises your signal backlog into a structured report: focus areas, technology clusters, and a pulse reading on your radar.

```bash
# Full backlog digest
vger track digest

# Digest only spotted signals, enrich any that haven't been enriched yet
vger track digest --status spotted --enrich

# Export a Markdown report
vger track digest --output ~/tech-review-2026-04.md

# Filter by category
vger track digest --category security
```

| Flag | Description |
|------|-------------|
| `--status` | Only digest signals with this status |
| `--category` | Only digest signals in this category |
| `--enrich` | AI-enrich unenriched signals before synthesising |
| `--output` | Write Markdown report to this file path |

> Requires `GEMINI_API_KEY`. Uses [Genkit](https://github.com/firebase/genkit) for typed structured output.

---

## TYPICAL WORKFLOW

### Single video

```bash
# 1. Find talks you want to analyse
vger list --channel @cncf --search "kubecon 2024" --max 20

# 2. Scan a talk — Gemini analyses the full video
vger scan https://www.youtube.com/watch?v=TALK_ID

# 3. Ask follow-up questions from the cache (instant)
vger ask https://www.youtube.com/watch?v=TALK_ID \
  "Which technologies are worth prioritising for my platform team?"

# 4. Use an analytical lens — no question needed
vger ask --lens architect https://www.youtube.com/watch?v=TALK_ID
vger ask --lens radar     https://www.youtube.com/watch?v=TALK_ID

# 5. Dig deeper into something specific (re-reads video)
vger ask --deep https://www.youtube.com/watch?v=TALK_ID \
  "What did they say about multi-cluster networking?"
```

### Full playlist

```bash
# 1. Find the playlist you want
vger list --channel @cncf --playlists --search kubecon

# 2. Scan all videos in the playlist (resumable, cached per video)
vger scan --playlist PLj6h78yzYM2P... --concurrency 3

# 3. Get an instant overview — technology radar across all talks
vger digest --playlist PLj6h78yzYM2P...

# 4. Add AI synthesis — learning path and priority talks
vger digest --playlist PLj6h78yzYM2P... --ai

# 5. Export a shareable Markdown report
vger digest --playlist PLj6h78yzYM2P... --ai --output kubecon2024.md

# 6. Drill into a specific talk
vger ask https://www.youtube.com/watch?v=TALK_ID \
  "What was shown in the live demo?"
```

---

### Cross-playlist / cross-event search

Scan multiple editions of a conference and search across all of them at once.
V'Ger records the playlist title as an event tag on every video it scans.

```bash
# Scan multiple KubeCon editions
vger scan --playlist PLj6h78yzYM2N...  # KubeCon EU 2025
vger scan --playlist PLj6h78yzYM2P...  # KubeCon NA 2024

# Browse everything you have scanned (no YouTube API call)
vger list --cached

# All KubeCon talks (matches playlist title)
vger list --cached --tags kubecon

# eBPF talks from any scanned playlist
vger list --cached --tags ebpf

# eBPF talks specifically from KubeCon
vger list --cached --tags kubecon --search ebpf

# Find talks on Cilium across any event
vger list --cached --tags cilium
```

---

### Tech signal tracking + research

Connect your scanned talks with your signal backlog and research topics across
everything V'Ger knows.

```bash
# After scanning a playlist, research a specific topic
vger research "eBPF"

# Discover unscanned talks too
vger research "eBPF" --discover

# Apply a lens for a focused perspective
vger research "service mesh" --lens architect

# Save as Markdown for sharing or archiving
vger research "WASM in Kubernetes" --output wasm-brief.md
```

---

### Tech signal tracking

Capture technologies as you encounter them — blog posts, talks, colleague recommendations.
Enrich them later and synthesise your backlog when it's time to prioritise.

```bash
# 1. Quick capture from a tweet or article (AI extracts the fields)
vger track add --ai "read a blog post about HolmesGPT for AI-driven k8s remediation https://..."

# 2. Browse your backlog
vger track list --status spotted

# 3. Review a signal in detail
vger track show 0001

# 4. AI-enrich it with context, alternatives, and stack fit
vger track enrich 0001

# 5. Link it to a conference talk you scanned
vger track link 0001 --video https://www.youtube.com/watch?v=abc123

# 6. Update investigation status as you make progress
vger track status 0001 evaluating

# 7. Synthesise your backlog into an actionable digest
vger track digest --status spotted --enrich --output ~/tech-review.md
```

**With tech-signals git repo** (set `TECHDR_DIR` or place repo at `~/code/github.com/costap/tech-signals`):
every add, enrich, status change, and link is automatically git-committed with a structured message —
giving you a full audit trail of your technology evaluation history.

```bash
export TECHDR_DIR=~/code/github.com/costap/tech-signals
vger track add --ai "eBPF replacing sidecars in Cilium 1.15..."
# → creates signals/2026/0002-2026-04-04-ebpf-sidecarless.md
# → git commit: "signal: 0002 eBPF replacing sidecars in Cilium 1.15"
```

---

## CACHE

Analysis results are stored in `~/.vger/cache/<video-id>.json`.

Each cache entry includes:
- Video metadata (title, channel, duration, publish date)
- Structured report (summary, technologies)
- Detailed notes (exhaustive narrative for follow-up Q&A)

Cache entries do not expire — conference talks don't change.
Use `vger scan --refresh <url>` to force a re-analysis.

---

## GLOBAL FLAGS

These flags apply to all commands:

| Flag | Env var | Default | Description |
|------|---------|---------|-------------|
| `--gemini-key` | `GEMINI_API_KEY` | — | Gemini API key |
| `--youtube-key` | `YOUTUBE_API_KEY` | — | YouTube Data API key |
| `--model` | — | `gemini-2.5-flash` | Gemini model to use |

---

## SHELL COMPLETION

V'GER supports tab-completion for all commands and flags in Fish, Bash, Zsh, and PowerShell.

**Fish (recommended)**

```fish
vger completion fish > ~/.config/fish/completions/vger.fish
```

Fish loads completions from that directory automatically — no `config.fish` edit needed. Completions are active immediately (or after `source ~/.config/fish/completions/vger.fish`).

**Bash**

```bash
vger completion bash > /etc/bash_completion.d/vger
# or for the current session only:
source <(vger completion bash)
```

**Zsh**

```zsh
vger completion zsh > "${fpath[1]}/_vger"
```

**PowerShell**

```powershell
vger completion powershell | Out-String | Invoke-Expression
```

**What gets completed:**
- All subcommands (`scan`, `ask`, `digest`, `completion`, …)
- All flag names for each command
- `--lens` flag values: `architect`, `engineer`, `radar`, `brief` (with descriptions)

---

## ARCHITECTURE

V'Ger uses an **onion (ports-and-adapters) architecture**:

```
cmd/vger/          — binary entry point
internal/
  domain/          — model structs and port interfaces (zero external deps)
  agent/           — analysis pipeline orchestration
  adapters/
    youtube/       — YouTube Data API v3 client
    gemini/        — Gemini multimodal analyser, Q&A, and signal AI (enrich/add)
    genkit/        — Genkit flows for typed structured output (track digest)
    cache/         — local JSON file cache (~/.vger/cache/)
    signals/       — signal store adapters: JSONStore + MarkdownStore (tech-signals)
  cli/             — Cobra commands with LCARS terminal styling
docs/              — architecture options, dataflow diagrams, technical spec
```

Video analysis is performed via **Gemini's native YouTube URL passthrough** —
no video is downloaded. Gemini processes audio, slides, on-screen code, and
speaker names directly from the stream.

---

## RELEASING A NEW VERSION

Releases are fully automated via [GoReleaser](https://goreleaser.com/) and GitHub Actions.

```bash
git tag v0.1.0
git push --tags
```

This triggers `.github/workflows/release.yml`, which:
1. Cross-compiles binaries for all 6 platforms
2. Creates a GitHub Release with all assets and checksums
3. Pushes an updated Homebrew formula to `costap/homebrew-tap`

### One-time setup for maintainers

**Create the Homebrew tap repo:**

```bash
gh repo create costap/homebrew-tap --public --description "Homebrew tap for costap tools"
```

**Add the required GitHub secret:**

Go to `github.com/costap/vger → Settings → Secrets → Actions` and add:

| Secret | Value |
|--------|-------|
| `HOMEBREW_TAP_GITHUB_TOKEN` | A GitHub PAT with `repo` scope on `costap/homebrew-tap` |

The built-in `GITHUB_TOKEN` is used automatically for creating the release itself — no extra setup needed.

---

## STARDATE LOG

| Stardate | Event |
|----------|-------|
| 1262.4 | V'Ger bootstrap — project initialised |
| 1262.5 | YouTube Data API and Gemini multimodal integration |
| 1262.6 | Channel listing, analysis caching, follow-up Q&A |
| 1262.7 | `vger track` — technology signal tracking with AI enrichment, Genkit digest, and Markdown/git store integration |

---

*"V'Ger has been damaged... but it has not been destroyed."*
