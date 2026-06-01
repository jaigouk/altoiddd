---
title: New Project
description: Full walkthrough of bootstrapping a new project with alto — from README to tickets
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

## Step 2: Preview with `alto init`

Run `alto init` to see what alto will create:

```bash
alto init
```

alto shows a preview of every file it plans to create or install. Nothing is written until you confirm:

```
Detecting tools...
  Found: Claude Code (global config at ~/.claude/)
  Found: Beads (already installed)

Global settings scan:
  OK — no conflicts detected

Project files:
  CREATE  alto-scaffold/config.toml
  CREATE  alto-scaffold/knowledge/ddd/...          (12 files)
  CREATE  .claude/CLAUDE.md
  CREATE  .claude/agents/developer.md
  ...

Proceed? [y/N]
```

If alto detects conflicts between your global AI tool settings and what it wants to set locally, it shows each conflict and lets you choose a resolution. See [AI Tool Integration](/guides/ai-tool-integration) for details.

Use `--dry-run` to see the preview without any confirmation prompt:

```bash
alto init --dry-run
```

## Step 3: Guided discovery

After you confirm the preview, alto starts the Domain Storytelling discovery flow. This is the same flow you can run independently with `alto guide`.

### Persona detection

alto first asks which role best describes you:

- **Solo Developer** — building a project with AI assistance
- **Team Lead** — setting up conventions for a team
- **AI Tool Switcher** — using multiple AI coding tools
- **Product Owner** — defining what to build, not how
- **Domain Expert** — describing a business problem

Your choice determines the **register** — the language level alto uses in its storytelling conversation. Developers get technical DDD terminology. Product owners and domain experts get plain business language. The same Domain Storytelling flow extracts the same domain knowledge either way.

### Domain Storytelling

alto uses Domain Storytelling to discover your domain. Instead of abstract questions, alto proposes concrete stories about how your system works and invites you to refine them.

**StorytellingFlow** is selected by your DiscoveryMode:
- **ModeRapid** — 3 domain stories (default, ~15 minutes)
- **ModeThorough** — 5+ domain stories (~30-60 minutes, for complex domains)

Each story progresses through four **NarrationPhases**:

| Phase | What happens |
|-------|-------------|
| Opening | alto sets the scene: who is the actor, what are they trying to do |
| Narration | alto proposes **StorySentences** in the form [Actor] [activity] [WorkObject] |
| Deepening | alto probes for business rules, edge cases, and annotations |
| Closing | alto synthesizes what was learned and transitions to the next story |

For each proposed StorySentence, alto runs a **ConfirmSentence** loop: you can accept, reject, or edit the sentence. Mid-story synthesis checkpoints occur every 3 sentences to ensure alignment.

After all sentences in a story are confirmed, alto runs **ProposeStory** — a full replay of the complete story for your final confirmation.

Business rule **annotations** can be attached to any sentence (e.g., "invoice must be approved within 30 days").

Boundary detection produces **BoundedContextSketch[]** with confidence scores, which alto uses to generate your bounded context map.

Use `--no-tui` for plain stdin/stdout mode (useful for screen readers or scripted input):

```bash
alto guide --no-tui
```

## Step 4: Artifact generation

Once discovery completes, alto generates artifacts in a pipeline:

```
Domain stories
  → PRD (docs/PRD.md)
  → DDD artifacts (docs/DDD.md)
  → Architecture doc (docs/ARCHITECTURE.md)
  → Fitness tests (arch-go.yml)
  → Beads tickets (.beads/)
  → Tool configs (.claude/, .cursor/, etc.)
```

Each stage previews its output before writing. You approve or adjust at every step.

### What each artifact contains

**PRD** — Product requirements derived from your domain stories. Includes personas, scenarios, capabilities, and constraints.

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
alto check

# See what tickets are ready for implementation
bd ready
```

## Tips

- Respond to stories in your own language. alto builds the ubiquitous language from your words, not from developer jargon.
- The complexity budget matters. Not every subdomain needs full DDD treatment — let alto classify subdomains so you invest effort where it counts.
- Review the generated ubiquitous language glossary in `docs/DDD.md`. If a term doesn't match how you talk about the domain, correct it now. Code will use these names exactly.
