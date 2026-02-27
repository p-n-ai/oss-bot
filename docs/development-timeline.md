# oss-bot — Daily Development Timeline

> **Repository:** `p-n-ai/oss-bot`
> **Focus:** Contribution tooling for KSSM Matematik content
> **Duration:** 6 weeks (starts late — main build in Weeks 4-6)

---

## Scope for oss-bot

oss-bot owns **contribution tooling**: the CLI validator, AI content generation pipeline, GitHub bot (@oss-bot), and the web contribution portal (contribute.opensyllabus.org). It's the bridge between human contributors and the oss data repository.

**Key insight:** oss-bot is NOT needed for Weeks 1-3. During the validation phase, content is created manually by the Education Lead + AI assistance. oss-bot ships when the system needs to scale contribution beyond the core team — that's Week 4 onward.

**Build order:**
1. CLI validator (Week 4) — validates content locally before committing
2. AI content generation pipeline (Week 4-5) — generates teaching notes, assessments, examples from prompts
3. GitHub bot (Week 5-6) — automates PR creation from issue comments
4. Web portal (Week 6) — teacher-friendly contribution form

---

## WEEKS 1-3 — NO oss-bot WORK

oss-bot repo does not exist yet. All curriculum content is created directly in the oss repo by the Education Lead using AI assistance (Claude/ChatGPT) + manual editing.

**Why wait:** Building tooling before validating the content format is premature. If the schema changes based on student feedback (it will), the tooling would need to be rewritten. Let the content stabilize first.

---

## WEEK 4 — CLI TOOL + AI GENERATION PIPELINE

### Day 16 (Mon) — Initialize Repo + CLI Scaffold

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W4D16-1` | Initialize Go 1.22 repo: `cmd/oss/main.go` (CLI), `cmd/bot/main.go` (server), `internal/{ai,generator,validator,parser,github,api}` | 🤖 |
| `B-W4D16-2` | CLI scaffold using `cobra`: root command + `validate`, `generate`, `quality` subcommands | 🤖 |
| `B-W4D16-3` | `internal/validator/validator.go` — load JSON Schema from oss repo, compile at startup, validate YAML files in-process (no external deps) | 🤖 |
| `B-W4D16-4` | `oss validate [path]` command: validate all YAML in a directory tree against oss schemas. Exit code 0/1. Colored output showing pass/fail per file. | 🤖 |

### Day 17 (Tue) — Validation Tools

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W4D17-1` | `internal/validator/bloom.go` — verify bloom_levels match assessment question verbs | 🤖 |
| `B-W4D17-2` | `internal/validator/prerequisites.go` — check prerequisite graph for cycles across all KSSM forms | 🤖 |
| `B-W4D17-3` | `internal/validator/duplicates.go` — flag >85% similar assessment questions (cosine similarity on tokenized text) | 🤖 |
| `B-W4D17-4` | `internal/validator/quality.go` — auto-assess quality level (0-5) based on present fields | 🤖 |
| `B-W4D17-5` | `oss quality [path]` command: print quality report for all topics | 🤖 |

### Day 18 (Wed) — AI Content Generation Pipeline

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W4D18-1` | `internal/ai/` — same Provider interface as pai-bot (OpenAI, Anthropic, Ollama) | 🤖 |
| `B-W4D18-2` | `internal/generator/context.go` — Context Builder: load target topic + parent subject + syllabus + prerequisites + siblings + schema rules. Build ~8K token context. | 🤖 |
| `B-W4D18-3` | Create `prompts/teaching_notes.md` — template with variables: {{topic}}, {{subject}}, {{prerequisites}}, {{style_examples}}. Encodes pedagogical best practices for KSSM. | 🤖 |
| `B-W4D18-4` | Create `prompts/assessments.md` — template for generating quiz questions with rubrics, hints, distractors, Bloom's levels, KSSM exam format | 🤖 |
| `B-W4D18-5` | 🧑 Review and heavily edit prompt templates — these define teaching quality | 🧑 Education Lead (2hr) |

### Day 19 (Thu) — Generation Commands

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W4D19-1` | `internal/generator/teaching_notes.go` — generate teaching notes for a topic: build context → inject into template → call AI → parse markdown → validate → write file | 🤖 |
| `B-W4D19-2` | `internal/generator/assessments.go` — generate N assessment questions: build context → inject → call AI → parse YAML → validate against schema → retry if invalid → write file | 🤖 |
| `B-W4D19-3` | `internal/generator/examples.go` — generate worked examples with step-by-step solutions | 🤖 |
| `B-W4D19-4` | `oss generate teaching-notes --topic F2-01 --syllabus kssm-tingkatan2` — generate and write to correct path | 🤖 |
| `B-W4D19-5` | `oss generate assessments --topic F2-01 --count 5 --difficulty medium` — generate N questions | 🤖 |

### Day 20 (Fri) — Translation + Testing

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W4D20-1` | Create `prompts/translation.md` — preserve YAML structure exactly, translate only human-readable fields, use correct BM mathematical terminology | 🤖 |
| `B-W4D20-2` | `internal/generator/translator.go` + `oss translate --topic F1-01 --to ms` — generates locale file | 🤖 |
| `B-W4D20-3` | Test full pipeline: generate teaching notes + assessments + examples + translation for 1 Form 2 topic. Compare AI-generated vs manually-written quality. | 🤖🧑 |
| `B-W4D20-4` | 🧑 Education Lead evaluates: is AI-generated content quality acceptable with light editing? What needs to improve in prompts? | 🧑 Education Lead |

**Week 4 Output:** Working CLI with validate, generate, quality, translate commands. Prompt templates for KSSM content.

---

## WEEK 5 — GITHUB BOT + PDF IMPORT

### Day 21 (Mon) — GitHub App Setup

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W5D21-1` | `internal/github/app.go` — GitHub App authentication (JWT from private key, installation token exchange) | 🤖 |
| `B-W5D21-2` | `internal/github/webhook.go` — webhook handler: verify HMAC signature, parse issue_comment events, extract @oss-bot commands | 🤖 |
| `B-W5D21-3` | `internal/parser/command.go` — parse bot commands: `add teaching notes`, `add N assessments`, `translate`, `scaffold`, `quality` | 🤖 |
| `B-W5D21-4` | 🧑 Register GitHub App: p-n-ai org, webhook URL, permissions (Issues R/W, PRs R/W, Contents R/W) | 🧑 Human |

### Day 22 (Tue) — Bot → PR Pipeline

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W5D22-1` | `internal/github/pr.go` — create branch, commit files, open PR with labels and description | 🤖 |
| `B-W5D22-2` | `internal/github/contents.go` — read existing topic files from oss repo via GitHub Contents API | 🤖 |
| `B-W5D22-3` | Bot command flow: `@oss-bot add teaching notes for F2-01` → load topic from GitHub → run generation pipeline → create branch → commit files → open PR with provenance:ai-generated label | 🤖 |
| `B-W5D22-4` | Bot responds to issue with PR link: "I've generated teaching notes for F2-01 and opened #PR. Please review for accuracy." | 🤖 |

### Day 23 (Wed) — PDF Import + More Commands

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W5D23-1` | Create `prompts/pdf_import.md` — extract curriculum structure from PDF, infer Bloom's levels from verbs, map prerequisites | 🤖 |
| `B-W5D23-2` | `internal/parser/pdf.go` — PDF text extraction (ledongthuc/pdf) | 🤖 |
| `B-W5D23-3` | `internal/generator/scaffolder.go` — `oss import --pdf ./kssm-spec.pdf --board malaysia --level form3` → generate full syllabus scaffold | 🤖 |
| `B-W5D23-4` | `@oss-bot quality` command — responds with quality report for the topic in the issue | 🤖 |

### Day 24 (Thu) — Contribution Parser + Feedback API

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W5D24-1` | Create `prompts/contribution_parser.md` — parse natural language teacher input into structured YAML, preserve teacher's voice | 🤖 |
| `B-W5D24-2` | `internal/parser/contribution.go` — teacher writes "My students always confuse the negative sign when expanding brackets" → structured misconception entry | 🤖 |
| `B-W5D24-3` | `POST /api/feedback` — endpoint for pai-bot to submit observed patterns (misconception frequency, explanation effectiveness) | 🤖 |
| `B-W5D24-4` | Feedback handler: receive structured feedback → run generation pipeline → create PR with provenance:ai-observed label | 🤖 |

### Day 25 (Fri) — Docker + Testing

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W5D25-1` | Dockerfile: multi-stage Go build for both CLI binary and bot server | 🤖 |
| `B-W5D25-2` | `docker-compose.yml`: bot server + webhook tunnel (for dev) | 🤖 |
| `B-W5D25-3` | README.md: CLI installation (go install + pre-built binaries), GitHub App setup, bot deployment | 🤖 |
| `B-W5D25-4` | Test end-to-end: create GitHub issue → comment @oss-bot add teaching notes for F3-02 → verify PR is created with valid content | 🤖🧑 |
| `B-W5D25-5` | 🧑 Education Lead reviews 3 AI-generated PRs: would you approve these? What needs improvement? | 🧑 Education Lead |

**Week 5 Output:** Working GitHub bot that generates content and opens PRs. CLI with validate/generate/translate/import. Feedback API for pai-bot.

---

## WEEK 6 — WEB PORTAL + LAUNCH

### Day 26 (Mon) — Web Portal Scaffold

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W6D26-1` | Scaffold `web/`: Next.js 14 + TypeScript + shadcn/ui + Tailwind | 🤖 |
| `B-W6D26-2` | Contribution form: Select form (F1/F2/F3) → Select topic → Contribution type (teaching notes/example/assessment/correction/translation) → Content textarea | 🤖 |
| `B-W6D26-3` | `POST /api/preview` — AI structures the natural language input into proper YAML/markdown, returns preview | 🤖 |

### Day 27 (Tue) — Submit + Preview Flow

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W6D27-1` | Preview component: show structured output with syntax highlighting, diff against existing content | 🤖 |
| `B-W6D27-2` | `POST /api/submit` — on confirmation, create GitHub PR with attribution to the contributor | 🤖 |
| `B-W6D27-3` | Real-time schema validation in preview: show green checkmarks for valid fields, red for issues | 🤖 |

### Day 28 (Wed) — Curricula Browser

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W6D28-1` | `GET /api/curricula` — list all syllabi, subjects, topics from the oss repo | 🤖 |
| `B-W6D28-2` | Browse page: tree view of KSSM → Form 1/2/3 → Subject → Topic. Quality level badges. "Contribute" button per topic. | 🤖 |
| `B-W6D28-3` | Topic detail page: show existing content (teaching notes, examples, assessments) with "Improve this" buttons | 🤖 |

### Day 29 (Thu) — Deploy + Documentation

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W6D29-1` | Deploy bot + web portal: Docker on VPS, configure GitHub App webhook URL | 🤖 |
| `B-W6D29-2` | CONTRIBUTING.md: 3 ways to contribute (web form, @oss-bot, CLI), screenshot walkthrough | 🤖 |
| `B-W6D29-3` | 🧑 Test web portal with 2 teachers: can they contribute without knowing Git? | 🧑 Education Lead |

### Day 30 (Fri) — Launch + Report

| Task ID | Task | Owner |
|---------|------|-------|
| `B-W6D30-1` | 🧑 Announce web portal in launch materials: "contribute.opensyllabus.org — teachers can contribute without Git" | 🧑 Human |
| `B-W6D30-2` | 🧑 Write oss-bot section of 6-week report: AI generation quality, bot PRs created, web portal usage | 🧑 Human |

**Week 6 Output:** Web portal live at contribute.opensyllabus.org. GitHub bot responding to @oss-bot. CLI distributed as pre-built binary.

---

## Task Count Summary

| Week | 🤖 Claude Code | 🧑 Human | Total |
|------|----------------|----------|-------|
| 1-3 | 0 | 0 | 0 (no oss-bot work) |
| 4 | 16 | 2 | 18 |
| 5 | 14 | 2 | 16 |
| 6 | 10 | 2 | 12 |
| **Total** | **40** | **6** | **46** |

---

## Performance Targets

| Operation | Target |
|-----------|--------|
| `oss validate` (full repo) | <2s |
| Teaching notes generation | <15s |
| Assessment generation (5 questions) | <10s |
| PDF import (50-page syllabus) | <60s |
| Bot webhook → PR created | <30s |
| Web portal preview | <5s |
| CLI startup | <100ms |
