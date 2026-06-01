# alto-scaffold/ Scaffold — Ubiquitous Language

This document defines the terms used across this scaffold's commands, agents, templates,
and skills. Mirrors mattpocock/skills' `CONTEXT.md` role ("helps agents decode the jargon
used in the project").

## Core terms (in this order)

### Scaffold
The `alto-scaffold/` tree shipped to downstream consumers. It contains workflow assets that any
project can adopt: commands, agents, templates, skills. Lifecycle folders track maturity.

### Workflow Asset
Any `.md` file under `alto-scaffold/` (command, agent, template, or skill). Each asset carries
YAML frontmatter declaring its `name`, `description`, `kind`, `phase`, and required tools.

### GENERIC
A workflow asset that is usable in any project without modification. GENERIC assets carry
no project-specific paths, language references, or ticket prefixes. They are the canonical
shippable artifact downstream consumers receive.

### OVERLAY
Project-specific addenda extracted from a workflow asset into a sibling `.project.md` file.
The OVERLAY carries language-specific source layouts, project-specific paths, tooling
references (lint configs, test commands), and any other context that does not generalise.

### `.project.md` sibling
The file naming convention `<asset>.project.md` carrying project-local content next to
a GENERIC asset. Example: `alto-scaffold/commands/groom.md` (GENERIC) lives next to
`alto-scaffold/commands/groom.project.md` (OVERLAY). Claude Code automatically merges sibling
`.md` files when invoking a skill, so the OVERLAY loads at invocation time.

## Scope clarifications

- **`.claude/skills/` retains personal Claude Code skills.** The 18 non-symlink design/craft
  skills under `.claude/skills/` (adapt, animate, audit, etc.) and the vendored `gstack/`
  bundle (symlinks) are NOT migrated. They are personal Claude Code assets, not alto
  workflow scaffold.
- **`alto-scaffold/skills/` is reserved for shipped alto workflow skills only.** It is empty at
  migration time (placeholder `.gitkeep` only); follow-up tickets may add scaffold skills.

## Platform caveats

- **Windows POSIX symlinks.** The Phase-5 symlink bridge (`.claude/commands/*.md ->
  ../../alto-scaffold/commands/*.md`) uses `ln -s`. On Windows, Claude Code may not resolve POSIX
  symlinks created without administrator privileges. Windows users defer to the
  `additionalDirectories` settings.json mechanism (`alto init --with-scaffold` follow-up).

## File-tree contract

```
alto-scaffold/
├── CONTEXT.md            # this file
├── commands/             # invocable workflows (one .md per command)
├── agents/               # personas (one .md per agent)
├── templates/            # documentation + ticket templates
├── skills/               # reserved for shipped alto skills (empty)
└── lifecycle/
    ├── in-progress/      # assets under design, not yet stable
    └── deprecated/       # retained for migration, not for new use
```
