# OSS Bot — Business Plan

*Last updated: February 2026*

---

## Executive Summary

OSS Bot is the AI-powered tooling layer for the [Open School Syllabus](https://github.com/p-n-ai/oss). It provides three interfaces — a GitHub bot, a CLI tool, and a web contribution portal — that enable anyone to contribute structured curriculum content without needing to write YAML, understand JSON Schemas, or use Git.

OSS Bot exists to solve one problem: **the contribution bottleneck.** Teachers have the pedagogical knowledge that makes AI tutoring effective. But they can't contribute that knowledge into a machine-readable curriculum repository if doing so requires technical skills. OSS Bot bridges this gap — teachers describe their expertise in natural language, and the bot structures it into schema-valid, review-ready content.

---

## Problem

### The Contribution Bottleneck

Open School Syllabus (OSS) needs three things to succeed at scale: comprehensive curricula, rich teaching insights, and continuous improvement. All three depend on contributions from educators — people who teach these subjects daily and understand what works.

But the OSS repository is YAML files in a Git repository. The typical contributor profile (a 15-year veteran math teacher) has never used Git, doesn't know what YAML is, and would abandon the process after seeing a JSON Schema validation error. This is not a knowledge gap to be "solved with documentation." It is a fundamental interface mismatch.

### Without OSS Bot, OSS Growth Is Capped

Without tooling, OSS can only grow as fast as the core team can manually structure content. The Pandai team can seed 2–3 curricula. Getting to 50+ curricula requires thousands of educators contributing. Each additional educator who finds the contribution process too complex is a curriculum that never gets added.

### The Content Generation Paradox

AI can generate curriculum content at scale — but only if given proper context. An AI asked to "create teaching notes for quadratic equations" produces generic, mediocre output. An AI given the existing topic YAML, related prerequisites, the curriculum's assessment structure, and examples of what "good" looks like in OSS produces high-quality, schema-compliant content. OSS Bot encapsulates this context-building logic so every generation is grounded in the right information.

---

## Solution

OSS Bot provides three interfaces to the same AI-powered content pipeline:

### 1. GitHub Bot (`@oss-bot`)

A GitHub App installed on the OSS repository. Community members mention `@oss-bot` in issues or PR comments with natural language commands:

```
@oss-bot add teaching notes for cambridge/igcse/mathematics-0580/topics/algebra/05-quadratic-equations
```

The bot reads the existing topic, generates content, validates it against the schema, and opens a PR. All contributions go through standard GitHub review — bot creates, humans approve.

**Use cases:**
- Generate teaching notes, assessments, or worked examples for existing topics
- Translate topics to new languages
- Scaffold entire new syllabi from descriptions or PDF attachments
- Enrich topics with classroom experience described in natural language

### 2. CLI Tool (`oss`)

A command-line tool for developers and power users who work with local OSS clones:

```bash
oss validate                                          # Validate all YAML
oss generate teaching-notes cambridge/.../quadratics  # Generate content
oss import --pdf ./syllabus.pdf --board cambridge     # Import from PDF
oss translate --topic ... --to ms                     # Translate
oss quality cambridge/igcse/mathematics-0580          # Quality report
oss contribute "I teach this topic by..."             # Natural language
```

**Use cases:**
- Validate local changes before committing
- Batch-generate content for new syllabi
- Import curricula from official PDF documents
- Run quality analysis across a syllabus
- Contribute via terminal (for developers who prefer CLI)

### 3. Web Portal (contribute.p-n-ai.org)

A simple web form for teachers who don't use Git or the command line:

1. Select a syllabus and topic (or "Add new syllabus")
2. Choose contribution type (teaching notes, assessment, correction, translation)
3. Type contribution in natural language
4. Preview the structured output
5. Submit — a PR is created on their behalf

**Use cases:**
- Teacher adds misconceptions they've observed in 8 years of teaching
- Teacher contributes practice questions from their classroom
- Teacher corrects an inaccuracy in existing content
- Teacher translates a topic to their local language

---

## Target Users

| User | Interface | What They Do | Volume |
|------|-----------|-------------|--------|
| **Teachers** (primary) | Web portal | Contribute teaching notes, misconceptions, assessments in natural language | High — thousands potential |
| **Curriculum designers** | CLI + GitHub | Structure new syllabi, review AI-generated content | Medium |
| **Developers** | CLI | Validate, import, generate while building on OSS | Medium |
| **P&AI Bot** (automated) | API | Submit data-backed improvement suggestions from student interaction data | High — automated |
| **Community members** | GitHub Bot | Request content, report issues, contribute via comments | Growing over time |

---

## Product Strategy

### Phase 1: Validation Tools (Weeks 0–4)

Before any generation, OSS needs reliable validation. The first deliverable is the `oss validate` CLI command that checks all YAML against JSON Schemas. This runs in CI on every PR and locally before commits.

**Deliverables:**
- `oss validate` — schema validation for all file types
- GitHub Action that blocks invalid PRs
- Quality report: `oss quality` shows completeness per syllabus

### Phase 2: Content Generation (Weeks 4–5)

AI-powered content generation for the most common contribution types:

**Deliverables:**
- `oss generate teaching-notes` — generates teaching notes from topic YAML
- `oss generate assessments` — generates practice questions with rubrics
- `oss import --pdf` — imports curriculum structure from PDF documents
- `oss translate` — generates translations matching source structure

### Phase 3: GitHub Bot (Week 6)

The bot that enables community contributions directly in GitHub:

**Deliverables:**
- GitHub App responding to `@oss-bot` mentions
- All generation commands available via bot comments
- Auto-PR creation with labels, quality assessment, reviewer assignment

### Phase 4: Web Portal (Week 6+)

The web contribution form for non-technical educators:

**Deliverables:**
- Simple form at contribute.p-n-ai.org
- Natural language → structured YAML preview
- PR creation on behalf of contributors
- Contribution tracking (optional GitHub sign-in)

### Phase 5: P&AI Feedback Pipeline (Month 2+)

Automated contributions from P&AI Bot's student interaction data:

**Deliverables:**
- API endpoint for P&AI Bot to submit observed patterns
- Auto-generation of improvement issues (misconceptions, teaching notes, engagement data)
- All auto-generated content tagged `ai-observed`, requiring educator review

---

## AI Content Generation Pipeline

Every generation follows the same pipeline regardless of interface:

```
Input (natural language, command, or PDF)
    │
    ▼
┌─────────────────────────┐
│  1. Context Building     │  Load topic YAML, related topics,
│                          │  existing content, schema rules,
│                          │  quality standards
└──────────┬──────────────┘
           │
           ▼
┌─────────────────────────┐
│  2. AI Generation        │  Pedagogical prompt + schema-aware
│                          │  output constraints + style matching
│                          │  to existing content
└──────────┬──────────────┘
           │
           ▼
┌─────────────────────────┐
│  3. Validation           │  JSON Schema check, Bloom's taxonomy
│                          │  verification, prerequisite graph
│                          │  consistency, duplicate detection
└──────────┬──────────────┘
           │
           ▼
┌─────────────────────────┐
│  4. Quality Assessment   │  Self-assess quality level,
│                          │  identify what's missing for
│                          │  the next level
└──────────┬──────────────┘
           │
           ▼
┌─────────────────────────┐
│  5. Output               │  Write files, open PR, add labels,
│                          │  request reviewers, include
│                          │  provenance metadata
└─────────────────────────┘
```

**Context building is the key differentiator.** A generic AI call produces mediocre content. OSS Bot builds rich context by loading the target topic's full YAML, reading related topics for consistency, loading the schema to constrain output structure, and referencing existing high-quality content as style examples. This context-aware generation is what makes the output publishable, not just plausible.

---

## Generation Quality

### Quality Safeguards

| Safeguard | What It Catches | Automated? |
|-----------|----------------|-----------|
| Schema validation | Structural errors (missing fields, wrong types) | ✅ Yes — CI blocks |
| Bloom's taxonomy check | Misaligned cognitive levels (e.g., "remember" verb for "analyze" objective) | ✅ Yes |
| Prerequisite graph validation | Circular dependencies, missing prerequisites | ✅ Yes |
| Duplicate detection | Content too similar to existing questions/notes | ✅ Yes |
| Copyright check | Content resembling copyrighted exam papers | ✅ Yes |
| Pedagogical review | Teaching quality, accuracy, cultural sensitivity | ❌ No — requires human educator |
| Native speaker review | Translation quality | ❌ No — requires native speaker |

### Provenance Tracking

Every generated piece of content includes provenance metadata:

```yaml
provenance: ai-generated          # or: human, ai-assisted, ai-observed
generated_by: oss-bot v1.0.0
generation_context:
  model: claude-sonnet-4-20250514
  source_topic: 05-quadratic-equations
  prompt_template: teaching_notes_v2
  generation_date: 2026-03-15
```

Downstream consumers (like P&AI Bot) can filter by provenance if they want only human-verified content.

---

## Technology Decisions

### Why Go?

OSS Bot is written in Go to match P&AI Bot's stack. Single binary distribution (no runtime dependencies), fast execution for CLI operations, and excellent AI SDK support. Same team, same language, same tooling.

### Why AI-Powered (Not Template-Based)?

Templates produce predictable but rigid output. Curriculum content is inherently varied — a teaching note for quadratic equations is structurally different from one for photosynthesis. AI generation with schema constraints produces content that is both structurally valid (schema compliance) and pedagogically natural (language quality).

### Why Three Interfaces?

Different users have different comfort zones. Forcing teachers onto GitHub loses 90% of potential contributors. Forcing developers onto a web form frustrates them. Three interfaces, one pipeline — maximum reach, minimal duplication.

---

## Key Metrics

### Contribution Pipeline Metrics

| Metric | Week 6 | Month 3 | Month 6 | Month 12 |
|--------|--------|---------|---------|----------|
| PRs generated by bot | 10 | 50 | 200+ | 1,000+ |
| PRs merged (acceptance rate) | 70%+ | 75%+ | 80%+ | 85%+ |
| Unique contributors using bot/CLI/web | 5 | 20 | 100+ | 500+ |
| Web portal contributions | 0 | 10 | 50+ | 200+ |
| AI-observed improvements from P&AI | 0 | 10 | 50+ | 200+ |

### Generation Quality Metrics

| Metric | Target |
|--------|--------|
| Schema validation pass rate (first generation) | ≥95% |
| PR acceptance rate (after human review) | ≥75% |
| Average revisions before merge | <2 |
| Time from contribution to merged PR | <48 hours |
| Educator satisfaction with generated content | ≥4/5 |

### Impact Metrics

| Metric | Description | Target (Month 6) |
|--------|------------|-------------------|
| Content acceleration factor | How much faster OSS grows with bot vs. without | 5x |
| Teacher contribution rate | % of teachers who complete a contribution after starting | ≥60% via web portal |
| Cross-curriculum coverage | New curricula per month attributed to bot scaffolding | 3+ |
| Translation coverage | Languages per curriculum attributed to bot translation | 3+ per curriculum |

---

## Relationship to Other Repositories

### OSS Bot → OSS (Open School Syllabus)

OSS Bot operates *on* the OSS repository. It generates content, validates it, and creates PRs. It is the primary growth engine for OSS. Without OSS Bot, OSS growth is limited to manual contributions. With OSS Bot, any teacher can contribute in natural language.

### OSS Bot → P&AI Bot

P&AI Bot is both a consumer of OSS content and a contributor via OSS Bot. The feedback pipeline works as follows:

1. P&AI Bot teaches students using OSS content
2. P&AI Bot observes patterns (misconceptions, effective explanations, engagement data)
3. P&AI Bot calls OSS Bot's API to submit improvement suggestions
4. OSS Bot validates and structures the suggestions into PRs
5. Educators review and merge the improvements
6. OSS content improves → P&AI Bot teaches better → cycle repeats

### Independence

OSS Bot can be used independently of P&AI Bot. Any organization that forks OSS can install their own OSS Bot instance to manage contributions to their fork. The GitHub App, CLI, and web portal all work against any OSS-compatible repository.

---

## Deployment & Operations

### Hosted (by Pandai)

The default deployment. Pandai runs the GitHub bot and web portal:

- **GitHub Bot:** GitHub App installed on `p-n-ai/oss`, webhook handler on a small server
- **Web Portal:** Static Next.js site on Vercel/Netlify with serverless API for AI generation
- **CLI:** Distributed as a Go binary via GitHub Releases

**Cost:** ~$50/month (server for webhook handler + AI API costs for generation)

### Self-Hosted (for OSS Forks)

Organizations that fork OSS for their own curriculum can run their own OSS Bot:

```bash
git clone https://github.com/p-n-ai/oss-bot.git
cd oss-bot && cp .env.example .env
# Configure: GitHub App credentials, AI API key, target repo
docker compose up -d
```

This enables schools, governments, and organizations to manage their own curriculum contributions with the same AI-powered tooling.

---

## Sustainability

OSS Bot is lightweight to maintain:

| Cost | Amount | Notes |
|------|--------|-------|
| Server hosting | ~$20/month | Webhook handler + web portal backend |
| AI API costs | ~$30/month | Generation calls (scales with contribution volume) |
| GitHub App hosting | Free | GitHub provides free hosting for Apps |
| Development time | 3–5 hrs/week | Bug fixes, prompt improvements, new commands |

**Total: ~$50/month + engineering time**

Costs scale linearly with contribution volume — more contributions mean more AI calls. At high volume (1,000+ PRs/month), AI costs could reach $200–500/month, offset by the enormous value of the generated content.

---

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Generated content is low quality | Medium | High | Schema validation catches structural issues. Prompt engineering with examples of high-quality content. Human review required for all merges. Continuous prompt iteration based on rejection reasons. |
| Teachers don't use the web portal | Medium | Medium | User testing with real teachers. Iterate on UX. Fall back to GitHub Issues as lowest-friction alternative. |
| AI hallucination introduces factual errors | Medium | High | Educator review required. Generated content tagged with provenance. Quality guidelines document common AI mistakes for reviewers. |
| GitHub API rate limits | Low | Low | Batch operations. Cache API responses. The bot operates within reasonable call volumes. |
| AI costs spike with volume | Low | Low | Route generation to cheaper models when possible. Ollama option for self-hosted deployments. Budget caps in configuration. |
| Competing tool emerges | Low | Low | OSS Bot is tightly integrated with OSS's specific schema and workflow. General-purpose tools won't match this specificity. |

---

## Execution Timeline

| Week | Milestone |
|------|-----------|
| Week 0–2 | `oss validate` CLI tool + GitHub CI integration |
| Week 3–4 | `oss generate` commands (teaching notes, assessments) |
| Week 4 | `oss import --pdf` for curriculum importing |
| Week 5 | `oss translate` and `oss quality` commands |
| Week 5 | `oss contribute` (natural language → structured PR) |
| Week 6 | GitHub Bot (`@oss-bot`) responding to commands in OSS repo |
| Week 6 | Web contribution portal (contribute.p-n-ai.org) |
| Month 2 | P&AI → OSS Bot feedback pipeline (automated improvement suggestions) |
| Month 3 | Prompt optimization based on acceptance rates |
| Month 4 | Self-hosting documentation and simplified setup |
| Month 6 | API for third-party platforms to submit contributions |

---

## Vision

**Make it effortless for any educator on Earth to contribute their teaching wisdom to an open, AI-ready curriculum — in their own language, in their own words, without ever touching a line of code.**

The world's best teaching strategies live in the heads of individual teachers. OSS Bot is the bridge that turns those strategies into structured data that AI agents can use to teach millions of students. Every teacher who contributes multiplies their impact from one classroom to every classroom that uses OSS.

---

*A [Pandai](https://pandai.org) initiative. Tooling for the world's open curriculum.*
