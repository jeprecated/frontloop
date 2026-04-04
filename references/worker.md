# Worker Workflow

## Before Starting

Move the task file from `ready/` to `in_progress/`.

## During

Follow the task's acceptance criteria exactly. Respect design decisions — they were pre-approved by a human. Do not modify Goal, Acceptance Criteria, or Design Decisions sections.

If the task cannot be completed as described, append a **Blocked** section explaining why and move the file to `.frontloop/clarify/` instead.

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

2. Move the file to `.frontloop/done/` (remove the priority prefix from the filename).
3. Commit changes with version control.
