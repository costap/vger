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
- Submit the video directly to **Gemini 2.5 Flash** for multimodal analysis
- Extract a structured summary and a technology radar with CNCF project context
- Cache all analysis locally so follow-up questions are instant and free
- Answer follow-up questions from cached context — or re-submit the video for deep analysis

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

### `vger scan` — Analyse a video

Submits a conference video to Gemini for full multimodal analysis.
Results are cached to `~/.vger/cache/` for instant follow-up queries.

```bash
vger scan <youtube-url>
```

**Example:**

```bash
vger scan https://www.youtube.com/watch?v=H06qrNmGqyE
```

**Output includes:**
- Video title, channel, duration, and stardate
- 3–5 sentence technical summary
- Technology radar: each identified project with description, CNCF stage, why it matters, and where to learn more

**Flags:**

| Flag | Description |
|------|-------------|
| `--refresh` | Force re-analysis even if a cached result exists |
| `--model` | Override the Gemini model (default: `gemini-2.5-flash`) |

---

### `vger list` — Browse a channel's videos

Lists videos from a YouTube channel ordered by publish date, newest first.
Supports searching by title and description to find specific events.

```bash
vger list --channel <channel-id-or-handle> [--search <query>] [--max <n>]
```

**Examples:**

```bash
# Browse the CNCF channel
vger list --channel @cncf

# Find all KubeCon talks
vger list --channel @cncf --search kubecon

# Narrow to KubeCon 2024, more results
vger list --channel @cncf --search "kubecon 2024" --max 50
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--channel` | required | Channel ID (`UCxx...`) or handle (`@cncf`) |
| `--search` | — | Filter by title/description keyword |
| `--max` | 25 | Maximum results (1–50) |

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

**Examples:**

```bash
# Fast — answered from cached notes
vger ask https://www.youtube.com/watch?v=H06qrNmGqyE \
  "Who were the speakers and what companies do they work for?"

vger ask https://www.youtube.com/watch?v=H06qrNmGqyE \
  "Which of the technologies mentioned are production-ready today?"

# Deep — Gemini re-reads the full video
vger ask --deep https://www.youtube.com/watch?v=H06qrNmGqyE \
  "What exact kubectl command did they run in the live demo?"
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--deep` | Re-submit the video to Gemini for video-grounded answers |

> Note: `vger scan` must be run at least once before `vger ask` can be used.
> Use `vger scan --refresh` to refresh the cache and pick up the deep notes format
> if you scanned before the notes feature was added.

---

## TYPICAL WORKFLOW

```bash
# 1. Find KubeCon talks you want to analyse
vger list --channel @cncf --search "kubecon 2024" --max 20

# 2. Scan a talk — Gemini analyses the full video
vger scan https://www.youtube.com/watch?v=TALK_ID

# 3. Ask follow-up questions from the cache (instant)
vger ask https://www.youtube.com/watch?v=TALK_ID \
  "Which technologies are worth prioritising for my platform team?"

# 4. Dig deeper into something specific (re-reads video)
vger ask --deep https://www.youtube.com/watch?v=TALK_ID \
  "What did they say about multi-cluster networking?"
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

## ARCHITECTURE

V'Ger uses an **onion (ports-and-adapters) architecture**:

```
cmd/vger/          — binary entry point
internal/
  domain/          — model structs and port interfaces (zero external deps)
  agent/           — analysis pipeline orchestration
  adapters/
    youtube/       — YouTube Data API v3 client
    gemini/        — Gemini multimodal analyser and Q&A
    cache/         — local JSON file cache (~/.vger/cache/)
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

---

*"V'Ger has been damaged... but it has not been destroyed."*
