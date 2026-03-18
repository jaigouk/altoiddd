---
title: CLI Reference
description: Complete reference for all alto CLI commands, flags, and usage
sidebar:
  order: 5
---

## `alto init`

Bootstrap a new project from a README idea, or rescue an existing project.

```bash
alto init [flags]
```

| Flag | Description |
|------|-------------|
| `-y`, `--yes` | Skip confirmation prompt |
| `--dry-run` | Show plan without executing |
| `--existing` | Rescue an existing project (branch-based scaffolding) |

**Examples:**

```bash
# Interactive bootstrap with preview
alto init

# Quick bootstrap, skip confirmation
alto init -y

# Rescue an existing project
alto init --existing

# Preview rescue mode without writing anything
alto init --existing --dry-run
```

---

## `alto guide`

Run the 10-question guided DDD discovery flow.

```bash
alto guide [flags]
```

| Flag | Description |
|------|-------------|
| `--no-tui` | Disable TUI prompts, use plain stdin/stdout (accessibility, CI) |

Orchestrates persona detection, 10 guided questions with playback loops, and artifact generation.

**Examples:**

```bash
# Interactive guided discovery
alto guide

# Plain text mode (screen readers, scripted input)
alto guide --no-tui
```

---

## `alto detect`

Scan for installed AI coding tools and their global settings.

```bash
alto detect [project-dir]
```

Detects Claude Code, Cursor, Roo Code, OpenCode, and reports global config locations and potential conflicts. If `project-dir` is omitted, uses the current directory.

---

## `alto gap`

Analyze a project for structural gaps without modifying anything.

```bash
alto gap [project-dir]
```

Compares the project against a fully structured alto project and reports missing docs, tooling, structure, and conflicts. Read-only — nothing is written.

---

## `alto check`

Run quality gates.

```bash
alto check [flags]
```

| Flag | Description |
|------|-------------|
| `--gate <name>` | Run a specific gate: `lint`, `types`, `tests`, `fitness` |

Runs all quality gates by default: `go vet`, `golangci-lint`, `go test -race`, and architecture fitness tests.

**Examples:**

```bash
# Run all gates
alto check

# Run only lint
alto check --gate lint

# Run only tests
alto check --gate tests
```

---

## `alto generate`

Generate project artifacts. Has four subcommands.

### `alto generate artifacts`

Generate DDD artifacts (PRD, DDD.md, ARCHITECTURE.md).

```bash
alto generate artifacts
```

### `alto generate configs`

Generate tool-native configurations for AI coding tools.

```bash
alto generate configs
```

### `alto generate fitness`

Generate architecture fitness tests.

```bash
alto generate fitness
```

### `alto generate tickets`

Generate dependency-ordered beads tickets from DDD artifacts.

```bash
alto generate tickets
```

---

## `alto fitness`

Architecture fitness testing commands.

### `alto fitness generate`

Generate `arch-go.yml` configuration from the bounded context map.

```bash
alto fitness generate [flags]
```

| Flag | Description |
|------|-------------|
| `--preview` | Show what would be generated without writing |
| `--brownfield` | Use 80% compliance threshold for existing projects |
| `--dir <path>` | Project directory (default: current directory) |

Reads `.alto/bounded_context_map.yaml` and generates arch-go rules based on subdomain classification.

**Examples:**

```bash
# Generate with confirmation
alto fitness generate

# Preview only
alto fitness generate --preview

# Existing project with relaxed thresholds
alto fitness generate --brownfield
```

---

## `alto doc-health`

Check documentation freshness and health.

```bash
alto doc-health [project-dir]
```

Reports stale documents (based on `last_reviewed` frontmatter), broken references, and missing metadata.

---

## `alto doc-review`

Manage documentation review status.

### `alto doc-review list`

List documents due for review.

```bash
alto doc-review list
```

### `alto doc-review mark`

Mark a document as reviewed (updates `last_reviewed` frontmatter).

```bash
alto doc-review mark <doc-path>
```

### `alto doc-review mark-all`

Mark all stale documents as reviewed.

```bash
alto doc-review mark-all
```

---

## `alto kb`

Knowledge base operations.

### `alto kb lookup`

Look up a topic in the RLM knowledge base.

```bash
alto kb lookup <topic>
```

**Examples:**

```bash
alto kb lookup "bounded contexts"
alto kb lookup "claude-code agents"
```

### `alto kb drift`

Detect drift in knowledge base entries.

```bash
alto kb drift [tool]
```

Compares documented tool conventions against current versions and flags discrepancies.

---

## `alto persona`

Manage AI agent persona configurations.

### `alto persona list`

List all available persona definitions.

```bash
alto persona list
```

### `alto persona generate`

Generate persona configuration files for a specific AI tool.

```bash
alto persona generate <persona-name> [flags]
```

| Flag | Description |
|------|-------------|
| `--tool <name>` | Target tool: `claude-code`, `cursor`, `roo-code`, `opencode` (default: `claude-code`) |
| `-y`, `--yes` | Skip confirmation prompt |

**Examples:**

```bash
# Generate developer persona for Claude Code
alto persona generate developer

# Generate tech-lead persona for Cursor
alto persona generate tech-lead --tool cursor
```

---

## `alto ticket-health`

Show ripple review report for tickets needing attention.

```bash
alto ticket-health
```

Lists tickets flagged with `review_needed`, oldest `last_reviewed` dates, and context diffs from triggering closures.

---

## `alto ticket-verify`

Verify quantitative claims in a ticket.

```bash
alto ticket-verify <ticket-id>
```

Parses a ticket for bold number patterns (e.g., **14 findings**) and runs associated commands from code blocks to verify the claims match reality.

---

## `alto version`

Print the current alto version.

```bash
alto version
```

---

## `alto completion`

Generate shell autocompletion scripts.

```bash
alto completion [bash|zsh|fish|powershell]
```

Follow the printed instructions to enable tab completion for your shell.
