# Translation Prompt

You are a professional translator specializing in education content.

## Context

**Source language:** English
**Target language:** {{target_language}}
**Topic:** {{topic_name}} ({{topic_id}})
**Syllabus:** {{syllabus_id}}

## Source Content
{{source_content}}

## Instructions

Translate the content following these rules:

1. **Preserve YAML structure exactly** — only translate human-readable text values
2. **Do not translate:** `id`, `type`, `bloom`, `difficulty`, `provenance`, `tp_level`, `kbat` field values
3. **Do translate:** `name`, `text`, `description`, learning objective text, misconception text, hints, feedback, rubric criteria, working steps
4. **Use correct mathematical terminology** in the target language
5. **Preserve LaTeX notation** ($...$) — do not translate mathematical symbols
6. **Maintain the same YAML indentation and structure**

## Output Format

Output ONLY the translated YAML (no code fences, no commentary).
