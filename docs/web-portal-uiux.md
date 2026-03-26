# Web Portal — UI/UX Specification & Wireframes

> **Status:** Planned (Week 6 of [development timeline](development-timeline.md))
>
> **Reference:** [docs/technical-plan.md](technical-plan.md) for architecture, [docs/user-guide.md](user-guide.md) for user flows
>
> **URL:** `contribute.p-n-ai.org`

This document provides UI/UX specifications, layout wireframes, and interaction patterns for the OSS Bot contribution web portal. The design system is shared with the [P&AI Bot admin panel](../../pnai-pai-bot/docs/admin-panel-uiux.md) to maintain visual consistency across the P&AI ecosystem.

---

## Table of Contents

- [Design System](#design-system)
- [Shell & Navigation](#shell--navigation)
- [Landing Page](#landing-page)
- [Contribution Flow](#contribution-flow)
  - [Step 1 — Select Syllabus & Topic](#step-1--select-syllabus--topic)
  - [Step 2 — Choose Contribution Type](#step-2--choose-contribution-type)
  - [Step 3 — Provide Content](#step-3--provide-content)
  - [Step 4 — Preview & Validate](#step-4--preview--validate)
  - [Step 5 — Submit](#step-5--submit)
- [Contribution History](#contribution-history)
- [Quality Report View](#quality-report-view)
- [Shared Components](#shared-components)
- [Responsive Behavior](#responsive-behavior)
- [Interaction Patterns](#interaction-patterns)
- [Technical Notes](#technical-notes)

---

## Design System

Shared with the P&AI Bot admin panel to ensure ecosystem-wide visual consistency.

### Colors (OKLch via CSS custom properties)

| Token | Light | Dark | Usage |
|-------|-------|------|-------|
| `--primary` | Slate 950 | White | Headings, primary text |
| `--accent` | Sky 700 | Sky 300 | Links, active states, eyebrow labels |
| `--success` | Emerald | Emerald 400/18 | Quality level ≥ 3, validation pass |
| `--warning` | Amber | Amber 400/18 | Quality level 1–2, validation warnings |
| `--danger` | Rose | Rose 400/18 | Validation errors, quality level 0 |
| `--surface` | White/85 | Slate 950/60 | Card backgrounds |
| `--surface-dark` | Slate 950 | Slate 900/90 | Hero aside, dark cards |

### Typography

- **Eyebrow:** `text-xs font-semibold uppercase tracking-[0.22em]` — sky-700 (light) / sky-300 (dark)
- **Page title:** `text-3xl font-semibold tracking-tight`
- **Card title:** `text-xl tracking-tight`
- **Body:** `text-sm leading-6`
- **Label:** `text-xs uppercase tracking-[0.18em]` — slate-500 / slate-400

### Spacing & Radius

- Page sections: `space-y-6`
- Card border radius: `rounded-[28px]`
- Inner containers: `rounded-[24px]` or `rounded-2xl`
- Card shadow (light): `shadow-[0_18px_60px_rgba(15,23,42,0.05)]`
- Card shadow (dark): `shadow-[0_24px_80px_rgba(2,8,23,0.35)]`

### Component Library (shadcn/ui)

Installed: `Button`, `Card`, `Dialog`, `Input`, `Label`, `Select`, `Textarea`, `Badge`, `Tabs`, `Progress`, `Accordion`, `Tooltip`

Custom components: `PageHero`, `StatCard`, `StatePanel`, `StepIndicator`, `ContributionShell`

---

## Shell & Navigation

The contribution portal uses a simpler shell than the pai-bot admin panel — no sidebar, top navigation only. The focus is on a linear contribution flow.

### Desktop Layout (≥ 1024px)

```
┌───────────────────────────────────────────────────────────────────────┐
│                        max-w-[1200px] centered                        │
├───────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │  🌿 Open School Syllabus                                        │  │
│  │                                                                 │  │
│  │  [Home]  [Contribute]  [Quality Reports]  [My Contributions]    │  │
│  │                                                    🌙  [Sign in]│  │
│  └─────────────────────────────────────────────────────────────────┘  │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │                                                                 │  │
│  │                       PAGE CONTENT                              │  │
│  │                       (max-w-4xl for forms)                     │  │
│  │                       (max-w-6xl for tables)                    │  │
│  │                                                                 │  │
│  └─────────────────────────────────────────────────────────────────┘  │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────────────┐  │
│  │  Footer: P&AI · GitHub · License (Apache 2.0)                   │  │
│  └─────────────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────────────┘
```

### Mobile Layout (< 1024px)

```
┌──────────────────────────────────┐
│ 🌿 OSS  ☰                  🌙    │  ← sticky top bar
├──────────────────────────────────┤
│ ┌──────────────────────────────┐ │  ← collapsible nav
│ │  Home                        │ │     (slides down on ☰ tap)
│ │  Contribute                  │ │
│ │  Quality Reports             │ │
│ │  My Contributions            │ │
│ │  Sign in with GitHub         │ │
│ └──────────────────────────────┘ │
├──────────────────────────────────┤
│                                  │
│       PAGE CONTENT               │
│       (full width, px-4)         │
│                                  │
└──────────────────────────────────┘
```

### Navigation Items

| Item | Route | Auth Required | Description |
|------|-------|---------------|-------------|
| Home | `/` | No | Landing page with curriculum overview |
| Contribute | `/contribute` | No | Start a new contribution |
| Quality Reports | `/quality` | No | Browse syllabus quality reports |
| My Contributions | `/contributions` | Yes (GitHub) | Track past contributions and PR status |
| Sign in | — | — | GitHub OAuth (optional) |

---

## Landing Page

**Route:** `/`
**Access:** Public

```
┌────────────────────────────────────────────────────────────────────┐
│  OPEN SCHOOL SYLLABUS                                              │
│  "Contribute curriculum knowledge"                                 │
│  Help build free, open, AI-ready curriculum for           ┌──────┐ │
│  every student. No Git or YAML knowledge needed.          │Topics│ │
│  Share your teaching expertise in plain language.         │1,240 │ │
│                                                           │across│ │
│  [Start contributing →]                                   │ 24   │ │
│                                                           │syllab│ │
│                                                           └──────┘ │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐    │
│  │ Syllabi    │  │ Topics     │  │ Contribut- │  │ Languages  │    │
│  │    24      │  │   1,240    │  │ ions       │  │    8       │    │
│  │ Across 6   │  │ Structured │  │    342     │  │ Translated │    │
│  │ countries  │  │ topics     │  │ PRs merged │  │ so far     │    │
│  └────────────┘  └────────────┘  └────────────┘  └────────────┘    │
│                                                                    │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ HOW IT WORKS                                                 │  │
│  │                                                              │  │
│  │  ① Select            ② Contribute           ③ Review       │  │
│  │                                                              │  │
│  │  Pick a syllabus     Type, paste a URL,     An educator      │  │
│  │  and topic from      or upload a file.      reviews your     │  │
│  │  the curriculum      Write in any           contribution     │  │
│  │  tree.               language.              and merges it.   │  │
│  │                                                              │  │
│  │                    [Start contributing →]                    │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                    │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ THREE WAYS TO CONTRIBUTE                                     │  │
│  │                                                              │  │
│  │  ┌──────────────────┐ ┌──────────────────┐ ┌──────────────┐  │  │
│  │  │ 🔗 Paste a URL   │ │ ✏️ Type or paste │ │ 📄 Upload     │  │  │
│  │  │                  │ │                  │ │              │  │  │
│  │  │ Link to a curric-│ │ Write in your    │ │ PDF, Word,   │  │  │
│  │  │ ulum page,       │ │ own words, in    │ │ PowerPoint,  │  │  │
│  │  │ syllabus PDF,    │ │ any language.    │ │ images, or   │  │  │
│  │  │ or textbook page.│ │ Share teaching   │ │ text files.  │  │  │
│  │  │                  │ │ experience.      │ │              │  │  │
│  │  └──────────────────┘ └──────────────────┘ └──────────────┘  │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                    │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ RECENT CONTRIBUTIONS (dark card)                             │  │
│  │                                                              │  │
│  │  "Add teaching notes for IGCSE Quadratic Equations"          │  │
│  │   by cikgu_aminah · merged 2 hours ago · quality:level-3     │  │
│  │                                                              │  │
│  │  "Translate KSSM Algebra to Malay"                           │  │
│  │   by @contributor42 · in review · quality:level-4            │  │
│  │                                                              │  │
│  │  "Import CBSE Class 10 Mathematics from PDF"                 │  │
│  │   by anonymous · merged 1 day ago · quality:level-2          │  │
│  └──────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────┘
```

**Interactions:**
- "Start contributing" → navigates to `/contribute`
- Click recent contribution → opens GitHub PR in new tab
- Sign in with GitHub → enables "My Contributions" tracking

---

## Contribution Flow

The contribution process follows a linear 5-step wizard. A `StepIndicator` at the top shows progress.

### Step Indicator

```
┌────────────────────────────────────────────────────────────────────┐
│  ● Select  ─── ● Type  ─── ○ Content  ─── ○ Preview  ─── ○ Submit  │
│  Step 1         Step 2       Step 3         Step 4         Step 5  │
└────────────────────────────────────────────────────────────────────┘
```

- Completed steps: `●` with sky-700 fill + connector line
- Current step: `●` with sky-700 fill, label in bold
- Future steps: `○` with slate-300 outline

---

### Step 1 — Select Syllabus & Topic

**Route:** `/contribute`
**Access:** Public

```
┌───────────────────────────────────────────────────────────────────┐
│  CONTRIBUTE                                                       │
│  "Select a syllabus and topic"                                    │
│  Browse the curriculum tree to find the area you        ┌───────┐ │
│  want to improve, or add a new syllabus.                │Step   │ │
│                                                         │1 of 5 │ │
│                                                         └───────┘ │
│  ● Select ─── ○ Type ─── ○ Content ─── ○ Preview ─── ○ Submit     │
├───────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────────────┐  ┌───────────────────────────┐  │
│  │ SYLLABUS BROWSER             │  │ SELECTED TOPIC            │  │
│  │                              │  │                           │  │
│  │ 🔍 Search syllabi...         │  │ ┌───────────────────────┐ │  │
│  │                              │  │ │ No topic selected     │ │  │
│  │ ┌──────────────────────────┐ │  │ │                       │ │  │
│  │ │ ▾ Malaysia / KSSM        │ │  │ │ Browse the curriculum │ │  │
│  │ │   ▾ Mathematics          │ │  │ │ tree and select a     │ │  │
│  │ │     ▾ Algebra            │ │  │ │ topic to contribute   │ │  │
│  │ │       ▸ F1-05 Expressions│ │  │ │ to.                   │ │  │
│  │ │       ▸ F1-06 Linear Eq. │ │  │ └───────────────────────┘ │  │
│  │ │       ▸ F1-07 Inequalit. │ │  │                           │  │
│  │ │     ▸ Geometry           │ │  │                           │  │
│  │ │     ▸ Statistics         │ │  │                           │  │
│  │ │ ▸ UK / Cambridge         │ │  │                           │  │
│  │ │ ▸ India / CBSE           │ │  │                           │  │
│  │ │ ▸ Singapore / MOE        │ │  │                           │  │
│  │ └──────────────────────────┘ │  │                           │  │
│  │                              │  │                           │  │
│  │ [+ Add new syllabus]         │  │                           │  │
│  └──────────────────────────────┘  └───────────────────────────┘  │
│                                                                   │
│                                            [Back]  [Next →]       │
└───────────────────────────────────────────────────────────────────┘
```

**After selecting a topic:**

```
┌────────────────────────────────┐ 
│  ┌───────────────────────────┐ │
│  │ SELECTED TOPIC            │ │
│  │                           │ │
│  │ Malaysia / KSSM           │ │
│  │ Mathematics > Algebra     │ │
│  │                           │ │
│  │ F1-06 Linear Equations    │ │
│  │ Persamaan Linear          │ │
│  │                           │ │
│  │ Quality: Level 2          │ │
│  │ ████████░░░░ Structured   │ │
│  │                           │ │
│  │ What's needed for Level 3:│ │
│  │ • Teaching notes          │ │
│  │ • ≥ 5 assessments         │ │
│  │ • Worked examples         │ │
│  │                           │ │
│  │ Existing content:         │ │
│  │ ✅ Learning objectives    │ │
│  │ ✅ Schema-valid structure │ │
│  │ ⬜ Teaching notes         │ │
│  │ ⬜ Assessments (0/5)      │ │
│  │ ⬜ Worked examples        │ │
│  └───────────────────────────┘ │
└────────────────────────────────┘
```

**Create New Curriculum sub-flow (triggered by "+ Add new syllabus"):**

```
┌──────────────────────────────────────────────────────────────────┐
│  ADD NEW CURRICULUM                                              │
│                                                                  │
│  Country:                                                        │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │ Select country or type a new one                        ▼ │  │
│  │ Malaysia                                                  │  │
│  │ United Kingdom                                            │  │
│  │ India                                                     │  │
│  │ Singapore                                                 │  │
│  │ + Add new country: "Kenya"                                │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  Syllabus name (e.g. KSSM, GCSE, CBSE, CBC, JEE):               │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │                                                            │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  Subject:                                                        │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │ Select subject or type a new one                        ▼ │  │
│  │ Mathematics                                               │  │
│  │ Physics                                                   │  │
│  │ Chemistry                                                 │  │
│  │ + Add new subject: "Biology"                              │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  Topic name (optional — can be scaffolded later):                │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │                                                            │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  This will scaffold a new curriculum directory structure in the  │
│  OSS repository. An educator will review the PR before merging. │
│                                                                  │
│                                  [Cancel]  [Create & continue →] │
└──────────────────────────────────────────────────────────────────┘
```

The "Create New Curriculum" flow supports adding entirely new countries, syllabi, and subjects — not just selecting from existing ones. The scaffolder generates the directory structure (`country/syllabus/subject/topic/`) with stub YAML files. After creation, the user continues through the contribution wizard with the new topic pre-selected.

**Interactions:**
- Tree nodes expand/collapse on click
- Search filters tree in real-time
- Selecting a topic shows quality summary in the right panel
- "+ Add new syllabus" opens the Create New Curriculum dialog (above)
- Country and subject fields support both selection from existing values and free-text entry for new ones
- "Create & continue" scaffolds the new curriculum path and pre-selects it in the tree
- "Next" is disabled until a topic is selected

---

### Step 2 — Choose Contribution Type

**Route:** `/contribute?step=2`
**Access:** Public

```
┌──────────────────────────────────────────────────────────────────┐
│  CONTRIBUTE                                                      │
│  "What would you like to contribute?"                            │
│  F1-06 Linear Equations                                ┌───────┐ │
│  Malaysia / KSSM / Mathematics / Algebra               │Step   │ │
│                                                        │2 of 5 │ │
│                                                        └───────┘ │
│  ● Select ─── ● Type ─── ○ Content ─── ○ Preview ─── ○ Submit    │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                                                             │ │
│  │  ┌────────────────────────┐  ┌────────────────────────┐     │ │
│  │  │ 📖 Teaching Notes      │  │ 📝 Assessments         │     │ │
│  │  │                        │  │                        │     │ │
│  │  │ Explanations, engage-  │  │ Practice questions     │     │ │
│  │  │ ment hooks, worked     │  │ with rubrics, hints,   │     │ │
│  │  │ examples, misconception│  │ and common wrong       │     │ │
│  │  │ warnings, teaching tips│  │ answers.               │     │ │
│  │  │                        │  │                        │     │ │
│  │  │ ⬜ Not yet contributed │  │ ⬜ 0 of 5 minimum      │     │ │
│  │  └────────────────────────┘  └────────────────────────┘     │ │
│  │                                                             │ │
│  │  ┌────────────────────────┐  ┌────────────────────────┐     │ │
│  │  │ 🔄 Translation         │  │ 🔧 Correction          │     │ │
│  │  │                        │  │                        │     │ │
│  │  │ Translate this topic   │  │ Fix errors in existing │     │ │
│  │  │ into another language. │  │ content: typos, inaccu-│     │ │
│  │  │ Preserves structure,   │  │ racies, or outdated    │     │ │
│  │  │ translates text.       │  │ information.           │     │ │
│  │  │                        │  │                        │     │ │
│  │  │ 🌐 8 languages so far  │  │ ✅ Existing content    │     │ │
│  │  └────────────────────────┘  └────────────────────────┘     │ │
│  │                                                             │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                  │
│                                        [← Back]  [Next →]        │
└──────────────────────────────────────────────────────────────────┘
```

**Interactions:**
- Cards highlight on hover with `ring-2 ring-sky-500`
- Click card to select (shows checkmark, sky border)
- Status badge shows what's needed vs what exists
- "Correction" only appears if topic has existing content
- "Next" is disabled until a type is selected
- For Translation, selecting "Next" shows a language picker inline

**Translation sub-step (shown inline when Translation is selected):**
```
┌────────────────────────────────────────┐
│  Target language:                      │
│  ┌──────────────────────────────────┐  │
│  │ Select language                 ▼│  │
│  │ Bahasa Melayu (ms)               │  │
│  │ العربية (ar)                     │  │
│  │ Español (es)                     │  │
│  │ हिन्दी (hi)                         │  │
│  │ 中文 (zh)                        │  │
│  │ தமிழ் (ta)                        │  │
│  └──────────────────────────────────┘  │
└────────────────────────────────────────┘
```

---

### Step 3 — Provide Content

**Route:** `/contribute?step=3`
**Access:** Public

This step supports three input methods: URL, Text, and Upload. Tabs switch between them.

```
┌───────────────────────────────────────────────────────────────────┐
│  CONTRIBUTE                                                       │
│  "Provide your content"                                           │
│  Teaching Notes for F1-06 Linear Equations             ┌───────┐  │
│  Write in any language. The AI structures your         │Step   │  │
│  input into schema-valid curriculum content.           │3 of 5 │  │
│                                                        └───────┘  │
│                                                                   │
│  ● Select ─── ● Type ─── ● Content ─── ○ Preview ─── ○ Submit     │
├───────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │  [🔗 URL]    [✏️ Text]    [📄 Upload]                         │ │
│  │                                                              │ │
│  │  ── Text tab (active) ──────────────────────────────────     │ │
│  │                                                              │ │
│  │  ┌──────────────────────────────────────────────────────┐    │ │
│  │  │                                                      │    │ │
│  │  │  Type or paste your content here. Write in any       │    │ │
│  │  │  language — the AI handles structuring.              │    │ │
│  │  │                                                      │    │ │
│  │  │  Examples:                                           │    │ │
│  │  │  • Teaching explanations and tips                    │    │ │
│  │  │  • Common student misconceptions you've observed     │    │ │
│  │  │  • Practice questions you use in class               │    │ │
│  │  │  • Worked examples with step-by-step solutions       │    │ │
│  │  │                                                      │    │ │
│  │  │                                                      │    │ │
│  │  │                                                      │    │ │
│  │  │                                                      │    │ │
│  │  │                                                      │    │ │
│  │  └──────────────────────────────────────────────────────┘    │ │
│  │  0 characters                              Min: 50 chars     │ │
│  │                                                              │ │
│  └──────────────────────────────────────────────────────────────┘ │
│                                                                   │
│                                        [← Back]  [Preview →]      │
└───────────────────────────────────────────────────────────────────┘
```

**URL tab:**

```
│  ── URL tab (active) ───────────────────────────────────────    │
│                                                                 │
│  ┌──────────────────────────────────────────────────────┐       │
│  │ https://                                             │       │
│  └──────────────────────────────────────────────────────┘       │
│                                                                 │
│  Paste a link to a curriculum page, syllabus PDF,               │
│  textbook publisher page, or any educational resource.          │
│                                                                 │
│  Good sources:                                                  │
│  • Government curriculum specification pages                    │
│  • Textbook publisher table-of-contents pages                   │
│  • University course outlines                                   │
│  • Online syllabus PDFs and documents                           │
│                                                                 │
│  [Fetch & extract]                                              │
│                                                                 │
│  ┌──────────────────────────────────────────────────────┐       │
│  │  ⏳ Fetching page content...                         │       │
│  │  Extracting curriculum structure...                  │       │
│  │  ████████████████░░░░░░░░░░  65%                     │       │
│  └──────────────────────────────────────────────────────┘       │
```

**Upload tab:**

```
│  ── Upload tab (active) ────────────────────────────────────    │
│                                                                 │
│  ┌──────────────────────────────────────────────────────┐       │
│  │                                                      │       │
│  │          ┌──────────┐                                │       │
│  │          │  📄      │                                │       │
│  │          └──────────┘                                │       │
│  │                                                      │       │
│  │     Drag and drop a file here, or click to browse    │       │
│  │                                                      │       │
│  │     PDF · DOCX · PPTX · TXT · PNG · JPG              │       │
│  │     Max file size: 25 MB                             │       │
│  │                                                      │       │
│  └──────────────────────────────────────────────────────┘       │
│                                                                 │
│  ┌──────────────────────────────────────────────────────┐       │
│  │  ☑ Use AI Vision (for handwriting, diagrams,         │       │
│  │    complex layouts)                                  │       │
│  └──────────────────────────────────────────────────────┘       │
│                                                                 │
│  After upload:                                                  │
│  ┌──────────────────────────────────────────────────────┐       │
│  │  📄 curriculum-spec.pdf                    ✕ Remove  │       │
│  │  2.3 MB · PDF · 12 pages                             │       │
│  │                                                      │       │
│  │  ⏳ Extracting content...                            │       │
│  │  ████████████░░░░░░░░░░░░░░  48%                     │       │
│  └──────────────────────────────────────────────────────┘       │
```

**Bulk Import Progress (for large files, 50+ pages):**

When a user uploads a large document (e.g., a full syllabus PDF), the portal shows a multi-stage progress UI instead of a simple spinner:

```
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  📄 national-curriculum-maths.pdf                  ✕ Remove  │  │
│  │  8.7 MB · PDF · 124 pages                                   │  │
│  │                                                              │  │
│  │  ┌──────────────────────────────────────────────────────┐    │  │
│  │  │                                                      │    │  │
│  │  │  ✅ Upload complete                                  │    │  │
│  │  │  ████████████████████████████████  100%               │    │  │
│  │  │                                                      │    │  │
│  │  │  ✅ Text extraction                                  │    │  │
│  │  │  ████████████████████████████████  100%               │    │  │
│  │  │                                                      │    │  │
│  │  │  ✅ Analyzing structure...                           │    │  │
│  │  │  Found 12 topics across 4 subjects                   │    │  │
│  │  │                                                      │    │  │
│  │  │  ⏳ Generating content...                            │    │  │
│  │  │  ████████████████░░░░░░░░░░░░░░░░  Generating        │    │  │
│  │  │  topic 3 of 12: "Quadratic Equations"                │    │  │
│  │  │                                                      │    │  │
│  │  │  ○ Final review                                      │    │  │
│  │  │                                                      │    │  │
│  │  └──────────────────────────────────────────────────────┘    │  │
│  └──────────────────────────────────────────────────────────────┘  │
```

After bulk generation completes, the preview (Step 4) shows all generated files in an accordion:

```
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │  BULK IMPORT RESULTS — 12 topics generated                   │  │
│  │                                                              │  │
│  │  ▾ Algebra / Quadratic Equations          L3 ✅ Valid        │  │
│  │    (expanded: rendered preview + YAML)                       │  │
│  │                                                              │  │
│  │  ▸ Algebra / Simultaneous Equations       L3 ✅ Valid        │  │
│  │  ▸ Algebra / Inequalities                 L2 ⚠ 1 warning    │  │
│  │  ▸ Geometry / Triangles                   L3 ✅ Valid        │  │
│  │  ▸ Geometry / Circles                     L3 ✅ Valid        │  │
│  │  ▸ ... (7 more)                                             │  │
│  │                                                              │  │
│  │  Summary: 11 passed · 1 warning · 0 errors                  │  │
│  │                                                              │  │
│  │                              [← Edit]  [Submit all →]        │  │
│  └──────────────────────────────────────────────────────────────┘  │
```

Progress updates are streamed in real-time via SSE (see [Technical Notes](#technical-notes) below).

**Interactions:**
- Tab switch retains previously entered content (user can combine URL + text)
- Text area shows placeholder hints that change by contribution type
- Character counter turns amber below minimum, emerald when sufficient
- Upload shows progress bar during extraction
- For large files (50+ pages), the multi-stage bulk import progress UI is shown automatically
- "Use AI Vision" toggle appears only for image files
- "Preview" button triggers AI generation pipeline

---

### Step 4 — Preview & Validate

**Route:** `/contribute?step=4`
**Access:** Public

```
┌──────────────────────────────────────────────────────────────────┐
│  CONTRIBUTE                                                      │
│  "Preview your contribution"                                     │
│  Teaching Notes for F1-06 Linear Equations             ┌───────┐ │
│  Review the structured output before submitting.       │Quality│ │
│                                                        │Level 3│ │
│                                                        └───────┘ │
│  ● Select ─── ● Type ─── ● Content ─── ● Preview ─── ○ Submit    │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐  │
│  │ Quality    │  │ Validation │  │ Bloom's    │  │ Duplicates │  │
│  │ Level 3    │  │ ✅ Passed  │  │ ✅ Valid   │  │ ✅ None     │  │
│  │ Teachable  │  │ Schema OK  │  │ Levels OK  │  │ All unique │  │
│  └────────────┘  └────────────┘  └────────────┘  └────────────┘  │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │  [Rendered]  [YAML Source]                                  │ │
│  │                                                             │ │
│  │  ── Rendered view (active) ──────────────────────────────   │ │
│  │                                                             │ │
│  │  TEACHING NOTES                                             │ │
│  │  F1-06 Linear Equations                                     │ │
│  │                                                             │ │
│  │  Engagement Hook                                            │ │
│  │  ┌──────────────────────────────────────────────────────┐   │ │
│  │  │ "Imagine you and your friend go to the canteen.      │   │ │
│  │  │  Together you spend RM12 on food. Your friend spent  │   │ │
│  │  │  RM2 more than you. How much did each of you spend?" │   │ │
│  │  └──────────────────────────────────────────────────────┘   │ │
│  │                                                             │ │
│  │  Key Concepts                                               │ │
│  │  • Variable as unknown: representing quantities with        │ │
│  │    letters (x, y) to form equations                         │ │
│  │  • Balancing principle: performing the same operation       │ │
│  │    on both sides maintains equality                         │ │
│  │  • Solution verification: substituting back to check        │ │
│  │                                                             │ │
│  │  Common Misconceptions                                      │ │
│  │  ┌──────────────────────────────────────────────────────┐   │ │
│  │  │ ⚠ "Moving to the other side changes the sign"        │   │ │
│  │  │   Students think terms teleport. Reinforce that      │   │ │
│  │  │   we perform inverse operations on both sides.       │   │ │
│  │  └──────────────────────────────────────────────────────┘   │ │
│  │  ┌──────────────────────────────────────────────────────┐   │ │
│  │  │ ⚠ "x = 3 means x is always 3"                        │   │ │
│  │  │   Students confuse solution with identity. Show that │   │ │
│  │  │   x = 3 only satisfies this specific equation.       │   │ │
│  │  └──────────────────────────────────────────────────────┘   │ │
│  │                                                             │ │
│  │  ... (scrollable)                                           │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                  │
│                                  [← Edit]  [Submit →]            │
└──────────────────────────────────────────────────────────────────┘
```

**YAML Source tab:**

```
│  ── YAML Source ────────────────────────────────────────────    │
│                                                                 │
│  ┌──────────────────────────────────────────────────────┐       │
│  │  1 │ # Teaching Notes — F1-06 Linear Equations       │       │
│  │  2 │ engagement_hook:                                │       │
│  │  3 │   text: "Imagine you and your friend go to..."  │       │
│  │  4 │   type: real_world_scenario                     │       │
│  │  5 │                                                 │       │
│  │  6 │ key_concepts:                                   │       │
│  │  7 │   - name: Variable as unknown                   │       │
│  │  8 │     description: "Representing quantities..."   │       │
│  │  9 │   - name: Balancing principle                   │       │
│  │ 10 │     description: "Performing the same..."       │       │
│  │ ...│                                                 │       │
│  └──────────────────────────────────────────────────────┘       │
│                                                                 │
│  Syntax highlighting: code block with line numbers              │
│  Font: JetBrains Mono or system monospace                       │
```

**Validation failure state:**

```
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐  │
│  │ Quality    │  │ Validation │  │ Bloom's    │  │ Duplicates │  │
│  │ Level 1    │  │ ❌ Failed  │  │ ⚠ Warning  │  │ ✅ None     │  │
│  │ Basic      │  │ 2 errors   │  │ 1 warning  │  │ All unique │  │
│  └────────────┘  └────────────┘  └────────────┘  └────────────┘  │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │ VALIDATION ERRORS                                           │ │
│  │                                                             │ │
│  │  ❌ Missing required field: engagement_hook.type            │ │
│  │     Expected one of: real_world_scenario, puzzle,           │ │
│  │     challenge, curiosity                                    │ │
│  │                                                             │ │
│  │  ❌ key_concepts[2].description exceeds 500 characters      │ │
│  │     Current: 523 characters                                 │ │
│  │                                                             │ │
│  │  ⚠ Assessment Q3 uses "Evaluate" verb but is tagged as      │ │
│  │    difficulty:easy (Bloom's level mismatch)                 │ │
│  │                                                             │ │
│  │                           [← Back to edit]                  │ │
│  └─────────────────────────────────────────────────────────────┘ │
```

**Interactions:**
- "Rendered" and "YAML Source" tabs toggle view
- Validation errors are shown inline with line references
- "Edit" returns to Step 3 with content preserved
- "Submit" is disabled if blocking validation errors exist (warnings are OK)
- Quality level badge uses color coding: emerald (3+), amber (1-2), rose (0)

---

### Step 5 — Submit

**Route:** `/contribute?step=5`
**Access:** Public

```
┌───────────────────────────────────────────────────────────────────┐
│  CONTRIBUTE                                                       │
│  "Contribution submitted"                                         │
│  Teaching Notes for F1-06 Linear Equations             ┌───────┐  │
│                                                        │  ✅   │  │
│                                                        └───────┘  │
│  ● Select ─── ● Type ─── ● Content ─── ● Preview ─── ● Submit     │
├───────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │                                                              │ │
│  │                    ✅ Contribution submitted!                │ │
│  │                                                              │ │
│  │  Your content has been structured, validated, and submitted  │ │
│  │  as a pull request to the Open School Syllabus repository.   │ │
│  │                                                              │ │
│  │  ┌──────────────────────────────────────────────────────┐    │ │
│  │  │  PR #247                                             │    │ │
│  │  │  "Add teaching notes for F1-06 Linear Equations"     │    │ │
│  │  │                                                      │    │ │
│  │  │  Branch: oss-bot/teaching-notes-F1-06-20260326T1400  │    │ │
│  │  │  Quality: Level 3 (Teachable)                        │    │ │
│  │  │  Provenance: ai-assisted                             │    │ │
│  │  │  Reviewer: Requested (CODEOWNERS)                    │    │ │
│  │  │                                                      │    │ │
│  │  │  [View on GitHub ↗]                                  │    │ │
│  │  └──────────────────────────────────────────────────────┘    │ │
│  │                                                              │ │
│  │  What happens next:                                          │ │
│  │  1. An educator with subject expertise will review your PR   │ │
│  │  2. They may approve, request changes, or leave feedback     │ │
│  │  3. Once approved, your contribution is merged into OSS      │ │
│  │                                                              │ │
│  │  ┌──────────────────────────────────────┐                    │ │
│  │  │  Sign in with GitHub to track your   │                    │ │
│  │  │  contribution and get notified of    │                    │ │
│  │  │  review status.                      │                    │ │
│  │  │                                      │                    │ │
│  │  │  [Sign in with GitHub]               │                    │ │
│  │  └──────────────────────────────────────┘                    │ │
│  │                                                              │ │
│  │  [← Contribute more]         [View all contributions →]      │ │
│  │                                                              │ │
│  └──────────────────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────────────────┘
```

**Submitting state (shown briefly before success):**

```
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │                                                              │ │
│  │                    ⏳ Submitting...                          │ │
│  │                                                              │ │
│  │  Creating branch...                              ✅          │ │
│  │  Committing files with provenance metadata...    ⏳          │ │
│  │  Opening pull request...                         ○           │ │
│  │  Requesting reviewers...                         ○           │ │
│  │                                                              │ │
│  └──────────────────────────────────────────────────────────────┘ │
```

**Interactions:**
- "View on GitHub" opens PR in new tab
- "Contribute more" resets wizard to Step 1
- "Sign in with GitHub" triggers OAuth flow
- If user is already signed in, attribution line shows their GitHub handle

---

## Contribution History

**Route:** `/contributions`
**Access:** Requires GitHub sign-in

```
┌───────────────────────────────────────────────────────────────────┐
│  MY CONTRIBUTIONS                                                 │
│  "Your curriculum contributions"                                  │
│  Track the status of your pull requests and see        ┌───────┐  │
│  your impact on the Open School Syllabus.              │Total  │  │
│                                                        │  12   │  │
│                                                        │merged │  │
│                                                        └───────┘  │
├───────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌────────────┐   │
│  │ Total      │  │ Merged     │  │ In Review  │  │ Quality    │   │
│  │   15       │  │    12      │  │     2      │  │ Avg L3.2   │   │
│  │ PRs created│  │ Accepted   │  │ Awaiting   │  │ Across all │   │
│  └────────────┘  └────────────┘  └────────────┘  └────────────┘   │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │  [All]  [In Review]  [Merged]  [Changes Requested]          │  │
│  │                                                             │  │
│  │  Title               │ Topic    │ Type     │ Quality│ Status│  │
│  │  ────────────────────┼──────────┼──────────┼────────┼───────│  │
│  │  Add teaching notes  │ F1-06    │ Teaching │ L3     │ 🟢 M  │  │
│  │  5 assessments for   │ F1-05    │ Assess.  │ L3     │ 🟡 R  │  │
│  │  Translate F2-01 to  │ F2-01    │ Transl.  │ L4     │ 🟢 M  │  │
│  │  Import CBSE Math    │ multiple │ Import   │ L2     │ 🟡 R  │  │
│  │  Fix typo in F1-07   │ F1-07    │ Correct. │ —      │ 🟢 M  │  │
│  │                                                             │  │
│  │  Status: 🟢 Merged  🟡 In Review  🔴 Changes Requested       │  │
│  │                                                             │  │
│  │  Click row → opens GitHub PR in new tab                     │  │
│  └─────────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────────┘
```

---

## Quality Report View

**Route:** `/quality`
**Access:** Public

```
┌───────────────────────────────────────────────────────────────────┐
│  QUALITY REPORTS                                                  │
│  "Syllabus quality overview"                                      │
│  See which topics need contributions and what's        ┌───────┐  │
│  required to reach the next quality level.             │Avg    │  │
│                                                        │Level  │  │
│                                                        │ 2.4   │  │
│                                                        └───────┘  │
├───────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │  Country:  [All ▼]  Syllabus:  [All ▼]  Subject:  [All ▼]    │ │
│  │                                                              │ │
│  │  QUALITY HEATMAP                                             │ │
│  │                                                              │ │
│  │  Topic                    │Obj│Notes│Assess│Examp│Trans│ Lvl │ │
│  │  ─────────────────────────┼───┼─────┼──────┼─────┼─────┼─────│ │
│  │  F1-05 Expressions        │ ✅│  ✅ │  ✅   │  ✅ │  ✅ │  5  │ │
│  │  F1-06 Linear Equations   │ ✅│  ⬜ │  ⬜   │  ⬜ │  ⬜ │  2  │ │
│  │  F1-07 Inequalities       │ ✅│  ✅ │  ✅   │  ⬜ │  ⬜ │  3  │ │
│  │  F2-01 Patterns           │ ✅│  ✅ │  ✅   │  ✅ │  ⬜ │  4  │ │
│  │  F2-02 Factorisation      │ ✅│  ⬜ │  ⬜   │  ⬜ │  ⬜ │  2  │ │
│  │  F2-03 Formulae           │ ⬜│  ⬜ │  ⬜   │  ⬜ │  ⬜ │  0  │ │
│  │  F3-01 Indices            │ ✅│  ✅ │  ✅   │  ✅ │  ✅ │  5  │ │
│  │  F3-09 Straight Lines     │ ✅│  ✅ │  ⬜   │  ⬜ │  ⬜ │  2  │ │
│  │                                                             │ │
│  │  Level colors: ■ 5 emerald  ■ 4 lime  ■ 3 sky               │ │
│  │                ■ 2 amber    ■ 1 amber  ■ 0 rose             │ │
│  │                                                             │ │
│  │  Click topic → opens contribution wizard (Step 1 prefilled) │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │ CONTRIBUTION OPPORTUNITIES (dark card)                      │ │
│  │                                                             │ │
│  │ Topics most in need of contributions:                       │ │
│  │                                                             │ │
│  │ 1. F2-03 Formulae — Level 0 (stub only, needs everything)   │ │
│  │    [Contribute →]                                           │ │
│  │                                                             │ │
│  │ 2. F1-06 Linear Equations — Level 2 (needs teaching notes)  │ │
│  │    [Contribute →]                                           │ │
│  │                                                             │ │
│  │ 3. F3-09 Straight Lines — Level 2 (needs assessments)       │ │
│  │    [Contribute →]                                           │ │
│  └─────────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────────┘
```

**Interactions:**
- Country / Syllabus / Subject dropdowns filter the heatmap (cascade: selecting a country narrows syllabus options)
- Supports browsing any curriculum worldwide (e.g., Malaysia KSSM, India CBSE, UK GCSE, Kenya CBC)
- Click topic row → navigates to `/contribute` with topic pre-selected
- "Contribute" links pre-fill the wizard with the suggested contribution type
- Heatmap cells are colored by quality level

---

## Shared Components

### PageHero

Identical pattern to P&AI Bot admin panel. Used at the top of every page.

```
┌───────────────────────────────────────────────────────────────────┐
│  EYEBROW LABEL                                                    │
│  "Page Title"                                                     │
│  Description text that explains what this page      ┌───────────┐ │
│  shows and what actions are available.              │ Dark aside│ │
│                                                     │ with key  │ │
│  [Optional action buttons]                          │ metric    │ │
│                                                     └───────────┘ │
│  [Optional child content like breadcrumb links]                   │
└───────────────────────────────────────────────────────────────────┘
```

### StatCard

Compact metric display, same as P&AI Bot.

```
┌─────────────────┐
│ 📊 Title        │
│    42           │
│ Explanatory note│
└─────────────────┘
```

### StatePanel

Loading, empty, and error states within cards.

```
Loading:                     Empty:                      Error:
┌──────────────────┐   ┌──────────────────┐   ┌──────────────────┐
│                  │   │                  │   │                  │
│  ⏳ Generating...│   │  📭 No data yet  │   │  ⚠ Generation     │
│  Structuring your│   │  Select a topic  │   │  failed          │
│  contribution... │   │  to get started. │   │  [Retry]         │
│                  │   │                  │   │                  │
└──────────────────┘   └──────────────────┘   └──────────────────┘
```

### StepIndicator

Horizontal step progress bar (described above in Contribution Flow).

### QualityBadge

Inline colored badge showing quality level.

```
Level 5:  [ L5 Excellent  ] emerald bg
Level 4:  [ L4 Complete   ] lime bg
Level 3:  [ L3 Teachable  ] sky bg
Level 2:  [ L2 Structured ] amber bg
Level 1:  [ L1 Basic      ] amber bg
Level 0:  [ L0 Stub       ] rose bg
```

### ValidationStatus

Inline status indicator for validation checks.

```
✅ Passed:  [ ✅ Schema valid ] emerald text
⚠ Warning:  [ ⚠ Bloom's mismatch ] amber text
❌ Failed:   [ ❌ 2 schema errors ] rose text
```

---

## Responsive Behavior

| Breakpoint | Behavior |
|------------|----------|
| `< 768px` (mobile) | Single column, hamburger menu, full-width cards, stacked input tabs |
| `768px–1023px` (tablet) | 2-column stat grids, side-by-side syllabus browser + topic panel |
| `≥ 1024px` (desktop) | Full layout, 2-column contribution flow, 4-column stat grids |
| `≥ 1280px` (xl) | Wider content area, quality heatmap without horizontal scroll |

### Key responsive patterns:
- **Stat cards:** `grid md:grid-cols-2 xl:grid-cols-4`
- **Two-panel layouts (syllabus browser):** `grid xl:grid-cols-[1fr_1fr]`, stacks on mobile
- **Contribution type cards:** `grid md:grid-cols-2`, stacks on mobile
- **Quality heatmap:** Horizontal scroll on mobile (`overflow-x-auto`, `min-w-[700px]`)
- **Step indicator:** Horizontal on desktop, vertical on mobile (`flex-col` below `md`)
- **Input tabs:** Full-width tabs on mobile

---

## Interaction Patterns

### Contribution Flow Navigation
- Steps are navigable via "Back" / "Next" buttons at bottom of each step
- Step indicator allows clicking completed steps to return (but not skipping ahead)
- Browser back/forward works within the wizard (URL query param `?step=N`)
- Unsaved content triggers a browser `beforeunload` confirmation

### AI Generation Loading
1. User clicks "Preview" on Step 3
2. Loading overlay with progress steps:
   - "Building context..." (context builder loads topic + neighbors)
   - "Generating content..." (AI provider call, streaming)
   - "Validating..." (schema + Bloom's + duplicates)
3. On success → transition to Step 4 (Preview)
4. On failure → error message with "Retry" and "Edit input" options

For bulk imports (large files with multiple topics), the progress UI shows per-topic generation status ("Generating topic 3/12..."). All progress updates are delivered via SSE (see [Technical Notes](#technical-notes)).

### File Upload
1. User drops file or clicks browse
2. File type validated client-side (extension + MIME)
3. File uploaded to Go backend
4. Progress bar shows extraction progress
5. For images: auto-detection of OCR vs AI Vision (toggle to override)
6. Extracted text shown in collapsible "Extracted content" section
7. User can edit extracted text before proceeding

### GitHub Sign-in (Optional)
- OAuth flow via GitHub App
- Minimal scopes: `read:user` only
- On sign-in: PR attribution includes GitHub handle
- "My Contributions" page becomes accessible
- Sign-in state persisted in `localStorage`

### Theme
- System preference detected on load
- Manual toggle via sun/moon icon in top nav
- Preference stored in `localStorage`
- All components support light and dark via Tailwind `dark:` classes

### Data Loading
- Client components use TanStack Query with loading/error states
- Syllabus tree loaded once, cached for session
- Preview generation uses streaming response for progress updates via SSE
- Every data section has empty, loading, and error states via `StatePanel`

---

## Technical Notes

### Progress Streaming via SSE

All long-running operations (AI generation, bulk import, file extraction) stream real-time progress to the frontend using **Server-Sent Events (SSE)**.

**Endpoint:** `GET /api/progress/:jobId`

**Flow:**
1. Client initiates a generation or import via `POST /api/preview` or `POST /api/submit`
2. Server responds immediately with `{ "jobId": "abc123" }`
3. Client opens an SSE connection to `GET /api/progress/abc123`
4. Server pushes events as the pipeline progresses:

```
event: progress
data: {"stage": "upload", "percent": 100, "message": "Upload complete"}

event: progress
data: {"stage": "extraction", "percent": 65, "message": "Extracting text..."}

event: progress
data: {"stage": "analysis", "percent": 100, "message": "Found 12 topics across 4 subjects"}

event: progress
data: {"stage": "generation", "percent": 25, "current": 3, "total": 12, "message": "Generating topic 3/12: Quadratic Equations"}

event: progress
data: {"stage": "validation", "percent": 100, "message": "Validation complete"}

event: complete
data: {"result": { ... }}
```

5. Client closes the SSE connection on `complete` or `error` event

**Stages:** `upload` | `extraction` | `analysis` | `generation` | `validation` | `submission`

The SSE approach avoids polling, provides instant UI updates, and works well with the existing TanStack Query setup (via `useEventSource` or a custom hook). For single-topic generation, stages progress quickly. For bulk imports (50+ pages, multiple topics), the `generation` stage reports per-topic progress.

### Global Curriculum Scope

The portal is designed for curricula from around the world. The curriculum browser uses a **Country -> Syllabus -> Subject -> Topic** hierarchy. Examples include:

- Malaysia / KSSM / Mathematics / Algebra
- India / CBSE / Physics / Mechanics
- United Kingdom / GCSE / Chemistry / Organic Chemistry
- Kenya / CBC / Mathematics / Number Patterns
- Singapore / MOE / Mathematics / Calculus

Users can contribute to any existing curriculum or create entirely new ones via the "Create New Curriculum" flow (see [Step 1](#step-1--select-syllabus--topic)).

---

## Tech Stack Summary

Aligned with P&AI Bot admin panel for ecosystem consistency:

| Component | Technology | Version | Notes |
|-----------|-----------|---------|-------|
| **Framework** | Next.js (App Router) | 16 | Matches P&AI Bot admin panel |
| **Language** | TypeScript | 5.x | Type safety, form handling |
| **UI Components** | shadcn/ui | latest | Same component library as P&AI Bot |
| **Styling** | Tailwind CSS | 3.x | Same utility classes, shared design tokens |
| **Form Handling** | React Hook Form + Zod | latest | Step wizard validation |
| **State** | TanStack Query | v5 | Server state for preview and submission |
| **Code Highlighting** | Shiki or Prism | latest | YAML syntax highlighting in preview |
| **Auth** | GitHub OAuth (optional) | — | Via GitHub App, `read:user` scope |

**Not used (differs from P&AI Bot):**
- **Refine** — not needed; the portal is a linear contribution wizard, not a CRUD admin panel
- **Recharts/Tremor** — no charting needed; quality heatmap uses HTML table with colored cells

---

## File Reference

| Component / Page | File Path |
|------------------|-----------|
| Contribution Shell | `web/src/components/contribution-shell.tsx` |
| Landing Page | `web/src/app/page.tsx` |
| Contribution Wizard | `web/src/app/contribute/page.tsx` |
| Quality Reports | `web/src/app/quality/page.tsx` |
| My Contributions | `web/src/app/contributions/page.tsx` |
| Syllabus Picker | `web/src/components/syllabus-picker.tsx` |
| Create New Curriculum | `web/src/components/create-curriculum-dialog.tsx` |
| Topic Picker | `web/src/components/topic-picker.tsx` |
| Contribution Form | `web/src/components/contribution-form.tsx` |
| YAML Preview | `web/src/components/yaml-preview.tsx` |
| Step Indicator | `web/src/components/step-indicator.tsx` |
| Quality Badge | `web/src/components/quality-badge.tsx` |
| Validation Status | `web/src/components/validation-status.tsx` |
| File Upload | `web/src/components/file-upload.tsx` |
| Bulk Import Progress | `web/src/components/bulk-import-progress.tsx` |
| Submission Status | `web/src/components/submission-status.tsx` |
| SSE Progress Hook | `web/src/lib/use-progress.ts` |
| API Client | `web/src/lib/api.ts` |
| Auth (GitHub OAuth) | `web/src/lib/auth.ts` |
| Shared Components | `web/src/components/` (page-hero, stat-card, state-panel) |
| shadcn/ui Components | `web/src/components/ui/` |
