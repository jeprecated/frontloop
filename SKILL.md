---
name: frontloop
description: File-based task queue for agent loops. Defines the v2 epic-first .frontloop/ directory structure, task markdown format, and lifecycle. Tasks are markdown files with YAML frontmatter that move between per-epic status directories (clarify when review is needed, ready, in_progress, done). Use when the user mentions frontloop, task queues, or agent task pipelines. See references/clarify.md for the human-review workflow and references/worker.md for the execution workflow.
---

# Frontloop

Markdown task files move between status directories inside an epic. The directory is the status; the top-level epic directory is the bucket that keeps related ordered tasks together.

**Task creation rule:** Put new tasks directly in `.frontloop/<epic>/ready/` when the goal, acceptance criteria, and relevant decisions are clear enough for an agent to execute without more human input. Use `.frontloop/<epic>/clarify/` only when there are open questions, missing decisions, or an explicit need for human review. Never create new tasks directly in `in_progress/` or `done/`. If no epic is specified, use the built-in `default` epic.

```text
.frontloop/
├── default/                 # built-in bucket for unscoped tasks
│   ├── epic.md
│   ├── clarify/             # default tasks needing human review
│   ├── ready/               # actionable default tasks ready to work
│   ├── in_progress/         # task currently being worked on
│   └── done/                # completed tasks with summaries
├── <epic>/                  # e.g. checkout-redesign
│   ├── epic.md
│   ├── clarify/             # tasks needing human review
│   ├── ready/               # actionable tasks ready to work
│   ├── in_progress/
│   └── done/
└── _archive/                # completed epics; ignored by active workflows
```

Active task path format:

```text
.frontloop/<epic>/<status>/<task>.md
```

`_archive/` is reserved. Do not treat archived epics as active work, do not create new tasks there, and ignore `_archive/` during status, clarify, work, and gather flows.

## Filename Conventions

Within each epic:

- **clarify/**: `task-name.md`
- **ready/**: `NNNN-task-name.md` where NNNN is a zero-padded 4-digit number (0001-9999). Alphabetical sort gives correct execution order. Suggested ranges: critical=0001-2499, high=2500-4999, medium=5000-7499, low=7500-9999. Agents can pick any number — ranges are a guide, not enforced. Duplicate numbers are allowed because filenames remain unique.
- **in_progress/**: Same filename as in ready/.
- **done/**: Same filename as in ready/in_progress; preserve the numeric prefix so completed and archived epics remain readable in execution order.

## Creating Tasks

When creating frontloop tasks — whether via `/add`, `/gather`, a Pi tool, or manually — choose the initial status intentionally:

- **Create in `.frontloop/<epic>/ready/`** when the task is actionable now: the goal and acceptance criteria are clear, there are no open Questions, and any important choices are already captured as Design Decisions or Implementation Notes. This is the right choice when the user says they want to work on the task immediately. Use a 4-digit ordering prefix in the filename, e.g. `2500-task-name.md`.
- **Create in `.frontloop/<epic>/clarify/`** when the task needs human answers, prioritisation, or more detail before a worker can execute it. Include a **Questions** section only for clarify tasks.

If the user does not specify an epic, use `.frontloop/default/ready/` for ready work or `.frontloop/default/clarify/` for tasks needing review.

Epic membership comes from the path. Do not add or rely on an `epic:` field in task frontmatter.

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
- **Blocked** (temporary) — Why the task could not be completed; blocked tasks return to that epic's `clarify/` queue.
- **Completion Summary** (done/ only) — What was done, what files changed.

## Epic Lifecycle

- `default/` is the built-in active epic for unscoped tasks.
- Create named epics explicitly before writing tasks to them.
- Archive completed epics under `.frontloop/_archive/YYYY-MM-DD-<epic>/` only after their `clarify/`, `ready/`, and `in_progress/` queues are empty.
- Active workflows ignore `_archive/` by default.

## Migrating Legacy Queues

The old flat layout was `.frontloop/{clarify,ready,in_progress,done}/`. Migrate it into the v2 `default` epic with:

```bash
fl migrate epic-layout
```

After migration, legacy tasks live under `.frontloop/default/<status>/`.

## Workflows

- **Human review**: See [references/clarify.md](references/clarify.md)
- **Task execution**: See [references/worker.md](references/worker.md)

## Commands

| Command | Purpose |
|---------|---------|
| `/init` | Create the v2 `.frontloop/` tree with `default/` and `_archive/` |
| `/status` | Show active queue state grouped by epic |
| `/clarify` | Review tasks in one or more epic `clarify/` queues with a human |
| `/work` | Pick up and execute the next ready task, optionally within a specific epic |
| `/add` | Create a new task in `ready/` or `clarify/` depending on whether more review is needed |
| `/gather` | Collect feature ideas from user, batch-create ready tasks or clarify tasks with questions |
