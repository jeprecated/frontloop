# Frontloop v2 epic layout specification

Status: proposed implementation contract  
Date: 2026-06-21

This document specifies the v2 frontloop filesystem layout. It is the contract for moving from a single flat queue to epic-scoped queues while preserving the core frontloop rule that directory movement represents task status.

## Goals

- Keep related ordered tasks for the same epic together.
- Prevent tasks from different epics from becoming visually or structurally mixed after workspace merges.
- Preserve a simple status model: within an epic, the status directory is the task status.
- Keep completed epic history available without clogging the active queue.

## Active layout

A v2 `.frontloop/` tree stores each active epic as a top-level directory:

```text
.frontloop/
├── default/
│   ├── epic.md
│   ├── clarify/
│   ├── ready/
│   ├── in_progress/
│   └── done/
├── checkout-redesign/
│   ├── epic.md
│   ├── clarify/
│   ├── ready/
│   ├── in_progress/
│   └── done/
└── _archive/
```

The active task path format is:

```text
.frontloop/<epic>/<status>/<task>.md
```

Where:

- `<epic>` is an active epic slug such as `default` or `checkout-redesign`.
- `<status>` is one of `clarify`, `ready`, `in_progress`, or `done`.
- `<task>.md` is the task markdown file.

## Default epic

`default/` is the built-in epic bucket for tasks without an explicit epic.

Commands that create tasks without an epic argument must write to:

```text
.frontloop/default/clarify/<task>.md
```

`default` is a normal active epic for listing, task movement, and statistics. It exists so unscoped tasks do not need a separate legacy layout.

## Reserved archive directory

`_archive/` is reserved for completed epics.

Normal active-queue commands must ignore `_archive/` by default, including task listing, status/statistics, task creation, and work selection. Archived epics are retained for historical inspection, not active scheduling.

Top-level names beginning with `_` are reserved and must not be treated as active epic slugs.

## Active epic structure

Every active epic directory must contain:

```text
.frontloop/<epic>/
├── epic.md
├── clarify/
├── ready/
├── in_progress/
└── done/
```

### `epic.md`

`epic.md` stores human-readable epic metadata. It is not the source of truth for which tasks belong to the epic; the task path is.

Example:

```markdown
---
title: Checkout redesign
status: active
created_at: 2026-06-21
completed_at:
---

## Goal

Describe the outcome this epic is meant to deliver.
```

At minimum, implementations should be able to create and preserve `epic.md`. Additional metadata can be added later as long as path-based task membership remains authoritative.

## Task metadata and epic membership

Task files keep their existing task-focused frontmatter:

```markdown
---
title: Render checkout review page
priority: high
---

## Goal

Render the review page for checkout.
```

The task path is the source of truth for epic membership:

```text
.frontloop/checkout-redesign/ready/0020-render-review-page.md
```

This task belongs to the `checkout-redesign` epic because it is under `.frontloop/checkout-redesign/`. A task-level `epic:` frontmatter field is not required and should not be used as the authoritative source of membership.

## Status and movement rules

Within each epic, task status is represented by movement between the four status directories:

```text
.frontloop/<epic>/clarify/<task>.md
.frontloop/<epic>/ready/<task>.md
.frontloop/<epic>/in_progress/<task>.md
.frontloop/<epic>/done/<task>.md
```

Moving a task between statuses must preserve the epic. For example:

```text
.frontloop/checkout-redesign/clarify/render-review-page.md
→ .frontloop/checkout-redesign/ready/0020-render-review-page.md
→ .frontloop/checkout-redesign/in_progress/0020-render-review-page.md
→ .frontloop/checkout-redesign/done/0020-render-review-page.md
```

A status move must not silently move the task to another epic.

## Filename and ordering rules

V2 preserves numeric/order prefixes in `ready/`, `in_progress/`, and `done/` so archived epics remain readable in their original task order.

Recommended conventions:

- `clarify/`: `task-name.md`
- `ready/`: `NNNN-task-name.md`
- `in_progress/`: same filename as `ready/`
- `done/`: preserve the `NNNN-` prefix from `ready/` and `in_progress/`

The exact numeric ranges are a scheduling convention, not a membership mechanism. Epic membership is always path-based.

## Archive layout and lifecycle

When an epic is complete, it can move from the active top-level area to `_archive/`:

```text
.frontloop/_archive/YYYY-MM-DD-<epic>/
├── epic.md
├── clarify/
├── ready/
├── in_progress/
└── done/
```

Archive rules:

- Only active epic directories can be archived.
- Normal active commands ignore archived epics.
- An epic is archivable only when `clarify/`, `ready/`, and `in_progress/` are empty.
- `done/` may contain completed tasks.
- Archiving should update `epic.md` to mark the epic archived and set `completed_at`.
- `default` should not be archived unless a future specification explicitly allows it.

## Migration contract from v1 flat queues

The legacy layout is:

```text
.frontloop/
├── clarify/
├── ready/
├── in_progress/
└── done/
```

Migration to v2 moves all legacy flat tasks into the `default` epic:

```text
.frontloop/clarify/<task>.md        → .frontloop/default/clarify/<task>.md
.frontloop/ready/<task>.md          → .frontloop/default/ready/<task>.md
.frontloop/in_progress/<task>.md    → .frontloop/default/in_progress/<task>.md
.frontloop/done/<task>.md           → .frontloop/default/done/<task>.md
```

Migration requirements:

- Preserve task filenames.
- Preserve task file contents.
- Create `.frontloop/default/epic.md` if it does not exist.
- Create `.frontloop/_archive/` if it does not exist.
- Refuse or clearly report conflicts when both the legacy source and v2 destination contain the same task path.
- Do not delete or hide legacy data unless all moves needed for that status completed successfully.
- Commands that detect a legacy flat layout should guide users to run the migration instead of treating the queue as a valid v2 root.

## Compatibility notes and known inconsistencies

The current repository documentation and Go implementation do not fully agree about ready-queue filename prefixes:

- Existing docs describe ready tasks as `NNNN-task-name.md` with a four-digit ordering prefix.
- Current Go code maps priorities to short prefixes such as `1-`, `2-`, `3-`, and `4-`.

V2 resolves this specification-level conflict by standardizing on numeric/order prefixes in `ready/`, `in_progress/`, and `done/`, with examples using four digits. Follow-up implementation tasks should update code and user-facing documentation to match this spec.

Until the v2 implementation and documentation updates land, older docs may still describe the flat `.frontloop/{clarify,ready,in_progress,done}` layout. Treat this document as the v2 implementation contract.
