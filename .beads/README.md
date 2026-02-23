# Beads Issue Tracking

This project uses **Beads v0.55.4** with an embedded **Dolt** database for issue tracking.

## How It Works

```
.beads/
├── metadata.json      # Backend config (Dolt)
├── issues.jsonl       # Git-tracked JSONL export (source of truth for portability)
├── dolt/              # Embedded Dolt database (gitignored, rebuilt from JSONL)
└── .gitignore         # Managed by bd
```

- **Dolt** is the primary database — version-controlled SQL with cell-level merges.
- **issues.jsonl** is the git-portable export, auto-synced by git hooks.
- There is no daemon — Dolt runs embedded in the `bd` binary.

## Sync Model

Git hooks handle all sync automatically:

| Hook | Direction | What happens |
|------|-----------|-------------|
| `pre-commit` | Dolt → JSONL | Exports DB, stages `issues.jsonl` |
| `post-merge` | JSONL → Dolt | Imports after `git pull` |
| `pre-push` | Dolt → JSONL | Prevents pushing stale JSONL |
| `post-checkout` | JSONL → Dolt | Rebuilds DB on branch switch |

For manual export: `bd export`

> **Note:** `bd sync` is deprecated in the Dolt backend. Hooks replace it.

## Quick Reference

```bash
bd ready                         # Find available work
bd show <id>                     # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>                    # Complete work
bd export                        # Manual Dolt → JSONL export
bd list --status=open            # All open issues
bd blocked                       # Show blocked issues
bd label add <id> <label>        # Add label
bd label remove <id> <label>     # Remove label
```

## More Info

- Internal reference: `docs/reference/beads-knowledge.md`
- Upstream: [github.com/steveyegge/beads](https://github.com/steveyegge/beads)
