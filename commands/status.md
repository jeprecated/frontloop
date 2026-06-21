---
description: Show the current frontloop task queue state grouped by epic
argument-hint: [optional epic]
---

# Frontloop Status

Display the state of active `.frontloop/` task queues grouped by epic.

## Precondition

Check that `.frontloop/default/` exists. If no `.frontloop/` exists, tell the user to run `/frontloop-init`.

If the repository still has the legacy flat `.frontloop/{clarify,ready,in_progress,done}/` layout, tell the user to run:

```bash
fl migrate epic-layout
```

Do not include `.frontloop/_archive/` in active status output.

## Arguments

`{{arg}}`

If an epic slug is provided, show only that active epic. Otherwise, show all active epics.

## Execution

1. Discover active epics by listing top-level `.frontloop/` directories that contain `clarify/`, `ready/`, `in_progress/`, and `done/`.
2. Ignore reserved directories, especially `.frontloop/_archive/` and any top-level name beginning with `_`.
3. For each selected epic, read `.md` files in its four status directories.
4. Parse YAML frontmatter to extract `title` and `priority`.
5. Display each epic separately.

Display the results in this format:

```text
=== Frontloop Status ===

EPIC: <epic> [(default)]

IN PROGRESS (N):
  <filename>  [<priority>]  <title>

READY (N):
  <filename>  [<priority>]  <title>

NEEDS CLARIFICATION (N):
  <filename>  [<priority>]  <title>

DONE (N):
  <filename>  [<priority>]  <title>
```

### Rules

- **Epic grouping**: Never mix tasks from different epics into one flat visual list.
- **Default epic**: Label `default` clearly, for example `EPIC: default (default)`.
- **In Progress**: Show all active in-progress tasks for the epic.
- **Ready**: Sort files alphabetically within the epic. The 4-digit prefix gives the intended order.
- **Needs Clarification**: Show all tasks in that epic's `clarify/` queue.
- **Done**: Show the 5 most recently modified files per epic. If more than 5, append `... and N more` for that epic.
- Empty sections show `(empty)` instead of file listings.
- Filenames are shown without the `.md` extension.
