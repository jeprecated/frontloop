# Frontloop Plugin Commands Implementation Plan

> **Historical note:** This dated plan describes the pre-v2 legacy flat queue layout. Current frontloop behavior uses the epic-first layout documented in `../../frontloop-v2-epic-layout.md`.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the frontloop repo into a plugin marketplace with 5 slash commands replacing the shell scripts.

**Architecture:** Single-plugin marketplace. Commands are markdown files in `commands/` registered via `.claude-plugin/marketplace.json`. The existing SKILL.md stays for auto-invocation but drops its scripts reference in favor of a commands table.

**Tech Stack:** Claude Code plugin system (marketplace.json, command markdown files), jj for version control.

---

## File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `.claude-plugin/marketplace.json` | Registers the plugin, its commands and skill |
| Create | `commands/frontloop-init.md` | `/frontloop-init` slash command |
| Create | `commands/frontloop-status.md` | `/frontloop-status` slash command |
| Create | `commands/frontloop-clarify.md` | `/frontloop-clarify` slash command |
| Create | `commands/frontloop-work.md` | `/frontloop-work` slash command |
| Create | `commands/frontloop-add.md` | `/frontloop-add` slash command |
| Modify | `SKILL.md` | Remove scripts section, add commands table |
| Delete | `scripts/init.sh` | Replaced by `/frontloop-init` |
| Delete | `scripts/status.sh` | Replaced by `/frontloop-status` |
| Delete | `scripts/` | Empty after script removal |
| Delete | `frontloop.skill` | Zip archive, no longer needed |

---

### Task 1: Create marketplace.json

**Files:**
- Create: `.claude-plugin/marketplace.json`

- [ ] **Step 1: Create the `.claude-plugin/` directory**

```bash
mkdir -p .claude-plugin
```

- [ ] **Step 2: Write marketplace.json**

```json
{
  "name": "frontloop",
  "owner": { "name": "jeprecated" },
  "plugins": [
    {
      "name": "frontloop",
      "source": "./",
      "description": "File-based task queue for agent loops with slash commands for queue management",
      "version": "1.0.0",
      "author": { "name": "jeprecated" },
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

- [ ] **Step 3: Commit**

```bash
jj commit -m "add marketplace.json for frontloop plugin"
```

---

### Task 2: Create `/frontloop-init` command

**Files:**
- Create: `commands/frontloop-init.md`

- [ ] **Step 1: Create the `commands/` directory**

```bash
mkdir -p commands
```

- [ ] **Step 2: Write `commands/frontloop-init.md`**

```markdown
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
```

- [ ] **Step 3: Commit**

```bash
jj commit -m "add /frontloop-init command"
```

---

### Task 3: Create `/frontloop-status` command

**Files:**
- Create: `commands/frontloop-status.md`

- [ ] **Step 1: Write `commands/frontloop-status.md`**

```markdown
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
```

- [ ] **Step 2: Commit**

```bash
jj commit -m "add /frontloop-status command"
```

---

### Task 4: Create `/frontloop-add` command

**Files:**
- Create: `commands/frontloop-add.md`

- [ ] **Step 1: Write `commands/frontloop-add.md`**

```markdown
---
description: Create a new task in the frontloop clarify queue
argument-hint: [task description]
---

# Frontloop Add

Create a new task file in `.frontloop/clarify/`.

## Precondition

Check that `.frontloop/` exists. If not, tell the user to run `/frontloop-init`.

## Arguments

`{{arg}}`

If arguments are provided, use them as the initial task description to seed the conversation below.

## Execution

Gather the following from the user (skip any that are already clear from the arguments):

1. **Title** - short human-readable name for the task
2. **Goal** - what the task achieves (1-3 sentences)
3. **Priority** - critical, high, medium, or low
4. **Acceptance Criteria** - concrete checklist items for completion
5. **Design Decisions** (optional) - any pre-answered choices
6. **Implementation Notes** (optional) - hints, relevant files, constraints

## File Creation

Derive the filename from the title using kebab-case (e.g., "Quote age tracking" becomes `quote-age-tracking.md`).

Write to `.frontloop/clarify/<filename>.md`:

```markdown
---
title: <title>
priority: <priority>
---

## Goal

<goal>

## Acceptance Criteria

- <criterion 1>
- <criterion 2>
- ...

## Design Decisions

- <decision 1>

## Implementation Notes

<notes>
```

Omit the Design Decisions and Implementation Notes sections entirely if the user didn't provide them.

## Output

Confirm the file was created and show its path. Run `/frontloop-status` to show the updated queue.
```

- [ ] **Step 2: Commit**

```bash
jj commit -m "add /frontloop-add command"
```

---

### Task 5: Create `/frontloop-clarify` command

**Files:**
- Create: `commands/frontloop-clarify.md`
- Reference: `references/clarify.md`

- [ ] **Step 1: Write `commands/frontloop-clarify.md`**

```markdown
---
description: Review tasks in the clarify queue with a human
---

# Frontloop Clarify

Run the human review workflow on tasks in `.frontloop/clarify/`.

## Precondition

Check that `.frontloop/clarify/` exists. If it doesn't, tell the user to run `/frontloop-init`.

If no `.md` files are in `clarify/`, report "No tasks need clarification" and exit.

## Workflow

Read the full workflow from `references/clarify.md` in the frontloop skill directory.

For each `.md` file in `.frontloop/clarify/`:

### 1. Present the task

Read the file. Show the title, goal, and acceptance criteria to the user.

### 2. Handle questions

If the task has a **Questions** section:
- Present each question one at a time
- Show the lettered options and the recommendation
- Record the user's answer

If the task has no Questions section:
- Ask the user if the goal and acceptance criteria are clear enough for an agent to execute independently

### 3. Update the file

- Write chosen answers into the **Design Decisions** section as concise statements
- **Remove the entire Questions section** — keep only the decisions
- Add any implementation guidance the user provides to **Implementation Notes**

### 4. Promote or keep

Ask the user if this task is ready to work on.

- **If yes**: Move to `.frontloop/ready/` with a 4-digit numeric prefix. Pick a number based on priority and relation to other tasks. Suggested ranges: critical=0001-2499, high=2500-4999, medium=5000-7499, low=7500-9999. The filename becomes `NNNN-<original-filename>`.
- **If no**: Leave in `clarify/`. Ask what's missing and add it to the file.

## Output

After processing all tasks, run `/frontloop-status` to show the updated queue.
```

- [ ] **Step 2: Commit**

```bash
jj commit -m "add /frontloop-clarify command"
```

---

### Task 6: Create `/frontloop-work` command

**Files:**
- Create: `commands/frontloop-work.md`
- Reference: `references/worker.md`

- [ ] **Step 1: Write `commands/frontloop-work.md`**

```markdown
---
description: Pick up and execute the next ready task from the frontloop queue
---

# Frontloop Work

Pick up the highest-priority task from `.frontloop/ready/` and execute it.

## Precondition

Check that `.frontloop/ready/` exists. If it doesn't, tell the user to run `/frontloop-init`.

If no `.md` files are in `ready/`, report "No tasks ready for work" and exit.

If any `.md` files are in `in_progress/`, report "A task is already in progress" and show its title. Ask the user if they want to continue that task or abandon it (move back to `ready/`).

## Workflow

Read the full workflow from `references/worker.md` in the frontloop skill directory.

### 1. Pick the task

List `.md` files in `ready/` sorted alphabetically. The first file is the highest priority (filenames are prefixed with a 4-digit number; lowest number = highest priority).

Read the file. Present the title, goal, acceptance criteria, and any design decisions to the user.

### 2. Move to in_progress

Move the file from `ready/` to `in_progress/` (keep the same filename including priority prefix).

### 3. Execute

Follow the task's acceptance criteria exactly. Respect design decisions — they were pre-approved by a human. Do not modify Goal, Acceptance Criteria, or Design Decisions sections.

If the task cannot be completed as described:
- Append a **Blocked** section explaining why
- Move the file back to `.frontloop/clarify/` (strip priority prefix)
- Report the blocker and exit

### 4. Complete

Append a **Completion Summary** section to the task file:

```markdown
## Completion Summary

- <what was done, one bullet per change>

### Files Changed

- <path> (new/modified/deleted)
```

Move the file to `.frontloop/done/` and strip the priority prefix from the filename.

Commit changes with version control.

### 5. Report

Run `/frontloop-status` to show the updated queue.
```

- [ ] **Step 2: Commit**

```bash
jj commit -m "add /frontloop-work command"
```

---

### Task 7: Update SKILL.md

**Files:**
- Modify: `SKILL.md` (lines 73-80)

- [ ] **Step 1: Read current SKILL.md**

Verify the file still has the Scripts section starting around line 73.

- [ ] **Step 2: Replace Scripts section with Commands section**

Remove lines 73-80 (the Scripts section) and replace with:

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

- [ ] **Step 3: Commit**

```bash
jj commit -m "update SKILL.md: replace scripts with commands table"
```

---

### Task 8: Delete old files

**Files:**
- Delete: `scripts/init.sh`
- Delete: `scripts/status.sh`
- Delete: `scripts/` (directory)
- Delete: `frontloop.skill`

- [ ] **Step 1: Delete scripts and zip archive**

```bash
rm scripts/init.sh scripts/status.sh
rmdir scripts
rm frontloop.skill
```

- [ ] **Step 2: Verify deletions**

```bash
ls scripts/ 2>&1    # should say "No such file or directory"
ls frontloop.skill 2>&1  # should say "No such file or directory"
```

- [ ] **Step 3: Commit**

```bash
jj commit -m "remove shell scripts and skill zip, replaced by plugin commands"
```

---

### Task 9: Verify the plugin structure

- [ ] **Step 1: Check final directory structure**

```bash
find . -not -path './.jj/*' -not -path './.git/*' -not -path './.frontloop/*' | sort
```

Expected output:
```
.
./.claude-plugin
./.claude-plugin/marketplace.json
./.gitignore
./commands
./commands/frontloop-add.md
./commands/frontloop-clarify.md
./commands/frontloop-init.md
./commands/frontloop-status.md
./commands/frontloop-work.md
./docs
./docs/superpowers
./docs/superpowers/plans
./docs/superpowers/plans/2026-04-04-frontloop-plugin-commands.md
./docs/superpowers/specs
./docs/superpowers/specs/2026-04-04-frontloop-plugin-commands-design.md
./references
./references/clarify.md
./references/worker.md
./SKILL.md
```

- [ ] **Step 2: Validate marketplace.json is valid JSON**

```bash
cat .claude-plugin/marketplace.json | python3 -m json.tool > /dev/null
```

Expected: no output (success).

- [ ] **Step 3: Verify each command file has valid frontmatter**

Read each file in `commands/` and confirm it starts with `---` frontmatter containing at least a `description` field.

- [ ] **Step 4: Confirm SKILL.md no longer references scripts**

Search SKILL.md for "scripts" or "init.sh" or "status.sh". Should find no matches.
