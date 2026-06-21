---
description: Pick up and execute the next ready task from an epic frontloop queue
argument-hint: [optional epic]
---

# Frontloop Work

Pick up a ready task from `.frontloop/<epic>/ready/` and execute it.

## Precondition

Check that `.frontloop/default/` exists. If it doesn't, tell the user to run `/init`.

If the repository still uses the legacy flat layout, tell the user to run `fl migrate epic-layout` before working tasks.

Ignore `.frontloop/_archive/`; archived epics are historical and not active work queues.

## Arguments

`{{arg}}`

If an epic slug is provided, work only within `.frontloop/<epic>/`. If no epic is provided:

- If exactly one active epic has ready tasks, use that epic.
- If multiple active epics have ready tasks, show the grouped ready list and ask the user which epic to work.
- If no active epic has ready tasks, report "No tasks ready for work" and exit.

If any `.md` files are in the selected epic's `in_progress/`, report "A task is already in progress" and show its title. Ask the user if they want to continue that task or abandon it by moving it back to the same epic's `ready/` queue.

## Workflow

Read the full workflow from `references/worker.md` in the frontloop skill directory.

### 1. Pick the task

List `.md` files in `.frontloop/<epic>/ready/` sorted alphabetically. The first file is the highest priority for that epic (filenames are prefixed with a 4-digit number, e.g. `0100-`, `2500-`; lowest number = highest priority).

Read the file. Present the epic, title, goal, acceptance criteria, and any design decisions to the user.

### 2. Move to in_progress

Move the file from `.frontloop/<epic>/ready/` to `.frontloop/<epic>/in_progress/`. Keep the same filename, including the priority/order prefix.

### 3. Execute

Follow the task's acceptance criteria exactly. Respect design decisions — they were pre-approved by a human. Do not modify Goal, Acceptance Criteria, or Design Decisions sections.

If the task cannot be completed as described:

- Append a **Blocked** section explaining why
- Move the file back to `.frontloop/<epic>/clarify/`
- Strip the priority/order prefix when returning to `clarify/`
- Report the blocker and exit

### 4. Complete

Append a **Completion Summary** section to the task file:

```markdown
## Completion Summary

- <what was done, one bullet per change>

### Files Changed

- <path> (new/modified/deleted)
```

Move the file to `.frontloop/<epic>/done/`. Preserve the priority/order prefix in `done/` so completed and archived epics remain readable in execution order.

Commit changes with the project's version-control workflow when appropriate.

### 5. Report

Run `/status` to show the updated queue grouped by epic.
