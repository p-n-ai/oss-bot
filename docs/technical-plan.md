# Technical Plan — OSS Bot

> **Repository:** `p-n-ai/oss-bot`
> **License:** Apache 2.0
> **Last updated:** February 2026

---

## 1. Architecture Overview

OSS Bot is a **tooling layer** for the [Open School Syllabus](https://github.com/p-n-ai/oss) repository. It provides three interfaces — a GitHub Bot, a CLI tool, and a web portal — all powered by a shared AI content generation pipeline. The core problem it solves: teachers have pedagogical knowledge but can't contribute to a Git/YAML repository. OSS Bot bridges this gap.

```
┌───────────────────────────────────────────────────────────────┐
│  Input Interfaces                                             │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────────┐ │
│  │ GitHub Bot   │  │ CLI Tool     │  │ Web Portal          │ │
│  │ (@oss-bot)   │  │ (oss)        │  │ (contribute.        │ │
│  │              │  │              │  │  opensyllabus.org)   │ │
│  │ Webhook      │  │ Binary       │  │ Next.js             │ │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬──────────┘ │
│         │                 │                      │            │
│         └─────────────────┼──────────────────────┘            │
│                           ▼                                   │
│  ┌────────────────────────────────────────────────────────┐   │
│  │  Shared Pipeline                                       │   │
│  │                                                        │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │   │
│  │  │ Context      │  │ AI Content   │  │ Validation   │ │   │
│  │  │ Builder      │  │ Generator    │  │ Engine       │ │   │
│  │  │              │  │              │  │              │ │   │
│  │  │ Load topic   │  │ Pedagogical  │  │ JSON Schema  │ │   │
│  │  │ Load related │  │ prompts      │  │ Bloom's check│ │   │
│  │  │ Load schema  │  │ Schema-aware │  │ Prereq graph │ │   │
│  │  │ Load existing│  │ Style match  │  │ Duplicate det│ │   │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘ │   │
│  │         └─────────────────┼──────────────────┘         │   │
│  └────────────────────────────┼───────────────────────────┘   │
│                               ▼                               │
│  ┌────────────────────────────────────────────────────────┐   │
│  │  Output Layer                                          │   │
│  │  ├── Write YAML/Markdown files                         │   │
│  │  ├── Open GitHub Pull Request (via GitHub API)         │   │
│  │  ├── Add provenance labels + quality assessment        │   │
│  │  └── Request appropriate reviewers                     │   │
│  └────────────────────────────────────────────────────────┘   │
└───────────────────────────────────────────────────────────────┘
         │
         ▼
    ┌─────────┐
    │ p-n-ai/ │
    │ oss     │  (Pull Requests land here)
    └─────────┘
```

---

## 2. Tech Stack

### 2.1 Backend (Go — Shared Core)

OSS Bot uses Go to match the P&AI Bot stack, enabling code sharing and consistent AI provider interfaces across the ecosystem.

| Component | Technology | Version | Rationale |
|-----------|-----------|---------|-----------|
| **Language** | Go | ≥1.22 | Matches P&AI Bot stack. Single binary for CLI distribution. Goroutines for concurrent API calls. |
| **HTTP Router** | Go stdlib `net/http` | 1.22+ | Webhook handler for GitHub events. Minimal dependencies. |
| **AI Providers** | Custom interface | — | Same provider abstraction as P&AI Bot. Supports OpenAI, Anthropic, Ollama. |
| **GitHub API** | `google/go-github` | v6 | GitHub App authentication, PR creation, issue commenting, file operations. |
| **YAML Parsing** | `go-yaml/yaml` | v3 | Read and write OSS curriculum YAML files. |
| **JSON Schema** | `santhosh-tekuri/jsonschema` | v5 | In-process schema validation (no shelling out to ajv). |
| **PDF Parsing** | `ledongthuc/pdf` or `unidoc/unipdf` | latest | Extract text and structure from curriculum PDF documents. |
| **CLI Framework** | `spf13/cobra` | v1 | Industry-standard Go CLI framework. Subcommands, flags, help generation. |
| **Configuration** | Environment variables | — | All config via `OSS_` prefixed env vars. |
| **Testing** | Go stdlib `testing` | — | Table-driven tests. Mock AI providers for deterministic output. |

### 2.2 AI Provider Interface

```go
// Shared with P&AI Bot — same interface, same providers
type AIProvider interface {
    Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
    StreamComplete(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
    Models() []ModelInfo
}
```

| Provider | Best For | Model | Config |
|----------|---------|-------|--------|
| **OpenAI** | General content generation, assessments | GPT-4o, GPT-4o-mini | `OSS_AI_PROVIDER=openai` |
| **Anthropic** | Teaching notes, nuanced pedagogy, natural language parsing | Claude Sonnet, Claude Haiku | `OSS_AI_PROVIDER=anthropic` |
| **Ollama** | Free/offline usage, privacy-sensitive deployments | Llama 3, Mistral, Gemma | `OSS_AI_PROVIDER=ollama` |

### 2.3 Web Portal (Next.js)

| Component | Technology | Version | Rationale |
|-----------|-----------|---------|-----------|
| **Framework** | Next.js (App Router) | 14 | Consistent with P&AI Bot admin panel. SSR for SEO. |
| **Language** | TypeScript | 5.x | Type safety, form handling. |
| **UI Components** | shadcn/ui | latest | Consistent design system across P&AI ecosystem. |
| **Styling** | Tailwind CSS | 3.x | Utility-first, matching other P&AI frontends. |
| **Form Handling** | React Hook Form + Zod | latest | Validated forms for contribution input. |
| **State** | React Query (TanStack Query) | v5 | Server state for contribution preview and submission. |

### 2.4 GitHub App Integration

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Authentication** | GitHub App (JWT + installation tokens) | Authenticate as the bot to create PRs, comment on issues |
| **Webhook Handler** | Go `net/http` | Receives `issue_comment` and `pull_request` events |
| **Webhook Verification** | HMAC-SHA256 | Verify webhook payloads from GitHub |
| **Webhook Forwarding (dev)** | smee.io | Forward GitHub webhooks to localhost during development |

---

## 3. Three Interfaces, One Pipeline

### 3.1 GitHub Bot (`@oss-bot`)

**Runtime:** Long-running Go server receiving GitHub webhooks.

**Trigger:** Mention `@oss-bot` in any issue or PR comment in the `p-n-ai/oss` repository.

**Commands:**

| Command | Example | Action |
|---------|---------|--------|
| `add teaching notes` | `@oss-bot add teaching notes for cambridge/igcse/.../05-quadratic-equations` | Generate `.teaching.md`, open PR |
| `add N assessments` | `@oss-bot add 5 assessments for .../03-simultaneous-equations difficulty:medium` | Generate `.assessments.yaml`, open PR |
| `translate` | `@oss-bot translate .../01-expressions to ms` | Generate translated YAML in `locales/ms/`, open PR tagged `needs-native-review` |
| `scaffold syllabus` | `@oss-bot scaffold syllabus india/cbse/mathematics-class10` | Create directory structure + Level 0 stubs, open PR with completeness checklist |
| `import --pdf` | `@oss-bot import --pdf [attached PDF]` | Extract structure from PDF, generate Level 0–1 stubs, open PR |
| `enrich` | `@oss-bot enrich .../05-quadratic-equations` (with natural language body) | Parse teacher's experience into structured misconceptions, teaching notes, open PR |
| `quality` | `@oss-bot quality cambridge/igcse/mathematics-0580` | Comment with quality report (no PR) |

**Webhook flow:**

```
GitHub issue_comment event
    │
    ▼
Webhook handler (Go)
    ├── Verify HMAC signature
    ├── Parse command from comment body
    ├── Authenticate as GitHub App (installation token)
    │
    ▼
Command router
    ├── Extract topic path, options, body text
    ├── Load current OSS content from repo (via GitHub Contents API)
    │
    ▼
Shared pipeline (Context → Generate → Validate)
    │
    ▼
Output
    ├── Create branch: oss-bot/{command}-{topic}-{timestamp}
    ├── Commit generated files
    ├── Open PR with description, provenance label, quality assessment
    ├── Request reviewers (educators from CODEOWNERS)
    └── React to original comment with 👍 + link to PR
```

### 3.2 CLI Tool (`oss`)

**Runtime:** Single Go binary. No server required. Runs locally against a cloned OSS repository.

**Distribution:**

```bash
# From source
go install github.com/p-n-ai/oss-bot/cmd/oss@latest

# Pre-built binary (Linux/macOS/Windows)
curl -sSL https://github.com/p-n-ai/oss-bot/releases/latest/download/oss-$(uname -s)-$(uname -m) \
  -o /usr/local/bin/oss && chmod +x /usr/local/bin/oss
```

**Commands:**

| Command | Purpose | Requires AI? |
|---------|---------|-------------|
| `oss validate` | Schema validation of all YAML files | No |
| `oss validate --file <path>` | Validate a single file | No |
| `oss validate --syllabus <path>` | Validate an entire syllabus | No |
| `oss generate teaching-notes <topic-path>` | Generate AI teaching notes | Yes |
| `oss generate assessments <topic-path> --count N --difficulty <level>` | Generate assessment questions | Yes |
| `oss generate examples <topic-path> --count N` | Generate worked examples | Yes |
| `oss import --pdf <file> --board <board> --level <level> --subject <subject>` | Import curriculum from PDF | Yes |
| `oss translate --topic <path> --to <lang>` | Translate topic to target language | Yes |
| `oss translate --syllabus <path> --to <lang>` | Translate all topics in a syllabus | Yes |
| `oss quality <syllabus-path>` | Quality report for a syllabus | No |
| `oss contribute "<natural language>"` | Parse natural language into structured PR | Yes |

**CLI architecture:**

```
cmd/oss/main.go
    │
    ├── cobra root command
    │   ├── validate subcommand  → internal/validator/
    │   ├── generate subcommand  → internal/generator/
    │   ├── import subcommand    → internal/parser/pdf.go → internal/generator/
    │   ├── translate subcommand → internal/generator/translator.go
    │   ├── quality subcommand   → internal/validator/ (quality assessment mode)
    │   └── contribute subcommand→ internal/parser/contribution.go → internal/generator/
    │
    └── All subcommands share:
        ├── internal/ai/         (AI provider interface)
        ├── internal/validator/  (JSON Schema validation)
        └── internal/github/     (optional: PR creation if --pr flag set)
```

### 3.3 Web Portal (`contribute.opensyllabus.org`)

**Runtime:** Next.js application served via Docker container. Calls the Go backend API for AI generation and GitHub PR creation.

**Architecture:**

```
Browser (Teacher)
    │
    ▼
Next.js Frontend (TypeScript)
    ├── Step 1: Select syllabus + topic (or "Add new syllabus")
    ├── Step 2: Type contribution in natural language (any language)
    ├── Step 3: Preview structured output (YAML/Markdown rendered)
    ├── Step 4: Confirm and submit
    │
    ▼
Go API Backend (same binary as GitHub bot)
    ├── POST /api/preview   → Run pipeline, return structured preview
    ├── POST /api/submit    → Create branch, commit, open PR
    └── GET  /api/curricula → List available syllabi and topics
    │
    ▼
GitHub (p-n-ai/oss)
    └── PR created with attribution to the contributor
```

**Web portal features:**

- No GitHub account required (PR is created by the bot, with attribution in the PR description)
- Supports plain language input in any language
- Real-time preview of structured output before submission
- Schema validation before submission (invalid content blocked)
- Optional GitHub sign-in to track contributions across sessions

---

## 4. AI Content Generation Pipeline

This is the shared core of all three interfaces. The pipeline's differentiator is **context building** — it doesn't just ask an LLM to generate content, it provides rich context about the topic, its neighbors, and the quality standards expected.

### 4.1 Pipeline Stages

```
Input (command + topic path + optional natural language)
    │
    ▼
┌──────────────────────────────────────┐
│  Stage 1: Context Builder            │
│                                      │
│  Load from OSS repository:           │
│  ├── Target topic YAML               │
│  ├── Parent subject + syllabus       │
│  ├── Prerequisite topics (content)   │
│  ├── Sibling topics (for style)      │
│  ├── Existing teaching notes/examples│
│  ├── JSON Schema rules               │
│  └── Quality level standards         │
│                                      │
│  Build context window (~8K tokens)   │
└──────────────────┬───────────────────┘
                   │
                   ▼
┌──────────────────────────────────────┐
│  Stage 2: AI Generation              │
│                                      │
│  Select prompt template:             │
│  ├── prompts/teaching_notes.md       │
│  ├── prompts/assessments.md          │
│  ├── prompts/examples.md             │
│  ├── prompts/translation.md          │
│  └── prompts/contribution_parser.md  │
│                                      │
│  Inject context into template        │
│  Call AI provider (streaming)        │
│  Parse structured output (YAML/MD)   │
└──────────────────┬───────────────────┘
                   │
                   ▼
┌──────────────────────────────────────┐
│  Stage 3: Validation                 │
│                                      │
│  ├── JSON Schema validation          │
│  ├── Bloom's taxonomy level check    │
│  ├── Prerequisite graph integrity    │
│  ├── Duplicate detection (vs exist.) │
│  ├── Copyright check (flag verbatim) │
│  └── Quality level self-assessment   │
│                                      │
│  If validation fails:                │
│  ├── Retry with error feedback (1x)  │
│  └── If still fails: report error    │
└──────────────────┬───────────────────┘
                   │
                   ▼
┌──────────────────────────────────────┐
│  Stage 4: Output                     │
│                                      │
│  ├── Write files to branch           │
│  ├── Add provenance metadata         │
│  │   provenance: ai-generated        │
│  │   model: claude-sonnet-4-20250514 │
│  │   generated_at: 2026-02-27T14:00Z │
│  │   context_topics: [01-expr, ...]  │
│  ├── Open GitHub PR                  │
│  ├── Add labels (provenance, quality)│
│  ├── Add quality self-assessment     │
│  └── Request educator reviewers      │
└──────────────────────────────────────┘
```

### 4.2 Prompt Templates

Prompt templates live in `prompts/` as Markdown files with template variables. Each template encodes pedagogical best practices and output format requirements.

| Template | Purpose | Key Instructions |
|----------|---------|-----------------|
| `teaching_notes.md` | Generate `.teaching.md` files | Write for AI chat delivery. Start with engagement hook, not definition. Include scaffolding for when student is stuck. End with forward look. |
| `assessments.md` | Generate `.assessments.yaml` files | Include worked solutions, rubrics, progressive hints, and common wrong answers with targeted feedback. Distribute across Bloom's levels. |
| `examples.md` | Generate `.examples.yaml` files | Worked examples with step-by-step solutions. Progressive difficulty. Connect to real-world contexts. |
| `translation.md` | Translate topic files | Preserve structure exactly. Translate only human-readable text fields. Use mathematically correct terminology in target language. |
| `contribution_parser.md` | Parse natural language into structured data | Identify contribution type (misconception, teaching note, assessment, etc.). Extract structured fields. Preserve teacher's voice where possible. |
| `pdf_import.md` | Extract curriculum structure from PDF | Identify subjects, topics, learning objectives. Infer Bloom's levels from specification verbs. Map prerequisite relationships. |

### 4.3 Context Building Strategy

Context building is what makes OSS Bot's output pedagogically coherent rather than generic. For each generation request, the context builder assembles:

```go
type GenerationContext struct {
    // Target
    Topic          Topic           // The topic being generated for
    Subject        Subject         // Parent subject
    Syllabus       Syllabus        // Parent syllabus

    // Neighbors (for style matching and coherence)
    Prerequisites  []Topic         // Topics this one depends on
    Siblings       []Topic         // Other topics in the same subject
    ExistingNotes  string          // Current teaching notes (if any)
    ExistingAssess []Assessment    // Current assessments (if any)

    // Standards
    Schema         json.RawMessage // JSON Schema for output validation
    QualityRules   QualityRules    // What's needed for each quality level
    StyleExamples  []string        // Example outputs from sibling topics (for tone matching)
}
```

This context is injected into the prompt template before the AI call, ensuring the output matches the existing style, references correct prerequisites, and meets schema requirements.

---

## 5. Project Structure

```
oss-bot/
├── cmd/
│   ├── oss/                         # CLI entrypoint
│   │   └── main.go                  # cobra root + subcommands
│   └── bot/                         # GitHub Bot + Web Portal entrypoint
│       └── main.go                  # HTTP server (webhooks + API + static)
├── internal/
│   ├── ai/                          # AI provider interface (shared with P&AI Bot)
│   │   ├── provider.go              # Provider interface definition
│   │   ├── openai.go                # OpenAI implementation
│   │   ├── anthropic.go             # Anthropic implementation
│   │   └── ollama.go                # Ollama implementation
│   ├── generator/                   # Content generation (Stage 2)
│   │   ├── context.go               # Context builder (Stage 1)
│   │   ├── teaching_notes.go        # Teaching notes generator
│   │   ├── assessments.go           # Assessment question generator
│   │   ├── examples.go              # Worked examples generator
│   │   ├── translator.go            # Topic translation
│   │   ├── scaffolder.go            # New syllabus scaffolding
│   │   └── importer.go              # PDF → structured curriculum import
│   ├── validator/                   # Schema validation (Stage 3)
│   │   ├── validator.go             # JSON Schema validation engine
│   │   ├── bloom.go                 # Bloom's taxonomy level verification
│   │   ├── prerequisites.go         # Prerequisite graph cycle detection
│   │   ├── duplicates.go            # Duplicate content detection
│   │   └── quality.go               # Quality level auto-assessment
│   ├── parser/                      # Input parsing
│   │   ├── command.go               # Parse @oss-bot commands from comments
│   │   ├── contribution.go          # Natural language → structured contribution
│   │   └── pdf.go                   # PDF text extraction and structure detection
│   ├── github/                      # GitHub API integration
│   │   ├── app.go                   # GitHub App authentication (JWT + installation tokens)
│   │   ├── webhook.go               # Webhook handler + HMAC verification
│   │   ├── pr.go                    # PR creation, labeling, reviewer assignment
│   │   └── contents.go              # Read/write files via GitHub Contents API
│   └── api/                         # Web portal API
│       ├── router.go                # HTTP routes for web portal backend
│       ├── preview.go               # POST /api/preview handler
│       ├── submit.go                # POST /api/submit handler
│       └── curricula.go             # GET /api/curricula handler
├── web/                             # Contribution web portal (Next.js)
│   ├── src/
│   │   ├── app/
│   │   │   ├── page.tsx             # Landing / syllabus selector
│   │   │   ├── contribute/
│   │   │   │   └── page.tsx         # Contribution form
│   │   │   └── preview/
│   │   │       └── page.tsx         # Preview and submit
│   │   ├── components/
│   │   │   ├── syllabus-picker.tsx
│   │   │   ├── topic-picker.tsx
│   │   │   ├── contribution-form.tsx
│   │   │   ├── yaml-preview.tsx     # Rendered YAML preview
│   │   │   └── submission-status.tsx
│   │   └── lib/
│   │       └── api.ts               # API client for Go backend
│   ├── package.json
│   ├── next.config.js
│   ├── tailwind.config.ts
│   └── tsconfig.json
├── prompts/                         # AI prompt templates (Markdown)
│   ├── teaching_notes.md
│   ├── assessments.md
│   ├── examples.md
│   ├── translation.md
│   ├── contribution_parser.md
│   └── pdf_import.md
├── deploy/
│   ├── docker/
│   │   ├── Dockerfile               # Multi-stage: Go build + Web build
│   │   └── Dockerfile.dev           # Development with hot reload
│   └── helm/
│       └── oss-bot/                 # Helm chart (optional, for K8s deployment)
│           ├── Chart.yaml
│           ├── values.yaml
│           └── templates/
├── scripts/
│   ├── setup.sh                     # First-time setup
│   └── test-webhook.sh              # Send test webhook payload locally
├── docker-compose.yml               # Local dev: Go server + Web portal + Ollama
├── Makefile                         # Dev shortcuts
├── .env.example                     # All configuration documented
├── .github/
│   └── workflows/
│       ├── ci.yml                   # Test + lint + build on every PR
│       └── release.yml              # Build binaries + Docker image on tag
├── go.mod
├── go.sum
└── README.md
```

---

## 6. Quality Safeguards

Every piece of generated content passes through automated and human quality gates before reaching the OSS repository.

### 6.1 Automated Checks (Block PR if Failed)

| Check | Tool | Description |
|-------|------|-------------|
| **JSON Schema validation** | `santhosh-tekuri/jsonschema` | Every YAML file must validate against its schema |
| **Bloom's taxonomy verification** | Custom Go code | Learning objectives must use valid Bloom's verbs. Assessment difficulty must align with Bloom's level. |
| **Prerequisite graph integrity** | Custom Go code | No circular dependencies. All referenced topic IDs must exist. |
| **Duplicate detection** | Embedding similarity | New assessments compared against existing ones. Flag if >85% similar. |
| **YAML syntax** | `go-yaml/yaml` parser | Valid YAML structure |

### 6.2 Automated Checks (Warning, Reviewer Decides)

| Check | Description |
|-------|-------------|
| **Quality level not decreased** | Warn if a change would lower a topic's quality level |
| **Copyright flag** | Flag any content that appears to be verbatim from a known source |
| **Translation completeness** | Warn if translation is missing fields present in source |
| **Assessment balance** | Warn if all assessments are same difficulty level |

### 6.3 Human Review (Required)

| Content Type | Required Reviewer |
|-------------|-------------------|
| `provenance: ai-generated` | At least 1 educator with subject expertise |
| `provenance: ai-observed` | At least 1 educator with subject expertise |
| Translations | Native speaker of target language |
| New syllabus scaffold | Educator familiar with the curriculum |
| `provenance: human` or `ai-assisted` | Standard PR review |

### 6.4 Provenance Metadata

Every generated file includes provenance metadata:

```yaml
# Added automatically by OSS Bot
_metadata:
  provenance: ai-generated
  model: claude-sonnet-4-20250514
  generator: oss-bot v0.1.0
  generated_at: "2026-02-27T14:00:00Z"
  context_topics:
    - 01-expressions
    - 02-linear-equations
    - 04-inequalities
  reviewed_by: null                   # Filled after human review
```

---

## 7. P&AI Bot Feedback Pipeline

OSS Bot exposes an API endpoint that P&AI Bot calls to submit data-driven improvement suggestions.

```
P&AI Bot (teaching students)
    │
    │  Observes: "73% of students make sign errors on topic X"
    │  Observes: "Explanation A works 20% better than B"
    │  Observes: "Students keep asking about Y (not in syllabus)"
    │
    ▼
POST /api/feedback
{
  "type": "misconception_observed",
  "topic_path": "cambridge/igcse/mathematics-0580/topics/algebra/05-quadratic-equations",
  "data": {
    "misconception": "Students write (x+3)(x-3) = x² - 9 but fail with (x+3)(x-2)",
    "frequency": 0.73,
    "sample_size": 142,
    "observed_period": "2026-01-01/2026-02-27"
  }
}
    │
    ▼
OSS Bot Pipeline
    ├── Context build (load current topic data)
    ├── AI generation (structure observation into misconception entry)
    ├── Validation (schema check, duplicate check)
    ├── Quality self-assessment
    │
    ▼
GitHub PR
    ├── Branch: oss-bot/ai-observed-{topic}-{timestamp}
    ├── Label: provenance:ai-observed
    ├── Description includes: sample size, frequency, observation period
    └── Auto-requests educator reviewer
```

**Feedback types supported:**

| Type | P&AI Observation | OSS Bot Action |
|------|-----------------|----------------|
| `misconception_observed` | High-frequency error pattern | Add to `teaching.common_misconceptions` |
| `explanation_effectiveness` | One explanation outperforms another | Update teaching notes with more effective approach |
| `content_gap` | Students ask about topic not in syllabus | Open issue flagging potential content gap |
| `difficulty_mismatch` | Students consistently fail/ace a topic | Suggest difficulty level adjustment |
| `engagement_hook_effectiveness` | One hook drives more session completion | Reorder engagement hooks by effectiveness |

---

## 8. Infrastructure & Deployment

### 8.1 Hosted by Pandai (Default)

| Component | Technology | Cost |
|-----------|-----------|------|
| **GitHub Bot** | Go binary on Docker | ~$10/month (small VPS) |
| **Web Portal** | Next.js on Docker | ~$10/month (same VPS) |
| **Ollama** (optional) | Self-hosted LLM | ~$30/month (if GPU instance) |
| **Total** | | **~$20–50/month** |

Both the Go server and Next.js portal run as Docker containers via `docker-compose.yml` on a single VPS. Scales with contribution volume — at low volume (<100 contributions/month), a $10 VPS is sufficient.

### 8.2 Self-Hosted (For OSS Forks)

Organizations that fork OSS can run their own OSS Bot instance:

```bash
git clone https://github.com/p-n-ai/oss-bot.git
cd oss-bot
cp .env.example .env
# Edit .env: GitHub App credentials, AI API key, target repo
docker compose up -d
```

**GitHub App setup:**

1. Create GitHub App at `github.com/settings/apps/new`
2. Webhook URL: `https://your-bot.example.com/webhook`
3. Permissions: Issues (R/W), Pull Requests (R/W), Contents (R/W)
4. Subscribe to events: `issue_comment`, `pull_request`
5. Install on your OSS fork
6. Set App ID, private key, webhook secret in `.env`

### 8.3 Docker Build

```dockerfile
# Stage 1: Build Go binaries
FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /oss-bot ./cmd/bot
RUN CGO_ENABLED=0 go build -o /oss-cli ./cmd/oss

# Stage 2: Build web portal
FROM node:20-alpine AS web-builder
WORKDIR /web
COPY web/package*.json ./
RUN npm ci
COPY web/ .
RUN npm run build

# Stage 3: Final image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=go-builder /oss-bot /usr/local/bin/oss-bot
COPY --from=go-builder /oss-cli /usr/local/bin/oss
COPY --from=web-builder /web/.next /web/.next
COPY --from=web-builder /web/public /web/public
COPY prompts/ /prompts/
ENTRYPOINT ["oss-bot"]
```

---

## 9. Configuration Reference

All configuration via environment variables with `OSS_` prefix.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `OSS_GITHUB_APP_ID` | Yes (bot) | — | GitHub App ID |
| `OSS_GITHUB_PRIVATE_KEY_PATH` | Yes (bot) | — | Path to GitHub App private key `.pem` file |
| `OSS_GITHUB_WEBHOOK_SECRET` | Yes (bot) | — | Webhook secret for HMAC verification |
| `OSS_REPO_OWNER` | Yes | `p-n-ai` | GitHub org/user owning the OSS repo |
| `OSS_REPO_NAME` | Yes | `oss` | OSS repository name |
| `OSS_REPO_PATH` | Yes (CLI) | `./oss` | Local path to OSS clone (CLI only) |
| `OSS_AI_PROVIDER` | Yes | — | `openai`, `anthropic`, or `ollama` |
| `OSS_AI_API_KEY` | No* | — | API key for chosen provider |
| `OSS_AI_OLLAMA_URL` | No | `http://ollama:11434` | Ollama server URL |
| `OSS_AI_MODEL` | No | Provider default | Override default model selection |
| `OSS_WEB_PORT` | No | `3001` | Web portal HTTP port |
| `OSS_BOT_PORT` | No | `8090` | Webhook handler HTTP port |
| `OSS_LOG_LEVEL` | No | `info` | `debug`, `info`, `warn`, `error` |
| `OSS_PROMPTS_DIR` | No | `./prompts` | Path to prompt template directory |
| `OSS_GITHUB_TOKEN` | No (CLI) | — | Personal access token for CLI PR creation |

*Not needed for Ollama.

---

## 10. Key Go Libraries

| Library | Purpose | Import Path |
|---------|---------|-------------|
| cobra | CLI framework | `github.com/spf13/cobra` |
| go-github | GitHub API client | `github.com/google/go-github/v62` |
| go-yaml | YAML parsing/writing | `gopkg.in/yaml.v3` |
| jsonschema | JSON Schema validation | `github.com/santhosh-tekuri/jsonschema/v5` |
| jwt | GitHub App JWT auth | `github.com/golang-jwt/jwt/v5` |
| slog | Structured logging | `log/slog` (stdlib) |

---

## 11. Development Workflow

```bash
# Prerequisites: Go 1.22+, Node.js 20+, Docker

# First-time setup
make setup                           # Copies .env.example, clones OSS repo

# CLI development
go run ./cmd/oss validate            # Run validator against local OSS clone
go run ./cmd/oss generate teaching-notes cambridge/igcse/.../05-quadratic-equations

# Bot development
npx smee -u https://smee.io/your-channel -p 8090   # Forward webhooks
go run ./cmd/bot                                     # Start webhook handler

# Web portal development
cd web && npm install && npm run dev  # Start at http://localhost:3001

# Run tests
make test                            # All Go tests
make lint                            # golangci-lint

# Build CLI binary
make build-cli                       # Output: ./bin/oss

# Build Docker image
make docker                          # Build multi-stage image
```

---

## 12. Performance Considerations

| Metric | Target | Approach |
|--------|--------|----------|
| AI generation latency | <15s for teaching notes, <10s for assessments | Streaming responses. Context limited to ~8K tokens to keep within fast completion range. |
| Validation latency | <500ms per file | In-process JSON Schema validation (no subprocess). Schema compiled once at startup. |
| PDF import | <60s for a 50-page PDF | Parallel page extraction. Streaming AI processing per section. |
| Concurrent webhook handling | 50 simultaneous events | Go goroutines. Each webhook processed independently. |
| CLI startup time | <100ms | Single Go binary, no runtime. |

---

## 13. Related Repositories

| Repository | Relationship |
|-----------|-------------|
| [p-n-ai/oss](https://github.com/p-n-ai/oss) | Target repository. OSS Bot creates PRs against this repo. Reads content for context building. |
| [p-n-ai/pai-bot](https://github.com/p-n-ai/pai-bot) | Calls OSS Bot's feedback API to submit data-driven improvement suggestions from student interactions. Shares AI provider interface code. |
