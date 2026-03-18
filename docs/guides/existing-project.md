---
title: Existing Project
description: Apply alto structure to an existing codebase with rescue mode and gap analysis
sidebar:
  order: 4
---

alto can analyze and scaffold existing projects without disrupting your current code. Two tools handle this: `alto init --existing` for full rescue mode, and `alto gap` for read-only analysis.

## Gap analysis (read-only)

Run `alto gap` to see what your project is missing compared to a fully structured alto project — without changing anything:

```bash
cd your-existing-project
alto gap
```

The gap report shows:

- **Missing docs** — PRD, DDD.md, ARCHITECTURE.md
- **Missing tooling** — `.claude/`, `.beads/`, quality gate configs
- **Missing structure** — DDD layers, test mirrors, bounded context directories
- **Conflicts** — existing files that would conflict with alto defaults

This is a safe, non-destructive scan. Use it to evaluate what rescue mode would do before committing to it.

You can also point `alto gap` at a specific directory:

```bash
alto gap /path/to/project
```

## Rescue mode (`alto init --existing`)

Rescue mode applies alto structure to an existing project on a separate branch. It never touches your current branch.

### Prerequisites

Your git working tree must be clean (no uncommitted changes):

```bash
git status
# On branch main
# nothing to commit, working tree clean
```

If you have uncommitted changes, alto refuses to run. Commit or stash first.

### Running rescue mode

```bash
alto init --existing
```

alto performs these steps in order:

1. **Branch creation** — creates an `alto/init` branch. If the branch already exists, alto aborts. Delete the existing branch first or use a clean repository.
2. **Project scan** — analyzes your code, docs, configs, and folder structure.
3. **Gap analysis** — compares against a fully seeded project.
4. **Gap report** — shows what's missing, what conflicts, and what alto proposes to add.
5. **Guided questions** — asks clarifying DDD questions about your existing domain (bounded contexts, ubiquitous language).
6. **Artifact generation** — generates missing artifacts adapted to your domain language.
7. **Agent adaptation** — configures AI agent personas using your domain terms.
8. **Test gate** — runs your existing test suite. If any test fails, all changes are rolled back.
9. **Ready for review** — you review the branch diff and merge when satisfied.

### Safety rules

| Rule | Behavior |
|------|----------|
| Branch isolation | All changes go to `alto/init`, never your current branch |
| Clean tree required | Refuses to run with uncommitted changes |
| Never overwrites | Existing files are skipped, not replaced |
| Conflict rename | If alto needs to create a file that exists, it suffixes: `filename_alto.md` |
| Zero test regression | Runs your test suite after scaffolding — rolls back on any failure |
| Never merges | You merge the branch manually after review |

### Reviewing the result

After rescue mode completes, review the branch diff:

```bash
git diff main..alto/init
```

If you're satisfied, merge:

```bash
git checkout main
git merge alto/init
```

If you're not satisfied, delete the branch and start over:

```bash
git branch -D alto/init
```

### Preview mode

Use `--dry-run` to see what rescue mode would do without creating a branch or writing files:

```bash
alto init --existing --dry-run
```

## When to use gap vs rescue

| Scenario | Command |
|----------|---------|
| "What's missing in my project?" | `alto gap` |
| "I want to adopt alto structure" | `alto init --existing` |
| "Show me the plan before I commit" | `alto init --existing --dry-run` |
