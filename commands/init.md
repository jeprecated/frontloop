---
description: Create the v2 .frontloop/ epic task queue in the current project
---

# Frontloop Init

Create the v2 `.frontloop/` directory structure for an epic-scoped file-based task queue.

## Execution

1. Check whether `.frontloop/` already exists.

2. If it does not exist, create the v2 tree:

```bash
mkdir -p .frontloop/default/{clarify,ready,in_progress,done}
mkdir -p .frontloop/_archive
```

3. Create `.frontloop/default/epic.md` if it does not already exist:

```markdown
---
title: Default
slug: default
status: active
created_at: <YYYY-MM-DD>
completed_at:
---

## Goal

```

4. Verify the directories:

```bash
ls -d .frontloop/default/{clarify,ready,in_progress,done} .frontloop/_archive
```

5. Report what was created:

```text
Initialized v2 .frontloop task queue:
  default/      - built-in epic for unscoped tasks
    clarify/    - new tasks needing human review
    ready/      - reviewed tasks, prioritized and ready to work
    in_progress/ - task currently being worked on
    done/       - completed tasks
  _archive/     - completed epics ignored by active workflows
```

## Existing Queues

If `.frontloop/default/` already exists, report that the v2 queue is already initialized and show the active epic directories.

If the old flat layout exists with top-level `.frontloop/{clarify,ready,in_progress,done}/`, do **not** keep using it as active state. Tell the user to migrate it into the default epic:

```bash
fl migrate epic-layout
```

Migration preserves task filenames and contents by moving legacy tasks to `.frontloop/default/<status>/` and creates `_archive/`.
