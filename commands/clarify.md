---
description: Review tasks in the clarify queue with a human
---

# Frontloop Clarify

Run the human review workflow on tasks in `.frontloop/clarify/`.

## Precondition

Check that `.frontloop/clarify/` exists. If it doesn't, tell the user to run `/frontloop-init`.

If no `.md` files are in `clarify/`, report "No tasks need clarification" and exit.

## Workflow

Read the full workflow from `references/clarify.md` in the frontloop skill directory.

For each `.md` file in `.frontloop/clarify/`:

### 1. Present the task

Read the file. Show the title, goal, and acceptance criteria to the user.

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

- **If yes**: Move to `.frontloop/ready/` with priority prefix. Read the `priority` frontmatter to determine the prefix: `1-` (critical), `2-` (high), `3-` (medium), `4-` (low). The filename becomes `<prefix><original-filename>`.
- **If no**: Leave in `clarify/`. Ask what's missing and add it to the file.

## Output

After processing all tasks, run `/frontloop-status` to show the updated queue.
