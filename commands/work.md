---
description: Pick up and execute the next ready task from the frontloop queue
---

# Frontloop Work

Pick up the highest-priority task from `.frontloop/ready/` and execute it.

## Precondition

Check that `.frontloop/ready/` exists. If it doesn't, tell the user to run `/frontloop-init`.

If no `.md` files are in `ready/`, report "No tasks ready for work" and exit.

If any `.md` files are in `in_progress/`, report "A task is already in progress" and show its title. Ask the user if they want to continue that task or abandon it (move back to `ready/`).

## Workflow

Read the full workflow from `references/worker.md` in the frontloop skill directory.

### 1. Pick the task

List `.md` files in `ready/` sorted alphabetically. The first file is the highest priority (filenames are prefixed with a 4-digit number, e.g. `0100-`, `2500-`; lowest number = highest priority).

Read the file. Present the title, goal, acceptance criteria, and any design decisions to the user.

### 2. Move to in_progress

Move the file from `ready/` to `in_progress/` (keep the same filename including priority prefix).

### 3. Execute

Follow the task's acceptance criteria exactly. Respect design decisions — they were pre-approved by a human. Do not modify Goal, Acceptance Criteria, or Design Decisions sections.

If the task cannot be completed as described:
- Append a **Blocked** section explaining why
- Move the file back to `.frontloop/clarify/` (strip priority prefix)
- Report the blocker and exit

### 4. Complete

Append a **Completion Summary** section to the task file:

```markdown
## Completion Summary

- <what was done, one bullet per change>

### Files Changed

- <path> (new/modified/deleted)
```

Move the file to `.frontloop/done/` and strip the priority prefix from the filename.

Commit changes with version control.

### 5. Report

Run `/frontloop-status` to show the updated queue.
