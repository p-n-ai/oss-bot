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

OSS Bot is the tooling layer for [Open School Syllabus](https://github.com/p-n-ai/oss). It provides three ways to contribute curriculum content without needing to manually write YAML:

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

#### Import from PDF

```
@oss-bot import --pdf [attached PDF]
```

Attach a curriculum PDF to the issue. The bot extracts the structure, maps it to OSS schema, infers Bloom's taxonomy levels from specification verbs, and creates topic stubs. Opens a PR tagged `needs-educator-review`.

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
# Set your AI provider (the bot needs an AI model to generate content)
export OSS_AI_PROVIDER=openai          # openai | anthropic | ollama
export OSS_AI_API_KEY=sk-...           # Not needed for Ollama
export OSS_REPO_PATH=./oss             # Path to your local OSS clone
export OSS_GITHUB_TOKEN=ghp_...        # For creating PRs (optional)
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
```

Generated files are written to the local OSS clone. Review them, then commit and PR.

#### Import from PDF

```bash
# Import a curriculum from a PDF document
oss import --pdf ./cambridge-igcse-maths-2025-syllabus.pdf --board cambridge --level igcse --subject mathematics

# Import with more guidance
oss import --pdf ./kssm-form4.pdf --board malaysia --level kssm --subject mathematics --language ms
```

The importer:
1. Extracts text and structure from the PDF
2. Identifies subjects, topics, and learning objectives
3. Infers Bloom's taxonomy levels from specification verbs ("State" → remember, "Calculate" → apply, "Evaluate" → evaluate)
4. Maps prerequisite relationships between topics
5. Generates Level 0-1 topic stubs
6. Outputs to the correct directory structure

#### Translate

```bash
# Translate all topics in a syllabus to Malay
oss translate --syllabus cambridge/igcse/mathematics-0580 --to ms

# Translate a single topic
oss translate --topic cambridge/igcse/mathematics-0580/topics/algebra/05-quadratic-equations --to ar

# Translate with a specific model (better for less-common languages)
oss translate --topic ... --to hi --model claude-sonnet-4-20250514
```

Translations are placed in the `locales/{lang}/` directory, matching the source file structure exactly.

#### Analyze Quality

```bash
# Show quality overview for a syllabus
oss quality cambridge/igcse/mathematics-0580
```

Output:
```
Cambridge IGCSE Mathematics 0580 — Quality Report

Topics: 8 total
  ⭐⭐⭐⭐ Complete (4):   0  (0%)
  ⭐⭐⭐   Teachable (3):  5  (63%)
  ⭐⭐     Structured (2): 2  (25%)
  ⭐       Basic (1):      1  (12%)
  ⬜       Stub (0):       0  (0%)

Missing for next level:
  02-linear-equations (Level 2 → 3): needs teaching notes, assessments
  06-fractions (Level 1 → 2): needs teaching sequence, misconceptions

Translations: 1 language (ms) — 3/8 topics translated
Cross-curriculum links: 4/8 topics linked to universal concepts
```

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
Input (natural language, PDF, or structured command)
    │
    ▼
┌──────────────────────┐
│  Context Builder      │
│  - Load topic YAML    │
│  - Load related topics│
│  - Load existing content│
│  - Load schema rules  │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  AI Generation        │
│  - Pedagogical prompt │
│  - Schema-aware output│
│  - Style matching     │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  Validation           │
│  - JSON Schema check  │
│  - Bloom's taxonomy   │
│  - Prerequisite graph │
│  - Duplicate detection│
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  Output               │
│  - Write YAML/MD file │
│  - Open GitHub PR     │
│  - Add quality labels │
│  - Request reviewers  │
└──────────────────────┘
```

### AI Provider Support

OSS Bot uses the same AI providers as P&AI Bot:

| Provider | Best For | Setup |
|----------|---------|-------|
| OpenAI (GPT-4o) | General content generation | `OSS_AI_PROVIDER=openai` |
| Anthropic (Claude) | Teaching notes, nuanced pedagogy | `OSS_AI_PROVIDER=anthropic` |
| Ollama (Llama 3) | Free, self-hosted, privacy-sensitive | `OSS_AI_PROVIDER=ollama` |

---

## Project Structure

```
oss-bot/
├── cmd/
│   ├── oss/                     # CLI entrypoint
│   │   └── main.go
│   └── bot/                     # GitHub Bot entrypoint
│       └── main.go
├── internal/
│   ├── ai/                      # AI provider interface
│   │   ├── provider.go
│   │   ├── openai.go
│   │   ├── anthropic.go
│   │   └── ollama.go
│   ├── generator/               # Content generation
│   │   ├── teaching_notes.go
│   │   ├── assessments.go
│   │   ├── examples.go
│   │   ├── translator.go
│   │   └── importer.go          # PDF import
│   ├── validator/               # Schema validation
│   │   └── validator.go
│   ├── github/                  # GitHub API + webhook handling
│   │   ├── bot.go               # GitHub App webhook handler
│   │   └── pr.go                # PR creation helpers
│   └── parser/                  # Natural language → structured data
│       ├── contribution.go
│       └── pdf.go               # PDF text extraction
├── web/                         # Contribution web portal
│   ├── src/
│   │   └── app/                 # Next.js pages
│   ├── package.json
│   └── next.config.js
├── prompts/                     # AI prompt templates
│   ├── teaching_notes.md
│   ├── assessments.md
│   ├── translation.md
│   └── contribution_parser.md
├── deploy/
│   └── docker/
│       └── Dockerfile
├── docker-compose.yml
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
docker compose up -d
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
| `OSS_WEB_PORT` | No | Web portal port (default: `3001`) |

*Not needed for Ollama.

---

## Development

### Prerequisites

- Go 1.22+
- Node.js 20+ (for the web portal)
- A local clone of [p-n-ai/oss](https://github.com/p-n-ai/oss)

### Local Development

```bash
# Run the CLI locally
go run ./cmd/oss validate --repo-path ../oss

# Run the GitHub bot locally (with smee.io for webhook forwarding)
npx smee -u https://smee.io/your-channel -p 8090
go run ./cmd/bot

# Run the web portal
cd web && npm install && npm run dev
```

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
  A <a href="https://pandai.app">Pandai</a> initiative.
</p>
