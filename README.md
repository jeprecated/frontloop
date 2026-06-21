# frontloop

File-based task queue for AI agent loops. Tasks are Markdown files with YAML frontmatter that move through status directories inside an epic: `clarify` → `ready` → `in_progress` → `done`.

The active v2 layout is epic-first:

```text
.frontloop/
├── default/                 # built-in bucket for unscoped tasks
│   ├── epic.md
│   ├── clarify/             # new tasks always start here
│   ├── ready/               # reviewed and prioritised
│   ├── in_progress/         # currently being worked on
│   └── done/                # completed tasks
├── checkout-redesign/       # another active epic
│   ├── epic.md
│   ├── clarify/
│   ├── ready/
│   ├── in_progress/
│   └── done/
└── _archive/                # archived completed epics, ignored by active commands
```

Active task paths use:

```text
.frontloop/<epic>/<status>/<task>.md
```

`default/` is the epic used when no explicit epic is provided. `_archive/` stores completed epics and is ignored by normal active-queue commands.

## Ways to use frontloop

### Claude Code plugin

```bash
claude plugin marketplace add jeprecated/frontloop
claude plugin enable frontloop
```

Slash commands for managing tasks inside agent conversations:

| Command | Purpose |
|---------|---------|
| `/init` | Create the v2 `.frontloop/` tree with `default/` and `_archive/` |
| `/status` | Show active queue state grouped by epic |
| `/clarify` | Review tasks in an epic's `clarify/` queue with a human |
| `/work` | Pick up and execute the next ready task, optionally within one epic |
| `/add` | Create a new task in `.frontloop/<epic>/clarify/` |
| `/gather` | Collect feature ideas and batch-create clarify tasks |

### `fl` CLI

A standalone Go binary for managing queues from the terminal. The CLI app lives in `apps/fl/`; see [`apps/fl/README.md`](apps/fl/README.md) for full command reference.

```bash
fl init                                      # create v2 .frontloop/ tree
fl migrate epic-layout                      # migrate old flat queues into default/
fl epic new checkout-redesign               # create an active epic
fl idea --epic checkout-redesign "render review page"
fl idea "small unscoped task"               # goes to default/clarify/
fl stats                                    # grouped by active epic
fl stats --epic checkout-redesign           # one epic only
fl epic archive checkout-redesign           # archive a completed epic
fl move                                     # interactive TUI to move tasks
```

### Pi extension

A Pi extension app lives in `apps/pi-extension/`. It adds Pi-native `/fl-*` commands, frontloop tools for the agent, a compact queue status in the footer, and active-task context injection.

```bash
pi install ./apps/pi-extension
# or test without installing
pi -e ./apps/pi-extension
```

The Pi extension intentionally targets only the v2 epic-first layout and does not include legacy flat-queue migration support.

## Install

### Homebrew

```bash
brew install jeprecated/tap/fl
```

### Scoop (Windows)

```powershell
scoop bucket add jeprecated https://github.com/jeprecated/scoop
scoop install fl
```

### Nix

```bash
nix profile install github:jeprecated/frontloop
```

Or add to a flake:

```nix
{
  inputs.frontloop.url = "github:jeprecated/frontloop";
  # then use inputs.frontloop.packages.${system}.default
}
```

### From source

```bash
go install github.com/jeprecated/frontloop/apps/fl/cmd/fl@latest
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

Frontmatter fields: `title` (required), `priority` (required — `critical`, `high`, `medium`, `low`). Epic membership comes from the path, not task frontmatter.

Body sections: **Goal** (required), **Acceptance Criteria** (required), **Design Decisions** (optional), **Implementation Notes** (optional).

### Filename conventions

Within each epic:

- **clarify/**: `task-name.md`
- **ready/**: `NNNN-task-name.md` where `NNNN` is a zero-padded ordering prefix
- **in_progress/**: keeps the `NNNN-` prefix from ready
- **done/**: preserves the `NNNN-` prefix so completed/archived epics remain readable in execution order

Suggested prefix ranges: critical `0001-2499`, high `2500-4999`, medium `5000-7499`, low `7500-9999`.

## Epic lifecycle

Create active epics with `fl epic new <slug>`. When an epic is complete, archive it with `fl epic archive <slug>`. Archiving is allowed only when that epic has no tasks in `clarify/`, `ready/`, or `in_progress/`; completed tasks may remain in `done/`.

Archived epics move to:

```text
.frontloop/_archive/YYYY-MM-DD-<epic>/
```

Normal status, idea, move, and work flows ignore `_archive/`.

## Migrating old queues

The legacy flat layout was:

```text
.frontloop/{clarify,ready,in_progress,done}/
```

Run:

```bash
fl migrate epic-layout
```

This moves legacy tasks into `.frontloop/default/<status>/`, preserves filenames and contents, creates `default/epic.md`, and creates `_archive/`.

For the full v2 filesystem contract, see [`docs/frontloop-v2-epic-layout.md`](docs/frontloop-v2-epic-layout.md).

## License

[MIT](LICENSE)
