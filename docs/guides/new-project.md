---
title: New Project
description: Full walkthrough of bootstrapping a new project with alty — from README to tickets
sidebar:
  order: 3
---

This guide covers the complete flow of turning a project idea into a structured, production-ready project. For the abbreviated version, see the [Quickstart](/guides/quickstart).

## Step 1: Write your README

Create a new directory and describe your idea in 4-5 sentences:

```bash
mkdir invoice-tracker && cd invoice-tracker
git init
```

Write a `README.md` with your project idea. Be specific about what the software does, who uses it, and what problem it solves:

```markdown
# Invoice Tracker

A web service for freelancers to create, send, and track invoices.
Clients receive invoices via email with a payment link.
The system tracks payment status and sends automated reminders for overdue invoices.
Freelancers see a dashboard with revenue summaries and outstanding amounts.
Supports multiple currencies with automatic exchange rate lookup.
```

## Step 2: Preview with `alty init`

Run `alty init` to see what alty will create:

```bash
alty init
```

alty shows a preview of every file it plans to create or install. Nothing is written until you confirm:

```
Detecting tools...
  Found: Claude Code (global config at ~/.claude/)
  Found: Beads (already installed)

Global settings scan:
  OK — no conflicts detected

Project files:
  CREATE  .alty/config.toml
  CREATE  .alty/knowledge/ddd/...          (12 files)
  CREATE  .claude/CLAUDE.md
  CREATE  .claude/agents/developer.md
  ...

Proceed? [y/N]
```

If alty detects conflicts between your global AI tool settings and what it wants to set locally, it shows each conflict and lets you choose a resolution. See [AI Tool Integration](/guides/ai-tool-integration) for details.

Use `--dry-run` to see the preview without any confirmation prompt:

```bash
alty init --dry-run
```

## Step 3: Guided discovery

After you confirm the preview, alty starts the guided DDD discovery flow. This is the same flow you can run independently with `alty guide`.

### Persona detection

alty first asks which role best describes you:

- **Solo Developer** — building a project with AI assistance
- **Team Lead** — setting up conventions for a team
- **AI Tool Switcher** — using multiple AI coding tools
- **Product Owner** — defining what to build, not how
- **Domain Expert** — describing a business problem

Your choice determines the **register** — the language level alty uses in its questions. Developers get technical DDD terminology. Product owners and domain experts get plain business language. The same questions extract the same domain knowledge either way.

### The 10 questions

alty asks 10 questions in 5 phases:

| Phase | Questions | What it captures |
|-------|-----------|-----------------|
| Seed | Q1-Q2 | Core idea, tech stack |
| Actors | Q3-Q4 | Who uses the system, what roles exist |
| Story | Q5-Q6 | Key workflows, step-by-step processes |
| Events | Q7-Q8 | What happens in the domain, business rules |
| Boundaries | Q9-Q10 | Where the model splits, what terms are ambiguous |

After every 3-4 questions, alty plays back its understanding for you to confirm or correct. This playback loop prevents misunderstandings from compounding.

You can skip any question with an explicit acknowledgment, but alty requires at least 5 key questions (Q1, Q3, Q4, Q9, Q10) for a viable domain model.

Use `--no-tui` for plain stdin/stdout mode (useful for screen readers or scripted input):

```bash
alty guide --no-tui
```

## Step 4: Artifact generation

Once discovery completes, alty generates artifacts in a pipeline:

```
Discovery answers
  → PRD (docs/PRD.md)
  → DDD artifacts (docs/DDD.md)
  → Architecture doc (docs/ARCHITECTURE.md)
  → Fitness tests (arch-go.yml)
  → Beads tickets (.beads/)
  → Tool configs (.claude/, .cursor/, etc.)
```

Each stage previews its output before writing. You approve or adjust at every step.

### What each artifact contains

**PRD** — Product requirements derived from your answers. Includes personas, scenarios, capabilities, and constraints.

**DDD.md** — Domain model with:
- Domain stories (step-by-step business process narratives)
- Ubiquitous language glossary (terms that must match your code exactly)
- Bounded contexts with responsibilities
- Aggregate designs with invariants (for core subdomains)
- Subdomain classification (Core / Supporting / Generic)

**ARCHITECTURE.md** — Technical architecture informed by the domain model. Layer rules, dependency direction, port/adapter patterns.

**Fitness tests** — Executable architecture tests generated from your bounded context map. Core subdomains get strict rules; generic subdomains get minimal boundary checks.

**Beads tickets** — Dependency-ordered implementation tickets. Core subdomain tickets include full acceptance criteria, TDD phases, and SOLID mapping. Supporting tickets get standard detail. Generic tickets are stubs.

## Step 5: Start building

With your project seeded, hand it to your AI coding tool. The generated agent personas understand your domain model, enforce quality gates, and follow TDD.

```bash
# Check that quality gates are configured
alty check

# See what tickets are ready for implementation
bd ready
```

## Tips

- Answer questions in your own language. alty builds the ubiquitous language from your words, not from developer jargon.
- The complexity budget matters. Not every subdomain needs full DDD treatment — let alty classify subdomains so you invest effort where it counts.
- Review the generated ubiquitous language glossary in `docs/DDD.md`. If a term doesn't match how you talk about the domain, correct it now. Code will use these names exactly.
