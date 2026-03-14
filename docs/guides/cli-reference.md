---
title: CLI Reference
description: Complete reference for all alty CLI commands, flags, and usage
sidebar:
  order: 5
---

## `alty init`

Bootstrap a new project from a README idea, or rescue an existing project.

```bash
alty init [flags]
```

| Flag | Description |
|------|-------------|
| `-y`, `--yes` | Skip confirmation prompt |
| `--dry-run` | Show plan without executing |
| `--existing` | Rescue an existing project (branch-based scaffolding) |

**Examples:**

```bash
# Interactive bootstrap with preview
alty init

# Quick bootstrap, skip confirmation
alty init -y

# Rescue an existing project
alty init --existing

# Preview rescue mode without writing anything
alty init --existing --dry-run
```

---

## `alty guide`

Run the 10-question guided DDD discovery flow.

```bash
alty guide [flags]
```

| Flag | Description |
|------|-------------|
| `--no-tui` | Disable TUI prompts, use plain stdin/stdout (accessibility, CI) |

Orchestrates persona detection, 10 guided questions with playback loops, and artifact generation.

**Examples:**

```bash
# Interactive guided discovery
alty guide

# Plain text mode (screen readers, scripted input)
alty guide --no-tui
```

---

## `alty detect`

Scan for installed AI coding tools and their global settings.

```bash
alty detect [project-dir]
```

Detects Claude Code, Cursor, Roo Code, OpenCode, and reports global config locations and potential conflicts. If `project-dir` is omitted, uses the current directory.

---

## `alty gap`

Analyze a project for structural gaps without modifying anything.

```bash
alty gap [project-dir]
```

Compares the project against a fully structured alty project and reports missing docs, tooling, structure, and conflicts. Read-only — nothing is written.

---

## `alty check`

Run quality gates.

```bash
alty check [flags]
```

| Flag | Description |
|------|-------------|
| `--gate <name>` | Run a specific gate: `lint`, `types`, `tests`, `fitness` |

Runs all quality gates by default: `go vet`, `golangci-lint`, `go test -race`, and architecture fitness tests.

**Examples:**

```bash
# Run all gates
alty check

# Run only lint
alty check --gate lint

# Run only tests
alty check --gate tests
```

---

## `alty generate`

Generate project artifacts. Has four subcommands.

### `alty generate artifacts`

Generate DDD artifacts (PRD, DDD.md, ARCHITECTURE.md).

```bash
alty generate artifacts
```

### `alty generate configs`

Generate tool-native configurations for AI coding tools.

```bash
alty generate configs
```

### `alty generate fitness`

Generate architecture fitness tests.

```bash
alty generate fitness
```

### `alty generate tickets`

Generate dependency-ordered beads tickets from DDD artifacts.

```bash
alty generate tickets
```

---

## `alty fitness`

Architecture fitness testing commands.

### `alty fitness generate`

Generate `arch-go.yml` configuration from the bounded context map.

```bash
alty fitness generate [flags]
```

| Flag | Description |
|------|-------------|
| `--preview` | Show what would be generated without writing |
| `--brownfield` | Use 80% compliance threshold for existing projects |
| `--dir <path>` | Project directory (default: current directory) |

Reads `.alty/bounded_context_map.yaml` and generates arch-go rules based on subdomain classification.

**Examples:**

```bash
# Generate with confirmation
alty fitness generate

# Preview only
alty fitness generate --preview

# Existing project with relaxed thresholds
alty fitness generate --brownfield
```

---

## `alty doc-health`

Check documentation freshness and health.

```bash
alty doc-health [project-dir]
```

Reports stale documents (based on `last_reviewed` frontmatter), broken references, and missing metadata.

---

## `alty doc-review`

Manage documentation review status.

### `alty doc-review list`

List documents due for review.

```bash
alty doc-review list
```

### `alty doc-review mark`

Mark a document as reviewed (updates `last_reviewed` frontmatter).

```bash
alty doc-review mark <doc-path>
```

### `alty doc-review mark-all`

Mark all stale documents as reviewed.

```bash
alty doc-review mark-all
```

---

## `alty kb`

Knowledge base operations.

### `alty kb lookup`

Look up a topic in the RLM knowledge base.

```bash
alty kb lookup <topic>
```

**Examples:**

```bash
alty kb lookup "bounded contexts"
alty kb lookup "claude-code agents"
```

### `alty kb drift`

Detect drift in knowledge base entries.

```bash
alty kb drift [tool]
```

Compares documented tool conventions against current versions and flags discrepancies.

---

## `alty persona`

Manage AI agent persona configurations.

### `alty persona list`

List all available persona definitions.

```bash
alty persona list
```

### `alty persona generate`

Generate persona configuration files for a specific AI tool.

```bash
alty persona generate <persona-name> [flags]
```

| Flag | Description |
|------|-------------|
| `--tool <name>` | Target tool: `claude-code`, `cursor`, `roo-code`, `opencode` (default: `claude-code`) |
| `-y`, `--yes` | Skip confirmation prompt |

**Examples:**

```bash
# Generate developer persona for Claude Code
alty persona generate developer

# Generate tech-lead persona for Cursor
alty persona generate tech-lead --tool cursor
```

---

## `alty ticket-health`

Show ripple review report for tickets needing attention.

```bash
alty ticket-health
```

Lists tickets flagged with `review_needed`, oldest `last_reviewed` dates, and context diffs from triggering closures.

---

## `alty ticket-verify`

Verify quantitative claims in a ticket.

```bash
alty ticket-verify <ticket-id>
```

Parses a ticket for bold number patterns (e.g., **14 findings**) and runs associated commands from code blocks to verify the claims match reality.

---

## `alty version`

Print the current alty version.

```bash
alty version
```

---

## `alty completion`

Generate shell autocompletion scripts.

```bash
alty completion [bash|zsh|fish|powershell]
```

Follow the printed instructions to enable tab completion for your shell.
