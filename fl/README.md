# fl

Command-line tool for managing a [frontloop](https://github.com/ohare93/frontloop) task queue. Tasks live as Markdown files in a `.frontloop/` directory tree; `fl` lets you capture ideas, inspect queue state, and move tasks between queues — all from the terminal.

## Commands

### `fl idea`

Quickly capture a task idea into the clarify queue.

```bash
fl idea "add retry logic to the API client"
fl idea -p high "fix login race condition"
fl idea --priority critical "database is down"
```

Flags:
- `-p`, `--priority` — `critical`, `high`, `medium` (default), or `low`

The description is slugified into a filename and written to `.frontloop/clarify/`.

### `fl stats`

Show queue state at a glance.

```bash
fl stats
fl stats --no-color   # pipe-safe, no ANSI codes
fl stats --no-color | grep READY
```

Displays counts and task names for all four queues: IN PROGRESS, READY, NEEDS CLARIFICATION, and DONE (capped at 5 most-recent).

### `fl move`

Interactively move tasks between queues with a TUI.

```bash
fl move
```

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

## Installation

### From source (requires Go 1.21+)

```bash
go install github.com/ohare93/fl/cmd/fl@latest
```

### With devbox (recommended for development)

```bash
devbox shell          # enter the devbox environment
devbox run build      # builds ./fl
devbox run test       # runs all tests
```

The built binary is placed in the current directory as `./fl`.

## Directory structure

`fl` searches upward from the current directory for a `.frontloop/` tree:

```
.frontloop/
├── clarify/      # new ideas waiting for human review
├── ready/        # reviewed and prioritised (NNNN-task-name.md)
├── in_progress/  # currently being worked on
└── done/         # completed tasks
```

If no `.frontloop/` directory is found, commands print a helpful error:

```
no .frontloop directory found (run fl init to create one)
```

## Development

Dependencies are managed with [devbox](https://www.jetify.com/devbox/) and [Go modules](https://go.dev/ref/mod).

```bash
devbox shell
devbox run test    # go test -v ./...
devbox run build   # go build -o fl ./cmd/fl
```
