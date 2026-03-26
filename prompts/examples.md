# Worked Examples Generation Prompt

You are an expert educator creating worked examples for a topic in the {{syllabus_id}} curriculum.

## Context

**Topic:** {{topic_name}} ({{topic_id}})
**Subject:** {{subject_id}}
**Syllabus:** {{syllabus_id}}
**Difficulty:** {{difficulty}}

### Learning Objectives
{{learning_objectives}}

## Instructions

Generate 3 worked examples as YAML. Each example must include:
- A `real_world_analogy` to make the concept relatable
- A `misconception_alert` highlighting the common mistake students make
- A `scenario` presenting the problem in context
- Step-by-step `working` with numbered steps and explanations

Output format (YAML):

```yaml
topic_id: {{topic_id}}
provenance: ai-generated
description: "Worked examples for {{topic_name}}"
worked_examples:
  - id: WE-01
    topic: "Section or subtopic name"
    difficulty: easy
    real_world_analogy: "A relatable analogy grounded in local context"
    misconception_alert: "Common mistake students make and why"
    scenario: "The problem statement in real-world context"
    working: |
      Step 1: [First step with explanation]
      Step 2: [Next step]
      Step 3: [Final step with answer]
```

## Requirements
- Cover a progression from easy to medium to hard
- Each example should target different learning objectives where possible
- Use real-world scenarios relevant to the curriculum's local context
- Working must be broken into clearly numbered steps with explanations
- The misconception_alert should describe a REAL student error pattern
- Use $LaTeX$ notation for mathematical expressions
