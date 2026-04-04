---
name: frontloop
description: File-based task queue for agent loops. Defines the .frontloop/ directory structure, task markdown format, and lifecycle. Tasks are markdown files with YAML frontmatter that move between directories (clarify → ready → in_progress → done). Use when the user mentions frontloop, task queues, or agent task pipelines. See references/clarify.md for the human-review workflow and references/worker.md for the execution workflow.
---

# Frontloop

Markdown task files move between directories. The directory is the status. The filename is the id.

**IMPORTANT: New tasks ALWAYS go in `clarify/`. Never create tasks directly in `ready/`, `in_progress/`, or `done/`.** Only the `/clarify` command promotes tasks from `clarify/` to `ready/` after human review.

```
.frontloop/
├── clarify/      # ALL new tasks start here. NEVER skip this step.
├── ready/        # Reviewed and ready. Only /clarify moves tasks here.
├── in_progress/  # Task currently being worked on.
└── done/         # Completed tasks with summaries.
```

### Filename Conventions

- **clarify/**: `task-name.md`
- **ready/**: `{priority}-task-name.md` where priority is `1` (critical), `2` (high), `3` (medium), `4` (low). Alphabetical sort gives highest priority first.
- **in_progress/**: Same filename as in ready/.
- **done/**: `task-name.md` (priority prefix removed).

## Creating Tasks

When creating frontloop tasks — whether via `/add`, `/gather`, or manually — **always write them to `.frontloop/clarify/`**. Tasks must be reviewed by a human (`/clarify`) before they can move to `ready/`. There are no exceptions.

## Task File Format

```markdown
---
title: Quote-age tracking and stale-data guards
priority: critical
---

## Goal

Reject trades when inputs are stale.

## Acceptance Criteria

- Quotes have explicit freshness metadata
- Profiles define max staleness per input type
- Commands fail with machine-readable stale-data error

## Design Decisions

- Enforce at execution boundary, not inside strategies

## Implementation Notes

Optional freeform guidance for the worker.
```

### Frontmatter

| Field | Required | Values |
|-------|----------|--------|
| `title` | yes | Human-readable name |
| `priority` | yes | `critical`, `high`, `medium`, `low` |

### Body Sections

- **Goal** (required) — What the task achieves, 1-3 sentences.
- **Acceptance Criteria** (required) — Concrete checklist for completion.
- **Design Decisions** (optional) — Pre-answered choices so the worker doesn't need to ask.
- **Implementation Notes** (optional) — Freeform hints, relevant files, constraints.
- **Questions** (clarify/ only) — Specific questions with lettered options and a recommendation. Removed once answered.
- **Completion Summary** (done/ only) — What was done, what files changed.

## Workflows

- **Human review**: See [references/clarify.md](references/clarify.md)
- **Task execution**: See [references/worker.md](references/worker.md)

## Commands

| Command | Purpose |
|---------|---------|
| `/init` | Create `.frontloop/` directories in the current project |
| `/status` | Show queue state |
| `/clarify` | Review tasks in `clarify/` with a human |
| `/work` | Pick up and execute the next ready task |
| `/add` | Create a new task in `clarify/` |
| `/gather` | Collect feature ideas from user, batch-create tasks in `clarify/` |
