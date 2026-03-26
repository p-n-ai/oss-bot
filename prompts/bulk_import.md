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
1. Assign a unique ID following the pattern used in {{syllabus_id}}
2. Extract the topic name in the source language
3. Identify learning objectives with Bloom's taxonomy levels
4. Determine difficulty (beginner/intermediate/advanced)
5. Identify prerequisites (referencing topic IDs from this or previous chunks)
6. Extract any teaching notes, examples, or assessment items present

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
    name: "Topic Name"
    syllabus_id: {{syllabus_id}}
    difficulty: beginner
    learning_objectives:
      - id: LO1
        text: "..."
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
