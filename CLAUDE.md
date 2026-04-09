# CLAUDE.md — OSS Bot

This file provides context for Claude Code when working on this repository.

## Project Overview

OSS Bot is the AI-powered tooling layer for the [Open School Syllabus (OSS)](https://github.com/p-n-ai/oss). It provides three interfaces to contribute structured curriculum content:

1. **GitHub Bot** (`@oss-bot`) — comment-driven content generation in the OSS repo
2. **CLI Tool** (`oss`) — local validation, generation, import, translation
3. **Web Portal** (`contribute.p-n-ai.org`) — teacher-friendly contribution form

All tools share a single AI content generation pipeline: Context Building -> AI Generation -> Content Merge -> Progress Reporting -> Validation -> Output (PR).

## Tech Stack

### Backend (Go)
- **Language:** Go >= 1.22
- **CLI Framework:** `spf13/cobra`
- **GitHub API:** stdlib `net/http` (direct REST calls — no third-party GitHub client library; consistent with existing bot code)
- **YAML:** `gopkg.in/yaml.v3`
- **JSON Schema:** `santhosh-tekuri/jsonschema/v5`
- **JWT Auth:** `golang-jwt/jwt/v5`
- **PDF Parsing (CLI):** `ledongthuc/pdf` (lightweight, standalone)
- **Document Parsing (Server):** Apache Tika via `google/go-tika` (PDF, DOCX, PPTX, XLSX, HTML — Docker sidecar)
- **Logging:** `log/slog` (stdlib)
- **HTTP:** `net/http` (stdlib, Go 1.22+ router)
- **Config:** Environment variables with `OSS_` prefix

### Web Portal (TypeScript)
- **Framework:** Next.js 15 (App Router)
- **UI:** shadcn/ui + Tailwind CSS
- **Forms:** React Hook Form + Zod
- **State:** TanStack Query v5

### AI Providers
- OpenAI (GPT-4o) — general content generation
- Anthropic (Claude) — teaching notes, pedagogy
- Ollama (Llama 3) — free, self-hosted option
- Reasoning Models (DeepSeek R1, Kimi K2.5, Qwen 3.5, o3-mini) via OpenRouter — complex analysis: bulk import, content merge, prerequisite mapping

## Project Structure

```
oss-bot/
├── cmd/
│   ├── oss/main.go                  # CLI entrypoint (cobra)
│   └── bot/main.go                  # GitHub Bot + Web Portal server
├── internal/
│   ├── ai/                          # AI provider interface (shared with P&AI Bot)
│   │   ├── provider.go              # Provider interface
│   │   ├── openai.go
│   │   ├── anthropic.go
│   │   ├── ollama.go
│   │   └── reasoning.go             # Reasoning model provider
│   ├── pipeline/                    # Shared orchestrator (ALL interfaces use this)
│   │   ├── pipeline.go              # Execute(ctx, Request) → Result
│   │   ├── bulk.go                  # Bulk import orchestrator (parallel workers)
│   │   └── progress.go              # Progress reporting (CLI bar, Bot comment, Web SSE)
│   ├── output/                      # Output writers
│   │   ├── writer.go                # Writer interface + LocalWriter (CLI)
│   │   └── github.go                # GitHubWriter (Bot + Web Portal)
│   ├── generator/                   # Content generation (individual generators)
│   │   ├── context.go               # Context builder (loads topic, schema, etc.)
│   │   ├── teaching_notes.go
│   │   ├── assessments.go
│   │   ├── examples.go
│   │   ├── translator.go
│   │   ├── scaffolder.go
│   │   ├── importer.go              # Document -> curriculum import
│   │   ├── enrich.go               # Topic YAML enrichment (Level 2 fields: teaching, engagement_hooks)
│   │   └── merge.go                 # Content merge (append + dedup assessments, additive teaching notes)
│   ├── validator/                   # Schema validation
│   │   ├── validator.go             # JSON Schema engine
│   │   ├── resolver.go              # Per-subject schema resolution (subject override + global fallback)
│   │   ├── bloom.go                 # Bloom's taxonomy checks
│   │   ├── prerequisites.go         # Prerequisite graph cycle detection
│   │   ├── duplicates.go            # Duplicate content detection
│   │   └── quality.go               # Quality level assessment
│   ├── parser/                      # Input parsing + content extraction
│   │   ├── command.go               # Parse @oss-bot commands
│   │   ├── contribution.go          # Natural language -> structured data
│   │   ├── document.go              # ContentExtractor interface + InputSource
│   │   ├── pdf.go                   # PDF text extraction (Go-native, for CLI)
│   │   ├── tika.go                  # Apache Tika client (multi-format, for server)
│   │   ├── url.go                   # URL fetcher (web page → text)
│   │   ├── image.go                 # Image extraction (OCR + AI Vision)
│   │   └── chunker.go              # Large document chunking
│   ├── github/                      # GitHub API integration
│   │   ├── app.go                   # GitHub App auth (JWT + installation tokens)
│   │   ├── webhook.go               # Webhook handler + HMAC verification
│   │   ├── pr.go                    # PR creation, labels, reviewers
│   │   └── contents.go              # Read/write via GitHub Contents API
│   └── api/                         # Web portal backend (thin layer → delegates to pipeline)
│       ├── router.go
│       ├── preview.go               # POST /api/preview → pipeline.Execute(ModePreview)
│       ├── submit.go                # POST /api/submit → pipeline.Execute(ModeCreatePR)
│       └── curricula.go             # GET /api/curricula
├── web/                             # Next.js web portal
│   ├── src/app/                     # Pages (App Router)
│   ├── src/components/              # UI components
│   └── src/lib/                     # API client
├── prompts/                         # AI prompt templates (Markdown)
│   ├── teaching_notes.md
│   ├── assessments.md
│   ├── examples.md
│   ├── translation.md
│   ├── contribution_parser.md
│   ├── document_import.md           # Curriculum import (PDF, DOCX, PPTX, HTML)
│   ├── bulk_import.md               # Multi-topic extraction from large docs
│   └── content_merge.md             # AI-driven content comparison
├── deploy/
│   └── docker/
│       ├── Dockerfile               # Multi-stage: Go + Web build
│       └── Dockerfile.dev
├── scripts/
│   ├── setup.sh
│   └── test-webhook.sh
├── docker-compose.yml
├── Makefile
└── .env.example
```

## Common Commands

```bash
# CLI development
go run ./cmd/oss validate                    # Validate all YAML in local OSS clone
go run ./cmd/oss validate --file <path>      # Validate single file
go run ./cmd/oss import --pdf <file>                 # Import from PDF — whole-PDF mode (default, more robust)
go run ./cmd/oss import --pdf <file> --chunk TAJUK   # Import from PDF — chunk mode (split by keyword)
go run ./cmd/oss import --pdf <file> --from-text "1. Fungsi\n2. Algebra"  # Whole-PDF with topic hints
go run ./cmd/oss import --file <file>                # Import from any format (requires Tika)
go run ./cmd/oss generate teaching-notes <topic-path>
go run ./cmd/oss generate assessments <topic-path> --count 5 --difficulty medium
go run ./cmd/oss generate all --syllabus <id> --subject-grade <id>  # Generate all 4 types for every topic (teaching-notes, assessments, examples, topic enrichment)
go run ./cmd/oss translate --topic <path> --to <lang>
go run ./cmd/oss quality <syllabus-path>                           # Or: --syllabus <id> --subject-grade <id>
go run ./cmd/oss scaffold syllabus --country india --name JEE
go run ./cmd/oss scaffold subject --syllabus india-jee --name Chemistry --grade 11
go run ./cmd/oss import --pdf textbook.pdf --syllabus india-jee --subject-grade india-jee-chemistry-class-11
go run ./cmd/oss import --pdf textbook.pdf --syllabus india-jee --subject-grade india-jee-chemistry-class-11 --chunk "Chapter "

# Bot development
npx smee -u https://smee.io/<channel> -p 8090   # Forward webhooks locally
go run ./cmd/bot                                  # Start webhook handler

# Web portal development
cd web && npm install && npm run dev              # Start at localhost:3001

# Testing
go test ./...
make test
make lint                                         # golangci-lint

# Build
make build-cli                                    # Output: ./bin/oss
make docker                                       # Multi-stage Docker image (includes Tika sidecar)
make setup                                        # First-time setup
```

## Key Architecture Patterns

### AI Provider Interface
All AI providers implement the same interface (shared with P&AI Bot):
```go
type AIProvider interface {
    Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
    StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
    Models() []ModelInfo
}
```

### Unified Pipeline Orchestrator
All three interfaces (CLI, Bot, Web Portal) call the same `pipeline.Execute()` function. No generation logic is duplicated across interfaces — each interface is a thin adapter that parses its input format and delegates to the shared pipeline.

```go
// internal/pipeline/pipeline.go
result, err := pipeline.Execute(ctx, pipeline.Request{
    TopicPath:        "mathematics/algebra/03-simultaneous-equations",
    ContributionType: "teaching_notes",
    Content:          extractedText,
    Mode:             pipeline.ModeCreatePR, // or ModePreview, ModeWriteFS
    Source:           "bot",                 // or "cli", "web"
})
```

**Execution modes:**
- `ModePreview` — generate + validate, return structured output (Web Portal preview, CLI dry-run)
- `ModeWriteFS` — generate + validate + write files to filesystem (CLI default)
- `ModeCreatePR` — generate + validate + create GitHub PR (Bot, Web submit, CLI `--pr`)

**Output writers** (`internal/output/`):
- `LocalWriter` — writes to filesystem (CLI)
- `GitHubWriter` — creates PRs via GitHub API (Bot, Web Portal)

### Content Generation Pipeline
The pipeline executes 7 stages regardless of interface:
1. **Context Builder** — loads topic YAML, related topics, schema rules, style examples (~8K tokens)
2. **AI Generation** — injects context + resolved JSON Schema into prompt template, calls AI provider
3. **Schema Validation** — validates generated YAML against the resolved schema (per-subject override or global fallback). If validation fails, retries once with error feedback injected into the prompt
4. **Content Merge** — compare new vs existing, additive by default
5. **Progress Reporting** — real-time status updates
6. **Bloom Validation** — Bloom's taxonomy level checks on learning objectives
7. **Output** — based on execution mode: write files, open PR, or return preview

### Per-Subject Schema Resolution
Each subject can have its own JSON Schema overrides in a `schemas/` directory. Resolution is per-schema-file:
1. Check `{subjectID}/schemas/{type}.schema.json` (subject-level override)
2. If not found, fall back to `{repoPath}/schema/{type}.schema.json` (global)

This allows different subjects (e.g., English vs Math) to have different schema requirements while sharing a common base. The `scaffold subject` command copies the 6 global schemas into the new subject's `schemas/` directory for customization.

Schema files: `assessments.schema.json`, `concept.schema.json`, `examples.schema.json`, `subject.schema.json`, `syllabus.schema.json`, `topic.schema.json`.

**Contribution types:** `teaching_notes`, `assessments`, `examples`, `topic_enrich`

The `topic_enrich` type is special — instead of creating a companion file, it uses AI to generate structured Level 2 quality fields (`teaching.sequence`, `teaching.common_misconceptions`, `engagement_hooks`) and merges them into the existing topic YAML. This is included automatically in `oss generate all`.

### Provenance Metadata
All generated content includes provenance tracking:
```yaml
_metadata:
  provenance: ai-generated    # human | ai-assisted | ai-observed
  model: <model-name>
  generator: oss-bot v<version>
  generated_at: "<ISO-8601>"
```

### Document Parsing (Hybrid Approach)
The project uses a hybrid approach for document import:
- **CLI (`oss import --pdf`):** Uses `ledongthuc/pdf` (Go-native). Lightweight, no external dependencies. PDF-only.
- **Server (Bot + Web Portal):** Uses Apache Tika as a Docker sidecar via `google/go-tika`. Supports PDF, DOCX, PPTX, XLSX, HTML, and 1000+ other formats.

Both implementations share the `DocumentParser` interface in `internal/parser/document.go`:
```go
type DocumentParser interface {
    Extract(ctx context.Context, input []byte, mimeType string) (string, error)
}
```

### Content Merge Strategy
When new content targets a topic with existing content:
- **Assessments/Examples:** Append + dedup (>95% skip, 85-95% flag in PR)
- **Teaching Notes:** Additive only — keep all existing sections, add new knowledge
- **Principle:** Never silently drop content. Additive by default.

### Prompt Templates
Located in `prompts/` as Markdown files with template variables (e.g., `{{topic}}`, `{{prerequisites}}`). These encode pedagogical best practices and output format requirements. All prompts are curriculum-agnostic — use `{{syllabus_id}}` template variables, never hardcode specific curricula.

## Environment Variables

All config uses `OSS_` prefix. Key variables:

| Variable | Required | Description |
|----------|----------|-------------|
| `OSS_AI_PROVIDER` | Yes | `openai`, `anthropic`, or `ollama` |
| `OSS_AI_API_KEY` | No* | API key for chosen provider |
| `OSS_REPO_PATH` | Yes (CLI) | Path to local OSS clone |
| `OSS_GITHUB_APP_ID` | Yes (bot) | GitHub App ID |
| `OSS_GITHUB_PRIVATE_KEY_PATH` | Yes (bot) | Path to GitHub App private key |
| `OSS_GITHUB_WEBHOOK_SECRET` | Yes (bot) | Webhook HMAC secret |
| `OSS_REPO_OWNER` | Yes | GitHub org/user (default: `p-n-ai`) |
| `OSS_REPO_NAME` | Yes | Repository name (default: `oss`) |
| `OSS_TIKA_URL` | No | Tika server URL (default: `http://tika:9998`, server only) |
| `OSS_AI_REASONING_PROVIDER` | No | Reasoning model provider (default: `openrouter`) |
| `OSS_AI_REASONING_API_KEY` | No | OpenRouter API key for reasoning models |
| `OSS_AI_REASONING_MODEL` | No | Reasoning model on OpenRouter (default: `deepseek/deepseek-r1`) |
| `OSS_WORKER_COUNT` | No | Parallel workers for bulk import (default: 3) |

*Not needed for Ollama.

## ID Conventions

All entity IDs follow **[`docs/id-conventions.md`](docs/id-conventions.md)**. Core principle:

> **IDs use the official MOE language of the country. English is always added to content fields (`name_en`, `text_en`) for interoperability.**

| Entity | Format | Example |
|--------|--------|---------|
| Country | `{english-name}` slug | `malaysia`, `india` |
| Syllabus | `{country}-{board}` | `malaysia-kssm` |
| Grade | `{grade-name}` in MOE language | `tingkatan-1`, `class-10` |
| Subject | `{syllabus}-{subject}` in MOE language | `malaysia-kssm-matematik` |
| Subject Grade | `{subject_id}-{grade}` | `malaysia-kssm-matematik-tingkatan-1` |
| Topic | `{PREFIX}{grade_num}-{NN}` (prefix from English name) | `MT1-01`, `MT3-09`, `PHY12-03` |

Every generated YAML must include both `name` (MOE language) and `name_en` (English).

## Related Repositories

| Repository | Relationship |
|-----------|-------------|
| [p-n-ai/oss](https://github.com/p-n-ai/oss) | Target data repo. OSS Bot creates PRs here. Reads content for context. |
| [p-n-ai/pai-bot](https://github.com/p-n-ai/pai-bot) | AI learning companion. Shares AI provider interface. Submits feedback via `/api/feedback`. |

## Development Workflow

### Daily Implementation — Required Reading

**MANDATORY:** Before starting any day's implementation work, you MUST read and cross-reference BOTH of these documents:

1. **[`docs/development-timeline.md`](docs/development-timeline.md)** — the daily task breakdown with task IDs, ownership, and sequencing
2. **[`docs/implementation-guide.md`](docs/implementation-guide.md)** — the step-by-step executable instructions with code templates, tests, file paths, entry/exit criteria, and validation commands

These two documents are complementary and both are required:
- The **timeline** tells you WHAT to build each day and in what order (task IDs like `B-W4D16-1`)
- The **implementation guide** tells you HOW to build it (exact file paths, code, tests, validation steps)

**Do not implement from the timeline alone** — it lacks the detail needed for correct implementation. **Do not implement from the guide alone** — you may miss sequencing dependencies and ownership context from the timeline.

Workflow for each day:
1. Read the day's section in `docs/development-timeline.md` for task overview
2. Read the matching day's section in `docs/implementation-guide.md` for detailed instructions
3. Check **entry criteria** in the guide before starting
4. Follow the TDD workflow below for each task
5. Run **validation commands** from the guide
6. Verify all **exit criteria** checkboxes before moving to the next day

### Updating Task Status in the Timeline

When completing tasks, update the corresponding row in `docs/development-timeline.md`:

| Column | Description |
|--------|-------------|
| Task ID | e.g. `B-W4D16-1` — do not modify |
| Task | Task description — do not modify |
| Owner | `🤖` or `🧑` — do not modify |
| Status | `⬜` = not started, `✅` = completed |
| Remark | Add any notes (e.g. deviations, blockers, decisions made) |

Mark each task's Status as `✅` as soon as it is done. Add context in the Remark column when relevant.

### Test-Driven Development (TDD)
This project follows a strict test-first approach. Every feature goes through this cycle:

1. **Write tests first** — before implementing any feature, write unit tests that define expected behavior
2. **Implement** — write the minimum code to make the tests pass
3. **Run package tests** — verify the new feature works (`go test ./internal/<package>/...`)
4. **Run full test suite** — run `go test ./...` to ensure nothing is broken across the entire codebase
5. **Never skip step 4** — every completed feature must pass the full suite before moving on

**Testing conventions:**
- Use stdlib `testing` with table-driven tests and `t.Run()` subtests
- **Mock AI providers** for deterministic test output — never call real AI APIs in tests
- **Test files** live alongside source: `validator.go` → `validator_test.go`

```bash
# Run all tests (REQUIRED after every feature)
go test ./...

# Run tests for a specific package
go test ./internal/validator/...

# Run with verbose output
go test -v ./...

# Run a specific test
go test -run TestValidateSchema ./internal/validator/...
```

## Development Notes

- **Build order priority:** Validator -> Generator -> GitHub Bot -> Web Portal
- **Week 4 complete** — CLI fully functional with validate, generate, quality, translate. 58 tests passing across 5 packages (validator, ai, generator, pipeline, output).
- The AI provider interface is shared with P&AI Bot; keep implementations compatible
- All PRs from the bot require human review before merging
- Schema validation must block invalid content — never submit invalid YAML
- Prompt templates in `prompts/` are critical for output quality; changes need educator review
- The web portal calls the Go backend API; both run as Docker containers via docker-compose
