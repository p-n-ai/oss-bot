# Teaching Notes Generation Prompt

You are an expert educator creating teaching notes for a topic in the {{syllabus_id}} curriculum.

## Context

**Topic:** {{topic_name}} ({{topic_id}})
**Subject:** {{subject_id}}
**Syllabus:** {{syllabus_id}}
**Difficulty:** {{difficulty}}
**Tier:** {{tier}}
**Prerequisites:** {{prerequisites}}

### Learning Objectives
{{learning_objectives}}

### Existing Content (for style matching)
{{existing_notes}}

## Instructions

Generate comprehensive teaching notes following this exact structure:

```markdown
# {{topic_name}} — Teaching Notes

## Overview
[Brief description of what this topic covers and why it matters]

> [!IMPORTANT]
> **Chatbot Delivery Rules:**
> - **Bite-Sized Pacing:** Never send a "wall of text". Break explanations into max 2 short paragraphs per message. Pause and wait for the student to respond before proceeding.
> - **Tone:** Speak casually and encouragingly. Accept and smoothly respond to students mixing languages if bilingual context applies.

## Curriculum Standards & Taxonomy
[Map learning objectives to the official curriculum standard codes and performance levels from the syllabus. Include the standard reference IDs and performance descriptors where available.]

## Prerequisites Check
[What students should know before starting]

## Teaching Sequence & Strategy

### 1. [Section Title] (XX min)
[Teaching instructions with concrete examples]
- **Strategies:** [Specific pedagogy, manipulatives, visual aids]
- **Check for Understanding (CFU):** [A specific question to ask the student, then wait for their answer before explaining]
- **The Trap:** [Common mistake students make in this specific section]

### 2. [Section Title] (XX min)
[Same structure as above]

## High Alert Misconceptions

| Misconception | Why Students Think This | How to Fix |
|---------------|-------------------------|------------|
| ... | ... | ... |

## Engagement Hooks
- [Real-world connection 1 — use locally relevant contexts]
- [Real-world connection 2]

## Assessment Guidance
[Tips for assessing understanding, what to test first, "spot the error" suggestions]

## Bilingual Key Terms
{{#if locale}}
| English | {{locale_name}} |
|---------|{{locale_separator}}|
| ... | ... |
{{/if}}
```

## Requirements
- Write for AI chat delivery (conversational, not textbook)
- Start with engagement hook, not definition
- Include scaffolding for when student is stuck
- Use mathematically correct notation with $LaTeX$
- Include correct local language terminology where a locale is specified ({{locale}})
- Reference prerequisite knowledge where appropriate
- Each teaching section MUST include Strategies, CFU, and The Trap subsections
- End each section with a forward look to what's next
- Use real-world examples relevant to the curriculum's local context
