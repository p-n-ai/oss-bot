# Content Import Prompt

You are an expert at extracting curriculum structure from educational sources.

## Context

**Syllabus:** {{syllabus_id}}
**Source content (pre-extracted text):**
{{document_text}}

**Source type:** {{source_type}}
<!-- "url", "pdf", "docx", "pptx", "txt", "image_ocr", "image_vision", "text" -->
**Image extraction method (if image):** {{image_method}}
<!-- "ocr", "vision", or empty -->
**Source URL (if applicable):** {{source_url}}
**Source format:** {{source_format}}
**Target board:** {{board}}
**Target level:** {{level}}

## Instructions

Extract the curriculum structure from the source text and output as YAML:

1. Identify subjects/strands
2. Identify individual topics within each subject
3. For each topic, determine:
   - A unique ID (format: XX-NN)
   - Name in source language
   - Learning objectives (with Bloom's levels inferred from verbs)
   - Difficulty (beginner/intermediate/advanced)
   - Prerequisites (which topics should come before)

## Bloom's Taxonomy Reference

Infer Bloom's levels from verbs in the objectives:
- **remember** — list, recall, name, state, define
- **understand** — explain, describe, summarise, classify, compare
- **apply** — solve, use, calculate, demonstrate, construct
- **analyse** — differentiate, examine, break down, contrast, investigate
- **evaluate** — assess, justify, critique, judge, defend
- **create** — design, develop, formulate, produce, plan

## Output Format (YAML)

```yaml
subjects:
  - id: subject-id
    name: "Subject Name"
    topics:
      - id: XX-01
        name: "Topic Name"
        difficulty: beginner
        learning_objectives:
          - id: LO1
            text: "..."
            bloom: understand
        prerequisites: []
```

## Notes

- Use `{{syllabus_id}}` as the `syllabus_id` field on every topic
- Do not hardcode curriculum-specific codes; derive all IDs from the source
- If the source language is not English, preserve the original names and add an `en_name` field
- Prerequisites should reference topic IDs within this import batch only
