---
title: Quickstart
description: Get started with alto in 5 minutes — from install to your first structured project
sidebar:
  order: 1
---

Get from zero to a fully structured project in under 5 minutes.

## Prerequisites

- A project idea (4-5 sentences is enough)

## Install

Download the latest binary for your platform from the [releases page](https://github.com/jaigouk/altoiddd/releases):

```bash
# macOS (Apple Silicon)
curl -L -o /usr/local/bin/alto https://github.com/jaigouk/altoiddd/releases/latest/download/alto-darwin-arm64
chmod +x /usr/local/bin/alto

# macOS (Intel)
curl -L -o /usr/local/bin/alto https://github.com/jaigouk/altoiddd/releases/latest/download/alto-darwin-amd64
chmod +x /usr/local/bin/alto

# Linux (amd64)
curl -L -o /usr/local/bin/alto https://github.com/jaigouk/altoiddd/releases/latest/download/alto-linux-amd64
chmod +x /usr/local/bin/alto
```

On Windows, download `alto-windows-amd64.exe` from the [releases page](https://github.com/jaigouk/altoiddd/releases) and add it to your `PATH`.

Or install from source if you have Go 1.26+:

```bash
go install github.com/jaigouk/altoiddd/cmd/alto@latest
```

Verify the installation:

```bash
alto version
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

Run `alto init` with the `-y` flag to skip confirmation prompts:

```bash
alto init -y
```

alto detects your installed AI coding tools, then walks you through 10 guided DDD discovery questions. Answer each one in plain language — alto adapts its vocabulary to your expertise level.

## What you get

After answering the questions, alto generates:

| Artifact | Purpose |
|----------|---------|
| `docs/PRD.md` | Product requirements derived from your answers |
| `docs/DDD.md` | Domain model — bounded contexts, aggregates, ubiquitous language |
| `docs/ARCHITECTURE.md` | Technical architecture informed by the domain model |
| `.alto/` | Project config, knowledge base, doc maintenance registry |
| `.claude/agents/` | AI agent personas (developer, tech-lead, QA, etc.) |
| `.beads/` | Dependency-ordered tickets ready for implementation |

## Next steps

- Run `alto guide` to re-enter the guided discovery flow
- Run `alto gap` to see what's missing from your project structure
- Run `alto check` to verify quality gates pass
- Read the [New Project Guide](/guides/new-project) for the full walkthrough
- Read [Concepts](/guides/concepts) to understand why alto enforces DDD before coding
