# OSS ID Conventions

This document defines the canonical identifier formats for all entities in the Open School Syllabus (OSS) data model. All contributors, tools, and AI-generated content **must** follow these conventions for consistency and cross-referencing.

---

## Language Principle

> **IDs use the official language of the country's Ministry of Education. English is always included in the content for interoperability.**

This means:
- A Malaysian KSSM subject ID is `malaysia-kssm-matematik-tingkatan-1` (Malay — KPM's official language), not `mathematics-form-1`
- The YAML file at that path includes `name: "Matematik"` **and** `name_en: "Mathematics"` for tools and integrations that need English
- An Indian CBSE subject ID is `india-cbse-mathematics-class-10` (English — CBSE's official language)

The bot follows this principle when generating IDs and scaffolding new content.

### Official MOE Languages by Country

| Country | Official MOE Language | Example Subject ID |
|---------|----------------------|-------------------|
| Malaysia | Malay (Bahasa Melayu) | `malaysia-kssm-matematik-tingkatan-3` |
| Indonesia | Indonesian (Bahasa Indonesia) | `indonesia-k13-matematika-kelas-10` |
| India | English | `india-cbse-mathematics-class-10` |
| Singapore | English | `singapore-o-level-mathematics` |
| UK | English | `uk-cambridge-igcse-mathematics-0580` |
| Australia | English | `australia-ac-mathematics-year-10` |
| Nigeria | English | `nigeria-waec-mathematics` |
| UAE | Arabic | `uae-moe-riyaadiyaat-saff-7` |
| Japan | Japanese | `japan-mext-suugaku-chugaku-1` |

For countries not listed, use the primary language of instruction as declared by the national MOE.

---

## Summary

| Entity | Format | Example |
|--------|--------|---------|
| Country | `{name}` | `malaysia`, `india`, `uk` |
| Syllabus | `{country}-{board}` | `malaysia-kssm`, `india-cbse` |
| Grade | `{grade-name}` in MOE language | `tingkatan-1`, `class-10`, `year-9` |
| Subject | `{syllabus}-{subject}-{grade}` | `malaysia-kssm-matematik-tingkatan-1` |
| Topic | `{PREFIX}{grade_num}-{NN}` | `MT1-01`, `SA2-03`, `PHY12-01` |
| Learning Objective | `LO{N}` | `LO1`, `LO2` |
| Assessment Question | `Q{N}` | `Q1`, `Q5` |
| Worked Example | `WE-{NN}` | `WE-01`, `WE-12` |

All IDs use **lowercase kebab-case** (hyphens, no underscores, no spaces) except Topic IDs which use an uppercase prefix.

---

## Country ID

**Format:** `{common-english-name}` — lowercase, kebab-case

Country IDs are always in English (they identify the country, not a language):

```
malaysia
india
uk
us
singapore
indonesia
australia
nigeria
south-africa
new-zealand
```

Use the common English name, not the ISO code. Multi-word countries use hyphens.

---

## Syllabus ID

**Format:** `{country}-{board}` or `{country}-{board}-{curriculum}`

The board abbreviation uses the MOE's own abbreviation for the curriculum:

```
malaysia-kssm          # Kurikulum Standard Sekolah Menengah
malaysia-kssm-v2       # Revised KSSM (if versioned)
malaysia-uec           # Unified Examination Certificate
india-cbse             # Central Board of Secondary Education
india-jee              # Joint Entrance Examination
indonesia-k13          # Kurikulum 2013
uk-cambridge-igcse     # Cambridge IGCSE
uk-aqa-gcse            # AQA GCSE
singapore-o-level      # Singapore-Cambridge O-Level
us-common-core         # US Common Core
```

Rules:
- Use the board's own abbreviation in the MOE language where it has one (`kssm` not `ssmk`)
- If the board has multiple curricula, append the curriculum level: `uk-cambridge-igcse` vs `uk-cambridge-alevel`
- Never include subject or grade in the syllabus ID

---

## Grade ID

**Format:** `{grade-name}` — in the MOE's official language, slugified

```
# Malaysia KSSM (Malay — KPM official)
tingkatan-1
tingkatan-2
tingkatan-3
tingkatan-4
tingkatan-5

# Indonesia Kurikulum 2013 (Indonesian — Kemendikbud official)
kelas-7
kelas-8
kelas-10

# India CBSE (English — CBSE official)
class-9
class-10
class-11
class-12

# UK Cambridge IGCSE (English — no grade, use assessment tier)
core
extended

# UK GCSEs (English)
year-10
year-11

# Japan MEXT (Japanese — MEXT official)
chugaku-1      # 中学1年 = Junior High Year 1
chugaku-2
koko-1         # 高校1年 = Senior High Year 1
```

Rules:
- Always use the MOE's official language term for the grade (e.g. `tingkatan` not `form` for KSSM)
- Slugify: lowercase, spaces and special chars become hyphens
- For exam-based syllabi with no grade structure (e.g. JEE), omit the grade from the subject ID

---

## Subject ID

**Format:** `{syllabus-id}-{subject}-{grade-id}`

Subject names use the **official MOE language**:

```
# Malaysia KSSM — Malay subject names (KPM official)
malaysia-kssm-matematik-tingkatan-1
malaysia-kssm-matematik-tingkatan-3
malaysia-kssm-sains-tingkatan-2
malaysia-kssm-fizik-tingkatan-4
malaysia-kssm-kimia-tingkatan-4
malaysia-kssm-biologi-tingkatan-4
malaysia-kssm-sejarah-tingkatan-1
malaysia-kssm-bahasa-melayu-tingkatan-3

# Indonesia K13 — Indonesian subject names (Kemendikbud official)
indonesia-k13-matematika-kelas-10
indonesia-k13-fisika-kelas-11
indonesia-k13-biologi-kelas-10

# India CBSE — English subject names (CBSE official)
india-cbse-mathematics-class-10
india-cbse-physics-class-12
india-cbse-chemistry-class-11

# UK Cambridge IGCSE — English subject names + code
uk-cambridge-igcse-mathematics-0580
uk-cambridge-igcse-physics-0625
```

Rules:
- Use the official subject name from the MOE curriculum document, slugified
- Keep the syllabus subject code if the board uses one (e.g. `0580` for IGCSE Maths)
- Grade is always the last component

---

## Topic ID

**Format:** `{PREFIX}{grade_num}-{NN}`

- **PREFIX** — 2 uppercase letters derived from the **English subject name** (language-neutral, for tooling)
- **grade_num** — the grade number only (e.g. `1` for Tingkatan 1, `12` for Class 12)
- **NN** — 2-digit zero-padded sequence number within the subject+grade, matching chapter/topic order in the official syllabus document

```
# Malaysia KSSM Matematik Tingkatan 1 → prefix MT, grade 1
MT1-01    # Chapter 1
MT1-02    # Chapter 2
MT1-15    # Chapter 15

# Malaysia KSSM Matematik Tingkatan 3 → prefix MT, grade 3
MT3-01
MT3-09

# Malaysia KSSM Sains Tingkatan 2 → prefix SA, grade 2
SA2-01
SA2-07

# Malaysia KSSM Fizik Tingkatan 4 → prefix PH, grade 4
PH4-01

# India CBSE Physics Class 12 → prefix PH, grade 12
PH12-01

# Cambridge IGCSE Mathematics (no grade) → prefix MT
MT-01
MT-14
```

**Prefix table** (derived from English name — language-neutral):

| Subject (any language) | English equivalent | Prefix |
|------------------------|-------------------|--------|
| Matematik, Matematika, Mathematics | Mathematics | `MT` |
| Sains, Ilmu Pengetahuan Alam, Science | Science | `SC` |
| Fizik, Fisika, Physics | Physics | `PH` |
| Kimia, Kimia, Chemistry | Chemistry | `CH` |
| Biologi, Biologi, Biology | Biology | `BI` |
| Sejarah, History | History | `HI` |
| Geografi, Geography | Geography | `GE` |
| Bahasa Melayu | Malay Language | `BM` |
| Indonesian | Indonesian Language | `ID` |
| English Language | English | `EN` |
| Bahasa Arab, Arabic | Arabic | `AR` |
| Other | First 2 consonants of English name | — |

Topic prefixes are always derived from English to keep them consistent across languages. This means `MT3-01` is unambiguous whether the subject is called "Matematik" (Malaysia), "Matematika" (Indonesia), or "Mathematics" (India).

---

## Official Reference (`official_ref`)

MOEs and exam boards often assign their own codes, chapter numbers, or reference strings to topics in official curriculum documents. These are captured in the optional `official_ref` field on every topic YAML.

`official_ref` is **read-only provenance** — it records the board's identifier as printed in the source document. It never replaces the OSS `id` and is never used for cross-referencing between topics (use `id` for that).

### Examples

| Board | Source reference | `official_ref` value |
|-------|-----------------|----------------------|
| Malaysia KSSM | "Bab 9: Garis Lurus" | `"Bab 9"` |
| Malaysia KSSM | Topic code printed on past paper | `"F4-T9"` |
| India CBSE | "Chapter 12: Linear Programming" | `"Chapter 12"` |
| Cambridge IGCSE | Section code in syllabus | `"C2.5"` |
| Singapore O-Level | Syllabus section number | `"2.3"` |
| Japan MEXT | Unit reference in 学習指導要領 | `"第3章第2節"` |
| Nigeria WAEC | WAEC syllabus item code | `"SS2/MAT/03"` |

### Rules

- Store the reference **exactly as it appears** in the official document — do not translate or normalise it.
- If the board uses multiple reference formats (e.g. a chapter number *and* a paper code), pick the most stable/official one. Add the others as a YAML list: `official_ref: ["C2.5", "0580/C2.5"]`.
- Omit `official_ref` entirely if the source document has no formal identifier for the topic.
- `official_ref` is a string (or list of strings) — never a structured object.

### Topic YAML with `official_ref`

```yaml
id: MT3-09
official_ref: "Bab 9"              # as printed in KSSM document; omit if absent
name: Garis Lurus
name_en: Straight Lines
subject_id: malaysia-kssm-matematik-tingkatan-3
syllabus_id: malaysia-kssm
country_id: malaysia
language: ms
```

```yaml
id: MT-05
official_ref: "C2.5"              # Cambridge IGCSE 0580 syllabus section
name: Coordinate Geometry
name_en: Coordinate Geometry
subject_id: uk-cambridge-igcse-mathematics-0580
syllabus_id: uk-cambridge-igcse
country_id: uk
language: en
```

---

## English in Content

Every entity YAML includes English alongside the official language. This enables integrations, search, and cross-country mapping without requiring translation.

### Subject YAML

```yaml
id: malaysia-kssm-matematik-tingkatan-3
name: Matematik Tingkatan 3         # official MOE language
name_en: Mathematics Form 3         # English for interoperability
syllabus_id: malaysia-kssm
grade_id: tingkatan-3
country_id: malaysia
language: ms                        # BCP 47 language code
```

### Topic YAML

```yaml
id: MT3-09
official_ref: "Bab 9"              # board's own chapter/topic code; omit if absent
name: Garis Lurus                   # official MOE language
name_en: Straight Lines             # English for interoperability
subject_id: malaysia-kssm-matematik-tingkatan-3
syllabus_id: malaysia-kssm
country_id: malaysia
language: ms

learning_objectives:
  - id: LO1
    text: Menentukan kecerunan garis lurus     # MOE language
    text_en: Determine the gradient of a straight line  # English
    bloom: apply
```

### Teaching Notes and Assessments

Teaching notes (`.teaching.md`) and assessments (`.assessments.yaml`) default to the MOE language. English versions are stored in the `translations/en/` directory using the same filename.

```
topics/
├── MT3-09.yaml
├── MT3-09.teaching.md              # Malay (primary)
├── MT3-09.assessments.yaml         # Malay (primary)
└── translations/
    └── en/
        ├── MT3-09.teaching.md      # English translation
        └── MT3-09.assessments.yaml # English translation
```

---

## Learning Objective ID

**Format:** `LO{N}` — sequential integer, no padding

```
LO1
LO2
LO3
```

Scoped within a single topic. Reset to `LO1` for each topic.

---

## Assessment Question ID

**Format:** `Q{N}` — sequential integer, no padding

```
Q1
Q2
Q10
```

Scoped within a single assessments file.

---

## Worked Example ID

**Format:** `WE-{NN}` — zero-padded 2-digit sequence

```
WE-01
WE-02
WE-12
```

Scoped within a single examples file.

---

## Folder and File Naming

### Rule: folder name = entity ID

Every folder and file in `curricula/` is named after the `id` field of the YAML it contains. This makes any path self-describing and allows tools to resolve entities from paths without opening files.

| Level | Folder/file name | Equals |
|-------|-----------------|--------|
| Country | `malaysia/` | `country_id` |
| Syllabus | `malaysia-kssm/` | `syllabus_id` |
| Subject | `malaysia-kssm-matematik-tingkatan-3/` | `subject_id` |
| Topic file | `MT3-09.yaml` | `topic_id` + `.yaml` |
| Teaching notes | `MT3-09.teaching.md` | `topic_id` + `.teaching.md` |
| Assessments | `MT3-09.assessments.yaml` | `topic_id` + `.assessments.yaml` |
| Examples | `MT3-09.examples.yaml` | `topic_id` + `.examples.yaml` |

### Fixed folder names (never change)

| Folder | Purpose |
|--------|---------|
| `curricula/` | Root of all curriculum content |
| `topics/` | Contains all topic files within a subject |
| `translations/` | Contains translated versions of content files |
| `translations/{lang}/` | One folder per BCP 47 language code (`en`, `ms`, `ta`, `zh-hans`) |

### Folder naming rules

1. **Always use the full entity ID** — never abbreviate. Subject folder is `malaysia-kssm-matematik-tingkatan-3`, not `matematik` or `math-t3`.
2. **Lowercase kebab-case** — no uppercase letters, no underscores, no spaces.
3. **No version suffixes in folder names** — versioning is handled inside `syllabus.yaml` via a `version` field.
4. **Language of folder names** follows the ID convention — MOE language for grades and subjects, English for countries.

---

## Directory Structure

```
curricula/
└── {country_id}/                              # e.g. malaysia/
    └── {syllabus_id}/                         # e.g. malaysia-kssm/
        ├── syllabus.yaml                      # id: malaysia-kssm
        └── {subject_id}/                      # e.g. malaysia-kssm-matematik-tingkatan-3/
            ├── subject.yaml                   # id: malaysia-kssm-matematik-tingkatan-3
            └── topics/
                ├── {topic_id}.yaml            # e.g. MT3-09.yaml
                ├── {topic_id}.teaching.md
                ├── {topic_id}.assessments.yaml
                ├── {topic_id}.examples.yaml
                └── translations/
                    └── {lang}/                # e.g. en/
                        ├── {topic_id}.teaching.md
                        └── {topic_id}.assessments.yaml
```

**Full example — Malaysia KSSM Matematik Tingkatan 3:**

```
curricula/
└── malaysia/
    └── malaysia-kssm/
        ├── syllabus.yaml
        ├── malaysia-kssm-matematik-tingkatan-1/
        │   ├── subject.yaml
        │   └── topics/
        │       ├── MT1-01.yaml
        │       └── MT1-01.teaching.md
        └── malaysia-kssm-matematik-tingkatan-3/
            ├── subject.yaml
            └── topics/
                ├── MT3-01.yaml
                ├── MT3-01.teaching.md
                ├── MT3-01.assessments.yaml
                ├── MT3-09.yaml
                ├── MT3-09.teaching.md
                ├── MT3-09.assessments.yaml
                └── translations/
                    └── en/
                        ├── MT3-09.teaching.md
                        └── MT3-09.assessments.yaml
```

---

## Validation

The OSS validator enforces these patterns via JSON Schema:

```yaml
topic_id:    pattern: "^[A-Z]{2,3}[0-9]*-[0-9]{2}$"
syllabus_id: pattern: "^[a-z][a-z0-9-]+$"
subject_id:  pattern: "^[a-z][a-z0-9-]+$"
country_id:  pattern: "^[a-z][a-z0-9-]+$"
grade_id:    pattern: "^[a-z][a-z0-9-]+$"
```

---

## AI Generation Note

When the OSS Bot generates IDs or scaffolds new content, it follows these conventions:
- IDs use the official MOE language of the target country
- Every generated YAML includes both `name` (MOE language) and `name_en` (English)
- Every generated learning objective includes both `text` and `text_en`
- Topic prefixes always use the English-derived prefix table (language-neutral)
- If the source document contains a formal chapter/section code, capture it verbatim in `official_ref`; omit the field if no such code exists

If you are writing prompt templates or reviewing AI output, refer to this document. Misformatted IDs are caught by the validator and the PR will be blocked.
