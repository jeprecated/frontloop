# Frontloop Plugin Commands

Turn the frontloop repo from a standalone skill with shell scripts into a plugin marketplace with registered slash commands.

## Problem

Agents ignore the helper scripts (`scripts/init.sh`, `scripts/status.sh`) and improvise with raw `ls` commands instead. Shell scripts are invisible to the plugin system — they're just docs that agents can skip.

## Solution

Replace shell scripts with plugin **commands** (markdown files registered in `marketplace.json`). Commands are first-class in the plugin system — they show up as slash commands and get invoked via the Skill tool. Add three new commands for common workflows (clarify, work, add).

## Repo Structure

```
frontloop/
├── .claude-plugin/
│   └── marketplace.json
├── SKILL.md
├── references/
│   ├── clarify.md
│   └── worker.md
└── commands/
    ├── frontloop-init.md
    ├── frontloop-status.md
    ├── frontloop-clarify.md
    ├── frontloop-work.md
    └── frontloop-add.md
```

### Deleted

- `scripts/init.sh`
- `scripts/status.sh`
- `scripts/` directory
- `frontloop.skill` (zip archive from skill-creator)

## Commands

### `/frontloop-init`

Creates `.frontloop/{clarify,ready,in_progress,done}` directories in the current project. Reports what was created. Idempotent — safe to run if directories already exist.

### `/frontloop-status`

Reads all four directories under `.frontloop/`. For each `.md` file, parses YAML frontmatter to extract `title` and `priority`. Displays a formatted summary grouped by status:

- **In Progress** — currently active tasks
- **Ready** — queued tasks sorted by priority (1=critical first)
- **Needs Clarification** — tasks awaiting human review
- **Done** — last 5 completed tasks

Shows counts per section. If `.frontloop/` doesn't exist, tells the user to run `/frontloop-init`.

### `/frontloop-clarify`

Runs the human review workflow defined in `references/clarify.md`. For each task in `clarify/`:

1. Reads the file, presents title and goal
2. If questions exist, presents each one at a time with options and recommendation
3. Records answers as design decisions
4. Removes the Questions section
5. Asks if the task is ready to work on
6. If yes, moves to `ready/` with priority prefix (`1-`, `2-`, `3-`, `4-`)
7. If no, leaves in `clarify/`

### `/frontloop-work`

Picks up the next task for execution. Follows `references/worker.md`:

1. Reads `ready/` directory sorted alphabetically (highest priority first)
2. Presents the top task's title, goal, and acceptance criteria
3. Moves the file from `ready/` to `in_progress/`
4. Executes the task following acceptance criteria and design decisions
5. On completion: appends Completion Summary section, moves to `done/` (strips priority prefix), commits with version control
6. If blocked: appends Blocked section, moves back to `clarify/`

### `/frontloop-add`

Creates a new task. Asks the user for:

1. Title
2. Goal (1-3 sentences)
3. Priority (critical/high/medium/low)
4. Acceptance criteria (checklist items)
5. Any design decisions or implementation notes (optional)

Writes a properly formatted markdown file to `clarify/` with YAML frontmatter and body sections. Filename derived from title (kebab-case).

## marketplace.json

```json
{
  "name": "frontloop",
  "owner": { "name": "ohare93" },
  "plugins": [
    {
      "name": "frontloop",
      "source": "./",
      "description": "File-based task queue for agent loops with slash commands for queue management",
      "version": "1.0.0",
      "author": { "name": "ohare93" },
      "commands": [
        "./commands/frontloop-init.md",
        "./commands/frontloop-status.md",
        "./commands/frontloop-clarify.md",
        "./commands/frontloop-work.md",
        "./commands/frontloop-add.md"
      ],
      "skills": [
        "./"
      ]
    }
  ]
}
```

Note: The skill is referenced as `"./"` since SKILL.md lives at the repo root (not in a `skills/` subdirectory).

## SKILL.md Changes

Remove the Scripts section. Replace with a Commands section:

```markdown
## Commands

| Command | Purpose |
|---------|---------|
| `/frontloop-init` | Create `.frontloop/` directories in the current project |
| `/frontloop-status` | Show queue state |
| `/frontloop-clarify` | Review tasks in `clarify/` with a human |
| `/frontloop-work` | Pick up and execute the next ready task |
| `/frontloop-add` | Create a new task in `clarify/` |
```

## What Stays the Same

- SKILL.md frontmatter and description (auto-invocation behavior unchanged)
- Task file format (frontmatter, body sections)
- Filename conventions (priority prefixes, kebab-case)
- Workflow definitions in `references/clarify.md` and `references/worker.md`
- `.frontloop/` directory structure
- `.gitignore`
