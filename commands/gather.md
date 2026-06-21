---
description: Gather feature ideas from the user and add them to an epic's clarify queue
argument-hint: [optional epic or brainstorming topic]
---

# Frontloop Gather

Interactively collect feature ideas from the user and create task files in `.frontloop/<epic>/clarify/` for each one. If no epic is specified, use `default`.

## Precondition

Check that `.frontloop/default/` exists. If not, tell the user to run `/init`.

If the repository still uses the legacy flat layout, tell the user to run `fl migrate epic-layout` before gathering tasks.

Ignore `.frontloop/_archive/`; archived epics are not active task destinations.

## Arguments

`{{arg}}`

If arguments name an active epic, use it as the target epic. Otherwise use arguments to frame the gathering session, and default task creation to the `default` epic unless the user chooses another active epic.

Do not silently create a new epic from a typo. Ask before using a non-existent epic, and recommend `fl epic new <slug>` if a new epic is needed.

## Execution

### 1. Start the session

Tell the user you're ready to collect feature ideas. If arguments were provided, acknowledge the focus area and target epic.

Explain:

- They can describe ideas in any level of detail — a single sentence is fine
- You'll create a task for each one with your own questions and unknowns filled in
- New tasks will be written to `.frontloop/<epic>/clarify/`
- Say "done" or "that's all" when finished

### 2. Collect ideas

Wait for the user to describe a feature idea. For each idea:

- Acknowledge it briefly (one line)
- Note the target epic if it differs from the default for the session
- Ask if they have another idea

Keep collecting until the user signals they're done.

### 3. Create tasks

For each idea collected, create a task file in `.frontloop/<epic>/clarify/`. Derive the filename from your chosen title using kebab-case. Ensure uniqueness within that epic's `clarify/` directory.

For each task, **you** fill in:

- **Title** — a short, clear name you derive from the idea
- **Goal** — what the feature achieves, based on what the user said (1-3 sentences)
- **Priority** — your best guess: critical, high, medium, or low
- **Acceptance Criteria** — concrete checklist items you infer from the idea
- **Questions** — things you'd need answered before implementation. Format as lettered options with a recommendation, so they're ready for `/clarify`:

```markdown
### Q1: <question>
- a) <option>
- b) <option>
- c) <option>
- **Recommendation**: <your pick and why>
```

Write to `.frontloop/<epic>/clarify/<filename>.md`:

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

Do **not** ask the user these questions during gather. The questions are for the `/clarify` step.

Epic membership is represented by the path; do not add an `epic:` field to task frontmatter.

## Output

After creating all tasks, show a summary of what was created, including epic, title, and priority for each task. Run `/status` to show the updated queue grouped by epic.
