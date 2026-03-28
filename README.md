<p align="center">
  <h1 align="center">OSS Bot</h1>
  <p align="center">
    <strong>AI-powered tools to build, enrich, and maintain the Open School Syllabus</strong>
  </p>
  <p align="center">
    GitHub Bot · CLI Tools · Web Contribution Portal
  </p>
  <p align="center">
    <a href="#github-bot">GitHub Bot</a> ·
    <a href="#cli-tools">CLI Tools</a> ·
    <a href="#web-portal">Web Portal</a> ·
    <a href="#how-it-works">How It Works</a>
  </p>
  <p align="center">
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License"></a>
    <img src="https://img.shields.io/badge/go-%3E%3D1.22-00ADD8.svg" alt="Go Version">
  </p>
</p>

---

## What is OSS Bot?

OSS Bot is the tooling layer for [Open School Syllabus](https://github.com/p-n-ai/oss) — an open, structured curriculum for schools around the world. Whether it's Cambridge IGCSE, Malaysia's KSSM, India's CBSE, or any national or international syllabus, OSS Bot provides three ways to contribute curriculum content without needing to manually write YAML:

1. **GitHub Bot** — Comment on issues/PRs in the OSS repo and the bot generates content
2. **CLI Tools** — Import curricula from PDFs, generate assessments, translate topics
3. **Web Portal** — Teachers contribute in plain language through a simple form

All tools validate generated content against OSS's JSON Schemas before submission. All contributions go through GitHub's standard PR review process — the bot creates, humans approve.

---

## GitHub Bot

Install the bot on the [p-n-ai/oss](https://github.com/p-n-ai/oss) repository. Mention `@oss-bot` in any issue or PR comment.

### Commands

#### Generate Teaching Notes

```
@oss-bot add teaching notes for cambridge/igcse/mathematics-0580/topics/algebra/05-quadratic-equations
```

The bot reads the topic YAML, generates pedagogically-sound teaching notes, and opens a PR.

#### Generate Assessments

```
@oss-bot add 5 assessments for cambridge/igcse/mathematics-0580/topics/algebra/03-simultaneous-equations difficulty:medium
```

Generates assessment questions with worked solutions, rubrics, hints, and distractors. Opens a PR.

#### Translate a Topic

```
@oss-bot translate cambridge/igcse/mathematics-0580/topics/algebra/01-expressions to ms
```

Generates a Malay translation of the topic, matching the source structure exactly. Opens a PR tagged `needs-native-review`.

#### Scaffold a New Syllabus

```
@oss-bot scaffold syllabus india/cbse/mathematics-class10
```

Creates the directory structure, syllabus.yaml, subject files, and topic stubs (Level 0) based on publicly available curriculum information. Opens a PR with a completeness checklist.

#### Import from Document

```
@oss-bot import [attached file]
```

Attach a curriculum document (PDF, DOCX, PPTX, XLSX, or HTML) to the issue. The bot extracts the structure using Apache Tika, maps it to OSS schema, infers Bloom's taxonomy levels from specification verbs, and creates topic stubs. Opens a PR tagged `needs-educator-review`.

#### Check Topic Quality

```
@oss-bot quality cambridge/igcse/mathematics-0580/topics/algebra/05-quadratic-equations
```

The bot reads the topic YAML and reports the current quality level (0–5) along with which fields are present or missing.

#### Enrich from Classroom Experience

```
@oss-bot enrich cambridge/igcse/mathematics-0580/topics/algebra/05-quadratic-equations

I've taught this topic for 8 years. Students consistently struggle with negative
factors — they can factorise x²+5x+6 but fail on x²-x-6 because they don't
think to try negative numbers. I always start with a number puzzle: "Find two
numbers that multiply to -6 and add to -1." This clicks for about 80% of students.

For the quadratic formula, I teach them to write out a=-1, b=5, c=-6 explicitly
before substituting. This reduces sign errors by half.
```

The bot parses the natural language into structured contributions: misconception entries, teaching sequence updates, engagement hooks. Opens a PR crediting the teacher.

### Bot Behavior

- All generated content includes `provenance: ai-generated` or `provenance: ai-assisted`
- All PRs require human review before merging
- The bot validates all YAML against JSON Schemas — invalid content is never submitted
- PRs include a self-assessment: estimated quality level and what's missing for the next level
- Translation PRs are auto-tagged `needs-native-review`

---

## CLI Tools

### Installation

```bash
# From source
go install github.com/p-n-ai/oss-bot/cmd/oss@latest

# Or download the binary
curl -sSL https://github.com/p-n-ai/oss-bot/releases/latest/download/oss-$(uname -s)-$(uname -m) -o /usr/local/bin/oss
chmod +x /usr/local/bin/oss
```

### Configuration

```bash
# Required — AI provider for content generation
export OSS_AI_PROVIDER=openai          # openai | anthropic | ollama
export OSS_AI_API_KEY=sk-...           # Not needed for Ollama
export OSS_REPO_PATH=./oss             # Path to your local OSS clone

# Optional — reasoning model for bulk import and content merge (recommended)
export OSS_AI_REASONING_API_KEY=sk-or-...           # OpenRouter API key
export OSS_AI_REASONING_MODEL=deepseek/deepseek-r1  # default; also: moonshotai/kimi-k2.5, qwen/qwen3.5, openai/o3-mini

# Optional — for creating PRs
export OSS_GITHUB_TOKEN=ghp_...
```

### Commands

#### Validate

Check all YAML files against the schemas:

```bash
# Validate entire repository
oss validate

# Validate a specific syllabus
oss validate --syllabus cambridge/igcse/mathematics-0580

# Validate a single file
oss validate --file curricula/cambridge/igcse/mathematics-0580/topics/algebra/05-quadratic-equations.yaml
```

Output:
```
✅ syllabus.yaml — valid
✅ subjects/algebra.yaml — valid
✅ topics/algebra/01-expressions.yaml — valid
❌ topics/algebra/05-quadratic-equations.yaml — missing required field: mastery.minimum_score
✅ topics/algebra/05-quadratic-equations.assessments.yaml — valid

4/5 files valid. 1 error.
```

#### Generate Content

```bash
# Generate teaching notes for a topic
oss generate teaching-notes cambridge/igcse/mathematics-0580/topics/algebra/05-quadratic-equations

# Generate assessment questions
oss generate assessments cambridge/igcse/mathematics-0580/topics/algebra/05-quadratic-equations --count 5 --difficulty medium

# Generate worked examples
oss generate examples cambridge/igcse/mathematics-0580/topics/algebra/05-quadratic-equations --count 3

# Generate all content types (teaching-notes, assessments, examples, topic enrichment) for every topic in a subject-grade
oss generate all --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-4

# With more parallel workers
oss generate all --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-4 --workers 5

# Dry run — list topics that would be processed without generating
oss generate all --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-4 --dry-run
```

Generated files are written to the local OSS clone. Review them, then commit and PR.

#### Batch Generate (Post-Import)

After importing topics from a PDF, generate all supporting content (teaching notes, assessments, examples, and topic enrichment) for every topic in one command:

```bash
oss generate all --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-4
```

**Flags**

| Flag | Default | Description |
|------|---------|-------------|
| `--syllabus` | *(required)* | Syllabus ID — disambiguates when the same subject-grade exists under multiple syllabuses |
| `--subject-grade` | *(required)* | Subject grade ID, e.g. `malaysia-kssm-matematik-tingkatan-4` |
| `--workers` | `3` | Number of parallel AI workers |
| `--dry-run` | `false` | List discovered topics without generating anything |

**Typical workflow:**

```bash
# 1. Scaffold the subject directory
oss scaffold subject \
  --syllabus malaysia-kssm \
  --id malaysia-kssm-matematik \
  --grade-id malaysia-kssm-matematik-tingkatan-4 \
  --country malaysia

# 2. Import topic structure from the curriculum PDF
oss import --pdf DSKP-KSSM-Matematik-Tingkatan-4.pdf \
           --syllabus malaysia-kssm \
           --subject-grade malaysia-kssm-matematik-tingkatan-4

# 3. Generate teaching notes, assessments, examples, and enrich topic YAML for all imported topics
oss generate all --syllabus malaysia-kssm \
                 --subject-grade malaysia-kssm-matematik-tingkatan-4
```

The command discovers all `.yaml` topic files in the subject-grade's `topics/` directory, extracts their IDs, and runs the generation pipeline for each topic in parallel. For each topic, it generates 4 content types:
1. **Teaching notes** (`.teaching.md`) — pedagogical guide for educators
2. **Assessments** (`.assessments.yaml`) — questions with rubrics and hints
3. **Examples** (`.examples.yaml`) — worked examples at varying difficulty
4. **Topic enrichment** — adds structured Level 2 fields (`teaching.sequence`, `teaching.common_misconceptions`, `engagement_hooks`) into the topic YAML to advance quality from Level 1 to Level 3

Progress is reported as each generation completes.

#### Import from PDF

Extract curriculum topics from a PDF and generate structured OSS YAML files.

```bash
oss import --pdf <file> --syllabus <id> [flags]
```

**Flags**

| Flag | Default | Description |
|------|---------|-------------|
| `--pdf` | *(required)* | Path to the PDF file |
| `--syllabus` | *(required)* | Target syllabus ID, e.g. `malaysia-kssm` |
| `--subject-grade` | `""` | Subject grade ID for correct file naming, e.g. `malaysia-kssm-matematik-tingkatan-4`. When set, output files follow the OSS topic ID convention (`MT4-01.yaml`, `PH12-03.yaml`, etc.) and are placed under `{subject_id}/{subject_grade_id}/topics/` |
| `--workers` | `3` | Number of parallel AI workers — each processes one topic concurrently |
| `--chunk-size` | `2000` | Max tokens per chunk for the generic chunker. Lower values force more splits on dense documents |
| `--force` | `false` | **Replace** existing topic files outright. Default (`false`) AI-merges new content into the existing file without losing any objectives |
| `--pr` | `false` | Create a GitHub PR instead of writing files to the local filesystem |

**Examples**

```bash
# Typical workflow: scaffold the subject first, then import
oss scaffold subject \
  --syllabus malaysia-kssm \
  --id malaysia-kssm-matematik \
  --grade-id malaysia-kssm-matematik-tingkatan-4 \
  --country malaysia

oss import --pdf DSKP-KSSM-Matematik-Tingkatan-4.pdf \
           --syllabus malaysia-kssm \
           --subject-grade malaysia-kssm-matematik-tingkatan-4

# More workers for a large document
oss import --pdf textbook.pdf \
           --syllabus india-cbse \
           --subject-grade india-cbse-physics-class-12 \
           --workers 5

# Re-import from scratch — replace all existing files
oss import --pdf DSKP.pdf --syllabus malaysia-kssm \
           --subject-grade malaysia-kssm-matematik-tingkatan-4 --force

# Open a GitHub PR directly instead of writing to disk
oss import --pdf DSKP.pdf --syllabus malaysia-kssm \
           --subject-grade malaysia-kssm-matematik-tingkatan-4 --pr
```

**What it does**

1. Extracts text from the PDF using Go-native `ledongthuc/pdf` (no external dependencies)
2. **DSKP auto-detection** — if `BIDANG PEMBELAJARAN` / `TAJUK` markers are found (Malaysian KSSM curriculum documents), splits exactly on those boundaries for per-chapter accuracy; otherwise falls back to generic heading-based chunking
3. Processes each topic in parallel using the configured number of workers
4. Names output files using the [OSS ID convention](docs/id-conventions.md) — e.g. `MT4-01.yaml` for Matematik Tingkatan 4 Chapter 1, `PH12-03.yaml` for Physics Class 12 Chapter 3
5. Each generated file includes: `id`, `official_ref`, `name` (MOE language), `name_en` (English translation), `subject_grade_id`, `subject_id`, `syllabus_id`, `country_id`, `language`, `difficulty`, `tier`, `learning_objectives` (with SP codes and Bloom's levels in English), `prerequisites`, `mastery`, `provenance: ai-assisted`
6. **Existing file handling:**
   - `--force` not set (default): AI compares the existing file with the newly extracted content and produces a single merged YAML — new objectives are added, duplicates are skipped, identity fields (`id`, `subject_id`, etc.) are preserved
   - `--force` set: existing file is overwritten with the freshly generated content

**Using a reasoning model (recommended)**

Set `OSS_AI_REASONING_API_KEY` to route both extraction and merge calls through a reasoning model via [OpenRouter](https://openrouter.ai). This produces significantly better Bloom's level inference and objective extraction for complex documents.

```bash
export OSS_AI_REASONING_API_KEY=sk-or-...
export OSS_AI_REASONING_MODEL=deepseek/deepseek-r1   # optional — this is the default

oss import --pdf DSKP-KSSM-Matematik-Tingkatan-4.pdf \
           --syllabus malaysia-kssm \
           --subject-grade malaysia-kssm-matematik-tingkatan-4
```

Expected output:
```
Using reasoning model: deepseek/deepseek-r1
Extracting text from DSKP-KSSM-Matematik-Tingkatan-4.pdf...
Extracted 30112 characters
Detected DSKP format: 10 topics (BIDANG PEMBELAJARAN/TAJUK structure)
Split into 10 chunks
Starting bulk import of 10 items
  [1/10] 1.0 FUNGSI DAN PERSAMAAN KUADRATIK: done
  [2/10] 2.0 POLINOMIAL: done
  ...
  wrote: .../topics/MT4-01.yaml
  wrote: .../topics/MT4-02.yaml
  ...
Processed 10/10 chunks in 45s — wrote 10 new, merged 0 existing file(s)
```

When `OSS_AI_REASONING_API_KEY` is not set, the standard `OSS_AI_PROVIDER` is used as a transparent fallback.

#### Translate

```bash
# Translate all topics in a subject-grade to English
oss translate --syllabus malaysia-kssm \
              --subject-grade malaysia-kssm-matematik-tingkatan-5 \
              --to en

# With more parallel workers
oss translate --syllabus malaysia-kssm \
              --subject-grade malaysia-kssm-matematik-tingkatan-5 \
              --to en --workers 5

# Translate a single topic
oss translate --topic MT5-01 --to en
```

**Flags**

| Flag | Default | Description |
|------|---------|-------------|
| `--topic` | | Topic ID for single-topic translation |
| `--syllabus` | | Syllabus ID for batch translation (e.g. `malaysia-kssm`) |
| `--subject-grade` | | Subject grade ID for batch translation (e.g. `malaysia-kssm-matematik-tingkatan-5`) |
| `--to` | *(required)* | Target language code: `en`, `ms`, `zh`, `ta` |
| `--workers` | `3` | Number of parallel workers (batch mode only) |

Provide either `--topic` for a single topic, or `--syllabus` and `--subject-grade` for batch translation. Translations are written into each topic YAML under the `translations` field.

#### Analyze Quality

```bash
# Show quality overview by path
oss quality /path/to/topics

# Show quality overview by flags
oss quality --syllabus malaysia-kssm --subject-grade malaysia-kssm-matematik-tingkatan-5
```

**Flags**

| Flag | Default | Description |
|------|---------|-------------|
| `[path]` | `$OSS_REPO_PATH` or `.` | Positional argument — directory to scan |
| `--syllabus` | | Syllabus ID (e.g. `malaysia-kssm`) |
| `--subject-grade` | | Subject grade ID (e.g. `malaysia-kssm-matematik-tingkatan-5`) |

Provide either a positional path or `--syllabus`/`--subject-grade` flags (the flags resolve the topics directory automatically).

Output:
```
=== Quality Level Report ===
Level 5 (Gold): 0 topics
Level 4 (Complete): 0 topics
Level 3 (Teachable): 5 topics
Level 2 (Structured): 2 topics
Level 1 (Basic): 1 topics
Level 0 (Stub): 0 topics

⚠️  Overclaimed quality levels:
  MT5-01: claims Level 1, actual Level 0
```

#### Scaffold a New Curriculum

```bash
# Create a new syllabus for any country
oss scaffold syllabus --id india-jee --country india

# Create a new subject + subject_grade within a syllabus
oss scaffold subject \
  --syllabus malaysia-kssm \
  --id malaysia-kssm-matematik \
  --grade-id malaysia-kssm-matematik-tingkatan-3 \
  --country malaysia
```

The `scaffold subject` command creates the three-level directory structure:

```
curricula/malaysia/malaysia-kssm/
└── malaysia-kssm-matematik/                         # subject (grade-less)
    ├── subject.yaml                                 # id: malaysia-kssm-matematik
    └── malaysia-kssm-matematik-tingkatan-3/         # subject_grade (with grade)
        ├── subject-grade.yaml                    # id: malaysia-kssm-matematik-tingkatan-3
        └── topics/
            ├── MT3-01.yaml
            ├── MT3-02.yaml
            └── ...
```

**Flags**

| Flag | Required | Description |
|------|----------|-------------|
| `--syllabus` | Yes | Syllabus ID, e.g. `malaysia-kssm` |
| `--id` | Yes | Subject ID — grade-less, e.g. `malaysia-kssm-matematik` |
| `--grade-id` | No | Subject grade ID — with grade, e.g. `malaysia-kssm-matematik-tingkatan-3`. If omitted, defaults to `--id` (for grade-less syllabi like IGCSE) |
| `--country` | No | Country code, e.g. `malaysia` |
| `--from-file` | No | Path to subject document for AI-assisted topic extraction |

See [ID conventions](docs/id-conventions.md) for the full naming rules.

#### Contribute via CLI

```bash
# Add a contribution in natural language
oss contribute "I teach IGCSE Math. For quadratic equations, I've found that starting with the discriminant helps students understand WHY some equations have no solution before they learn the formula. About 60% of my students grasp the concept faster this way."

# The CLI will:
# 1. Identify the relevant topic (quadratic equations in IGCSE)
# 2. Parse the insight into structured data (teaching sequence suggestion)
# 3. Show a preview of the proposed change
# 4. On confirmation, commit to a branch and open a PR
```

---

## Web Portal

**contribute.p-n-ai.org** — for teachers who don't use Git or the command line.

### How It Works

1. **Select** — Choose a syllabus and topic (or "Add new syllabus")
2. **Contribute** — Type in plain language. Examples:
   - "Here are 3 common mistakes students make when solving simultaneous equations..."
   - "I'd suggest teaching this topic by starting with a real-world example about..."
   - "Here are 5 practice questions I use in my classroom..."
3. **Preview** — The AI structures your input into proper YAML/Markdown and shows a preview
4. **Submit** — On confirmation, a PR is opened on the OSS repo on your behalf
5. **Track** — Get a link to your PR so you can follow the review process

### Features

- No account required (PR is created by the bot, with attribution to you)
- Supports plain language input in any language
- Shows a real-time preview of the structured output
- Validates against schemas before submission
- Tracks your contributions across sessions (optional sign-in via GitHub)

---

## How It Works

### AI Content Generation Pipeline

```
Input (natural language, document [PDF/DOCX/PPTX/XLSX/HTML], or structured command)
    │
    ▼
┌─────────────────────────┐
│  Document Extraction    │  (CLI: Go-native PDF | Server: Apache Tika)
└──────────┬──────────────┘
           │
           ▼
┌─────────────────────────┐
│  Context Builder        │
│  - Load topic YAML      │
│  - Load related topics  │
│  - Load existing content│
│  - Load schema rules    │
└──────────┬──────────────┘
           │
           ▼
┌─────────────────────────┐
│  AI Generation          │
│  - Pedagogical prompt   │
│  - Schema-aware output  │
│  - Style matching       │
└──────────┬──────────────┘
           │
           ▼
┌─────────────────────────┐
│  Content Merge          │
│  - Compare new vs       │
│    existing content     │
│  - Append/dedup         │
│    assessments          │
│  - Additive teaching    │
│    notes                │
└──────────┬──────────────┘
           │
           ▼
┌─────────────────────────┐
│  Progress Reporter      │
│  - CLI progress bar     │
│  - Bot comment updates  │
│  - Web SSE stream       │
└──────────┬──────────────┘
           │
           ▼
┌────────────────────────┐
│  Validation            │
│  - JSON Schema check   │
│  - Bloom's taxonomy    │
│  - Prerequisite graph  │
│  - Duplicate detection │
└──────────┬─────────────┘
           │
           ▼
┌────────────────────────┐
│  Output                │
│  - Write YAML/MD file  │
│  - Open GitHub PR      │
│  - Add quality labels  │
│  - Request reviewers   │
└────────────────────────┘
```

### AI Provider Support

OSS Bot uses the same AI providers as P&AI Bot:

| Provider | Best For | Setup |
|----------|---------|-------|
| OpenAI (GPT-4o) | General content generation | `OSS_AI_PROVIDER=openai` |
| Anthropic (Claude) | Teaching notes, nuanced pedagogy | `OSS_AI_PROVIDER=anthropic` |
| Ollama (Llama 3) | Free, self-hosted, privacy-sensitive | `OSS_AI_PROVIDER=ollama` |
| Reasoning (DeepSeek R1, Kimi K2.5, Qwen 3.5, o3-mini) | Complex analysis: bulk import, content merge | `OSS_AI_REASONING_PROVIDER=openrouter` (via [OpenRouter](https://openrouter.ai)) |

---

## Project Structure

```
oss-bot/
├── cmd/
│   ├── oss/                     # CLI entrypoint
│   │   └── main.go
│   └── bot/                     # GitHub Bot + Web Portal server
│       └── main.go
├── internal/
│   ├── ai/                      # AI provider interface (shared with P&AI Bot)
│   │   ├── provider.go
│   │   ├── openai.go
│   │   ├── anthropic.go
│   │   ├── ollama.go
│   │   └── reasoning.go            # Reasoning model provider
│   ├── generator/               # Content generation pipeline
│   │   ├── context.go           # Context builder
│   │   ├── teaching_notes.go
│   │   ├── assessments.go
│   │   ├── examples.go
│   │   ├── translator.go
│   │   ├── scaffolder.go        # New syllabus scaffolding
│   │   ├── importer.go          # Document import (PDF, DOCX, PPTX, HTML)
│   │   └── merge.go             # Content merge logic
│   ├── validator/               # Schema validation
│   │   ├── validator.go         # JSON Schema engine
│   │   ├── bloom.go             # Bloom's taxonomy checks
│   │   ├── prerequisites.go     # Prerequisite graph integrity
│   │   ├── duplicates.go        # Duplicate content detection
│   │   └── quality.go           # Quality level assessment
│   ├── pipeline/                # Shared orchestrator
│   │   ├── pipeline.go          # Execute(ctx, Request) → Result
│   │   ├── bulk.go              # Bulk import orchestrator
│   │   └── progress.go          # Progress reporting
│   ├── parser/                  # Input parsing + document extraction
│   │   ├── command.go           # Parse @oss-bot commands
│   │   ├── contribution.go      # Natural language → structured data
│   │   ├── document.go          # DocumentParser interface
│   │   ├── pdf.go               # Go-native PDF extraction (CLI)
│   │   ├── tika.go              # Apache Tika multi-format extraction (server)
│   │   └── chunker.go           # Large document chunking
│   ├── github/                  # GitHub API integration
│   │   ├── app.go               # GitHub App authentication
│   │   ├── webhook.go           # Webhook handler + HMAC verification
│   │   ├── pr.go                # PR creation, labels, reviewers
│   │   └── contents.go          # Read/write via GitHub Contents API
│   └── api/                     # Web portal backend
│       ├── router.go
│       ├── preview.go           # POST /api/preview
│       ├── submit.go            # POST /api/submit
│       └── curricula.go         # GET /api/curricula
├── web/                         # Contribution web portal (Next.js)
│   ├── src/
│   │   ├── app/                 # Next.js pages (App Router)
│   │   ├── components/          # UI components
│   │   └── lib/                 # API client
│   ├── package.json
│   └── next.config.js
├── prompts/                     # AI prompt templates
│   ├── teaching_notes.md
│   ├── assessments.md
│   ├── examples.md
│   ├── translation.md
│   ├── contribution_parser.md
│   ├── document_import.md       # Curriculum import (PDF, DOCX, PPTX, HTML)
│   ├── bulk_import.md           # Bulk import prompt template
│   └── content_merge.md         # Content merge prompt template
├── scripts/                     # Dev scripts
│   ├── setup.sh
│   └── test-webhook.sh
├── deploy/
│   └── docker/
│       ├── Dockerfile
│       └── Dockerfile.dev
├── docker-compose.yml
├── Makefile
├── .env.example
└── README.md
```

---

## Self-Hosting the Bot

If you fork OSS for your own curriculum project, you can run your own OSS Bot instance.

### Quick Start

```bash
git clone https://github.com/p-n-ai/oss-bot.git
cd oss-bot
cp .env.example .env
# Edit .env with your GitHub App credentials and AI API key
docker compose up -d    # Starts bot, web portal, and Apache Tika sidecar
```

### GitHub App Setup

1. Create a GitHub App at `github.com/settings/apps/new`
2. Set the webhook URL to your bot's endpoint (e.g., `https://your-bot.example.com/webhook`)
3. Permissions needed: Issues (read/write), Pull Requests (read/write), Contents (read/write)
4. Subscribe to events: Issue comment, Pull request
5. Install the app on your OSS fork
6. Add the App ID, private key, and webhook secret to `.env`

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `OSS_GITHUB_APP_ID` | Yes | GitHub App ID |
| `OSS_GITHUB_PRIVATE_KEY_PATH` | Yes | Path to the GitHub App private key |
| `OSS_GITHUB_WEBHOOK_SECRET` | Yes | Webhook secret for verifying GitHub events |
| `OSS_REPO_OWNER` | Yes | GitHub org/user (e.g., `p-n-ai`) |
| `OSS_REPO_NAME` | Yes | Repository name (e.g., `oss`) |
| `OSS_AI_PROVIDER` | Yes | `openai`, `anthropic`, or `ollama` |
| `OSS_AI_API_KEY` | No* | API key for the chosen provider |
| `OSS_AI_OLLAMA_URL` | No | Ollama base URL (default: `http://ollama:11434`) |
| `OSS_TIKA_URL` | No | Apache Tika server URL (default: `http://tika:9998`) |
| `OSS_AI_REASONING_PROVIDER` | No | Reasoning model provider (default: `openrouter`) |
| `OSS_AI_REASONING_API_KEY` | No | OpenRouter API key for reasoning models |
| `OSS_AI_REASONING_MODEL` | No | Model on OpenRouter (default: `deepseek/deepseek-r1`) |
| `OSS_WORKER_COUNT` | No | Parallel workers for bulk import (default: `3`) |
| `OSS_WEB_PORT` | No | Web portal port (default: `3001`) |

*Not needed for Ollama.

---

## Development

### Prerequisites

- Go 1.22+
- Node.js 20+ (for the web portal)
- Docker (for Apache Tika sidecar — required for multi-format document import)
- A local clone of [p-n-ai/oss](https://github.com/p-n-ai/oss)

### Local Development

```bash
# Run the CLI locally
go run ./cmd/oss validate --repo-path ../oss

# Run the web portal
cd web && npm install && npm run dev
```

### Running the Bot Locally

The bot requires two terminals — one for the webhook tunnel, one for the server.

**Get a smee.io channel** (one-time): visit [smee.io](https://smee.io) → "Start a new channel". Copy the URL — it stays the same across restarts.

**Terminal 1 — webhook tunnel** (forwards GitHub events to your local machine):
```bash
smee -u https://smee.io/<your-channel> -p 8090 --path /webhook
```

> `--path /webhook` is required. Without it, smee forwards to `/` and all events are silently dropped.

**Terminal 2 — bot server**:
```bash
set -a && source .env && set +a
go run ./cmd/bot
```

Set the smee URL as your GitHub App's **Webhook URL** in App settings. Then comment `@oss-bot add teaching notes for <topic-id>` on any issue in the target repo.

**Smoke-test without a live GitHub App:**
```bash
OSS_GITHUB_WEBHOOK_SECRET=<your-secret> ./scripts/test-webhook.sh
```

### Feedback API (pai-bot integration)

The bot exposes a `POST /api/feedback` endpoint for [pai-bot](https://github.com/p-n-ai/pai-bot) to submit observed learning patterns as curriculum contributions.

```bash
curl -X POST http://localhost:8090/api/feedback \
  -H "Content-Type: application/json" \
  -d '{
    "topic_path": "mathematics/algebra/03-simultaneous-equations",
    "content_type": "teaching_notes",
    "observation": "Students consistently struggle with the substitution step when coefficients are fractions.",
    "source": "pai-bot"
  }'
```

Generated content uses `provenance: ai-observed` and is reviewed before merging.

### Running Tests

```bash
go test ./...
```

---

## Related Repositories

| Repository | Description |
|-----------|-------------|
| [p-n-ai/oss](https://github.com/p-n-ai/oss) | The curriculum data repository that this bot operates on |
| [p-n-ai/pai-bot](https://github.com/p-n-ai/pai-bot) | AI learning companion that consumes OSS curriculum data |

---

## License

OSS Bot is licensed under the [Apache License 2.0](LICENSE).

---

<p align="center">
  <strong>Making it easy for anyone to contribute to the world's open curriculum.</strong>
  <br>
  A <a href="https://pandai.org">Pandai</a> initiative. Built with ❤️ by educators and AI, for everyone.
</p>
