# Beads (bd) Internal Knowledge Reference

> **Version**: v0.55.4 | **Backend**: Embedded Dolt | **Prefix**: `alty-k7m`
> **Source**: Distilled from [github.com/steveyegge/beads](https://github.com/steveyegge/beads) docs

## Architecture Overview

Beads is a lightweight issue tracker with first-class dependency support. It uses an embedded Dolt database (version-controlled SQL) with JSONL export for git portability.

```
.beads/
├── metadata.json      # Backend config: {"database": "dolt", "jsonl_export": "issues.jsonl"}
├── issues.jsonl       # Git-tracked JSONL (source of truth for portability)
├── dolt/              # Embedded Dolt database (gitignored, rebuilt from JSONL)
├── config.yaml        # Tool-level settings (optional)
└── .gitignore         # Managed by bd
```

### Dolt vs SQLite

| Feature | SQLite (legacy) | Dolt (current) |
|---------|----------------|----------------|
| Version control | Via JSONL export | Native cell-level |
| Multi-writer | Single process | Server mode available |
| Merge conflicts | Line-based JSONL | Cell-level 3-way merge |
| Daemon | Supported | Disabled (single-process) |
| History | Git commits | Dolt commits + Git |

### Sync Flow (Dolt + Git Hooks)

```
Pre-commit hook:  Dolt DB → export → issues.jsonl → git stage
Post-merge hook:  git pull → issues.jsonl → import → Dolt DB (branch-then-merge)
Pre-push hook:    Dolt DB → export → issues.jsonl (prevents stale push)
Post-checkout:    issues.jsonl → import → Dolt DB
```

The hooks are embedded in the `bd` binary. Install with `bd hooks install`.

## Essential Commands

### Finding Work

```bash
bd ready                          # Issues with no active blockers
bd ready -n 25                    # Show more results
bd list --status=open             # All open issues
bd list --status=in_progress      # Active work
bd blocked                        # Show blocked issues
bd stale --days 30                # Issues not updated in 30 days
```

### Creating Issues

```bash
# Basic creation (ALWAYS follow with bd update --description)
bd create "Title" -t task -p 2

# Create with description inline
bd create "Title" -t bug -p 1 -d "Description text"

# Create with description from file (avoids shell escaping)
bd create "Title" --body-file=description.md -p 1

# Create with labels
bd create "Title" -t feature -p 1 -l backend,auth

# Create child of epic (auto-numbered: k7m.1, k7m.2, etc.)
bd create "Child task" -p 1 --parent k7m

# Create and link discovered work
bd create "Found bug" -t bug -p 1 --deps discovered-from:k7m.5
```

### Updating Issues

```bash
# Status changes
bd update <id> --status in_progress
bd update <id> --claim               # Atomic claim (fails if already claimed)

# Field updates (DO NOT use bd edit - it opens $EDITOR)
bd update <id> --title "New title"
bd update <id> --description "New description"
bd update <id> --design "Design notes"
bd update <id> --notes "Additional notes"
bd update <id> --acceptance "Acceptance criteria"
bd update <id> --priority 1

# Batch updates
bd update k7m.1 k7m.2 k7m.3 --priority 0
```

### Closing Issues

```bash
bd close <id>                        # Simple close
bd close <id> --reason "Done"        # Close with reason
bd close k7m.1 k7m.2 k7m.3          # Batch close (more efficient)
bd reopen <id> --reason "Reopening"  # Reopen
```

### Dependencies

```bash
bd dep add <issue> <depends-on>      # issue depends on depends-on
bd dep add <issue> <depends-on> --type discovered-from
bd dep tree <id>                     # Show dependency tree
bd dep cycles                        # Detect circular dependencies
```

### Labels

```bash
bd label add <id> <label>            # Add label
bd label add k7m.1 k7m.2 urgent     # Batch add
bd label remove <id> <label>         # Remove label
bd label list <id>                   # List labels on issue
bd label list-all                    # All labels in use
bd list --label backend,auth         # Filter AND (must have ALL)
bd list --label-any frontend,backend # Filter OR (has ANY)
```

### Filtering & Search

```bash
bd list --title "auth"               # Title substring
bd list --desc-contains "implement"  # Description search
bd list --no-assignee                # Unassigned
bd list --empty-description          # Missing descriptions
bd list --priority-min 0 --priority-max 1  # P0 and P1 only
bd list --created-after 2026-01-01   # Date filters
bd list --type bug --status open     # Combine filters
bd query label=review_needed         # Query language
```

### Sync & Data (v0.55.4 with Dolt)

```bash
# Git hooks handle most sync automatically. Manual commands:
bd import -i .beads/issues.jsonl     # Import JSONL into Dolt
bd import --force -i .beads/issues.jsonl  # Force re-import
bd export                            # Export Dolt to JSONL

# Dolt version control
bd vc log                            # View Dolt commit history
bd vc diff HEAD~1 HEAD               # Diff between Dolt commits

# Health & diagnostics
bd doctor                            # Check for issues
bd doctor --fix --yes                # Auto-fix what's fixable
bd doctor --deep                     # Full validation
bd hooks list                        # Show installed hooks
bd hooks install                     # Install/reinstall hooks
bd info                              # Database info
```

### Project Setup

```bash
bd config set beads.role maintainer  # Set role (via git config)
bd config set create.require-description true  # Enforce descriptions
bd config set validation.on-create warn        # Template validation
bd setup claude                      # Install Claude Code hooks
bd setup claude --check              # Verify installation
```

## Issue Schema

### Types
- `bug` — Something broken
- `feature` — New functionality
- `task` — Work item (tests, docs, refactoring)
- `epic` — Large feature with child issues
- `chore` — Maintenance work

### Statuses
- `open` — Ready to be worked on
- `in_progress` — Currently being worked on
- `blocked` — Waiting on dependencies
- `deferred` — Deliberately put on ice
- `closed` — Completed
- `tombstone` — Deleted (suppresses resurrections)
- `pinned` — Stays open indefinitely

### Priorities
- `0` — Critical (security, data loss, broken builds)
- `1` — High (major features, important bugs)
- `2` — Medium (nice-to-have, minor bugs)
- `3` — Low (polish, optimization)
- `4` — Backlog (future ideas)

### Dependency Types
- `blocks` — Hard dependency (affects ready queue)
- `related` — Soft relationship (informational)
- `parent-child` — Epic/subtask hierarchy
- `discovered-from` — Work discovered during other work
- `conditional-blocks` — Runs only if dependency fails
- `waits-for` — Waits for all children of dependency

## Molecules (Work Graphs)

Molecules = epics with execution semantics. Children are parallel by default; only explicit `blocks` dependencies create sequence.

```bash
# Create molecule (just an epic)
bd create "Auth System" -t epic -p 1        # → k7m.XX
bd create "Design API" -p 1 --parent k7m.XX # → k7m.XX.1
bd create "Implement" -p 1 --parent k7m.XX  # → k7m.XX.2
bd dep add k7m.XX.2 k7m.XX.1               # Implement waits for Design

# Protos (reusable templates)
bd formula list                             # List templates
bd mol pour <proto-id> --var key=value      # Instantiate
bd mol bond <A> <B>                         # Combine work graphs

# Wisps (ephemeral, not exported to JSONL)
bd mol wisp <proto-id> --var key=value      # Create ephemeral
bd mol squash <id> --summary "Done"         # Compress to permanent
bd mol burn <id>                            # Discard without trace
```

## Git Integration

### Merge Conflict Resolution

With hash-based IDs (v0.20.1+), conflicts are rare. When they occur:

```bash
# Option 1: Accept remote
git checkout --theirs .beads/issues.jsonl
bd import -i .beads/issues.jsonl

# Option 2: Keep local
git checkout --ours .beads/issues.jsonl
bd import -i .beads/issues.jsonl

# Then commit the merge
git add .beads/issues.jsonl && git commit
```

### Intelligent Merge Driver

Auto-configured during `bd init`. Provides field-level 3-way merge:
- Timestamps → max value
- Dependencies/Labels → union
- Status/Priority → 3-way merge
- Comments → append with dedup

### Worktree Support

All worktrees share the same `.beads` database in the main repository. Dolt backend has no daemon, so no worktree-daemon conflicts.

## Key Configuration

### Config File Locations (precedence order)
1. Command-line flags (`--json`, `--no-daemon`, etc.)
2. Environment variables (`BD_JSON`, `BD_NO_DAEMON`, etc.)
3. `.beads/config.yaml` (project-specific)
4. `~/.config/bd/config.yaml` (user-specific)

### Important Environment Variables

| Variable | Purpose |
|----------|---------|
| `BD_ACTOR` | Override actor name for audit trail |
| `BD_JSON` | Always output JSON |
| `BD_SYNC_MODE` | Sync mode: `git-portable`, `dolt-native`, `belt-and-suspenders` |
| `BD_NO_DAEMON` | Force direct mode |
| `BD_DOLT_AUTO_COMMIT` | Control Dolt history commits |

### Actor Identity Resolution
1. `--actor` flag
2. `BD_ACTOR` env
3. `BEADS_ACTOR` env
4. `git config user.name`
5. `$USER`
6. `"unknown"`

## Troubleshooting

### Common Issues

**JSONL out of sync with Dolt:**
```bash
bd doctor                          # Check status
bd import --force -i .beads/issues.jsonl  # Force re-import
```

**Database corruption / rebuild:**
```bash
rm -rf .beads/dolt                 # Remove broken Dolt
bd import -i .beads/issues.jsonl   # Rebuild from JSONL (auto-bootstraps)
```

**Hooks not firing:**
```bash
bd hooks list                      # Check installed
bd hooks install --force           # Reinstall
git config core.hooksPath          # Check for override
```

**Lock contention:**
```bash
# Dolt embedded is single-writer. If locked:
ls .beads/dolt/.dolt/noms/LOCK    # Check for stale lock
bd doctor --fix                    # May clear stale locks
```

**Fresh clone bootstrap:**
On first `bd` command after cloning, JSONL is auto-imported into a fresh Dolt database. No manual steps needed. Verify with `bd vc log` (should show "Bootstrap from JSONL").

## Agent Workflow Patterns

### Claim and Complete
```bash
bd ready                           # Find work
bd update <id> --status in_progress  # Claim
# ... do work ...
bd close <id> --reason "Done"      # Complete
```

### Discover and Link
```bash
# While working on k7m.5, discover a bug:
bd create "Found auth bug" -t bug -p 1 --deps discovered-from:k7m.5
```

### Session End
Git hooks handle export/import automatically on commit/pull. For manual operations:
```bash
bd export                          # Export Dolt → JSONL
git add .beads/issues.jsonl && git commit  # Commit changes
```

## Key Docs (Upstream)

| Document | What It Covers |
|----------|---------------|
| [AGENT_INSTRUCTIONS.md](https://github.com/steveyegge/beads/blob/main/AGENT_INSTRUCTIONS.md) | Agent workflow, session protocol, "land the plane" |
| [CLI_REFERENCE.md](https://github.com/steveyegge/beads/blob/main/docs/CLI_REFERENCE.md) | Complete command reference |
| [DOLT.md](https://github.com/steveyegge/beads/blob/main/docs/DOLT.md) | Dolt backend, federation, server mode |
| [GIT_INTEGRATION.md](https://github.com/steveyegge/beads/blob/main/docs/GIT_INTEGRATION.md) | Merge conflicts, hooks, worktrees |
| [SYNC.md](https://github.com/steveyegge/beads/blob/main/docs/SYNC.md) | 3-way merge, pull-first sync, concurrency |
| [CONFIG.md](https://github.com/steveyegge/beads/blob/main/docs/CONFIG.md) | All configuration options |
| [LABELS.md](https://github.com/steveyegge/beads/blob/main/docs/LABELS.md) | Label patterns, operational state |
| [MOLECULES.md](https://github.com/steveyegge/beads/blob/main/docs/MOLECULES.md) | Work graphs, protos, bonding, wisps |
| [PROTECTED_BRANCHES.md](https://github.com/steveyegge/beads/blob/main/docs/PROTECTED_BRANCHES.md) | Sync branch workflows |

## Version History (Our Upgrade Path)

| Version | Key Change |
|---------|-----------|
| v0.49.6 | Our original version (SQLite + daemon) |
| v0.53.0 | Daemon fully removed for Dolt backend |
| v0.55.4 | Current version (Dolt embedded, no daemon, cell-level merge) |

**Upgrade lesson**: v0.49.6 had a critical race condition where multiple daemon processes re-imported stale JSONL, silently reverting label removals. Fixed by upgrading to v0.55.4 which eliminates the daemon entirely for Dolt backend.
