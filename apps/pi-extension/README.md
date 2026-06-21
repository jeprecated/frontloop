# Frontloop Pi extension

Pi extension for working with frontloop queues from inside Pi.

This extension targets the v2 epic-first frontloop layout only:

```text
.frontloop/<epic>/{clarify,ready,in_progress,done}/
```

It intentionally does not include legacy flat-queue support or migration flows.

## Install locally

From the repository root:

```bash
pi install ./apps/pi-extension
```

For one-off testing without installing:

```bash
pi -e ./apps/pi-extension
```

## Commands

| Command | Purpose |
|---------|---------|
| `/fl-status [epic]` | Show active queue state grouped by epic |
| `/fl-add` | Create a task in an epic's `clarify/` queue through a small UI wizard |
| `/fl-work [epic]` | Move the next ready task to `in_progress/` and send its context to the agent |
| `/fl-complete [epic]` | Append a completion summary and move the active task to `done/` |
| `/fl-block [epic]` | Append a blocker reason and return the active task to `clarify/` |

## Agent tools

| Tool | Purpose |
|------|---------|
| `frontloop_status` | Inspect active queue state |
| `frontloop_create_task` | Create a new task in `clarify/` |
| `frontloop_start_task` | Move the next ready task to `in_progress/` |
| `frontloop_complete_task` | Complete the active task and move it to `done/` |
| `frontloop_block_task` | Mark the active task blocked and return it to `clarify/` |

The extension also shows a compact footer status when a v2 queue is present and injects active in-progress task context into agent turns.
