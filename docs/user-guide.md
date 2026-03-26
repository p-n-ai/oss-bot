# User Guide — Contributing to the Open School Syllabus

This document explains how educators, developers, and community members can contribute structured curriculum content to the [Open School Syllabus (OSS)](https://github.com/p-n-ai/oss) using OSS Bot.

OSS Bot provides three ways to contribute — choose whichever fits your workflow:

| Interface | Best For | Requires |
|-----------|----------|----------|
| [Web Portal](#web-portal) | Teachers and non-technical contributors | A browser |
| [GitHub Bot](#github-bot) | Community members already on GitHub | A GitHub account |
| [CLI Tool](#cli-tool) | Developers and power users | Go installed locally |

All three interfaces feed into the same AI-powered pipeline, so the quality, validation, and review process is identical regardless of how you contribute.

---

## Input Methods

You can provide source material in three ways. Every interface (Web Portal, GitHub Bot, CLI) supports all three.

### URL — Import from a web page

Paste a link to a curriculum page, syllabus PDF hosted online, textbook publisher page, or any publicly accessible educational resource. OSS Bot fetches the content, extracts the relevant curriculum structure, and converts it into structured YAML.

**Good sources:**
- Government curriculum specification pages (e.g., CBSE, Cambridge, KSSM)
- Textbook publisher table-of-contents pages
- University course outlines
- Online syllabus PDFs and documents

**What happens:** The bot fetches the page, extracts text (rendering JavaScript if needed), identifies topics, learning objectives, and Bloom's taxonomy levels from specification verbs, then structures everything into schema-valid YAML.

### Text — Copy and paste

Paste or type content directly. This can be structured content (a syllabus outline, a list of learning objectives) or freeform natural language (your teaching experience, notes, explanations). Write in any language.

**Examples of what you can paste:**
- A syllabus table copied from a PDF or Word document
- Your own teaching notes written in plain language
- A list of assessment questions you use in class
- A description of common student misconceptions you have observed
- Content in any language — the AI handles structuring and can translate if needed

**What happens:** The AI parses your text, identifies the content type (teaching notes, assessments, objectives, etc.), and structures it into the correct YAML format. Freeform teacher experience is mapped to specific fields like `common_misconceptions`, `engagement_hooks`, and `teaching_tips`.

### Upload — Attach a file

Upload a document containing curriculum content. Supported formats:

| Format | Extension | Notes |
|--------|-----------|-------|
| PDF | `.pdf` | Syllabus specifications, textbook excerpts, exam papers |
| Word | `.docx` | Curriculum documents, lesson plans, scheme of work |
| PowerPoint | `.pptx` | Lecture slides, topic breakdowns |
| Plain text | `.txt` | Simple text files, exported notes |
| Image | `.png`, `.jpg`, `.jpeg` | Photos of syllabuses, whiteboard notes, textbook pages (OCR extracted) |

**What happens:** The bot extracts text from the document (using OCR for images), identifies curriculum structure, maps content to topics and learning objectives, infers Bloom's levels, and produces structured YAML. For multi-page documents, the bot processes the entire file and may generate multiple topics.

---

## Web Portal

**URL:** `contribute.p-n-ai.org`

The web portal is the simplest way to contribute. No GitHub account, no terminal, no YAML knowledge required.

### Step-by-step

1. **Open the portal** at `contribute.p-n-ai.org`.

2. **Select a syllabus and topic.** Browse the curriculum tree to find the subject area you want to improve. You can also select "Add new syllabus" to propose an entirely new curriculum.

3. **Choose a contribution type:**
   - **Teaching Notes** — explanations, engagement hooks, worked examples, tips for delivering a topic
   - **Assessments** — practice questions with rubrics, hints, and common wrong answers
   - **Corrections** — fix errors in existing content (typos, inaccuracies, outdated information)
   - **Translations** — translate a topic into another language

4. **Provide your content** using any of the three input methods:

   **Paste a URL:**
   Enter a link to a curriculum page or online document. The portal fetches and extracts the content automatically.

   **Type or paste text:**
   Write in your own words, in any language. For example:

   > "When I teach simultaneous equations, I start with a real-world example like splitting a restaurant bill. Students often confuse which variable to eliminate first. I find that colour-coding the variables on the board helps. A common mistake is forgetting to flip the sign when subtracting equations."

   **Upload a file:**
   Drag and drop or browse to upload a PDF, DOCX, PPTX, TXT, or image file. The portal extracts the content and shows a processing indicator while the document is parsed.

5. **Preview the result.** The portal shows a live preview of the structured output. A green checkmark means the content passes schema validation. A red indicator means something needs adjustment — the portal will tell you what.

6. **Submit.** Once the preview looks right, confirm your submission. OSS Bot will:
   - Create a branch in the OSS repository
   - Commit your structured content with provenance metadata
   - Open a pull request attributed to you
   - Provide a link so you can track the PR

7. **Wait for review.** An educator with subject expertise will review your PR. They may approve it, request changes, or leave feedback. You will receive updates if you signed in with GitHub.

### Tips

- You do not need a GitHub account to submit. The bot creates the PR on your behalf.
- If you do sign in with GitHub, your contributions are attributed to your account and you can track them.
- You can write in any language — the system handles structuring and, if needed, translation.
- You can combine input methods — paste a URL and add your own teaching notes as text in the same contribution.
- For images, ensure the text is legible. High-resolution photos of printed syllabuses work well; blurry whiteboard photos may not.
- The preview step catches most issues before submission. If the preview shows errors, adjust your input and try again.

---

## GitHub Bot

**Handle:** `@oss-bot`

If you already work in the [p-n-ai/oss](https://github.com/p-n-ai/oss) repository, you can trigger content generation directly from issue or pull request comments.

### Available Commands

#### Generate teaching notes

```
@oss-bot add teaching notes for <topic-path>
```

Example:

```
@oss-bot add teaching notes for mathematics/algebra/03-simultaneous-equations
```

The bot generates structured teaching notes (engagement hooks, explanations, common misconceptions, worked examples) and opens a PR.

#### Generate assessments

```
@oss-bot add <count> assessments for <topic-path> difficulty:<level>
```

Example:

```
@oss-bot add 5 assessments for mathematics/algebra/01-expressions difficulty:medium
```

Generates practice questions with rubrics, progressive hints, worked solutions, and common wrong answers. Difficulty levels: `easy`, `medium`, `hard`.

#### Translate a topic

```
@oss-bot translate <topic-path> to <language-code>
```

Example:

```
@oss-bot translate mathematics/algebra/03-simultaneous-equations to AR
```

Translates all human-readable fields while preserving the YAML structure exactly. Language codes follow ISO 639-1 (e.g., `AR` for Arabic, `ML` for Malayalam, `ES` for Spanish).

#### Scaffold a new syllabus

```
@oss-bot scaffold syllabus <curriculum-path>
```

Example:

```
@oss-bot scaffold syllabus science/physics/mechanics
```

Creates the full directory structure and Level 0 stubs (topic names, empty fields) for a new syllabus. You fill in the details from there.

#### Import from a URL

```
@oss-bot import <url>
```

Example:

```
@oss-bot import https://www.cambridgeinternational.org/programmes/cambridge-igcse-mathematics-0580
```

The bot fetches the page, extracts curriculum structure, identifies topics and learning objectives, infers Bloom's taxonomy levels, and opens a PR with the structured result.

#### Import from an uploaded file

```
@oss-bot import
```

Attach a PDF, DOCX, PPTX, TXT, or image file to your comment. The bot extracts text (using OCR for images), identifies curriculum structure, and opens a PR with the structured result.

Supported attachments: `.pdf`, `.docx`, `.pptx`, `.txt`, `.png`, `.jpg`, `.jpeg`

#### Enrich a topic with text

```
@oss-bot enrich <topic-path>
<your natural language contribution>
```

Example:

```
@oss-bot enrich mathematics/algebra/03-simultaneous-equations
My students always struggle with the elimination method. I find that starting
with graphical solutions helps them understand what "solving simultaneously"
actually means. A common mistake is dropping negative signs when subtracting
one equation from another.
```

The bot parses your natural language into structured fields (teaching notes, misconceptions, tips) and opens a PR. Write in any language.

#### Run a quality report

```
@oss-bot quality <syllabus-path>
```

Example:

```
@oss-bot quality mathematics/algebra
```

Generates a quality assessment of the syllabus — which topics are at what quality level, what is missing to reach the next level. Posts the report as a comment (no PR created).

### What happens after you run a command

1. The bot reacts to your comment with a thumbs-up to acknowledge receipt.
2. It loads the topic and its context from the repository (related topics, schema, style examples).
3. The AI generates structured content using the appropriate prompt template.
4. The output is validated against the JSON Schema, checked for Bloom's taxonomy alignment, prerequisite graph integrity, and duplicate content.
5. If validation fails, the bot retries once with error feedback. If it still fails, the bot reports the error as a comment.
6. On success, the bot creates a branch (`oss-bot/<command>-<topic>-<timestamp>`), commits the files with provenance metadata, and opens a PR.
7. The bot replies to your comment with a link to the PR.
8. Subject-matter educators from CODEOWNERS are automatically requested as reviewers.

---

## CLI Tool

**Binary:** `oss`

The CLI is for developers who want to validate, generate, or import content locally before pushing to GitHub.

### Installation

```bash
# From source
go install github.com/p-n-ai/oss-bot/cmd/oss@latest

# Or download a pre-built binary
curl -LO https://github.com/p-n-ai/oss-bot/releases/latest/download/oss-$(uname -s)-$(uname -m)
chmod +x oss-*
sudo mv oss-* /usr/local/bin/oss
```

### Setup

```bash
# Point to your local OSS clone
export OSS_REPO_PATH=/path/to/oss

# Set your AI provider
export OSS_AI_PROVIDER=openai      # or: anthropic, ollama
export OSS_AI_API_KEY=sk-...       # not needed for ollama
```

### Commands

#### Validate content

No AI provider needed. Runs schema validation, Bloom's checks, prerequisite graph validation, and duplicate detection against your local files.

```bash
# Validate everything
oss validate

# Validate a single file
oss validate --file mathematics/algebra/03-simultaneous-equations/topic.yaml

# Validate an entire syllabus
oss validate --syllabus mathematics/algebra
```

#### Generate teaching notes

```bash
oss generate teaching-notes mathematics/algebra/03-simultaneous-equations
```

Writes structured teaching notes to the correct file in your local OSS clone. Review the output, then commit and push as usual.

#### Generate assessments

```bash
oss generate assessments mathematics/algebra/01-expressions --count 5 --difficulty medium
```

#### Generate worked examples

```bash
oss generate examples mathematics/algebra/03-simultaneous-equations --count 3
```

#### Translate

```bash
# Single topic
oss translate --topic mathematics/algebra/03-simultaneous-equations --to AR

# Entire syllabus
oss translate --syllabus mathematics/algebra --to ES
```

#### Import from a URL

```bash
oss import --url https://example.org/curriculum-specification --board CBSE --level secondary --subject mathematics
```

Fetches the page, extracts curriculum structure, and writes structured YAML files to your local clone.

#### Import from a file

```bash
# PDF (Go-native, no external dependencies)
oss import --pdf curriculum.pdf --board CBSE --level secondary --subject mathematics

# Word, PowerPoint, text (requires Apache Tika running locally or via Docker)
oss import --file syllabus.docx --board CBSE --level secondary --subject mathematics

# Image — extracts text via OCR
oss import --file whiteboard-photo.jpg --board CBSE --level secondary --subject mathematics
```

Supported formats: `.pdf`, `.docx`, `.pptx`, `.txt`, `.png`, `.jpg`, `.jpeg`

The import command extracts text from the source (using OCR for images), identifies curriculum structure, infers Bloom's levels, and writes structured YAML files to your local clone.

#### Quality report

```bash
oss quality mathematics/algebra
```

Prints a quality assessment: which topics are at what level, what is missing to reach the next level.

#### Contribute with natural language

```bash
oss contribute "When teaching simultaneous equations, I always start with graphical solutions..."
```

Parses your natural language into structured YAML fields.

### Local workflow

A typical CLI workflow looks like this:

```bash
# 1. Clone the OSS repo
git clone https://github.com/p-n-ai/oss.git
cd oss

# 2. Create a branch
git checkout -b add-teaching-notes-simultaneous-equations

# 3. Generate content
oss generate teaching-notes mathematics/algebra/03-simultaneous-equations

# 4. Validate
oss validate --file mathematics/algebra/03-simultaneous-equations/topic.yaml

# 5. Review the generated files, make manual edits if needed

# 6. Commit and push
git add .
git commit -m "Add teaching notes for simultaneous equations"
git push origin add-teaching-notes-simultaneous-equations

# 7. Open a PR on GitHub
```

---

## What Gets Generated

Regardless of which interface you use, all generated content includes **provenance metadata** so reviewers and downstream tools know where the content came from:

```yaml
_metadata:
  provenance: ai-generated
  model: gpt-4o
  generator: oss-bot v0.1.0
  generated_at: "2026-03-26T10:00:00Z"
  context_topics:
    - 01-expressions
    - 02-linear-equations
  reviewed_by: null
```

The `provenance` field can be:

| Value | Meaning |
|-------|---------|
| `human` | Written entirely by a human contributor |
| `ai-assisted` | Human-written with AI help (e.g., via the web portal or `oss contribute`) |
| `ai-generated` | Generated by the AI pipeline from a bot command or CLI generation |
| `ai-observed` | Generated from P&AI Bot's learning analytics (student struggle patterns) |

After a human reviewer approves the PR, the `reviewed_by` field is updated with their GitHub handle.

---

## Validation and Quality Gates

Every contribution passes through four automated checks before a PR is created:

1. **JSON Schema validation** — all fields match required types, formats, and constraints. Invalid YAML is never submitted.

2. **Bloom's taxonomy verification** — assessment questions use verbs that match their cognitive level. A "Remember"-level question cannot use "Evaluate" or "Analyze" verbs.

3. **Prerequisite graph integrity** — no circular dependencies. All referenced topic IDs must exist in the syllabus.

4. **Duplicate detection** — new content is compared against existing content using cosine similarity. Items with >85% similarity are flagged.

If any check fails, the pipeline retries once with the error feedback injected into the AI prompt. If it still fails, the error is reported and no PR is created.

### Quality levels

The bot self-assesses every topic's quality on a scale of 0 to 5:

| Level | Name | What it means |
|-------|------|---------------|
| 0 | Stub | Topic name only, no content |
| 1 | Basic | Learning objectives defined |
| 2 | Structured | Learning objectives + schema-valid structure |
| 3 | Teachable | Complete teaching notes + assessments |
| 4 | Complete | Teaching notes + assessments + examples + misconceptions |
| 5 | Excellent | All of the above + translations + cross-curriculum links |

The quality level is included in the PR description and as a label (e.g., `quality:level-3`), so reviewers can see at a glance what the contribution achieves.

---

## The Review Process

All contributions require human educator review before merging into the OSS repository. The bot never auto-merges.

1. The PR is created with labels indicating provenance and quality level.
2. Subject-matter educators listed in the repository's CODEOWNERS file are automatically requested as reviewers.
3. For translations, native speakers of the target language are requested in addition to subject experts.
4. Reviewers can approve, request changes, or leave inline comments.
5. If changes are requested, the contributor (or the bot, if prompted) can update the PR.
6. Once approved, a maintainer merges the PR into the main branch.

---

## Input Methods at a Glance

| Input method | Web Portal | GitHub Bot | CLI |
|--------------|-----------|------------|-----|
| **URL** | Paste link in input field | `@oss-bot import <url>` | `oss import --url <url>` |
| **Text** | Type or paste in text area | Write in comment body | `oss contribute "..."` |
| **Upload file** | Drag-and-drop or browse | Attach file to comment | `oss import --file <path>` |

Supported upload formats: PDF, DOCX, PPTX, TXT, PNG, JPG/JPEG

## Contribution Types at a Glance

| What you want to do | Web Portal | GitHub Bot | CLI |
|---------------------|-----------|------------|-----|
| Add teaching notes | Select topic, type notes | `@oss-bot add teaching notes for <path>` | `oss generate teaching-notes <path>` |
| Add assessments | Select topic, choose "Assessment" | `@oss-bot add N assessments for <path>` | `oss generate assessments <path>` |
| Add worked examples | Select topic, choose "Teaching Notes" | `@oss-bot enrich <path>` | `oss generate examples <path>` |
| Translate content | Select topic, choose "Translation" | `@oss-bot translate <path> to <lang>` | `oss translate --topic <path> --to <lang>` |
| Import from URL | Paste URL in input field | `@oss-bot import <url>` | `oss import --url <url>` |
| Import from file | Upload PDF, DOCX, PPTX, TXT, or image | `@oss-bot import` + attachment | `oss import --file <path>` |
| Scaffold a new syllabus | Select "Add new syllabus" | `@oss-bot scaffold syllabus <path>` | Manual directory creation |
| Fix an error | Select topic, choose "Correction" | Edit file, open PR manually | Edit file locally, push PR |
| Check quality | Automatic during preview | `@oss-bot quality <path>` | `oss quality <path>` |
| Validate content | Automatic during preview | Automatic before PR | `oss validate` |
