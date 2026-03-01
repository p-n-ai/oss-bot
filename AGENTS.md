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
3. `docs/development-timeline.md` + `docs/implementation-guide.md` (sequencing, milestones, and executable instructions — **always use both together**)
4. `docs/business-plan.md` (product context and goals)
5. `README.md` (public overview)

`CLAUDE.md` and this file should stay aligned with the docs above.

## 4. Daily Implementation — Required Reading

**MANDATORY:** Before starting any day's implementation work, you MUST read and cross-reference BOTH of these documents:

1. **`docs/development-timeline.md`** — the daily task breakdown with task IDs, ownership, and sequencing
2. **`docs/implementation-guide.md`** — step-by-step executable instructions with code templates, tests, file paths, entry/exit criteria, and validation commands

These two documents are complementary and both are required:
- The **timeline** tells you WHAT to build each day and in what order (task IDs like `B-W4D16-1`)
- The **implementation guide** tells you HOW to build it (exact file paths, code, tests, validation steps)

**Do not implement from the timeline alone** — it lacks the detail needed for correct implementation. **Do not implement from the guide alone** — you may miss sequencing dependencies and ownership context from the timeline.

For each day:
1. Read the day's section in both documents
2. Check **entry criteria** in the implementation guide before starting
3. Follow the TDD workflow (see below) for each task
4. Run **validation commands** from the implementation guide
5. Verify all **exit criteria** checkboxes before moving to the next day

## 5. Rules for Coding Agents

- Start by verifying actual files on disk before referencing project structure.
- Treat future-state architecture in docs as **planned**, not already implemented.
- If asked to "implement" or "scaffold", create missing directories/files explicitly.
- Keep new code consistent with the plan:
  - Go 1.22+ backend
  - Cobra CLI
  - Next.js + TypeScript web app
  - `OSS_`-prefixed environment configuration
- Keep claims in docs accurate to repository reality. Mark roadmap content as planned.

### Test-Driven Development (TDD) — Mandatory

Every feature must follow this strict cycle:

1. **Write tests first** — define expected behavior with unit tests before writing implementation code.
2. **Implement** — write the minimum code to make the tests pass.
3. **Run package tests** — verify the new feature works (`go test ./internal/<package>/...`).
4. **Run full test suite** — run `go test ./...` to confirm no existing tests are broken.
5. **Never skip step 4** — a feature is not complete until the full suite passes.

Testing conventions:
- Use stdlib `testing` with table-driven tests and `t.Run()` subtests.
- Mock AI providers for deterministic output — never call real AI APIs in tests.
- Test files live alongside source: `validator.go` → `validator_test.go`.

## 6. Documentation Update Expectations

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

## 7. If You Add Initial Code

If a task introduces first implementation files, also update these docs in the same task:

- `README.md`: add "Current Status" and runnable commands that actually work
- `AGENTS.md`: update the "Repository Status" section and existing paths
- `CLAUDE.md`: keep in sync with the implemented structure

## 8. Suggested Validation Checklist

Before finishing any task:

1. Confirm every referenced path exists.
2. Confirm command examples are runnable in current repo state (or marked planned).
3. Confirm no doc claims implementation that is not present.
4. Confirm `AGENTS.md` and `CLAUDE.md` are not contradicting each other.
