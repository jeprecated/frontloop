# Clarify Workflow

Tasks that need human review arrive in `.frontloop/<epic>/clarify/`. This workflow reviews them with a human and either promotes them to that same epic's `ready/` queue or leaves them for further discussion. Tasks that are already actionable may be created directly in `ready/` and skip this workflow.

`default` is the built-in epic for unscoped tasks. Ignore `.frontloop/_archive/`; archived epics are historical and are not active clarify queues.

## Process

1. Select the active epic or epics to clarify. If no epic is specified, scan all active epics and group tasks by epic.
2. List files in `.frontloop/<epic>/clarify/`, sorted by filename within each epic.
3. For each task file:
   a. Read the file. Present the epic, title, goal, acceptance criteria, and each question to the human.
   b. For each question, show the options and the triage agent's recommendation.
   c. Record the human's answer.
4. After all questions are answered, update the file:
   - Write the chosen answers into the **Design Decisions** section as concise statements.
   - **Remove the entire Questions section** and all unchosen options. Keep only the decisions. This minimizes token cost for the worker.
   - Add any implementation guidance the human provides to **Implementation Notes**.
5. If the task is clear enough to execute, move it within the same epic to `.frontloop/<epic>/ready/` with a 4-digit numeric prefix. Pick a number based on the task's priority and relation to other tasks in that epic. Suggested ranges: critical=0001-2499, high=2500-4999, medium=5000-7499, low=7500-9999. Duplicate numbers are allowed. The filename becomes `NNNN-<original-filename>`.
6. If the human says the task needs more work or isn't ready, leave it in `.frontloop/<epic>/clarify/`.

Do not move a task between epics during clarification unless the human explicitly asks for that change.

## Tasks Without Questions

Some tasks in `clarify/` may not have a Questions section — they may just lack enough detail. In that case:

- Ask the human if the goal and acceptance criteria are clear enough for an agent to execute independently.
- If yes, move to the same epic's `ready/` queue.
- If no, ask what's missing and add it to the file while leaving it in the same epic's `clarify/` queue.

## Example

Before (in `.frontloop/checkout-redesign/clarify/render-review-page.md`):

```markdown
## Questions

1. **Where should review data be loaded?**
   - (A) Inside the page component
   - (B) In the route loader — recommended, keeps rendering simple
   - (C) Both

2. **Should missing optional totals be a warning or hard error?**
   - (A) Warning with placeholder
   - (B) Hard error — recommended
```

After (moved to `.frontloop/checkout-redesign/ready/2500-render-review-page.md`):

```markdown
## Design Decisions

- Load review data in the route loader, not inside the page component.
- Missing optional totals are a hard error.
```

Questions section is gone. Only the decisions remain. The task stays in the `checkout-redesign` epic.
