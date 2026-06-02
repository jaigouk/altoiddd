# alto-scaffold/

The alto workflow scaffold: commands, agents, templates, and skills you drop into a
project so Claude Code (and OpenCode) share a common playbook for DDD + TDD + SOLID
work. Author once here; consumers copy or symlink in.

For vocabulary used throughout this scaffold (Scaffold, Workflow Asset, GENERIC, OVERLAY,
`.project.md` sibling), see [CONTEXT.md](CONTEXT.md). This file owns procedures only.

## Install

### Recommended — `alto init --with-scaffold`

```bash
alto init \
  --with-scaffold \
  --project-name=<your-project> \
  --ticket-prefix=<PREFIX> \
  --issue-tracker=beads \
  --primary-tool=claude   # or: opencode
```

This writes `alto-scaffold/` into the target repo, substitutes project-name / ticket-prefix
into the [GENERIC](CONTEXT.md#generic) assets, and wires the symlink bridge below.
Supported `--primary-tool` values: `claude`, `opencode`. (Cursor / Roo are rejected
today; track follow-up tickets.)

Add `--force` to overwrite an existing `alto-scaffold/` tree.

### Manual copy

```bash
cp -r alto-scaffold/ <your-repo>/alto-scaffold/
```

Then expose the assets to Claude Code by one of:

- **Symlink bridge (POSIX)** — `.claude/commands/<name>.md -> ../../alto-scaffold/commands/<name>.md`.
  Claude Code reads the symlinked file directly. Repeat for `agents/`, `skills/`.
- **`additionalDirectories` (Windows + cross-platform fallback)** — add
  `"<repo>/alto-scaffold"` to `.claude/settings.json`'s `additionalDirectories`. Required on
  Windows; see [CONTEXT.md § Platform caveats](CONTEXT.md#platform-caveats).

## Update

When alto ships a new scaffold version:

```bash
alto init --with-scaffold --force ...   # re-runs the writer; preserves .project.md siblings
```

The contract: **GENERIC `.md` files are alto's, [OVERLAY](CONTEXT.md#overlay) `.project.md`
files are yours.** Updates overwrite GENERIC; `.project.md` siblings are not touched.

If you manually copied the tree, run a structured diff (`diff -r` or `rsync --dry-run`)
between the upstream `alto-scaffold/` and your local copy, then merge — staying clear of
`.project.md` files.

## Customize

Never edit a GENERIC asset in place. Write a sibling `<asset>.project.md` next to it:

```
alto-scaffold/commands/groom.md            # GENERIC — alto's
alto-scaffold/commands/groom.project.md    # OVERLAY — yours
```

Claude Code merges sibling `.md` files at invocation time, so the OVERLAY loads
automatically. Full mechanics are in [CONTEXT.md § `.project.md` sibling](CONTEXT.md#projectmd-sibling).

OVERLAYs are the right home for: language-specific paths (`internal/...` vs `src/...`),
lint commands (`golangci-lint run` vs `ruff check`), ticket prefix patterns, and any other
project-local context that does not generalise.

## Discover what's here

List every asset with its declared purpose:

```bash
grep -rh '^description:' alto-scaffold/{commands,agents,templates,skills} 2>/dev/null
```

Per-asset purpose lives in YAML frontmatter (`name`, `description`, `kind`, `phase`,
`when_to_use`, `tools`). There is no central inventory file — frontmatter is the
source of truth.

For the tree layout, see [CONTEXT.md § File-tree contract](CONTEXT.md#file-tree-contract).

## Where this lives in a consuming project

```
<your-repo>/
├── .claude/
│   ├── commands/        # symlinks → ../../alto-scaffold/commands/*.md   (POSIX)
│   ├── agents/          # symlinks → ../../alto-scaffold/agents/*.md
│   └── settings.json    # additionalDirectories: ["<repo>/alto-scaffold"] (Windows)
└── alto-scaffold/
    ├── CONTEXT.md
    ├── README.md        # this file
    ├── commands/  agents/  templates/  scripts/  skills/  lifecycle/
```

## Scripts

Shell scripts that scaffold assets call live in [`scripts/`](scripts/). They ship inside
the scaffold so consumers don't have to copy them separately, and `alto init --with-scaffold`
writes them with `0o755` so they're immediately runnable.

| Script | Used by | Purpose |
|---|---|---|
| `scripts/bd-ripple` | `commands/launch-team.md`, `agents/tech-lead.md`, the After-Close Protocol | Flag open dependents/siblings with `review_needed` after a ticket closes |

Scaffold assets reference scripts by their canonical path `alto-scaffold/scripts/<name>`,
not `bin/<name>`. The scaffold has no dependency on the consuming project's `bin/` layout.

When the alto CLI grows native subcommands that replace a script (e.g. a future
`alto ripple <id>`), the script becomes the fallback and the asset references the
CLI command — but for now, the bash script is the canonical implementation.

## Not in scope

- **Personal Claude Code skills** under `.claude/skills/` (design/craft skills, vendored
  bundles) are not scaffold assets and are not shipped from `alto-scaffold/`. See
  [CONTEXT.md § Scope clarifications](CONTEXT.md#scope-clarifications).
- **A consuming project's `CLAUDE.md`** governs that project. `alto-scaffold/` does not
  install or modify it.
