---
description: Gather feature ideas from the user and add them to the clarify queue
argument-hint: [optional topic or area to brainstorm about]
---

# Frontloop Gather

Interactively collect feature ideas from the user and create task files in `.frontloop/clarify/` for each one.

## Precondition

Check that `.frontloop/` exists. If not, tell the user to run `/frontloop-init`.

## Arguments

`{{arg}}`

If arguments are provided, use them to frame the gathering session (e.g., a specific area of the product to brainstorm about).

## Execution

### 1. Start the session

Tell the user you're ready to collect feature ideas. If arguments were provided, acknowledge the focus area.

Explain:
- They can describe ideas in any level of detail — a single sentence is fine
- You'll create a task for each one with your own questions and unknowns filled in
- Say "done" or "that's all" when finished

### 2. Collect ideas

Wait for the user to describe a feature idea. For each idea:

- Acknowledge it briefly (one line)
- Ask if they have another idea

Keep collecting until the user signals they're done.

### 3. Create tasks

For each idea collected, create a task file in `.frontloop/clarify/`. Derive the filename from your chosen title using kebab-case.

For each task, **you** fill in:

- **Title** — a short, clear name you derive from the idea
- **Goal** — what the feature achieves, based on what the user said (1-3 sentences)
- **Priority** — your best guess: critical, high, medium, or low
- **Acceptance Criteria** — concrete checklist items you infer from the idea
- **Questions** — things you'd need answered before implementation. Format as lettered options with a recommendation, so they're ready for `/frontloop-clarify`:

```markdown
### Q1: <question>
- a) <option>
- b) <option>
- c) <option>
- **Recommendation**: <your pick and why>
```

Write to `.frontloop/clarify/<filename>.md`:

```markdown
---
title: <title>
priority: <priority>
---

## Goal

<goal>

## Acceptance Criteria

- <criterion 1>
- <criterion 2>
- ...

## Questions

### Q1: <question>
- a) <option>
- b) <option>
- **Recommendation**: <pick>

### Q2: ...
```

Do **not** ask the user these questions during gather. The questions are for the `/frontloop-clarify` step.

## Output

After creating all tasks, show a summary of what was created (title + priority for each). Run `/frontloop-status` to show the updated queue.
