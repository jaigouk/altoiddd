---
title: Quickstart
description: Get started with alty in 5 minutes — from install to your first structured project
sidebar:
  order: 1
---

Get from zero to a fully structured project in under 5 minutes.

## Prerequisites

- Go 1.26 or later
- A project idea (4-5 sentences is enough)

## Install

```bash
go install github.com/alty-cli/alty/cmd/alty@latest
```

Verify the installation:

```bash
alty version
```

## Create your first project

Create a directory and write a short README describing your idea:

```bash
mkdir my-project && cd my-project
cat > README.md << 'EOF'
A CLI tool that helps restaurant owners manage daily specials.
Owners enter dishes with prices and dietary tags.
The tool generates a formatted menu board and posts it to a shared display.
It tracks which specials sell out and suggests reorders.
EOF
```

Run `alty init` with the `-y` flag to skip confirmation prompts:

```bash
alty init -y
```

alty detects your installed AI coding tools, then walks you through 10 guided DDD discovery questions. Answer each one in plain language — alty adapts its vocabulary to your expertise level.

## What you get

After answering the questions, alty generates:

| Artifact | Purpose |
|----------|---------|
| `docs/PRD.md` | Product requirements derived from your answers |
| `docs/DDD.md` | Domain model — bounded contexts, aggregates, ubiquitous language |
| `docs/ARCHITECTURE.md` | Technical architecture informed by the domain model |
| `.alty/` | Project config, knowledge base, doc maintenance registry |
| `.claude/agents/` | AI agent personas (developer, tech-lead, QA, etc.) |
| `.beads/` | Dependency-ordered tickets ready for implementation |

## Next steps

- Run `alty guide` to re-enter the guided discovery flow
- Run `alty gap` to see what's missing from your project structure
- Run `alty check` to verify quality gates pass
- Read the [New Project Guide](/guides/new-project) for the full walkthrough
- Read [Concepts](/guides/concepts) to understand why alty enforces DDD before coding
