# frontloop

File-based task queue for AI agent loops. Tasks are Markdown files with YAML frontmatter that move between directories: `clarify` → `ready` → `in_progress` → `done`. The directory is the status. The filename is the id.

```
.frontloop/
├── clarify/      # new ideas waiting for human review
├── ready/        # reviewed and prioritised (NNNN-task-name.md)
├── in_progress/  # currently being worked on
└── done/         # completed tasks
```

## Two ways to use frontloop

### Claude Code plugin

```bash
claude plugin marketplace add ohare93/frontloop
claude plugin enable frontloop
```

Slash commands for managing tasks inside agent conversations:

| Command | Purpose |
|---------|---------|
| `/init` | Create `.frontloop/` directories in the current project |
| `/status` | Show queue state |
| `/clarify` | Review tasks in `clarify/` with a human |
| `/work` | Pick up and execute the next ready task |
| `/add` | Create a new task in `clarify/` |
| `/gather` | Collect feature ideas from user, batch-create tasks |

### `fl` CLI

A standalone Go binary for managing queues from the terminal. See [`fl/README.md`](fl/README.md) for full command reference.

```bash
fl init                          # create .frontloop/ tree
fl idea "add retry logic"        # capture an idea
fl idea -p high "fix login bug"  # with priority
fl stats                         # view queue state
fl move                          # interactive TUI to move tasks
```

## Install

### Homebrew

```bash
brew install ohare93/tap/fl
```

### Scoop (Windows)

```powershell
scoop bucket add ohare93 https://github.com/ohare93/scoop
scoop install fl
```

### Nix

```bash
nix profile install github:ohare93/frontloop
```

Or add to a flake:

```nix
{
  inputs.frontloop.url = "github:ohare93/frontloop";
  # then use inputs.frontloop.packages.${system}.default
}
```

### From source

```bash
go install github.com/ohare93/frontloop/fl/cmd/fl@latest
```

## Task format

```markdown
---
title: Quote-age tracking and stale-data guards
priority: critical
---

## Goal

Reject trades when inputs are stale.

## Acceptance Criteria

- Quotes have explicit freshness metadata
- Profiles define max staleness per input type
- Commands fail with machine-readable stale-data error
```

Frontmatter fields: `title` (required), `priority` (required — `critical`, `high`, `medium`, `low`).

Body sections: **Goal** (required), **Acceptance Criteria** (required), **Design Decisions** (optional), **Implementation Notes** (optional).

### Filename conventions

- **clarify/**: `task-name.md`
- **ready/**: `NNNN-task-name.md` (4-digit priority prefix, e.g. `0001-critical-fix.md`)
- **in_progress/**: keeps the `NNNN-` prefix from ready
- **done/**: `task-name.md` (prefix removed)

## License

[MIT](LICENSE)
