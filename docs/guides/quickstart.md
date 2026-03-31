---
title: Quickstart
description: Get started with alto in 5 minutes — from install to your first structured project
sidebar:
  order: 1
---

<div class="slide-container">
<iframe src="/altoiddd/alto-5min.html" style="width:100%; aspect-ratio:16/9; border:1px solid var(--sl-color-gray-5); border-radius:8px; box-shadow:0 2px 12px rgba(0,0,0,0.08);" allowfullscreen></iframe>
<p style="text-align:center; font-size:0.85em; color:var(--sl-color-gray-3); margin-top:8px;">
Use arrow keys to navigate slides. <a href="/altoiddd/alto-5min.html" target="_blank">Open fullscreen</a>
</p>
</div>

## Install

```bash
go install github.com/jaigouk/altoiddd/cmd/alto@latest
```

Or download a binary from the [releases page](https://github.com/jaigouk/altoiddd/releases). See [Installation](/altoiddd/guides/installation) for platform-specific instructions and optional tools.

## Create your first project

```bash
mkdir my-project && cd my-project
git init
cat > README.md << 'EOF'
A CLI tool that helps restaurant owners manage daily specials.
Owners enter dishes with prices and dietary tags.
The tool generates a formatted menu board and posts it to a shared display.
It tracks which specials sell out and suggests reorders.
EOF
```

Bootstrap the project and start the guided discovery flow:

```bash
alto init -y
alto guide
```

alto reads your README, detects installed AI tools, and runs a Domain Storytelling conversation to discover your domain. You answer in plain language — alto adapts to your expertise level.

## What you get

| Artifact | Purpose |
|----------|---------|
| `docs/PRD.md` | Product requirements derived from your domain stories |
| `docs/DDD.md` | Domain model — bounded contexts, aggregates, ubiquitous language |
| `docs/ARCHITECTURE.md` | Technical architecture informed by the domain model |
| `.alto/` | Project config, knowledge base, doc maintenance registry |
| `.claude/agents/` | AI agent personas (developer, tech-lead, QA, etc.) |
| `.beads/` | Dependency-ordered tickets ready for implementation |

## Next steps

- `alto guide --existing` — run discovery on an existing codebase
- `alto gap` — analyze project for structural gaps
- `alto check` — verify quality gates pass
- [New Project Guide](/altoiddd/guides/new-project) — full walkthrough
- [Concepts](/altoiddd/guides/concepts) — why alto enforces DDD before coding
