# AGENTS.md — OSS Bot

Agent guidance for working in this repository.

## 1. Repository Status (Current Reality)

This repository is currently **documentation-only**. There is no implementation code yet.

Current files:

- `README.md`
- `CLAUDE.md`
- `docs/technical-plan.md`
- `docs/development-timeline.md`
- `docs/business-plan.md`
- `LICENSE`

Do not assume paths like `cmd/`, `internal/`, `web/`, `prompts/`, `scripts/`, `docker-compose.yml`, or `Makefile` exist unless they are created in the same task.

## 2. Project Intent

OSS Bot is planned as an AI tooling layer for [p-n-ai/oss](https://github.com/p-n-ai/oss) with three interfaces:

1. GitHub bot (`@oss-bot`)
2. CLI tool (`oss`)
3. Web contribution portal

All three are planned to share one pipeline: context building -> AI generation -> validation -> PR output.

## 3. Source-of-Truth Documents

When requirements conflict, use this precedence:

1. User request in the active task
2. `docs/technical-plan.md` (architecture and technical direction)
3. `docs/development-timeline.md` (sequencing and milestones)
4. `docs/business-plan.md` (product context and goals)
5. `README.md` (public overview)

`CLAUDE.md` and this file should stay aligned with the docs above.

## 4. Rules for Coding Agents

- Start by verifying actual files on disk before referencing project structure.
- Treat future-state architecture in docs as **planned**, not already implemented.
- If asked to "implement" or "scaffold", create missing directories/files explicitly.
- Keep new code consistent with the plan:
  - Go 1.22+ backend
  - Cobra CLI
  - Next.js + TypeScript web app
  - `OSS_`-prefixed environment configuration
- Keep claims in docs accurate to repository reality. Mark roadmap content as planned.

## 5. Documentation Update Expectations

When editing docs:

- Maintain consistency across `README.md`, `CLAUDE.md`, and files in `docs/`.
- Use explicit status labels where useful:
  - `Planned`
  - `In Progress`
  - `Implemented`
- Avoid command examples that require missing files unless clearly marked as future usage.
- Keep references to external repos accurate:
  - `p-n-ai/oss`
  - `p-n-ai/pai-bot`

## 6. If You Add Initial Code

If a task introduces first implementation files, also update these docs in the same task:

- `README.md`: add "Current Status" and runnable commands that actually work
- `AGENTS.md`: update the "Repository Status" section and existing paths
- `CLAUDE.md`: keep in sync with the implemented structure

## 7. Suggested Validation Checklist

Before finishing any task:

1. Confirm every referenced path exists.
2. Confirm command examples are runnable in current repo state (or marked planned).
3. Confirm no doc claims implementation that is not present.
4. Confirm `AGENTS.md` and `CLAUDE.md` are not contradicting each other.
