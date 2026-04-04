---
description: Show the current frontloop task queue state
---

# Frontloop Status

Display the state of the `.frontloop/` task queue.

## Precondition

Check that `.frontloop/` exists. If not, tell the user to run `/frontloop-init`.

## Execution

Read all `.md` files in each of the four directories. For each file, parse the YAML frontmatter to extract `title` and `priority`.

Display the results in this format:

```
=== Frontloop Status ===

IN PROGRESS (N):
  * <title>

READY (N):
  <filename>  [<priority>]  <title>

NEEDS CLARIFICATION (N):
  <filename>  [<priority>]  <title>

DONE (N):
  <filename>  <title>
```

### Rules

- **In Progress**: Show all tasks. There should be 0 or 1.
- **Ready**: Sort files alphabetically (priority prefix ensures highest priority first). Show all.
- **Needs Clarification**: Show all.
- **Done**: Show the 5 most recently modified files. If more than 5, append `... and N more`.
- Empty sections show `(empty)` instead of file listings.
- Filenames are shown without the `.md` extension.
