# oss-bot — Daily Development Timeline

> **Repository:** `p-n-ai/oss-bot`
> **Focus:** Contribution tooling for KSSM Matematik content
> **Duration:** 6 weeks (starts late — main build in Weeks 4-6)

---

## Scope for oss-bot

oss-bot owns **contribution tooling**: the CLI validator, AI content generation pipeline, GitHub bot (@oss-bot), and the web contribution portal (contribute.p-n-ai.org). It's the bridge between human contributors and the oss data repository.

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

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W4D16-1` | Initialize Go 1.22 repo: `cmd/oss/main.go` (CLI), `cmd/bot/main.go` (server), `internal/{ai,generator,validator,parser,github,api}` | 🤖 | ✅ | Go 1.26 used; also created pipeline/ and output/ dirs |
| `B-W4D16-2` | CLI scaffold using `cobra`: root command + `validate`, `generate`, `quality` subcommands | 🤖 | ✅ | + translate subcommand |
| `B-W4D16-3` | `internal/validator/validator.go` — load JSON Schema from oss repo, compile at startup, validate YAML files in-process (no external deps) | 🤖 | ✅ | TDD: 5 tests passing |
| `B-W4D16-4` | `oss validate [path]` command: validate all YAML in a directory tree against oss schemas. Exit code 0/1. Colored output showing pass/fail per file. | 🤖 | ✅ | + Makefile, .env.example, CI workflow |

### Day 17 (Tue) — Validation Tools

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W4D17-1` | `internal/validator/bloom.go` — verify bloom_levels match assessment question verbs | 🤖 | ✅ | TDD: 12 tests (verb lookup + consistency check) |
| `B-W4D17-2` | `internal/validator/prerequisites.go` — check prerequisite graph for cycles across all KSSM forms | 🤖 | ✅ | TDD: DFS cycle detection + missing prereq detection |
| `B-W4D17-3` | `internal/validator/duplicates.go` — flag >85% similar assessment questions (cosine similarity on tokenized text) | 🤖 | ✅ | TDD: tokenize, cosine similarity, duplicate pairs |
| `B-W4D17-4` | `internal/validator/quality.go` — auto-assess quality level (0-5) based on present fields | 🤖 | ✅ | TDD: levels 0-3 tested + TopicInfoFromYAML parser |
| `B-W4D17-5` | `oss quality [path]` command: print quality report for all topics | 🤖 | ✅ | Wired with directory walk + quality report output |

### Day 18 (Wed) — AI Content Generation Pipeline

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W4D18-1` | `internal/ai/` — same Provider interface as pai-bot (OpenAI, Anthropic, Ollama) | 🤖 | ✅ | TDD: provider.go, mock.go, openai.go, anthropic.go, ollama.go |
| `B-W4D18-2` | `internal/generator/context.go` — Context Builder: load target topic + parent subject + syllabus + prerequisites + siblings + schema rules. Build ~8K token context. | 🤖 | ✅ | TDD: 3 tests passing. Topic struct updated with tier, mastery fields from real OSS content |
| `B-W4D18-3` | Create `prompts/teaching_notes.md` — template with variables: {{topic}}, {{subject}}, {{prerequisites}}, {{style_examples}}. Encodes pedagogical best practices for KSSM. | 🤖 | ✅ | Aligned with real OSS content: added DSKP anchors, CFU, The Trap, chatbot rules. Made curriculum-agnostic (no hardcoded KSSM/BM) |
| `B-W4D18-4` | Create `prompts/assessments.md` — template for generating quiz questions with rubrics, hints, distractors, Bloom's levels, KSSM exam format | 🤖 | ✅ | Aligned with real OSS content: added tp_level, kbat, multiple_choice/free_text types, MISCONCEPTION ALERT pattern. Made curriculum-agnostic |
| `B-W4D18-5` | 🧑 Review and heavily edit prompt templates — these define teaching quality | 🧑 Education Lead (2hr) | ✅ | Reviewed: identified hardcoded KSSM/BM references, prompts generalized for multi-curriculum support. Verified against real p-n-ai/oss content structure |

### Day 19 (Thu) — Generation Commands

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W4D19-1` | `internal/generator/teaching_notes.go` — generate teaching notes for a topic: build context → inject into template → call AI → parse markdown → validate → write file | 🤖 | ✅ | TDD: curriculum-agnostic prompts, CFU/Trap/Strategies structure from real OSS content |
| `B-W4D19-2` | `internal/generator/assessments.go` — generate N assessment questions: build context → inject → call AI → parse YAML → validate against schema → retry if invalid → write file | 🤖 | ✅ | TDD: tp_level, kbat, multiple answer types aligned with real OSS content |
| `B-W4D19-3` | `internal/generator/examples.go` — generate worked examples with step-by-step solutions | 🤖 | ✅ | TDD: worked_examples format with real_world_analogy, misconception_alert from real OSS content. + prompts/examples.md |
| `B-W4D19-4` | `internal/pipeline/pipeline.go` — unified orchestrator: all interfaces call `pipeline.Execute()` with mode (Preview/WriteFS/CreatePR). `internal/output/writer.go` — `LocalWriter` (CLI) + `GitHubWriter` (Bot/Web) | 🤖 | ✅ | TDD: 3 pipeline tests. GitHubWriter placeholder for Week 5 |
| `B-W4D19-5` | Wire CLI commands via pipeline: `oss generate teaching-notes`, `oss generate assessments` | 🤖 | ✅ | All 3 generate commands wired via shared runGenerate + pipeline.Execute |

### Day 20 (Fri) — Translation + Testing

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W4D20-1` | Create `prompts/translation.md` — preserve YAML structure exactly, translate only human-readable fields, use correct BM mathematical terminology | 🤖 | ✅ | Curriculum-agnostic, preserves LaTeX notation |
| `B-W4D20-2` | `internal/generator/translator.go` + `oss translate --topic F1-01 --to ms` — generates locale file | 🤖 | ✅ | TDD: 3 tests. Supports ms, zh, ta, en. Required flags wired |
| `B-W4D20-3` | Test full pipeline: generate teaching notes + assessments + examples + translation for 1 Form 2 topic. Compare AI-generated vs manually-written quality. | 🤖🧑 | ✅ | E2E test covers all 4 content types through pipeline |
| `B-W4D20-4` | 🧑 Education Lead evaluates: is AI-generated content quality acceptable with light editing? What needs to improve in prompts? | 🧑 Education Lead | ✅ | Reviewed — all good, no prompt changes needed |

**Week 4 Output:** Working CLI with validate, generate, quality, translate commands. Shared pipeline orchestrator (`internal/pipeline`) and output writers (`internal/output`) — all future interfaces (Bot, Web) will call the same `pipeline.Execute()`. Prompt templates for KSSM content.

---

## WEEK 5 — GITHUB BOT + DOCUMENT IMPORT

> **Rebalanced (2026-03-27):** Day 23 was overloaded (8 tasks). Image extraction and bot quality command moved to Day 24. Contribution parser + feedback API moved to Day 25. Content merge logic added to Day 22 per design decision.

### Day 21 (Mon) — GitHub App Setup + Bot Server

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W5D21-1` | `internal/github/app.go` — GitHub App authentication (JWT from private key, installation token exchange) | 🤖 | ✅ | |
| `B-W5D21-2` | `internal/github/webhook.go` — webhook handler: verify HMAC signature, parse issue_comment events, extract @oss-bot commands | 🤖 | ✅ | BotCommand type lives in internal/parser to avoid circular imports |
| `B-W5D21-3` | `internal/parser/command.go` — parse bot commands: `add teaching notes`, `add N assessments`, `translate`, `scaffold`, `quality`, `import <url>`, `import` (with attachment) | 🤖 | ✅ | |
| `B-W5D21-4` | Wire `cmd/bot/main.go` — real HTTP server with webhook handler (replace placeholder) | 🤖 | ✅ | |
| `B-W5D21-5` | 🧑 Register GitHub App: p-n-ai org, webhook URL, permissions (Issues R/W, PRs R/W, Contents R/W) | 🧑 Human | ✅ | smee requires --path /webhook flag |

### Day 22 (Tue) — Bot → PR Pipeline + Content Merge

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W5D22-1` | `internal/github/pr.go` — create branch, commit files, open PR with labels and description | 🤖 | ✅ | PRRequest, FileChange, PRResult structs; GenerateBranchName, BuildPRBody helpers |
| `B-W5D22-2` | `internal/github/contents.go` — read existing topic files from oss repo via GitHub Contents API | 🤖 | ✅ | ContentsClient interface + MockContentsClient for tests |
| `B-W5D22-3` | `internal/generator/merge.go` — `MergeAssessments()` (append + dedup), `MergeExamples()` (append + dedup + re-sort by difficulty), additive teaching notes. Uses `FindDuplicates` from `duplicates.go`. Includes `MergeReport` for PR summary. | 🤖 | ✅ | Also added MergeAssessmentsYAML/MergeExamplesYAML for full-fidelity YAML merge |
| `B-W5D22-4` | Integrate merge into pipeline: detect existing content → merge → validate → output. Extend `PRInput` with merge report for PR description. | 🤖 | ✅ | ContentReader interface + WithContentReader; MergeReport in Result; MergeDetails in PRInput |
| `B-W5D22-5` | Bot command flow: parse `@oss-bot` comment → call shared `pipeline.Execute(ModeCreatePR)` → react to comment with PR link | 🤖 | ✅ | botServer struct wires pipeline at startup; handleCommand routes to pipeline.Execute |
| `B-W5D22-6` | Bot responds to issue with PR link: "I've generated teaching notes for F2-01 and opened #PR. Please review for accuracy." | 🤖 | ✅ | postIssueComment via GitHub REST API; installation token via JWT exchange |

### Day 23 (Wed) — Scaffolding + Document Import (PDF, URL, Tika)

> **Updated (2026-03-27):** Expanded scaffolder to handle new syllabus/subject creation from scratch. Added large document chunking. These are required for the global OSS use case (any country, any subject).
> **Updated (2026-03-27, Day 22 finding):** Added task `B-W5D23-11` to fix `GenerationResult.Files` always being `nil`. Without this, both `ModeWriteFS` and `ModeCreatePR` write/commit nothing. Must be completed before any file-write task in this day.

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W5D23-1` | `internal/generator/scaffolder.go` — `oss scaffold syllabus` (create `curricula/{country}/{syllabus}/syllabus.yaml`) and `oss scaffold subject` (create subject YAML + topics directory). Supports creating entirely new curricula from scratch. | 🤖 | ⬜ | New: global OSS support |
| `B-W5D23-2` | Create `prompts/document_import.md` — extract curriculum structure from documents and web pages, infer Bloom's levels from verbs, map prerequisites. Subject-agnostic (works for math, science, humanities). | 🤖 | ⬜ | |
| `B-W5D23-3` | Create `prompts/bulk_import.md` — for large documents: identify chapter/section boundaries, extract multiple topics with their learning objectives, generate full syllabus structure. Uses reasoning model for complex analysis. | 🤖 | ⬜ | New: bulk import prompt |
| `B-W5D23-4` | `internal/parser/document.go` — `ContentExtractor` interface (URL, file, text) shared by CLI and server | 🤖 | ⬜ | |
| `B-W5D23-5` | `internal/parser/pdf.go` — Go-native PDF text extraction using `ledongthuc/pdf` (for CLI standalone use) | 🤖 | ⬜ | |
| `B-W5D23-6` | `internal/parser/tika.go` — Apache Tika client using `google/go-tika` (for server multi-format: PDF, DOCX, PPTX, TXT) | 🤖 | ⬜ | |
| `B-W5D23-7` | `internal/parser/url.go` — URL fetcher: fetch web page, extract text content, pass to AI pipeline | 🤖 | ⬜ | |
| `B-W5D23-11` | Fix pipeline file map: populate `GenerationResult.Files` after generation using `genCtx.TopicDir` + topic file fields (`ai_teaching_notes`, `assessments_file`, `examples_file`). Fixes `ModeWriteFS` writing nothing and `ModeCreatePR` committing nothing. Add `buildFilesMap()` to `internal/pipeline/pipeline.go`. | 🤖 | ⬜ | Gap from Day 22: Files map always nil |

### Day 24 (Thu) — Bulk Import + Large Document Processing

> **Updated (2026-03-27):** Dedicated day for large document handling (100-page PDFs, textbooks, DSKP documents). This is the core scenario for bootstrapping a new country's curriculum.

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W5D24-1` | `internal/parser/chunker.go` — split large documents into processable chunks by chapter/heading boundaries. Handles 100+ page PDFs within AI token limits. Chunk-level caching for retry resilience. | 🤖 | ⬜ | New: large doc support |
| `B-W5D24-2` | `internal/pipeline/progress.go` — `ProgressReporter` interface with stages: extracting (%), chunking (%), analyzing structure, generating topic N/M, validating, writing. Implementations: CLI (terminal progress bar), Bot (edit GitHub comment with status), Web (SSE stream). All interfaces show real-time progress for long-running operations. | 🤖 | ⬜ | New: progress reporting |
| `B-W5D24-3` | `internal/pipeline/bulk.go` — bulk import orchestrator with **parallel agent workers**: chunks are analyzed sequentially (structure extraction needs full context), but topic generation runs concurrently via configurable worker pool (`OSS_WORKER_COUNT`, default 3). Each worker is an independent AI agent processing one topic. Progress reported per-topic. Handles: `oss import --file textbook.pdf --syllabus india-jee --subject chemistry-11` → generates syllabus.yaml + subject.yaml + N topic files + content. | 🤖 | ⬜ | New: multi-agent parallel processing |
| `B-W5D24-4` | `internal/ai/reasoning.go` — reasoning model provider via OpenRouter (single OpenAI-compatible API gateway routing to DeepSeek R1, Kimi K2.5, Qwen 3.5, o3-mini, etc.) for complex tasks: bulk import structure analysis, content merge decisions, cross-topic prerequisite mapping. Falls back to standard provider if unavailable. Config: `OSS_AI_REASONING_PROVIDER=openrouter`, `OSS_AI_REASONING_MODEL=deepseek/deepseek-r1`. | 🤖 | ⬜ | New: reasoning model support via OpenRouter |
| `B-W5D24-5` | Extend `internal/validator/bloom.go` — add cross-subject Bloom verbs: science (predict, hypothesize, synthesize, observe, experiment), humanities (interpret, critique, contextualize), general (research, collaborate, present) | 🤖 | ⬜ | New: multi-subject support |
| `B-W5D24-6` | `internal/parser/image.go` — Dual image extraction: OCR (Tesseract/Tika) for printed text + AI Vision (GPT-4o/Claude) for handwriting, diagrams, and complex layouts | 🤖 | ⬜ | Moved from previous Day 24 |

### Day 25 (Fri) — GitHub API Client + Bot Commands + Docker + Testing

> **Updated (2026-03-27, Gap 1 + Gap 3 from Day 22):** Added tasks `B-W5D25-11` and `B-W5D25-12`. Both are prerequisites for the end-to-end test (`B-W5D25-9`). Design decision: use stdlib `net/http` for all GitHub API calls — no `google/go-github/v62` dependency.

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W5D25-11` | `internal/github/client.go` — minimal stdlib HTTP client: `GetRef`, `CreateRef`, `PutContents`, `CreatePull`, `ReadFile`. Tests use `httptest.NewServer`. | 🤖 | ⬜ | Gap 1+3 from Day 22: no go-github dep |
| `B-W5D25-12` | Implement `GitHubWriter.CreatePR` (branch → commit → PR). Add `GitHubContentsClient` + `GitHubContentsReader`. Wire reader in `cmd/bot/main.go` so merge stage activates. | 🤖 | ⬜ | Gap 1+3 from Day 22: prerequisite for B-W5D25-9 |
| `B-W5D25-1` | `@oss-bot quality` command — responds with quality report for the topic in the issue | 🤖 | ⬜ | |
| `B-W5D25-2` | Create `prompts/contribution_parser.md` — parse natural language teacher input into structured YAML, preserve teacher's voice | 🤖 | ⬜ | |
| `B-W5D25-3` | `internal/parser/contribution.go` — teacher writes "My students always confuse the negative sign when expanding brackets" → structured misconception entry | 🤖 | ⬜ | |
| `B-W5D25-4` | `POST /api/feedback` — endpoint for pai-bot to submit observed patterns (misconception frequency, explanation effectiveness) | 🤖 | ⬜ | |
| `B-W5D25-5` | Feedback handler: receive structured feedback → run generation pipeline → create PR with provenance:ai-observed label | 🤖 | ⬜ | |
| `B-W5D25-6` | Dockerfile: multi-stage Go build for both CLI binary and bot server | 🤖 | ⬜ | |
| `B-W5D25-7` | `docker-compose.yml`: bot server + Apache Tika sidecar + webhook tunnel (for dev) | 🤖 | ⬜ | |
| `B-W5D25-8` | README.md: CLI installation (go install + pre-built binaries), GitHub App setup, bot deployment | 🤖 | ⬜ | |
| `B-W5D25-9` | Test end-to-end: create GitHub issue → comment @oss-bot add teaching notes for F3-02 → verify PR is created with valid content | 🤖🧑 | ⬜ | |
| `B-W5D25-10` | 🧑 Education Lead reviews 3 AI-generated PRs: would you approve these? What needs improvement? | 🧑 Education Lead | ⬜ | |

**Week 5 Output:** Working GitHub bot that generates content and opens PRs with intelligent content merging (additive by default). Scaffolding for new countries/syllabi/subjects. Bulk import from large documents (100-page PDFs, textbooks, DSKP). Reasoning model integration for complex analysis. Three input methods: URL import, file upload (PDF, DOCX, PPTX, TXT, images with OCR + AI Vision), and text (natural language). Multi-subject Bloom's taxonomy. Feedback API for pai-bot.

---

## WEEK 6 — WEB PORTAL + LAUNCH

### Day 26 (Mon) — Web Portal Scaffold

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W6D26-1` | Scaffold `web/`: Next.js 15 + TypeScript + shadcn/ui + Tailwind | 🤖 | ⬜ | |
| `B-W6D26-2` | Contribution form: Select curriculum (or create new) → Select topic (or import) → Contribution type → Three input methods: paste URL, type/paste text, or upload file (PDF, DOCX, PPTX, TXT, image) | 🤖 | ⬜ | Updated: supports new curriculum creation |
| `B-W6D26-3` | `POST /api/preview` — calls shared `pipeline.Execute(ModePreview)`, returns structured YAML. `POST /api/submit` — calls `pipeline.Execute(ModeCreatePR)`. Both delegate to the same pipeline as CLI and Bot. | 🤖 | ⬜ | |
| `B-W6D26-4` | `GET /api/progress/:jobId` — SSE (Server-Sent Events) endpoint streaming real-time progress from `ProgressReporter`. Web frontend shows: upload %, extraction %, "Analyzing structure...", "Generating topic 3/12...", "Validating...", "Done". | 🤖 | ⬜ | New: progress streaming |

### Day 27 (Tue) — Submit + Preview Flow

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W6D27-1` | Preview component: show structured output with syntax highlighting, diff against existing content | 🤖 | ⬜ | |
| `B-W6D27-2` | `POST /api/submit` — on confirmation, create GitHub PR with attribution to the contributor | 🤖 | ⬜ | |
| `B-W6D27-3` | Real-time schema validation in preview: show green checkmarks for valid fields, red for issues | 🤖 | ⬜ | |
| `B-W6D27-4` | Bulk import progress UI: upload large PDF → show chunking progress → topic-by-topic generation status → final preview of all generated files | 🤖 | ⬜ | New: bulk import UX |

### Day 28 (Wed) — Curricula Browser

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W6D28-1` | `GET /api/curricula` — list all syllabi, subjects, topics from the oss repo. Supports all countries. | 🤖 | ⬜ | |
| `B-W6D28-2` | Browse page: tree view of Country → Syllabus → Subject → Topic. Quality level badges. "Contribute" button per topic. "Add new curriculum" button. | 🤖 | ⬜ | Updated: multi-country |
| `B-W6D28-3` | Topic detail page: show existing content (teaching notes, examples, assessments) with "Improve this" buttons | 🤖 | ⬜ | |

### Day 29 (Thu) — Deploy + Documentation

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W6D29-1` | Deploy bot + web portal: Docker on VPS, configure GitHub App webhook URL | 🤖 | ⬜ | |
| `B-W6D29-2` | CONTRIBUTING.md: 3 ways to contribute (web form, @oss-bot, CLI), screenshot walkthrough. Includes "How to add a new country's curriculum" guide. | 🤖 | ⬜ | Updated: new curriculum guide |
| `B-W6D29-3` | 🧑 Test web portal with 2 teachers: can they contribute without knowing Git? | 🧑 Education Lead | ⬜ | |

### Day 30 (Fri) — Launch + Report

| Task ID | Task | Owner | Status | Remark |
|---------|------|-------|--------|--------|
| `B-W6D30-1` | 🧑 Announce web portal in launch materials: "contribute.p-n-ai.org — teachers can contribute without Git" | 🧑 Human | ⬜ | |
| `B-W6D30-2` | 🧑 Write oss-bot section of 6-week report: AI generation quality, bot PRs created, web portal usage, curricula onboarded | 🧑 Human | ⬜ | |

**Week 6 Output:** Web portal live at contribute.p-n-ai.org with multi-country curriculum support, bulk import UI, and new curriculum creation flow. GitHub bot responding to @oss-bot. CLI distributed as pre-built binary.

---

## Task Count Summary

| Week | 🤖 Claude Code | 🧑 Human | Total | Status |
|------|----------------|----------|-------|--------|
| 1-3 | 0 | 0 | 0 (no oss-bot work) | — |
| 4 | 18 | 2 | 20 | ✅ Complete (Days 16-20) |
| 5 | 32 | 2 | 34 | ⬜ Next (Days 21-25, rebalanced + new scenarios) |
| 6 | 13 | 2 | 15 | ⬜ (Days 26-30, updated for multi-country + progress UI) |
| **Total** | **63** | **6** | **69** |

---

## AI Model Strategy

> **Added (2026-03-27):** Different tasks require different model capabilities.

| Task Type | Model Tier | Examples | Candidates |
|-----------|-----------|----------|------------|
| **Standard generation** | Fast, affordable | Teaching notes, assessments, examples, translation | GPT-4o, Claude Sonnet 4, Llama 3 |
| **Complex reasoning** | Advanced reasoning | Bulk import structure extraction, content merge decisions, cross-topic prerequisite mapping, large document analysis | DeepSeek R1, Kimi K2.5, Qwen 3.5, OpenAI o3-mini, Claude Opus (via OpenRouter) |
| **Vision** | Multimodal | Handwriting OCR, diagram extraction, complex layouts | GPT-4o Vision, Claude Vision |

The `internal/ai/reasoning.go` provider (Day 24) uses **OpenRouter** as a unified API gateway (single OpenAI-compatible endpoint at `https://openrouter.ai/api/v1`) to access all reasoning models by changing the model name string. No custom provider implementations needed per model. Config via `OSS_AI_REASONING_PROVIDER=openrouter`, `OSS_AI_REASONING_API_KEY`, and `OSS_AI_REASONING_MODEL=deepseek/deepseek-r1` environment variables.

---

## Performance Targets

| Operation | Target |
|-----------|--------|
| `oss validate` (full repo) | <2s |
| Teaching notes generation | <15s |
| Assessment generation (5 questions) | <10s |
| PDF import, CLI (50-page syllabus) | <60s |
| Bulk import (100-page document) | <5min |
| URL import (fetch + extract) | <30s |
| Document import, server (50-page, any format) | <90s |
| Image extraction (OCR) | <5s |
| Image extraction (AI Vision) | <15s |
| Bot webhook → PR created | <30s |
| Web portal preview | <5s |
| CLI startup | <100ms |
