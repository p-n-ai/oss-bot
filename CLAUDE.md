# CLAUDE.md — OSS Bot

This file provides context for Claude Code when working on this repository.

## Project Overview

OSS Bot is the AI-powered tooling layer for the [Open School Syllabus (OSS)](https://github.com/p-n-ai/oss). It provides three interfaces to contribute structured curriculum content:

1. **GitHub Bot** (`@oss-bot`) — comment-driven content generation in the OSS repo
2. **CLI Tool** (`oss`) — local validation, generation, import, translation
3. **Web Portal** (`contribute.p-n-ai.org`) — teacher-friendly contribution form

All tools share a single AI content generation pipeline: Context Building -> AI Generation -> Validation -> Output (PR).

## Tech Stack

### Backend (Go)
- **Language:** Go >= 1.22
- **CLI Framework:** `spf13/cobra`
- **GitHub API:** `google/go-github/v62`
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
│   │   └── ollama.go
│   ├── generator/                   # Content generation pipeline
│   │   ├── context.go               # Context builder (loads topic, schema, etc.)
│   │   ├── teaching_notes.go
│   │   ├── assessments.go
│   │   ├── examples.go
│   │   ├── translator.go
│   │   ├── scaffolder.go
│   │   └── importer.go              # Document -> curriculum import (PDF, DOCX, PPTX, HTML)
│   ├── validator/                   # Schema validation
│   │   ├── validator.go             # JSON Schema engine
│   │   ├── bloom.go                 # Bloom's taxonomy checks
│   │   ├── prerequisites.go         # Prerequisite graph cycle detection
│   │   ├── duplicates.go            # Duplicate content detection
│   │   └── quality.go               # Quality level assessment
│   ├── parser/                      # Input parsing + document extraction
│   │   ├── command.go               # Parse @oss-bot commands
│   │   ├── contribution.go          # Natural language -> structured data
│   │   ├── document.go              # DocumentParser interface
│   │   ├── pdf.go                   # PDF text extraction (Go-native, for CLI)
│   │   └── tika.go                  # Apache Tika client (multi-format, for server)
│   ├── github/                      # GitHub API integration
│   │   ├── app.go                   # GitHub App auth (JWT + installation tokens)
│   │   ├── webhook.go               # Webhook handler + HMAC verification
│   │   ├── pr.go                    # PR creation, labels, reviewers
│   │   └── contents.go              # Read/write via GitHub Contents API
│   └── api/                         # Web portal backend
│       ├── router.go
│       ├── preview.go               # POST /api/preview
│       ├── submit.go                # POST /api/submit
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
│   └── document_import.md           # Curriculum import (PDF, DOCX, PPTX, HTML)
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
go run ./cmd/oss import --pdf <file>                 # Import from PDF (CLI, Go-native)
go run ./cmd/oss import --file <file>                # Import from any format (requires Tika)
go run ./cmd/oss generate teaching-notes <topic-path>
go run ./cmd/oss generate assessments <topic-path> --count 5 --difficulty medium
go run ./cmd/oss translate --topic <path> --to <lang>
go run ./cmd/oss quality <syllabus-path>

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

### Content Generation Pipeline
Every generation follows 4 stages regardless of interface:
1. **Context Builder** — loads topic YAML, related topics, schema rules, style examples (~8K tokens)
2. **AI Generation** — injects context into prompt template, calls AI provider
3. **Validation** — JSON Schema check, Bloom's taxonomy, prerequisite graph, duplicate detection
4. **Output** — write files, open PR with provenance labels and quality assessment

If validation fails, the pipeline retries once with error feedback before reporting failure.

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

### Prompt Templates
Located in `prompts/` as Markdown files with template variables (e.g., `{{topic}}`, `{{prerequisites}}`). These encode pedagogical best practices and output format requirements.

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

*Not needed for Ollama.

## Related Repositories

| Repository | Relationship |
|-----------|-------------|
| [p-n-ai/oss](https://github.com/p-n-ai/oss) | Target data repo. OSS Bot creates PRs here. Reads content for context. |
| [p-n-ai/pai-bot](https://github.com/p-n-ai/pai-bot) | AI learning companion. Shares AI provider interface. Submits feedback via `/api/feedback`. |

## Development Workflow

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
- **No code exists yet** — this repo is currently documentation-only (plans, README, business docs)
- The AI provider interface is shared with P&AI Bot; keep implementations compatible
- All PRs from the bot require human review before merging
- Schema validation must block invalid content — never submit invalid YAML
- Prompt templates in `prompts/` are critical for output quality; changes need educator review
- The web portal calls the Go backend API; both run as Docker containers via docker-compose
