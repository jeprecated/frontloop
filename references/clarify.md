# Clarify Workflow

New tasks arrive in `.frontloop/clarify/`. This workflow reviews them with a human and either promotes them to `ready/` or leaves them for further discussion.

## Process

1. List files in `.frontloop/clarify/`.
2. For each task file:
   a. Read the file. Present the title, goal, and each question to the human.
   b. For each question, show the options and the triage agent's recommendation.
   c. Record the human's answer.
3. After all questions are answered, update the file:
   - Write the chosen answers into the **Design Decisions** section as concise statements.
   - **Remove the entire Questions section** and all unchosen options. Keep only the decisions. This minimizes token cost for the worker.
   - Add any implementation guidance the human provides to **Implementation Notes**.
4. If the task is clear enough to execute, move it to `.frontloop/ready/` with a 4-digit numeric prefix. Pick a number based on the task's priority and relation to other tasks. Suggested ranges: critical=0001-2499, high=2500-4999, medium=5000-7499, low=7500-9999. Duplicate numbers are allowed. The filename becomes `NNNN-<original-filename>`.
5. If the human says the task needs more work or isn't ready, leave it in `clarify/`.

## Tasks Without Questions

Some tasks in `clarify/` may not have a Questions section — they may just lack enough detail. In that case:
- Ask the human if the goal and acceptance criteria are clear enough for an agent to execute independently.
- If yes, move to `ready/`.
- If no, ask what's missing and add it to the file.

## Example

Before (in `clarify/`):

```markdown
## Questions

1. **Where should staleness be enforced?**
   - (A) Strategy evaluation boundary
   - (B) Trade execution boundary — recommended, keeps strategies pure
   - (C) Both

2. **Should stale data be a warning or hard error?**
   - (A) Warning with flag
   - (B) Hard error — recommended
```

After (moved to `ready/`):

```markdown
## Design Decisions

- Enforce staleness at the trade execution boundary, not inside strategies.
- Stale data is a hard error — commands fail closed.
```

Questions section is gone. Only the decisions remain.
