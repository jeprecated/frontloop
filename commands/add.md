---
description: Create a new task in an epic's frontloop ready or clarify queue
argument-hint: [optional epic] [task description]
---

# Frontloop Add

Create a new task file in `.frontloop/<epic>/ready/` when it is actionable now, or `.frontloop/<epic>/clarify/` when it needs human review. If no epic is specified, use the `default` epic.

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
8. **Initial Status** - use `ready` if the task has no open questions and can be worked immediately (especially if the user wants to work on it next); use `clarify` if human answers or more detail are still needed

## File Creation

Derive the filename from the title using kebab-case (e.g., "Quote age tracking" becomes `quote-age-tracking.md`).

- For `ready`, add a 4-digit ordering prefix based on priority and relation to other ready tasks in the epic. Suggested ranges: critical=0001-2499, high=2500-4999, medium=5000-7499, low=7500-9999. Ensure the filename is unique within `.frontloop/<epic>/ready/`.
- For `clarify`, do not add the numeric prefix. Ensure the filename is unique within `.frontloop/<epic>/clarify/`.

Write to `.frontloop/<epic>/<ready-or-clarify>/<filename>.md`:

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

Omit the Design Decisions and Implementation Notes sections entirely if the user didn't provide them. Only include a Questions section for tasks created in `clarify/`; never create a ready task with open Questions.

Epic membership is represented by the path; do not add an `epic:` field to task frontmatter.

## Output

Confirm the file was created and show its path. Run `/status` to show the updated queue grouped by epic.
