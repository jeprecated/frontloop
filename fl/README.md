# fl

Command-line tool for managing a [frontloop](https://github.com/jeprecated/frontloop) task queue. Tasks live as Markdown files in a v2 epic-first `.frontloop/` tree; `fl` lets you capture ideas, inspect queue state, manage epics, archive completed epics, and move tasks between queues from the terminal.

## Directory structure

`fl` searches upward from the current directory for a `.frontloop/` tree. New queues use the v2 epic layout:

```text
.frontloop/
├── default/
│   ├── epic.md
│   ├── clarify/
│   ├── ready/
│   ├── in_progress/
│   └── done/
├── checkout-redesign/
│   ├── epic.md
│   ├── clarify/
│   ├── ready/
│   ├── in_progress/
│   └── done/
└── _archive/
```

Active task paths are `.frontloop/<epic>/<status>/<task>.md`. `default/` receives tasks when no epic is specified. `_archive/` contains completed epics and is ignored by normal active-queue commands.

Within an epic, task status is still represented by directory movement: `clarify` → `ready` → `in_progress` → `done`.

## Commands

### `fl init`

Create a v2 `.frontloop/` directory tree in the current directory.

```bash
fl init
```

Creates:

```text
.frontloop/default/{clarify,ready,in_progress,done}/
.frontloop/default/epic.md
.frontloop/_archive/
```

Safe to run multiple times — existing task files and existing `epic.md` metadata are left untouched. If legacy flat task files are present, `fl init` asks you to run `fl migrate epic-layout` instead.

### `fl migrate epic-layout`

Move an old flat queue into the v2 `default` epic.

```bash
fl migrate epic-layout
```

Migrates:

```text
.frontloop/<status>/<task>.md → .frontloop/default/<status>/<task>.md
```

The command preserves filenames and file contents, creates `default/epic.md`, creates `_archive/`, and refuses destination conflicts.

### `fl epic new`

Create a new active epic.

```bash
fl epic new checkout-redesign
```

Creates `.frontloop/checkout-redesign/` with `epic.md` plus all four status directories. Epic slugs use lower-case letters, digits, and hyphens. Reserved names such as `_archive`, `clarify`, `ready`, `in_progress`, and `done` are rejected.

### `fl epic list`

List active epics.

```bash
fl epic list
```

Archived epics under `_archive/` are not listed.

### `fl epic archive`

Archive a completed active epic.

```bash
fl epic archive checkout-redesign
```

The epic must have no tasks in `clarify/`, `ready/`, or `in_progress/`. Completed tasks may remain in `done/`. The command moves the epic to `.frontloop/_archive/YYYY-MM-DD-<slug>/`, updates `epic.md` with archived metadata, and prints manual restore guidance. The `default` epic cannot be archived.

### `fl idea`

Quickly capture a task idea into an epic's clarify queue.

```bash
fl idea "add retry logic to the API client"                       # default epic
fl idea -p high "fix login race condition"                        # default epic, high priority
fl idea --priority critical "database is down"                    # default epic
fl idea --epic checkout-redesign "render checkout review page"    # selected epic
```

Flags:

- `-p`, `--priority` — `critical`, `high`, `medium` (default), or `low`
- `--epic` — active epic slug to receive the task; defaults to `default`

The description is slugified into a filename and written to `.frontloop/<epic>/clarify/`. `--epic` must name an existing active epic; use `fl epic new <slug>` first.

### `fl stats`

Show active queue state grouped by epic.

```bash
fl stats
fl stats --epic checkout-redesign
fl stats --no-color   # pipe-safe, no ANSI codes
fl stats --no-color | grep READY
```

Displays counts and task names for each active epic's IN PROGRESS, READY, NEEDS CLARIFICATION, and DONE queues. Done lists are capped at 5 most-recent tasks per epic. `_archive/` is ignored.

### `fl move`

Interactively move active tasks between queues with a TUI.

```bash
fl move
```

The TUI shows each task's epic, status, filename, and title so duplicate filenames in different epics are unambiguous. Moves preserve epic membership.

Keys:

- `j` / `↓` — move cursor down
- `k` / `↑` — move cursor up
- `Enter` — select a task and choose its destination queue
- `Esc` — cancel destination selection
- `q` / `Ctrl+C` — quit

### `fl completion`

Generate shell completion scripts.

```bash
# bash — add to ~/.bashrc
source <(fl completion bash)

# zsh — add to ~/.zshrc
source <(fl completion zsh)

# fish — one-time setup
fl completion fish > ~/.config/fish/completions/fl.fish
```

## Filename conventions

Within each active epic:

- `clarify/`: `task-name.md`
- `ready/`: `NNNN-task-name.md`
- `in_progress/`: same filename as ready
- `done/`: preserves the `NNNN-` prefix from ready/in_progress

Default priority prefixes are four-digit ranges: critical `0001-`, high `2500-`, medium `5000-`, low `7500-`.

## Error handling

If no `.frontloop/` directory is found, commands print a helpful error:

```text
no .frontloop directory found (run fl init to create one)
```

If an old flat queue is detected, v2-only commands tell you to run:

```bash
fl migrate epic-layout
```

## Installation

### From source (requires Go 1.24+)

```bash
go install github.com/jeprecated/frontloop/fl/cmd/fl@latest
```

### With Nix for development

```bash
nix shell nixpkgs#go -c sh -lc 'cd fl && go test ./...'
nix build
```

## Development

Dependencies are managed with [Go modules](https://go.dev/ref/mod). From the `fl/` directory:

```bash
go test ./...
go build -o fl ./cmd/fl
```
