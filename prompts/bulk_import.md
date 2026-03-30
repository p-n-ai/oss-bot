# Bulk Import Prompt

You are extracting curriculum structure from a large educational document.

## Context

**Syllabus:** {{syllabus_id}}
**Source content (chunk {{chunk_index}} of {{total_chunks}}):**
{{document_chunk}}

**Source type:** {{source_type}}
**Previously extracted topics (for continuity):**
{{previous_topics}}

## Instructions

Extract all curriculum topics from this chunk. For each topic:
1. Assign a unique ID: `{PREFIX}{grade_num}-{NN}` (e.g. `MT3-01`). Prefix is always from
   the English subject name — language-neutral: MT=Mathematics, SC=Science, PHY=Physics,
   CHM=Chemistry, BIO=Biology, HIS=History, GEO=Geography, BM=Malay/Indonesian, ENG=English.
2. Set `official_ref` to the chapter/section/topic code printed in the source document (e.g. `"Bab 9"`, `"C2.5"`, `"Chapter 12"`). Omit if no formal code is present.
3. Set `name` in the MOE's official language and `name_en` in English.
4. Extract the topic name in the source language
5. Identify learning objectives with Bloom's taxonomy levels; include `text_en` (English) alongside each `text`.
6. Determine difficulty (beginner/intermediate/advanced)
7. Identify prerequisites (referencing topic IDs from this or previous chunks)
8. Extract any teaching notes, examples, or assessment items present

Maintain consistency with topics extracted from previous chunks.

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
topics:
  - id: XX-NN
    official_ref: "Chapter N"   # board's code from source doc; omit if absent
    name: "Topic Name"          # MOE official language
    name_en: "Topic Name"       # English
    syllabus_id: {{syllabus_id}}
    difficulty: beginner
    learning_objectives:
      - id: 1.0.1
        text: "..."
        text_en: "... (English)"
        bloom: understand
    prerequisites: []
    teaching_notes: "..." # if present in source
    examples: []          # if present in source
    assessments: []       # if present in source
```

## Notes

- Do not repeat topics already listed in `{{previous_topics}}`
- If a chapter boundary is detected, note it as a comment in the YAML
- If the source language is not English, preserve original names and add an `en_name` field
- Prerequisites must only reference IDs already seen in this or previous chunks
- Use `{{syllabus_id}}` as the `syllabus_id` on every topic — do not hardcode curriculum names
