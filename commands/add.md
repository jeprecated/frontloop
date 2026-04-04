---
description: Create a new task in the frontloop clarify queue
argument-hint: [task description]
---

# Frontloop Add

Create a new task file in `.frontloop/clarify/`.

## Precondition

Check that `.frontloop/` exists. If not, tell the user to run `/frontloop-init`.

## Arguments

`{{arg}}`

If arguments are provided, use them as the initial task description to seed the conversation below.

## Execution

Gather the following from the user (skip any that are already clear from the arguments):

1. **Title** - short human-readable name for the task
2. **Goal** - what the task achieves (1-3 sentences)
3. **Priority** - critical, high, medium, or low
4. **Acceptance Criteria** - concrete checklist items for completion
5. **Design Decisions** (optional) - any pre-answered choices
6. **Implementation Notes** (optional) - hints, relevant files, constraints

## File Creation

Derive the filename from the title using kebab-case (e.g., "Quote age tracking" becomes `quote-age-tracking.md`).

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

## Design Decisions

- <decision 1>

## Implementation Notes

<notes>
```

Omit the Design Decisions and Implementation Notes sections entirely if the user didn't provide them.

## Output

Confirm the file was created and show its path. Run `/frontloop-status` to show the updated queue.
