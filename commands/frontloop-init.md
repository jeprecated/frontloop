---
description: Create .frontloop/ task queue directories in the current project
---

# Frontloop Init

Create the `.frontloop/` directory structure for a file-based task queue.

## Execution

1. Create the directories:

```bash
mkdir -p .frontloop/{clarify,ready,in_progress,done}
```

2. Verify they were created:

```bash
ls -d .frontloop/*/
```

3. Report what was created:

```
Initialized .frontloop/ task queue:
  clarify/      - new tasks needing human review
  ready/        - reviewed tasks, prioritized and ready to work
  in_progress/  - task currently being worked on
  done/         - completed tasks
```

If `.frontloop/` already exists, report that it's already initialized and show which directories exist.
