---
description: Create a new task in an epic's frontloop clarify queue
argument-hint: [optional epic] [task description]
---

# Frontloop Add

Create a new task file in `.frontloop/<epic>/clarify/`. If no epic is specified, use `.frontloop/default/clarify/`.

## Precondition

Check that `.frontloop/default/` exists. If not, tell the user to run `/init`.

If the repository still uses the legacy flat layout, tell the user to run `fl migrate epic-layout` before adding tasks.

Do not create tasks under `.frontloop/_archive/`.

## Arguments

`{{arg}}`

If arguments include a clear epic slug, use that active epic. Otherwise ask for the target epic only if it is ambiguous; default to `default`.

Do not silently create a new epic from a typo. If the requested epic does not exist, ask the user whether to create it first with `fl epic new <slug>` or use `default`.

If a task description is provided, use it to seed the conversation below.

## Execution

Gather the following from the user (skip any that are already clear from the arguments):

1. **Epic** - active epic slug; default is `default`
2. **Title** - short human-readable name for the task
3. **Goal** - what the task achieves (1-3 sentences)
4. **Priority** - critical, high, medium, or low
5. **Acceptance Criteria** - concrete checklist items for completion
6. **Design Decisions** (optional) - any pre-answered choices
7. **Implementation Notes** (optional) - hints, relevant files, constraints

## File Creation

Derive the filename from the title using kebab-case (e.g., "Quote age tracking" becomes `quote-age-tracking.md`). Ensure the filename is unique within the target epic's `clarify/` directory.

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

## Design Decisions

- <decision 1>

## Implementation Notes

<notes>
```

Omit the Design Decisions and Implementation Notes sections entirely if the user didn't provide them.

Epic membership is represented by the path; do not add an `epic:` field to task frontmatter.

## Output

Confirm the file was created and show its path. Run `/status` to show the updated queue grouped by epic.
