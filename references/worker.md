# Worker Workflow

Workers execute tasks from `.frontloop/<epic>/ready/`. The epic is part of the task path and must be preserved for every status move.

`default` is the built-in epic for unscoped tasks. Ignore `.frontloop/_archive/`; archived epics are historical and are not active work queues.

## Before Starting

1. Select the active epic to work. If no epic is specified and multiple epics have ready tasks, ask the user which epic to work rather than mixing them.
2. Pick the first file in `.frontloop/<epic>/ready/` sorted alphabetically. The 4-digit filename prefix defines order within that epic.
3. Move the task file from `.frontloop/<epic>/ready/` to `.frontloop/<epic>/in_progress/`.
4. Keep the same filename, including the numeric prefix.

## During

Follow the task's acceptance criteria exactly. Respect design decisions — they were pre-approved by a human. Do not modify Goal, Acceptance Criteria, or Design Decisions sections.

If the task cannot be completed as described:

1. Append a **Blocked** section explaining why.
2. Move the file back to `.frontloop/<epic>/clarify/`.
3. Strip the numeric priority/order prefix when returning to `clarify/`.
4. Report the blocker.

Do not move the task to another epic unless the human explicitly requests that.

## After Completing

1. Append a **Completion Summary** section to the task file:

```markdown
## Completion Summary

- Implemented configurable TTL per input type in profile config
- Trade execution boundary rejects stale inputs before strategy evaluation

### Files Changed

- internal/freshness/check.go (new)
- internal/execution/run.go (modified)
- internal/execution/run_test.go (modified)
```

2. Move the file to `.frontloop/<epic>/done/`.
3. Preserve the numeric prefix in `done/` so completed tasks remain readable in execution order, especially after the epic is archived.
4. Commit changes with the project's version-control workflow when appropriate.
