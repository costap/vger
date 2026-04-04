# Epic: `vger ask` — Follow-up Q&A on Cached Videos

Covers improvements to how users query analysed content — single video, multi-video, and interactive.

---

## Stories

### T1-2 · `vger ask --all` — Question across the full cache
**Priority:** Tier 1 · Effort: M

**As a** user with a large cache of scanned talks,  
**I want to** ask a single question and get an answer drawing from all cached analyses,  
**so that** I can query my entire knowledge base at once.

**Acceptance criteria:**
- `--all` flag selects all cached videos as context (no `--video` required)
- Optional `--query <term>` pre-filters to relevance-scored cache hits before sending to Gemini
- Gemini answer cites which videos informed which parts of the response
- Context window managed by batching: summaries first, full reports if within token budget

**Implementation notes:**
- `internal/cli/ask.go` — new `--all` mode, call `AskAcrossCache()`
- `internal/adapters/gemini/qa.go` — new `AskAcrossCache(reports []domain.CachedAnalysis, question string)` method
- `internal/adapters/cache/search.go` — reuse `Search()` for relevance pre-filter

---

### T2-1 · Custom lenses via `~/.vger/lenses.yaml`
**Priority:** Tier 2 · Effort: M

**As a** user with a specific analytical context,  
**I want to** define my own `--lens` presets in a config file,  
**so that** I can apply tailored prompts without retyping them.

**Acceptance criteria:**
- `~/.vger/lenses.yaml` loaded at startup, merged with built-in lenses
- Schema: `name`, `role`, `description`, `default_questions []string`
- Custom lenses appear in `--lens` tab-completion alongside built-ins
- Name collision with a built-in lens: user's definition wins with a warning

**Implementation notes:**
- New `internal/cli/lenses.go` — config file loading and merge
- `internal/cli/ask.go` — pass merged lens map to prompt builder
- Fish completion script regenerated to pick up dynamic lenses (or use `__complete`)

---

### · Compose multiple lenses (`--lens architect,radar`)
**Priority:** Unscheduled · Effort: S

**As a** user,  
**I want to** combine two lens presets in one ask,  
**so that** I get both an architectural analysis and a tech-radar verdict in one response.

**Acceptance criteria:**
- `--lens` accepts comma-separated list
- Role contexts and default questions merged in order
- Output sections clearly labelled per lens

---

### · `--output <file>` — Save Q&A to Markdown
**Priority:** Unscheduled · Effort: S

**As a** user,  
**I want to** save the ask response to a Markdown file,  
**so that** I can include it in notes or share it with colleagues.

**Acceptance criteria:**
- `--output <path>` writes formatted Markdown (video title, question, response, timestamp)
- If file exists, appends with separator rather than overwriting

---

### · Context injection — team/stack context
**Priority:** Unscheduled · Effort: M

**As a** user with a specific tech stack,  
**I want** vger to always be aware of my context (e.g. "we run on AWS EKS with Istio"),  
**so that** ask and research answers are tailored to my environment.

**Acceptance criteria:**
- `~/.vger/config.yaml` supports `user_context: |` multiline string
- Context prepended to all Gemini prompts (ask, research, digest)
- `vger config set user_context "..."` command to set it

---

### T3-1 · Multi-turn interactive ask session (`--interactive`)
**Priority:** Tier 3 · Effort: L

**As a** user exploring a talk,  
**I want** an interactive chat loop against a cached video,  
**so that** I can ask follow-up questions without re-running the command.

**Acceptance criteria:**
- `vger ask --interactive --video <url>` opens a REPL-style prompt
- Turn history maintained in-memory for the session duration
- `/exit`, `/save`, `/clear` slash-commands in the loop
- Multi-video mode: `--all --interactive` for full-cache chat

**Implementation notes:**
- In-memory turn history `[]gemini.Content` passed back each call
- Bubble Tea or simple `bufio.Scanner` loop for input

---

## Implementation Map

```
internal/cli/ask.go           — --all flag, --output, --interactive mode
internal/cli/lenses.go        — custom lens loading from ~/.vger/lenses.yaml
internal/adapters/gemini/qa.go — AskAcrossCache() method
```
