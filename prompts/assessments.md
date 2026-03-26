# Assessment Generation Prompt

You are an expert educator creating assessment questions for a topic in the {{syllabus_id}} curriculum.

## Context

**Topic:** {{topic_name}} ({{topic_id}})
**Subject:** {{subject_id}}
**Syllabus:** {{syllabus_id}}
**Difficulty:** {{difficulty}}
**Count:** {{count}} questions
**Target difficulty:** {{target_difficulty}}

### Learning Objectives
{{learning_objectives}}

## Instructions

Generate {{count}} assessment questions as YAML. Each question must include:
- Worked solution (`answer.working`)
- Mark scheme (`rubric`) with partial marks
- Progressive hints (at least 1, more for harder questions)
- Answer type: `exact`, `multiple_choice`, or `free_text` as appropriate

Output format (YAML):

```yaml
topic_id: {{topic_id}}
provenance: ai-generated

questions:
  # Group questions by learning objective or subchapter
  - id: Q1
    text: "Question text. Supports $LaTeX$ notation."
    difficulty: easy          # easy | medium | hard
    learning_objective: LO1   # Must match a learning objective ID from the topic
    tp_level: 2               # Performance/mastery level (use syllabus scale)
    kbat: false               # true if higher-order thinking (analyze/evaluate/create)
    answer:
      type: exact             # exact | multiple_choice | free_text
      value: "correct answer"
      working: |
        Step-by-step solution
    marks: 2
    rubric:
      - marks: 1
        criteria: "First mark criterion"
      - marks: 1
        criteria: "Second mark criterion"
    hints:
      - level: 1
        text: "Gentle nudge"
      - level: 2
        text: "More explicit help — address the specific misconception"

  # For multiple choice questions:
  - id: Q2
    text: "Question stem\n\nA) Option A\nB) Option B\nC) Option C\nD) Option D"
    difficulty: medium
    learning_objective: LO1
    tp_level: 3
    kbat: false
    answer:
      type: multiple_choice
      value: "B"
      working: |
        Step-by-step solution showing why B is correct
    marks: 1
    rubric:
      - marks: 1
        criteria: "Correctly selects option B."
    hints:
      - level: 1
        text: "MISCONCEPTION ALERT: [Describe the common trap]"
      - level: 2
        text: "Try [specific approach]"
```

## Requirements
- Distribute questions across available learning objectives
- Difficulty spread: mix of easy, medium, hard per {{target_difficulty}}
- Follow the exam format and conventions of the {{syllabus_id}} curriculum
- Include LaTeX for mathematical notation ($...$)
- Each question must test a single concept clearly
- Include `tp_level` matching the syllabus performance/mastery scale
- Set `kbat: true` for questions at analyze/evaluate/create Bloom's levels
- Use a mix of answer types: `exact`, `multiple_choice`, and `free_text`
- Hints should address specific misconceptions (prefix with "MISCONCEPTION ALERT:" where relevant)
- Rubric must support partial marks — show what earns each mark
- Group questions by subchapter/learning objective with YAML comments
