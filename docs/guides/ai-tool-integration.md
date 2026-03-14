---
title: AI Tool Integration
description: How alty works with Claude Code, Cursor, Roo Code, and OpenCode
sidebar:
  order: 6
---

alty generates domain-aware configurations for multiple AI coding tools from a single domain model. The same bounded contexts, ubiquitous language, and quality gates apply regardless of which tool you use.

## Supported tools

| Tool | Config Location | Status |
|------|----------------|--------|
| Claude Code | `.claude/` | Fully supported |
| Cursor | `.cursor/` | Supported |
| Roo Code | `.roo/` | Planned |
| OpenCode | `.opencode/` | Planned |

Detect which tools are installed on your system:

```bash
alty detect
```

## How it works

During `alty init`, alty detects your installed tools and generates configurations for each one. The generation pipeline:

1. **Domain model** (from guided discovery) provides bounded contexts, ubiquitous language, and subdomain classifications
2. **Tool Translation** maps domain concepts to each tool's native config format
3. **Knowledge Base** supplies current tool conventions and format requirements
4. **Preview** shows exactly what will be generated before writing

You can also regenerate configs after the initial bootstrap:

```bash
alty generate configs
```

## Claude Code integration

alty generates a complete Claude Code workspace:

```
.claude/
├── CLAUDE.md              # Project instructions, conventions, quality gates
├── agents/                # AI agent personas
│   ├── developer.md       # Implementation-focused agent
│   ├── tech-lead.md       # Architecture review agent
│   ├── qa-engineer.md     # Testing and coverage agent
│   ├── researcher.md      # Spike and research agent
│   ├── project-manager.md # Ticket and backlog agent
│   └── white-hacker.md    # Security audit agent
└── commands/              # Slash commands
```

### CLAUDE.md

The generated `CLAUDE.md` contains:

- **Project overview** derived from your PRD
- **Ubiquitous language** — terms that must match code exactly
- **Architecture rules** — layer boundaries, dependency direction
- **Quality gates** — `go vet`, `golangci-lint`, `go test -race`
- **Beads workflow** — how to claim, work, and close tickets
- **Grooming checklist** — steps agents must follow before starting a ticket

### Agent personas

Each agent persona is tuned to your domain model. The developer agent knows your bounded context boundaries. The tech-lead agent enforces your architecture rules. The QA agent understands your domain invariants.

Generate or regenerate a specific persona:

```bash
# List available personas
alty persona list

# Generate a persona for Claude Code
alty persona generate developer

# Generate for a different tool
alty persona generate developer --tool cursor
```

## Cursor integration

alty generates Cursor-compatible rules and agent definitions in `.cursor/`. The same domain knowledge — bounded contexts, ubiquitous language, quality gates — is expressed in Cursor's rule format.

## Global settings detection

AI coding tools have global configs (e.g., `~/.claude/CLAUDE.md`) that can override local project settings. During `alty init`, alty scans for these and reports conflicts:

```
Global settings scan:
  OK  ~/.claude/CLAUDE.md — compatible with alty defaults
  CONFLICT  ~/.claude/settings.json has allowedTools restrictions
            Global restricts: Edit, Write require approval
            Local:  alty agents expect Edit, Write available
            WARNING: agents may hit permission prompts

            Options:
              [1] Keep global (agents will prompt for permissions)
              [2] Update global to allow Edit, Write
              [3] Note in local CLAUDE.md that agents need these tools
```

You choose the resolution per conflict. alty never silently creates local settings that will be overridden by global ones.

## Multi-tool workflows

If you use multiple AI coding tools on the same project, alty generates configs for all detected tools. The key guarantee: **same conventions, different formats**.

- Ubiquitous language is identical across all tool configs
- Quality gates run the same commands regardless of tool
- Bounded context boundaries are enforced the same way
- Agent personas have the same domain knowledge

This means you can switch between Claude Code and Cursor mid-project without losing architectural guardrails.

## MCP server

alty exposes its guided bootstrap and knowledge base as MCP (Model Context Protocol) tools. This allows AI coding tools that support MCP to call alty's capabilities directly:

- **Guide tools** — run guided discovery questions programmatically
- **Knowledge tools** — look up DDD patterns, tool conventions
- **Challenge tools** — probe domain models for gaps
- **Ticket tools** — verify ticket claims, check health

The MCP server is a separate entry point (`alty-mcp`) that shares the same application core as the CLI.
