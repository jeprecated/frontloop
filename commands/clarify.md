---
description: Review tasks in epic clarify queues with a human
argument-hint: [optional epic]
---

# Frontloop Clarify

Run the human review workflow on tasks in `.frontloop/<epic>/clarify/`.

## Precondition

Check that `.frontloop/default/` exists. If it doesn't, tell the user to run `/frontloop-init`.

If the repository still uses the legacy flat layout, tell the user to run `fl migrate epic-layout` before clarifying tasks.

Ignore `.frontloop/_archive/`; archived epics are historical and not active clarify queues.

## Arguments

`{{arg}}`

If an epic slug is provided, process only `.frontloop/<epic>/clarify/`. Otherwise process all active epic clarify queues, grouped by epic.

If no `.md` files are in the selected active `clarify/` queues, report "No tasks need clarification" and exit.

## Workflow

Read the full workflow from `references/clarify.md` in the frontloop skill directory.

For each `.md` file in `.frontloop/<epic>/clarify/`:

### 1. Present the task

Read the file. Show the epic, title, goal, and acceptance criteria to the user.

### 2. Handle questions

If the task has a **Questions** section:

- Present each question one at a time
- Show the lettered options and the recommendation
- Record the user's answer

If the task has no Questions section:

- Ask the user if the goal and acceptance criteria are clear enough for an agent to execute independently

### 3. Update the file

- Write chosen answers into the **Design Decisions** section as concise statements
- **Remove the entire Questions section** — keep only the decisions
- Add any implementation guidance the user provides to **Implementation Notes**

### 4. Promote or keep

Ask the user if this task is ready to work on.

- **If yes**: Move the task within the same epic to `.frontloop/<epic>/ready/` with a 4-digit numeric prefix. Pick a number based on priority and relation to other tasks in that epic. Suggested ranges: critical=0001-2499, high=2500-4999, medium=5000-7499, low=7500-9999. Duplicate numbers are allowed. The filename becomes `NNNN-<original-filename>`.
- **If no**: Leave it in `.frontloop/<epic>/clarify/`. Ask what's missing and add it to the file.

Never move a task to another epic during clarification unless the user explicitly asks for an epic change.

## Output

After processing all tasks, run `/frontloop-status` to show the updated queue grouped by epic.
